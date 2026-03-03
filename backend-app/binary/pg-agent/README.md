# PostgreSQL CLI Agent

An interactive REPL agent for PostgreSQL that interprets mixed input types (human instructions, raw SQL, or a combination) and executes queries against a database.

## Features

- **Interactive REPL** - Continuous command session with PostgreSQL
- **Natural Language to SQL** - Uses OpenAI's LLM to convert natural language to PostgreSQL queries
- **Mixed Input Support** - Type SQL queries directly or describe what you want in natural language
- **SQL Preview** - Always shows generated SQL before execution
- **Destructive Operation Protection** - DROP, DELETE, TRUNCATE require confirmation
- **Transaction Support** - BEGIN, COMMIT, ROLLBACK commands
- **Command History** - Persistent history with up/down arrow navigation
- **Schema Awareness** - LLM uses database schema for context-aware query generation

## Installation

1. Build the agent:
```bash
make pg-agent-build
```

2. Or run directly:
```bash
go run backend-app/binary/pg-agent/main.go
```

## Usage

1. Run the agent:
```bash
./bin/pg-agent
```

2. Enter your PostgreSQL credentials:
```
Enter PostgreSQL connection information:
PostgreSQL URI (leave blank for individual fields):
Host [localhost]:
Port [5432]:
Database Name: mydb
Username: postgres
Password: ******
SSL Mode [disable]:
```

3. (Optional) Enter your OpenAI API key for natural language to SQL conversion:
```
Enter OpenAI API key (leave blank to skip LLM features): sk-...
```

4. Start using the agent!

## Commands

### SQL Queries
```bash
# Direct SQL
SELECT * FROM users;

# Natural language
show me all users

# Counting
count how many users we have

# Create table
create a table called products with id serial primary key, name text, price decimal
```

### Transaction Support
```bash
BEGIN
INSERT INTO users (name, email) VALUES ('test', 'test@example.com')
COMMIT

-- or rollback
ROLLBACK
```

### Special Commands

| Command | Description |
|---------|-------------|
| `.help` | Show help and available commands |
| `.exit` | Exit the REPL |
| `.tables` | List all tables in the database |
| `.schema` | Show complete database schema |
| `.desc <table>` | Describe table structure |
| `.explain <sql>` | Show query execution plan |
| `.history` | Show command history |
| `.db` | Show current database |
| `.user` | Show current user |
| `.ver` | Show PostgreSQL version |

## Examples

### Query Data
```bash
pg> SELECT * FROM users;
```

### Natural Language Query
```bash
pg> show me all users with email ending in @gmail.com

Query Preview:
SELECT * FROM users WHERE email LIKE '%@gmail.com';

Execute this query [y/N]: y
```

### Create Table
```bash
pg> create a table called users with id serial primary key, name text not null, email text unique

Query Preview:
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    email TEXT UNIQUE
);

Execute this query [y/N]: y
```

### Destructive Operation (Requires Confirmation)
```bash
pg> DROP TABLE users;

Query Preview:
DROP TABLE users;

Warning: This DROP operation is destructive and cannot be undone.
Execute this destructive operation [y/N]: y
```

### Transaction Example
```bash
pg(tx)> BEGIN
pg(tx)> INSERT INTO users (name, email) VALUES ('Alice', 'alice@example.com')
pg(tx)> INSERT INTO users (name, email) VALUES ('Bob', 'bob@example.com')
pg(tx)> COMMIT
Success: 2 rows affected
```

## Architecture

The agent follows the portfolio's hexagonal architecture pattern:

```
backend-app/
├── binary/pg-agent/          # Entry point
├── domain/agent/             # Domain layer
│   ├── dto/                  # Data transfer objects
│   ├── handler/              # REPL interaction handler
│   ├── service/              # Business logic
│   └── repository/           # Data access
├── infrastructure/           # Infrastructure layer
│   ├── database/postgres/    # PostgreSQL client
│   ├── llm/                 # OpenAI integration
│   └── repl/                # REPL utilities
└── pkg/                     # Shared packages
    └── parser/              # Input type detection
```

## Configuration

### Environment Variables

- `OPENAI_API_KEY` - OpenAI API key for natural language to SQL conversion

### Connection Options

You can connect using either:
1. **Connection URI**: `postgresql://user:password@host:port/database`
2. **Individual fields**: Host, Port, Database, Username, Password

## Dependencies

- `github.com/jmoiron/sqlx` - Database operations
- `github.com/lib/pq` - PostgreSQL driver
- `github.com/openai/openai-go` - OpenAI API
- `github.com/chzyer/readline` - REPL interface
- `github.com/jedib0t/go-pretty/v6` - Table formatting
- `github.com/fatih/color` - Color output

## Testing

Run tests:
```bash
go test ./backend-app/domain/agent/... -v
go test ./backend-app/infrastructure/llm/... -v
go test ./backend-app/infrastructure/repl/... -v
go test ./backend-app/pkg/parser/... -v
```

## License

MIT
