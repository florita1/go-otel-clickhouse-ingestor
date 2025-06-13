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

func Init(port string) {
	prometheus.MustRegister(IngestedEventCount)

	http.Handle("/metrics", promhttp.Handler())

	log.Printf("Prometheus metrics available at http://localhost:%s/metrics", port)

	go func() {
		if err := http.ListenAndServe(":" + port, nil); err != nil {
			log.Fatalf("Failed to start metrics endpoint: %v", err)
		}
	} ()
}