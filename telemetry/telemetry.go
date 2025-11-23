package telemetry

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/zipkin"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

func InitializeTelemetryTracing(appName, tracerCollectorEndpoint string) func() {
	// Initialize OpenTelemetry SDK
	ctx := context.Background()
	exporter, err := zipkin.New(tracerCollectorEndpoint)
	if err != nil {
		return func() {}
	}

	res, err := resource.New(
		context.Background(),
		resource.WithAttributes(
			semconv.ServiceName(appName),
			semconv.ServiceVersion("v1.0.0"),
			semconv.DeploymentEnvironment("local"),
		),
	)
	if err != nil {
		panic(err)
	}

	// tracer provider
	tracerProvider := sdktrace.NewTracerProvider(
		// sdktrace.WithBatcher(exporter), // batch send to collector
		sdktrace.WithSyncer(exporter), // always send to colletor, good for debugging not for production
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.AlwaysSample()), // again, not for production, use sampling deliberately
		// sdktrace.WithSampler(sdktrace.TraceIDRatioBased(0.1)), // this is good choice for start, sampling 10% traffic in local sampler
	)

	otel.SetTracerProvider(tracerProvider)

	return func() {
		tracerProvider.Shutdown(ctx)
	}

}
