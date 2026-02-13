package metrics

import (
	"context"
	"sync/atomic"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

// Instruments holds common metric instruments
type Instruments struct {
	// Counter
	RequestCounter metric.Int64Counter

	// Gauge
	ActiveUsersGauge   metric.Int64ObservableGauge
	ResponseTimeGauge  metric.Float64ObservableGauge
	latestResponseTime atomic.Value

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

	instruments := &Instruments{}

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

	responseTimeGauge, err := m.Float64ObservableGauge(
		"http_response_time_latest_seconds",
		metric.WithDescription("Latest HTTP response time in seconds"),
		metric.WithUnit("s"),
		metric.WithFloat64Callback(func(ctx context.Context, observer metric.Float64Observer) error {
			if v := instruments.latestResponseTime.Load(); v != nil {
				observer.Observe(v.(float64))
			}
			return nil
		}),
	)
	if err != nil {
		return nil, err
	}

	instruments.latestResponseTime.Store(float64(0))

	instruments.RequestCounter = reqCounter
	instruments.RequestDuration = duration
	instruments.ResponseTimeGauge = responseTimeGauge

	return instruments, nil
}

// SetResponseTime sets the latest response time for gauge observation
func (i *Instruments) SetResponseTime(duration float64) {
	if i.ResponseTimeGauge == nil {
		return
	}
	i.latestResponseTime.Store(duration)
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
	if i.ResponseTimeGauge != nil {
		i.SetResponseTime(duration)
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
