package models

import (
	"time"

	"github.com/shopspring/decimal"
)

// Operation - модель операции
type Operation struct {
	ID          uint64
	UserID      uint64
	Type        OperationType
	Status      OperationStatus
	Amount      decimal.Decimal
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	OrderNumber *string // номер заказа, если операция связана с заказом
	PromoID     *uint64 // id промо-кампании, если операция связана с промо-кодом
}

// OperationType - тип операции
type OperationType string

const (
	OrderAccrual    OperationType = "order_accrual"
	OrderWithdrawal OperationType = "order_withdrawal"
	PromoAccrual    OperationType = "promo_accrual"
)

// OperationStatus - статус исполнения операции
type OperationStatus string

const (
	StatusNew        OperationStatus = "NEW"
	StatusProcessing OperationStatus = "PROCESSING"
	StatusInvalid    OperationStatus = "INVALID"
	StatusProcessed  OperationStatus = "PROCESSED"
	StatusCanceled   OperationStatus = "CANCELED"
)

// CanTransit - проверяет возможность перехода из статуса from в статус to.
// В проекте не используется, но может пригодиться в будущем.
func (s *OperationStatus) CanTransit(to OperationStatus) bool {
	if *s == to {
		return true
	}
	_, ok := statusGraph[*s][to]
	return ok
}

var statusGraph = map[OperationStatus]edges{
	StatusNew: {
		StatusProcessing: edge{},
		StatusInvalid:    edge{},
		StatusProcessed:  edge{},
		StatusCanceled:   edge{},
	},
	StatusProcessing: {
		StatusProcessed: edge{},
		StatusInvalid:   edge{},
		StatusCanceled:  edge{},
	},
	StatusInvalid:   {},
	StatusProcessed: {},
	StatusCanceled:  {},
}

type edge struct{}
type edges map[OperationStatus]edge
