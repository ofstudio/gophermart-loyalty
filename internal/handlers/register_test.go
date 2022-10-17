package handlers

import (
	"github.com/stretchr/testify/mock"
	"gophermart-loyalty/internal/app"
	"net/http"
)

func (suite *handlersSuite) TestRegister() {
	reqBody := `{"login":"test","password":"q123456"}`

	suite.Run("success", func() {
		suite.repo.On("UserCreate", mock.Anything, mock.Anything).
			Return(nil).Once()

		res := suite.httpJSONRequest("POST", "/register", reqBody, "")
		defer res.Body.Close()
		suite.Equal(http.StatusOK, res.StatusCode)
		resJSON := suite.parseJSON(res.Body)
		suite.Equal("Bearer", resJSON["token_type"])
		suite.NotEmpty(resJSON["access_token"])
		suite.Equal(60., resJSON["expires_in"])
	})

	suite.Run("failed to bind request", func() {
		res := suite.httpJSONRequest("POST", "/register", `{malformed json`, "")
		defer res.Body.Close()
		suite.Equal(http.StatusBadRequest, res.StatusCode)
		resJSON := suite.parseJSON(res.Body)
		suite.Equal(1003., resJSON["code"])
	})

	suite.Run("user already exists", func() {
		suite.repo.On("UserCreate", mock.Anything, mock.Anything).
			Return(app.ErrUserAlreadyExists).Once()

		res := suite.httpJSONRequest("POST", "/register", reqBody, "")
		defer res.Body.Close()
		suite.Equal(http.StatusConflict, res.StatusCode)
		resJSON := suite.parseJSON(res.Body)
		suite.Equal(1100., resJSON["code"])
	})

	suite.Run("internal error", func() {
		suite.repo.On("UserCreate", mock.Anything, mock.Anything).
			Return(app.ErrInternal).Once()

		res := suite.httpJSONRequest("POST", "/register", reqBody, "")
		defer res.Body.Close()
		suite.Equal(http.StatusInternalServerError, res.StatusCode)
		resJSON := suite.parseJSON(res.Body)
		suite.Equal(1000., resJSON["code"])
	})

	suite.Run("invalid login", func() {
		res := suite.httpJSONRequest("POST", "/register", `{"login":"x","password":"q123456"}`, "")
		defer res.Body.Close()
		suite.Equal(http.StatusBadRequest, res.StatusCode)
		resJSON := suite.parseJSON(res.Body)
		suite.Equal(1101., resJSON["code"])
	})

	suite.Run("invalid password", func() {
		res := suite.httpJSONRequest("POST", "/register", `{"login":"test","password":"1"}`, "")
		defer res.Body.Close()
		suite.Equal(http.StatusBadRequest, res.StatusCode)
		resJSON := suite.parseJSON(res.Body)
		suite.Equal(1102., resJSON["code"])
	})
}
