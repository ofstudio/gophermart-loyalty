package app

import (
	"context"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"gophermart-loyalty/internal/config"
	"gophermart-loyalty/internal/handlers"
	"gophermart-loyalty/internal/integrations"
	"gophermart-loyalty/internal/logger"
	"gophermart-loyalty/internal/repo"
	"gophermart-loyalty/internal/usecases"
)

// stopTimeout - таймаут остановки приложения
const stopTimeout = 5 * time.Second

// App - приложение
type App struct {
	cfg    *config.Config
	log    logger.Log
	server *http.Server
}

func NewApp(cfg *config.Config, log logger.Log) *App {
	return &App{cfg: cfg, log: log}
}

// Start - запуск приложения.
// ctx - контекст для остановки приложения.
func (a *App) Start(ctx context.Context) error {
	// Создаём репозиторий
	repository, err := repo.NewPGXRepo(&a.cfg.DB, a.log)
	if err != nil {
		return err
	}

	// Создаем юзкейсы
	useCases := usecases.NewUseCases(repository, a.log)

	// Создаём сервер
	h := handlers.NewHandlers(&a.cfg.Auth, useCases, a.log)
	r := chi.NewRouter()
	r.Use(middleware.RealIP)
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Mount("/api/user", h.InitRoutes())
	a.server = &http.Server{
		Addr:    a.cfg.RunAddress,
		Handler: r,
	}

	// Запускаем интеграции
	integrations.NewIntegrationAccrual(&a.cfg.IntegrationAccrual, useCases, a.log).Start(ctx)
	integrations.NewIntegrationShopStub(useCases, a.log).Start(ctx)

	// Горутина для остановки HTTP-сервера
	serverStopped := make(chan struct{})
	go func() {
		<-ctx.Done()
		a.stopServer()
		close(serverStopped)
	}()

	// Запускаем сервер
	a.log.Info().Msgf("starting server on %s", a.server.Addr)
	if err = a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}

	// Ждём сигнала остановки HTTP-сервера
	<-serverStopped

	// Закрываем репозиторий
	repository.Close()

	a.log.Info().Msg("application stopped")
	return nil
}

// stopServer - остановка HTTP-сервера
func (a *App) stopServer() {
	ctx, cancel := context.WithTimeout(context.Background(), stopTimeout)
	defer cancel()
	if err := a.server.Shutdown(ctx); err != nil {
		a.log.Error().Err(err).Msg("server shutdown error")
	}
	a.log.Info().Msg("server stopped")
}
