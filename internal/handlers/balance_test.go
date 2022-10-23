package handlers

import (
	"net/http"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/mock"

	"gophermart-loyalty/internal/app"
	"gophermart-loyalty/internal/models"
)

func (suite *handlersSuite) TestBalanceGet() {
	suite.Run("success", func() {
		suite.repo.On("UserGetByID", mock.Anything, uint64(1)).
			Return(&models.User{
				ID:        1,
				Login:     "test",
				Balance:   decimal.NewFromFloat(100.34),
				Withdrawn: decimal.NewFromFloat(20.2),
			}, nil).Once()

		token := suite.validJWTToken(1)
		res := suite.httpJSONRequest(http.MethodGet, "/balance", "", token)
		defer res.Body.Close()
		suite.Equal(http.StatusOK, res.StatusCode)
		resJSON := suite.parseJSON(res.Body)
		suite.Equal(100.34, resJSON["current"])
		suite.Equal(20.2, resJSON["withdrawn"])
	})

	suite.Run("non existing user", func() {
		suite.repo.On("UserGetByID", mock.Anything, uint64(100)).
			Return(nil, app.ErrNotFound).Once()

		token := suite.validJWTToken(100)
		res := suite.httpJSONRequest(http.MethodGet, "/balance", "", token)
		defer res.Body.Close()
		suite.Equal(http.StatusInternalServerError, res.StatusCode)
	})

	suite.Run("unauthorized", func() {
		res := suite.httpJSONRequest(http.MethodGet, "/balance", "", "invalid token")
		defer res.Body.Close()
		suite.Equal(http.StatusUnauthorized, res.StatusCode)
	})
}

func (suite *handlersSuite) TestBalanceHistoryGet() {
	data := []*models.Operation{
		{
			ID:          3,
			UserID:      1,
			Type:        models.OrderWithdrawal,
			Status:      models.StatusProcessing,
			Amount:      decimal.NewFromFloat(-100.34),
			Description: "Description 3",
			CreatedAt:   time.Date(2022, 1, 3, 0, 0, 0, 0, time.UTC),
			UpdatedAt:   time.Date(2022, 1, 3, 1, 0, 0, 0, time.UTC),
			OrderNumber: strPtr("12345678903"),
			PromoID:     nil,
		},
		{
			ID:          2,
			UserID:      1,
			Type:        models.OrderAccrual,
			Status:      models.StatusProcessed,
			Amount:      decimal.NewFromFloat(500.),
			Description: "Description 2",
			CreatedAt:   time.Date(2022, 1, 2, 0, 0, 0, 0, time.UTC),
			UpdatedAt:   time.Date(2022, 1, 2, 1, 0, 0, 0, time.UTC),
			OrderNumber: strPtr("9278923470"),
			PromoID:     nil,
		},
		{
			ID:          1,
			UserID:      1,
			Type:        models.PromoAccrual,
			Status:      models.StatusProcessed,
			Amount:      decimal.NewFromFloat(100.),
			Description: "Description 1",
			CreatedAt:   time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC),
			UpdatedAt:   time.Date(2022, 1, 1, 1, 0, 0, 0, time.UTC),
			OrderNumber: nil,
			PromoID:     uint64Ptr(10),
		},
	}

	suite.Run("success", func() {
		suite.repo.On("BalanceHistoryGetByID", mock.Anything, uint64(1)).
			Return(data, nil).Once()

		token := suite.validJWTToken(1)
		res := suite.httpJSONRequest(http.MethodGet, "/balance/history", "", token)
		defer res.Body.Close()
		suite.Equal(http.StatusOK, res.StatusCode)
		resJSON := suite.parseJSONList(res.Body)
		suite.Equal(3, len(resJSON))
		suite.Equal(-100.34, resJSON[0]["amount"])
		suite.Equal("Description 3", resJSON[0]["description"])
		suite.Equal("2022-01-03T01:00:00Z", resJSON[0]["processed_at"])
		_, ok := resJSON[2]["number"]
		suite.False(ok)
	})

	suite.Run("empty", func() {
		suite.repo.On("BalanceHistoryGetByID", mock.Anything, uint64(1)).
			Return(nil, app.ErrNotFound).Once()

		token := suite.validJWTToken(1)
		res := suite.httpJSONRequest(http.MethodGet, "/balance/history", "", token)
		defer res.Body.Close()
		suite.Equal(http.StatusNoContent, res.StatusCode)
	})

	suite.Run("unauthorized", func() {
		res := suite.httpJSONRequest(http.MethodGet, "/balance/history", "", "invalid token")
		defer res.Body.Close()
		suite.Equal(http.StatusUnauthorized, res.StatusCode)
	})

	suite.Run("internal error", func() {
		suite.repo.On("BalanceHistoryGetByID", mock.Anything, uint64(1)).
			Return(nil, app.ErrInternal).Once()

		token := suite.validJWTToken(1)
		res := suite.httpJSONRequest(http.MethodGet, "/balance/history", "", token)
		defer res.Body.Close()
		suite.Equal(http.StatusInternalServerError, res.StatusCode)
	})
}
