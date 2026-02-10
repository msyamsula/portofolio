package logger

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/log"
)

// otelWriter is a log writer that uses OpenTelemetry logs
type otelWriter struct {
	logger log.Logger
}

// SetOTELLogger sets an OpenTelemetry logger as the output destination
// This allows formatted logs to be exported via OTLP
func SetOTELLogger(otelLogger log.Logger) {
	w := &otelWriter{logger: otelLogger}
	logFunc = func(format string, args ...any) {
		msg := fmt.Sprintf(format, args...)
		w.emit(msg)
	}
}

// emit sends a log record via OpenTelemetry
func (w *otelWriter) emit(formattedLog string) {
	// Create a new log record
	var r log.Record
	r.SetBody(log.StringValue(formattedLog))
	w.logger.Emit(context.Background(), r)
}

// ResetToStdout resets the log output back to stdout
func ResetToStdout() {
	logFunc = func(format string, args ...any) {
		fmt.Printf(format+"\n", args...)
	}
}

// setLogFunc is a variable for testing purposes
var setLogFunc = func(fn func(format string, args ...any)) {
	logFunc = fn
}
