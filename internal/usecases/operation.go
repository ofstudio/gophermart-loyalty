package usecases

import (
	"context"
	"errors"
	"fmt"
	"github.com/shopspring/decimal"
	"gophermart-loyalty/internal/app"
	"gophermart-loyalty/internal/models"
	"gophermart-loyalty/internal/repo"
	"gophermart-loyalty/pkg/luhn"
	"time"
)

// OrderAccrualPrepare - создает модель операции начисления по заказу.
func (u *UseCases) OrderAccrualPrepare(ctx context.Context, userID uint64, orderNumber string) (*models.Operation, error) {
	if err := u.orderNumberValidate(orderNumber); err != nil {
		u.log.WithReqID(ctx).Error().Err(err).Msg("invalid order number")
		return nil, err
	}
	return &models.Operation{
		UserID:      userID,
		Type:        models.OrderAccrual,
		OrderNumber: &orderNumber,
		Status:      models.StatusNew,
		Description: fmt.Sprintf("Начисление баллов за заказ %s", orderNumber),
	}, nil
}

// OrderWithdrawalPrepare - создает модель операции списания по заказу.
func (u *UseCases) OrderWithdrawalPrepare(ctx context.Context, userID uint64, orderNumber string, amount decimal.Decimal) (*models.Operation, error) {
	if err := u.orderNumberValidate(orderNumber); err != nil {
		u.log.WithReqID(ctx).Error().Err(err).Msg("invalid order number")
		return nil, err
	}
	return &models.Operation{
		UserID:      userID,
		Type:        models.OrderWithdrawal,
		OrderNumber: &orderNumber,
		Status:      models.StatusNew,
		Amount:      amount,
		Description: fmt.Sprintf("Списание баллов за заказ %s", orderNumber),
	}, nil
}

// PromoAccrualPrepare - создает модель операции начисления по промо-коду.
func (u *UseCases) PromoAccrualPrepare(ctx context.Context, userID uint64, promoCode string) (*models.Operation, error) {
	// получаем промокод
	promo, err := u.repo.PromoGetByCode(context.Background(), promoCode)
	if err != nil {
		u.log.WithReqID(ctx).Error().Err(err).Msg("failed to get promo")
		return nil, err
	}
	// проверяем, что промокод активен в данный момент
	now := time.Now()
	if now.Before(promo.NotBefore) || now.After(promo.NotAfter) {
		u.log.WithReqID(ctx).Error().Err(err).Msg("promo is not active")
		return nil, app.ErrNotFound
	}

	return &models.Operation{
		UserID:      userID,
		Type:        models.PromoAccrual,
		PromoID:     &promo.ID,
		Amount:      promo.Reward,
		Status:      models.StatusProcessed,
		Description: fmt.Sprintf("Начисление баллов по промо-коду %s", promoCode),
	}, nil
}

// OperationCreate - создает операцию в репозитории.
func (u *UseCases) OperationCreate(ctx context.Context, op *models.Operation) error {
	if err := u.repo.OperationCreate(ctx, op); err != nil {
		u.log.WithReqID(ctx).Error().Err(err).Msg("failed to create operation")
		return err
	}
	u.log.WithReqID(ctx).Info().Uint64("operation_id", op.ID).Msg("operation created")
	return nil
}

// OperationGetByType - список операций пользователя по заданному типу
func (u *UseCases) OperationGetByType(ctx context.Context, userID uint64, t models.OperationType) ([]*models.Operation, error) {
	ops, err := u.repo.OperationGetByType(ctx, userID, t)
	if errors.Is(err, app.ErrNotFound) {
		return []*models.Operation{}, nil
	} else if err != nil {
		u.log.WithReqID(ctx).Error().Err(err).Msg("failed to get operations")
		return nil, err
	}
	return ops, nil
}

// OperationUpdateFurther - вызывает Repo.OperationUpdateFurther.
func (u *UseCases) OperationUpdateFurther(ctx context.Context, opType models.OperationType, updateFunc repo.UpdateFunc) (*models.Operation, error) {
	return u.repo.OperationUpdateFurther(ctx, opType, updateFunc)
}

// orderNumberValidate - валидирует номер заказа.
func (u *UseCases) orderNumberValidate(orderNumber string) error {
	const orderNumberMaxLen = 512
	if len(orderNumber) == 0 {
		return app.ErrBadRequest
	}
	if len(orderNumber) > orderNumberMaxLen || !luhn.Check(orderNumber) {
		return app.ErrOperationOrderNumberInvalid
	}
	return nil
}
