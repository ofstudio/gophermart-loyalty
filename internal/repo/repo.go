package repo

import (
	"database/sql"
	"fmt"

	_ "github.com/jackc/pgx/v4/stdlib"

	"gophermart-loyalty/internal/config"
	"gophermart-loyalty/internal/logger"
	"gophermart-loyalty/internal/repo/migrations"
)

// PGXRepo - репозиторий для работы с Postgres.
type PGXRepo struct {
	db         *sql.DB
	log        logger.Log
	statements []*sql.Stmt
}

// NewPGXRepo - создает новый репозиторий
func NewPGXRepo(cfg *config.DB, log logger.Log) (*PGXRepo, error) {
	var err error

	// Подключаемся к БД
	db, err := sql.Open("pgx", cfg.URI)
	if err != nil {
		return nil, fmt.Errorf("db open error: %w", err)
	}
	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("db connection error: %w", err)
	}
	log.Info().Msg("db connected")

	// Запускаем миграцию БД
	ver, err := migrations.Migrate(db)
	if err != nil {
		return nil, err
	}

	// Проверяем что версия БД соответствует заданной
	if ver != cfg.RequiredVersion {
		return nil, fmt.Errorf("db version mismatch: got %d, want %d", ver, cfg.RequiredVersion)
	}
	log.Info().Msgf("db migrated to version %d", ver)

	// Подготавливаем стейтменты
	stmts, err := prepareStatements(db)
	if err != nil {
		return nil, err
	}
	log.Info().Msg("db statements prepared")
	log.Info().Msg("repo created")
	return &PGXRepo{db: db, log: log, statements: stmts}, nil
}

// DB - возвращает соединение с БД.
func (r *PGXRepo) DB() *sql.DB {
	return r.db
}

// Close - закрывает соединение с БД.
func (r *PGXRepo) Close() {
	if err := r.db.Close(); err != nil {
		r.log.Error().Err(err).Msg("failed to close db connection")
	}
	r.log.Info().Msg("repo closed")
}
