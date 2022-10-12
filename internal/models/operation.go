package models

import (
	"github.com/shopspring/decimal"
	"time"
)

type Operation struct {
	ID          uint64
	UserID      uint64
	Type        OperationType
	Status      OperationStatus
	Amount      decimal.Decimal
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	OrderNumber *string
	PromoID     *uint64
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
