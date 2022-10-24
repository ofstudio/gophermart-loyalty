package handlers

import (
	"errors"
	"net/http"

	"github.com/go-chi/render"

	"gophermart-loyalty/internal/errs"
	"gophermart-loyalty/internal/middleware"
)

// balanceGet - получение текущего баланса пользователя.
// Формат запроса:
//    GET /api/user/balance HTTP/1.1
//    Content-Length: 0
//    Authorization: Bearer <token>
//
// Возможные коды ответа:
//    200 — успешная обработка запроса
//    401 — пользователь не авторизован
//    500 — внутренняя ошибка сервера
//
// Формат ответа:
//    HTTP/1.1 200 OK
//    Content-Type: application/json
//
//    {
//    	"current": 500.5,
//    	"withdrawn": 42
//    }
func (h *Handlers) balanceGet(w http.ResponseWriter, r *http.Request) {
	// Получаем пользователя из контекста
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		_ = render.Render(w, r, errs.ErrResponseUnauthorized)
		return
	}

	// Запрашиваем пользователя
	user, err := h.useCases.UserGetByID(r.Context(), userID)
	if errors.Is(err, errs.ErrNotFound) {
		// Если пользователь не найден — возвращаем 500
		_ = render.Render(w, r, errs.ErrResponseInternal)
		return
	} else if err != nil {
		_ = render.Render(w, r, errs.NewErrResponse(err))
		return
	}

	// Отправляем ответ
	_ = render.Render(w, r, &BalanceResponse{
		Current:   user.Balance,
		Withdrawn: user.Withdrawn,
	})
}

// balanceHistoryGet - запрос истории операций по балансу пользователя.
// В ответе отображается только список тех операций, которые были изменяют баланс пользователя.
// Формат запроса:
//    GET /api/user/balance/history HTTP/1.1
//    Content-Length: 0
//    Authorization: Bearer <token>
//
// Возможные коды ответа:
//    200 — успешная обработка запроса
//	  204 — история операций пуста
//    401 — пользователь не авторизован
//    500 — внутренняя ошибка сервера
//
// Формат ответа:
//    HTTP/1.1 200 OK
//    Content-Type: application/json
//
//    [
//        {
//    	    "amount": -300,
//    	    "number": "12345678903",
//    	    "description": "Списание баллов за заказ 12345678903",
//          "processed_at": "2020-01-03T00:00:00Z"
//        },
//        {
//    	    "amount": 500.5,
//    	    "number": "9278923470",
//    	    "description": "Начисление баллов за заказ 9278923470",
//          "processed_at": "2020-01-02T00:00:00Z"
//        },
//        {
//    	    "amount": 100,
//    	    "description": "Начисление баллов по промо-коду WELCOME2020",
//          "processed_at": "2020-01-01T00:00:00Z"
//        }
//    ]
func (h *Handlers) balanceHistoryGet(w http.ResponseWriter, r *http.Request) {
	// Получаем пользователя из контекста
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		_ = render.Render(w, r, errs.ErrResponseUnauthorized)
		return
	}

	// Запрашиваем историю операций пользователя
	history, err := h.useCases.BalanceHistoryGetByID(r.Context(), userID)
	if err != nil {
		_ = render.Render(w, r, errs.NewErrResponse(err))
		return
	}

	// Если история пуста, возвращаем 204 No Content
	if len(history) == 0 {
		render.NoContent(w, r)
		return
	}

	// Отправляем ответ
	_ = render.RenderList(w, r, newBalanceHistoryResponse(history))
}
