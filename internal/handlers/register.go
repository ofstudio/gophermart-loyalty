package handlers

import (
	"github.com/go-chi/render"
	"net/http"
)

// RegisterRequest - render.Binder для запроса на регистрацию
type RegisterRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

func (req *RegisterRequest) Bind(_ *http.Request) error {
	return nil
}

func (h *Handlers) register(w http.ResponseWriter, r *http.Request) {
	data := &RegisterRequest{}
	if err := render.Bind(r, data); err != nil {
		h.log.Debug().Err(err).Msg("failed to bind request")
		_ = render.Render(w, r, ErrBadRequest)
		return
	}

	if _, err := h.useCases.UserCreate(r.Context(), data.Login, data.Password); err != nil {
		_ = render.Render(w, r, NewErrResponse(err))
		return
	}

	w.WriteHeader(http.StatusCreated)
}
