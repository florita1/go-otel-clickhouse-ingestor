package ingestion

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"log"
	"os"
	"net/url"

	"go.opentelemetry.io/otel"
	"github.com/florita1/ingestion-service/internal/model"
)

func InsertEvent(ctx context.Context, event model.Event) error {
    tr := otel.Tracer("ingestion-service")
    ctx, span := tr.Start(ctx, "insertToClickHouse")
    defer span.End()

    clickhouseHost := os.Getenv("CLICKHOUSE_HOST")
	if clickhouseHost == "" {
		clickhouseHost = "localhost"
	}
    query := url.QueryEscape("INSERT INTO events_db.events FORMAT JSONEachRow")
    clickhouseEndpoint := fmt.Sprintf("http://%s:8123/?query=%s", clickhouseHost, query)

    event.Timestamp = event.Timestamp.UTC() // normalize if needed
    formatted := struct {
    	Timestamp string `json:"timestamp"`
    	UserID    string `json:"user_id"`
    	Action    string `json:"action"`
    	Payload   string `json:"payload"`
    }{
    	Timestamp: event.Timestamp.Format("2006-01-02 15:04:05"),
    	UserID:    event.UserID,
    	Action:    event.Action,
    	Payload:   event.Payload,
    }
    body, err := json.Marshal(formatted)
    if err != nil {
    	return fmt.Errorf("marshal error: %w", err)
    }
    log.Println(formatted)
    req, err := http.NewRequest("POST", clickhouseEndpoint, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("request creation error: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

    username := os.Getenv("CLICKHOUSE_USER")
    password := os.Getenv("CLICKHOUSE_PASSWORD")

    if username != "" && password != "" {
    	req.SetBasicAuth(username, password)
    }

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("http post error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("clickhouse returned status: %s", resp.Status)
	}

	return nil
}