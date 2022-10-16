package repo

import (
	"context"
	"gophermart-loyalty/internal/models"
)

// stmtBalanceHistoryGetByID - возвращает список операций пользователя, учитывающихся в балансе.
//    $1 - user_id
// Возвращает id, user_id, op_type, status, amount, description,
// order_number, promo_id, created_at, updated_at операции.
var stmtBalanceHistoryGetByID = registerStmt(`
	SELECT id, user_id, op_type, status, amount, description, order_number, promo_id, created_at, updated_at
	FROM operations
	WHERE user_id = $1 AND (
	    (status = 'PROCESSED' AND amount >= 0)
	    OR 
	    (status NOT IN ('INVALID', 'CANCELED') AND amount < 0)
	)
	ORDER BY updated_at DESC
`)

// BalanceHistoryGetByID - возвращает список операций пользователя, учитывающихся в балансе.
func (r *PGXRepo) BalanceHistoryGetByID(ctx context.Context, userID uint64) ([]*models.Operation, error) {
	rows, err := r.stmts[stmtBalanceHistoryGetByID].QueryContext(ctx, userID)
	if err != nil {
		return nil, r.handleError(ctx, err)
	}
	//goland:noinspection GoUnhandledErrorResult
	defer rows.Close()

	ops, err := r.operationScanRows(ctx, rows)
	if err != nil {
		return nil, err
	}
	return ops, nil
}
