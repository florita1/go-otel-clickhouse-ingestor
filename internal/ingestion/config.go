package ingestion

import (
	"os"
	"strings"
)

type Mode string
const (
	ModeSynthetic Mode = "synthetic"
	ModeCDC       Mode = "cdc"
)

type Config struct {
	Mode          Mode
	Brokers       []string
	Topic         string
	GroupID       string
	ClickHouseDB   string
	ClickHouseTable string
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" { return v }
	return def
}

func LoadConfig() Config {
	return Config{
		Mode:            Mode(getenv("MODE", "synthetic")),
		Brokers:         strings.Split(getenv("REDPANDA_BROKERS", "redpanda:9093"), ","),
		Topic:           getenv("TOPIC", "dbserver1.app.users"),
		GroupID:         getenv("GROUP_ID", "wal-cdc-ingestor"),
		ClickHouseDB:    getenv("CLICKHOUSE_DB", "appdb"),
		ClickHouseTable: getenv("CLICKHOUSE_TABLE", "app.users_cur"),
	}
}
