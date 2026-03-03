package readline

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/chzyer/readline"
)

var (
	// ErrInterrupt is returned when user interrupts input with Ctrl+C
	ErrInterrupt = errors.New("interrupted")

	// ErrEOF is returned when EOF is reached
	ErrEOF = errors.New("EOF")
)

// Config holds configuration for the readline instance
type Config struct {
	Prompt      string
	HistoryFile string
}

// Readline wraps the chzyer/readline instance
type Readline struct {
	instance *readline.Instance
	config   Config
}

// New creates a new readline instance
func New(cfg Config) (*Readline, error) {
	var err error

	rl, err := readline.NewEx(&readline.Config{
		Prompt:          cfg.Prompt,
		HistoryFile:     cfg.HistoryFile,
		HistorySearchFold: true,
		AutoComplete:    nil,
		InterruptPrompt:  "^C",
		EOFPrompt:       "exit",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create readline: %w", err)
	}

	return &Readline{
		instance: rl,
		config:   cfg,
	}, nil
}

// Read reads a single line of input
func (r *Readline) Read() (string, error) {
	line, err := r.instance.Readline()
	if err != nil {
		if errors.Is(err, readline.ErrInterrupt) {
			return "", ErrInterrupt
		}
		if err == io.EOF {
			return "", ErrEOF
		}
		return "", fmt.Errorf("read error: %w", err)
	}
	return strings.TrimSpace(line), nil
}

// ReadMultiline reads input until an empty line or semicolon is encountered
func (r *Readline) ReadMultiline() (string, error) {
	var builder strings.Builder
	lineNumber := 0

	for {
		prompt := r.config.Prompt
		if lineNumber > 0 {
			prompt = "... "
		}

		r.instance.SetPrompt(prompt)

		line, err := r.instance.Readline()
		if err != nil {
			if errors.Is(err, readline.ErrInterrupt) {
				return "", ErrInterrupt
			}
			if err == io.EOF {
				return "", ErrEOF
			}
			return "", fmt.Errorf("read error: %w", err)
		}

		trimmed := strings.TrimSpace(line)

		// Empty line ends multiline input
		if trimmed == "" && builder.Len() > 0 {
			break
		}

		// Semicolon at end ends multiline input
		if strings.HasSuffix(trimmed, ";") {
			if builder.Len() > 0 {
				builder.WriteString(" ")
			}
			builder.WriteString(trimmed)
			break
		}

		if builder.Len() > 0 {
			builder.WriteString("\n")
		}
		builder.WriteString(line)
		lineNumber++
	}

	return builder.String(), nil
}

// ReadPassword reads password input without echoing
func (r *Readline) ReadPassword(prompt string) (string, error) {
	// For password, we just use a simple prompt since readline doesn't have password mode
	fmt.Print(prompt)
	var password string
	_, err := fmt.Scanln(&password)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(password), nil
}

// SetPrompt updates the current prompt
func (r *Readline) SetPrompt(prompt string) {
	r.instance.SetPrompt(prompt)
	r.config.Prompt = prompt
}

// Write writes output to the readline instance
func (r *Readline) Write(data []byte) (int, error) {
	return r.instance.Stdout().Write(data)
}

// WriteString writes a string to the readline instance
func (r *Readline) WriteString(s string) (int, error) {
	return r.Write([]byte(s))
}

// Println prints a line to the readline instance
func (r *Readline) Println(a ...any) {
	r.WriteString(fmt.Sprintln(a...))
}

// Printf prints formatted output to the readline instance
func (r *Readline) Printf(format string, a ...any) {
	r.WriteString(fmt.Sprintf(format, a...))
}

// Close closes the readline instance
func (r *Readline) Close() error {
	return r.instance.Close()
}

// History returns the command history
func (r *Readline) History() []string {
	// The readline package doesn't expose GetHistory directly
	// Return empty for now
	return []string{}
}

// Stdout returns the stdout writer
func (r *Readline) Stdout() io.Writer {
	return r.instance.Stdout()
}

// Stderr returns the stderr writer
func (r *Readline) Stderr() io.Writer {
	return r.instance.Stderr()
}
