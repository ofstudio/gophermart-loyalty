package main

import (
	"context"
	"syscall"

	"github.com/rs/zerolog"

	"gophermart-loyalty/internal/app"
	"gophermart-loyalty/internal/config"
	"gophermart-loyalty/internal/logger"
	"gophermart-loyalty/pkg/shutdown"
)

func main() {
	// Логгер
	log := logger.NewLogger(zerolog.InfoLevel)

	// Читаем конфигурацию
	cfg, err := config.Compose(config.NewDefault, config.NewFromEnv, config.NewFromCLI)
	if err != nil {
		log.Fatal().Err(err).Msg("error while loading config")
	}

	// Контекст для остановки приложения
	ctx, cancel := shutdown.ContextWithShutdown(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// Запускаем приложение
	application := app.NewApp(cfg, log)
	if err = application.Start(ctx); err != nil {
		log.Fatal().Err(err).Msg("error while starting application")
	}

	log.Info().Msg("exiting")
}
