package handlers

import (
	"github.com/stretchr/testify/mock"
	"gophermart-loyalty/internal/app"
	"gophermart-loyalty/internal/models"
	"net/http"
)

func (suite *handlersSuite) TestRegister() {
	reqBody := `{"login":"test","password":"q123456"}`

	suite.Run("success", func() {
		suite.useCases.
			On("UserCreate", mock.Anything, "test", "q123456").
			Return(&models.User{ID: 1, Login: "123456"}, nil).Once()

		res := suite.httpRequest("POST", "/register", reqBody, "")
		defer res.Body.Close()
		suite.Equal(http.StatusCreated, res.StatusCode)
	})

	suite.Run("failed to bind request", func() {
		res := suite.httpRequest("POST", "/register", `{malformed json`, "")
		defer res.Body.Close()
		suite.Equal(http.StatusBadRequest, res.StatusCode)
		resJSON := suite.parseJSON(res.Body)
		suite.Equal(1003., resJSON["code"])
	})

	suite.Run("user already exists", func() {
		suite.useCases.
			On("UserCreate", mock.Anything, "test", "q123456").
			Return(nil, app.ErrUserAlreadyExists).Once()

		res := suite.httpRequest("POST", "/register", reqBody, "")
		defer res.Body.Close()
		suite.Equal(http.StatusConflict, res.StatusCode)
		resJSON := suite.parseJSON(res.Body)
		suite.Equal(1100., resJSON["code"])
	})

	suite.Run("internal error", func() {
		suite.useCases.
			On("UserCreate", mock.Anything, "test", "q123456").
			Return(nil, app.ErrInternal).Once()

		res := suite.httpRequest("POST", "/register", reqBody, "")
		defer res.Body.Close()
		suite.Equal(http.StatusInternalServerError, res.StatusCode)
		resJSON := suite.parseJSON(res.Body)
		suite.Equal(1000., resJSON["code"])
	})

	suite.Run("invalid login", func() {
		suite.useCases.
			On("UserCreate", mock.Anything, "test", "q123456").
			Return(nil, app.ErrUserLoginInvalid).Once()

		res := suite.httpRequest("POST", "/register", reqBody, "")
		defer res.Body.Close()
		suite.Equal(http.StatusBadRequest, res.StatusCode)
		resJSON := suite.parseJSON(res.Body)
		suite.Equal(1101., resJSON["code"])
	})

	suite.Run("invalid password", func() {
		suite.useCases.
			On("UserCreate", mock.Anything, "test", "q123456").
			Return(nil, app.ErrUserPassInvalid).Once()

		res := suite.httpRequest("POST", "/register", reqBody, "")
		defer res.Body.Close()
		suite.Equal(http.StatusBadRequest, res.StatusCode)
		resJSON := suite.parseJSON(res.Body)
		suite.Equal(1102., resJSON["code"])
	})
}
