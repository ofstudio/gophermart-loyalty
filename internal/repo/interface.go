package repo

import (
	"context"

	"gophermart-loyalty/internal/models"
)

// Repo - интерфейс репозитория
type Repo interface {
	UserRepo
	OperationRepo
	PromoRepo
}

type UserRepo interface {
	// UserCreate - создает пользователя по логину и хэшу пароля
	UserCreate(ctx context.Context, u *models.User) error
	// UserGetByID - возвращает пользователя по id.
	UserGetByID(ctx context.Context, userID uint64) (*models.User, error)
	// UserGetByLogin - возвращает пользователя по логину.
	UserGetByLogin(ctx context.Context, login string) (*models.User, error)
	// UserBalanceHistoryGetByID - возвращает список операций пользователя, учитывающихся в балансе.
	UserBalanceHistoryGetByID(ctx context.Context, userID uint64) ([]*models.Operation, error)
}

type OperationRepo interface {
	// OperationCreate - создает операцию и обновляет баланс пользователя.
	OperationCreate(ctx context.Context, op *models.Operation) error
	// OperationUpdateFurther - берет самую старую операцию заданного типа,
	// которая находится не в конечном статусе, вызывает для нее коллбэк updateOp, обновляет операцию
	// и обновляет баланс пользователя.
	OperationUpdateFurther(ctx context.Context, opType models.OperationType, updateFunc UpdateFunc) (*models.Operation, error)
	// OperationGetByType - возвращает список операций пользователя заданного типа.
	OperationGetByType(ctx context.Context, userID uint64, t models.OperationType) ([]*models.Operation, error)
}

type PromoRepo interface {
	// PromoCreate - создает промо-кампанию.
	PromoCreate(ctx context.Context, p *models.Promo) error
	// PromoGetByCode - возвращает промо-кампанию по ее промо-коду.
	PromoGetByCode(ctx context.Context, code string) (*models.Promo, error)
}
