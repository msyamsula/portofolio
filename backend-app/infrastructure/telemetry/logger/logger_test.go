package logger

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewLogger(t *testing.T) {
	cfg := ConsoleConfig{
		Level:      DebugLevel,
		Format:     TextFormat,
		TimeFormat: "2006-01-02T15:04:05Z",
	}

	logger := New(cfg)
	assert.NotNil(t, logger)
	assert.Equal(t, cfg.Level, logger.cfg.Level)
	assert.Equal(t, cfg.Format, logger.cfg.Format)
	assert.Equal(t, cfg.TimeFormat, logger.cfg.TimeFormat)
}

func TestDefaultLogger(t *testing.T) {
	logger := Default()
	assert.NotNil(t, logger)
	assert.Equal(t, InfoLevel, logger.cfg.Level)
	assert.Equal(t, TextFormat, logger.cfg.Format)
}

func TestLevelString(t *testing.T) {
	tests := []struct {
		level    Level
		expected string
	}{
		{DebugLevel, "DEBUG"},
		{InfoLevel, "INFO"},
		{WarnLevel, "WARN"},
		{ErrorLevel, "ERROR"},
		{Level(99), "UNKNOWN"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.level.String())
		})
	}
}

func TestLoggerInfo(t *testing.T) {
	originalLogFunc := logFunc
	originalFormatTime := formatTime
	originalGetCaller := getCaller
	defer func() {
		logFunc = originalLogFunc
		formatTime = originalFormatTime
		getCaller = originalGetCaller
	}()

	var capturedOutput string
	logFunc = func(format string, args ...interface{}) {
		capturedOutput = fmt.Sprintf(format, args...)
	}

	formatTime = func(layout string) string {
		return "2024-01-01T12:00:00Z"
	}

	getCaller = func(skip int) (file string, line int) {
		return "logger_test.go", 42
	}

	logger := New(ConsoleConfig{Level: InfoLevel, Format: TextFormat})
	logger.Info("test message", map[string]interface{}{"key": "value"})

	assert.Contains(t, capturedOutput, "[2024-01-01T12:00:00Z]")
	assert.Contains(t, capturedOutput, "[logger_test.go]")
	assert.Contains(t, capturedOutput, "[42]")
	assert.Contains(t, capturedOutput, "test message")
	assert.Contains(t, capturedOutput, `"key":"value"`)
}

func TestLoggerErrorWithMetadata(t *testing.T) {
	originalLogFunc := logFunc
	originalFormatTime := formatTime
	originalGetCaller := getCaller
	defer func() {
		logFunc = originalLogFunc
		formatTime = originalFormatTime
		getCaller = originalGetCaller
	}()

	var capturedOutput string
	logFunc = func(format string, args ...interface{}) {
		capturedOutput = fmt.Sprintf(format, args...)
	}

	formatTime = func(layout string) string {
		return "2024-01-01T12:00:00Z"
	}

	getCaller = func(skip int) (file string, line int) {
		return "app.go", 100
	}

	testErr := errors.New("test error")
	logger := New(ConsoleConfig{Level: InfoLevel, Format: TextFormat})
	logger.ErrorError("operation failed", testErr, map[string]interface{}{"user_id": 123})

	assert.Contains(t, capturedOutput, "[2024-01-01T12:00:00Z]")
	assert.Contains(t, capturedOutput, "[app.go]")
	assert.Contains(t, capturedOutput, "[100]")
	assert.Contains(t, capturedOutput, "operation failed")
	assert.Contains(t, capturedOutput, `"user_id":123`)
	assert.Contains(t, capturedOutput, "test error")
}

func TestLoggerDebugDisabled(t *testing.T) {
	originalLogFunc := logFunc
	defer func() { logFunc = originalLogFunc }()

	called := false
	logFunc = func(format string, args ...interface{}) {
		called = true
	}

	logger := New(ConsoleConfig{Level: InfoLevel, Format: TextFormat})
	logger.Debug("debug message", nil)

	assert.False(t, called, "Debug should not be logged when level is Info")
}

func TestLoggerDebugEnabled(t *testing.T) {
	originalLogFunc := logFunc
	defer func() { logFunc = originalLogFunc }()

	called := false
	logFunc = func(format string, args ...interface{}) {
		called = true
	}

	logger := New(ConsoleConfig{Level: DebugLevel, Format: TextFormat})
	logger.Debug("debug message", nil)

	assert.True(t, called, "Debug should be logged when level is Debug")
}

func TestLoggerJSONFormat(t *testing.T) {
	originalLogFunc := logFunc
	originalFormatTime := formatTime
	originalGetCaller := getCaller
	defer func() {
		logFunc = originalLogFunc
		formatTime = originalFormatTime
		getCaller = originalGetCaller
	}()

	var capturedOutput string
	logFunc = func(format string, args ...interface{}) {
		capturedOutput = fmt.Sprintf(format, args...)
	}

	formatTime = func(layout string) string {
		return "2024-01-01T12:00:00Z"
	}

	getCaller = func(skip int) (file string, line int) {
		return "handler.go", 50
	}

	testErr := errors.New("db error")
	logger := New(ConsoleConfig{Level: InfoLevel, Format: JSONFormat})
	logger.InfoError("database query failed", testErr, map[string]interface{}{
		"query": "SELECT * FROM users",
		"rows":  0,
	})

	var logEntry map[string]interface{}
	err := json.Unmarshal([]byte(capturedOutput), &logEntry)
	require.NoError(t, err)

	assert.Equal(t, "2024-01-01T12:00:00Z", logEntry["time"])
	assert.Equal(t, "INFO", logEntry["level"])
	assert.Equal(t, "handler.go", logEntry["file"])
	assert.Equal(t, float64(50), logEntry["line"]) // JSON numbers are float64
	assert.Equal(t, "database query failed", logEntry["message"])
	assert.Equal(t, "db error", logEntry["error"])

	metadata, ok := logEntry["metadata"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "SELECT * FROM users", metadata["query"])
	assert.Equal(t, float64(0), metadata["rows"])
}

func TestLoggerWithNilMetadata(t *testing.T) {
	originalLogFunc := logFunc
	originalFormatTime := formatTime
	originalGetCaller := getCaller
	defer func() {
		logFunc = originalLogFunc
		formatTime = originalFormatTime
		getCaller = originalGetCaller
	}()

	var capturedOutput string
	logFunc = func(format string, args ...interface{}) {
		capturedOutput = fmt.Sprintf(format, args...)
	}

	formatTime = func(layout string) string {
		return "2024-01-01T12:00:00Z"
	}

	getCaller = func(skip int) (file string, line int) {
		return "main.go", 1
	}

	logger := New(ConsoleConfig{Level: InfoLevel, Format: TextFormat})
	logger.Info("simple message", nil)

	assert.Contains(t, capturedOutput, "[2024-01-01T12:00:00Z]")
	assert.Contains(t, capturedOutput, "[main.go]")
	assert.Contains(t, capturedOutput, "[1]")
	assert.Contains(t, capturedOutput, "simple message")
	// Should not have trailing comma for empty metadata
	assert.NotContains(t, capturedOutput, ", ,")
}

func TestLoggerWarn(t *testing.T) {
	originalLogFunc := logFunc
	originalFormatTime := formatTime
	originalGetCaller := getCaller
	defer func() {
		logFunc = originalLogFunc
		formatTime = originalFormatTime
		getCaller = originalGetCaller
	}()

	var buf bytes.Buffer
	logFunc = func(format string, args ...interface{}) {
		buf.WriteString(fmt.Sprintf(format, args...))
	}

	formatTime = func(layout string) string {
		return "2024-01-01T12:00:00Z"
	}

	getCaller = func(skip int) (file string, line int) {
		return "service.go", 200
	}

	logger := New(ConsoleConfig{Level: WarnLevel, Format: TextFormat})
	logger.Warn("deprecated API usage", map[string]interface{}{"api": "v1"})

	output := buf.String()
	fmt.Println(output)
	assert.Contains(t, output, "deprecated API usage")
	assert.Contains(t, output, `"api":"v1"`)
}

func TestLoggerLevelFiltering(t *testing.T) {
	originalLogFunc := logFunc
	defer func() { logFunc = originalLogFunc }()

	var logCalls []string
	logFunc = func(format string, args ...interface{}) {
		logCalls = append(logCalls, format)
	}

	logger := New(ConsoleConfig{Level: WarnLevel, Format: TextFormat})

	logger.Debug("debug", nil)
	logger.Info("info", nil)
	logger.Warn("warn", nil)
	logger.Error("error", nil)

	assert.Len(t, logCalls, 2, "Only Warn and Error should be logged")
}

func TestLoggerGetCaller(t *testing.T) {
	originalGetCaller := getCaller
	defer func() { getCaller = originalGetCaller }()

	getCaller = func(skip int) (file string, line int) {
		assert.Equal(t, 3, skip)
		return "test_file.go", 999
	}

	logger := New(ConsoleConfig{Level: InfoLevel, Format: TextFormat})
	logger.Info("test", nil)
}

func TestLoggerComplexMetadata(t *testing.T) {
	originalLogFunc := logFunc
	originalFormatTime := formatTime
	originalGetCaller := getCaller
	defer func() {
		logFunc = originalLogFunc
		formatTime = originalFormatTime
		getCaller = originalGetCaller
	}()

	var capturedOutput string
	logFunc = func(format string, args ...interface{}) {
		capturedOutput = fmt.Sprintf(format, args...)
	}

	formatTime = func(layout string) string {
		return "2024-01-01T12:00:00Z"
	}

	getCaller = func(skip int) (file string, line int) {
		return "api.go", 10
	}

	metadata := map[string]interface{}{
		"user_id":  12345,
		"username": "testuser",
		"active":   true,
		"balance":  99.99,
		"tags":     []string{"vip", "verified"},
		"nested":   map[string]string{"key": "value"},
	}

	logger := New(ConsoleConfig{Level: InfoLevel, Format: TextFormat})
	logger.Info("user login", metadata)

	assert.Contains(t, capturedOutput, "user login")
	assert.Contains(t, capturedOutput, `"user_id":12345`)
	assert.Contains(t, capturedOutput, `"username":"testuser"`)
}

func TestLoggerErrorWithoutMetadata(t *testing.T) {
	originalLogFunc := logFunc
	originalFormatTime := formatTime
	originalGetCaller := getCaller
	defer func() {
		logFunc = originalLogFunc
		formatTime = originalFormatTime
		getCaller = originalGetCaller
	}()

	var capturedOutput string
	logFunc = func(format string, args ...interface{}) {
		capturedOutput = fmt.Sprintf(format, args...)
	}

	formatTime = func(layout string) string {
		return "2024-01-01T12:00:00Z"
	}

	getCaller = func(skip int) (file string, line int) {
		return "worker.go", 25
	}

	testErr := errors.New("connection timeout")
	logger := New(ConsoleConfig{Level: ErrorLevel, Format: TextFormat})
	logger.ErrorError("request failed", testErr, nil)

	assert.Contains(t, capturedOutput, "request failed")
	assert.Contains(t, capturedOutput, "connection timeout")
}
