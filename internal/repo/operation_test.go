package repo

import (
	"context"
	"errors"
	"fmt"
	"github.com/shopspring/decimal"
	"gophermart-loyalty/internal/app"
	"gophermart-loyalty/internal/models"
	"sync"
	"time"
)

func (suite *repoSuite) TestOperationCreate() {

	suite.Run("OrderAccrual", func() {
		suite.NoError(suite.repo.OperationCreate(suite.ctx(), testOA(1, "01", 100, models.StatusNew)))
		suite.NoError(suite.repo.OperationCreate(suite.ctx(), testOA(1, "02", 100, models.StatusProcessing)))
		suite.NoError(suite.repo.OperationCreate(suite.ctx(), testOA(1, "03", 100, models.StatusCanceled)))
		suite.NoError(suite.repo.OperationCreate(suite.ctx(), testOA(1, "04", 100, models.StatusInvalid)))
		suite.NoError(suite.repo.OperationCreate(suite.ctx(), testOA(1, "05", 100, models.StatusProcessed)))
	})

	suite.Run("OrderWithdrawal", func() {
		suite.NoError(suite.repo.OperationCreate(suite.ctx(), testOW(1, "06", -10, models.StatusNew)))
		suite.NoError(suite.repo.OperationCreate(suite.ctx(), testOW(1, "07", -10, models.StatusProcessing)))
		suite.NoError(suite.repo.OperationCreate(suite.ctx(), testOW(1, "08", -10, models.StatusCanceled)))
		suite.NoError(suite.repo.OperationCreate(suite.ctx(), testOW(1, "09", -10, models.StatusInvalid)))
		suite.NoError(suite.repo.OperationCreate(suite.ctx(), testOW(1, "05", -10, models.StatusProcessed)))
	})

	suite.Run("PromoAccrual", func() {
		suite.NoError(suite.repo.OperationCreate(suite.ctx(), testPA(1, 1, 100, models.StatusProcessed)))
	})

	suite.Run("Balance check", func() {
		u, err := suite.repo.UserGetByID(suite.ctx(), 1)
		suite.NoError(err)
		suite.Equal("170", u.Balance.String())
		suite.Equal("30", u.Withdrawn.String())
	})

}

func (suite *repoSuite) TestOperationCreate_constraints() {

	suite.Run("must_refs_user constraint", func() {
		err := suite.repo.OperationCreate(suite.ctx(), testOA(1000, "10", 100, models.StatusNew))
		suite.ErrorIs(err, app.ErrOperationUserNotExists)
	})

	suite.Run("amount_valid_sign constraint", func() {
		err := suite.repo.OperationCreate(suite.ctx(), testOA(1, "10", -100, models.StatusProcessed))
		suite.ErrorIs(err, app.ErrOperationAmountInvalid)

		err = suite.repo.OperationCreate(suite.ctx(), testOW(1, "20", 100, models.StatusProcessed))
		suite.ErrorIs(err, app.ErrOperationAmountInvalid)

		err = suite.repo.OperationCreate(suite.ctx(), testPA(1, 1, -100, models.StatusProcessed))
		suite.ErrorIs(err, app.ErrOperationAmountInvalid)
	})

	suite.Run("order_belongs_to_user constraint", func() {
		err := suite.repo.OperationCreate(suite.ctx(), testOA(1, "10", 100, models.StatusNew))
		suite.NoError(err)

		err = suite.repo.OperationCreate(suite.ctx(), testOA(2, "10", 100, models.StatusNew))
		suite.ErrorIs(err, app.ErrOperationOrderNotBelongs)
	})

	suite.Run("order_unique_for_op_type constraint", func() {
		err := suite.repo.OperationCreate(suite.ctx(), testOA(1, "100", 100, models.StatusNew))
		suite.NoError(err)
		err = suite.repo.OperationCreate(suite.ctx(), testOA(1, "100", 100, models.StatusNew))
		suite.ErrorIs(err, app.ErrOperationOrderUsed)
	})

	suite.Run("balance_not_negative constraint", func() {
		err := suite.repo.OperationCreate(suite.ctx(), testOA(3, "60", 100, models.StatusProcessed))
		suite.NoError(err)
		err = suite.repo.OperationCreate(suite.ctx(), testOW(3, "60", -100, models.StatusNew))
		suite.NoError(err)
		err = suite.repo.OperationCreate(suite.ctx(), testOW(3, "70", -150, models.StatusProcessing))
		suite.ErrorIs(err, app.ErrUserBalanceNegative)
	})

	suite.Run("must_refs_promo constraint", func() {
		err := suite.repo.OperationCreate(suite.ctx(), testPA(1, 100500, 100, models.StatusProcessed))
		suite.ErrorIs(err, app.ErrOperationPromoNotExists)
	})

	suite.Run("promo_unique_for_user constraint", func() {
		err := suite.repo.OperationCreate(suite.ctx(), testPA(2, 1, 100, models.StatusProcessed))
		suite.NoError(err)
		err = suite.repo.OperationCreate(suite.ctx(), testPA(2, 1, 100, models.StatusProcessed))
		suite.ErrorIs(err, app.ErrOperationPromoUsed)
	})

	suite.Run("operation_valid_attrs constraint", func() {
		promoID := uint64(1)
		orderNumber := "200"
		op := &models.Operation{
			UserID:      1,
			Type:        models.OrderAccrual,
			Status:      models.StatusNew,
			Amount:      decimal.NewFromInt(100),
			Description: "test",
			OrderNumber: nil,
			PromoID:     nil,
		}

		err := suite.repo.OperationCreate(suite.ctx(), op)
		suite.ErrorIs(err, app.ErrOperationAttrsInvalid)

		op.PromoID = &promoID
		err = suite.repo.OperationCreate(suite.ctx(), op)
		suite.ErrorIs(err, app.ErrOperationAttrsInvalid)

		op.OrderNumber = &orderNumber
		err = suite.repo.OperationCreate(suite.ctx(), op)
		suite.ErrorIs(err, app.ErrOperationAttrsInvalid)

		op.PromoID = nil
		err = suite.repo.OperationCreate(suite.ctx(), op)
		suite.NoError(err)
	})

}

func (suite *repoSuite) TestOperationGetByType() {

	suite.Run("populate user 1", func() {
		suite.NoError(suite.repo.OperationCreate(suite.ctx(), testOA(1, "10", 100, models.StatusProcessed)))
		suite.NoError(suite.repo.OperationCreate(suite.ctx(), testOW(1, "20", -10, models.StatusProcessing)))
		suite.NoError(suite.repo.OperationCreate(suite.ctx(), testOW(1, "30", -10, models.StatusProcessed)))
		suite.NoError(suite.repo.OperationCreate(suite.ctx(), testPA(1, 1, 100, models.StatusProcessed)))
	})

	suite.Run("populate user 2", func() {
		suite.NoError(suite.repo.OperationCreate(suite.ctx(), testOA(2, "40", 100, models.StatusCanceled)))
	})

	suite.Run("get OrderAccrual for user 1", func() {
		ops, err := suite.repo.OperationGetByType(suite.ctx(), 1, models.OrderAccrual)
		suite.NoError(err)
		suite.Len(ops, 1)
		suite.Equal("10", *ops[0].OrderNumber)
	})

	suite.Run("get OrderWithdrawal for user 1", func() {
		ops, err := suite.repo.OperationGetByType(suite.ctx(), 1, models.OrderWithdrawal)
		suite.NoError(err)
		suite.Len(ops, 2)
		suite.Equal("30", *ops[0].OrderNumber)
		suite.Equal("20", *ops[1].OrderNumber)
	})

	suite.Run("get PromoAccrual for user 1", func() {
		ops, err := suite.repo.OperationGetByType(suite.ctx(), 1, models.PromoAccrual)
		suite.NoError(err)
		suite.Len(ops, 1)
		suite.Equal(uint64(1), *ops[0].PromoID)
	})

	suite.Run("get OrderAccrual for user 2", func() {
		ops, err := suite.repo.OperationGetByType(suite.ctx(), 2, models.OrderAccrual)
		suite.NoError(err)
		suite.Equal("40", *ops[0].OrderNumber)
	})

	suite.Run("get OrderWithdrawal for user 2", func() {
		ops, err := suite.repo.OperationGetByType(suite.ctx(), 2, models.OrderWithdrawal)
		suite.NoError(err)
		suite.Len(ops, 0)
	})

	suite.Run("get OrderAccrual for user 3", func() {
		ops, err := suite.repo.OperationGetByType(suite.ctx(), 3, models.OrderAccrual)
		suite.NoError(err)
		suite.Len(ops, 0)
	})

}

func (suite *repoSuite) TestOperationGetBalance() {

	suite.Run("populate user 1", func() {
		suite.NoError(suite.repo.OperationCreate(suite.ctx(), testOA(1, "10", 100, models.StatusNew)))
		suite.NoError(suite.repo.OperationCreate(suite.ctx(), testOA(1, "20", 100, models.StatusProcessing)))
		suite.NoError(suite.repo.OperationCreate(suite.ctx(), testOA(1, "30", 100, models.StatusProcessed)))
		suite.NoError(suite.repo.OperationCreate(suite.ctx(), testOA(1, "40", 100, models.StatusInvalid)))
		suite.NoError(suite.repo.OperationCreate(suite.ctx(), testOA(1, "50", 100, models.StatusCanceled)))

		suite.NoError(suite.repo.OperationCreate(suite.ctx(), testPA(1, 1, 100, models.StatusProcessed)))

		suite.NoError(suite.repo.OperationCreate(suite.ctx(), testOW(1, "10", -10, models.StatusNew)))
		suite.NoError(suite.repo.OperationCreate(suite.ctx(), testOW(1, "20", -10, models.StatusProcessing)))
		suite.NoError(suite.repo.OperationCreate(suite.ctx(), testOW(1, "30", -10, models.StatusProcessed)))
		suite.NoError(suite.repo.OperationCreate(suite.ctx(), testOW(1, "40", -10, models.StatusInvalid)))
		suite.NoError(suite.repo.OperationCreate(suite.ctx(), testOW(1, "50", -10, models.StatusCanceled)))
	})

	suite.Run("populate user 2", func() {
		suite.NoError(suite.repo.OperationCreate(suite.ctx(), testOA(2, "60", 100, models.StatusNew)))
		suite.NoError(suite.repo.OperationCreate(suite.ctx(), testOA(2, "70", 100, models.StatusProcessing)))
		suite.NoError(suite.repo.OperationCreate(suite.ctx(), testOA(2, "80", 100, models.StatusInvalid)))
		suite.NoError(suite.repo.OperationCreate(suite.ctx(), testOA(2, "90", 100, models.StatusCanceled)))
	})

	suite.Run("get balance operations for user 1", func() {
		ops, err := suite.repo.OperationGetBalance(suite.ctx(), 1)
		suite.NoError(err)
		suite.Len(ops, 5)
		suite.Equal("30", *ops[0].OrderNumber)
		suite.Equal("20", *ops[1].OrderNumber)
		suite.Equal("10", *ops[2].OrderNumber)
		suite.Equal(uint64(1), *ops[3].PromoID)
		suite.Equal("30", *ops[4].OrderNumber)
	})

	suite.Run("get balance operations for user 2", func() {
		ops, err := suite.repo.OperationGetBalance(suite.ctx(), 2)
		suite.NoError(err)
		suite.Len(ops, 0)
	})

	suite.Run("get balance operations for user 3", func() {
		ops, err := suite.repo.OperationGetBalance(suite.ctx(), 3)
		suite.NoError(err)
		suite.Len(ops, 0)
	})

}

func (suite *repoSuite) TestOperationUpdateFurther() {

	suite.Run("populate operations", func() {
		// Создаем операции для 3 пользователей
		for i := 0; i < 300; i++ {
			uid := uint64(i%3 + 1)
			num := fmt.Sprintf("%06d", i)
			suite.NoError(suite.repo.OperationCreate(suite.ctx(), testOA(uid, num, 1, models.StatusNew)))
		}
	})

	suite.Run("update operations", func() {
		// Запускаем воркеры для обновления операций
		wg := &sync.WaitGroup{}
		ctx, cancel := context.WithTimeout(suite.ctx(), 5*time.Second)
		defer cancel()
		for i := 0; i < 8; i++ {
			wg.Add(1)
			go suite.updateWorker(ctx, wg, i)
		}
		wg.Wait()

		// Проверяем балансы пользователей после обновления операций
		u, err := suite.repo.UserGetByID(suite.ctx(), 1)
		suite.NoError(err)
		suite.Equal("0", u.Balance.String())
		u, err = suite.repo.UserGetByID(suite.ctx(), 2)
		suite.NoError(err)
		suite.Equal("0", u.Balance.String())
		u, err = suite.repo.UserGetByID(suite.ctx(), 3)
		suite.NoError(err)
		suite.Equal("100", u.Balance.String())
	})
}

// updateWorker - воркер, который обновляет операции в очереди на обновление.
func (suite *repoSuite) updateWorker(ctx context.Context, wg *sync.WaitGroup, pid int) {
	defer wg.Done()
	for {
		select {
		case <-ctx.Done():
			suite.T().Errorf("worker timeout: %d", pid)
			return
		default:
			err := suite.repo.OperationUpdateFurther(ctx, models.OrderAccrual, suite.updateFunc)
			if errors.Is(err, app.ErrNotFound) {
				return
			}
			suite.NoError(err)
		}
	}
}

// UpdateFunc - коллбэк обновления операции.
// В результате работы все статусы NEW и PROCESSING обновятся
//    user_id = 1 => CANCELED
//    user_id = 2 => INVALID
//    user_id = 3 => PROCESSED
func (suite *repoSuite) updateFunc(_ context.Context, op *models.Operation) error {
	if op.Status == models.StatusNew {
		op.Status = models.StatusProcessing
		return nil
	}
	if op.Status == models.StatusProcessing {
		switch int(op.UserID % 3) {
		case 0:
			op.Status = models.StatusProcessed
		case 1:
			op.Status = models.StatusCanceled
		case 2:
			op.Status = models.StatusInvalid
		}
	}
	return nil
}
