package ingestion

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"net/url"
	"time"
	"strings"

	"go.opentelemetry.io/otel"
	"github.com/florita1/ingestion-service/internal/metrics"
	"github.com/florita1/ingestion-service/internal/model"
	"github.com/florita1/ingestion-service/internal/logging"
)

func chHostPort() string {
  host := os.Getenv("CLICKHOUSE_HOST")
  if host == "" { host = "localhost" }
  if strings.Contains(host, ":") { return host }
  return host + ":8123"
}

func doJSONEachRowPOST(ctx context.Context, endpoint string, payload any) error {
  tr := otel.Tracer("ingestion-service")
  ctx, span := tr.Start(ctx, "clickhouse.post")
  defer span.End()

  start := time.Now()

  body, err := json.Marshal(payload)
  if err != nil {
    metrics.InsertErrors.Inc()
    logging.WithTrace(ctx, "marshal error: %v", err)
    return fmt.Errorf("marshal error: %w", err)
  }

  req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(body))
  if err != nil {
    metrics.InsertErrors.Inc()
    logging.WithTrace(ctx, "request creation error: %v", err)
    return fmt.Errorf("request creation error: %w", err)
  }
  req.Header.Set("Content-Type", "application/json")

  if u, p := os.Getenv("CLICKHOUSE_USER"), os.Getenv("CLICKHOUSE_PASSWORD"); u != "" && p != "" {
    req.SetBasicAuth(u, p)
  }

  resp, err := (&http.Client{}).Do(req)
  metrics.InsertLatency.Observe(time.Since(start).Seconds())

  if err != nil {
    metrics.InsertErrors.Inc()
    logging.WithTrace(ctx, "http post error: %v", err)
    return fmt.Errorf("http post error: %w", err)
  }
  defer resp.Body.Close()

  if resp.StatusCode != http.StatusOK {
    metrics.InsertErrors.Inc()
    logging.WithTrace(ctx, "clickhouse status error: %s", resp.Status)
    return fmt.Errorf("clickhouse returned status: %s", resp.Status)
  }

  metrics.RowsInserted.Inc()
  return nil
}

func InsertEvent(ctx context.Context, event model.Event) error {
  clickhouse := chHostPort()
  query := url.QueryEscape("INSERT INTO events_db.events FORMAT JSONEachRow")
  endpoint := fmt.Sprintf("http://%s/?query=%s", clickhouse, query)

  event.Timestamp = event.Timestamp.UTC()
  payload := struct {
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

  if err := doJSONEachRowPOST(ctx, endpoint, payload); err != nil {
    logging.WithTrace(ctx, "InsertEvent failed: %v", err)
    return err
  }
  logging.WithTrace(ctx, "Inserted event")
  return nil
}

func InsertCDCUser(ctx context.Context, row model.CDCUserRow) error {
  clickhouse := chHostPort()
  db := os.Getenv("CLICKHOUSE_DB")
  if db == "" { db = "appdb" }
  table := os.Getenv("CLICKHOUSE_TABLE")
  if table == "" { table = "app.users_cur" }

  q := url.QueryEscape(fmt.Sprintf("INSERT INTO %s FORMAT JSONEachRow", table))
  endpoint := fmt.Sprintf("http://%s/?database=%s&query=%s", clickhouse, db, q)

  ts := row.Ts.UTC().Format("2006-01-02 15:04:05")
  payload := struct {
    ID        uint64 `json:"id"`
    Name      string `json:"name"`
    Email     string `json:"email"`
    IsDeleted uint8  `json:"is_deleted"`
    Op        uint8  `json:"_op"`
    LSN       uint64 `json:"_lsn"`
    TS        string `json:"_ts"`
  }{
    ID: row.ID, Name: row.Name, Email: row.Email,
    IsDeleted: row.IsDeleted, Op: row.Op, LSN: row.LSN, TS: ts,
  }

  if err := doJSONEachRowPOST(ctx, endpoint, payload); err != nil {
    logging.WithTrace(ctx, "InsertCDCUser failed: %v", err)
    return err
  }
  logging.WithTrace(ctx, "Inserted CDC user row")
  return nil
}

