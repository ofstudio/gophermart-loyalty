package handlers

import (
	"github.com/go-chi/render"
	"net/http"
)

type RegisterRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

func (req *RegisterRequest) Bind(_ *http.Request) error {
	return nil
}

// register - регистрация пользователя.
// Формат запроса:
//    POST /api/user/register HTTP/1.1
//    Content-Type: application/json
//
//    {
//        "login": "<login>",
//        "password": "<password>"
//    }
//
// Возможные коды ответа:
//    200 — пользователь успешно зарегистрирован и аутентифицирован
//    400 — неверный формат запроса
//    409 — логин уже занят
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
func (h *Handlers) register(w http.ResponseWriter, r *http.Request) {
	data := &RegisterRequest{}
	if err := render.Bind(r, data); err != nil {
		_ = render.Render(w, r, ErrBadRequest)
		return
	}

	user, err := h.useCases.UserCreate(r.Context(), data.Login, data.Password)
	if err != nil {
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
	render.Status(r, http.StatusCreated)
	_ = render.Render(w, r, &LoginResponse{
		AccessToken: token,
		TokenType:   "Bearer",
		ExpiresIn:   int64(h.cfg.TTL.Seconds()),
	})
}
