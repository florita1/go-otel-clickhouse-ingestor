# ingestion-service

## Usage

With flags:
./ingestion-service --clickhouse-url=http://localhost:8123 --rate=5 --duration=60

With environment variables:
export CLICKHOUSE_URL=http://localhost:8123
export EVENT_RATE=5
export INGESTION_DURATION=60
./ingestion-service