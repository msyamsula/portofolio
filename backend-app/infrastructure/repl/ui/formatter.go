package ui

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/jedib0t/go-pretty/v6/table"
)

var (
	// Color styles
	successColor = color.New(color.FgGreen).SprintFunc()
	errorColor   = color.New(color.FgRed).SprintFunc()
	warnColor    = color.New(color.FgYellow).SprintFunc()
	infoColor    = color.New(color.FgCyan).SprintFunc()
	sqlColor     = color.New(color.FgMagenta).SprintFunc()
	highlightColor = color.New(color.Bold, color.FgWhite).SprintFunc()
)

// Formatter handles output formatting for the REPL
type Formatter struct{}

// NewFormatter creates a new formatter instance
func NewFormatter() *Formatter {
	return &Formatter{}
}

// FormatTable formats a result set as a table
func (f *Formatter) FormatTable(headers []string, rows [][]string) string {
	if len(rows) == 0 {
		return "(no rows returned)"
	}

	t := table.NewWriter()
	t.SetOutputMirror(nil)
	t.SetStyle(table.StyleRounded)
	t.SetAutoIndex(false)

	// Convert headers to table.Row
	headerRow := make(table.Row, len(headers))
	for i, h := range headers {
		headerRow[i] = h
	}
	t.AppendHeader(headerRow)

	// Convert each row to table.Row
	for _, row := range rows {
		rowSlice := make(table.Row, len(row))
		for i, r := range row {
			rowSlice[i] = r
		}
		t.AppendRow(rowSlice)
	}

	return t.Render()
}

// FormatSuccess formats a success message
func (f *Formatter) FormatSuccess(msg string) string {
	return successColor(msg)
}

// FormatError formats an error message
func (f *Formatter) FormatError(msg string) string {
	return errorColor(msg)
}

// FormatWarn formats a warning message
func (f *Formatter) FormatWarn(msg string) string {
	return warnColor(msg)
}

// FormatInfo formats an info message
func (f *Formatter) FormatInfo(msg string) string {
	return infoColor(msg)
}

// FormatSQL formats a SQL query
func (f *Formatter) FormatSQL(sql string) string {
	return sqlColor(sql)
}

// FormatHighlight highlights text
func (f *Formatter) FormatHighlight(text string) string {
	return highlightColor(text)
}

// FormatRowCount formats a row count message
func (f *Formatter) FormatRowCount(count int) string {
	if count == 0 {
		return f.FormatInfo("(no rows)")
	}
	return f.FormatInfo(fmt.Sprintf("(%d row%s)", count, pluralize(count)))
}

// FormatQueryPreview formats a SQL query preview
func (f *Formatter) FormatQueryPreview(sql string) string {
	header := f.FormatHighlight("Query Preview:")
	return fmt.Sprintf("%s\n%s", header, f.FormatSQL(sql))
}

// FormatConfirmation formats a confirmation prompt
func (f *Formatter) FormatConfirmation(msg string) string {
	return fmt.Sprintf("%s [y/N]", f.FormatWarn(msg))
}

// FormatDestructiveWarning formats a warning for destructive operations
func (f *Formatter) FormatDestructiveWarning(operation string) string {
	warning := fmt.Sprintf("Warning: This %s operation is destructive and cannot be undone.", operation)
	return f.FormatWarn(warning)
}

// FormatKeyValue formats key-value pairs
func (f *Formatter) FormatKeyValue(pairs map[string]string) string {
	var builder strings.Builder
	for k, v := range pairs {
		builder.WriteString(f.FormatInfo(k))
		builder.WriteString(": ")
		builder.WriteString(v)
		builder.WriteString("\n")
	}
	return builder.String()
}

// FormatCommandHelp formats a command help entry
func (f *Formatter) FormatCommandHelp(cmd, desc string) string {
	return fmt.Sprintf("  %-15s %s", f.FormatHighlight(cmd), desc)
}

// FormatWelcome formats the welcome message
func (f *Formatter) FormatWelcome() string {
	return `
╔════════════════════════════════════════════════════════════╗
║          PostgreSQL CLI Agent - Interactive REPL             ║
╚════════════════════════════════════════════════════════════╝

Type SQL queries or natural language descriptions directly.
Type .help for available commands or .exit to quit.
`
}

// FormatConnectionSuccess formats a successful connection message
func (f *Formatter) FormatConnectionSuccess(database, user string) string {
	return f.FormatSuccess(fmt.Sprintf("Connected to %s as %s", database, user))
}

// FormatConnectionError formats a connection error message
func (f *Formatter) FormatConnectionError(err error) string {
	return f.FormatError(fmt.Sprintf("Connection failed: %v", err))
}

// pluralize returns the plural form of a word if count != 1
func pluralize(count int) string {
	if count == 1 {
		return ""
	}
	return "s"
}
