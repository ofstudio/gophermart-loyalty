package handlers

import (
	"net/http"
	"time"

	"github.com/go-chi/render"
	"github.com/shopspring/decimal"

	"gophermart-loyalty/internal/models"
)

// RegisterRequest - запрос на регистрацию пользователя Handlers.register.
type RegisterRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

func (req *RegisterRequest) Bind(_ *http.Request) error {
	return nil
}

// LoginRequest - запрос на аутентификацию пользователя Handlers.login.
type LoginRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

func (req *LoginRequest) Bind(_ *http.Request) error {
	return nil
}

// LoginResponse - ответ на запрос аутентификации пользователя Handlers.login.
type LoginResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int64  `json:"expires_in"`
}

func (res *LoginResponse) Render(_ http.ResponseWriter, _ *http.Request) error {
	return nil
}

// BalanceResponse - ответ на запрос баланса пользователя Handlers.balanceGet.
type BalanceResponse struct {
	Current   decimal.Decimal `json:"current"`
	Withdrawn decimal.Decimal `json:"withdrawn"`
}

func (b *BalanceResponse) Render(_ http.ResponseWriter, _ *http.Request) error {
	return nil
}

// BalanceHistoryResponse - ответ на запрос истории баланса пользователя Handlers.balanceHistoryGet.
type BalanceHistoryResponse struct {
	Amount      decimal.Decimal `json:"amount"`
	OrderNumber *string         `json:"number,omitempty"`
	Description string          `json:"description"`
	ProcessedAt string          `json:"processed_at"`
}

func (b *BalanceHistoryResponse) Render(_ http.ResponseWriter, _ *http.Request) error {
	return nil
}

func newBalanceHistoryResponse(ops []*models.Operation) []render.Renderer {
	list := make([]render.Renderer, len(ops))
	for i, op := range ops {
		list[i] = &BalanceHistoryResponse{
			Amount:      op.Amount,
			OrderNumber: op.OrderNumber,
			Description: op.Description,
			ProcessedAt: op.UpdatedAt.Format(timeFmt),
		}
	}
	return list
}

// OrderWithdrawalCreateRequest - запрос на создание операции списания бонусов Handlers.orderWithdrawalCreate.
type OrderWithdrawalCreateRequest struct {
	OrderNumber string          `json:"order"`
	Amount      decimal.Decimal `json:"sum"`
}

func (o *OrderWithdrawalCreateRequest) Bind(_ *http.Request) error {
	o.Amount = o.Amount.Neg()
	return nil
}

// OrderAccrualListResponse - ответ на запрос истории начислений бонусов Handlers.orderAccrualList.
type OrderAccrualListResponse struct {
	OrderNumber *string                `json:"number"`
	Status      models.OperationStatus `json:"status"`
	Amount      decimal.Decimal        `json:"accrual,omitempty"`
	CreatedAt   string                 `json:"uploaded_at"`
}

func (o *OrderAccrualListResponse) Render(_ http.ResponseWriter, _ *http.Request) error {
	return nil
}

func newOrderAccrualListResponse(ops []*models.Operation) []render.Renderer {
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

// OrderWithdrawalListResponse - ответ на запрос истории списаний бонусов Handlers.orderWithdrawalList.
type OrderWithdrawalListResponse struct {
	OrderNumber *string                `json:"order"`
	Status      models.OperationStatus `json:"status"`
	Amount      decimal.Decimal        `json:"sum"`
	UpdatedAt   time.Time              `json:"processed_at"`
}

func (o *OrderWithdrawalListResponse) Render(_ http.ResponseWriter, _ *http.Request) error {
	o.Amount = o.Amount.Neg() // меняем знак
	return nil
}
