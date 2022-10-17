package usecases

import (
	"gophermart-loyalty/internal/logger"
	"gophermart-loyalty/internal/repo"
)

// UseCases - набор бизнес-логики.
type UseCases struct {
	repo repo.Repo
	log  logger.Log
}

func NewUseCases(repo repo.Repo, log logger.Log) *UseCases {
	return &UseCases{
		repo: repo,
		log:  log,
	}
}
