package handlers

import (
	"net/http"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/mock"

	"gophermart-loyalty/internal/errs"
	"gophermart-loyalty/internal/models"
)

func (suite *handlersSuite) TestOrderAccrualCreate() {
	suite.Run("success", func() {
		suite.repo.On("OperationCreate", mock.Anything, mock.Anything).
			Return(nil).Once()
		token := suite.validJWTToken(1)
		res := suite.httpPlainTextRequest("POST", "/orders", "12345678903", token)
		defer res.Body.Close()
		suite.Equal(http.StatusAccepted, res.StatusCode)
	})

	suite.Run("invalid order number", func() {
		token := suite.validJWTToken(1)
		res := suite.httpPlainTextRequest("POST", "/orders", "invalid", token)
		defer res.Body.Close()
		suite.Equal(http.StatusUnprocessableEntity, res.StatusCode)
		resJSON := suite.parseJSON(res.Body)
		suite.Equal(1203., resJSON["code"])
	})

	suite.Run("duplicate order id", func() {
		suite.repo.On("OperationCreate", mock.Anything, mock.Anything).
			Return(errs.ErrOperationOrderUsed).Once()
		token := suite.validJWTToken(1)
		res := suite.httpPlainTextRequest("POST", "/orders", "12345678903", token)
		defer res.Body.Close()
		suite.Equal(http.StatusOK, res.StatusCode)
		body := suite.getBody(res.Body)
		suite.Equal(0, len(body))
	})

	suite.Run("no request body ", func() {
		token := suite.validJWTToken(1)
		res := suite.httpPlainTextRequest("POST", "/orders", "", token)
		defer res.Body.Close()
		suite.Equal(http.StatusBadRequest, res.StatusCode)
		resJSON := suite.parseJSON(res.Body)
		suite.Equal(1003., resJSON["code"])
	})

	suite.Run("order number belongs to another user", func() {
		suite.repo.On("OperationCreate", mock.Anything, mock.Anything).
			Return(errs.ErrOperationOrderNotBelongs).Once()
		token := suite.validJWTToken(2)
		res := suite.httpPlainTextRequest("POST", "/orders", "12345678903", token)
		defer res.Body.Close()
		suite.Equal(http.StatusConflict, res.StatusCode)
		resJSON := suite.parseJSON(res.Body)
		suite.Equal(1204., resJSON["code"])
	})

	suite.Run("internal error", func() {
		suite.repo.On("OperationCreate", mock.Anything, mock.Anything).
			Return(errs.ErrInternal).Once()
		token := suite.validJWTToken(1)
		res := suite.httpPlainTextRequest("POST", "/orders", "12345678903", token)
		defer res.Body.Close()
		suite.Equal(http.StatusInternalServerError, res.StatusCode)
		resJSON := suite.parseJSON(res.Body)
		suite.Equal(1000., resJSON["code"])
	})
}

func (suite *handlersSuite) TestOrderWithdrawalCreate() {
	suite.Run("success", func() {
		reqBody := `{"order":"12345678903","sum":100}`
		suite.repo.On("OperationCreate", mock.Anything, mock.Anything).
			Return(nil).Once()
		token := suite.validJWTToken(1)
		res := suite.httpJSONRequest("POST", "/balance/withdraw", reqBody, token)
		defer res.Body.Close()
		suite.Equal(http.StatusOK, res.StatusCode)
		resBody := suite.getBody(res.Body)
		suite.Equal(0, len(resBody))
	})

	suite.Run("invalid order number", func() {
		reqBody := `{"order":"invalid","sum":100}`
		token := suite.validJWTToken(1)
		res := suite.httpJSONRequest("POST", "/balance/withdraw", reqBody, token)
		defer res.Body.Close()
		suite.Equal(http.StatusUnprocessableEntity, res.StatusCode)
		resJSON := suite.parseJSON(res.Body)
		suite.Equal(1203., resJSON["code"])
	})

	suite.Run("bad request", func() {
		reqBody := `{"notjson`
		token := suite.validJWTToken(1)
		res := suite.httpJSONRequest("POST", "/balance/withdraw", reqBody, token)
		defer res.Body.Close()
		suite.Equal(http.StatusBadRequest, res.StatusCode)
		resJSON := suite.parseJSON(res.Body)
		suite.Equal(1003., resJSON["code"])
	})

	suite.Run("insufficient funds", func() {
		reqBody := `{"order":"12345678903","sum":100}`
		suite.repo.On("OperationCreate", mock.Anything, mock.Anything).
			Return(errs.ErrUserBalanceNegative).Once()
		token := suite.validJWTToken(1)
		res := suite.httpJSONRequest("POST", "/balance/withdraw", reqBody, token)
		defer res.Body.Close()
		suite.Equal(http.StatusPaymentRequired, res.StatusCode)
		resJSON := suite.parseJSON(res.Body)
		suite.Equal(1105., resJSON["code"])
	})

	suite.Run("order number belongs to another user", func() {
		reqBody := `{"order":"12345678903","sum":100}`
		suite.repo.On("OperationCreate", mock.Anything, mock.Anything).
			Return(errs.ErrOperationOrderNotBelongs).Once()
		token := suite.validJWTToken(2)
		res := suite.httpJSONRequest("POST", "/balance/withdraw", reqBody, token)
		defer res.Body.Close()
		suite.Equal(http.StatusConflict, res.StatusCode)
		resJSON := suite.parseJSON(res.Body)
		suite.Equal(1204., resJSON["code"])
	})

	suite.Run("internal error", func() {
		reqBody := `{"order":"12345678903","sum":100}`
		suite.repo.On("OperationCreate", mock.Anything, mock.Anything).
			Return(errs.ErrInternal).Once()
		token := suite.validJWTToken(1)
		res := suite.httpJSONRequest("POST", "/balance/withdraw", reqBody, token)
		defer res.Body.Close()
		suite.Equal(http.StatusInternalServerError, res.StatusCode)
		resJSON := suite.parseJSON(res.Body)
		suite.Equal(1000., resJSON["code"])
	})
}

func (suite *handlersSuite) TestPromoAccrualCreate() {
	suite.Run("success", func() {
		suite.repo.On("PromoGetByCode", mock.Anything, "WELCOME2022").
			Return(&models.Promo{
				ID:          1,
				Code:        "WELCOME2022",
				Description: "Test",
				Reward:      decimal.NewFromInt(100),
				NotBefore:   time.Now().Add(-time.Hour),
				NotAfter:    time.Now().Add(time.Hour),
			}, nil).Once()

		suite.repo.On("OperationCreate", mock.Anything, mock.Anything).
			Return(nil).Once()

		token := suite.validJWTToken(1)
		res := suite.httpPlainTextRequest("POST", "/promos", "WELCOME2022", token)
		defer res.Body.Close()
		suite.Equal(http.StatusOK, res.StatusCode)
		resBody := suite.getBody(res.Body)
		suite.Equal(0, len(resBody))
	})

	suite.Run("promo not found", func() {
		suite.repo.On("PromoGetByCode", mock.Anything, "WELCOME2022").
			Return(nil, errs.ErrNotFound).Once()

		token := suite.validJWTToken(1)
		res := suite.httpPlainTextRequest("POST", "/promos", "WELCOME2022", token)
		defer res.Body.Close()
		suite.Equal(http.StatusNotFound, res.StatusCode)
		resJSON := suite.parseJSON(res.Body)
		suite.Equal(1001., resJSON["code"])
	})

	suite.Run("promo expired", func() {
		suite.repo.On("PromoGetByCode", mock.Anything, "WELCOME2022").
			Return(&models.Promo{
				ID:          1,
				Code:        "WELCOME2022",
				Description: "Test",
				Reward:      decimal.NewFromInt(100),
				NotBefore:   time.Now().Add(-time.Hour),
				NotAfter:    time.Now().Add(-time.Hour),
			}, nil).Once()

		token := suite.validJWTToken(1)
		res := suite.httpPlainTextRequest("POST", "/promos", "WELCOME2022", token)
		defer res.Body.Close()
		suite.Equal(http.StatusNotFound, res.StatusCode)
		resJSON := suite.parseJSON(res.Body)
		suite.Equal(1001., resJSON["code"])
	})

	suite.Run("promo not active yet", func() {
		suite.repo.On("PromoGetByCode", mock.Anything, "WELCOME2022").
			Return(&models.Promo{
				ID:          1,
				Code:        "WELCOME2022",
				Description: "Test",
				Reward:      decimal.NewFromInt(100),
				NotBefore:   time.Now().Add(time.Hour),
				NotAfter:    time.Now().Add(time.Hour),
			}, nil).Once()

		token := suite.validJWTToken(1)
		res := suite.httpPlainTextRequest("POST", "/promos", "WELCOME2022", token)
		defer res.Body.Close()
		suite.Equal(http.StatusNotFound, res.StatusCode)
		resJSON := suite.parseJSON(res.Body)
		suite.Equal(1001., resJSON["code"])
	})

	suite.Run("internal error", func() {
		suite.repo.On("PromoGetByCode", mock.Anything, "WELCOME2022").
			Return(nil, errs.ErrInternal).Once()

		token := suite.validJWTToken(1)
		res := suite.httpPlainTextRequest("POST", "/promos", "WELCOME2022", token)
		defer res.Body.Close()
		suite.Equal(http.StatusInternalServerError, res.StatusCode)
		resJSON := suite.parseJSON(res.Body)
		suite.Equal(1000., resJSON["code"])
	})
}

func (suite *handlersSuite) TestOrderAccrualList() {
	suite.Run("success", func() {
		suite.repo.On("OperationGetByType", mock.Anything, uint64(1), models.OrderAccrual).
			Return([]*models.Operation{
				{
					ID:          1,
					UserID:      1,
					OrderNumber: strPtr("12345678901"),
					Type:        models.OrderAccrual,
					Status:      models.StatusNew,
					Amount:      decimal.NewFromInt(100),
					CreatedAt:   time.Now(),
				},
				{
					ID:          2,
					UserID:      1,
					OrderNumber: strPtr("12345678902"),
					Type:        models.OrderAccrual,
					Status:      models.StatusProcessing,
					Amount:      decimal.NewFromInt(100),
					CreatedAt:   time.Now(),
				},
			}, nil).Once()

		token := suite.validJWTToken(1)
		res := suite.httpPlainTextRequest("GET", "/orders", "", token)
		defer res.Body.Close()
		suite.Equal(http.StatusOK, res.StatusCode)
		resJSON := suite.parseJSONList(res.Body)
		suite.Equal(2, len(resJSON))
	})

	suite.Run("no content", func() {
		suite.repo.On("OperationGetByType", mock.Anything, uint64(1), models.OrderAccrual).
			Return(nil, errs.ErrNotFound).Once()

		token := suite.validJWTToken(1)
		res := suite.httpPlainTextRequest("GET", "/orders", "", token)
		defer res.Body.Close()
		suite.Equal(http.StatusNoContent, res.StatusCode)
	})

	suite.Run("internal error", func() {
		suite.repo.On("OperationGetByType", mock.Anything, uint64(1), models.OrderAccrual).
			Return(nil, errs.ErrInternal).Once()

		token := suite.validJWTToken(1)
		res := suite.httpPlainTextRequest("GET", "/orders", "", token)
		defer res.Body.Close()
		suite.Equal(http.StatusInternalServerError, res.StatusCode)
		resJSON := suite.parseJSON(res.Body)
		suite.Equal(1000., resJSON["code"])
	})
}

func (suite *handlersSuite) TestOrderWithdrawalList() {
	suite.Run("success", func() {
		suite.repo.On("OperationGetByType", mock.Anything, uint64(1), models.OrderWithdrawal).
			Return([]*models.Operation{
				{
					ID:          1,
					UserID:      1,
					OrderNumber: strPtr("12345678901"),
					Type:        models.OrderWithdrawal,
					Status:      models.StatusNew,
					Amount:      decimal.NewFromInt(-100),
					CreatedAt:   time.Now(),
				},
				{
					ID:          2,
					UserID:      1,
					OrderNumber: strPtr("12345678902"),
					Type:        models.OrderWithdrawal,
					Status:      models.StatusProcessing,
					Amount:      decimal.NewFromInt(-100),
					CreatedAt:   time.Now(),
				},
			}, nil).Once()

		token := suite.validJWTToken(1)
		res := suite.httpPlainTextRequest("GET", "/withdrawals", "", token)
		defer res.Body.Close()
		suite.Equal(http.StatusOK, res.StatusCode)
		resJSON := suite.parseJSONList(res.Body)
		suite.Equal(2, len(resJSON))
		suite.Equal(100., resJSON[0]["sum"])
	})

	suite.Run("no content", func() {
		suite.repo.On("OperationGetByType", mock.Anything, uint64(1), models.OrderWithdrawal).
			Return(nil, errs.ErrNotFound).Once()

		token := suite.validJWTToken(1)
		res := suite.httpPlainTextRequest("GET", "/withdrawals", "", token)
		defer res.Body.Close()
		suite.Equal(http.StatusNoContent, res.StatusCode)
	})

	suite.Run("internal error", func() {
		suite.repo.On("OperationGetByType", mock.Anything, uint64(1), models.OrderWithdrawal).
			Return(nil, errs.ErrInternal).Once()

		token := suite.validJWTToken(1)
		res := suite.httpPlainTextRequest("GET", "/withdrawals", "", token)
		defer res.Body.Close()
		suite.Equal(http.StatusInternalServerError, res.StatusCode)
		resJSON := suite.parseJSON(res.Body)
		suite.Equal(1000., resJSON["code"])
	})
}
