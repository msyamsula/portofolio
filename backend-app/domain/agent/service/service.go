package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/msyamsula/portofolio/backend-app/domain/agent/dto"
	"github.com/msyamsula/portofolio/backend-app/domain/agent/repository"
	postgresInfra "github.com/msyamsula/portofolio/backend-app/infrastructure/database/postgres"
	llmdto "github.com/msyamsula/portofolio/backend-app/infrastructure/llm/dto"
	llmservice "github.com/msyamsula/portofolio/backend-app/infrastructure/llm/service"
	"github.com/msyamsula/portofolio/backend-app/infrastructure/repl/readline"
	"github.com/msyamsula/portofolio/backend-app/pkg/parser"
)

// agentService implements the Service interface
type agentService struct {
	repo         repository.Repository
	llm          llmservice.Service
	rl           *readline.Readline
	tx           *sqlx.Tx
	schema       string
	schemaLoaded bool
}

// New creates a new agent service
func New(repo repository.Repository) Service {
	return &agentService{
		repo: repo,
	}
}

// Connect establishes a database connection
func (s *agentService) Connect(ctx context.Context, cfg dto.ConnectionConfig) (*sqlx.DB, error) {
	var postgresCfg postgresInfra.Config

	if cfg.URI != "" {
		postgresCfg = postgresInfra.Config{
			URI: cfg.URI,
		}
	} else {
		postgresCfg = postgresInfra.Config{
			Host:     cfg.Host,
			Port:     cfg.Port,
			Database: cfg.Database,
			User:     cfg.User,
			Password: cfg.Password,
			SSLMode:  cfg.SSLMode,
		}
	}

	db, err := postgresInfra.NewClient(ctx, postgresCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Test connection
	if err := postgresInfra.TestConnection(ctx, db); err != nil {
		db.Close()
		return nil, fmt.Errorf("connection test failed: %w", err)
	}

	// Create new repository with this connection
	s.repo = repository.NewPostgresRepository(db)

	return db, nil
}

// ProcessInput processes user input and returns the response
func (s *agentService) ProcessInput(ctx context.Context, input string) (dto.QueryResponse, error) {
	result := s.ParseInput(input)

	switch result.Type {
	case parser.TypeCommand:
		return s.handleCommand(ctx, result.Command)

	case parser.TypeSQL:
		return s.ExecuteSQL(ctx, result.SQL)

	case parser.TypeNaturalLanguage:
		sql, err := s.GenerateSQL(ctx, result.NaturalLang)
		if err != nil {
			return dto.QueryResponse{Error: fmt.Sprintf("Failed to generate SQL: %v", err)}, nil
		}
		return s.ExecuteSQL(ctx, sql)

	case parser.TypeMixed:
		// For mixed input, use the natural language part
		sql, err := s.GenerateSQL(ctx, result.NaturalLang)
		if err != nil {
			return dto.QueryResponse{Error: fmt.Sprintf("Failed to generate SQL: %v", err)}, nil
		}
		return s.ExecuteSQL(ctx, sql)

	default:
		return dto.QueryResponse{Error: "Unknown input type"}, nil
	}
}

// GenerateSQL generates SQL from natural language
func (s *agentService) GenerateSQL(ctx context.Context, naturalLang string) (string, error) {
	if s.llm == nil || s.llm.GetAPIKey() == "" {
		return "", fmt.Errorf("LLM service not initialized. Please set an OpenAI API key")
	}

	schema, err := s.GetSchemaForPrompt(ctx)
	if err != nil {
		// Continue without schema
		schema = ""
	}

	req := llmdto.SQLRequest{
		NaturalLanguage: naturalLang,
		Schema:          schema,
	}

	resp, err := s.llm.GenerateSQL(ctx, req)
	if err != nil {
		return "", fmt.Errorf("LLM generation failed: %w", err)
	}

	return resp.SQL, nil
}

// ExecuteSQL executes a SQL query
func (s *agentService) ExecuteSQL(ctx context.Context, sql string) (dto.QueryResponse, error) {
	sql = strings.TrimSpace(sql)

	// Check for transaction commands
	if isCmd, cmd := parser.IsTransactionCommand(sql); isCmd {
		return s.handleTransactionCommand(ctx, cmd)
	}

	// Use transaction if active
	if s.tx != nil {
		return s.executeInTransaction(ctx, sql)
	}

	// Check if it's a SELECT or other query
	upper := strings.ToUpper(sql)
	if strings.HasPrefix(upper, "SELECT") ||
		strings.HasPrefix(upper, "SHOW") ||
		strings.HasPrefix(upper, "EXPLAIN") ||
		strings.HasPrefix(upper, "DESCRIBE") ||
		strings.HasPrefix(upper, "\\") {
		result, err := s.repo.ExecuteQuery(ctx, sql)
		if err != nil {
			return dto.QueryResponse{Error: fmt.Sprintf("Query error: %v", err)}, nil
		}
		return dto.QueryResponse{Result: &result}, nil
	}

	// It's an UPDATE/INSERT/DELETE
	affected, err := s.repo.ExecuteUpdate(ctx, sql)
	if err != nil {
		return dto.QueryResponse{Error: fmt.Sprintf("Update error: %v", err)}, nil
	}

	return dto.QueryResponse{Affected: affected}, nil
}

// executeInTransaction executes a query within the active transaction
func (s *agentService) executeInTransaction(ctx context.Context, sql string) (dto.QueryResponse, error) {
	upper := strings.ToUpper(sql)

	if strings.HasPrefix(upper, "SELECT") ||
		strings.HasPrefix(upper, "SHOW") ||
		strings.HasPrefix(upper, "EXPLAIN") ||
		strings.HasPrefix(upper, "DESCRIBE") ||
		strings.HasPrefix(upper, "\\") {
		// Execute query using transaction
		result, err := s.repo.ExecuteQuery(ctx, sql)
		if err != nil {
			return dto.QueryResponse{Error: fmt.Sprintf("Query error: %v", err)}, nil
		}
		return dto.QueryResponse{Result: &result}, nil
	}

	// Execute update using transaction
	affected, err := s.repo.ExecuteUpdate(ctx, sql)
	if err != nil {
		return dto.QueryResponse{Error: fmt.Sprintf("Update error: %v", err)}, nil
	}

	return dto.QueryResponse{Affected: affected}, nil
}

// handleTransactionCommand handles BEGIN/COMMIT/ROLLBACK commands
func (s *agentService) handleTransactionCommand(ctx context.Context, cmd string) (dto.QueryResponse, error) {
	switch cmd {
	case "BEGIN":
		if s.tx != nil {
			return dto.QueryResponse{Error: "Already in a transaction"}, nil
		}
		err := s.BeginTransaction(ctx)
		if err != nil {
			return dto.QueryResponse{Error: fmt.Sprintf("Transaction error: %v", err)}, nil
		}
		return dto.QueryResponse{}, nil
	case "COMMIT":
		err := s.CommitTransaction(ctx)
		if err != nil {
			return dto.QueryResponse{Error: fmt.Sprintf("Commit error: %v", err)}, nil
		}
		return dto.QueryResponse{}, nil
	case "ROLLBACK":
		err := s.RollbackTransaction(ctx)
		if err != nil {
			return dto.QueryResponse{Error: fmt.Sprintf("Rollback error: %v", err)}, nil
		}
		return dto.QueryResponse{}, nil
	default:
		return dto.QueryResponse{Error: fmt.Sprintf("Unknown transaction command: %s", cmd)}, nil
	}
}

// handleCommand handles special REPL commands
func (s *agentService) handleCommand(ctx context.Context, cmd string) (dto.QueryResponse, error) {
	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		return dto.QueryResponse{}, nil
	}

	command := strings.ToLower(parts[0])
	args := parts[1:]

	switch command {
	case "exit":
		return dto.QueryResponse{Result: &dto.QueryResult{Columns: []string{}, Rows: [][]string{{"exit"}}}}, nil

	case "help":
		// Help is handled by the handler
		return dto.QueryResponse{}, nil

	case "tables":
		tables, err := s.GetTables(ctx)
		if err != nil {
			return dto.QueryResponse{Error: fmt.Sprintf("Error: %v", err)}, nil
		}
		rows := make([][]string, len(tables))
		for i, t := range tables {
			rows[i] = []string{t.Name, t.Schema}
		}
		return dto.QueryResponse{Result: &dto.QueryResult{
			Columns:  []string{"Table", "Schema"},
			Rows:     rows,
			RowCount: len(tables),
		}}, nil

	case "schema":
		schema, err := s.GetFullSchema(ctx)
		if err != nil {
			return dto.QueryResponse{Error: fmt.Sprintf("Error: %v", err)}, nil
		}
		// Format as table
		rows := make([][]string, 0)
		for _, table := range schema.Tables {
			for _, col := range schema.Columns[table.Name] {
				rows = append(rows, []string{table.Name, col.Name, col.Type, fmt.Sprintf("%t", col.Nullable)})
			}
		}
		return dto.QueryResponse{Result: &dto.QueryResult{
			Columns:  []string{"Table", "Column", "Type", "Nullable"},
			Rows:     rows,
			RowCount: len(rows),
		}}, nil

	case "desc":
		if len(args) == 0 {
			return dto.QueryResponse{Error: "Usage: .desc <table_name>"}, nil
		}
		cols, err := s.DescribeTable(ctx, args[0])
		if err != nil {
			return dto.QueryResponse{Error: fmt.Sprintf("Error: %v", err)}, nil
		}
		rows := make([][]string, len(cols))
		for i, c := range cols {
			rows[i] = []string{c.Name, c.Type, fmt.Sprintf("%t", c.Nullable), fmt.Sprintf("%t", c.IsPrimaryKey)}
		}
		return dto.QueryResponse{Result: &dto.QueryResult{
			Columns:  []string{"Column", "Type", "Nullable", "PK"},
			Rows:     rows,
			RowCount: len(cols),
		}}, nil

	case "explain":
		if len(args) == 0 {
			return dto.QueryResponse{Error: "Usage: .explain <sql>"}, nil
		}
		sql := strings.Join(args, " ")
		result, err := s.ExplainQuery(ctx, sql)
		if err != nil {
			return dto.QueryResponse{Error: fmt.Sprintf("Error: %v", err)}, nil
		}
		return dto.QueryResponse{Result: &dto.QueryResult{
			Columns:  []string{"Plan"},
			Rows:     [][]string{{result.Plan}},
			RowCount: 1,
		}}, nil

	case "history":
		if s.rl != nil {
			history := s.rl.History()
			rows := make([][]string, len(history))
			for i, h := range history {
				rows[i] = []string{fmt.Sprintf("%d", i+1), h}
			}
			return dto.QueryResponse{Result: &dto.QueryResult{
				Columns:  []string{"#", "Command"},
				Rows:     rows,
				RowCount: len(history),
			}}, nil
		}
		return dto.QueryResponse{}, nil

	case "db":
		info, err := s.GetConnectionInfo(ctx)
		if err != nil {
			return dto.QueryResponse{Error: fmt.Sprintf("Error: %v", err)}, nil
		}
		return dto.QueryResponse{Result: &dto.QueryResult{
			Columns:  []string{"Key", "Value"},
			Rows:     [][]string{{"Database", info.Database}},
			RowCount: 1,
		}}, nil

	case "user":
		info, err := s.GetConnectionInfo(ctx)
		if err != nil {
			return dto.QueryResponse{Error: fmt.Sprintf("Error: %v", err)}, nil
		}
		return dto.QueryResponse{Result: &dto.QueryResult{
			Columns:  []string{"Key", "Value"},
			Rows:     [][]string{{"User", info.User}},
			RowCount: 1,
		}}, nil

	case "ver":
		info, err := s.GetConnectionInfo(ctx)
		if err != nil {
			return dto.QueryResponse{Error: fmt.Sprintf("Error: %v", err)}, nil
		}
		return dto.QueryResponse{Result: &dto.QueryResult{
			Columns:  []string{"Version"},
			Rows:     [][]string{{info.Version}},
			RowCount: 1,
		}}, nil

	default:
		return dto.QueryResponse{Error: fmt.Sprintf("Unknown command: .%s. Type .help for available commands.", command)}, nil
	}
}

// GetConnectionInfo retrieves connection information
func (s *agentService) GetConnectionInfo(ctx context.Context) (dto.ConnectionInfo, error) {
	return s.repo.GetConnectionInfo(ctx)
}

// GetTables retrieves all tables
func (s *agentService) GetTables(ctx context.Context) ([]dto.Table, error) {
	return s.repo.GetTables(ctx)
}

// DescribeTable describes a table structure
func (s *agentService) DescribeTable(ctx context.Context, tableName string) ([]dto.Column, error) {
	return s.repo.DescribeTable(ctx, tableName)
}

// GetFullSchema retrieves the complete schema
func (s *agentService) GetFullSchema(ctx context.Context) (dto.Schema, error) {
	return s.repo.GetFullSchema(ctx)
}

// ExplainQuery explains a query execution plan
func (s *agentService) ExplainQuery(ctx context.Context, sql string) (dto.ExplainResult, error) {
	return s.repo.ExplainQuery(ctx, sql)
}

// BeginTransaction begins a new transaction
func (s *agentService) BeginTransaction(ctx context.Context) error {
	tx, err := s.repo.BeginTransaction(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	s.tx = tx
	return nil
}

// CommitTransaction commits the current transaction
func (s *agentService) CommitTransaction(ctx context.Context) error {
	if s.tx == nil {
		return fmt.Errorf("no active transaction")
	}
	err := s.tx.Commit()
	s.tx = nil
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}

// RollbackTransaction rolls back the current transaction
func (s *agentService) RollbackTransaction(ctx context.Context) error {
	if s.tx == nil {
		return fmt.Errorf("no active transaction")
	}
	err := s.tx.Rollback()
	s.tx = nil
	if err != nil {
		return fmt.Errorf("failed to rollback transaction: %w", err)
	}
	return nil
}

// InTransaction returns true if currently in a transaction
func (s *agentService) InTransaction() bool {
	return s.tx != nil
}

// SetLLMService sets the LLM service
func (s *agentService) SetLLMService(llm llmservice.Service) {
	s.llm = llm
}

// SetReadline sets the readline instance
func (s *agentService) SetReadline(rl *readline.Readline) {
	s.rl = rl
}

// GetSchemaForPrompt returns the schema formatted for LLM prompting
func (s *agentService) GetSchemaForPrompt(ctx context.Context) (string, error) {
	if s.schemaLoaded {
		return s.schema, nil
	}

	schema, err := s.GetFullSchema(ctx)
	if err != nil {
		return "", err
	}

	// Format schema for LLM
	var builder strings.Builder
	for _, table := range schema.Tables {
		builder.WriteString(fmt.Sprintf("Table %s:\n", table.Name))
		for _, col := range schema.Columns[table.Name] {
			nullStr := ""
			if col.Nullable {
				nullStr = " NULL"
			} else {
				nullStr = " NOT NULL"
			}
			pkStr := ""
			if col.IsPrimaryKey {
				pkStr = " PRIMARY KEY"
			}
			builder.WriteString(fmt.Sprintf("  %s %s%s%s\n", col.Name, col.Type, nullStr, pkStr))
		}
	}

	s.schema = builder.String()
	s.schemaLoaded = true
	return s.schema, nil
}

// ParseInput parses user input and determines type
func (s *agentService) ParseInput(input string) parser.Result {
	return parser.Parse(input)
}

// IsDestructive checks if SQL is destructive
func (s *agentService) IsDestructive(sql string) bool {
	return parser.IsDestructiveOperation(sql)
}
