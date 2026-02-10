package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"

	_ "github.com/lib/pq"
)

// Config holds the PostgreSQL configuration
type Config struct {
	User     string
	Host     string
	Password string
	Database string
	Port     string
	SSLMode  string
}

// ConnectFunc is a function that creates a database connection
type ConnectFunc func(ctx context.Context, driverName, dataSourceName string) (*sqlx.DB, error)

// connectFunc is the default connection function
var connectFunc ConnectFunc = func(ctx context.Context, driverName, dataSourceName string) (*sqlx.DB, error) {
	return sqlx.ConnectContext(ctx, driverName, dataSourceName)
}

// configurePoolFunc configures the connection pool
type configurePoolFunc func(db *sqlx.DB)

var configurePool configurePoolFunc = func(db *sqlx.DB) {
	db.SetMaxIdleConns(3)
	db.SetMaxOpenConns(10)
	db.SetConnMaxIdleTime(5 * time.Second)
	db.SetConnMaxLifetime(-1)
}

// NewPostgresClient creates a new PostgreSQL client connection
func NewPostgresClient(ctx context.Context, cfg Config) (*sqlx.DB, error) {
	sslmode := cfg.SSLMode
	if sslmode == "" {
		if cfg.Database == "production" {
			sslmode = "require"
		} else {
			sslmode = "disable"
		}
	}

	connectionString := fmt.Sprintf(
		"user=%s dbname=%s sslmode=%s password=%s host=%s port=%s",
		cfg.User,
		cfg.Database,
		sslmode,
		cfg.Password,
		cfg.Host,
		cfg.Port,
	)

	db, err := connectFunc(ctx, "postgres", connectionString)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to postgres: %w", err)
	}

	configurePool(db)
	return db, nil
}
