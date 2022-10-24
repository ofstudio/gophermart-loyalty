package handlers

import (
	"io/ioutil"
	"net/http"

	"github.com/go-chi/render"

	"gophermart-loyalty/internal/errs"
)

func decodePlainText(r *http.Request) (string, error) {
	contentType := render.GetRequestContentType(r)
	if contentType != render.ContentTypePlainText {
		return "", errs.ErrBadRequest
	}
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
