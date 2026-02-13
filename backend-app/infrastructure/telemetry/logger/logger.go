package logger

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime"
	"time"

	otellog "go.opentelemetry.io/otel/log"
)

// Level represents log level
type Level int

const (
	// DebugLevel logs are typically voluminous, and are usually disabled in production
	DebugLevel Level = iota
	// InfoLevel is the default logging priority
	InfoLevel
	// WarnLevel logs are more important than Info, but don't need individual human review
	WarnLevel
	// ErrorLevel logs are high-priority. If an application is running smoothly, it shouldn't generate any error-level logs
	ErrorLevel
)

// String returns the string representation of the log level
func (l Level) String() string {
	switch l {
	case DebugLevel:
		return "DEBUG"
	case InfoLevel:
		return "INFO"
	case WarnLevel:
		return "WARN"
	case ErrorLevel:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// Format represents the output format
type Format int

const (
	// TextFormat is human-readable text format
	TextFormat Format = iota
	// JSONFormat is machine-readable JSON format
	JSONFormat
)

// Config holds the logger configuration
type Config struct {
	ServiceName       string
	CollectorEndpoint string
	Insecure         bool
	Environment      string
	LogsEnabled      bool

	// Console settings
	Level      Level
	Format     Format
	TimeFormat string
}

var (
	cfg            Config
	otelLogger     otellog.Logger
	exporterClient *ExporterClient
)

// Init initializes the logger with the given configuration
func Init(ctx context.Context, c Config) error {
	cfg = c

	if c.TimeFormat == "" {
		cfg.TimeFormat = time.RFC3339
	}

	if c.LogsEnabled {
		exporterCfg := ExporterConfig{
			ServiceName:       c.ServiceName,
			CollectorEndpoint: c.CollectorEndpoint,
			Insecure:         c.Insecure,
			Environment:      c.Environment,
		}
		client, err := NewExporterClient(ctx, exporterCfg)
		if err != nil {
			return fmt.Errorf("failed to initialize logs exporter: %w", err)
		}
		exporterClient = client
		otelLogger = client.OTELLogger(c.ServiceName)
		SetOTELLogger(otelLogger)
	}

	return nil
}

// Shutdown gracefully shuts down the logger
func Shutdown(ctx context.Context) error {
	if exporterClient != nil {
		return exporterClient.Shutdown(ctx)
	}
	return nil
}

// logFunc is the function that writes the log output
var logFunc = func(format string, args ...any) {
	fmt.Printf(format+"\n", args...)
}

// shouldLog determines if a message should be logged based on level
func shouldLog(level Level) bool {
	return level >= cfg.Level
}

// formatTime formats the current time
var formatTime = func(layout string) string {
	return time.Now().Format(layout)
}

// getCaller gets the file and line number of the caller
var getCaller = func(skip int) (file string, line int) {
	_, file, line, _ = runtime.Caller(skip)
	// Trim the file path for readability
	for i := len(file) - 1; i > 0; i-- {
		if file[i] == '/' {
			file = file[i+1:]
			break
		}
	}
	return
}

// formatMessage formats the log message according to the configured format
func formatMessage(level Level, file string, line int, msg string, metadata map[string]any, err error) string {
	timestamp := formatTime(cfg.TimeFormat)

	if cfg.Format == JSONFormat {
		entry := map[string]any{
			"time":     timestamp,
			"level":    level.String(),
			"file":     file,
			"line":     line,
			"message":  msg,
			"metadata": metadata,
		}
		if err != nil {
			entry["error"] = err.Error()
		}
		jsonBytes, _ := json.Marshal(entry)
		return string(jsonBytes)
	}

	// Text format: [time][level][file][line]: message, metadata, error
	metaStr := ""
	if len(metadata) > 0 {
		metaBytes, _ := json.Marshal(metadata)
		metaStr = ", " + string(metaBytes)
	}

	errStr := ""
	if err != nil {
		errStr = ", " + err.Error()
	}

	return fmt.Sprintf("[%s][%s][%s][%d]: %s%s%s", timestamp, level.String(), file, line, msg, metaStr, errStr)
}

// log writes a log message with the given level
func log(level Level, msg string, metadata map[string]any, err error) {
	if !shouldLog(level) {
		return
	}

	file, line := getCaller(3) // Skip: log, level function, caller
	formatted := formatMessage(level, file, line, msg, metadata, err)
	logFunc("%s", formatted)
}

// Debug logs a debug message
func Debug(msg string, metadata map[string]any) {
	log(DebugLevel, msg, metadata, nil)
}

// DebugError logs a debug message with an error
func DebugError(msg string, err error, metadata map[string]any) {
	log(DebugLevel, msg, metadata, err)
}

// Info logs an info message
func Info(msg string, metadata map[string]any) {
	log(InfoLevel, msg, metadata, nil)
}

// InfoError logs an info message with an error
func InfoError(msg string, err error, metadata map[string]any) {
	log(InfoLevel, msg, metadata, err)
}

// Warn logs a warning message
func Warn(msg string, metadata map[string]any) {
	log(WarnLevel, msg, metadata, nil)
}

// WarnError logs a warning message with an error
func WarnError(msg string, err error, metadata map[string]any) {
	log(WarnLevel, msg, metadata, err)
}

// Error logs an error message
func Error(msg string, metadata map[string]any) {
	log(ErrorLevel, msg, metadata, nil)
}

// ErrorError logs an error message with an error
func ErrorError(msg string, err error, metadata map[string]any) {
	log(ErrorLevel, msg, metadata, err)
}
