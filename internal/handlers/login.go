package handlers

import (
	"fmt"
	"github.com/go-chi/render"
	"github.com/golang-jwt/jwt/v4"
	"net/http"
)

type LoginRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

func (req *LoginRequest) Bind(_ *http.Request) error {
	return nil
}

type LoginResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int64  `json:"expires_in"`
}

func (res *LoginResponse) Render(_ http.ResponseWriter, _ *http.Request) error {
	return nil
}

// login - аутентификация пользователя.
// Формат запроса:
//     POST /api/user/login HTTP/1.1
//     Content-Type: application/json
//
//    {
//        "login": "<login>",
//        "password": "<password>"
//    }
//
// Возможные коды ответа:
//    200 — пользователь успешно аутентифицирован
//    400 — неверный формат запроса
//    401 — неверная пара логин/пароль
//    500 — внутренняя ошибка сервера
//
// Формат ответа:
//    HTTP/1.1 200 OK
//    Content-Type: application/json
//
//    {
//        "access_token": "<token>",
//        "token_type": "Bearer",
//        "expires_in": 3600
//    }
func (h *Handlers) login(w http.ResponseWriter, r *http.Request) {
	data := &LoginRequest{}
	if err := render.Bind(r, data); err != nil {
		h.log.Debug().Err(err).Msg("failed to bind request")
		_ = render.Render(w, r, ErrBadRequest)
		return
	}

	// Ищем пользователя по логину и паролю
	user, err := h.useCases.UserCheckLoginPass(r.Context(), data.Login, data.Password)
	if err != nil {
		h.log.Debug().Err(err).Msg("failed to check login and password")
		_ = render.Render(w, r, NewErrResponse(err))
		return
	}

	// Генерируем токен
	token, err := h.generateJWTToken(user.ID)
	if err != nil {
		h.log.Debug().Err(err).Msg("failed to generate token")
		_ = render.Render(w, r, ErrInternal)
		return
	}
	// Отправляем токен в ответе
	_ = render.Render(w, r, &LoginResponse{
		AccessToken: token,
		TokenType:   "Bearer",
		ExpiresIn:   int64(h.cfgAuth.TTL.Seconds()),
	})
}

// generateJWTToken - генерирует токен JWT для пользователя
func (h *Handlers) generateJWTToken(userID uint64) (string, error) {
	// Создаем claims
	now := jwt.TimeFunc()
	claims := jwt.MapClaims{
		"sub": userID,                        // subject
		"nbf": now.Unix(),                    // not before
		"iat": now.Unix(),                    // issued at
		"exp": now.Add(h.cfgAuth.TTL).Unix(), // expires at
	}

	// Создаем токен
	signingMethod := jwt.GetSigningMethod(h.cfgAuth.SigningAlg)
	if signingMethod == nil {
		return "", fmt.Errorf("unknown signing method: %s", h.cfgAuth.SigningAlg)
	}

	token := jwt.NewWithClaims(signingMethod, claims)
	// Подписываем токен
	return token.SignedString([]byte(h.cfgAuth.SigningKey))
}
