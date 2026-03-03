package parser

import (
	"regexp"
	"strings"
	"unicode"
)

// InputType represents the type of user input
type InputType int

const (
	// TypeUnknown is unknown input type
	TypeUnknown InputType = iota
	// TypeSQL is raw SQL input
	TypeSQL
	// TypeNaturalLanguage is natural language input
	TypeNaturalLanguage
	// TypeMixed is mixed SQL and natural language
	TypeMixed
	// TypeCommand is a special REPL command (.exit, .help, etc.)
	TypeCommand
)

// SQLKeywords contains common SQL keywords for detection
var SQLKeywords = []string{
	"SELECT", "INSERT", "UPDATE", "DELETE", "DROP", "CREATE", "ALTER", "TRUNCATE",
	"GRANT", "REVOKE", "BEGIN", "COMMIT", "ROLLBACK", "WITH", "EXPLAIN",
	"SHOW", "DESCRIBE", "DESC", "\\d", "\\dt", "\\l",
}

// CommandPrefix is the prefix for special commands
const CommandPrefix = "."

// Result represents the parsed input
type Result struct {
	Type        InputType
	SQL         string
	NaturalLang string
	Command     string
}

// Parse analyzes the input and determines its type
func Parse(input string) Result {
	input = strings.TrimSpace(input)
	if input == "" {
		return Result{Type: TypeUnknown}
	}

	// Check for special commands
	if strings.HasPrefix(input, CommandPrefix) {
		return Result{
			Type:    TypeCommand,
			Command: strings.ToLower(input[1:]),
		}
	}

	// // Check for PostgreSQL meta-commands
	// if strings.HasPrefix(input, "\\") {
	// 	return Result{
	// 		Type:    TypeCommand,
	// 		Command: input,
	// 	}
	// }

	// // Extract SQL from the input
	// sql := extractSQL(input)
	// natural := extractNaturalLanguage(input)

	// // Determine the type
	// if sql != "" && natural != "" {
	// 	return Result{
	// 		Type:        TypeMixed,
	// 		SQL:         sql,
	// 		NaturalLang: natural,
	// 	}
	// }

	// if sql != "" {
	// 	return Result{
	// 		Type: TypeSQL,
	// 		SQL:  sql,
	// 	}
	// }

	return Result{
		Type:        TypeNaturalLanguage,
		NaturalLang: input,
	}
}

// IsDestructiveOperation returns true if the SQL contains a destructive operation
func IsDestructiveOperation(sql string) bool {
	upper := strings.ToUpper(sql)

	destructivePatterns := []string{
		`\bDROP\s+(TABLE|INDEX|VIEW|DATABASE|SCHEMA)`,
		`\bDELETE\s+FROM`,
		`\bTRUNCATE\s+(TABLE|TABLES)`,
		`\bALTER\s+TABLE\s+\w+\s+DROP`,
	}

	for _, pattern := range destructivePatterns {
		if matched, _ := regexp.MatchString(pattern, upper); matched {
			return true
		}
	}

	return false
}

// IsTransactionCommand returns true if the input is a transaction command
func IsTransactionCommand(input string) (bool, string) {
	upper := strings.ToUpper(strings.TrimSpace(input))

	switch {
	case upper == "BEGIN" || upper == "BEGIN TRANSACTION" || upper == "START TRANSACTION":
		return true, "BEGIN"
	case upper == "COMMIT" || upper == "COMMIT TRANSACTION":
		return true, "COMMIT"
	case upper == "ROLLBACK" || upper == "ROLLBACK TRANSACTION":
		return true, "ROLLBACK"
	default:
		return false, ""
	}
}

// extractSQL extracts SQL statements from the input
func extractSQL(input string) string {
	// First, check if input starts with a SQL keyword
	words := strings.Fields(input)
	if len(words) == 0 {
		return ""
	}

	firstWord := strings.ToUpper(words[0])
	for _, keyword := range SQLKeywords {
		if firstWord == keyword {
			return input
		}
	}

	// If not, look for embedded SQL patterns
	// This is a simple heuristic - look for SELECT/INSERT/etc. in the input
	sqlPattern := regexp.MustCompile(`(?i)\b(SELECT|INSERT|UPDATE|DELETE|DROP|CREATE|ALTER|TRUNCATE|BEGIN|COMMIT|ROLLBACK|WITH)\s+`)

	matches := sqlPattern.FindAllStringIndex(input, -1)
	if len(matches) == 0 {
		return ""
	}

	// If we have SQL keywords, extract the SQL portion
	// This is a simple approach - assumes SQL is at the end or contains the keywords
	return input
}

// extractNaturalLanguage extracts natural language text from the input
func extractNaturalLanguage(input string) string {
	// Remove SQL keywords and look for natural language patterns
	words := strings.Fields(input)
	if len(words) == 0 {
		return ""
	}

	firstWord := strings.ToUpper(words[0])

	// Check if this starts with a SQL keyword
	for _, keyword := range SQLKeywords {
		if firstWord == keyword {
			return ""
		}
	}

	// Check if this looks like SQL but has natural language
	sqlPattern := regexp.MustCompile(`(?i)\b(SELECT|INSERT|UPDATE|DELETE|DROP|CREATE|ALTER|TRUNCATE|BEGIN|COMMIT|ROLLBACK|WITH)\s+`)

	// If no SQL keywords found, it's natural language
	if !sqlPattern.MatchString(input) {
		return input
	}

	// If we have SQL keywords, check for natural language before/after
	// Look for patterns like "show me all users" or "create a table called users with..."
	if looksLikeNaturalLanguage(input) {
		return input
	}

	return ""
}

// looksLikeNaturalLanguage checks if input looks like natural language
func looksLikeNaturalLanguage(input string) bool {
	// Look for natural language patterns
	// Common NL patterns in database queries:
	// - "show me", "get", "find", "list", "create a", "delete the"
	nlPatterns := []string{
		`\bshow me\b`,
		`\bget\b`,
		`\bfind\b`,
		`\blist\b`,
		`\bcreate a\b`,
		`\bcreate an\b`,
		`\bdelete the\b`,
		`\bdelete all\b`,
		`\bupdate the\b`,
		`\badd\b`,
		`\binsert\b`,
		`\bhow many\b`,
		`\bcount\b`,
		`\bwith\b`,
		`\bwhere\b`,
		`\bthat\b`,
		`\band\b`,
		`\bor\b`,
		`\bnot\b`,
	}

	upper := strings.ToUpper(input)
	for _, pattern := range nlPatterns {
		if matched, _ := regexp.MatchString(pattern, upper); matched {
			return true
		}
	}

	// Check for question words
	questionWords := []string{"what", "which", "where", "when", "who", "how", "why"}
	for _, word := range questionWords {
		if strings.HasPrefix(strings.ToLower(input), word+" ") {
			return true
		}
	}

	return false
}

// ExtractTableName extracts table name from SQL (basic extraction)
func ExtractTableName(sql string) []string {
	upper := strings.ToUpper(sql)

	var tables []string

	// Extract from CREATE TABLE
	createTable := regexp.MustCompile(`\bCREATE\s+TABLE\s+(?:IF\s+NOT\s+EXISTS\s+)?(\w+)`)
	matches := createTable.FindAllStringSubmatch(upper, -1)
	for _, match := range matches {
		if len(match) > 1 {
			tables = append(tables, match[1])
		}
	}

	// Extract from DROP TABLE
	dropTable := regexp.MustCompile(`\bDROP\s+TABLE\s+(?:IF\s+EXISTS\s+)?(\w+)`)
	matches = dropTable.FindAllStringSubmatch(upper, -1)
	for _, match := range matches {
		if len(match) > 1 {
			tables = append(tables, match[1])
		}
	}

	// Extract from SELECT ... FROM
	selectFrom := regexp.MustCompile(`\bFROM\s+([\w.]+)`)
	matches = selectFrom.FindAllStringSubmatch(upper, -1)
	for _, match := range matches {
		if len(match) > 1 {
			tables = append(tables, match[1])
		}
	}

	// Extract from INSERT INTO
	insertInto := regexp.MustCompile(`\bINSERT\s+INTO\s+([\w.]+)`)
	matches = insertInto.FindAllStringSubmatch(upper, -1)
	for _, match := range matches {
		if len(match) > 1 {
			tables = append(tables, match[1])
		}
	}

	// Extract from UPDATE ... SET
	updateTable := regexp.MustCompile(`\bUPDATE\s+([\w.]+)\s+SET`)
	matches = updateTable.FindAllStringSubmatch(upper, -1)
	for _, match := range matches {
		if len(match) > 1 {
			tables = append(tables, match[1])
		}
	}

	// Extract from DELETE FROM
	deleteFrom := regexp.MustCompile(`\bDELETE\s+FROM\s+([\w.]+)`)
	matches = deleteFrom.FindAllStringSubmatch(upper, -1)
	for _, match := range matches {
		if len(match) > 1 {
			tables = append(tables, match[1])
		}
	}

	return tables
}

// SanitizeInput removes excessive whitespace and normalizes input
func SanitizeInput(input string) string {
	// Remove extra whitespace
	input = strings.TrimSpace(input)

	// Replace multiple spaces with single space
	spaceRegex := regexp.MustCompile(`\s+`)
	input = spaceRegex.ReplaceAllString(input, " ")

	// Remove trailing semicolon
	input = strings.TrimSuffix(input, ";")

	return input
}

// IsValidSQL performs basic SQL validation
func IsValidSQL(sql string) bool {
	sql = strings.TrimSpace(sql)
	if sql == "" {
		return false
	}

	// Check if it starts with a known SQL keyword
	words := strings.Fields(sql)
	if len(words) == 0 {
		return false
	}

	firstWord := strings.ToUpper(words[0])
	for _, keyword := range SQLKeywords {
		if firstWord == keyword {
			return true
		}
	}

	// Check for PostgreSQL meta-commands
	if strings.HasPrefix(sql, "\\") {
		return true
	}

	return false
}

// HasBalancedParens checks if parentheses are balanced
func HasBalancedParens(s string) bool {
	count := 0
	for _, r := range s {
		switch r {
		case '(':
			count++
		case ')':
			count--
			if count < 0 {
				return false
			}
		}
	}
	return count == 0
}

// HasBalancedQuotes checks if quotes are balanced
func HasBalancedQuotes(s string) bool {
	singleQuote := false
	doubleQuote := false
	escape := false

	for _, r := range s {
		if escape {
			escape = false
			continue
		}

		switch r {
		case '\\':
			escape = true
		case '\'':
			if !doubleQuote {
				singleQuote = !singleQuote
			}
		case '"':
			if !singleQuote {
				doubleQuote = !doubleQuote
			}
		}
	}

	return !singleQuote && !doubleQuote
}

// IsCompleteStatement checks if a SQL statement is complete
func IsCompleteStatement(sql string) bool {
	sql = strings.TrimSpace(sql)
	if sql == "" {
		return false
	}

	// Check for trailing semicolon
	if strings.HasSuffix(sql, ";") {
		return true
	}

	// Check if we have balanced parens and quotes
	if !HasBalancedParens(sql) {
		return false
	}

	if !HasBalancedQuotes(sql) {
		return false
	}

	// Check if it ends with a keyword that might indicate completion
	upper := strings.ToUpper(sql)
	completionPatterns := []string{
		`\bTHEN\b`,
		`\bEND\b`,
		`\bDO\b`,
		`\bAS\b`,
	}

	for _, pattern := range completionPatterns {
		if matched, _ := regexp.MatchString(pattern, upper); matched {
			return true
		}
	}

	return true
}

// TrimComments removes SQL comments
func TrimComments(sql string) string {
	// Remove single-line comments
	sql = regexp.MustCompile(`--[^\n]*`).ReplaceAllString(sql, "")

	// Remove multi-line comments
	sql = regexp.MustCompile(`/\*.*?\*/`).ReplaceAllString(sql, "")

	return strings.TrimSpace(sql)
}

// NormalizeKeywords normalizes SQL keywords to uppercase
func NormalizeKeywords(sql string) string {
	result := make([]rune, 0, len(sql))
	wordStart := true

	for _, r := range sql {
		if wordStart && unicode.IsLetter(r) {
			result = append(result, unicode.ToUpper(r))
			wordStart = false
		} else if unicode.IsSpace(r) || r == '(' || r == ')' || r == ',' {
			result = append(result, r)
			wordStart = true
		} else {
			result = append(result, r)
		}
	}

	return string(result)
}
