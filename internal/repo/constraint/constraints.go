// Package constraint - ограничения (constraints) на уровне БД.
package constraint

import "github.com/jackc/pgconn"

type Constraint string

const (
	UsernameUnique       Constraint = "username_unique"
	MustRefsUser         Constraint = "must_refs_user"
	AmountValidSign      Constraint = "amount_valid_sign"
	OrderBelongsToUser   Constraint = "order_belongs_to_user"
	OrderUniqueForType   Constraint = "order_unique_for_op_type"
	MustRefsPromo        Constraint = "must_refs_promo"
	PromoUniqueForUser   Constraint = "promo_unique_for_user"
	OperationValidAttrs  Constraint = "operation_valid_attrs"
	BalanceNotNegative   Constraint = "balance_not_negative"
	WithdrawnNotNegative Constraint = "withdrawn_not_negative"
	PromoCodeUnique      Constraint = "promo_code_unique"
	PromoRewardPositive  Constraint = "promo_reward_positive"
)

// Violated - возвращает имя ограничения, если ошибка запроса к БД была вызвана нарушением ограничения БД.
func Violated(err error) (Constraint, bool) {
	if err == nil {
		return "", false
	}
	if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.ConstraintName != "" {
		return Constraint(pgErr.ConstraintName), true
	}
	return "", false
}
