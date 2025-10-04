package ingestion

import (
	"os"
)

type Mode string

const (
	ModeSynthetic Mode = "synthetic"
	ModeCDC       Mode = "cdc"
)

type Config struct {
	Mode    Mode
	Brokers []string
	Topic   string
	GroupID string

	ClickHouseHost  string
	ClickHouseUser  string
	ClickHousePass  string
	ClickHouseDB    string // used by CDC inserts
	ClickHouseTable string // used by CDC inserts (e.g., "app.users_cur")
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
