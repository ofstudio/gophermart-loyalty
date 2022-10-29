package repo

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jackc/pgconn"

	"gophermart-loyalty/internal/errs"
)

// constraintToAppError - соответствие нарушений ограничений БД ошибкам приложения
var constraintToAppError = map[string]*errs.Error{
	"username_unique":        errs.ErrUserAlreadyExists,     // логин должен быть уникальным
	"balance_not_negative":   errs.ErrUserBalanceNegative,   // общая сумма на счете не может быть отрицательной
	"withdrawn_not_negative": errs.ErrUserWithdrawnNegative, // общая сумма списаний не может быть отрицательной

	"operation_valid_attrs":    errs.ErrOperationAttrsInvalid,    // аттрибуты операции должны соответствовать типу операции
	"amount_valid_sign":        errs.ErrOperationAmountInvalid,   // зачисления должны иметь положительные значения, а списания - отрицательные
	"must_refs_user":           errs.ErrOperationUserNotExists,   // операция должна ссылаться на существующего пользователя
	"order_belongs_to_user":    errs.ErrOperationOrderNotBelongs, // номер заказа может принадлежать только одному пользователю
	"order_unique_for_op_type": errs.ErrOperationOrderUsed,       // по заказу возможна 1 операция списания баллов и 1 операция зачисления баллов
	"must_refs_promo":          errs.ErrNotFound,                 // операция зачисления по промо-кампании должна ссылаться на существующую промо-кампанию
	"promo_unique_for_user":    errs.ErrOperationPromoUsed,       // пользователь может воспользоваться промо-кампанией не более 1 раза

	"promo_code_unique":     errs.ErrPromoAlreadyExists,     // промо-кампания должна иметь уникальный код
	"promo_reward_positive": errs.ErrPromoRewardNotPositive, // вознаграждение за промо-кампанию должно быть положительным
	"promo_valid_period":    errs.ErrPromoPeriodInvalid,     // дата начала промо-кампании должна быть меньше даты окончания
}

func (r *PGXRepo) handleError(ctx context.Context, err error) error {
	log := r.log.WithReqID(ctx)

	// Если ошибки нет, то возвращаем nil
	if err == nil {
		return nil
	}

	// Если ошибка типа *errs.Error, то возвращаем ее
	if appErr, ok := err.(*errs.Error); ok {
		return appErr
	}

	// Если не найдены записи, то возвращаем ErrNotFound
	if errors.Is(err, sql.ErrNoRows) {
		return errs.ErrNotFound
	}

	// Проверяем, является ли ошибка нарушением ограничения БД
	if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.ConstraintName != "" {
		// Если нарушение ограничения известно, то возвращаем соответствующую ошибку
		if appErr, ok := constraintToAppError[pgErr.ConstraintName]; ok {
			return appErr
		}
	}

	// Если другая ошибка или неизвестное ограничение, то возвращаем ErrInternal
	log.Error().CallerSkipFrame(1).Err(err).Msg("repo internal error")
	return errs.ErrInternal
}
