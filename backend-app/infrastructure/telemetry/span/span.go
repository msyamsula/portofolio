package span

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// Config holds the trace configuration
type Config struct {
	ServiceName       string
	CollectorEndpoint string
	Insecure          bool
	SampleRate        float64
	Environment       string
}

// newExporter creates a new OTLP gRPC exporter
var newExporter = func(ctx context.Context, endpoint string, insecure bool) (sdktrace.SpanExporter, error) {
	opts := []otlptracegrpc.Option{
		otlptracegrpc.WithEndpoint(endpoint),
	}

	if insecure {
		opts = append(opts, otlptracegrpc.WithInsecure())
	}

	return otlptracegrpc.New(ctx, opts...)
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

// newTracerProvider creates a new tracer provider
var newTracerProvider = func(exporter sdktrace.SpanExporter, res *resource.Resource, sampleRate float64) *sdktrace.TracerProvider {
	return sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.TraceIDRatioBased(sampleRate)),
	)
}

// setGlobalTracerProvider sets the global tracer provider
var setGlobalTracerProvider = func(provider *sdktrace.TracerProvider) {
	otel.SetTracerProvider(provider)
}

// setGlobalPropagator is a no-op for testing purposes
// The propagator is set by the unified telemetry client in otel.go
var setGlobalPropagator = func() {
	// No-op - propagator is set by the unified client
}

// Client represents a trace client
type Client struct {
	TracerProvider *sdktrace.TracerProvider
	shutdownFunc   func(ctx context.Context) error
}

// NewClient creates a new trace client
func NewClient(ctx context.Context, cfg Config) (*Client, error) {
	exporter, err := newExporter(ctx, cfg.CollectorEndpoint, cfg.Insecure)
	if err != nil {
		return nil, fmt.Errorf("failed to create exporter: %w", err)
	}

	res, err := newResource(ctx, cfg.ServiceName, cfg.Environment)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	tracerProvider := newTracerProvider(exporter, res, cfg.SampleRate)

	setGlobalTracerProvider(tracerProvider)

	return &Client{
		TracerProvider: tracerProvider,
		shutdownFunc: func(ctx context.Context) error {
			return tracerProvider.Shutdown(ctx)
		},
	}, nil
}

// Shutdown gracefully shuts down the trace client
func (c *Client) Shutdown(ctx context.Context) error {
	if c.shutdownFunc != nil {
		return c.shutdownFunc(ctx)
	}
	return nil
}

// Tracer returns a tracer from the provider
func (c *Client) Tracer(name string) any {
	return otel.Tracer(name)
}
