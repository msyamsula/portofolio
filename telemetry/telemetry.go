package telemetry

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/zipkin"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

func InitializeTelemetryTracing(appName, tracerCollectorEndpoint string) func() {
	// Initialize OpenTelemetry SDK
	ctx := context.Background()
	exporter, err := zipkin.New(tracerCollectorEndpoint)
	if err != nil {
		return func() {}
	}

	// tracer provider
	tracerProvider := sdktrace.NewTracerProvider(
		// sdktrace.WithBatcher(exporter), // batch send to collector
		sdktrace.WithSyncer(exporter), // always send to colletor, good for debugging not for production
		sdktrace.WithResource(resource.Default()),
		sdktrace.WithSampler(sdktrace.AlwaysSample()), // again, not for production, use sampling deliberately
		// sdktrace.WithSampler(sdktrace.TraceIDRatioBased(0.1)), // this is good choice for start, sampling 10% traffic in local sampler
	)

	otel.SetTracerProvider(tracerProvider)

	return func() {
		tracerProvider.Shutdown(ctx)
	}

}
