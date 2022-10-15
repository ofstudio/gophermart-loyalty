package repo

import (
	"context"
	"database/sql"
	"gophermart-loyalty/internal/models"
)

// stmtUserCreate - создает пользователя.
//    $1 - username
//    $2 - pass_hash
// Возвращает id, balance, withdrawn, created_at, updated_at.
var stmtUserCreate = registerStmt(`
	INSERT INTO users (username, pass_hash) 
	VALUES ($1, $2) 
	RETURNING id, balance, withdrawn, created_at, updated_at
`)

// UserCreate - создает пользователя по логину и хэшу пароля
func (r *Repo) UserCreate(ctx context.Context, u *models.User) error {
	err := r.stmts[stmtUserCreate].
		QueryRowContext(ctx, u.Login, u.PassHash).
		Scan(&u.ID, &u.Balance, &u.Withdrawn, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return r.handleError(ctx, err)
	}
	return nil
}

// stmtUserGetByID - возвращает пользователя по id.
//    $1 - id
// Возвращает id, username, pass_hash, balance, withdrawn, created_at, updated_at.
var stmtUserGetByID = registerStmt(`
	SELECT id, username, pass_hash, balance, withdrawn, created_at, updated_at FROM users
	WHERE id = $1
`)

// UserGetByID - возвращает пользователя по id.
func (r *Repo) UserGetByID(ctx context.Context, userID uint64) (*models.User, error) {
	u := &models.User{}
	err := r.stmts[stmtUserGetByID].
		QueryRowContext(ctx, userID).
		Scan(&u.ID, &u.Login, &u.PassHash, &u.Balance, &u.Withdrawn, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, r.handleError(ctx, err)
	}
	return u, nil
}

// stmtUserGetByLogin - возвращает пользователя по логину.
//    $1 - username
// Возвращает id, username, pass_hash, balance, withdrawn, created_at, updated_at.
var stmtUserGetByLogin = registerStmt(`
	SELECT id, username, pass_hash, balance, withdrawn, created_at, updated_at  FROM users
	WHERE username = $1
`)

// UserGetByLogin - возвращает пользователя по логину.
func (r *Repo) UserGetByLogin(ctx context.Context, login string) (*models.User, error) {
	u := &models.User{}
	err := r.stmts[stmtUserGetByLogin].
		QueryRowContext(ctx, login).
		Scan(&u.ID, &u.Login, &u.PassHash, &u.Balance, &u.Withdrawn, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, r.handleError(ctx, err)
	}
	return u, nil
}

// stmtUserLock - блокирует пользователя для обновления другими транзакциями.
//    $1 - id пользователя
// Возвращает id пользователя.
// ВАЖНО: может вызываться только внутри транзакции.
var stmtUserLock = registerStmt(`
	SELECT id FROM users WHERE id = $1 FOR UPDATE 
`)

// userLockTx - блокирует пользователя для обновления другими транзакциями.
// ВАЖНО: может вызываться только внутри транзакции
func (r *Repo) userLockTx(ctx context.Context, tx *sql.Tx, userID uint64) error {
	if err := tx.Stmt(r.stmts[stmtUserLock]).
		QueryRowContext(ctx, userID).
		Scan(&sql.NullInt64{}); err != nil {
		return r.handleError(ctx, err)
	}
	return nil
}

// stmtUserUpdateBalance - обновляет баланс пользователя
//    $1 - id пользователя
// Возвращает id пользователя.
// ВАЖНО: может вызываться только внутри транзакции и только после вызова Repo.userLockTx
var stmtUserUpdateBalance = registerStmt(`
	WITH
	    total_accrued AS (
	    	SELECT coalesce(sum(amount), 0) AS val FROM operations
			WHERE user_id = $1 AND status = 'PROCESSED' AND amount > 0
		),
	    total_withdrawn AS (
	        SELECT coalesce(sum(amount), 0)  AS val  FROM operations
			WHERE  user_id = $1 AND status NOT IN ('INVALID', 'CANCELED') AND amount < 0
	    )
	UPDATE users
	SET
	    balance = total_accrued.val + total_withdrawn.val,
	    withdrawn = 0 - total_withdrawn.val,
		updated_at = now()
	FROM  total_accrued, total_withdrawn
	WHERE id = $1
	RETURNING id
`)

// userUpdateBalance - обновляет баланс пользователя.
// ВАЖНО: может вызываться только внутри транзакции и только после вызова Repo.userLockTx
func (r *Repo) userUpdateBalanceTx(ctx context.Context, tx *sql.Tx, userID uint64) error {
	err := tx.Stmt(r.stmts[stmtUserUpdateBalance]).
		QueryRowContext(ctx, userID).
		Scan(&sql.NullInt64{})
	if err != nil {
		return r.handleError(ctx, err)
	}
	return nil
}
