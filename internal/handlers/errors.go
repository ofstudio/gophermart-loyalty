package handlers

import (
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"gophermart-loyalty/internal/app"
	"net/http"
)

var (
	ErrNotFound     = NewErrResponse(app.ErrNotFound)
	ErrInternal     = NewErrResponse(app.ErrInternal)
	ErrBadRequest   = NewErrResponse(app.ErrBadRequest)
	ErrUnauthorized = NewErrResponse(app.ErrUnauthorized)
)

// ErrResponse - render.Renderer для ошибок
type ErrResponse struct {
	HTTPCode     int    `json:"-"`
	Code         int    `json:"code,omitempty"`
	ErrorMessage string `json:"error"`
	RequestID    string `json:"request_id,omitempty"`
}

// NewErrResponse - возвращает ErrResponse для соответствующей ошибки приложения.
func NewErrResponse(err error) render.Renderer {
	appErr, ok := err.(*app.Error)
	if !ok {
		appErr = app.ErrInternal
	}
	return &ErrResponse{
		HTTPCode:     appErr.HTTPCode,
		Code:         appErr.Code,
		ErrorMessage: appErr.Error(),
	}
}

func (e *ErrResponse) Render(_ http.ResponseWriter, r *http.Request) error {
	render.Status(r, e.HTTPCode)
	e.RequestID = middleware.GetReqID(r.Context())
	return nil
}
