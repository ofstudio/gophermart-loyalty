package handlers

import (
	"context"
	"fmt"
	"github.com/go-chi/render"
	"github.com/golang-jwt/jwt/v4"
	"github.com/golang-jwt/jwt/v4/request"
	"net/http"
)

// authMiddleware - middleware для проверки авторизации.
// Формат заголовка запроса:
//    Authorization: Bearer <JWT token>
// ...либо
//    Authorization: <JWT token>
func (h *Handlers) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Извлекаем ID пользователя из запроса
		userID, err := h.extractUserID(r)
		if err != nil {
			h.log.Debug().Err(err).Msg("failed to extract user ID")
			_ = render.Render(w, r, ErrUnauthorized)
			return
		}

		h.log.Debug().Uint64("user_id", userID).Msg("user ID extracted")
		// Добавляем в контекст запроса ID пользователя
		ctx := r.Context()
		ctx = context.WithValue(ctx, ctxUserID, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// extractUserID - извлекает ID пользователя из запроса
func (h *Handlers) extractUserID(r *http.Request) (uint64, error) {
	// Извлекаем токен из запроса
	token, err := h.extractJWTToken(r)
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
func (h *Handlers) extractJWTToken(r *http.Request) (*jwt.Token, error) {
	// Получаем JWT из заголовка запроса
	// AuthorizationHeaderExtractor - извлекает токен из заголовка Authorization
	// Если в заголовке пришел Bearer <token>, то извлекает только <token>
	return request.ParseFromRequest(r, request.AuthorizationHeaderExtractor, func(token *jwt.Token) (interface{}, error) {
		// Проверяем метод подписи токена
		if token.Method.Alg() != h.cfg.SigningAlg {
			return nil, fmt.Errorf(
				"unsupported signing method: %v, want %v",
				token.Method.Alg(),
				h.cfg.SigningAlg,
			)
		}

		return []byte(h.cfg.SigningKey), nil
	})
}

// getUserID - возвращает userID из контекста
func (h *Handlers) getUserID(ctx context.Context) (uint64, bool) {
	id, ok := ctx.Value(ctxUserID).(uint64)
	if !ok {
		return 0, false
	}
	return id, true
}

type contextKey struct {
	name string
}

var ctxUserID = &contextKey{"user_id"}
