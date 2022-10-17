package models

import (
	"github.com/shopspring/decimal"
	"time"
)

// User - модель пользователя
type User struct {
	ID        uint64
	Login     string
	PassHash  string
	Balance   decimal.Decimal
	Withdrawn decimal.Decimal
	CreatedAt time.Time
	UpdatedAt time.Time
}
