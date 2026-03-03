package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/msyamsula/portofolio/backend-app/domain/agent/dto"
)

// postgresRepository implements the Repository interface
type postgresRepository struct {
	db *sqlx.DB
}

// NewPostgresRepository creates a new PostgreSQL repository
func NewPostgresRepository(db *sqlx.DB) Repository {
	return &postgresRepository{db: db}
}

// GetConnectionInfo retrieves information about the current connection
func (r *postgresRepository) GetConnectionInfo(ctx context.Context) (dto.ConnectionInfo, error) {
	var info dto.ConnectionInfo

	// Get database name
	if err := r.db.GetContext(ctx, &info.Database, "SELECT current_database()"); err != nil {
		return dto.ConnectionInfo{}, fmt.Errorf("failed to get database: %w", err)
	}

	// Get current user
	if err := r.db.GetContext(ctx, &info.User, "SELECT current_user"); err != nil {
		return dto.ConnectionInfo{}, fmt.Errorf("failed to get user: %w", err)
	}

	// Get version
	if err := r.db.GetContext(ctx, &info.Version, "SELECT version()"); err != nil {
		return dto.ConnectionInfo{}, fmt.Errorf("failed to get version: %w", err)
	}

	return info, nil
}

// GetTables retrieves all tables in the current database
func (r *postgresRepository) GetTables(ctx context.Context) ([]dto.Table, error) {
	query := `
		SELECT
			table_name as name,
			table_schema as schema,
			obj_description((table_schema||'.'||table_name)::regclass) as comment
		FROM information_schema.tables
		WHERE table_schema = 'public'
		ORDER BY table_name
	`

	var tables []dto.Table
	err := r.db.SelectContext(ctx, &tables, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get tables: %w", err)
	}

	return tables, nil
}

// GetSchema retrieves the complete schema for a table
func (r *postgresRepository) GetSchema(ctx context.Context, tableName string) (dto.Column, error) {
	// This is a placeholder - use DescribeTable for full table schema
	return dto.Column{}, fmt.Errorf("use DescribeTable instead")
}

// GetFullSchema retrieves the complete database schema
func (r *postgresRepository) GetFullSchema(ctx context.Context) (dto.Schema, error) {
	tables, err := r.GetTables(ctx)
	if err != nil {
		return dto.Schema{}, err
	}

	columns := make(map[string][]dto.Column)
	for _, table := range tables {
		tableColumns, err := r.DescribeTable(ctx, table.Name)
		if err != nil {
			// Continue if we can't get columns for one table
			continue
		}
		columns[table.Name] = tableColumns
	}

	return dto.Schema{
		Tables:  tables,
		Columns: columns,
	}, nil
}

// DescribeTable returns information about a specific table
func (r *postgresRepository) DescribeTable(ctx context.Context, tableName string) ([]dto.Column, error) {
	query := `
		SELECT
			column_name as name,
			data_type as type,
			is_nullable as nullable,
			column_default as default_value,
			COALESCE(
				pg_get_serial_sequence(table_schema||'.'||table_name, column_name) IS NOT NULL,
				EXISTS (
					SELECT 1 FROM information_schema.table_constraints tc
					JOIN information_schema.key_column_usage kcu
					ON tc.constraint_name = kcu.constraint_name
					WHERE tc.constraint_type = 'PRIMARY KEY'
					AND tc.table_name = information_schema.columns.table_name
					AND kcu.column_name = information_schema.columns.column_name
				)
			) as is_primary_key
		FROM information_schema.columns
		WHERE table_name = $1
		ORDER BY ordinal_position
	`

	var columns []dto.Column
	err := r.db.SelectContext(ctx, &columns, query, tableName)
	if err != nil {
		return nil, fmt.Errorf("failed to describe table %s: %w", tableName, err)
	}

	return columns, nil
}

// ExecuteQuery executes a SQL query and returns the results
func (r *postgresRepository) ExecuteQuery(ctx context.Context, sql string) (dto.QueryResult, error) {
	rows, err := r.db.QueryxContext(ctx, sql)
	if err != nil {
		return dto.QueryResult{}, fmt.Errorf("query execution failed: %w", err)
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return dto.QueryResult{}, fmt.Errorf("failed to get columns: %w", err)
	}

	var resultRows [][]string
	for rows.Next() {
		// Create a slice of interface{} to hold each column value
		values := make([]any, len(columns))
		valuePtrs := make([]any, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return dto.QueryResult{}, fmt.Errorf("failed to scan row: %w", err)
		}

		// Convert values to strings
		row := make([]string, len(columns))
		for i, val := range values {
			if val == nil {
				row[i] = "NULL"
			} else {
				row[i] = fmt.Sprintf("%v", val)
			}
		}
		resultRows = append(resultRows, row)
	}

	return dto.QueryResult{
		Columns: columns,
		Rows:    resultRows,
		RowCount: len(resultRows),
	}, nil
}

// ExecuteUpdate executes a SQL update/insert/delete and returns affected rows
func (r *postgresRepository) ExecuteUpdate(ctx context.Context, sql string) (int64, error) {
	result, err := r.db.ExecContext(ctx, sql)
	if err != nil {
		return 0, fmt.Errorf("update execution failed: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get affected rows: %w", err)
	}

	return affected, nil
}

// BeginTransaction begins a new transaction
func (r *postgresRepository) BeginTransaction(ctx context.Context) (*sqlx.Tx, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	return tx, nil
}

// GetVersion returns the PostgreSQL version
func (r *postgresRepository) GetVersion(ctx context.Context) (string, error) {
	var version string
	err := r.db.GetContext(ctx, &version, "SELECT version()")
	if err != nil {
		return "", fmt.Errorf("failed to get version: %w", err)
	}
	return version, nil
}

// GetDatabase returns the current database name
func (r *postgresRepository) GetDatabase(ctx context.Context) (string, error) {
	var database string
	err := r.db.GetContext(ctx, &database, "SELECT current_database()")
	if err != nil {
		return "", fmt.Errorf("failed to get database: %w", err)
	}
	return database, nil
}

// GetUser returns the current user
func (r *postgresRepository) GetUser(ctx context.Context) (string, error) {
	var user string
	err := r.db.GetContext(ctx, &user, "SELECT current_user")
	if err != nil {
		return "", fmt.Errorf("failed to get user: %w", err)
	}
	return user, nil
}

// ExplainQuery returns the query execution plan
func (r *postgresRepository) ExplainQuery(ctx context.Context, sql string) (dto.ExplainResult, error) {
	explainSQL := "EXPLAIN " + sql

	rows, err := r.db.QueryContext(ctx, explainSQL)
	if err != nil {
		return dto.ExplainResult{}, fmt.Errorf("failed to explain query: %w", err)
	}
	defer rows.Close()

	var planLines []string
	for rows.Next() {
		var line string
		if err := rows.Scan(&line); err != nil {
			return dto.ExplainResult{}, fmt.Errorf("failed to scan explain output: %w", err)
		}
		planLines = append(planLines, line)
	}

	var plan strings.Builder
	for i, line := range planLines {
		if i > 0 {
			plan.WriteString("\n")
		}
		plan.WriteString(line)
	}

	return dto.ExplainResult{Plan: plan.String()}, nil
}

// TableExists checks if a table exists
func (r *postgresRepository) TableExists(ctx context.Context, tableName string) (bool, error) {
	var exists bool
	query := `
		SELECT EXISTS (
			SELECT 1 FROM information_schema.tables
			WHERE table_schema = 'public' AND table_name = $1
		)
	`
	err := r.db.GetContext(ctx, &exists, query, tableName)
	if err != nil {
		return false, fmt.Errorf("failed to check table existence: %w", err)
	}
	return exists, nil
}
