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
	db    *sql.DB
	log   logger.Log
	stmts []*sql.Stmt
}

// NewPGXRepo - создает новый репозиторий
func NewPGXRepo(cfg *config.DB, log logger.Log) (*PGXRepo, error) {
	var err error

	// Подключаемся к БД
	db, err := sql.Open("pgx", cfg.URI)
	if err != nil || db.Ping() != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}
	log.Info().Msg("connected to db")

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
	stmts, err := prepareStmts(db)
	if err != nil {
		return nil, err
	}
	log.Info().Msg("statements prepared")
	return &PGXRepo{db: db, log: log, stmts: stmts}, nil
}

// DB - возвращает соединение с БД.
func (r *PGXRepo) DB() *sql.DB {
	return r.db
}

// Close - закрывает соединение с БД.
func (r *PGXRepo) Close() error {
	return r.db.Close()
}
