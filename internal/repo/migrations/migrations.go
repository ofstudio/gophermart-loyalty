// Package migrations - миграции БД.
package migrations

import (
	"database/sql"
	"embed"
	"fmt"
	"github.com/pressly/goose/v3"
)

//go:embed *.sql
var embedMigrations embed.FS

func init() {
	goose.SetBaseFS(embedMigrations)
	goose.SetLogger(&nopLogger{})
}

func Migrate(db *sql.DB) (int64, error) {
	if err := goose.SetDialect("pgx"); err != nil {
		return -1, fmt.Errorf("failed to set migration dialect: %w", err)
	}
	if err := goose.Up(db, "."); err != nil {
		return -1, fmt.Errorf("failed to migrate database: %w", err)
	}
	ver, err := goose.GetDBVersion(db)
	if err != nil {
		return -1, fmt.Errorf("failed to get database migration version: %w", err)
	}
	return ver, nil
}

// nopLogger - copied from the future goose release
// https://github.com/pressly/goose/blob/2278d75a8657c9f9a9661d0b8d99e8b219e1a358/log.go#L37
type nopLogger struct{}

func (*nopLogger) Fatal(v ...interface{})                 {}
func (*nopLogger) Fatalf(format string, v ...interface{}) {}
func (*nopLogger) Print(v ...interface{})                 {}
func (*nopLogger) Println(v ...interface{})               {}
func (*nopLogger) Printf(format string, v ...interface{}) {}
