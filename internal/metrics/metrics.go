package metrics

import (
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var IngestedEventCount = prometheus.NewCounter(
	prometheus.CounterOpts{
		Name: "ingestion_generated_events_total",
		Help: "Total number of synthetic events generated",
	},
)

var InsertLatency = prometheus.NewHistogram(
	prometheus.HistogramOpts{
		Name:    "ingestion_clickhouse_insert_latency_seconds",
		Help:    "Latency of inserts to ClickHouse in seconds",
		Buckets: prometheus.DefBuckets,
	},
)

var InsertErrors = prometheus.NewCounter(
	prometheus.CounterOpts{
		Name: "ingestion_clickhouse_insert_errors_total",
		Help: "Total number of insert errors to ClickHouse",
	},
)

var RowsInserted = prometheus.NewCounter(
	prometheus.CounterOpts{
		Name: "ingestion_clickhouse_rows_inserted_total",
		Help: "Total number of rows inserted into ClickHouse",
	},
)

func Init(port string) {
	prometheus.MustRegister(
		IngestedEventCount,
		InsertLatency,
		InsertErrors,
		RowsInserted,
	)

	http.Handle("/metrics", promhttp.Handler())

	log.Printf("Prometheus metrics available at http://ingestion-service:%s/metrics", port)

	go func() {
		if err := http.ListenAndServe(":"+port, nil); err != nil {
			log.Fatalf("Failed to start metrics endpoint: %v", err)
		}
	}()
}
