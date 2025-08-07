package logging

import (
	"context"
	"log"
	"go.opentelemetry.io/otel/trace"
)

func WithTrace(ctx context.Context, format string, args ...any) {
	traceID := trace.SpanContextFromContext(ctx).TraceID().String()
	log.Printf("trace_id=%s "+format, append([]any{traceID}, args...)...)
}
