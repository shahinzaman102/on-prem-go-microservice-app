package tracing

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	sdkresource "go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
)

// Initializes OpenTelemetry tracing with OTLP (for Jaeger).
func InitTracer(ctx context.Context) (func(context.Context) error, error) {
	// Set up the OTLP trace exporter to send data to the collector
	exporter, err := otlptracehttp.New(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create OTLP trace exporter: %w", err)
	}

	// Set up trace provider with resource info
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(sdkresource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName("logger-service"),
		)),
	)

	// Register global trace provider
	otel.SetTracerProvider(tp)

	return tp.Shutdown, nil
}
