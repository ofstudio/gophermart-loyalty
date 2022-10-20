package integrations

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/shopspring/decimal"
	"gophermart-loyalty/internal/app"
	"gophermart-loyalty/internal/config"
	"gophermart-loyalty/internal/logger"
	"gophermart-loyalty/internal/models"
	"gophermart-loyalty/internal/usecases"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"sync"
	"time"
)

const (
	AccrualStopped = iota
	AccrualRunning
)

// Accrual - интеграция с системой начисления бонусов
type Accrual struct {
	status       int
	useCases     *usecases.UseCases
	log          logger.Log
	client       *accrualClient
	mu           sync.Mutex
	pollInterval time.Duration // pollInterval - тайминг между запросами к системе начисления
	retryAfter   time.Duration // retryAfter - тайминг ожидания после получения ошибки TooManyRequests
	timingCh     chan struct{} // timingCh - сигнал об изменении таймингов после получения ошибки TooManyRequests
}

func NewAccrual(c *config.IntegrationAccrual, u *usecases.UseCases, log logger.Log) *Accrual {
	return &Accrual{
		status:       AccrualStopped,
		useCases:     u,
		log:          log,
		pollInterval: c.PollInterval,
		retryAfter:   0,
		timingCh:     make(chan struct{}),
		client:       newAccrualClient(c.Address+"/api/orders/", c.Timeout),
	}
}

// Start - запускает интеграцию
func (a *Accrual) Start(ctx context.Context) {
	go a.poll(ctx)
	a.status = AccrualRunning
}

func (a *Accrual) Status() int {
	return a.status
}

// poll - цикл обновления необработанных заказов по начислению баллов.
// Тайминг между обновлениями задается в конфигурации и может
// адаптироваться к сервису начисления в случае ошибки HTTP 429 Too Many Requests
func (a *Accrual) poll(ctx context.Context) {
	a.log.Info().Msg("accrual poller started")
	for {
		select {
		case <-ctx.Done():
			a.log.Info().Msg("accrual poller stopped")
			a.status = AccrualStopped
			return
		case <-a.timingCh:
			// Тайминги обновлены, поэтому необходимо повторно вызвать pollTiming(),
			// чтобы установить новый таймаут между запросами
			continue
		case <-time.After(a.pollTiming()):
			go func() { _ = a.updateFurther(ctx) }()
		}
	}
}

// updateFurther - запрашивает необработанные операции по начислению баллов и обновляет их статусы
func (a *Accrual) updateFurther(ctx context.Context) error {
	op, err := a.useCases.OperationUpdateFurther(ctx, models.OrderAccrual, a.updateCallback)
	if err == app.ErrNotFound {
		a.log.Debug().Msg("accrual operation: nothing to update")
		return nil
	} else if err != nil {
		a.log.Error().Err(err).Msg("accrual operation update failed")
		return err
	}
	a.log.Info().Uint64("operation_id", op.ID).Msg("accrual operation updated")
	return nil
}

// updateCallback - функция обновления статуса операции для OperationUpdateFurther
func (a *Accrual) updateCallback(ctx context.Context, op *models.Operation) error {
	if op.OrderNumber == nil {
		a.log.Error().Uint64("operation_id", op.ID).Msg("order number is nil")
		return app.ErrInternal
	}

	// Получаем статус заказа из системы начисления
	res, err := a.client.request(ctx, *op.OrderNumber)
	if err != nil && err.HTTPStatus == http.StatusTooManyRequests {
		// Если получили ошибку TooManyRequests, то обновляем тайминги
		a.adjustPollTiming(err.RetryAfter, err.MaxRPM)
		return app.ErrIntegrationTooManyRequests
	} else if err != nil {
		a.log.Error().Uint64("operation_id", op.ID).Err(err).Msg("accrual operation request failed")
		return app.ErrIntegrationRequestFailed
	}

	a.log.Info().Uint64("operation_id", op.ID).Msg("accrual operation request success")
	// Обновляем данные операции
	op.Status = res.Status
	op.Amount = res.Amount
	return nil
}

// pollTiming - возвращает тайминг для следующего запроса к системе начисления
func (a *Accrual) pollTiming() time.Duration {
	a.mu.Lock()
	defer a.mu.Unlock()
	// Если установлен retryAfter для повторного запроса, то используем его	для очередного запроса
	if a.retryAfter > 0 {
		t := a.retryAfter
		a.retryAfter = 0 // сбрасываем retryAfter для последующих запросов
		return t
	}
	return a.pollInterval
}

// adjustPollTiming - корректирует тайминги запросов к системе начисления
func (a *Accrual) adjustPollTiming(retryAfter time.Duration, maxRPM int) {
	if maxRPM == 0 {
		a.log.Error().Msg("max rpm is zero")
		return
	}
	// Обновляем тайминги
	a.mu.Lock()
	defer a.mu.Unlock()
	a.pollInterval = time.Minute / time.Duration(maxRPM)
	a.retryAfter = retryAfter
	// Отправляем сигнал в канал, чтобы пересчитался pollTiming
	a.timingCh <- struct{}{}
	a.log.Info().
		Str("poll_interval", a.pollInterval.String()).
		Str("retry_after", a.retryAfter.String()).
		Msg("poll timing adjusted")
}

// accrualClient - клиент для работы с системой начисления бонусов
type accrualClient struct {
	address string
	client  *http.Client
}

func newAccrualClient(address string, timeout time.Duration) *accrualClient {
	return &accrualClient{
		address: address,
		client:  &http.Client{Timeout: timeout},
	}
}

// request - отправляет запрос к системе начисления бонусов
func (c *accrualClient) request(ctx context.Context, orderNumber string) (*accrualResponse, *accrualError) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.address+orderNumber, nil)
	if err != nil {
		return nil, &accrualError{error: err}
	}
	req.Header.Set("Content-Length", "0")
	res, err := c.client.Do(req)
	if err != nil {
		return nil, &accrualError{error: err}
	}
	//goland:noinspection ALL
	defer res.Body.Close()

	if res.StatusCode == http.StatusTooManyRequests {
		return nil, c.parseTooManyRequests(res)
	} else if res.StatusCode != http.StatusOK {
		return nil, &accrualError{error: errors.New(http.StatusText(res.StatusCode)), HTTPStatus: res.StatusCode}
	}

	accrualRes := &accrualResponse{}
	err = json.NewDecoder(res.Body).Decode(accrualRes)
	if err != nil {
		return nil, &accrualError{error: err}
	}
	return accrualRes, nil
}

var rpmRe = regexp.MustCompile(`^No more than (\d+) requests per minute allowed`)

// parseTooManyRequests - парсит данные ответа 429 Too Many Requests в поля accrualError
func (c *accrualClient) parseTooManyRequests(res *http.Response) *accrualError {
	var err accrualError
	err.error = errors.New(http.StatusText(res.StatusCode))
	err.HTTPStatus = res.StatusCode

	// извлекаем Retry-After из заголовка
	retryAfter, _ := strconv.Atoi(res.Header.Get("Retry-After"))
	err.RetryAfter = time.Duration(retryAfter) * time.Second

	// извлекаем максимальное количество запросов в минуту из тела ответа
	body, _ := ioutil.ReadAll(res.Body)
	if matches := rpmRe.FindStringSubmatch(string(body)); len(matches) > 1 {
		rpm, _ := strconv.Atoi(matches[1])
		err.MaxRPM = rpm
	}
	return &err
}

// accrualResponse - ответ системы начисления бонусов
type accrualResponse struct {
	OrderNumber string                 `json:"order"`
	Status      models.OperationStatus `json:"status"`
	Amount      decimal.Decimal        `json:"accrual"`
}

// accrualError - ошибка при обращении к системе начисления бонусов
type accrualError struct {
	error
	HTTPStatus int
	RetryAfter time.Duration
	MaxRPM     int
}
