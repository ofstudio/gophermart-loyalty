package repo

import (
	"database/sql"
	"github.com/jackc/pgconn"
	"gophermart-loyalty/internal/app"
)

// constraintToAppError - соответствие нарушений ограничений БД ошибкам приложения
var constraintToAppError = map[string]*app.Error{
	"username_unique":        app.ErrUserAlreadyExists,     // логин должен быть уникальным
	"balance_not_negative":   app.ErrUserBalanceNegative,   // общая сумма на счете не может быть отрицательной
	"withdrawn_not_negative": app.ErrUserWithdrawnNegative, // общая сумма списаний не может быть отрицательной

	"operation_valid_attrs":    app.ErrOperationAttrsInvalid,    // аттрибуты операции должны соответствовать типу операции
	"amount_valid_sign":        app.ErrOperationAmountInvalid,   // зачисления должны иметь положительные значения, а списания - отрицательные
	"must_refs_user":           app.ErrOperationUserNotExists,   // операция должна ссылаться на существующего пользователя
	"order_belongs_to_user":    app.ErrOperationOrderNotBelongs, // номер заказа может принадлежать только одному пользователю
	"order_unique_for_op_type": app.ErrOperationOrderUsed,       // по заказу возможна 1 операция списания баллов и 1 операция зачисления баллов
	"must_refs_promo":          app.ErrOperationPromoNotExists,  // операция зачисления по промо-кампании должна ссылаться на существующую промо-кампанию
	"promo_unique_for_user":    app.ErrOperationPromoUsed,       // пользователь может воспользоваться промо-кампанией не более 1 раза

	"promo_code_unique":     app.ErrPromoAlreadyExists,     // промо-кампания должна иметь уникальный код
	"promo_reward_positive": app.ErrPromoRewardNotPositive, // вознаграждение за промо-кампанию должно быть положительным
	"promo_valid_period":    app.ErrPromoPeriodInvalid,     // дата начала промо-кампании должна быть меньше даты окончания
}

func (r *Repo) appError(err error) *app.Error {
	// Если ошибки нет, то возвращаем nil
	if err == nil {
		return nil
	}

	// Если ошибка типа *app.Error, то возвращаем ее
	if appErr, ok := err.(*app.Error); ok {
		return appErr
	}

	// Если не найдены записи, то возвращаем ErrNotFound
	if err == sql.ErrNoRows {
		return app.ErrNotFound
	}

	// Проверяем, является ли ошибка нарушением ограничения БД
	if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.ConstraintName != "" {
		// Если нарушение ограничения известно, то возвращаем соответствующую ошибку
		if appErr, ok := constraintToAppError[pgErr.ConstraintName]; ok {
			return appErr
		} else {
			// Иначе выводим предупреждение в лог
			r.log.Warn().Str("constraint", pgErr.ConstraintName).Msg("unknown constraint violation")
		}
	}

	// Если другая ошибка или неизвестное ограничение, то возвращаем ErrInternal
	return app.ErrInternal
}
