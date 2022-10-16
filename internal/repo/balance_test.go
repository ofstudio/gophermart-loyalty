package repo

import "gophermart-loyalty/internal/models"

func (suite *pgxRepoSuite) TestBalanceHistoryGetByID() {

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
		ops, err := suite.repo.BalanceHistoryGetByID(suite.ctx(), 1)
		suite.NoError(err)
		suite.Len(ops, 5)
		suite.Equal("30", *ops[0].OrderNumber)
		suite.Equal("20", *ops[1].OrderNumber)
		suite.Equal("10", *ops[2].OrderNumber)
		suite.Equal(uint64(1), *ops[3].PromoID)
		suite.Equal("30", *ops[4].OrderNumber)
	})

	suite.Run("get balance operations for user 2", func() {
		ops, err := suite.repo.BalanceHistoryGetByID(suite.ctx(), 2)
		suite.NoError(err)
		suite.Len(ops, 0)
	})

	suite.Run("get balance operations for user 3", func() {
		ops, err := suite.repo.BalanceHistoryGetByID(suite.ctx(), 3)
		suite.NoError(err)
		suite.Len(ops, 0)
	})

}
