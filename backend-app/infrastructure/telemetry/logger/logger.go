package logger

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime"
	"time"
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

// ConsoleConfig holds the console logger configuration
type ConsoleConfig struct {
	Level      Level
	Format     Format
	TimeFormat string
}

// Config holds common configuration for logger
type Config struct {
	// Common settings
	ServiceName       string
	CollectorEndpoint string
	Insecure          bool
	Environment       string

	// Logs settings
	LogsEnabled bool

	// Console logger settings
	Level      Level
	Format     Format
	TimeFormat string
}

// Logger represents a logger instance
type Logger struct {
	cfg ConsoleConfig
}

// New creates a new logger with the given configuration
func New(cfg ConsoleConfig) *Logger {
	if cfg.TimeFormat == "" {
		cfg.TimeFormat = time.RFC3339
	}
	return &Logger{cfg: cfg}
}

// Default creates a logger with default configuration
func Default() *Logger {
	return New(ConsoleConfig{
		Level:      InfoLevel,
		Format:     TextFormat,
		TimeFormat: time.RFC3339,
	})
}

// logFunc is the function that writes the log output
var logFunc = func(format string, args ...any) {
	fmt.Printf(format+"\n", args...)
}

// shouldLog determines if a message should be logged based on level
func (l *Logger) shouldLog(level Level) bool {
	return level >= l.cfg.Level
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
func (l *Logger) formatMessage(level Level, file string, line int, msg string, metadata map[string]any, err error) string {
	timestamp := formatTime(l.cfg.TimeFormat)

	if l.cfg.Format == JSONFormat {
		logEntry := map[string]any{
			"time":     timestamp,
			"level":    level.String(),
			"file":     file,
			"line":     line,
			"message":  msg,
			"metadata": metadata,
		}
		if err != nil {
			logEntry["error"] = err.Error()
		}
		jsonBytes, _ := json.Marshal(logEntry)
		return string(jsonBytes)
	}

	// Text format: [time][file][line]: message, metadata, error
	metaStr := ""
	if len(metadata) > 0 {
		metaBytes, _ := json.Marshal(metadata)
		metaStr = ", " + string(metaBytes)
	}

	errStr := ""
	if err != nil {
		errStr = ", " + err.Error()
	}

	return fmt.Sprintf("[%s][%s][%d]: %s%s%s", timestamp, file, line, msg, metaStr, errStr)
}

// log writes a log message with the given level
func (l *Logger) log(level Level, msg string, metadata map[string]any, err error) {
	if !l.shouldLog(level) {
		return
	}

	file, line := getCaller(3) // Skip: log, log level function, caller
	formatted := l.formatMessage(level, file, line, msg, metadata, err)
	logFunc("%s", formatted)
}

// Debug logs a debug message
func (l *Logger) Debug(msg string, metadata map[string]any) {
	l.log(DebugLevel, msg, metadata, nil)
}

// DebugError logs a debug message with an error
func (l *Logger) DebugError(msg string, err error, metadata map[string]any) {
	l.log(DebugLevel, msg, metadata, err)
}

// Info logs an info message
func (l *Logger) Info(msg string, metadata map[string]any) {
	l.log(InfoLevel, msg, metadata, nil)
}

// InfoError logs an info message with an error
func (l *Logger) InfoError(msg string, err error, metadata map[string]any) {
	l.log(InfoLevel, msg, metadata, err)
}

// Warn logs a warning message
func (l *Logger) Warn(msg string, metadata map[string]any) {
	l.log(WarnLevel, msg, metadata, nil)
}

// WarnError logs a warning message with an error
func (l *Logger) WarnError(msg string, err error, metadata map[string]any) {
	l.log(WarnLevel, msg, metadata, err)
}

// Error logs an error message
func (l *Logger) Error(msg string, metadata map[string]any) {
	l.log(ErrorLevel, msg, metadata, nil)
}

// ErrorError logs an error message with an error
func (l *Logger) ErrorError(msg string, err error, metadata map[string]any) {
	l.log(ErrorLevel, msg, metadata, err)
}

// Client is a unified logger client that combines console logging and OTLP export
type Client struct {
	// Console logger
	Logger *Logger

	// OTLP exporter (optional)
	ExporterClient *ExporterClient
	exporterEnabled bool
}

// NewClient creates a new unified logger client
func NewClient(ctx context.Context, cfg Config) (*Client, error) {
	c := &Client{
		exporterEnabled: cfg.LogsEnabled,
	}

	// Initialize console logger
	consoleLogger := New(ConsoleConfig{
		Level:      cfg.Level,
		Format:     cfg.Format,
		TimeFormat: cfg.TimeFormat,
	})
	if consoleLogger == nil {
		consoleLogger = Default()
	}
	c.Logger = consoleLogger

	// Initialize OTLP exporter if enabled
	if cfg.LogsEnabled {
		exporterCfg := ExporterConfig{
			ServiceName:       cfg.ServiceName,
			CollectorEndpoint: cfg.CollectorEndpoint,
			Insecure:          cfg.Insecure,
			Environment:       cfg.Environment,
		}
		exporterClient, err := NewExporterClient(ctx, exporterCfg)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize logs exporter: %w", err)
		}
		c.ExporterClient = exporterClient

		// Set OTEL logger for the logger package
		otelLogger := exporterClient.OTELLogger(cfg.ServiceName)
		SetOTELLogger(otelLogger)
	}

	return c, nil
}

// Shutdown gracefully shuts down the logger client
func (c *Client) Shutdown(ctx context.Context) error {
	if c.exporterEnabled && c.ExporterClient != nil {
		if err := c.ExporterClient.Shutdown(ctx); err != nil {
			return fmt.Errorf("logger shutdown: %w", err)
		}
	}
	return nil
}

// DefaultConfig returns a logger config with sensible defaults
func DefaultConfig(serviceName string) Config {
	return Config{
		ServiceName:   serviceName,
		Level:         InfoLevel,
		Format:        TextFormat,
		TimeFormat:    time.RFC3339,
		LogsEnabled:   true,
		Insecure:      true,
		Environment:   "development",
	}
}

// Package-level logger for convenience
// Use SetDefaultLogger to set a custom logger
var defaultLogger = Default()

// SetDefaultLogger sets the package-level logger
func SetDefaultLogger(l *Logger) {
	defaultLogger = l
}

// GetDefaultLogger returns the package-level logger
func GetDefaultLogger() *Logger {
	return defaultLogger
}

// Package-level convenience functions for logging

// Debug logs a debug message using the default logger
func Debug(msg string, metadata map[string]any) {
	defaultLogger.Debug(msg, metadata)
}

// DebugError logs a debug message with an error using the default logger
func DebugError(msg string, err error, metadata map[string]any) {
	defaultLogger.DebugError(msg, err, metadata)
}

// Info logs an info message using the default logger
func Info(msg string, metadata map[string]any) {
	defaultLogger.Info(msg, metadata)
}

// InfoError logs an info message with an error using the default logger
func InfoError(msg string, err error, metadata map[string]any) {
	defaultLogger.InfoError(msg, err, metadata)
}

// Warn logs a warning message using the default logger
func Warn(msg string, metadata map[string]any) {
	defaultLogger.Warn(msg, metadata)
}

// WarnError logs a warning message with an error using the default logger
func WarnError(msg string, err error, metadata map[string]any) {
	defaultLogger.WarnError(msg, err, metadata)
}

// Error logs an error message using the default logger
func Error(msg string, metadata map[string]any) {
	defaultLogger.Error(msg, metadata)
}

// ErrorError logs an error message with an error using the default logger
func ErrorError(msg string, err error, metadata map[string]any) {
	defaultLogger.ErrorError(msg, err, metadata)
}
