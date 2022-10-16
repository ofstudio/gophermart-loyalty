package usecases

import (
	"context"
	"errors"
	"gophermart-loyalty/internal/app"
	"gophermart-loyalty/internal/models"
)

func (u *UseCases) BalanceHistoryGetByID(ctx context.Context, userID uint64) ([]*models.Operation, error) {
	list, err := u.repo.BalanceHistoryGetByID(ctx, userID)
	if errors.Is(err, app.ErrNotFound) {
		return nil, nil
	} else if err != nil {
		u.log.WithReqID(ctx).Error().Err(err).Msg("failed to get balance history")
		return nil, err
	}
	return list, nil
}
