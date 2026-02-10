package metrics

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/sdk/metric"
	metricdata "go.opentelemetry.io/otel/sdk/metric/metricdata"
	resourceSDK "go.opentelemetry.io/otel/sdk/resource"
)

type mockExporter struct {
	shutdownCalled bool
}

func (m *mockExporter) Temporality(kind metric.InstrumentKind) metricdata.Temporality {
	return metricdata.CumulativeTemporality
}

func (m *mockExporter) Aggregation(kind metric.InstrumentKind) metric.Aggregation {
	return metric.DefaultAggregationSelector(kind)
}

func (m *mockExporter) Export(ctx context.Context, res *metricdata.ResourceMetrics) error {
	return nil
}

func (m *mockExporter) Shutdown(ctx context.Context) error {
	m.shutdownCalled = true
	return nil
}

func (m *mockExporter) ForceFlush(ctx context.Context) error {
	return nil
}

func TestNewClientSuccess(t *testing.T) {
	originalExporter := newMetricExporter
	originalResource := newResource
	originalPeriodicReader := newPeriodicReader
	originalMeterProvider := newMeterProvider
	originalSetGlobalMeterProvider := setGlobalMeterProvider
	defer func() {
		newMetricExporter = originalExporter
		newResource = originalResource
		newPeriodicReader = originalPeriodicReader
		newMeterProvider = originalMeterProvider
		setGlobalMeterProvider = originalSetGlobalMeterProvider
	}()

	ctx := context.Background()

	var capturedExporter metric.Exporter
	var capturedRes *resourceSDK.Resource
	var capturedInterval time.Duration
	var capturedProvider *metric.MeterProvider

	newMetricExporter = func(ctx context.Context, endpoint string, insecure bool) (metric.Exporter, error) {
		assert.Equal(t, "localhost:4317", endpoint)
		assert.True(t, insecure)
		mockExp := &mockExporter{}
		capturedExporter = mockExp
		return mockExp, nil
	}

	newResource = func(ctx context.Context, serviceName, environment string) (*resourceSDK.Resource, error) {
		assert.Equal(t, "test-service", serviceName)
		assert.Equal(t, "test", environment)
		mockRes, _ := resourceSDK.New(ctx)
		capturedRes = mockRes
		return mockRes, nil
	}

	newPeriodicReader = func(exporter metric.Exporter, interval time.Duration) metric.Reader {
		assert.Same(t, capturedExporter, exporter)
		assert.Equal(t, 15*time.Second, interval)
		capturedInterval = interval
		return metric.NewPeriodicReader(exporter, metric.WithInterval(interval))
	}

	newMeterProvider = func(reader metric.Reader, res *resourceSDK.Resource) *metric.MeterProvider {
		assert.Same(t, capturedRes, res)
		mp := metric.NewMeterProvider(
			metric.WithResource(res),
			metric.WithReader(reader),
		)
		capturedProvider = mp
		return mp
	}

	setGlobalMeterProviderCalled := false
	setGlobalMeterProvider = func(mp *metric.MeterProvider) {
		setGlobalMeterProviderCalled = true
		assert.Same(t, capturedProvider, mp)
	}

	cfg := Config{
		ServiceName:       "test-service",
		CollectorEndpoint: "localhost:4317",
		Insecure:          true,
		PushInterval:      15 * time.Second,
		Environment:       "test",
	}

	client, err := NewClient(ctx, cfg)
	require.NoError(t, err)
	assert.NotNil(t, client)
	assert.NotNil(t, client.MeterProvider)
	assert.Same(t, capturedProvider, client.MeterProvider)
	assert.True(t, setGlobalMeterProviderCalled)
	assert.Equal(t, 15*time.Second, capturedInterval)
}

func TestNewClientExporterError(t *testing.T) {
	originalExporter := newMetricExporter
	defer func() { newMetricExporter = originalExporter }()

	ctx := context.Background()
	newMetricExporter = func(ctx context.Context, endpoint string, insecure bool) (metric.Exporter, error) {
		return nil, errors.New("exporter creation failed")
	}

	cfg := Config{
		ServiceName:       "test-service",
		CollectorEndpoint: "localhost:4317",
		Insecure:          true,
		PushInterval:      15 * time.Second,
		Environment:       "test",
	}

	client, err := NewClient(ctx, cfg)
	require.Error(t, err)
	assert.Nil(t, client)
	assert.Contains(t, err.Error(), "failed to create metric exporter")
}

func TestNewClientResourceError(t *testing.T) {
	originalExporter := newMetricExporter
	originalResource := newResource
	defer func() {
		newMetricExporter = originalExporter
		newResource = originalResource
	}()

	ctx := context.Background()
	newMetricExporter = func(ctx context.Context, endpoint string, insecure bool) (metric.Exporter, error) {
		return &mockExporter{}, nil
	}

	newResource = func(ctx context.Context, serviceName, environment string) (*resourceSDK.Resource, error) {
		return nil, errors.New("resource creation failed")
	}

	cfg := Config{
		ServiceName:       "test-service",
		CollectorEndpoint: "localhost:4317",
		Insecure:          true,
		PushInterval:      15 * time.Second,
		Environment:       "test",
	}

	client, err := NewClient(ctx, cfg)
	require.Error(t, err)
	assert.Nil(t, client)
	assert.Contains(t, err.Error(), "failed to create resource")
}

func TestClientShutdown(t *testing.T) {
	originalExporter := newMetricExporter
	originalResource := newResource
	originalPeriodicReader := newPeriodicReader
	originalMeterProvider := newMeterProvider
	originalSetGlobalMeterProvider := setGlobalMeterProvider
	defer func() {
		newMetricExporter = originalExporter
		newResource = originalResource
		newPeriodicReader = originalPeriodicReader
		newMeterProvider = originalMeterProvider
		setGlobalMeterProvider = originalSetGlobalMeterProvider
	}()

	ctx := context.Background()
	newMetricExporter = func(ctx context.Context, endpoint string, insecure bool) (metric.Exporter, error) {
		return &mockExporter{}, nil
	}

	newResource = func(ctx context.Context, serviceName, environment string) (*resourceSDK.Resource, error) {
		return resourceSDK.New(ctx)
	}

	newPeriodicReader = func(exporter metric.Exporter, interval time.Duration) metric.Reader {
		return metric.NewPeriodicReader(exporter, metric.WithInterval(interval))
	}

	newMeterProvider = func(reader metric.Reader, res *resourceSDK.Resource) *metric.MeterProvider {
		return metric.NewMeterProvider(
			metric.WithResource(res),
			metric.WithReader(reader),
		)
	}

	setGlobalMeterProvider = func(mp *metric.MeterProvider) {}

	cfg := Config{
		ServiceName:       "test-service",
		CollectorEndpoint: "localhost:4317",
		Insecure:          true,
		PushInterval:      15 * time.Second,
		Environment:       "test",
	}

	client, err := NewClient(ctx, cfg)
	require.NoError(t, err)

	err = client.Shutdown(ctx)
	require.NoError(t, err)
}

func TestClientSecureConnection(t *testing.T) {
	originalExporter := newMetricExporter
	originalResource := newResource
	originalPeriodicReader := newPeriodicReader
	originalMeterProvider := newMeterProvider
	originalSetGlobalMeterProvider := setGlobalMeterProvider
	defer func() {
		newMetricExporter = originalExporter
		newResource = originalResource
		newPeriodicReader = originalPeriodicReader
		newMeterProvider = originalMeterProvider
		setGlobalMeterProvider = originalSetGlobalMeterProvider
	}()

	ctx := context.Background()
	var insecureFlag bool

	newMetricExporter = func(ctx context.Context, endpoint string, insecure bool) (metric.Exporter, error) {
		insecureFlag = insecure
		return &mockExporter{}, nil
	}

	newResource = func(ctx context.Context, serviceName, environment string) (*resourceSDK.Resource, error) {
		return resourceSDK.New(ctx)
	}

	newPeriodicReader = func(exporter metric.Exporter, interval time.Duration) metric.Reader {
		return metric.NewPeriodicReader(exporter, metric.WithInterval(interval))
	}

	newMeterProvider = func(reader metric.Reader, res *resourceSDK.Resource) *metric.MeterProvider {
		return metric.NewMeterProvider(
			metric.WithResource(res),
			metric.WithReader(reader),
		)
	}

	setGlobalMeterProvider = func(mp *metric.MeterProvider) {}

	cfg := Config{
		ServiceName:       "test-service",
		CollectorEndpoint: "otel.prod.example.com:4317",
		Insecure:          false,
		PushInterval:      30 * time.Second,
		Environment:       "production",
	}

	client, err := NewClient(ctx, cfg)
	require.NoError(t, err)
	assert.NotNil(t, client)
	assert.False(t, insecureFlag)
}

func TestClientWithContext(t *testing.T) {
	originalExporter := newMetricExporter
	originalResource := newResource
	originalPeriodicReader := newPeriodicReader
	originalMeterProvider := newMeterProvider
	originalSetGlobalMeterProvider := setGlobalMeterProvider
	defer func() {
		newMetricExporter = originalExporter
		newResource = originalResource
		newPeriodicReader = originalPeriodicReader
		newMeterProvider = originalMeterProvider
		setGlobalMeterProvider = originalSetGlobalMeterProvider
	}()

	newMetricExporter = func(ctx context.Context, endpoint string, insecure bool) (metric.Exporter, error) {
		return &mockExporter{}, nil
	}

	newResource = func(ctx context.Context, serviceName, environment string) (*resourceSDK.Resource, error) {
		return resourceSDK.New(ctx)
	}

	newPeriodicReader = func(exporter metric.Exporter, interval time.Duration) metric.Reader {
		return metric.NewPeriodicReader(exporter, metric.WithInterval(interval))
	}

	newMeterProvider = func(reader metric.Reader, res *resourceSDK.Resource) *metric.MeterProvider {
		return metric.NewMeterProvider(
			metric.WithResource(res),
			metric.WithReader(reader),
		)
	}

	setGlobalMeterProvider = func(mp *metric.MeterProvider) {}

	cfg := Config{
		ServiceName:       "test-service",
		CollectorEndpoint: "localhost:4317",
		Insecure:          true,
		PushInterval:      15 * time.Second,
		Environment:       "test",
	}

	ctx := context.Background()
	client, err := NewClient(ctx, cfg)
	require.NoError(t, err)
	assert.NotNil(t, client)
}

func TestClientMeter(t *testing.T) {
	originalExporter := newMetricExporter
	originalResource := newResource
	originalPeriodicReader := newPeriodicReader
	originalMeterProvider := newMeterProvider
	originalSetGlobalMeterProvider := setGlobalMeterProvider
	defer func() {
		newMetricExporter = originalExporter
		newResource = originalResource
		newPeriodicReader = originalPeriodicReader
		newMeterProvider = originalMeterProvider
		setGlobalMeterProvider = originalSetGlobalMeterProvider
	}()

	ctx := context.Background()
	newMetricExporter = func(ctx context.Context, endpoint string, insecure bool) (metric.Exporter, error) {
		return &mockExporter{}, nil
	}

	newResource = func(ctx context.Context, serviceName, environment string) (*resourceSDK.Resource, error) {
		return resourceSDK.New(ctx)
	}

	newPeriodicReader = func(exporter metric.Exporter, interval time.Duration) metric.Reader {
		return metric.NewPeriodicReader(exporter, metric.WithInterval(interval))
	}

	newMeterProvider = func(reader metric.Reader, res *resourceSDK.Resource) *metric.MeterProvider {
		return metric.NewMeterProvider(
			metric.WithResource(res),
			metric.WithReader(reader),
		)
	}

	setGlobalMeterProvider = func(mp *metric.MeterProvider) {}

	cfg := Config{
		ServiceName:       "test-service",
		CollectorEndpoint: "localhost:4317",
		Insecure:          true,
		PushInterval:      15 * time.Second,
		Environment:       "test",
	}

	client, err := NewClient(ctx, cfg)
	require.NoError(t, err)

	meter := client.Meter("test-meter")
	assert.NotNil(t, meter)
}
