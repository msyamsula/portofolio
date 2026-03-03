package service

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/msyamsula/portofolio/backend-app/domain/agent/dto"
	llmservice "github.com/msyamsula/portofolio/backend-app/infrastructure/llm/service"
	"github.com/msyamsula/portofolio/backend-app/infrastructure/repl/readline"
	"github.com/msyamsula/portofolio/backend-app/pkg/parser"
)

// Service defines the interface for agent business logic
//
//go:generate mockgen -source=interface.go -destination=../../../mock/agent_service_mock.go -package=mock -mock_names Service=MockAgentService
type Service interface {
	// Connect establishes a database connection
	Connect(ctx context.Context, cfg dto.ConnectionConfig) (*sqlx.DB, error)

	// ProcessInput processes user input and returns the response
	ProcessInput(ctx context.Context, input string) (dto.QueryResponse, error)

	// GenerateSQL generates SQL from natural language
	GenerateSQL(ctx context.Context, naturalLang string) (string, error)

	// ExecuteSQL executes a SQL query
	ExecuteSQL(ctx context.Context, sql string) (dto.QueryResponse, error)

	// GetConnectionInfo retrieves connection information
	GetConnectionInfo(ctx context.Context) (dto.ConnectionInfo, error)

	// GetTables retrieves all tables
	GetTables(ctx context.Context) ([]dto.Table, error)

	// DescribeTable describes a table structure
	DescribeTable(ctx context.Context, tableName string) ([]dto.Column, error)

	// GetFullSchema retrieves the complete schema
	GetFullSchema(ctx context.Context) (dto.Schema, error)

	// ExplainQuery explains a query execution plan
	ExplainQuery(ctx context.Context, sql string) (dto.ExplainResult, error)

	// BeginTransaction begins a new transaction
	BeginTransaction(ctx context.Context) error

	// CommitTransaction commits the current transaction
	CommitTransaction(ctx context.Context) error

	// RollbackTransaction rolls back the current transaction
	RollbackTransaction(ctx context.Context) error

	// InTransaction returns true if currently in a transaction
	InTransaction() bool

	// SetLLMService sets the LLM service
	SetLLMService(llmservice.Service)

	// SetReadline sets the readline instance
	SetReadline(rl *readline.Readline)

	// GetSchemaForPrompt returns the schema formatted for LLM prompting
	GetSchemaForPrompt(ctx context.Context) (string, error)

	// ParseInput parses user input and determines type
	ParseInput(input string) parser.Result

	// IsDestructive checks if SQL is destructive
	IsDestructive(sql string) bool
}
