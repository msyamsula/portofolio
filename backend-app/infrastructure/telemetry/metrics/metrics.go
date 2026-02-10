package metrics

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	metricSDK "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

// Config holds the metrics configuration
type Config struct {
	ServiceName       string
	CollectorEndpoint string
	Insecure          bool
	PushInterval      time.Duration
	Environment       string
}

// newMetricExporter creates a new OTLP metric exporter
var newMetricExporter = func(ctx context.Context, endpoint string, insecure bool) (metricSDK.Exporter, error) {
	opts := []otlpmetricgrpc.Option{
		otlpmetricgrpc.WithEndpoint(endpoint),
	}

	if insecure {
		opts = append(opts, otlpmetricgrpc.WithInsecure())
	}

	return otlpmetricgrpc.New(ctx, opts...)
}

// newResource creates a new OpenTelemetry resource with standard attributes
var newResource = func(ctx context.Context, serviceName, environment string) (*resource.Resource, error) {
	return resource.New(
		ctx,
		resource.WithAttributes(
			semconv.ServiceName(serviceName),
			semconv.ServiceVersion(DefaultVersion),
			semconv.DeploymentEnvironment(environment),
		),
	)
}

// DefaultVersion is the default service version
const DefaultVersion = "v1.0.0"

// newPeriodicReader creates a new periodic reader
var newPeriodicReader = func(exporter metricSDK.Exporter, interval time.Duration) metricSDK.Reader {
	return metricSDK.NewPeriodicReader(
		exporter,
		metricSDK.WithInterval(interval),
	)
}

// newMeterProvider creates a new meter provider
var newMeterProvider = func(reader metricSDK.Reader, res *resource.Resource) *metricSDK.MeterProvider {
	return metricSDK.NewMeterProvider(
		metricSDK.WithResource(res),
		metricSDK.WithReader(reader),
	)
}

// setGlobalMeterProvider sets the global meter provider
var setGlobalMeterProvider = func(mp *metricSDK.MeterProvider) {
	otel.SetMeterProvider(mp)
}

// Client represents a metrics client
type Client struct {
	MeterProvider *metricSDK.MeterProvider
	shutdownFunc  func(ctx context.Context) error
}

// NewClient creates a new metrics client
func NewClient(ctx context.Context, cfg Config) (*Client, error) {
	exporter, err := newMetricExporter(ctx, cfg.CollectorEndpoint, cfg.Insecure)
	if err != nil {
		return nil, fmt.Errorf("failed to create metric exporter: %w", err)
	}

	res, err := newResource(ctx, cfg.ServiceName, cfg.Environment)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	reader := newPeriodicReader(exporter, cfg.PushInterval)
	meterProvider := newMeterProvider(reader, res)

	setGlobalMeterProvider(meterProvider)

	return &Client{
		MeterProvider: meterProvider,
		shutdownFunc: func(ctx context.Context) error {
			return meterProvider.Shutdown(ctx)
		},
	}, nil
}

// Shutdown gracefully shuts down the metrics client
func (c *Client) Shutdown(ctx context.Context) error {
	if c.shutdownFunc != nil {
		return c.shutdownFunc(ctx)
	}
	return nil
}

// Meter returns a meter from the provider
func (c *Client) Meter(name string) any {
	return otel.Meter(name)
}
