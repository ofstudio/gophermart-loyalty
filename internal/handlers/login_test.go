package handlers

import (
	"net/http"

	"github.com/golang-jwt/jwt/v4"
	"github.com/stretchr/testify/mock"

	"gophermart-loyalty/internal/app"
	"gophermart-loyalty/internal/models"
)

func (suite *handlersSuite) TestLogin() {
	reqBody := `{"login":"test","password":"test"}`
	passHash := suite.passHash("test")

	suite.Run("success", func() {
		suite.repo.On("UserGetByLogin", mock.Anything, "test").
			Return(&models.User{ID: 1, Login: "test", PassHash: passHash}, nil).Once()

		res := suite.httpJSONRequest(http.MethodPost, "/login", reqBody, "")
		defer res.Body.Close()
		suite.Equal(http.StatusOK, res.StatusCode)
		resJSON := suite.parseJSON(res.Body)

		// проверяем ответ
		suite.Equal("Bearer", resJSON["token_type"])
		suite.NotEmpty(resJSON["access_token"])
		suite.Equal(60., resJSON["expires_in"])

		// проверяем токен
		token, err := jwt.Parse(resJSON["access_token"].(string), func(token *jwt.Token) (interface{}, error) {
			return []byte(suite.cfg.SigningKey), nil
		})
		suite.NoError(err)
		suite.True(token.Valid)
		suite.Equal("HS256", token.Header["alg"])
		suite.Equal(1., token.Claims.(jwt.MapClaims)["sub"])
	})

	suite.Run("invalid login or password", func() {
		suite.repo.On("UserGetByLogin", mock.Anything, "test").
			Return(nil, app.ErrNotFound).Once()

		res := suite.httpJSONRequest(http.MethodPost, "/login", reqBody, "")
		defer res.Body.Close()
		suite.Equal(http.StatusUnauthorized, res.StatusCode)
		resJSON := suite.parseJSON(res.Body)
		suite.Equal(1103., resJSON["code"])
	})

	suite.Run("invalid request body", func() {
		res := suite.httpJSONRequest(http.MethodPost, "/login", "invalid", "")
		defer res.Body.Close()
		suite.Equal(http.StatusBadRequest, res.StatusCode)
		resJSON := suite.parseJSON(res.Body)
		suite.Equal(1003., resJSON["code"])
	})

	suite.Run("token generation error", func() {
		suite.repo.On("UserGetByLogin", mock.Anything, "test").
			Return(&models.User{ID: 1, Login: "test", PassHash: passHash}, nil).Once()

		// временно подменяем алгоритм подписи токена
		m := suite.handlers.cfg.SigningAlg
		suite.cfg.SigningAlg = ""
		defer func() { suite.handlers.cfg.SigningAlg = m }()

		res := suite.httpJSONRequest(http.MethodPost, "/login", reqBody, "")
		defer res.Body.Close()
		suite.Equal(http.StatusInternalServerError, res.StatusCode)
		resJSON := suite.parseJSON(res.Body)
		suite.Equal(1000., resJSON["code"])
	})
}
