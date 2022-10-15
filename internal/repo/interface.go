package repo

import (
	"context"
	"database/sql"
	"gophermart-loyalty/internal/models"
)

// RepoInterface - интерфейс репозитория.
type RepoInterface interface {
	UserCreate(ctx context.Context, u *models.User) error
	UserGetByID(ctx context.Context, userID uint64) (*models.User, error)
	UserGetByLogin(ctx context.Context, login string) (*models.User, error)
	DB() *sql.DB
	Close() error
	OperationCreate(ctx context.Context, op *models.Operation) error
	OperationUpdateFurther(ctx context.Context, opType models.OperationType, updateFunc UpdateFunc) error
	OperationGetByType(ctx context.Context, userID uint64, t models.OperationType) ([]*models.Operation, error)
	OperationGetBalance(ctx context.Context, userID uint64) ([]*models.Operation, error)
	PromoCreate(ctx context.Context, p *models.Promo) error
	PromoGetByCode(ctx context.Context, code string) (*models.Promo, error)
}
