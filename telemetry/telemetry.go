package telemetry

import (
	"context"
	"log"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

func InitializeTelemetryTracing(appName, jaegerHost string) func() {
	// Initialize OpenTelemetry SDK
	ctx := context.Background()
	exporterOptions := []otlptracegrpc.Option{
		otlptracegrpc.WithEndpoint(jaegerHost),
	}
	exporter, err := otlptracegrpc.New(ctx, exporterOptions...)
	if err != nil {
		log.Fatal("failed to create exporter")
	}

	// tracer provider
	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(resource.Default()),
	)

	otel.SetTracerProvider(tracerProvider)

	return func() {
		tracerProvider.Shutdown(ctx)
	}

}
