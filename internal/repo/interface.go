package repo

import (
	"context"
	"database/sql"
	"gophermart-loyalty/internal/models"
)

// Repo - интерфейс репозитория.
type Repo interface {
	// DB - возвращает соединение с БД.
	DB() *sql.DB
	// Close - закрывает соединение с БД.
	Close() error
	// UserCreate - создает пользователя по логину и хэшу пароля
	UserCreate(ctx context.Context, u *models.User) error
	// UserGetByID - возвращает пользователя по id.
	UserGetByID(ctx context.Context, userID uint64) (*models.User, error)
	// UserGetByLogin - возвращает пользователя по логину.
	UserGetByLogin(ctx context.Context, login string) (*models.User, error)
	// OperationCreate - создает операцию и обновляет баланс пользователя.
	OperationCreate(ctx context.Context, op *models.Operation) error
	// OperationUpdateFurther - берет самую старую операцию заданного типа,
	// которая находится не в конечном статусе, вызывает для нее коллбэк updateOp, обновляет операцию
	// и обновляет баланс пользователя.
	OperationUpdateFurther(ctx context.Context, opType models.OperationType, updateFunc UpdateFunc) error
	// OperationGetByType - возвращает список операций пользователя заданного типа.
	OperationGetByType(ctx context.Context, userID uint64, t models.OperationType) ([]*models.Operation, error)
	// BalanceHistoryGetByID - возвращает список операций пользователя, учитывающихся в балансе.
	BalanceHistoryGetByID(ctx context.Context, userID uint64) ([]*models.Operation, error)
	// PromoCreate - создает промо-кампанию.
	PromoCreate(ctx context.Context, p *models.Promo) error
	// PromoGetByCode - возвращает промо-кампанию по ее промо-коду.
	PromoGetByCode(ctx context.Context, code string) (*models.Promo, error)
}
