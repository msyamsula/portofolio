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

// otelSeverity maps internal log level to OTLP severity
func otelSeverity(level Level) otellog.Severity {
	switch level {
	case DebugLevel:
		return otellog.SeverityDebug
	case InfoLevel:
		return otellog.SeverityInfo
	case WarnLevel:
		return otellog.SeverityWarn
	case ErrorLevel:
		return otellog.SeverityError
	default:
		return otellog.SeverityInfo
	}
}

// SetOTELLogger sets an OpenTelemetry logger as the output destination
// This allows formatted logs to be exported via OTLP while preserving console output
func SetOTELLogger(otelLog otellog.Logger) {
	otelLogWriter = &otelWriter{logger: otelLog}
}

// emitStructured sends a structured log record via OpenTelemetry
func (w *otelWriter) emitStructured(ctx context.Context, level Level, msg string, file string, line int, metadata map[string]any, err error) {
	var r otellog.Record
	r.SetBody(otellog.StringValue(msg))
	r.SetSeverity(otelSeverity(level))
	r.SetSeverityText(level.String())

	// Add structured attributes
	attrs := []otellog.KeyValue{
		otellog.String("code.filepath", file),
		otellog.Int("code.lineno", line),
	}

	for k, v := range metadata {
		attrs = append(attrs, otellog.String(k, fmt.Sprintf("%v", v)))
	}

	if err != nil {
		attrs = append(attrs, otellog.String("error.message", err.Error()))
	}

	r.AddAttributes(attrs...)
	w.logger.Emit(ctx, r)
}

// ResetToStdout resets the log output back to stdout (used for testing)
func ResetToStdout() {
	otelLogWriter = nil
	logFunc = func(format string, args ...any) {
		fmt.Printf(format+"\n", args...)
	}
}
