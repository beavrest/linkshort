package db

import (
	"database/sql"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// NewPostgres opens a database/sql handle using the pgx stdlib driver.
// It does not verify connectivity — the caller should ping as needed.
func NewPostgres(dsn string) (*sql.DB, error) {
	return sql.Open("pgx", dsn)
}
