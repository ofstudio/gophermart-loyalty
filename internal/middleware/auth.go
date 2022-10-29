package middleware

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-chi/render"
	"github.com/golang-jwt/jwt/v4"
	"github.com/golang-jwt/jwt/v4/request"

	"gophermart-loyalty/internal/errs"
)

type contextKey struct {
	name string
}

var ctxUserID = &contextKey{"user_id"}

// Auth  - middleware для проверки авторизации.
// Формат заголовка запроса:
//    Authorization: Bearer <JWT token>
// ...либо
//    Authorization: <JWT token>
func Auth(alg, key string) func(next http.Handler) http.Handler {
	a := newAuthorizator(alg, key)
	return a.handler
}

// authorizator - хранит конфигурацию для авторизации
type authorizator struct {
	alg string
	key string
}

func newAuthorizator(alg, key string) *authorizator {
	return &authorizator{alg: alg, key: key}
}

// handler - хандлер для проверки авторизации
func (a *authorizator) handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Извлекаем ID пользователя из запроса
		userID, err := a.extractUserID(r)
		if err != nil {
			_ = render.Render(w, r, errs.ErrResponseUnauthorized)
			return
		}

		// Добавляем в контекст запроса ID пользователя
		ctx := r.Context()
		ctx = context.WithValue(ctx, ctxUserID, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// extractUserID - извлекает ID пользователя из запроса
func (a *authorizator) extractUserID(r *http.Request) (uint64, error) {
	// Извлекаем токен из запроса
	token, err := a.extractJWTToken(r)
	if err != nil {
		return 0, err
	}

	// Проверяем, что токен валидный
	if !token.Valid {
		return 0, fmt.Errorf("invalid token")
	}

	// Получаем claims из токена
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return 0, fmt.Errorf("claims extraction failed")
	}

	// Проверяем, что claims содержит число в поле "sub"
	sub, ok := claims["sub"].(float64)
	userID := uint64(sub)
	if !ok || userID <= 0 {
		return 0, fmt.Errorf("claims does not contain valid sub")
	}

	_, ok = claims["exp"].(float64)
	if !ok {
		return 0, fmt.Errorf("claims does not contain valid exp")
	}

	_, ok = claims["nbf"].(float64)
	if !ok {
		return 0, fmt.Errorf("claims does not contain valid nbf")
	}

	return userID, nil
}

// extractJWTToken - извлекает токен из запроса и проверяет алгоритм подписи
func (a *authorizator) extractJWTToken(r *http.Request) (*jwt.Token, error) {
	// Получаем JWT из заголовка запроса
	// AuthorizationHeaderExtractor - извлекает токен из заголовка Authorization
	// Если в заголовке пришел Bearer <token>, то извлекает только <token>
	return request.ParseFromRequest(r, request.AuthorizationHeaderExtractor, func(token *jwt.Token) (interface{}, error) {
		// Проверяем метод подписи токена
		if token.Method.Alg() != a.alg {
			return nil, fmt.Errorf("unsupported signing method: %v, want %v", token.Method.Alg(), a.alg)
		}

		return []byte(a.key), nil
	})
}

// GetUserID - возвращает userID из контекста
func GetUserID(ctx context.Context) (uint64, bool) {
	id, ok := ctx.Value(ctxUserID).(uint64)
	if !ok {
		return 0, false
	}
	return id, true
}
