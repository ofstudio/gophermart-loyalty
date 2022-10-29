package middleware

import (
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

// - [x] Успешная авторизация
// - [x] Нет заголовка с токеном
// - [x] Невалидный токен
// - [x] Токен подписан неверным ключом
// - [x] Токен подписан неверным алгоритмом
// - [x] Токен просрочен
// - [x] Токен используется раньше времени
// - [x] Токен не содержит поля sub
// - [x] Токен содержит невалидное поле sub
// - [x] Токен не содержит поле exp
// - [x] Токен не содержит поле nbf

func (suite *middlewareSuite) TestAuthMiddleware() {
	suite.Run("success", func() {
		claims := jwt.MapClaims{
			"sub": 1,
			"iat": time.Now().Unix(),
			"nbf": time.Now().Unix(),
			"exp": time.Now().Add(1 * time.Hour).Unix(),
		}
		token := suite.generateJWTToken(claims, "HS256", "test1234567890")
		res := suite.httpRequest(http.MethodGet, "/private", "", "", token)
		defer res.Body.Close()
		suite.Equal(http.StatusOK, res.StatusCode)
	})

	suite.Run("no token", func() {
		res := suite.httpRequest(http.MethodGet, "/private", "", "", "")
		defer res.Body.Close()
		suite.Equal(http.StatusUnauthorized, res.StatusCode)
	})

	suite.Run("invalid token", func() {
		res := suite.httpRequest(http.MethodGet, "/private", "", "", "invalid token")
		defer res.Body.Close()
		suite.Equal(http.StatusUnauthorized, res.StatusCode)
	})

	suite.Run("token signed with wrong key", func() {
		claims := jwt.MapClaims{
			"sub": 1,
			"iat": time.Now().Unix(),
			"nbf": time.Now().Unix(),
			"exp": time.Now().Add(1 * time.Hour).Unix(),
		}
		token := suite.generateJWTToken(claims, "HS256", "wrong key")

		res := suite.httpRequest(http.MethodGet, "/private", "", "", token)
		defer res.Body.Close()
		suite.Equal(http.StatusUnauthorized, res.StatusCode)
	})

	suite.Run("token signed with unsupported alg", func() {
		claims := jwt.MapClaims{
			"sub": 1,
			"iat": time.Now().Unix(),
			"nbf": time.Now().Unix(),
			"exp": time.Now().Add(1 * time.Hour).Unix(),
		}
		token := suite.generateJWTToken(claims, "HS384", "test1234567890")
		res := suite.httpRequest(http.MethodGet, "/private", "", "", token)
		defer res.Body.Close()
		suite.Equal(http.StatusUnauthorized, res.StatusCode)
	})

	suite.Run("token expired", func() {
		claims := jwt.MapClaims{
			"sub": 1,
			"iat": time.Now().Unix(),
			"nbf": time.Now().Unix(),
			"exp": time.Now().Add(-1 * time.Hour).Unix(),
		}
		token := suite.generateJWTToken(claims, "HS256", "test1234567890")

		res := suite.httpRequest(http.MethodGet, "/private", "", "", token)
		defer res.Body.Close()
		suite.Equal(http.StatusUnauthorized, res.StatusCode)
	})

	suite.Run("token used before time", func() {
		claims := jwt.MapClaims{
			"sub": 1,
			"iat": time.Now().Unix(),
			"nbf": time.Now().Add(1 * time.Hour).Unix(),
			"exp": time.Now().Add(2 * time.Hour).Unix(),
		}
		token := suite.generateJWTToken(claims, "HS256", "test1234567890")

		res := suite.httpRequest(http.MethodGet, "/private", "", "", token)
		defer res.Body.Close()
		suite.Equal(http.StatusUnauthorized, res.StatusCode)
	})

	suite.Run("token without sub claim", func() {
		claims := jwt.MapClaims{
			"iat": time.Now().Unix(),
			"nbf": time.Now().Unix(),
			"exp": time.Now().Add(1 * time.Hour).Unix(),
		}
		token := suite.generateJWTToken(claims, "HS256", "test1234567890")
		res := suite.httpRequest(http.MethodGet, "/private", "", "", token)
		defer res.Body.Close()
		suite.Equal(http.StatusUnauthorized, res.StatusCode)
	})

	suite.Run("token with invalid sub claim", func() {
		claims := jwt.MapClaims{
			"sub": "invalid",
			"iat": time.Now().Unix(),
			"nbf": time.Now().Unix(),
			"exp": time.Now().Add(1 * time.Hour).Unix(),
		}
		token := suite.generateJWTToken(claims, "HS256", "test1234567890")
		res := suite.httpRequest(http.MethodGet, "/private", "", "", token)
		defer res.Body.Close()
		suite.Equal(http.StatusUnauthorized, res.StatusCode)
	})

	suite.Run("token without exp claim", func() {
		claims := jwt.MapClaims{
			"sub": 1,
			"iat": time.Now().Unix(),
			"nbf": time.Now().Unix(),
		}
		token := suite.generateJWTToken(claims, "HS256", "test1234567890")
		res := suite.httpRequest(http.MethodGet, "/private", "", "", token)
		defer res.Body.Close()
		suite.Equal(http.StatusUnauthorized, res.StatusCode)
	})

	suite.Run("token without nbf claim", func() {
		claims := jwt.MapClaims{
			"sub": 1,
			"iat": time.Now().Unix(),
			"exp": time.Now().Add(1 * time.Hour).Unix(),
		}
		token := suite.generateJWTToken(claims, "HS256", "test1234567890")
		res := suite.httpRequest(http.MethodGet, "/private", "", "", token)
		defer res.Body.Close()
		suite.Equal(http.StatusUnauthorized, res.StatusCode)
	})

}

func (suite *middlewareSuite) generateJWTToken(claims jwt.Claims, alg, key string) string {
	signingMethod := jwt.GetSigningMethod(alg)
	suite.Require().NotNil(signingMethod)
	token := jwt.NewWithClaims(signingMethod, claims)
	tokenString, err := token.SignedString([]byte(key))
	suite.Require().NoError(err)
	return tokenString
}
