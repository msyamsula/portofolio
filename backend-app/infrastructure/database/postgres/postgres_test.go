package postgres

import (
	"context"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPostgresClientSuccess(t *testing.T) {
	originalConnect := connectFunc
	originalPool := configurePool
	defer func() {
		connectFunc = originalConnect
		configurePool = originalPool
	}()

	ctx := context.Background()
	mockDB := &sqlx.DB{}
	connectFunc = func(ctx context.Context, driverName, dataSourceName string) (*sqlx.DB, error) {
		assert.Equal(t, "postgres", driverName)
		assert.Contains(t, dataSourceName, "user=testuser")
		assert.Contains(t, dataSourceName, "sslmode=disable")
		return mockDB, nil
	}
	configurePool = func(db *sqlx.DB) {} // Skip pool config in test

	cfg := Config{
		User:     "testuser",
		Host:     "localhost",
		Password: "testpass",
		Database: "testdb",
		Port:     "5432",
	}

	db, err := NewPostgresClient(ctx, cfg)
	require.NoError(t, err)
	assert.Same(t, mockDB, db)
}

func TestNewPostgresClientConnectionError(t *testing.T) {
	originalConnect := connectFunc
	originalPool := configurePool
	defer func() {
		connectFunc = originalConnect
		configurePool = originalPool
	}()

	ctx := context.Background()
	connectFunc = func(ctx context.Context, driverName, dataSourceName string) (*sqlx.DB, error) {
		return nil, assert.AnError
	}
	configurePool = func(db *sqlx.DB) {} // Skip pool config in test

	cfg := Config{
		User:     "testuser",
		Host:     "localhost",
		Password: "testpass",
		Database: "testdb",
		Port:     "5432",
	}

	db, err := NewPostgresClient(ctx, cfg)
	require.Error(t, err)
	assert.Nil(t, db)
	assert.Contains(t, err.Error(), "failed to connect to postgres")
}

func TestNewPostgresClientProductionSSLMode(t *testing.T) {
	originalConnect := connectFunc
	originalPool := configurePool
	defer func() {
		connectFunc = originalConnect
		configurePool = originalPool
	}()

	ctx := context.Background()
	connectFunc = func(ctx context.Context, driverName, dataSourceName string) (*sqlx.DB, error) {
		assert.Contains(t, dataSourceName, "sslmode=require")
		return &sqlx.DB{}, nil
	}
	configurePool = func(db *sqlx.DB) {} // Skip pool config in test

	cfg := Config{
		User:     "testuser",
		Host:     "localhost",
		Password: "testpass",
		Database: "production",
		Port:     "5432",
	}

	_, err := NewPostgresClient(ctx, cfg)
	require.NoError(t, err)
}

func TestNewPostgresClientExplicitSSLMode(t *testing.T) {
	originalConnect := connectFunc
	originalPool := configurePool
	defer func() {
		connectFunc = originalConnect
		configurePool = originalPool
	}()

	ctx := context.Background()
	connectFunc = func(ctx context.Context, driverName, dataSourceName string) (*sqlx.DB, error) {
		assert.Contains(t, dataSourceName, "sslmode=verify-full")
		return &sqlx.DB{}, nil
	}
	configurePool = func(db *sqlx.DB) {} // Skip pool config in test

	cfg := Config{
		User:     "testuser",
		Host:     "localhost",
		Password: "testpass",
		Database: "production",
		Port:     "5432",
		SSLMode:  "verify-full",
	}

	_, err := NewPostgresClient(ctx, cfg)
	require.NoError(t, err)
}

func TestNewPostgresClientDevSSLMode(t *testing.T) {
	originalConnect := connectFunc
	originalPool := configurePool
	defer func() {
		connectFunc = originalConnect
		configurePool = originalPool
	}()

	ctx := context.Background()
	connectFunc = func(ctx context.Context, driverName, dataSourceName string) (*sqlx.DB, error) {
		assert.Contains(t, dataSourceName, "sslmode=disable")
		return &sqlx.DB{}, nil
	}
	configurePool = func(db *sqlx.DB) {} // Skip pool config in test

	cfg := Config{
		User:     "testuser",
		Host:     "localhost",
		Password: "testpass",
		Database: "development",
		Port:     "5432",
	}

	_, err := NewPostgresClient(ctx, cfg)
	require.NoError(t, err)
}

func TestNewPostgresClientConnectionStringFormat(t *testing.T) {
	originalConnect := connectFunc
	originalPool := configurePool
	defer func() {
		connectFunc = originalConnect
		configurePool = originalPool
	}()

	ctx := context.Background()
	var capturedDSN string
	connectFunc = func(ctx context.Context, driverName, dataSourceName string) (*sqlx.DB, error) {
		capturedDSN = dataSourceName
		return &sqlx.DB{}, nil
	}
	configurePool = func(db *sqlx.DB) {} // Skip pool config in test

	cfg := Config{
		User:     "testuser",
		Host:     "testhost",
		Password: "testpass",
		Database: "testdb",
		Port:     "5432",
		SSLMode:  "disable",
	}

	_, err := NewPostgresClient(ctx, cfg)
	require.NoError(t, err)

	expected := "user=testuser dbname=testdb sslmode=disable password=testpass host=testhost port=5432"
	assert.Equal(t, expected, capturedDSN)
}

func TestNewPostgresClientEmptyConfig(t *testing.T) {
	originalConnect := connectFunc
	originalPool := configurePool
	defer func() {
		connectFunc = originalConnect
		configurePool = originalPool
	}()

	ctx := context.Background()
	connectFunc = func(ctx context.Context, driverName, dataSourceName string) (*sqlx.DB, error) {
		assert.Contains(t, dataSourceName, "user=")
		assert.Contains(t, dataSourceName, "dbname=")
		return &sqlx.DB{}, nil
	}
	configurePool = func(db *sqlx.DB) {} // Skip pool config in test

	cfg := Config{
		User:     "",
		Host:     "",
		Password: "",
		Database: "",
		Port:     "",
	}

	_, err := NewPostgresClient(ctx, cfg)
	require.NoError(t, err)
}

func TestNewPostgresClientStagingSSLMode(t *testing.T) {
	originalConnect := connectFunc
	originalPool := configurePool
	defer func() {
		connectFunc = originalConnect
		configurePool = originalPool
	}()

	ctx := context.Background()
	connectFunc = func(ctx context.Context, driverName, dataSourceName string) (*sqlx.DB, error) {
		assert.Contains(t, dataSourceName, "sslmode=disable")
		return &sqlx.DB{}, nil
	}
	configurePool = func(db *sqlx.DB) {} // Skip pool config in test

	cfg := Config{
		User:     "testuser",
		Host:     "localhost",
		Password: "testpass",
		Database: "staging",
		Port:     "5432",
	}

	_, err := NewPostgresClient(ctx, cfg)
	require.NoError(t, err)
}

func TestNewPostgresClientWithContext(t *testing.T) {
	originalConnect := connectFunc
	originalPool := configurePool
	defer func() {
		connectFunc = originalConnect
		configurePool = originalPool
	}()

	connectFunc = func(ctx context.Context, driverName, dataSourceName string) (*sqlx.DB, error) {
		return &sqlx.DB{}, nil
	}
	configurePool = func(db *sqlx.DB) {} // Skip pool config in test

	cfg := Config{
		User:     "testuser",
		Host:     "localhost",
		Password: "testpass",
		Database: "testdb",
		Port:     "5432",
	}

	ctx := context.Background()
	_, err := NewPostgresClient(ctx, cfg)
	require.NoError(t, err)
}
