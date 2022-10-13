package handlers

import (
	"github.com/go-chi/chi/v5"
	"gophermart-loyalty/internal/config"
	"gophermart-loyalty/internal/logger"
	"gophermart-loyalty/internal/usecases"
)

type Handlers struct {
	useCases usecases.UseCasesInterface
	cfgAuth  *config.Auth
	log      logger.Log
}

func NewHandlers(u usecases.UseCasesInterface, c *config.Auth, log logger.Log) *Handlers {
	return &Handlers{
		useCases: u,
		cfgAuth:  c,
		log:      log,
	}
}

func (h *Handlers) Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/register", h.register)
	r.Post("/login", h.login)
	// todo add auth middleware
	r.Post("/orders", h.orderAccrualCreate)
	r.Get("/orders", h.orderAccrualList)
	r.Post("/promos", h.promoAccrualCreate)
	r.Get("/balance", h.balanceGet)
	r.Get("/balance/details", h.balanceDetailsGet)
	r.Post("/balance/withdraw", h.orderWithdrawalCreate)
	r.Get("/balance/withdrawals", h.orderWithdrawalList)

	return r
}
