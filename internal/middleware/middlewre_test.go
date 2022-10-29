package middleware

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestMiddlewareSuite(t *testing.T) {
	suite.Run(t, new(middlewareSuite))
}

type middlewareSuite struct {
	suite.Suite
	testServer *httptest.Server
}

func (suite *middlewareSuite) SetupTest() {
	mux := http.NewServeMux()
	mux.HandleFunc("/private", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Private area"))
	})
	r := Auth("HS256", "test1234567890")(mux)
	suite.testServer = httptest.NewServer(r)
}

func (suite *middlewareSuite) TearDownTest() {
	suite.testServer.Close()
}

func (suite *middlewareSuite) httpRequest(method, url, contentType, body, token string) *http.Response {
	url = fmt.Sprintf("%s%s", suite.testServer.URL, url)
	req, err := http.NewRequest(method, url, bytes.NewBuffer([]byte(body)))
	suite.NoError(err)
	if token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	res, err := http.DefaultClient.Do(req)
	suite.NoError(err)
	return res
}

func (suite *middlewareSuite) getBody(body io.Reader) []byte {
	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(body)
	suite.NoError(err)
	return buf.Bytes()
}
