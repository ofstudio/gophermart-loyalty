package handlers

import (
	"errors"
	"github.com/go-chi/render"
	"github.com/shopspring/decimal"
	"gophermart-loyalty/internal/app"
	"gophermart-loyalty/internal/models"
	"net/http"
	"time"
)

// orderAccrualCreate - загрузка номера заказа для зачисления баллов.
// Формат запроса:
//    POST /api/user/orders HTTP/1.1
//    Content-Type: text/plain
//
//    12345678903
//
// Возможные коды ответа:
//    200 — номер заказа уже был загружен этим пользователем
//    202 — новый номер заказа принят в обработку
//    400 — неверный формат запроса
//    401 — пользователь не аутентифицирован;
//    409 — номер заказа уже был загружен другим пользователем
//    422 — неверный формат номера заказа
//    500 — внутренняя ошибка сервера
func (h *Handlers) orderAccrualCreate(w http.ResponseWriter, r *http.Request) {
	// Получаем пользователя из контекста
	userID, ok := h.getUserID(r.Context())
	if !ok {
		_ = render.Render(w, r, ErrUnauthorized)
		return
	}

	// Получаем номер заказа из запроса
	orderNumber, err := decodePlainText(r)
	if err != nil {
		_ = render.Render(w, r, NewErrResponse(err))
		return
	}

	// Создаем модель операции
	op, err := h.useCases.OrderAccrualPrepare(r.Context(), userID, orderNumber)
	if err != nil {
		_ = render.Render(w, r, NewErrResponse(err))
		return
	}
	// Сохраняем операцию
	err = h.useCases.OperationCreate(r.Context(), op)
	if errors.Is(err, app.ErrOperationOrderUsed) {
		// если номер заказа уже использовался этим пользователем для начисления бонусов,
		// то возвращаем 200 ОК
		w.WriteHeader(http.StatusOK)
		return
	} else if err != nil {
		_ = render.Render(w, r, NewErrResponse(err))
		return
	}
	w.WriteHeader(http.StatusCreated)
}

type OrderWithdrawalCreateRequest struct {
	OrderNumber string          `json:"order"`
	Amount      decimal.Decimal `json:"sum"`
}

func (o *OrderWithdrawalCreateRequest) Bind(_ *http.Request) error {
	o.Amount = o.Amount.Neg()
	return nil
}

// orderWithdrawalCreate - создание операции списания бонусов.
// Формат запроса:
//    POST /api/user/withdraw HTTP/1.1
//    Content-Type: application/json
//
//    {
//	   "order": "2377225624",
//     "sum": 751
//    }
//
// Возможные коды ответа:
//    200 — успешная обработка запроса
//	  400 — неверный формат запроса
//    401 — пользователь не авторизован
//    402 — на счету недостаточно средств
//	  409 — номер заказа уже был загружен другим пользователем
//    422 — неверный номер заказа
//    500 — внутренняя ошибка сервера
func (h *Handlers) orderWithdrawalCreate(w http.ResponseWriter, r *http.Request) {
	// Получаем пользователя из контекста
	userID, ok := h.getUserID(r.Context())
	if !ok {
		_ = render.Render(w, r, ErrUnauthorized)
		return
	}

	// Получаем данные из запроса
	data := &OrderWithdrawalCreateRequest{}
	if err := render.Bind(r, data); err != nil {
		_ = render.Render(w, r, ErrBadRequest)
		return
	}

	// Создаем модель операции
	op, err := h.useCases.OrderWithdrawalPrepare(r.Context(), userID, data.OrderNumber, data.Amount)
	if err != nil {
		_ = render.Render(w, r, NewErrResponse(err))
		return
	}

	// Сохраняем операцию
	if err = h.useCases.OperationCreate(r.Context(), op); err != nil {
		_ = render.Render(w, r, NewErrResponse(err))
		return
	}
	w.WriteHeader(http.StatusOK)
}

// promoAccrualCreate - загрузка промо-кода для зачисления баллов.
// Формат запроса:
//    POST /api/user/promos HTTP/1.1
//    Content-Type: text/plain
//
//    WELCOME2020
//
// Возможные коды ответа:
//    200 — успешная обработка запроса
//    400 — неверный формат запроса
//    404 — промо-код не найден
//    409 — пользователь может воспользоваться промо-кампанией не более 1 раза
//    500 — внутренняя ошибка сервера
func (h *Handlers) promoAccrualCreate(w http.ResponseWriter, r *http.Request) {
	// Получаем пользователя из контекста
	userID, ok := h.getUserID(r.Context())
	if !ok {
		_ = render.Render(w, r, ErrUnauthorized)
		return
	}

	// Получаем промо-код из запроса
	promoCode, err := decodePlainText(r)
	if err != nil {
		_ = render.Render(w, r, NewErrResponse(err))
		return
	}

	// Создаем модель операции
	op, err := h.useCases.PromoAccrualPrepare(r.Context(), userID, promoCode)
	if err != nil {
		_ = render.Render(w, r, NewErrResponse(err))
		return
	}

	// Сохраняем операцию
	if err = h.useCases.OperationCreate(r.Context(), op); err != nil {
		_ = render.Render(w, r, NewErrResponse(err))
		return
	}
	w.WriteHeader(http.StatusOK)
}

type OrderAccrualListResponse struct {
	OrderNumber *string                `json:"number"`
	Status      models.OperationStatus `json:"status"`
	Amount      decimal.Decimal        `json:"accrual,omitempty"`
	CreatedAt   string                 `json:"uploaded_at"`
}

func (o *OrderAccrualListResponse) Render(_ http.ResponseWriter, _ *http.Request) error {
	return nil
}

func NewOrderAccrualListResponse(ops []*models.Operation) []render.Renderer {
	list := make([]render.Renderer, len(ops))
	for i, op := range ops {
		list[i] = &OrderAccrualListResponse{
			OrderNumber: op.OrderNumber,
			Status:      op.Status,
			Amount:      op.Amount,
			CreatedAt:   op.CreatedAt.Format(timeFmt),
		}
	}
	return list
}

// orderAccrualList - получение списка загруженных номеров заказов.
// Формат запроса:
//    GET /api/user/orders HTTP/1.1
//    Content-Length: 0
//
// Возможные коды ответа:
//    200 — успешная обработка запроса
//    204 — нет данных для ответа.
//    401 — пользователь не авторизован.
//    500 — внутренняя ошибка сервера.
//
// Формат ответа:
//    200 OK HTTP/1.1
//    Content-Type: application/json
//
//    [
//    	{
//            "number": "9278923470",
//            "status": "PROCESSED",
//            "accrual": 500,
//            "uploaded_at": "2020-12-10T15:15:45+03:00"
//        },
//        {
//            "number": "12345678903",
//            "status": "PROCESSING",
//            "uploaded_at": "2020-12-10T15:12:01+03:00"
//        },
//        {
//            "number": "346436439",
//            "status": "INVALID",
//            "uploaded_at": "2020-12-09T16:09:53+03:00"
//        }
//    ]
func (h *Handlers) orderAccrualList(w http.ResponseWriter, r *http.Request) {
	// Получаем пользователя из контекста
	userID, ok := h.getUserID(r.Context())
	if !ok {
		_ = render.Render(w, r, ErrUnauthorized)
		return
	}

	// получаем список операций начисления бонусов
	operations, err := h.useCases.OperationGetByType(r.Context(), userID, models.OrderAccrual)
	if err != nil {
		_ = render.Render(w, r, NewErrResponse(err))
		return
	}

	if len(operations) == 0 {
		render.NoContent(w, r)
		return
	}

	_ = render.RenderList(w, r, NewOrderAccrualListResponse(operations))

}

type OrderWithdrawalListResponse struct {
	OrderNumber *string         `json:"order"`
	Amount      decimal.Decimal `json:"sum"`
	UpdatedAt   time.Time       `json:"processed_at"`
}

func (o *OrderWithdrawalListResponse) Render(_ http.ResponseWriter, _ *http.Request) error {
	o.Amount = o.Amount.Neg() // меняем знак
	return nil
}

func NewOrderWithdrawalListResponse(ops []*models.Operation) []render.Renderer {
	list := make([]render.Renderer, len(ops))
	for i, op := range ops {
		list[i] = &OrderWithdrawalListResponse{
			OrderNumber: op.OrderNumber,
			Amount:      op.Amount,
			UpdatedAt:   op.UpdatedAt,
		}
	}
	return list
}

// orderWithdrawalList - получение информации о выводе средств.
// Формат запроса:
//    GET /api/user/balance/withdrawals HTTP/1.1
//    Content-Length: 0
//
// Возможные коды ответа:
//    200 — успешная обработка запроса
//    204 — нет ни одного списания
//    401 — пользователь не авторизован
//    500 — внутренняя ошибка сервера
//
// Формат ответа:
//    HTTP/1.1 200 OK
//    Content-Type: application/json
//
//    [
//        {
//            "order": "2377225624",
//            "sum": 500,
//            "processed_at": "2020-12-09T16:09:57+03:00"
//        }
//    ]
func (h *Handlers) orderWithdrawalList(w http.ResponseWriter, r *http.Request) {
	// Получаем пользователя из контекста
	userID, ok := h.getUserID(r.Context())
	if !ok {
		_ = render.Render(w, r, ErrUnauthorized)
		return
	}

	// Получаем список операций
	ops, err := h.useCases.OperationGetByType(r.Context(), userID, models.OrderWithdrawal)
	if err != nil {
		_ = render.Render(w, r, NewErrResponse(err))
		return
	}

	if len(ops) == 0 {
		render.NoContent(w, r)
		return
	}
	_ = render.RenderList(w, r, NewOrderWithdrawalListResponse(ops))
}
