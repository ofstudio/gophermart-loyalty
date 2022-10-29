package repo

import (
	"database/sql"
	"fmt"
)

var queries []string

// registerStatement - добавляет запрос в список запросов к БД.
// Возвращает индекс запроса в списке.
func registerStatement(query string) int {
	queries = append(queries, query)
	return len(queries) - 1
}

// prepareStatements - подготавливает стейтменты к БД.
// Возвращает список подготовленных стейтментов.
func prepareStatements(db *sql.DB) ([]*sql.Stmt, error) {
	statements := make([]*sql.Stmt, len(queries))
	for id, query := range queries {
		s, err := db.Prepare(query)
		if err != nil {
			return nil, fmt.Errorf("failed to prepare statement: %w in `%s`", err, query)
		}
		statements[id] = s
	}
	return statements, nil
}
