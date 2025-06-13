package main

import (
	"log"
	"time"
	"math/rand"
	"github.com/google/uuid"
	"encoding/json"
	"github.com/florita1/ingestion-service/internal/metrics"
)

// Event represents a synthetic user action.
type Event struct {
	Timestamp time.Time `json:"timestamp"`
	UserID    string    `json:"user_id"`
	Action    string    `json:"action"`
	Payload   string    `json:"payload"`
}

var actions = [] string {"login", "click", "purchase", "logout"}

func main() {
    metrics.Init("8080")

    rand.Seed(time.Now().UnixNano())

    ticker := time.NewTicker(1 * time.Second)
    defer ticker.Stop()

    log.Println("starting event generator")

    for range ticker.C {
        log.Println("generate event")
        event := generateEvent()
        jsonEvent, err := json.MarshalIndent(event, "", "  ")
        if err != nil {
        	log.Printf("Failed to serialize event: %v", err)
        	continue
        }
        log.Println(string(jsonEvent))
        metrics.IngestedEventCount.Inc()
    }
}

func generateEvent() Event {
    return Event {
        Timestamp:  time.Now(),
        UserID:     getUserId(),
        Action:     actions[rand.Intn(len(actions))],
        Payload:    "example-payload",
    }
}

func getUserId() string {
    return "user-" + uuid.NewString()
}