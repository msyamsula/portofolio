package span

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

type mockExporter struct{}

func (m *mockExporter) ExportSpans(ctx context.Context, spans []sdktrace.ReadOnlySpan) error {
	return nil
}

func (m *mockExporter) Shutdown(ctx context.Context) error {
	return nil
}

func (m *mockExporter) ForceFlush(ctx context.Context) error {
	return nil
}

func TestNewClientSuccess(t *testing.T) {
	originalExporter := newExporter
	originalResource := newResource
	originalTracerProvider := newTracerProvider
	originalSetTracerProvider := setGlobalTracerProvider
	defer func() {
		newExporter = originalExporter
		newResource = originalResource
		newTracerProvider = originalTracerProvider
		setGlobalTracerProvider = originalSetTracerProvider
	}()

	ctx := context.Background()

	var capturedExporter sdktrace.SpanExporter
	var capturedRes *resource.Resource
	var capturedSampleRate float64

	newExporter = func(ctx context.Context, endpoint string, insecure bool) (sdktrace.SpanExporter, error) {
		assert.Equal(t, "localhost:4317", endpoint)
		assert.True(t, insecure)
		mockExp := &mockExporter{}
		capturedExporter = mockExp
		return mockExp, nil
	}

	newResource = func(ctx context.Context, serviceName, environment string) (*resource.Resource, error) {
		assert.Equal(t, "test-service", serviceName)
		assert.Equal(t, "test", environment)
		mockRes, _ := resource.New(ctx)
		capturedRes = mockRes
		return mockRes, nil
	}

	newTracerProvider = func(exporter sdktrace.SpanExporter, res *resource.Resource, sampleRate float64) *sdktrace.TracerProvider {
		assert.Same(t, capturedExporter, exporter)
		assert.Same(t, capturedRes, res)
		assert.Equal(t, 1.0, sampleRate)
		capturedSampleRate = sampleRate
		tp := sdktrace.NewTracerProvider(
			sdktrace.WithBatcher(exporter),
			sdktrace.WithResource(res),
			sdktrace.WithSampler(sdktrace.AlwaysSample()),
		)
		return tp
	}

	setTracerProviderCalled := false
	var capturedProvider *sdktrace.TracerProvider
	setGlobalTracerProvider = func(provider *sdktrace.TracerProvider) {
		setTracerProviderCalled = true
		capturedProvider = provider
	}

	cfg := Config{
		ServiceName:       "test-service",
		CollectorEndpoint: "localhost:4317",
		Insecure:          true,
		SampleRate:        1.0,
		Environment:       "test",
	}

	client, err := NewClient(ctx, cfg)
	require.NoError(t, err)
	assert.NotNil(t, client)
	assert.NotNil(t, client.TracerProvider)
	assert.Same(t, capturedProvider, client.TracerProvider)
	assert.True(t, setTracerProviderCalled)
	assert.Equal(t, 1.0, capturedSampleRate)
}

func TestNewClientExporterError(t *testing.T) {
	originalExporter := newExporter
	defer func() { newExporter = originalExporter }()

	ctx := context.Background()
	newExporter = func(ctx context.Context, endpoint string, insecure bool) (sdktrace.SpanExporter, error) {
		return nil, errors.New("exporter creation failed")
	}

	cfg := Config{
		ServiceName:       "test-service",
		CollectorEndpoint: "localhost:4317",
		Insecure:          true,
		SampleRate:        1.0,
		Environment:       "test",
	}

	client, err := NewClient(ctx, cfg)
	require.Error(t, err)
	assert.Nil(t, client)
	assert.Contains(t, err.Error(), "failed to create exporter")
}

func TestNewClientResourceError(t *testing.T) {
	originalExporter := newExporter
	originalResource := newResource
	defer func() {
		newExporter = originalExporter
		newResource = originalResource
	}()

	ctx := context.Background()
	newExporter = func(ctx context.Context, endpoint string, insecure bool) (sdktrace.SpanExporter, error) {
		return &mockExporter{}, nil
	}

	newResource = func(ctx context.Context, serviceName, environment string) (*resource.Resource, error) {
		return nil, errors.New("resource creation failed")
	}

	cfg := Config{
		ServiceName:       "test-service",
		CollectorEndpoint: "localhost:4317",
		Insecure:          true,
		SampleRate:        1.0,
		Environment:       "test",
	}

	client, err := NewClient(ctx, cfg)
	require.Error(t, err)
	assert.Nil(t, client)
	assert.Contains(t, err.Error(), "failed to create resource")
}

func TestClientShutdown(t *testing.T) {
	originalExporter := newExporter
	originalResource := newResource
	originalTracerProvider := newTracerProvider
	originalSetTracerProvider := setGlobalTracerProvider
	defer func() {
		newExporter = originalExporter
		newResource = originalResource
		newTracerProvider = originalTracerProvider
		setGlobalTracerProvider = originalSetTracerProvider
	}()

	ctx := context.Background()

	newExporter = func(ctx context.Context, endpoint string, insecure bool) (sdktrace.SpanExporter, error) {
		return &mockExporter{}, nil
	}

	newResource = func(ctx context.Context, serviceName, environment string) (*resource.Resource, error) {
		return resource.New(ctx)
	}

	newTracerProvider = func(exporter sdktrace.SpanExporter, res *resource.Resource, sampleRate float64) *sdktrace.TracerProvider {
		tp := sdktrace.NewTracerProvider(
			sdktrace.WithBatcher(exporter),
			sdktrace.WithResource(res),
			sdktrace.WithSampler(sdktrace.AlwaysSample()),
		)
		return tp
	}

	setGlobalTracerProvider = func(provider *sdktrace.TracerProvider) {}

	cfg := Config{
		ServiceName:       "test-service",
		CollectorEndpoint: "localhost:4317",
		Insecure:          true,
		SampleRate:        1.0,
		Environment:       "test",
	}

	client, err := NewClient(ctx, cfg)
	require.NoError(t, err)

	err = client.Shutdown(ctx)
	require.NoError(t, err)
}

func TestClientSecureConnection(t *testing.T) {
	originalExporter := newExporter
	originalResource := newResource
	originalTracerProvider := newTracerProvider
	originalSetTracerProvider := setGlobalTracerProvider
	defer func() {
		newExporter = originalExporter
		newResource = originalResource
		newTracerProvider = originalTracerProvider
		setGlobalTracerProvider = originalSetTracerProvider
	}()

	ctx := context.Background()
	var insecureFlag bool

	newExporter = func(ctx context.Context, endpoint string, insecure bool) (sdktrace.SpanExporter, error) {
		insecureFlag = insecure
		return &mockExporter{}, nil
	}

	newResource = func(ctx context.Context, serviceName, environment string) (*resource.Resource, error) {
		return resource.New(ctx)
	}

	newTracerProvider = func(exporter sdktrace.SpanExporter, res *resource.Resource, sampleRate float64) *sdktrace.TracerProvider {
		tp := sdktrace.NewTracerProvider(
			sdktrace.WithBatcher(exporter),
			sdktrace.WithResource(res),
			sdktrace.WithSampler(sdktrace.AlwaysSample()),
		)
		return tp
	}

	setGlobalTracerProvider = func(provider *sdktrace.TracerProvider) {}

	cfg := Config{
		ServiceName:       "test-service",
		CollectorEndpoint: "jaeger.prod.example.com:4317",
		Insecure:          false,
		SampleRate:        0.1,
		Environment:       "production",
	}

	client, err := NewClient(ctx, cfg)
	require.NoError(t, err)
	assert.NotNil(t, client)
	assert.False(t, insecureFlag)
}

func TestClientWithContext(t *testing.T) {
	originalExporter := newExporter
	originalResource := newResource
	originalTracerProvider := newTracerProvider
	originalSetTracerProvider := setGlobalTracerProvider
	defer func() {
		newExporter = originalExporter
		newResource = originalResource
		newTracerProvider = originalTracerProvider
		setGlobalTracerProvider = originalSetTracerProvider
	}()

	newExporter = func(ctx context.Context, endpoint string, insecure bool) (sdktrace.SpanExporter, error) {
		return &mockExporter{}, nil
	}

	newResource = func(ctx context.Context, serviceName, environment string) (*resource.Resource, error) {
		return resource.New(ctx)
	}

	newTracerProvider = func(exporter sdktrace.SpanExporter, res *resource.Resource, sampleRate float64) *sdktrace.TracerProvider {
		tp := sdktrace.NewTracerProvider(
			sdktrace.WithBatcher(exporter),
			sdktrace.WithResource(res),
			sdktrace.WithSampler(sdktrace.AlwaysSample()),
		)
		return tp
	}

	setGlobalTracerProvider = func(provider *sdktrace.TracerProvider) {}

	cfg := Config{
		ServiceName:       "test-service",
		CollectorEndpoint: "localhost:4317",
		Insecure:          true,
		SampleRate:        1.0,
		Environment:       "test",
	}

	ctx := context.Background()
	client, err := NewClient(ctx, cfg)
	require.NoError(t, err)
	assert.NotNil(t, client)
}

func TestClientTracer(t *testing.T) {
	originalExporter := newExporter
	originalResource := newResource
	originalTracerProvider := newTracerProvider
	originalSetTracerProvider := setGlobalTracerProvider
	defer func() {
		newExporter = originalExporter
		newResource = originalResource
		newTracerProvider = originalTracerProvider
		setGlobalTracerProvider = originalSetTracerProvider
	}()

	ctx := context.Background()
	newExporter = func(ctx context.Context, endpoint string, insecure bool) (sdktrace.SpanExporter, error) {
		return &mockExporter{}, nil
	}

	newResource = func(ctx context.Context, serviceName, environment string) (*resource.Resource, error) {
		return resource.New(ctx)
	}

	newTracerProvider = func(exporter sdktrace.SpanExporter, res *resource.Resource, sampleRate float64) *sdktrace.TracerProvider {
		tp := sdktrace.NewTracerProvider(
			sdktrace.WithBatcher(exporter),
			sdktrace.WithResource(res),
			sdktrace.WithSampler(sdktrace.AlwaysSample()),
		)
		return tp
	}

	setGlobalTracerProvider = func(provider *sdktrace.TracerProvider) {}

	cfg := Config{
		ServiceName:       "test-service",
		CollectorEndpoint: "localhost:4317",
		Insecure:          true,
		SampleRate:        1.0,
		Environment:       "test",
	}

	client, err := NewClient(ctx, cfg)
	require.NoError(t, err)

	tracer := client.Tracer("test-tracer")
	assert.NotNil(t, tracer)
}
