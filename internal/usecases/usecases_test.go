package usecases

import (
	"context"
	"testing"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/suite"

	"gophermart-loyalty/internal/logger"
	"gophermart-loyalty/internal/mocks"
)

func TestUseCasesSuite(t *testing.T) {
	suite.Run(t, new(useCasesSuite))
}

type useCasesSuite struct {
	suite.Suite
	log      logger.Log
	repo     *mocks.Repo
	useCases *UseCases
}

func (suite *useCasesSuite) SetupSuite() {
	suite.log = logger.NewLogger(zerolog.DebugLevel)
}

func (suite *useCasesSuite) SetupTest() {
	suite.repo = mocks.NewRepo(suite.T())
	suite.useCases = NewUseCases(suite.repo, suite.log)
}

func (suite *useCasesSuite) ctx() context.Context {
	return context.WithValue(context.Background(), middleware.RequestIDKey, suite.T().Name())
}

func strPtr(s string) *string {
	return &s
}

func uint64Ptr(i uint64) *uint64 {
	return &i
}
