package repository

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/msyamsula/portofolio/backend-app/domain/agent/dto"
)

// Repository defines the interface for agent data access
//
//go:generate mockgen -source=interface.go -destination=../../../mock/agent_repository_mock.go -package=mock -mock_names Repository=MockAgentRepository
type Repository interface {
	// GetConnectionInfo retrieves information about the current connection
	GetConnectionInfo(ctx context.Context) (dto.ConnectionInfo, error)

	// GetTables retrieves all tables in the current database
	GetTables(ctx context.Context) ([]dto.Table, error)

	// GetSchema retrieves the complete schema for a table
	GetSchema(ctx context.Context, tableName string) (dto.Column, error)

	// GetFullSchema retrieves the complete database schema
	GetFullSchema(ctx context.Context) (dto.Schema, error)

	// DescribeTable returns information about a specific table
	DescribeTable(ctx context.Context, tableName string) ([]dto.Column, error)

	// ExecuteQuery executes a SQL query and returns the results
	ExecuteQuery(ctx context.Context, sql string) (dto.QueryResult, error)

	// ExecuteUpdate executes a SQL update/insert/delete and returns affected rows
	ExecuteUpdate(ctx context.Context, sql string) (int64, error)

	// BeginTransaction begins a new transaction
	BeginTransaction(ctx context.Context) (*sqlx.Tx, error)

	// GetVersion returns the PostgreSQL version
	GetVersion(ctx context.Context) (string, error)

	// GetDatabase returns the current database name
	GetDatabase(ctx context.Context) (string, error)

	// GetUser returns the current user
	GetUser(ctx context.Context) (string, error)

	// ExplainQuery returns the query execution plan
	ExplainQuery(ctx context.Context, sql string) (dto.ExplainResult, error)

	// TableExists checks if a table exists
	TableExists(ctx context.Context, tableName string) (bool, error)
}
