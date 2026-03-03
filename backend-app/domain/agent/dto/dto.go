package dto

// ConnectionConfig holds database connection configuration
type ConnectionConfig struct {
	URI      string
	Host     string
	Port     string
	Database string
	User     string
	Password string
	SSLMode  string
}

// ConnectionInfo holds active connection information
type ConnectionInfo struct {
	Database string
	User     string
	Host     string
	Port     string
	Version  string
}

// QueryResult represents the result of a SQL query
type QueryResult struct {
	Columns []string
	Rows    [][]string
	RowCount int
}

// Table represents a database table
type Table struct {
	Name    string
	Schema  string
	Comment string
}

// Column represents a table column
type Column struct {
	Name         string
	Type         string
	Nullable     bool
	DefaultValue string
	IsPrimaryKey bool
	IsForeignKey bool
	References   string
}

// Schema represents the database schema
type Schema struct {
	Tables  []Table
	Columns map[string][]Column
}

// Command represents a REPL command
type Command struct {
	Name   string
	Args   []string
	Raw    string
}

// QueryRequest represents a SQL query request
type QueryRequest struct {
	SQL      string
	Destructive bool
}

// QueryResponse represents a SQL query response
type QueryResponse struct {
	Result   *QueryResult
	Error    string
	Affected int64
}

// ExplainResult represents the result of EXPLAIN
type ExplainResult struct {
	Plan string
}

// HistoryEntry represents a command history entry
type HistoryEntry struct {
	Input    string
	Timestamp int64
	Result   *QueryResponse
}
