package postgres

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
)

// Database defines the interface for database operations
type Database interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
}

// Ensure *sqlx.DB implements the Database interface
var _ Database = (*sqlx.DB)(nil)
