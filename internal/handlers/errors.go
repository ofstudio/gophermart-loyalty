package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"

	"gophermart-loyalty/internal/errs"
)

var (
	ErrInternal     = NewErrResponse(errs.ErrInternal)
	ErrBadRequest   = NewErrResponse(errs.ErrBadRequest)
	ErrUnauthorized = NewErrResponse(errs.ErrUnauthorized)
)

// ErrResponse - render.Renderer для ответов с ошибками.
// Формат ответа:
//    HTTP/1.1 404 Not Found
//    Content-Type: application/json
//
//    {
//        "code": 1001,
//        "message": "Not found"",
//        "request_id": "<request id>"
//    }
type ErrResponse struct {
	HTTPCode  int    `json:"-"`
	Code      int    `json:"code,omitempty"`
	Message   string `json:"error"`
	RequestID string `json:"request_id,omitempty"`
}

func NewErrResponse(err error) render.Renderer {
	appErr, ok := err.(*errs.Error)
	if !ok {
		appErr = errs.ErrInternal
	}
	return &ErrResponse{
		HTTPCode: appErr.HTTPCode,
		Code:     appErr.Code,
		Message:  appErr.Error(),
	}
}

func (e *ErrResponse) Render(_ http.ResponseWriter, r *http.Request) error {
	render.Status(r, e.HTTPCode)
	e.RequestID = middleware.GetReqID(r.Context())
	return nil
}
