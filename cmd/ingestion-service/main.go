package main

import (
    "context"
	"log"
	"time"
	"math/rand"
	"github.com/google/uuid"
	"encoding/json"
	"github.com/florita1/ingestion-service/internal/metrics"
	"github.com/florita1/ingestion-service/internal/tracing"
	"go.opentelemetry.io/otel/trace"
	"github.com/florita1/ingestion-service/internal/model"
	"github.com/florita1/ingestion-service/internal/ingestion"
)


var actions = [] string {"login", "click", "purchase", "logout"}

func main() {
    ctx := context.Background()
    tracing.Init("ingestion-service")
    defer tracing.Shutdown(ctx)

    metrics.Init("8080")

    rand.Seed(time.Now().UnixNano())

    ticker := time.NewTicker(1 * time.Second)
    defer ticker.Stop()

    log.Println("starting event generator")

    for range ticker.C {
        log.Println("generate event")
        var span trace.Span
        ctx, span = tracing.Tracer.Start(ctx, "generateEvent")
        event := generateEvent()

        err := ingestion.InsertEvent(ctx, event)
        if err != nil {
        	log.Printf("Failed to insert event: %v", err)
        }

        span.End()
        jsonEvent, err := json.MarshalIndent(event, "", "  ")
        if err != nil {
        	log.Printf("Failed to serialize event: %v", err)
        	continue
        }
        log.Println(string(jsonEvent))
        metrics.IngestedEventCount.Inc()
    }
}

func generateEvent() model.Event {
    return model.Event {
        Timestamp:  time.Now(),
        UserID:     getUserId(),
        Action:     actions[rand.Intn(len(actions))],
        Payload:    "example-payload",
    }
}

func getUserId() string {
    return "user-" + uuid.NewString()
}