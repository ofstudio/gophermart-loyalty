package repo

import (
	"context"
	"database/sql"
	"fmt"
	_ "github.com/jackc/pgx/v4/stdlib"
	"gophermart-loyalty/internal/config"
	"gophermart-loyalty/internal/errs"
	"gophermart-loyalty/internal/logger"
	"gophermart-loyalty/internal/repo/constraint"
	"gophermart-loyalty/internal/repo/migrations"
)

// Repo - репозиторий для работы с БД.
type Repo struct {
	db    *sql.DB
	log   logger.Log
	stmts []*sql.Stmt
}

// NewRepo - создает новый репозиторий
func NewRepo(cfg config.DB, log logger.Log) (*Repo, error) {
	var err error

	// Подключаемся к БД
	db, err := sql.Open("pgx", cfg.DatabaseURI)
	if err != nil || db.Ping() != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}
	log.Info().Msg("connected to db")

	// Запускаем миграцию БД
	ver, err := migrations.Migrate(db)
	if err != nil {
		return nil, err
	}

	// Проверяем что версия БД соответствует заданной
	if ver != cfg.RequiredVersion {
		return nil, fmt.Errorf("db version mismatch: got %d, want %d", ver, cfg.RequiredVersion)
	}
	log.Info().Msgf("db migrated to version %d", ver)

	// Подготавливаем стейтменты
	stmts, err := prepareStmts(db)
	if err != nil {
		return nil, err
	}
	log.Info().Msg("statements prepared")
	return &Repo{db: db, log: log, stmts: stmts}, nil
}

// DB - возвращает соединение с БД.
func (r *Repo) DB() *sql.DB {
	return r.db
}

// Close - закрывает соединение с БД.
func (r *Repo) Close() error {
	return r.db.Close()
}

// constraintErr - возвращает бизнес-ошибку, соответствующую ограничению БД.
// Название нарушенного ограничения возможно получить с помощью constraint.Violated.
func (r *Repo) constraintErr(ctx context.Context, c constraint.Constraint) error {
	log := r.log.WithRequestID(ctx).With().Str("constraint", string(c)).Logger()
	switch c {
	// Имя пользователя (логин) должно быть уникальным
	case constraint.UsernameUnique:
		return errs.Duplicate

	// Операция должна ссылаться на существующего пользователя
	case constraint.MustRefsUser:
		return errs.NonExistingUser

	// Операция по промо-зачислению должна ссылаться на действительный промо
	case constraint.MustRefsPromo:
		return errs.NonExistingPromo

	// Зачисления должны иметь положительные значения, а списания - отрицательные
	case constraint.AmountValidSign:
		return errs.Validation

	// Заказ (номер заказа) может принадлежать только одному пользователю
	case constraint.OrderBelongsToUser:
		return errs.Conflict

	// По заказу возможна 1 операция списания баллов и 1 операция зачисления баллов
	case constraint.OrderUniqueForType:
		return errs.Duplicate

	// Баланс пользователя не должен быть отрицательным
	case constraint.BalanceNotNegative:
		return errs.InsufficientBalance

	// Пользователь может воспользоваться промо-кампанией не более 1 раза
	case constraint.PromoUniqueForUser:
		return errs.Duplicate

	// Промо-кампания должна иметь уникальный код
	case constraint.PromoCodeUnique:
		return errs.Duplicate

	// Вознаграждение за промо-кампанию должно быть положительным
	case constraint.PromoRewardPositive:
		return errs.Validation

	// Следующие кейсы могут возникать в случае некорректной работы
	// приложения, либо неконсистентных данных в БД.
	//    WithdrawnNotNegative - общая сумма списаний не может быть отрицательной.
	//    OperationValidAttrs - аттрибуты операции должны соответствовать типу операции
	case
		constraint.WithdrawnNotNegative,
		constraint.OperationValidAttrs:
		log.Error().Msg("unexpected constraint violation")
		return errs.Internal

	// Неизвестное ограничение БД
	default:
		log.Error().Msg("unknown constraint violation")
		return errs.Internal
	}
}
