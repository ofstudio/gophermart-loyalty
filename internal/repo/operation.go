package repo

import (
	"context"
	"database/sql"
	"errors"
	"gophermart-loyalty/internal/app"
	"gophermart-loyalty/internal/models"
)

// stmtOperationCreate - создает операцию.
//    $1 - user_id
//    $2 - op_type
//    $3 - status
//    $4 - amount
//    $5 - description
//    $6 - order_number
//    $7 - promo_id
// Возвращает id, created_at, updated_at операции.
// ВАЖНО: может вызываться только внутри транзакции и только после вызова Repo.userLockTx.
// После вызова необходимо обновить баланс пользователя при помощи Repo.userUpdateBalanceTx.
var stmtOperationCreate = registerStmt(`
	INSERT INTO operations (user_id, op_type, status, amount, description, order_number, promo_id)
	VALUES ($1, $2, $3, $4, $5, $6, $7)
	RETURNING id, created_at, updated_at
`)

// OperationCreate - создает операцию и обновляет баланс пользователя.
func (r *Repo) OperationCreate(ctx context.Context, op *models.Operation) error {

	tx, err := r.db.Begin()
	if err != nil {
		return r.handleError(ctx, err)
	}
	//goland:noinspection ALL
	defer tx.Rollback()

	// Блокируем запись пользователя для обновления
	if err = r.userLockTx(ctx, tx, op.UserID); err != nil {
		if errors.Is(err, app.ErrNotFound) {
			err = app.ErrOperationUserNotExists
		}
		return err
	}

	// Создаем операцию
	err = tx.Stmt(r.stmts[stmtOperationCreate]).
		QueryRowContext(ctx, op.UserID, op.Type, op.Status, op.Amount, op.Description, op.OrderNumber, op.PromoID).
		Scan(&op.ID, &op.CreatedAt, &op.UpdatedAt)
	if err != nil {
		return r.handleError(ctx, err)
	}

	// Обновляем баланс пользователя
	if err = r.userUpdateBalanceTx(ctx, tx, op.UserID); err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		return r.handleError(ctx, err)
	}

	return nil
}

type UpdateFunc func(ctx context.Context, operation *models.Operation) error

// stmtOperationLockFurther - ищет операцию самую старую операцию заданного типа,
// которая находится не в конечном статусе, и блокирует ее для обновления другими транзакциями.
//     $1 - op_type
// Возвращает id, user_id, op_type, status, amount, description, order_number, promo_id операции.
// ВАЖНО: может вызываться только внутри транзакции.
var stmtOperationLockFurther = registerStmt(`
		SELECT id, user_id, op_type, status, amount, description, order_number, promo_id
		FROM operations 
		WHERE status IN ('NEW', 'PROCESSING') AND op_type = $1
		ORDER BY updated_at
		FOR UPDATE SKIP LOCKED 
		LIMIT 1
`)

// stmtOperationUpdate - обновляет status и amount операции.
// ВАЖНО: может вызываться только внутри транзакции и только после вызова Repo.userLockTx.
// После вызова необходимо обновить баланс пользователя при помощи Repo.userUpdateBalanceTx.
var stmtOperationUpdate = registerStmt(`
	UPDATE operations
	SET status = $2, amount = $3, updated_at = now()
	WHERE id = $1
	RETURNING id
`)

// OperationUpdateFurther - берет самую старую операцию заданного типа,
// которая находится не в конечном статусе, вызывает для нее коллбэк updateOp, обновляет операцию
// и обновляет баланс пользователя.
func (r *Repo) OperationUpdateFurther(ctx context.Context, opType models.OperationType, updateOp UpdateFunc) error {

	tx, err := r.db.Begin()
	if err != nil {
		return r.handleError(ctx, err)
	}
	//goland:noinspection ALL
	defer tx.Rollback()

	// Находим операцию для обновления блокируем ее
	op := &models.Operation{}
	err = tx.Stmt(r.stmts[stmtOperationLockFurther]).
		QueryRowContext(ctx, opType).
		Scan(&op.ID, &op.UserID, &op.Type, &op.Status, &op.Amount, &op.Description, &op.OrderNumber, &op.PromoID)
	if err != nil {
		return r.handleError(ctx, err)
	}

	// Вызываем коллбэк для обновления данных операции
	if err = updateOp(ctx, op); err != nil {
		return err
	}

	// Блокируем запись пользователя для обновления
	if err = r.userLockTx(ctx, tx, op.UserID); err != nil {
		return err
	}

	// Обновляем операцию
	err = tx.Stmt(r.stmts[stmtOperationUpdate]).
		QueryRowContext(ctx, op.ID, op.Status, op.Amount).
		Scan(&sql.NullInt64{})
	if err != nil {
		return r.handleError(ctx, err)
	}

	// Обновляем баланс пользователя
	if err = r.userUpdateBalanceTx(ctx, tx, op.UserID); err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		return r.handleError(ctx, err)
	}

	return nil
}

// stmtOperationGetByType - возвращает список операций пользователя заданного типа.
//    $1 - user_id
//    $2 - op_type
// Возвращает id, user_id, op_type, status, amount, description,
// order_number, promo_id, created_at, updated_at операции.
var stmtOperationGetByType = registerStmt(`
	SELECT id, user_id, op_type, status, amount, description, order_number, promo_id, created_at, updated_at
	FROM operations
	WHERE user_id = $1 AND op_type = $2
	ORDER BY created_at DESC
`)

// OperationGetByType - возвращает список операций пользователя заданного типа.
func (r *Repo) OperationGetByType(ctx context.Context, userID int64, t models.OperationType) ([]models.Operation, error) {
	rows, err := r.stmts[stmtOperationGetByType].QueryContext(ctx, userID, t)
	if err != nil {
		return nil, r.handleError(ctx, err)
	}
	//goland:noinspection GoUnhandledErrorResult
	defer rows.Close()

	ops, err := r.operationScanRows(ctx, rows)
	if err != nil {
		return nil, r.handleError(ctx, err)
	}
	return ops, nil
}

// stmtOperationGetBalance - возвращает список операций пользователя, учитывающихся в балансе.
//    $1 - user_id
// Возвращает id, user_id, op_type, status, amount, description,
// order_number, promo_id, created_at, updated_at операции.
var stmtOperationGetBalance = registerStmt(`
	SELECT id, user_id, op_type, status, amount, description, order_number, promo_id, created_at, updated_at
	FROM operations
	WHERE user_id = $1 AND (
	    (status = 'PROCESSED' AND amount >= 0)
	    OR 
	    (status NOT IN ('INVALID', 'CANCELED') AND amount < 0)
	)
	ORDER BY updated_at DESC
`)

// OperationGetBalance - возвращает список операций пользователя, учитывающихся в балансе.
func (r *Repo) OperationGetBalance(ctx context.Context, userID int64) ([]models.Operation, error) {
	rows, err := r.stmts[stmtOperationGetBalance].QueryContext(ctx, userID)
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

func (r *Repo) operationScanRows(ctx context.Context, rows *sql.Rows) ([]models.Operation, error) {
	var ops []models.Operation
	for rows.Next() {
		select {
		case <-ctx.Done():
			return nil, r.handleError(ctx, ctx.Err())
		default:
		}
		op := models.Operation{}
		if err := rows.Scan(
			&op.ID,
			&op.UserID,
			&op.Type,
			&op.Status,
			&op.Amount,
			&op.Description,
			&op.OrderNumber,
			&op.PromoID,
			&op.CreatedAt,
			&op.UpdatedAt,
		); err != nil {
			return nil, err
		}
		ops = append(ops, op)
	}
	if err := rows.Err(); err != nil {
		return nil, r.handleError(ctx, err)
	}
	return ops, nil
}
