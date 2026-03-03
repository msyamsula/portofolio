package ui

import (
	"fmt"
	"strings"
)

// PromptState represents the current state of the REPL prompt
type PromptState int

const (
	// StateNormal is the normal prompt state
	StateNormal PromptState = iota
	// StateTransaction is in a transaction
	StateTransaction
	// StateMultiline is waiting for multiline input
	StateMultiline
)

// Prompt manages the REPL prompt
type Prompt struct {
	state     PromptState
	database  string
	user      string
	host      string
}

// NewPrompt creates a new prompt manager
func NewPrompt() *Prompt {
	return &Prompt{
		state: StateNormal,
	}
}

// SetState sets the current prompt state
func (p *Prompt) SetState(state PromptState) {
	p.state = state
}

// GetState returns the current prompt state
func (p *Prompt) GetState() PromptState {
	return p.state
}

// SetDatabase sets the current database name
func (p *Prompt) SetDatabase(db string) {
	p.database = db
}

// SetUser sets the current user
func (p *Prompt) SetUser(user string) {
	p.user = user
}

// SetHost sets the current host
func (p *Prompt) SetHost(host string) {
	p.host = host
}

// GetPrompt returns the current prompt string
func (p *Prompt) GetPrompt() string {
	f := NewFormatter()
	switch p.state {
	case StateTransaction:
		return f.FormatHighlight("pg(tx)> ") + " "
	case StateMultiline:
		return f.FormatHighlight("... ")
	default:
		return f.FormatHighlight("pg> ") + " "
	}
}

// GetPromptWithInfo returns the prompt with connection info
func (p *Prompt) GetPromptWithInfo() string {
	var info strings.Builder
	f := NewFormatter()

	if p.database != "" {
		info.WriteString(p.database)
	}
	if p.user != "" {
		if info.Len() > 0 {
			info.WriteString("@")
		}
		info.WriteString(p.user)
	}

	var suffix string
	switch p.state {
	case StateTransaction:
		suffix = "(tx)>"
	case StateMultiline:
		suffix = "..."
	default:
		suffix = ">"
	}

	if info.Len() > 0 {
		return f.FormatHighlight(fmt.Sprintf("%s%s ", info.String(), suffix)) + " "
	}
	return f.FormatHighlight(fmt.Sprintf("pg%s ", suffix)) + " "
}

// InTransaction returns true if currently in a transaction
func (p *Prompt) InTransaction() bool {
	return p.state == StateTransaction
}

// SetTransaction sets the transaction state
func (p *Prompt) SetTransaction(inTransaction bool) {
	if inTransaction {
		p.state = StateTransaction
	} else {
		p.state = StateNormal
	}
}

// SetMultiline sets the multiline state
func (p *Prompt) SetMultiline(isMultiline bool) {
	if isMultiline {
		p.state = StateMultiline
	} else if p.state == StateMultiline {
		p.state = StateNormal
	}
}

// FormatConnectionPrompt formats the connection prompt
func FormatConnectionPrompt(field string) string {
	switch strings.ToLower(field) {
	case "uri":
		return "PostgreSQL URI (leave blank to use individual fields): "
	case "host":
		return "PostgreSQL Host [localhost]: "
	case "port":
		return "PostgreSQL Port [5432]: "
	case "database":
		return "Database Name: "
	case "user":
		return "Username: "
	case "password":
		return "Password: "
	case "api_key":
		return "OpenAI API Key: "
	default:
		return fmt.Sprintf("%s: ", field)
	}
}

// FormatConfirmationPrompt formats a confirmation prompt
func FormatConfirmationPrompt(message string) string {
	f := NewFormatter()
	return fmt.Sprintf("%s [y/N]: ", f.FormatWarn(message))
}

// FormatYesNo formats a yes/no choice
func FormatYesNo(defaultYes bool) string {
	if defaultYes {
		return " [Y/n]: "
	}
	return " [y/N]: "
}
