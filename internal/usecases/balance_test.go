package usecases

import (
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/mock"

	"gophermart-loyalty/internal/errs"
	"gophermart-loyalty/internal/models"
)

func (suite *useCasesSuite) TestBalanceHistoryGetByID() {
	suite.Run("success", func() {
		ops := []*models.Operation{
			{
				ID:          1,
				UserID:      1,
				Type:        models.OrderAccrual,
				Status:      models.StatusNew,
				Amount:      decimal.NewFromInt(100),
				Description: "Начисление баллов за заказ 1",
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
		}
		suite.repo.On("BalanceHistoryGetByID", mock.Anything, uint64(1)).
			Return(ops, nil).Once()
		history, err := suite.useCases.BalanceHistoryGetByID(suite.ctx(), uint64(1))
		suite.NoError(err)
		suite.Equal(ops, history)
	})

	suite.Run("no operations", func() {
		suite.repo.On("BalanceHistoryGetByID", mock.Anything, uint64(1)).
			Return(nil, errs.ErrNotFound).Once()
		history, err := suite.useCases.BalanceHistoryGetByID(suite.ctx(), uint64(1))
		suite.NoError(err)
		suite.Nil(history)
	})

	suite.Run("internal error", func() {
		suite.repo.On("BalanceHistoryGetByID", mock.Anything, uint64(1)).
			Return(nil, errs.ErrInternal).Once()
		history, err := suite.useCases.BalanceHistoryGetByID(suite.ctx(), uint64(1))
		suite.ErrorIs(err, errs.ErrInternal)
		suite.Nil(history)
	})

}
