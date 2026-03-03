# PostgreSQL CLI Agent - Implementation Plan

## Context

Create an interactive CLI REPL agent for PostgreSQL that interprets mixed input types (human instructions, raw SQL, or a combination) and executes queries against a database. The agent uses OpenAI's LLM for natural language to SQL conversion while maintaining the portfolio's existing patterns (hexagonal architecture, sqlx with lib/pq, testify for testing).

## Requirements Summary

1. **Interactive REPL** - Continuous command session with PostgreSQL
2. **Connection Setup** - Prompt on startup for PostgreSQL credentials (URI, username, password)
3. **Input Types** - Human high-level instructions, pasted SQL (not well formatted), or mixed mode
4. **Always Preview SQL** - Show generated SQL before execution
5. **Always Confirm Destructive Ops** - DROP, DELETE, TRUNCATE require confirmation
6. **Full Error Details** - Display complete PostgreSQL error messages
7. **LLM Integration** - Use OpenAI API for natural language to SQL conversion, prompt for API key on startup
8. **Location** - New domain `agent/` within `backend-app/`, reusing existing `infrastructure/` and `pkg/`

## Project Structure

```
backend-app/
├── binary/
│   ├── http/          # Existing HTTP server
│   └── pg-agent/      # New: CLI REPL entry point
│       └── main.go
├── domain/
│   ├── ...existing domains...
│   └── agent/         # New: PostgreSQL CLI agent domain
│       ├── dto/
│       │   └── dto.go
│       ├── handler/
│       │   └── handler.go      # REPL interaction handler
│       ├── service/
│       │   ├── interface.go
│       │   └── service.go      # All execution logic: connection, SQL, schema, transactions
│       └── repository/
│           └── repository.go   # Database queries (schema info, results)
├── infrastructure/
│   ├── database/
│   │   └── postgres/           # REUSE: Existing PostgreSQL client
│   ├── llm/
│   │   ├── dto/
│   │   │   └── dto.go
│   │   ├── service/
│   │   │   ├── interface.go
│   │   │   └── service.go      # LLM integration (OpenAI API, prompt engineering)
│   │   └── prompt/
│   │       └── prompt.go       # System prompts for SQL generation
│   ├── repl/
│   │   ├── readline/
│   │   │   ├── readline.go     # New: Readline wrapper
│   │   │   └── readline_test.go
│   │   └── ui/
│   │       ├── formatter.go   # New: Output formatting
│   │       └── prompt.go      # New: Prompt management
│   └── logger/
│       └── logger.go           # REUSE: Existing logger
└── pkg/
    ├── common/
    │   └── env.go              # REUSE: Existing env helpers
    └── parser/
        └── parser.go           # NEW: Input type detection (SQL vs natural language) - pure algorithmic logic
```

## Key Libraries

```
github.com/jmoiron/sqlx          # Database operations (already in portfolio)
github.com/lib/pq                # PostgreSQL driver (already in portfolio)
github.com/openai/openai-go      # LLM API (already in portfolio)
github.com/chzyer/readline       # REPL interface (new)
github.com/jedib0t/go-pretty/v6   # Table formatting (new)
github.com/fatih/color           # Color output (new)
github.com/stretchr/testify      # Testing (already in portfolio)
```

## Implementation Steps

### Step 1: Foundation
- Create project structure with go.mod
- Implement PostgreSQL client using sqlx (pattern: `backend-app/infrastructure/database/postgres/postgres.go`)
- Implement readline wrapper for REPL
- Create UI utilities (formatter, prompt management)

### Step 2: Input Parser (pkg)
- Implement parser to detect input type (SQL vs natural language)
- Detect SQL keywords and classify operations

### Step 3: Agent Service (domain/agent/service/)
- Implement connection management (prompt for credentials, test connection)
- Implement SQL execution (execute queries, format results)
- Implement schema inspection (get tables, columns, types)
- Implement transaction support (BEGIN, COMMIT, ROLLBACK)
- Detect destructive operations (DROP, DELETE, TRUNCATE)

### Step 4: REPL Interface
- Implement REPL loop with prompt state management
- Add special commands (.exit, .help, .schema, .tables, .desc, .explain, .history)
- Implement confirmation prompts for destructive operations
- Add command history with persistence

### Step 5: LLM Integration
- Implement LLM service in infrastructure using OpenAI client
- Design system prompts for SQL generation
- Implement schema inspection to provide context to LLM
- Add error handling and retry logic

### Step 6: Testing & Documentation
- Write unit tests for all components
- Write integration tests for SQL execution
- Create README with usage examples

## REPL Interaction Flow

```
1. Startup: Prompt for PostgreSQL credentials + OpenAI API key
2. Test connection, load schema
3. REPL Loop:
   - Display prompt (pg> or pg(tx)>)
   - Read user input
   - Parse input type (SQL vs natural language)
   - If natural language: Call LLM service, extract SQL
   - Always preview SQL before execution
   - Check if destructive → confirm before execute
   - Execute SQL, format results
   - Display errors with full details
   - Return to prompt
```

## Special REPL Commands

```
.exit          - Exit the REPL
.help          - Show help and available commands
.schema        - Show database schema
.tables        - List all tables
.desc <table>  - Describe table structure
.explain <sql> - Explain query execution plan
.history       - Show command history
.db            - Show current database
.user          - Show current user
.ver           - Show PostgreSQL version
```

## Critical Files

- `backend-app/binary/pg-agent/main.go` - Entry point
- `backend-app/infrastructure/database/postgres/postgres.go` - PostgreSQL client (REUSE)
- `backend-app/infrastructure/llm/service/service.go` - LLM integration (NEW)
- `backend-app/infrastructure/repl/readline/readline.go` - Readline wrapper (NEW)
- `backend-app/infrastructure/repl/ui/formatter.go` - Output formatter (NEW)
- `backend-app/pkg/parser/parser.go` - Input type detection (NEW)
- `backend-app/domain/agent/service/service.go` - All execution logic (connection, SQL, schema, transactions)
- `backend-app/domain/agent/handler/handler.go` - REPL interaction
- `backend-app/domain/agent/repository/repository.go` - Database queries

## Portfolio Patterns to Follow

- `backend-app/infrastructure/database/postgres/postgres.go` - SQLx client pattern
- `backend-app/binary/http/main.go` - Application structure
- `backend-app/domain/friend/service/service.go` - Service interface pattern
- `backend-app/domain/friend/repository/repository.go` - Repository pattern
- `backend-app/infrastructure/database/postgres/postgres_test.go` - Testing pattern

## Verification

1. Run the agent: `go run backend-app/binary/pg-agent/main.go`
2. Enter credentials and connect to a test database
3. Test natural language: "create table users with id serial primary key, name text, email text unique"
4. Verify SQL preview is shown
5. Confirm and execute
6. Test raw SQL: `INSERT INTO users (name, email) VALUES ('test', 'test@example.com')`
7. Test destructive: `DROP TABLE users` - verify confirmation prompt
8. Test special commands: `.tables`, `.schema`, `.exit`
