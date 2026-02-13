package logger

import (
	"context"
	"fmt"

	otellog "go.opentelemetry.io/otel/log"
)

// otelWriter is a log writer that uses OpenTelemetry logs
type otelWriter struct {
	logger otellog.Logger
}

// SetOTELLogger sets an OpenTelemetry logger as the output destination
// This allows formatted logs to be exported via OTLP while preserving console output
func SetOTELLogger(otelLog otellog.Logger) {
	w := &otelWriter{logger: otelLog}

	// First reset to get the clean console function
	ResetToStdout()
	originalLogFunc := logFunc

	// Replace logFunc with dual-output version that sends to both console and OTLP
	logFunc = func(format string, args ...any) {
		// 1. Send to console (preserves stdout)
		originalLogFunc(format, args...)
		// 2. Send to OTLP
		msg := fmt.Sprintf(format, args...)
		w.emit(msg)
	}
}

// emit sends a log record via OpenTelemetry
func (w *otelWriter) emit(formattedLog string) {
	// Create a new log record
	var r otellog.Record
	r.SetBody(otellog.StringValue(formattedLog))
	w.logger.Emit(context.Background(), r)
}

// ResetToStdout resets the log output back to stdout
func ResetToStdout() {
	logFunc = func(format string, args ...any) {
		fmt.Printf(format+"\n", args...)
	}
}
