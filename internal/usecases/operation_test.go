package usecases

import (
	"context"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/mock"

	"gophermart-loyalty/internal/errs"
	"gophermart-loyalty/internal/models"
	"gophermart-loyalty/internal/repo"
)

func (suite *useCasesSuite) TestOrderAccrualPrepare() {
	suite.Run("success", func() {
		op, err := suite.useCases.OrderAccrualPrepare(suite.ctx(), 1, "2377225624")
		suite.NoError(err)
		suite.Equal(uint64(1), op.UserID)
		suite.Equal(models.OrderAccrual, op.Type)
		suite.NotNil(op.OrderNumber)
		suite.Equal("2377225624", *op.OrderNumber)
		suite.Contains(op.Description, "2377225624")
	})

	suite.Run("invalid order number", func() {
		op, err := suite.useCases.OrderAccrualPrepare(suite.ctx(), 1, "111")
		suite.ErrorIs(err, errs.ErrOperationOrderNumberInvalid)
		suite.Nil(op)
	})
}

func (suite *useCasesSuite) TestOrderWithdrawalPrepare() {
	suite.Run("success", func() {
		op, err := suite.useCases.OrderWithdrawalPrepare(suite.ctx(), 1, "2377225624", decimal.NewFromFloat(100))
		suite.NoError(err)
		suite.Equal(uint64(1), op.UserID)
		suite.Equal(models.OrderWithdrawal, op.Type)
		suite.NotNil(op.OrderNumber)
		suite.Equal("2377225624", *op.OrderNumber)
		suite.Contains(op.Description, "2377225624")
		suite.Equal(decimal.NewFromFloat(100), op.Amount)
	})

	suite.Run("invalid order number", func() {
		op, err := suite.useCases.OrderWithdrawalPrepare(suite.ctx(), 1, "111", decimal.NewFromFloat(100))
		suite.ErrorIs(err, errs.ErrOperationOrderNumberInvalid)
		suite.Nil(op)
	})
}

func (suite *useCasesSuite) TestPromoAccrualPrepare() {
	suite.Run("success", func() {
		suite.repo.On("PromoGetByCode", mock.Anything, "PROMO1").Return(&models.Promo{
			ID:        10,
			Code:      "PROMO1",
			Reward:    decimal.NewFromFloat(100),
			NotBefore: time.Now().Add(-time.Hour),
			NotAfter:  time.Now().Add(time.Hour),
		}, nil).Once()

		op, err := suite.useCases.PromoAccrualPrepare(suite.ctx(), 1, "PROMO1")
		suite.NoError(err)
		suite.Equal(uint64(1), op.UserID)
		suite.Equal(models.PromoAccrual, op.Type)
		suite.NotNil(op.PromoID)
		suite.Equal(uint64(10), *op.PromoID)
		suite.Contains(op.Description, "PROMO1")
		suite.Equal(decimal.NewFromFloat(100), op.Amount)
	})

	suite.Run("promo not found", func() {
		suite.repo.On("PromoGetByCode", mock.Anything, "PROMO1").
			Return(nil, errs.ErrNotFound).Once()
		op, err := suite.useCases.PromoAccrualPrepare(suite.ctx(), 1, "PROMO1")
		suite.ErrorIs(err, errs.ErrNotFound)
		suite.Nil(op)
	})

	suite.Run("promo expired", func() {
		suite.repo.On("PromoGetByCode", mock.Anything, "PROMO1").Return(&models.Promo{
			ID:        10,
			Code:      "PROMO1",
			Reward:    decimal.NewFromFloat(100),
			NotBefore: time.Now().Add(-time.Hour * 2),
			NotAfter:  time.Now().Add(-time.Hour),
		}, nil).Once()

		op, err := suite.useCases.PromoAccrualPrepare(suite.ctx(), 1, "PROMO1")
		suite.ErrorIs(err, errs.ErrNotFound)
		suite.Nil(op)
	})

	suite.Run("promo not active yet", func() {
		suite.repo.On("PromoGetByCode", mock.Anything, "PROMO1").Return(&models.Promo{
			ID:        10,
			Code:      "PROMO1",
			Reward:    decimal.NewFromFloat(100),
			NotBefore: time.Now().Add(time.Hour),
			NotAfter:  time.Now().Add(time.Hour * 2),
		}, nil).Once()

		op, err := suite.useCases.PromoAccrualPrepare(suite.ctx(), 1, "PROMO1")
		suite.ErrorIs(err, errs.ErrNotFound)
		suite.Nil(op)
	})
}

func (suite *useCasesSuite) TestOperationCreate() {
	suite.Run("success", func() {
		op := &models.Operation{
			UserID:      1,
			Type:        models.OrderAccrual,
			Status:      models.StatusNew,
			Amount:      decimal.NewFromFloat(100),
			Description: "Test",
			OrderNumber: strPtr("2377225624"),
			PromoID:     nil,
		}
		suite.repo.On("OperationCreate", mock.Anything, op).
			Return(nil).Once()

		err := suite.useCases.OperationCreate(suite.ctx(), op)
		suite.NoError(err)
	})

	suite.Run("invalid operation", func() {
		op := &models.Operation{
			UserID:      1,
			Type:        models.OrderWithdrawal,
			Status:      models.StatusProcessed,
			Amount:      decimal.NewFromFloat(-200),
			Description: "Test",
			OrderNumber: strPtr("2377225624"),
			PromoID:     nil,
		}
		suite.repo.On("OperationCreate", mock.Anything, op).
			Return(errs.ErrUserBalanceNegative).Once()

		err := suite.useCases.OperationCreate(suite.ctx(), op)
		suite.ErrorIs(err, errs.ErrUserBalanceNegative)
	})
}

func (suite *useCasesSuite) TestOperationGetByType() {
	suite.Run("success", func() {
		suite.repo.On("OperationGetByType", mock.Anything, uint64(1), models.OrderAccrual).
			Return([]*models.Operation{
				{
					ID:          1,
					UserID:      1,
					Type:        models.OrderAccrual,
					Status:      models.StatusNew,
					Amount:      decimal.NewFromFloat(100),
					Description: "Test",
					OrderNumber: strPtr("2377225624"),
					PromoID:     nil,
				},
				{
					ID:          2,
					UserID:      1,
					Type:        models.OrderAccrual,
					Status:      models.StatusNew,
					Amount:      decimal.NewFromFloat(100),
					Description: "Test",
					OrderNumber: strPtr("12345678903"),
					PromoID:     nil,
				},
			}, nil).Once()

		ops, err := suite.useCases.OperationGetByType(suite.ctx(), 1, models.OrderAccrual)
		suite.NoError(err)
		suite.Len(ops, 2)
	})

	suite.Run("not found", func() {
		suite.repo.On("OperationGetByType", mock.Anything, uint64(1), models.OrderAccrual).
			Return(nil, errs.ErrNotFound).Once()

		ops, err := suite.useCases.OperationGetByType(suite.ctx(), 1, models.OrderAccrual)
		suite.Nil(err)
		suite.Equal(0, len(ops))
	})
}

func (suite *useCasesSuite) TestOperationUpdateFurther() {
	suite.Run("success", func() {
		op := &models.Operation{
			ID:          1,
			UserID:      1,
			Type:        models.OrderAccrual,
			Status:      models.StatusNew,
			Amount:      decimal.NewFromFloat(100),
			Description: "Test",
			OrderNumber: strPtr("2377225624"),
			PromoID:     nil,
		}

		var updateFunc repo.UpdateFunc = func(ctx context.Context, op *models.Operation) error {
			suite.Equal(uint64(1), op.ID)
			suite.Equal(models.StatusNew, op.Status)
			op.Status = models.StatusProcessed
			return nil
		}

		suite.repo.On("OperationUpdateFurther", mock.Anything, models.OrderAccrual, mock.AnythingOfType("repo.UpdateFunc")).
			Return(&models.Operation{}, nil).Once().
			Run(func(args mock.Arguments) {
				updateFunc(suite.ctx(), op)
			})

		_, err := suite.useCases.OperationUpdateFurther(suite.ctx(), models.OrderAccrual, updateFunc)
		suite.NoError(err)
		suite.Equal(models.StatusProcessed, op.Status)
	})

	suite.Run("internal error", func() {
		suite.repo.On("OperationUpdateFurther", mock.Anything, models.OrderAccrual, mock.AnythingOfType("repo.UpdateFunc")).
			Return(nil, errs.ErrInternal).Once()

		_, err := suite.useCases.OperationUpdateFurther(suite.ctx(), models.OrderAccrual, nil)
		suite.ErrorIs(err, errs.ErrInternal)
	})
}
