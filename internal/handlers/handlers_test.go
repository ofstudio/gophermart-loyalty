package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/suite"
	"gophermart-loyalty/internal/config"
	"gophermart-loyalty/internal/logger"
	"gophermart-loyalty/internal/mocks"
	"gophermart-loyalty/internal/usecases"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHandlersSuite(t *testing.T) {
	suite.Run(t, new(handlersSuite))
}

type handlersSuite struct {
	suite.Suite
	log        logger.Log
	repo       *mocks.Repo
	useCases   *usecases.UseCases
	handlers   *Handlers
	cfg        *config.Auth
	testServer *httptest.Server
}

func (suite *handlersSuite) SetupSuite() {
	suite.log = logger.NewLogger(zerolog.DebugLevel)
	suite.cfg = &config.Auth{
		SigningAlg: "HS256",
		TTL:        60 * time.Second,
		SigningKey: "test123456789012345678901234567890",
	}
}

func (suite *handlersSuite) SetupTest() {
	suite.repo = mocks.NewRepo(suite.T())
	suite.useCases = usecases.NewUseCases(suite.repo, suite.log)
	suite.handlers = NewHandlers(suite.cfg, suite.useCases, suite.log)
	r := suite.handlers.Routes()

	// Дополнительный раут для тестирования авторизации
	r.Group(func(r chi.Router) {
		r.Use(suite.handlers.authMiddleware)
		r.Get("/test-auth", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
	})

	suite.testServer = httptest.NewServer(r)
}

func (suite *handlersSuite) TearDownTest() {
	suite.testServer.Close()
}

func (suite *handlersSuite) httpRequest(method, url, contentType, body, token string) *http.Response {
	url = fmt.Sprintf("%s%s", suite.testServer.URL, url)
	req, err := http.NewRequest(method, url, bytes.NewBuffer([]byte(body)))
	suite.NoError(err)
	if token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	}
	req.Header.Set("Content-Type", contentType)
	res, err := http.DefaultClient.Do(req)
	suite.NoError(err)
	return res
}

func (suite *handlersSuite) httpJSONRequest(method, url, body, token string) *http.Response {
	return suite.httpRequest(method, url, "application/json", body, token)
}

func (suite *handlersSuite) httpPlainTextRequest(method, url, body, token string) *http.Response {
	return suite.httpRequest(method, url, "text/plain", body, token)
}

func (suite *handlersSuite) parseJSON(body io.Reader) map[string]interface{} {
	var resJSON map[string]interface{}
	suite.NoError(json.Unmarshal(suite.getBody(body), &resJSON))
	return resJSON
}

func (suite *handlersSuite) parseJSONList(body io.Reader) []map[string]interface{} {
	var resJSON []map[string]interface{}
	suite.NoError(json.Unmarshal(suite.getBody(body), &resJSON))
	return resJSON
}

func (suite *handlersSuite) getBody(body io.Reader) []byte {
	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(body)
	suite.NoError(err)
	return buf.Bytes()
}

func strPtr(s string) *string {
	return &s
}

func uint64Ptr(i uint64) *uint64 {
	return &i
}
