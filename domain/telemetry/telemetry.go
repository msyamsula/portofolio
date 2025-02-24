package telemetry

import (
	"context"
	"log"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

func InitializeTelemetrySDK(appName, jaegerHost string) {
	// Initialize OpenTelemetry SDK
	ctx := context.Background()
	exporterOptions := []otlptracegrpc.Option{
		otlptracegrpc.WithEndpoint(jaegerHost),
		otlptracegrpc.WithInsecure(), // remove this for production
	}
	exporter, err := otlptracegrpc.New(ctx, exporterOptions...)
	if err != nil {
		log.Fatal("failed to create exporter")
	}

	attr := []attribute.KeyValue{
		{
			Key:   attribute.Key("opentelemetry.io/schemas"),
			Value: attribute.StringValue("1.7.0"),
		},
		{
			Key:   attribute.Key("service.name"),
			Value: attribute.StringValue(appName),
		},
	}

	resource, err := resource.New(ctx,
		resource.WithAttributes(
			attr...,
		),
		resource.WithSchemaURL(semconv.SchemaURL),
	)
	if err != nil {
		log.Fatal("failed to create resource")
	}

	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(resource),
	)

	otel.SetTracerProvider(tracerProvider)
	otel.SetTextMapPropagator(propagation.TraceContext{})
}
