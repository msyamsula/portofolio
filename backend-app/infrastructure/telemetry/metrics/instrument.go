package metrics

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

// Instruments holds common metric instruments
type Instruments struct {
	// Counter
	RequestCounter metric.Int64Counter

	// Gauge
	ActiveUsersGauge metric.Int64ObservableGauge

	// Histogram
	RequestDuration metric.Float64Histogram
}

// NewInstruments creates and registers common metric instruments
func NewInstruments(meter any) (*Instruments, error) {
	m, ok := meter.(metric.Meter)
	if !ok {
		// Return empty instruments if meter is not the right type
		return &Instruments{}, nil
	}

	reqCounter, err := m.Int64Counter("http_requests_total",
		metric.WithDescription("Total number of HTTP requests"),
		metric.WithUnit("1"),
	)
	if err != nil {
		return nil, err
	}

	duration, err := m.Float64Histogram("http_request_duration_seconds",
		metric.WithDescription("HTTP request duration in seconds"),
		metric.WithUnit("s"),
	)
	if err != nil {
		return nil, err
	}

	return &Instruments{
		RequestCounter:  reqCounter,
		RequestDuration: duration,
	}, nil
}

// RecordRequest records an HTTP request metric
func (i *Instruments) RecordRequest(ctx context.Context, method, path string, status int, duration float64) {
	if i.RequestCounter == nil {
		return
	}

	attrs := []attribute.KeyValue{
		attribute.String("method", method),
		attribute.String("path", path),
		attribute.Int64("status", int64(status)),
	}

	i.RequestCounter.Add(ctx, 1, metric.WithAttributes(attrs...))

	if i.RequestDuration != nil {
		i.RequestDuration.Record(ctx, duration, metric.WithAttributes(attrs...))
	}
}

// IncrementRequestCounter increments the request counter
func (i *Instruments) IncrementRequestCounter(ctx context.Context, method, path string, status int) {
	if i.RequestCounter == nil {
		return
	}

	attrs := []attribute.KeyValue{
		attribute.String("method", method),
		attribute.String("path", path),
		attribute.Int64("status", int64(status)),
	}

	i.RequestCounter.Add(ctx, 1, metric.WithAttributes(attrs...))
}

// RecordDuration records a duration value
func (i *Instruments) RecordDuration(ctx context.Context, name string, duration float64, attrs ...attribute.KeyValue) {
	if i.RequestDuration == nil {
		return
	}

	allAttrs := append([]attribute.KeyValue{
		attribute.String("name", name),
	}, attrs...)

	i.RequestDuration.Record(ctx, duration, metric.WithAttributes(allAttrs...))
}
