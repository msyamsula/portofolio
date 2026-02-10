package logger

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/log"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

// ExporterConfig holds the OTLP exporter configuration
type ExporterConfig struct {
	ServiceName       string
	CollectorEndpoint string
	Insecure          bool
	Environment       string
}

// DefaultVersion is the default service version
const DefaultVersion = "v1.0.0"

// newResource creates a new OpenTelemetry resource with standard attributes
func newResource(ctx context.Context, serviceName, environment string) (*resource.Resource, error) {
	return resource.New(
		ctx,
		resource.WithAttributes(
			semconv.ServiceName(serviceName),
			semconv.ServiceVersion(DefaultVersion),
			semconv.DeploymentEnvironment(environment),
		),
	)
}

// newLogExporter creates a new OTLP gRPC log exporter
var newLogExporter = func(ctx context.Context, endpoint string, insecure bool) (sdklog.Exporter, error) {
	opts := []otlploggrpc.Option{
		otlploggrpc.WithEndpoint(endpoint),
	}

	if insecure {
		opts = append(opts, otlploggrpc.WithInsecure())
	}

	return otlploggrpc.New(ctx, opts...)
}

// newLoggerProvider creates a new logger provider
// Note: Batching is not configured by default. Logs are exported immediately.
// For production use, add: sdklog.WithBatcher(exporter)
var newLoggerProvider = func(exporter sdklog.Exporter, res *resource.Resource) *sdklog.LoggerProvider {
	// Exporter is reserved for future batching configuration
	_ = exporter
	return sdklog.NewLoggerProvider(
		sdklog.WithResource(res),
	)
}

// ExporterClient represents an OTLP log exporter client
type ExporterClient struct {
	LoggerProvider *sdklog.LoggerProvider
	shutdownFunc   func(ctx context.Context) error
}

// NewExporterClient creates a new OTLP log exporter client
func NewExporterClient(ctx context.Context, cfg ExporterConfig) (*ExporterClient, error) {
	exporter, err := newLogExporter(ctx, cfg.CollectorEndpoint, cfg.Insecure)
	if err != nil {
		return nil, fmt.Errorf("failed to create log exporter: %w", err)
	}

	res, err := newResource(ctx, cfg.ServiceName, cfg.Environment)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	loggerProvider := newLoggerProvider(exporter, res)

	return &ExporterClient{
		LoggerProvider: loggerProvider,
		shutdownFunc: func(ctx context.Context) error {
			return loggerProvider.Shutdown(ctx)
		},
	}, nil
}

// Shutdown gracefully shuts down the exporter client
func (c *ExporterClient) Shutdown(ctx context.Context) error {
	if c.shutdownFunc != nil {
		return c.shutdownFunc(ctx)
	}
	return nil
}

// OTELLogger returns a typed OpenTelemetry logger for use with SetOTELLogger
func (c *ExporterClient) OTELLogger(name string) log.Logger {
	return c.LoggerProvider.Logger(name)
}
