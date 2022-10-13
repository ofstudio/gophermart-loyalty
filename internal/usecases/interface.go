package usecases

import (
	"context"
	"gophermart-loyalty/internal/models"
)

type UseCasesInterface interface {
	UserCreate(ctx context.Context, login, password string) (*models.User, error)
	UserCheckLoginPass(ctx context.Context, login, password string) (*models.User, error)
	UserGetByID(ctx context.Context, id uint64) (*models.User, error)
}
