package cmd

import (
    "fmt"
    "os"
    "strconv"
    "context"
	"log"
	"time"
	"math/rand"
	"strings"
	"github.com/florita1/ingestion-service/internal/metrics"
	"github.com/florita1/ingestion-service/internal/tracing"
	"go.opentelemetry.io/otel/trace"
	"github.com/florita1/ingestion-service/internal/ingestion"
	"github.com/florita1/ingestion-service/internal/generator"
	"github.com/florita1/ingestion-service/internal/logging"
	"github.com/spf13/cobra"
)

var (
	clickhouseHost string
	eventRate     int
	durationSec   int
    mode       string // synthetic | cdc
    brokersCSV string
    topic      string
    groupID    string
)

var rootCmd = &cobra.Command{
	Use:   "ingestion-service",
	Short: "Generate and stream synthetic events to ClickHouse",
	Run: func(cmd *cobra.Command, args []string) {
		switch strings.ToLower(mode) {
		case "", "synthetic":
			runIngestion() // existing synthetic path
		case "cdc":
			runCDC()
		default:
			fmt.Printf("unknown --mode=%s (expected synthetic|cdc)\n", mode)
			os.Exit(1)
		}
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	// Bind flags with fallback to environment variables
	rootCmd.Flags().StringVar(&clickhouseHost, "clickhouse-host", getEnv("CLICKHOUSE_HOST", "localhost"), "ClickHouse Host")
	rootCmd.Flags().IntVar(&eventRate, "rate", getEnvAsInt("EVENT_RATE", 5), "Events per second")
	rootCmd.Flags().IntVar(&durationSec, "duration", getEnvAsInt("INGESTION_DURATION", 60), "How long to run ingestion (in seconds)")

    rootCmd.Flags().StringVar(&mode, "mode", getEnv("MODE", "synthetic"), "synthetic|cdc")
    rootCmd.Flags().StringVar(&brokersCSV, "brokers", getEnv("REDPANDA_BROKERS", "redpanda.redpanda.svc.cluster.local:9093"), "comma-separated brokers")
    rootCmd.Flags().StringVar(&topic, "topic", getEnv("TOPIC", "dbserver1.app.users"), "Kafka topic")
    rootCmd.Flags().StringVar(&groupID, "group", getEnv("GROUP_ID", "wal-cdc-ingestor"), "Kafka consumer group")

}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func getEnvAsInt(key string, fallback int) int {
	valStr := os.Getenv(key)
	if valStr == "" {
		return fallback
	}
	val, err := strconv.Atoi(valStr)
	if err != nil {
		return fallback
	}
	return val
}

func runIngestion() {
	fmt.Printf("Starting ingestion\nClickHouse Host: %s\nRate: %d/sec\nDuration: %ds\n\n", clickhouseHost, eventRate, durationSec)

    ctx := context.Background()
    tracing.Init("ingestion-service")
    defer tracing.Shutdown(ctx)

    metrics.Init("8080")

    rand.Seed(time.Now().UnixNano())

    ticker := time.NewTicker(time.Second / time.Duration(eventRate))
    defer ticker.Stop()

    log.Println("starting event generator")

    timeout := time.After(time.Duration(durationSec) * time.Second)

    for {
    	select {
    	case <- ticker.C:
            logging.WithTrace(ctx, "Generate event")
            var span trace.Span
            ctx, span = tracing.Tracer.Start(ctx, "generateEvent")
            event := generator.GenerateEvent()

            err := ingestion.InsertEvent(ctx, event)
            if err != nil {
                logging.WithTrace(ctx, "Failed to insert event: %v", err)
            }

            span.End()
            logging.WithTrace(ctx, "Event generated with user_id = %s action = %s", event.UserID, event.Action)

            metrics.IngestedEventCount.Inc()
        case <- timeout:
            logging.WithTrace(ctx, "Ingestion complete")
        	select {} // keep metrics server alive for prometheus scraping
        }
    }
}

func runCDC() {
  fmt.Printf("Starting CDC consumer\nBrokers: %s\nTopic: %s\nGroup: %s\n\n", brokersCSV, topic, groupID)

  ctx := context.Background()
  tracing.Init("ingestion-service")
  defer tracing.Shutdown(ctx)
  metrics.Init("8080")

  cfg := ingestion.CDCConfig{
    Brokers: splitCSV(brokersCSV),
    Topic:   topic,
    GroupID: groupID,
  }
  if err := ingestion.RunCDC(ctx, cfg); err != nil {
    log.Fatalf("cdc error: %v", err)
  }
}

func splitCSV(s string) []string {
  var out []string
  for _, p := range strings.Split(s, ",") {
    if q := strings.TrimSpace(p); q != "" { out = append(out, q) }
  }
  return out
}

