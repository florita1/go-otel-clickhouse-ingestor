package cmd

import (
    "fmt"
    "os"
    "strconv"
    "context"
	"log"
	"time"
	"math/rand"
	"encoding/json"
	"github.com/florita1/ingestion-service/internal/metrics"
	"github.com/florita1/ingestion-service/internal/tracing"
	"go.opentelemetry.io/otel/trace"
	"github.com/florita1/ingestion-service/internal/ingestion"
	"github.com/florita1/ingestion-service/internal/generator"
	"github.com/spf13/cobra"
)

var (
	clickhouseURL string
	eventRate     int
	durationSec   int
)

var rootCmd = &cobra.Command{
	Use:   "ingestion-service",
	Short: "Generate and stream synthetic events to ClickHouse",
	Run: func(cmd *cobra.Command, args []string) {
		runIngestion()
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
	rootCmd.Flags().StringVar(&clickhouseURL, "clickhouse-url", getEnv("CLICKHOUSE_URL", "http://localhost:8123"), "ClickHouse HTTP URL")
	rootCmd.Flags().IntVar(&eventRate, "rate", getEnvAsInt("EVENT_RATE", 5), "Events per second")
	rootCmd.Flags().IntVar(&durationSec, "duration", getEnvAsInt("INGESTION_DURATION", 60), "How long to run ingestion (in seconds)")
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
	fmt.Printf("Starting ingestion\nClickHouse URL: %s\nRate: %d/sec\nDuration: %ds\n\n", clickhouseURL, eventRate, durationSec)

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
            log.Println("generate event")
            var span trace.Span
            ctx, span = tracing.Tracer.Start(ctx, "generateEvent")
            event := generator.GenerateEvent()

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
        case <- timeout:
            log.Println("Ingestion complete.")
        	select {} // keep metrics server alive for prometheus scraping
        }
    }
}
