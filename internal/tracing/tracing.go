package tracing

import (
	"context"
	"log"
	"os"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
)

var Tracer trace.Tracer
var traceProvider *sdktrace.TracerProvider

func Init(serviceName string) {
	ctx := context.Background()

	// Enable OTEL Go SDK debug logging
	os.Setenv("OTEL_GO_LOG_LEVEL", "debug")

	// Set OTLP endpoint from env or use default
	endpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	if endpoint == "" {
		endpoint = "localhost:4318"
	}

	// OTLP HTTP exporter (to Alloy)
	otlpExporter, err := otlptracehttp.New(ctx,
		otlptracehttp.WithEndpoint(endpoint),
		otlptracehttp.WithInsecure(),
	)
	if err != nil {
		log.Fatalf("Failed to create OTLP exporter: %v", err)
	}

	// Console exporter (for human-readable stdout)
	consoleExporter, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
	if err != nil {
		log.Fatalf("Failed to create console exporter: %v", err)
	}

	// Resource describing this service
	res := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceName(serviceName),
	)

	// Configure tracer provider with both exporters
	traceProvider = sdktrace.NewTracerProvider(
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(sdktrace.NewBatchSpanProcessor(otlpExporter)),   // → Alloy
		sdktrace.WithSpanProcessor(sdktrace.NewSimpleSpanProcessor(consoleExporter)), // → stdout
	)

	otel.SetTracerProvider(traceProvider)
	Tracer = traceProvider.Tracer(serviceName)

	log.Println("OpenTelemetry tracing initialized with OTLP + console exporters")
}

func Shutdown(ctx context.Context) {
	if traceProvider != nil {
		if err := traceProvider.Shutdown(ctx); err != nil {
			log.Printf("Error shutting down tracer: %v", err)
		} else {
			log.Println("Tracer shutdown complete")
		}
	}
}
