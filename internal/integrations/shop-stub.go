package integrations

import (
	"context"
	"strings"
	"time"

	"gophermart-loyalty/internal/errs"
	"gophermart-loyalty/internal/logger"
	"gophermart-loyalty/internal/models"
	"gophermart-loyalty/internal/usecases"
)

const (
	ShopStubStopped = iota
	ShopStubRunning
)

// IntegrationShopStub - эмулятор интеграции с магазином в части оплаты заказов баллами.
// Реализован в качестве демонстрации.
//
// Переводит операции по списанию баллов в конечный статус через 1 минуту после создания операции.
//
// Операции с номерами заказа, начинающимися с `000`, переводятся в статус CANCELED.
// Все остальные операции по списанию баллов переводятся в статус PROCESSED.
type IntegrationShopStub struct {
	status       int
	useCases     *usecases.UseCases
	log          logger.Log
	pollInterval time.Duration
}

func NewIntegrationShopStub(u *usecases.UseCases, log logger.Log) *IntegrationShopStub {
	return &IntegrationShopStub{
		status:       ShopStubStopped,
		useCases:     u,
		log:          log,
		pollInterval: 1 * time.Second,
	}
}

// Start - запускает эмулятор интеграции с магазином.
func (s *IntegrationShopStub) Start(ctx context.Context) {
	go s.poll(ctx)
	s.status = ShopStubRunning
}

// Status - возвращает статус эмулятора интеграции с магазином.
func (s *IntegrationShopStub) Status() int {
	return s.status
}

// poll - цикл обновления необработанных заказов по списанию баллов
func (s *IntegrationShopStub) poll(ctx context.Context) {
	s.log.Info().Msg("shop integration started")
	for {
		select {
		case <-ctx.Done():
			s.log.Info().Msg("shop integration stopped")
			s.status = ShopStubStopped
			return
		case <-time.After(s.pollInterval):
			go s.updateFurther(ctx)
		}
	}
}

// updateFurther - запрашивает необработанные операции по списанию баллов и обновляет их статусы
func (s *IntegrationShopStub) updateFurther(ctx context.Context) {
	op, err := s.useCases.OperationUpdateFurther(ctx, models.OrderWithdrawal, func(ctx context.Context, op *models.Operation) error {
		if op.OrderNumber == nil {
			s.log.Error().Uint64("operation_id", op.ID).Msg("order number is nil")
			return errs.ErrInternal
		}
		if time.Since(op.CreatedAt) > time.Minute {
			if strings.HasPrefix(*op.OrderNumber, "000") {
				op.Status = models.StatusCanceled
			} else {
				op.Status = models.StatusProcessed
			}
		} else {
			return errs.ErrNotFound
		}
		return nil
	})

	if err == errs.ErrNotFound {
		s.log.Debug().Msg("withdrawal operation: nothing to update")
		return
	}
	if err != nil {
		s.log.Error().Err(err).Msg("withdrawal operation update failed")
		return
	}
	s.log.Info().Uint64("operation_id", op.ID).Msg("withdrawal operation updated")
}
