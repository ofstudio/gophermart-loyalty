package usecases

import (
	"gophermart-loyalty/internal/logger"
	"gophermart-loyalty/internal/repo"
)

type UseCases struct {
	repo repo.RepoInterface
	log  logger.Log
}

func NewUseCases(repo repo.RepoInterface, log logger.Log) *UseCases {
	return &UseCases{
		repo: repo,
		log:  log,
	}
}
