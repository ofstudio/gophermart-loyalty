package integrations

import (
	"context"
	"gophermart-loyalty/internal/app"
	"gophermart-loyalty/internal/logger"
	"gophermart-loyalty/internal/models"
	"gophermart-loyalty/internal/usecases"
	"strings"
	"time"
)

const (
	ShopStubStopped = iota
	ShopStubRunning
)

type ShopStub struct {
	status       int
	useCases     *usecases.UseCases
	log          logger.Log
	pollInterval time.Duration
}

func NewShopStub(u *usecases.UseCases, log logger.Log) *ShopStub {
	return &ShopStub{
		status:       ShopStubStopped,
		useCases:     u,
		log:          log,
		pollInterval: 1 * time.Second,
	}
}

func (s *ShopStub) Start(ctx context.Context) {
	go s.poll(ctx)
	s.status = ShopStubRunning
}

func (s *ShopStub) Status() int {
	return s.status
}

func (s *ShopStub) poll(ctx context.Context) {
	s.log.Info().Msg("shop-stub poller started")
	for {
		select {
		case <-ctx.Done():
			s.log.Info().Msg("shop-stub poller stopped")
			s.status = ShopStubStopped
			return
		case <-time.After(s.pollInterval):
			go s.updateFurther(ctx)
		}
	}
}

func (s *ShopStub) updateFurther(ctx context.Context) {
	op, err := s.useCases.OperationUpdateFurther(ctx, models.OrderWithdrawal, func(ctx context.Context, op *models.Operation) error {
		if op.OrderNumber == nil {
			s.log.Error().Uint64("operation_id", op.ID).Msg("order number is nil")
			return app.ErrInternal
		}
		if time.Since(op.CreatedAt) > time.Minute {
			if strings.HasPrefix(*op.OrderNumber, "000") {
				op.Status = models.StatusCanceled
			} else {
				op.Status = models.StatusProcessed
			}
		}
		return nil
	})

	if err == app.ErrNotFound {
		s.log.Debug().Msg("withdrawal operation: nothing to update")
		return
	} else if err != nil {
		s.log.Error().Err(err).Msg("withdrawal operation update failed")
		return
	}
	s.log.Info().Uint64("operation_id", op.ID).Msg("withdrawal operation updated")
}
