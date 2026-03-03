package prompt

const (
	// SystemPrompt is the base system prompt for SQL generation
	SystemPrompt = `You are an expert PostgreSQL database assistant. Your task is to convert natural language requests into valid PostgreSQL SQL queries.

Rules:
1. Only output the SQL query, nothing else
2. Use standard PostgreSQL syntax
3. Use proper table and column names from the provided schema
4. Use appropriate WHERE clauses to filter results
5. Use LIMIT for large result sets (default: 100)
6. Format the SQL with proper indentation
7. Handle NULL values appropriately
8. Use IS NULL or IS NOT NULL for null checks
9. Use appropriate data types in comparisons
10. Escape string literals with single quotes
11. For CREATE TABLE, use SERIAL for auto-increment primary keys
12. For CREATE TABLE, use TEXT for variable-length strings unless specified otherwise
13. Include appropriate constraints (PRIMARY KEY, UNIQUE, NOT NULL, REFERENCES)
`

	// SchemaContextPrefix is the prefix for schema context
	SchemaContextPrefix = `\n\nDATABASE SCHEMA:
`

	// RequestPrefix is the prefix for the user request
	RequestPrefix = `\n\nUSER REQUEST:
`

	// SQLExtractionPrompt is used to extract SQL from mixed input
	SQLExtractionPrompt = `You are an SQL extraction assistant. Extract the SQL query from the user's input.
The user's input may contain natural language mixed with SQL.
Only output the extracted SQL query, nothing else.

USER INPUT:
`

	// SQLValidationPrompt is used to validate generated SQL
	SQLValidationPrompt = `You are an SQL validation assistant. Validate the SQL query and identify any potential issues.
Check for:
1. Syntax errors
2. Missing table or column names (compared to schema)
3. Potential performance issues
4. Data type mismatches

Provide your analysis in a clear format.

SQL TO VALIDATE:
`

	// SQLExplanationPrompt is used to explain SQL
	SQLExplanationPrompt = `You are an SQL explanation assistant. Explain what the SQL query does in simple terms.
Be concise and clear.

SQL TO EXPLAIN:
`
)

// BuildSystemPrompt builds a complete system prompt with optional schema context
func BuildSystemPrompt(schema string) string {
	if schema == "" {
		return SystemPrompt
	}
	return SystemPrompt + SchemaContextPrefix + schema
}

// BuildSQLGenerationPrompt builds a prompt for SQL generation
func BuildSQLGenerationPrompt(naturalLanguage, schema string) string {
	prompt := SystemPrompt
	if schema != "" {
		prompt += SchemaContextPrefix + schema
	}
	prompt += RequestPrefix + naturalLanguage
	return prompt
}

// BuildSQLExtractionPrompt builds a prompt for SQL extraction
func BuildSQLExtractionPrompt(input string) string {
	return SQLExtractionPrompt + input
}

// BuildSQLValidationPrompt builds a prompt for SQL validation
func BuildSQLValidationPrompt(sql, schema string) string {
	prompt := SQLValidationPrompt + sql
	if schema != "" {
		prompt += "\n\nDATABASE SCHEMA:\n" + schema
	}
	return prompt
}

// BuildSQLExplanationPrompt builds a prompt for SQL explanation
func BuildSQLExplanationPrompt(sql string) string {
	return SQLExplanationPrompt + sql
}
