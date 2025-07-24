package tracing

import (
	"context"
	"log"
	"os"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/semconv/v1.17.0"
)

var Tracer trace.Tracer
var traceProvider *sdktrace.TracerProvider

func Init(serviceName string) {
    ctx := context.Background()

    // Use env variable if set, else default to localhost:4318
    endpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
    if endpoint == "" {
    	endpoint = "localhost:4318"
    }

    exporter, err := otlptracehttp.New(ctx,
    	otlptracehttp.WithEndpoint(endpoint),
    	otlptracehttp.WithInsecure(),
    )

    if err != nil {
    	log.Fatalf("Failed to create OTLP exporter: %v", err)
    }

	res := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceName(serviceName),
	)

	traceProvider = sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)

	otel.SetTracerProvider(traceProvider)
	Tracer = traceProvider.Tracer(serviceName)

	log.Println("OpenTelemetry tracing initialized")
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