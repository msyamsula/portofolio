package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseCommand(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected InputType
		command  string
	}{
		{"exit command", ".exit", TypeCommand, "exit"},
		{"help command", ".help", TypeCommand, "help"},
		{"tables command", ".tables", TypeCommand, "tables"},
		{"schema command", ".schema", TypeCommand, "schema"},
		{"desc command", ".desc users", TypeCommand, "desc users"},
		{"explain command", ".explain SELECT * FROM users", TypeCommand, "explain select * from users"},
		{"postgres meta command", "\\dt", TypeCommand, "\\dt"},
		{"postgres meta desc", "\\d users", TypeCommand, "\\d users"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Parse(tt.input)
			assert.Equal(t, tt.expected, result.Type)
			assert.Equal(t, tt.command, result.Command)
		})
	}
}

func TestParseSQL(t *testing.T) {
	tests := []struct {
		name  string
		input string
		SQL   string
	}{
		{"SELECT", "SELECT * FROM users", "SELECT * FROM users"},
		{"INSERT", "INSERT INTO users (name) VALUES ('test')", "INSERT INTO users (name) VALUES ('test')"},
		{"UPDATE", "UPDATE users SET name = 'test' WHERE id = 1", "UPDATE users SET name = 'test' WHERE id = 1"},
		{"DELETE", "DELETE FROM users WHERE id = 1", "DELETE FROM users WHERE id = 1"},
		{"DROP", "DROP TABLE users", "DROP TABLE users"},
		{"CREATE", "CREATE TABLE users (id SERIAL PRIMARY KEY)", "CREATE TABLE users (id SERIAL PRIMARY KEY)"},
		{"BEGIN", "BEGIN TRANSACTION", "BEGIN TRANSACTION"},
		{"COMMIT", "COMMIT", "COMMIT"},
		{"ROLLBACK", "ROLLBACK", "ROLLBACK"},
		{"EXPLAIN", "EXPLAIN SELECT * FROM users", "EXPLAIN SELECT * FROM users"},
		{"SELECT with semicolon", "SELECT * FROM users;", "SELECT * FROM users"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Parse(tt.input)
			assert.Equal(t, TypeSQL, result.Type)
			assert.Equal(t, tt.SQL, result.SQL)
		})
	}
}

func TestParseNaturalLanguage(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		expectedType InputType
	}{
		{"show me", "show me all users", TypeNaturalLanguage},
		{"get all", "get all users from the database", TypeNaturalLanguage},
		{"find users", "find users with id greater than 10", TypeNaturalLanguage},
		{"create table", "create a table called users with id and name", TypeNaturalLanguage},
		{"count", "count how many users we have", TypeNaturalLanguage},
		{"what", "what is the total number of users", TypeNaturalLanguage},
		{"list", "list all products in the inventory", TypeNaturalLanguage},
		{"delete", "delete the user with email test@example.com", TypeNaturalLanguage},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Parse(tt.input)
			assert.Equal(t, tt.expectedType, result.Type)
		})
	}
}

func TestIsDestructiveOperation(t *testing.T) {
	tests := []struct {
		name     string
		sql      string
		expected bool
	}{
		{"DROP TABLE", "DROP TABLE users", true},
		{"DROP TABLE IF EXISTS", "DROP TABLE IF EXISTS users", true},
		{"DROP INDEX", "DROP INDEX idx_name", true},
		{"DELETE", "DELETE FROM users WHERE id = 1", true},
		{"DELETE FROM", "delete from users", true},
		{"TRUNCATE", "TRUNCATE TABLE users", true},
		{"TRUNCATE", "TRUNCATE users", true},
		{"SELECT", "SELECT * FROM users", false},
		{"INSERT", "INSERT INTO users (name) VALUES ('test')", false},
		{"UPDATE", "UPDATE users SET name = 'test'", false},
		{"CREATE", "CREATE TABLE users (id INT)", false},
		{"ALTER ADD", "ALTER TABLE users ADD COLUMN email TEXT", false},
		{"ALTER DROP", "ALTER TABLE users DROP COLUMN email", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsDestructiveOperation(tt.sql)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsTransactionCommand(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		isCommand  bool
		command    string
	}{
		{"BEGIN", "BEGIN", true, "BEGIN"},
		{"BEGIN TRANSACTION", "BEGIN TRANSACTION", true, "BEGIN"},
		{"START TRANSACTION", "START TRANSACTION", true, "BEGIN"},
		{"COMMIT", "COMMIT", true, "COMMIT"},
		{"COMMIT TRANSACTION", "COMMIT TRANSACTION", true, "COMMIT"},
		{"ROLLBACK", "ROLLBACK", true, "ROLLBACK"},
		{"ROLLBACK TRANSACTION", "ROLLBACK TRANSACTION", true, "ROLLBACK"},
		{"SELECT", "SELECT * FROM users", false, ""},
		{"INSERT", "INSERT INTO users VALUES (1)", false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isCmd, cmd := IsTransactionCommand(tt.input)
			assert.Equal(t, tt.isCommand, isCmd)
			assert.Equal(t, tt.command, cmd)
		})
	}
}

func TestExtractTableName(t *testing.T) {
	tests := []struct {
		name     string
		sql      string
		expected []string
	}{
		{"CREATE TABLE", "CREATE TABLE users (id INT)", []string{"USERS"}},
		{"DROP TABLE", "DROP TABLE users", []string{"USERS"}},
		{"SELECT FROM", "SELECT * FROM users WHERE id = 1", []string{"USERS"}},
		{"INSERT INTO", "INSERT INTO users (name) VALUES ('test')", []string{"USERS"}},
		{"UPDATE SET", "UPDATE users SET name = 'test' WHERE id = 1", []string{"USERS"}},
		{"DELETE FROM", "DELETE FROM users WHERE id = 1", []string{"USERS"}},
		{"JOIN query", "SELECT * FROM users JOIN orders ON users.id = orders.user_id", []string{"USERS", "ORDERS"}},
		{"Multiple tables", "SELECT * FROM users, orders WHERE users.id = orders.user_id", []string{"USERS", "ORDERS"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractTableName(tt.sql)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSanitizeInput(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"  SELECT * FROM users  ", "SELECT * FROM users"},
		{"SELECT   *   FROM   users", "SELECT * FROM users"},
		{"SELECT * FROM users;", "SELECT * FROM users"},
		{"  SELECT   *   FROM   users;  ", "SELECT * FROM users"},
		{"\n\tSELECT * FROM users\n\t", "SELECT * FROM users"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := SanitizeInput(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsValidSQL(t *testing.T) {
	tests := []struct {
		name     string
		sql      string
		expected bool
	}{
		{"SELECT", "SELECT * FROM users", true},
		{"INSERT", "INSERT INTO users VALUES (1)", true},
		{"UPDATE", "UPDATE users SET name = 'test'", true},
		{"DELETE", "DELETE FROM users", true},
		{"DROP", "DROP TABLE users", true},
		{"CREATE", "CREATE TABLE users (id INT)", true},
		{"BEGIN", "BEGIN TRANSACTION", true},
		{"lowercase select", "select * from users", true},
		{"random text", "show me all users", false},
		{"empty", "", false},
		{"postgres meta", "\\dt", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidSQL(tt.sql)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestHasBalancedParens(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"no parens", "", true},
		{"balanced", "(a OR b)", true},
		{"multiple balanced", "(a OR (b AND c))", true},
		{"unbalanced", "(a OR b", false},
		{"unbalanced 2", "a OR b)", false},
		{"nested balanced", "((a))", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := HasBalancedParens(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestHasBalancedQuotes(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"no quotes", "", true},
		{"balanced single", "'hello'", true},
		{"balanced double", `"hello"`, true},
		{"unbalanced single", "'hello", false},
		{"unbalanced double", `"hello`, false},
		{"mixed", `'hello' "world"`, true},
		{"escaped", "'don\\'t'", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := HasBalancedQuotes(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsCompleteStatement(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"SELECT * FROM users;", true},
		{"SELECT * FROM users", true},
		{"SELECT * FROM users WHERE", true},
		{"SELECT * FROM users WHERE (id = 1", false},
		{"SELECT * FROM users WHERE name = 'test", false},
		{"SELECT * FROM users WHERE (id = 1 OR id = 2)", true},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := IsCompleteStatement(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTrimComments(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"SELECT * FROM users -- this is a comment", "SELECT * FROM users"},
		{"SELECT * FROM users /* multi-line comment */ WHERE id = 1", "SELECT * FROM users WHERE id = 1"},
		{"-- line 1\nSELECT * FROM users", "SELECT * FROM users"},
		{"SELECT * FROM users WHERE id = 1 -- end", "SELECT * FROM users WHERE id = 1"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := TrimComments(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNormalizeKeywords(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"select * from users", "SELECT * FROM users"},
		{"SELECT * FROM users", "SELECT * FROM users"},
		{"SeLeCt * FrOm UsErS", "SELECT * FROM USERS"},
		{"SELECT name, email FROM users", "SELECT name, email FROM users"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := NormalizeKeywords(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
