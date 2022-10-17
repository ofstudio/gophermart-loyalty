package main

import (
	"context"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog"
	"gophermart-loyalty/internal/config"
	"gophermart-loyalty/internal/handlers"
	"gophermart-loyalty/internal/integrations"
	"gophermart-loyalty/internal/logger"
	"gophermart-loyalty/internal/repo"
	"gophermart-loyalty/internal/usecases"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var log = logger.NewLogger(zerolog.InfoLevel)

func main() {
	// Читаем конфигурацию
	cfg, err := config.Compose(config.Default, config.FromEnv, config.FromCLI)
	if err != nil {
		log.Fatal().Err(err).Msg("error while loading config")
	}

	// Создаём репозиторий, юзкейсы и интеграции
	repository, err := repo.NewPGXRepo(&cfg.DB, log)
	if err != nil {
		log.Fatal().Err(err).Msg("error while creating repository")
	}
	useCases := usecases.NewUseCases(repository, log)
	integrationAccrual := integrations.NewAccrual(&cfg.IntegrationAccrual, useCases, log)
	integrationShop := integrations.NewShopStub(useCases, log)

	// Создаём сервер
	h := handlers.NewHandlers(&cfg.Auth, useCases, log)
	r := chi.NewRouter()
	r.Use(middleware.Compress(5, "gzip"))
	r.Use(middleware.RealIP)
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Mount("/api/user", h.Routes())
	server := &http.Server{
		Addr:    cfg.RunAddress,
		Handler: r,
	}

	// Горутина для graceful-остановки приложения
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	exit := make(chan struct{})
	go func() {
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
		<-stop
		cancel()
		ctxSrv, cancelSrv := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancelSrv()
		_ = server.Shutdown(ctxSrv)
		log.Info().Msg("server stopped")
		close(exit)
	}()

	// Запускаем интеграции и сервер
	integrationAccrual.Start(ctx)
	integrationShop.Start(ctx)
	err = server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		log.Fatal().Err(err).Msg("error while starting server")
	}
	<-exit
	log.Info().Msg("exiting")
}
