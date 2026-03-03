package handler

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"

	"github.com/msyamsula/portofolio/backend-app/domain/agent/dto"
	agentservice "github.com/msyamsula/portofolio/backend-app/domain/agent/service"
	"github.com/msyamsula/portofolio/backend-app/infrastructure/repl/readline"
	"github.com/msyamsula/portofolio/backend-app/infrastructure/repl/ui"
	"github.com/msyamsula/portofolio/backend-app/pkg/parser"
	postgresInfra "github.com/msyamsula/portofolio/backend-app/infrastructure/database/postgres"
	infraLogger "github.com/msyamsula/portofolio/backend-app/infrastructure/telemetry/logger"
)

// Handler handles REPL interactions
type Handler struct {
	service agentservice.Service
	rl      *readline.Readline
	ui      *ui.Formatter
	prompt  *ui.Prompt
	running bool
}

// New creates a new handler
func New(service agentservice.Service) *Handler {
	return &Handler{
		service: service,
		ui:      ui.NewFormatter(),
		prompt:  ui.NewPrompt(),
	}
}

// Start starts REPL
func (h *Handler) Start(ctx context.Context) error {
	// Initialize readline
	rl, err := rlpkg.New(rlpkg.Config{
		Prompt:      h.prompt.GetPromptWithInfo(),
		HistoryFile: os.ExpandEnv("$HOME/.pg-agent-history"),
	})
	if err != nil {
		return fmt.Errorf("failed to initialize readline: %w", err)
	}
	defer rl.Close()

	h.rl = rl
	h.service.SetReadline(rl)

	// Print welcome message
	rl.Println(h.ui.FormatWelcome())

	// REPL loop
	h.running = true
	for h.running {
		if err := h.handleInput(ctx); err != nil {
			rl.Printf("%s\n", h.ui.FormatError(fmt.Sprintf("Error: %v", err)))
		}
	}

	return nil
}

// PromptForCredentials prompts user for database credentials
func (h *Handler) PromptForCredentials() (dto.ConnectionConfig, error) {
	var cfg dto.ConnectionConfig
	var input string

	fmt.Print("Enter PostgreSQL connection information:\n")

	// Check if user wants to use a connection URI
	fmt.Print("PostgreSQL URI (leave blank for individual fields): ")
	uri, _ := simplePrompt("")

	if uri != "" {
		cfg.URI = uri
		return cfg, nil
	}

	// Prompt for individual fields
	cfg.Host = h.promptWithDefault("Host", "localhost")
	cfg.Port = h.promptWithDefault("Port", "5432")
	cfg.Database = h.promptRequired("Database Name")
	cfg.User = h.promptRequired("Username")
	cfg.Password = h.promptPassword()

	cfg.SSLMode = h.promptWithDefault("SSL Mode", "disable")

	return cfg, nil
}

// PromptForAPIKey prompts user for OpenAI API key
func (h *Handler) PromptForAPIKey() string {
	fmt.Print("\nEnter OpenAI API key (leave blank to skip LLM features): ")
	key, _ := simplePrompt("")
	return strings.TrimSpace(key)
}

// handleInput handles a single input
func (h *Handler) handleInput(ctx context.Context) error {
	// Update prompt
	h.rl.SetPrompt(h.prompt.GetPromptWithInfo())

	// Read input
	input, err := h.rl.Read()
	if err != nil {
		if err == rlpkg.ErrInterrupt {
			h.rl.Println("^C")
			return nil
		}
		if err == rlpkg.ErrEOF {
			h.running = false
			return nil
		}
		return fmt.Errorf("read error: %w", err)
	}

	input = strings.TrimSpace(input)
	if input == "" {
		return nil
	}

	// Check for exit command
	if input == ".exit" || input == "exit" || input == "quit" || input == "\\q" {
		h.running = false
		return nil
	}

	// Check for help command
	if input == ".help" || input == "help" || input == "\\?" {
		h.showHelp()
		return nil
	}

	// Parse input type
	result := h.service.ParseInput(input)

	switch result.Type {
	case parser.TypeCommand:
		// Command needs special handling
		return h.handleCommand(ctx, input, result)

	case parser.TypeSQL:
		return h.handleSQL(ctx, input)

	case parser.TypeNaturalLanguage:
		return h.handleNaturalLanguage(ctx, input)

	case parser.TypeMixed:
		return h.handleNaturalLanguage(ctx, input)

	default:
		h.rl.Printf("%s\n", h.ui.FormatError("Unable to parse input. Type .help for available commands."))
		return nil
	}
}

// handleCommand handles special commands
func (h *Handler) handleCommand(ctx context.Context, input string, result parser.Result) error {
	cmd := strings.TrimPrefix(input, ".")
	cmd = strings.ToLower(cmd)
	args := strings.TrimSpace(strings.TrimPrefix(input, "."+cmd+" "))

	switch cmd {
	case "exit":
		h.running = false
		h.rl.Println("Goodbye!")
		return nil

	case "help":
		h.showHelp()
		return nil

	case "tables":
		return h.handleTablesCommand(ctx)

	case "schema":
		return h.handleSchemaCommand(ctx)

	case "desc":
		return h.handleDescCommand(ctx, args)

	case "explain":
		return h.handleExplainCommand(ctx, args)

	case "history":
		return h.handleHistoryCommand(ctx)

	case "db":
		return h.handleDBCommand(ctx)

	case "user":
		return h.handleUserCommand(ctx)

	case "ver":
		return h.handleVersionCommand(ctx)

	default:
		h.rl.Printf("%s\n", h.ui.FormatError(fmt.Sprintf("Unknown command: .%s. Type .help for available commands.", cmd)))
		return nil
}

// handleSQL handles SQL input
func (h *Handler) handleSQL(ctx context.Context, input string) error {
	// Preview SQL
	h.rl.Printf("%s\n", h.ui.FormatQueryPreview(input))

	// Check if it's an EXPLAIN command
	if strings.HasPrefix(strings.ToUpper(input), "EXPLAIN") {
		// For EXPLAIN, execute directly without confirmation
		response, err := h.service.ExecuteSQL(ctx, input)
		if err != nil {
			return err
		}
		h.displayResponse(response)
}

	// Check if destructive
	if h.service.IsDestructive(input) {
		h.rl.Println(h.ui.FormatDestructiveWarning("operation"))

		// Confirm
		if !h.confirm("Execute this destructive operation") {
			h.rl.Println("Operation cancelled.")
			return nil
		}
	}

	// Execute
	response, err := h.service.ExecuteSQL(ctx, input)
	if err != nil {
		return err
	}

	// Display result
	return h.displayResponse(response)
}

// handleNaturalLanguage handles natural language input
func (h *Handler) handleNaturalLanguage(ctx context.Context, input string) error {
	h.rl.Printf("%s\n", h.ui.FormatInfo(fmt.Sprintf("Processing: %s", input)))

	// Generate SQL
	sql, err := h.service.GenerateSQL(ctx, input)
	if err != nil {
		h.rl.Printf("%s\n", h.ui.FormatError(fmt.Sprintf("Failed to generate SQL: %v", err)))
		return nil
	}

	// Preview SQL
	h.rl.Printf("%s\n", h.ui.FormatQueryPreview(sql))

	// Confirm
	if !h.confirm("Execute this query") {
		h.rl.Println("Operation cancelled.")
		return nil
	}

	// Execute
	response, err := h.service.ExecuteSQL(ctx, sql)
	if err != nil {
		return err
	}

	// Display result
	return h.displayResponse(response)
}

// handleTablesCommand handles the .tables command
func (h *Handler) handleTablesCommand(ctx context.Context) error {
	tables, err := h.service.GetTables(ctx)
	if err != nil {
		return err
	}

	if len(tables) == 0 {
		h.rl.Println(h.ui.FormatInfo("(no tables found)"))
		return nil
	}

	rows := make([][]string, len(tables))
	for i, t := range tables {
		rows[i] = []string{t.Name, t.Schema}
	}

	response := dto.QueryResponse{
		Result: &dto.QueryResult{
			Columns:  []string{"Table", "Schema"},
			Rows:    rows,
			RowCount: len(tables),
		},
	}

	return h.displayResponse(response)
}

// handleSchemaCommand handles the .schema command
func (h *Handler) handleSchemaCommand(ctx context.Context) error {
	schema, err := h.service.GetFullSchema(ctx)
	if err != nil {
		return err
	}

	if len(schema.Tables) == 0 {
		h.rl.Println(h.ui.FormatInfo("(no tables found)"))
		return nil
	}

	// Format schema for display
	rows := make([][]string, 0)
	for _, table := range schema.Tables {
		for _, col := range schema.Columns[table.Name] {
			rows = append(rows, []string{
				table.Name,
				col.Name,
				col.Type,
				fmt.Sprintf("%t", col.Nullable),
			})
		}
	}

	response := dto.QueryResponse{
		Result: &dto.QueryResult{
			Columns:  []string{"Table", "Column", "Type", "Nullable"},
			Rows:    rows,
			RowCount: len(rows),
		},
	}

	return h.displayResponse(response)
}

// handleDescCommand handles the .desc command
func (h *Handler) handleDescCommand(ctx context.Context, args string) error {
	if args == "" {
		h.rl.Println(h.ui.FormatError("Usage: .desc <table_name>"))
		return nil
	}

	cols, err := h.service.DescribeTable(ctx, args)
	if err != nil {
		return err
	}

	if len(cols) == 0 {
		h.rl.Printf("%s\n", h.ui.FormatError(fmt.Sprintf("Table %s not found", args)))
		return nil
	}

	rows := make([][]string, len(cols))
	for i, c := range cols {
		rows[i] = []string{
			c.Name,
			c.Type,
			fmt.Sprintf("%t", c.Nullable),
			fmt.Sprintf("%t", c.IsPrimaryKey),
		}
	}

	response := dto.QueryResponse{
		Result: &dto.QueryResult{
			Columns:  []string{"Column", "Type", "Nullable", "PK"},
			Rows:    rows,
			RowCount: len(cols),
		},
	}

	return h.displayResponse(response)
}

// handleExplainCommand handles the .explain command
func (h *Handler) handleExplainCommand(ctx context.Context, args string) error {
	if args == "" {
		h.rl.Println(h.ui.FormatError("Usage: .explain <sql>"))
		return nil
	}

	// Parse SQL from args
	sql := strings.Join(strings.Fields(args), " ")

	result, err := h.service.ExplainQuery(ctx, sql)
	if err != nil {
		return err
	}

	response := dto.QueryResponse{
		Result: &dto.QueryResult{
			Columns:  []string{"Plan"},
			Rows:    [][]string{{result.Plan}},
			RowCount: 1,
		},
	}

	return h.displayResponse(response)
}

// handleHistoryCommand handles the .history command
func (h *Handler) handleHistoryCommand(ctx context.Context) error {
	history := h.rl.History()
	if len(history) == 0 {
		h.rl.Println(h.ui.FormatInfo("(no history)"))
		return nil
	}

	rows := make([][]string, len(history))
	for i, h := range history {
		rows[i] = []string{fmt.Sprintf("%d", i+1), h}
	}

	response := dto.QueryResponse{
		Result: &dto.QueryResult{
			Columns:  []string{"#", "Command"},
			Rows:    rows,
			RowCount: len(history),
		},
	}

	return h.displayResponse(response)
}

// handleDBCommand handles the .db command
func (h *Handler) handleDBCommand(ctx context.Context) error {
	info, err := h.service.GetConnectionInfo(ctx)
	if err != nil {
		return err
	}

	response := dto.QueryResponse{
		Result: &dto.QueryResult{
			Columns:  []string{"Key", "Value"},
			Rows:    [][]string{{"Database", info.Database}},
			RowCount: 1,
		},
	}

	return h.displayResponse(response)
}

// handleUserCommand handles the .user command
func (h *Handler) handleUserCommand(ctx context.Context) error {
	info, err := h.service.GetConnectionInfo(ctx)
	if err != nil {
		return err
	}

	response := dto.QueryResponse{
		Result: &dto.QueryResult{
			Columns:  []string{"Key", "Value"},
			Rows:    [][]string{{"User", info.User}},
			RowCount: 1,
		},
	}

	return h.displayResponse(response)
}

// handleVersionCommand handles the .ver command
func (h *Handler) handleVersionCommand(ctx context.Context) error {
	info, err := h.service.GetConnectionInfo(ctx)
	if err != nil {
		return err
	}

	response := dto.QueryResponse{
		Result: &dto.QueryResult{
			Columns:  []string{"Version"},
			Rows:    [][]string{{info.Version}},
			RowCount: 1,
		},
	}

	return h.displayResponse(response)
}

// displayResponse displays a query response
func (h *Handler) displayResponse(response dto.QueryResponse) error {
	if response.Error != "" {
		h.rl.Printf("%s\n", h.ui.FormatError(response.Error))
		return nil
	}

	if response.Result != nil {
		if len(response.Result.Rows) == 0 {
			h.rl.Println(h.ui.FormatInfo("(no results)"))
		} else {
			h.rl.Println(h.ui.FormatTable(response.Result.Columns, response.Result.Rows))
			h.rl.Println(h.ui.FormatRowCount(response.Result.RowCount))
		}
	}

	if response.Affected > 0 {
		h.rl.Printf("%s %d row%s affected\n", h.ui.FormatSuccess("Success:"), response.Affected, pluralize(response.Affected))
	}

	return nil
}

// confirm prompts for confirmation
func (h *Handler) confirm(message string) bool {
	prompt := h.ui.FormatConfirmation(message)
	h.rl.Printf("%s", prompt)

	response, err := h.rl.Read()
	if err != nil {
		return false
	}

	return strings.ToLower(response) == "y" || strings.ToLower(response) == "yes"
}

// showHelp displays help information
func (h *Handler) showHelp() {
	help := `
Available Commands:

SQL Queries:
  SELECT * FROM table_name     Execute a SELECT query
  INSERT INTO ...             Execute an INSERT query
  UPDATE ...                  Execute an UPDATE query
  DELETE FROM ...             Execute a DELETE query
  CREATE TABLE ...            Create a new table
  DROP TABLE table_name        Drop a table
  BEGIN                       Start a transaction
  COMMIT                      Commit current transaction
  ROLLBACK                    Rollback current transaction
  EXPLAIN <sql>              Show query execution plan

Natural Language:
  "show me all users"         Convert natural language to SQL
  "create table users..."      Generate table creation SQL
  "count how many users..."     Generate counting query

Special Commands:
  .help                       Show this help message
  .exit                       Exit REPL
  .tables                     List all tables
  .schema                     Show database schema
  .desc <table>               Describe table structure
  .explain <sql>              Show query execution plan
  .history                    Show command history
  .db                         Show current database
  .user                       Show current user
  .ver                        Show PostgreSQL version

Features:
  - Type SQL or natural language directly
  - All SQL is previewed before execution
  - Destructive operations require confirmation
  - Transaction support with BEGIN/COMMIT/ROLLBACK
  - Command history (up/down arrows)
  - Auto-completion (tab)
`
	h.rl.Println(help)
}

// promptWithDefault prompts with a default value
func (h *Handler) promptWithDefault(field, defaultValue string) string {
	prompt := fmt.Sprintf("%s [%s]: ", field, defaultValue)
	fmt.Print(prompt)
	response, _ := simplePrompt("")
	response = strings.TrimSpace(response)
	if response == "" {
		return defaultValue
	}
	return response
}

// promptRequired prompts for a required field
func (h *Handler) promptRequired(field string) string {
	for {
		fmt.Printf("%s: ", field)
		response, _ := simplePrompt("")
		response = strings.TrimSpace(response)
		if response != "" {
			return response
		}
		h.rl.Printf("%s\n", h.ui.FormatError(fmt.Sprintf("%s is required", field)))
	}
}

// promptPassword prompts for a password without echo
func (h *Handler) promptPassword() string {
	fmt.Print("Password: ")
	response, _ := simplePrompt("")
	return strings.TrimSpace(response)
}

// SetDatabase sets to current database name
func (h *Handler) SetDatabase(db string) {
	h.prompt.SetDatabase(db)
}

// SetUser sets to current user
func (h *Handler) SetUser(user string) {
	h.prompt.SetUser(user)
}

// SetTransaction sets to transaction state
func (h *Handler) SetTransaction(inTransaction bool) {
	h.prompt.SetTransaction(inTransaction)
}

// pluralize returns to plural form
func pluralize(count int64) string {
	if count == 1 {
		return ""
	}
	return "s"
}

// simplePrompt reads a line from stdin
func simplePrompt(prompt string) (string, error) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(prompt)
	line, err := reader.ReadString('\n')
	return strings.TrimSpace(line), err
}
