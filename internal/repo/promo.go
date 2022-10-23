package repo

import (
	"context"

	"gophermart-loyalty/internal/models"
)

// stmtPromoCreate - создает промо-кампанию.
//    $1 - code
//    $2 - description
//    $3 - reward
//    $4 - not_before
//	  $5 - not_after
// Возвращает id новой промо-кампании.
var stmtPromoCreate = registerStmt(`
	INSERT INTO promos (code, description, reward, not_before, not_after)
	VALUES ($1, $2, $3, $4, $5)
	RETURNING id
`)

// PromoCreate - создает промо-кампанию.
func (r *PGXRepo) PromoCreate(ctx context.Context, p *models.Promo) error {
	err := r.stmts[stmtPromoCreate].
		QueryRowContext(ctx, &p.Code, &p.Description, &p.Reward, &p.NotBefore, &p.NotAfter).
		Scan(&p.ID)
	if err != nil {
		return r.handleError(ctx, err)
	}
	return nil
}

// stmtPromoGetByCode - возвращает промо-кампанию по коду.
//    $1 - code
// Возвращает id, code, description, reward, valid_until, created_at.
var stmtPromoGetByCode = registerStmt(`
	SELECT id, code, description, reward, not_before, not_after, created_at 
	FROM promos
	WHERE code = $1
`)

// PromoGetByCode - возвращает промо-кампанию по ее промо-коду.
func (r *PGXRepo) PromoGetByCode(ctx context.Context, code string) (*models.Promo, error) {
	p := &models.Promo{}
	err := r.stmts[stmtPromoGetByCode].
		QueryRowContext(ctx, code).
		Scan(&p.ID, &p.Code, &p.Description, &p.Reward, &p.NotBefore, &p.NotAfter, &p.CreatedAt)
	if err != nil {
		return nil, r.handleError(ctx, err)
	}
	return p, nil
}
