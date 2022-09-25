package repo

import (
	"context"
	"database/sql"
	"gophermart-loyalty/internal/errs"
	"gophermart-loyalty/internal/models"
	"gophermart-loyalty/internal/repo/constraint"
)

// stmtPromoCreate - создает промо-кампанию.
//    $1 - code
//    $2 - description
//    $3 - reward
//    $4 - valid_until
// Возвращает id новой промо-кампании.
var stmtPromoCreate = registerStmt(`
	INSERT INTO promos (code, description, reward, valid_until)
	VALUES ($1, $2, $3, $4)
	RETURNING id
`)

// PromoCreate - создает промо-кампанию.
func (r *Repo) PromoCreate(ctx context.Context, p *models.Promo) error {
	log := r.log.WithRequestID(ctx).With().Str("code", p.Code).Logger()
	err := r.stmts[stmtPromoCreate].
		QueryRowContext(ctx, p.Code, p.Description, p.Reward, p.ValidUntil).
		Scan(&p.ID)
	if err != nil {
		if c, ok := constraint.Violated(err); ok {
			return r.constraintErr(ctx, c)
		}
		log.Error().Err(err).Msg("failed to create promo")
		return errs.Internal
	}

	log.Debug().Msg("promo created")
	return nil
}

// stmtPromoGetByCode - возвращает промо-кампанию по коду.
//    $1 - code
// Возвращает id, code, description, reward, valid_until, created_at.
var stmtPromoGetByCode = registerStmt(`
	SELECT id, code, description, reward, valid_until, created_at 
	FROM promos
	WHERE code = $1
`)

// PromoGetByCode - возвращает промо-кампанию по id.
func (r *Repo) PromoGetByCode(ctx context.Context, code string) (*models.Promo, error) {
	log := r.log.WithRequestID(ctx).With().Str("code", code).Logger()

	p := &models.Promo{}
	err := r.stmts[stmtPromoGetByCode].
		QueryRowContext(ctx, code).
		Scan(&p.ID, &p.Code, &p.Description, &p.Reward, &p.ValidUntil, &p.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, errs.NotFound
	} else if err != nil {
		log.Error().Err(err).Msg("failed to get promo")
		return nil, errs.Internal
	}
	log.Debug().Msg("promo retrieved")
	return p, nil
}
