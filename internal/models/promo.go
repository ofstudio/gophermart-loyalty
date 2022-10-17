package models

import (
	"github.com/shopspring/decimal"
	"time"
)

// Promo - модель промо-кампании
type Promo struct {
	ID          uint64
	Code        string
	Description string
	Reward      decimal.Decimal
	NotBefore   time.Time
	NotAfter    time.Time
	CreatedAt   time.Time
}
