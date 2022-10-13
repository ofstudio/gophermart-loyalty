package usecases

import (
	"context"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/suite"
	"gophermart-loyalty/internal/logger"
	"gophermart-loyalty/internal/mocks"
	"testing"
)

func TestUseCasesSuite(t *testing.T) {
	suite.Run(t, new(useCasesSuite))
}

type useCasesSuite struct {
	suite.Suite
	log      logger.Log
	repo     *mocks.RepoInterface
	useCases *UseCases
}

func (suite *useCasesSuite) SetupSuite() {
	suite.log = logger.NewLogger(zerolog.DebugLevel)
}

func (suite *useCasesSuite) SetupTest() {
	suite.repo = mocks.NewRepoInterface(suite.T())
	suite.useCases = NewUseCases(suite.repo, suite.log)
}

func (suite *useCasesSuite) ctx() context.Context {
	return context.WithValue(context.Background(), middleware.RequestIDKey, suite.T().Name())
}
