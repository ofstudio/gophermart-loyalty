package repo

import (
	"database/sql"
	"fmt"
)

var stmtQueries []string

// registerStmt - добавляет запрос в список запросов к БД.
// Возвращает индекс запроса в списке.
func registerStmt(query string) int {
	stmtQueries = append(stmtQueries, query)
	return len(stmtQueries) - 1
}

// prepareStmts - подготавливает стейтменты к БД.
// Возвращает список подготовленных стейтментов.
func prepareStmts(db *sql.DB) ([]*sql.Stmt, error) {
	statements := make([]*sql.Stmt, len(stmtQueries))
	for id, query := range stmtQueries {
		s, err := db.Prepare(query)
		if err != nil {
			return nil, fmt.Errorf("failed to prepare statement: %w in `%s`", err, query)
		}
		statements[id] = s
	}
	return statements, nil
}
