package usecases

import (
	"context"
	"errors"
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/shopspring/decimal"
	"gophermart-loyalty/internal/app"
	"gophermart-loyalty/internal/models"
	"gophermart-loyalty/internal/repo"
	"gophermart-loyalty/pkg/luhn"
	"time"
)

// OrderAccrualCreate - создание начисления по заказу.
func (u *UseCases) OrderAccrualCreate(ctx context.Context, userID uint64, orderNumber string) (*models.Operation, error) {
	// валидируем номер заказа
	if err := u.orderNumberValidate(ctx, orderNumber); err != nil {
		return nil, err
	}
	// создаем операцию
	op := &models.Operation{
		UserID:      userID,
		Type:        models.OrderAccrual,
		OrderNumber: &orderNumber,
		Status:      models.StatusNew,
		Description: fmt.Sprintf("Начисление баллов за заказ %s", orderNumber),
	}
	// сохраняем в репозиторий
	if err := u.operationCreate(ctx, op); err != nil {
		return nil, err
	}
	return op, nil
}

func (u *UseCases) OrderWithdrawalCreate(ctx context.Context, userID uint64, orderNumber string, amount decimal.Decimal) (*models.Operation, error) {
	// валидируем номер заказа
	if err := u.orderNumberValidate(ctx, orderNumber); err != nil {
		return nil, err
	}
	// создаем операцию
	op := &models.Operation{
		UserID:      userID,
		Type:        models.OrderWithdrawal,
		OrderNumber: &orderNumber,
		Status:      models.StatusNew,
		Amount:      amount,
		Description: fmt.Sprintf("Списание баллов за заказ %s", orderNumber),
	}
	// сохраняем в репозиторий
	if err := u.operationCreate(ctx, op); err != nil {
		return nil, err
	}
	return op, nil
}

func (u *UseCases) PromoAccrualCreate(ctx context.Context, userID uint64, promoCode string) (*models.Operation, error) {
	// получаем промокод
	promo, err := u.promoCodeValidate(promoCode)
	if err != nil {
		return nil, err
	}
	// создаем операцию
	op := &models.Operation{
		UserID:      userID,
		Type:        models.PromoAccrual,
		PromoID:     &promo.ID,
		Amount:      promo.Reward,
		Status:      models.StatusProcessed,
		Description: fmt.Sprintf("Начисление баллов по промо-коду %s", promoCode),
	}
	// сохраняем в репозиторий
	if err = u.repo.OperationCreate(ctx, op); err != nil {
		return nil, err
	}
	return op, nil
}

func (u *UseCases) operationCreate(ctx context.Context, op *models.Operation) error {
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

func (u *UseCases) OperationUpdateFurther(ctx context.Context, opType models.OperationType, updateFunc repo.UpdateFunc) error {
	// обновляем операцию
	var operationID uint64
	err := u.repo.OperationUpdateFurther(ctx, opType, func(ctx context.Context, operation *models.Operation) error {
		operationID = operation.ID
		return updateFunc(ctx, operation)
	})

	if errors.Is(err, app.ErrNotFound) {
		return nil
	} else if err != nil {
		u.log.WithReqID(ctx).Error().Err(err).Uint64("operation_id", operationID).Msg("failed to update operation")
		return err
	}

	log.Info().Uint64("operation_id", operationID).Msg("operation updated")
	return nil
}

const orderNumberMaxLen = 512

func (u *UseCases) orderNumberValidate(ctx context.Context, orderNumber string) error {
	if len(orderNumber) > orderNumberMaxLen {
		u.log.WithReqID(ctx).Error().Msg("order number is too long")
		return app.ErrOperationOrderNumberInvalid
	}
	if !luhn.Check(orderNumber) {
		u.log.WithReqID(ctx).Error().Msg("order number is invalid")
		return app.ErrOperationOrderNumberInvalid
	}
	return nil
}

func (u *UseCases) promoCodeValidate(promoCode string) (*models.Promo, error) {
	promo, err := u.repo.PromoGetByCode(context.Background(), promoCode)
	if err != nil {
		u.log.Error().Err(err).Msg("failed to get promo")
		return nil, err
	}
	// проверяем, что промокод активен в данный момент
	now := time.Now()
	if now.Before(promo.NotBefore) || now.After(promo.NotAfter) {
		u.log.Error().Err(err).Msg("promo is not active")
		return nil, app.ErrNotFound
	}
	return promo, nil
}
