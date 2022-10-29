package handlers

import (
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/shopspring/decimal"

	"gophermart-loyalty/internal/config"
	"gophermart-loyalty/internal/logger"
	"gophermart-loyalty/internal/middleware"
	"gophermart-loyalty/internal/usecases"
)

var timeFmt = time.RFC3339

func init() {
	decimal.MarshalJSONWithoutQuotes = true
}

// Handlers - HTTP-хандлеры для API
type Handlers struct {
	cfg      *config.Auth
	log      logger.Log
	useCases *usecases.UseCases
}

func NewHandlers(c *config.Auth, u *usecases.UseCases, log logger.Log) *Handlers {
	return &Handlers{
		useCases: u,
		cfg:      c,
		log:      log,
	}
}

func (h *Handlers) InitRoutes() chi.Router {
	r := chi.NewRouter()
	r.Post("/register", h.register)
	r.Post("/login", h.login)

	// Доступны только авторизованным пользователям
	r.Group(func(r chi.Router) {
		r.Use(middleware.Auth(h.cfg.SigningAlg, h.cfg.SigningKey))
		r.Post("/orders", h.orderAccrualCreate)
		r.Get("/orders", h.orderAccrualList)
		r.Post("/balance/withdraw", h.orderWithdrawalCreate)
		r.Get("/withdrawals", h.orderWithdrawalList)
		r.Post("/promos", h.promoAccrualCreate)
		r.Get("/balance", h.balanceGet)
		r.Get("/balance/history", h.balanceHistoryGet)
	})

	return r
}
