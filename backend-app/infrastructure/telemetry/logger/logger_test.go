package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	otellog "go.opentelemetry.io/otel/log"
)

// mockOTELLogger is a mock OpenTelemetry logger for testing
type mockOTELLogger struct {
	otellog.Logger
	emitCalled bool
	emitBody   string
}

func (m *mockOTELLogger) Emit(ctx context.Context, r otellog.Record) {
	m.emitCalled = true
	m.emitBody = r.Body().AsString()
}

func resetLoggerState() {
	// Reset to default console logger
	ResetToStdout()
}

func TestInit(t *testing.T) {
	resetLoggerState()
	ctx := context.Background()

	cfg := Config{
		ServiceName:       "test-service",
		CollectorEndpoint: "localhost:4317",
		Insecure:         true,
		Environment:      "test",
		LogsEnabled:      false, // Don't try to connect to OTLP in tests
		Level:            InfoLevel,
		Format:           TextFormat,
		TimeFormat:       "2006-01-02T15:04:05Z",
	}

	err := Init(ctx, cfg)
	assert.NoError(t, err)
}

func TestShutdown(t *testing.T) {
	resetLoggerState()
	ctx := context.Background()

	cfg := Config{
		ServiceName:  "test-service",
		LogsEnabled: false,
		Level:       InfoLevel,
	}

	err := Init(ctx, cfg)
	require.NoError(t, err)

	err = Shutdown(ctx)
	assert.NoError(t, err)
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

func TestInfo(t *testing.T) {
	resetLoggerState()
	originalLogFunc := logFunc
	originalFormatTime := formatTime
	originalGetCaller := getCaller
	defer func() {
		logFunc = originalLogFunc
		formatTime = originalFormatTime
		getCaller = originalGetCaller
	}()

	var capturedOutput string
	logFunc = func(format string, args ...any) {
		capturedOutput = fmt.Sprintf(format, args...)
	}

	formatTime = func(layout string) string {
		return "2024-01-01T12:00:00Z"
	}

	getCaller = func(skip int) (file string, line int) {
		return "logger_test.go", 42
	}

	// Initialize with Info level
	cfg := Config{Level: InfoLevel, Format: TextFormat}
	_ = Init(context.Background(), cfg)

	Info("test message", map[string]any{"key": "value"})

	assert.Contains(t, capturedOutput, "[2024-01-01T12:00:00Z]")
	assert.Contains(t, capturedOutput, "[INFO]")
	assert.Contains(t, capturedOutput, "[logger_test.go]")
	assert.Contains(t, capturedOutput, "[42]")
	assert.Contains(t, capturedOutput, "test message")
	assert.Contains(t, capturedOutput, `"key":"value"`)
}

func TestError(t *testing.T) {
	resetLoggerState()
	originalLogFunc := logFunc
	originalFormatTime := formatTime
	originalGetCaller := getCaller
	defer func() {
		logFunc = originalLogFunc
		formatTime = originalFormatTime
		getCaller = originalGetCaller
	}()

	var capturedOutput string
	logFunc = func(format string, args ...any) {
		capturedOutput = fmt.Sprintf(format, args...)
	}

	formatTime = func(layout string) string {
		return "2024-01-01T12:00:00Z"
	}

	getCaller = func(skip int) (file string, line int) {
		return "app.go", 100
	}

	testErr := errors.New("test error")
	cfg := Config{Level: InfoLevel, Format: TextFormat}
	_ = Init(context.Background(), cfg)

	Error("operation failed", map[string]any{"user_id": 123, "error": testErr})

	assert.Contains(t, capturedOutput, "[2024-01-01T12:00:00Z]")
	assert.Contains(t, capturedOutput, "[ERROR]")
	assert.Contains(t, capturedOutput, "[app.go]")
	assert.Contains(t, capturedOutput, "[100]")
	assert.Contains(t, capturedOutput, "operation failed")
	assert.Contains(t, capturedOutput, `"user_id":123`)
}

func TestErrorError(t *testing.T) {
	resetLoggerState()
	originalLogFunc := logFunc
	originalFormatTime := formatTime
	originalGetCaller := getCaller
	defer func() {
		logFunc = originalLogFunc
		formatTime = originalFormatTime
		getCaller = originalGetCaller
	}()

	var capturedOutput string
	logFunc = func(format string, args ...any) {
		capturedOutput = fmt.Sprintf(format, args...)
	}

	formatTime = func(layout string) string {
		return "2024-01-01T12:00:00Z"
	}

	getCaller = func(skip int) (file string, line int) {
		return "service.go", 200
	}

	testErr := errors.New("connection failed")
	cfg := Config{Level: InfoLevel, Format: TextFormat}
	_ = Init(context.Background(), cfg)

	ErrorError("database query failed", testErr, map[string]any{"query": "SELECT * FROM users"})

	assert.Contains(t, capturedOutput, "[2024-01-01T12:00:00Z]")
	assert.Contains(t, capturedOutput, "[ERROR]")
	assert.Contains(t, capturedOutput, "[service.go]")
	assert.Contains(t, capturedOutput, "[200]")
	assert.Contains(t, capturedOutput, "database query failed")
	assert.Contains(t, capturedOutput, `"query":"SELECT * FROM users"`)
	assert.Contains(t, capturedOutput, "connection failed")
}

func TestDebug(t *testing.T) {
	resetLoggerState()
	originalLogFunc := logFunc
	defer func() { logFunc = originalLogFunc }()

	called := false
	logFunc = func(format string, args ...any) {
		called = true
	}

	cfg := Config{Level: DebugLevel, Format: TextFormat}
	_ = Init(context.Background(), cfg)

	Debug("debug message", map[string]any{"trace": "12345"})

	assert.True(t, called, "Debug should be logged when level is Debug")
}

func TestDebugDisabled(t *testing.T) {
	resetLoggerState()
	originalLogFunc := logFunc
	defer func() { logFunc = originalLogFunc }()

	called := false
	logFunc = func(format string, args ...any) {
		called = true
	}

	cfg := Config{Level: InfoLevel, Format: TextFormat}
	_ = Init(context.Background(), cfg)

	Debug("debug message", nil)

	assert.False(t, called, "Debug should not be logged when level is Info")
}

func TestWarn(t *testing.T) {
	resetLoggerState()
	originalLogFunc := logFunc
	originalFormatTime := formatTime
	defer func() {
		logFunc = originalLogFunc
		formatTime = originalFormatTime
	}()

	var buf bytes.Buffer
	logFunc = func(format string, args ...any) {
		buf.WriteString(fmt.Sprintf(format, args...) + "\n")
	}

	formatTime = func(layout string) string {
		return "2024-01-01T12:00:00Z"
	}

	cfg := Config{Level: WarnLevel, Format: TextFormat}
	_ = Init(context.Background(), cfg)

	Warn("deprecated API usage", map[string]any{"api": "v1"})

	output := buf.String()
	assert.Contains(t, output, "[WARN]")
	assert.Contains(t, output, "deprecated API usage")
	assert.Contains(t, output, `"api":"v1"`)
}

func TestWarnError(t *testing.T) {
	resetLoggerState()
	originalLogFunc := logFunc
	originalFormatTime := formatTime
	defer func() {
		logFunc = originalLogFunc
		formatTime = originalFormatTime
	}()

	var capturedOutput string
	logFunc = func(format string, args ...any) {
		capturedOutput = fmt.Sprintf(format, args...)
	}

	formatTime = func(layout string) string {
		return "2024-01-01T12:00:00Z"
	}

	getCaller = func(skip int) (file string, line int) {
		return "cache.go", 50
	}

	testErr := errors.New("cache miss")
	cfg := Config{Level: InfoLevel, Format: TextFormat}
	_ = Init(context.Background(), cfg)

	WarnError("retry operation", testErr, map[string]any{"attempt": 3, "key": "user:123"})

	assert.Contains(t, capturedOutput, "[WARN]")
	assert.Contains(t, capturedOutput, "retry operation")
	assert.Contains(t, capturedOutput, `"attempt":3`)
	assert.Contains(t, capturedOutput, `"key":"user:123"`)
	assert.Contains(t, capturedOutput, "cache miss")
}

func TestJSONFormat(t *testing.T) {
	resetLoggerState()
	originalLogFunc := logFunc
	originalFormatTime := formatTime
	originalGetCaller := getCaller
	defer func() {
		logFunc = originalLogFunc
		formatTime = originalFormatTime
		getCaller = originalGetCaller
	}()

	var capturedOutput string
	logFunc = func(format string, args ...any) {
		capturedOutput = fmt.Sprintf(format, args...)
	}

	formatTime = func(layout string) string {
		return "2024-01-01T12:00:00Z"
	}

	getCaller = func(skip int) (file string, line int) {
		return "handler.go", 50
	}

	testErr := errors.New("db error")
	cfg := Config{Level: InfoLevel, Format: JSONFormat}
	_ = Init(context.Background(), cfg)

	InfoError("database query failed", testErr, map[string]any{
		"query": "SELECT * FROM users",
		"rows":  0,
	})

	var logEntry map[string]any
	err := json.Unmarshal([]byte(capturedOutput), &logEntry)
	require.NoError(t, err)

	assert.Equal(t, "2024-01-01T12:00:00Z", logEntry["time"])
	assert.Equal(t, "INFO", logEntry["level"])
	assert.Equal(t, "handler.go", logEntry["file"])
	assert.Equal(t, float64(50), logEntry["line"]) // JSON numbers are float64
	assert.Equal(t, "database query failed", logEntry["message"])
	assert.Equal(t, "db error", logEntry["error"])

	metadata, ok := logEntry["metadata"].(map[string]any)
	assert.True(t, ok)
	assert.Equal(t, "SELECT * FROM users", metadata["query"])
	assert.Equal(t, float64(0), metadata["rows"])
}

func TestWithNilMetadata(t *testing.T) {
	resetLoggerState()
	originalLogFunc := logFunc
	originalFormatTime := formatTime
	originalGetCaller := getCaller
	defer func() {
		logFunc = originalLogFunc
		formatTime = originalFormatTime
		getCaller = originalGetCaller
	}()

	var capturedOutput string
	logFunc = func(format string, args ...any) {
		capturedOutput = fmt.Sprintf(format, args...)
	}

	formatTime = func(layout string) string {
		return "2024-01-01T12:00:00Z"
	}

	getCaller = func(skip int) (file string, line int) {
		return "main.go", 1
	}

	cfg := Config{Level: InfoLevel, Format: TextFormat}
	_ = Init(context.Background(), cfg)

	Info("simple message", nil)

	assert.Contains(t, capturedOutput, "[2024-01-01T12:00:00Z]")
	assert.Contains(t, capturedOutput, "[main.go]")
	assert.Contains(t, capturedOutput, "[1]")
	assert.Contains(t, capturedOutput, "simple message")
	// Should not have trailing comma for empty metadata
	assert.NotContains(t, capturedOutput, ", ,")
}

func TestLevelFiltering(t *testing.T) {
	resetLoggerState()
	originalLogFunc := logFunc
	defer func() { logFunc = originalLogFunc }()

	var logCalls []string
	logFunc = func(format string, args ...any) {
		logCalls = append(logCalls, format)
	}

	cfg := Config{Level: WarnLevel, Format: TextFormat}
	_ = Init(context.Background(), cfg)

	Debug("debug", nil)
	Info("info", nil)
	Warn("warn", nil)
	Error("error", nil)

	assert.Len(t, logCalls, 2, "Only Warn and Error should be logged")
}

func TestGetCaller(t *testing.T) {
	resetLoggerState()
	originalGetCaller := getCaller
	originalLogFunc := logFunc
	defer func() {
		getCaller = originalGetCaller
		logFunc = originalLogFunc
	}()

	getCaller = func(skip int) (file string, line int) {
		assert.Equal(t, 3, skip, "Should skip 3 frames: log, level function, caller")
		return "test_file.go", 999
	}

	called := false
	logFunc = func(format string, args ...any) {
		called = true
	}

	cfg := Config{Level: InfoLevel, Format: TextFormat}
	_ = Init(context.Background(), cfg)

	Info("test", nil)
	assert.True(t, called, "logFunc should be called")
}

func TestComplexMetadata(t *testing.T) {
	resetLoggerState()
	originalLogFunc := logFunc
	originalFormatTime := formatTime
	defer func() {
		logFunc = originalLogFunc
		formatTime = originalFormatTime
	}()

	var capturedOutput string
	logFunc = func(format string, args ...any) {
		capturedOutput = fmt.Sprintf(format, args...)
	}

	formatTime = func(layout string) string {
		return "2024-01-01T12:00:00Z"
	}

	getCaller = func(skip int) (file string, line int) {
		return "api.go", 10
	}

	metadata := map[string]any{
		"user_id": 12345,
		"username": "testuser",
		"active":   true,
		"balance":  99.99,
		"tags":     []string{"vip", "verified"},
		"nested":   map[string]string{"key": "value"},
	}

	cfg := Config{Level: InfoLevel, Format: TextFormat}
	_ = Init(context.Background(), cfg)

	Info("user login", metadata)

	assert.Contains(t, capturedOutput, "user login")
	assert.Contains(t, capturedOutput, `"user_id":12345`)
	assert.Contains(t, capturedOutput, `"username":"testuser"`)
	assert.Contains(t, capturedOutput, `"active":true`)
	assert.Contains(t, capturedOutput, `"balance":99.99`)
}

func TestDualOutput(t *testing.T) {
	resetLoggerState()

	// Test that SetOTELLogger preserves console output
	// We verify this by checking the ResetToStdout function restores console output
	ResetToStdout()

	// After ResetToStdout, logFunc should write to stdout
	// We can verify this works by calling a log and checking it doesn't panic
	assert.NotPanics(t, func() {
		Info("test message", map[string]any{"key": "value"})
	}, "Logging should not panic after ResetToStdout")
}

func TestResetToStdout(t *testing.T) {
	// Initialize with Debug level so Debug logs work
	cfg := Config{Level: DebugLevel, Format: TextFormat}
	_ = Init(context.Background(), cfg)

	// Set a custom logFunc
	customCalled := false
	logFunc = func(format string, args ...any) {
		customCalled = true
	}

	// Call to verify custom is active
	Debug("custom debug", nil)
	assert.True(t, customCalled, "custom logFunc should be called")

	// Reset to stdout
	ResetToStdout()

	// After reset, logging should still work (no panic)
	assert.NotPanics(t, func() {
		Info("test", nil)
	}, "Logging should not panic after ResetToStdout")
}

func TestInfoError(t *testing.T) {
	resetLoggerState()
	originalLogFunc := logFunc
	originalFormatTime := formatTime
	defer func() {
		logFunc = originalLogFunc
		formatTime = originalFormatTime
	}()

	var capturedOutput string
	logFunc = func(format string, args ...any) {
		capturedOutput = fmt.Sprintf(format, args...)
	}

	formatTime = func(layout string) string {
		return "2024-01-01T12:00:00Z"
	}

	getCaller = func(skip int) (file string, line int) {
		return "worker.go", 25
	}

	testErr := errors.New("task failed")
	cfg := Config{Level: InfoLevel, Format: TextFormat}
	_ = Init(context.Background(), cfg)

	InfoError("background job failed", testErr, map[string]any{"job_id": 456})

	assert.Contains(t, capturedOutput, "[INFO]")
	assert.Contains(t, capturedOutput, "background job failed")
	assert.Contains(t, capturedOutput, `"job_id":456`)
	assert.Contains(t, capturedOutput, "task failed")
}

func TestDebugError(t *testing.T) {
	resetLoggerState()
	originalLogFunc := logFunc
	defer func() { logFunc = originalLogFunc }()

	called := false
	logFunc = func(format string, args ...any) {
		called = true
	}

	cfg := Config{Level: DebugLevel, Format: TextFormat}
	_ = Init(context.Background(), cfg)

	testErr := errors.New("debug error")
	DebugError("tracing request", testErr, map[string]any{"request_id": "abc-123"})

	assert.True(t, called, "DebugError should be logged when level is Debug")
}
