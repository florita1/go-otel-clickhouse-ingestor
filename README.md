# 🧪 Go Ingestion Service

A dual‑mode Go service that streams events into ClickHouse.

- **Synthetic mode** (default): generates mock JSON events for testing/observability
- **CDC mode**: consumes Debezium‑wrapped PostgreSQL WAL events from Redpanda and inserts into ClickHouse

Fully instrumented with **Prometheus** (metrics) and **OpenTelemetry** (traces).

---

## 🚀 Features

- Modular **Cobra** CLI with env‑var fallbacks
- **Dual mode**: `synthetic` or `cdc`
- Streams events to ClickHouse (via HTTP interface)
- Emits Prometheus metrics at `/metrics` (port configurable)
- Traces event generation and ingestion using OpenTelemetry
- Single source of truth config (`ingestion.Config`) — **envs by default, flags override**

> **Configuration precedence:** Environment variables are production‑standard (12‑Factor/Kubernetes). Any provided **CLI flags override** envs for local/dev convenience.

---

## ⚙️ Quickstart

### Synthetic Mode (default)

Generates random JSON events and inserts into `events_db.events`.

```bash
go run main.go   --mode=synthetic   --clickhouse-host=localhost   --rate=5   --duration=30
```

### CDC Mode

Consumes Debezium envelopes from Redpanda and inserts into ClickHouse (defaults to DB `appdb`, table `app.users_cur` unless overridden).

```bash
go run main.go   --mode=cdc   --clickhouse-host=localhost   --brokers=redpanda.redpanda.svc.cluster.local:9093   --topic=dbserver1.app.users   --group=wal-cdc-ingestor
```

> **ClickHouse host/port:** If `--clickhouse-host` (or `CLICKHOUSE_HOST`) lacks a port, `:8123` is assumed.  
> **Auth:** Basic auth is used only if **both** user and password are provided.

---

## 🔑 Flags & Environment Variables

| Flag                     | Env Var                | Description                                                             | Default                                         |
|--------------------------|------------------------|-------------------------------------------------------------------------|-------------------------------------------------|
| `--mode`                 | `MODE`                 | `synthetic` or `cdc`                                                    | `synthetic`                                     |
| `--clickhouse-host`      | `CLICKHOUSE_HOST`      | ClickHouse host or `host:port` (defaults to `:8123` if none)            | `localhost`                                     |
| `--clickhouse-user`      | `CLICKHOUSE_USER`      | ClickHouse username (optional)                                          | *(empty)*                                       |
| `--clickhouse-password`  | `CLICKHOUSE_PASSWORD`  | ClickHouse password (optional)                                          | *(empty)*                                       |
| `--clickhouse-db`        | `CLICKHOUSE_DB`        | Database for CDC inserts                                                | `appdb`                                         |
| `--clickhouse-table`     | `CLICKHOUSE_TABLE`     | Table for CDC inserts (e.g., `app.users_cur`)                           | `app.users_cur`                                 |
| `--rate`                 | `EVENT_RATE`           | Events/second (synthetic only)                                          | `5`                                             |
| `--duration`             | `INGESTION_DURATION`   | Run duration in seconds (synthetic only)                                | `60`                                            |
| `--brokers`              | `REDPANDA_BROKERS`     | Comma‑separated Redpanda/Kafka brokers                                  | `redpanda.redpanda.svc.cluster.local:9093`      |
| `--topic`                | `TOPIC`                | Topic with Debezium envelopes                                           | `dbserver1.app.users`                           |
| `--group`                | `GROUP_ID`             | Kafka consumer group                                                    | `wal-cdc-ingestor`                              |
| `--metrics-port`         | `METRICS_PORT`         | Metrics HTTP port                                                       | `8080`                                          |

---

## 📊 Metrics (Prometheus)

Metrics are served at:

```
http://localhost:8080/metrics
```

(or `http://<host>:<metrics-port>/metrics` if overridden)

Common metrics exposed include:

- `ingestion_generated_events_total` — synthetic events generated
- `ingestion_insert_latency_seconds` — insert latency histogram
- `ingestion_insert_errors_total` — insert error counter
- `ingestion_rows_inserted_total` — successful row inserts

Example scrape config:

```yaml
scrape_configs:
  - job_name: 'ingestion-service'
    static_configs:
      - targets: ['host.docker.internal:8080']
```

---

## 📈 Traces (OpenTelemetry)

Traces are exported via OTLP/HTTP to `localhost:4318` (commonly **Grafana Alloy** → **Tempo**).

Typical spans include:

- `generateEvent` (synthetic)
- `clickhouse.post` (HTTP insert span)

Search in Tempo:

```
service.name = "ingestion-service"
```

---

## 🔗 Integration (PostgreSQL WAL → ClickHouse)

This service fits into a WAL‑based CDC pipeline:

```
PostgreSQL (WAL) → Debezium → Redpanda → Ingestion Service → ClickHouse
```

- **Synthetic mode**: test ClickHouse ingestion & observability end‑to‑end without a source DB.
- **CDC mode**: consume Debezium‑wrapped WAL events from Redpanda and upsert into ClickHouse (ReplacingMergeTree recommended).

---

## 📁 Directory Structure

```
.
├── cmd/               # Cobra entrypoint + flags
│   └── root.go
├── internal/
│   ├── generator/     # synthetic event generator
│   ├── ingestion/     # ClickHouse + CDC logic (HTTP inserts, consumers, config)
│   ├── logging/       # logging helpers
│   ├── metrics/       # Prometheus setup & collectors
│   ├── model/         # event + CDC models
│   └── tracing/       # OTEL setup
├── main.go
├── Dockerfile
├── go.mod
├── go.sum
└── README.md
```

---

## 🛠️ Development

Install deps:

```bash
go mod tidy
```

Run locally (synthetic):

```bash
go run main.go --mode=synthetic --clickhouse-host=localhost
```

Run locally (cdc):

```bash
go run main.go --mode=cdc   --clickhouse-host=localhost   --brokers=localhost:9092   --topic=dbserver1.app.users   --group=wal-cdc-ingestor
```

---

## 🐳 Docker

Build multi‑arch and push to GHCR:

```bash
docker buildx build   --platform linux/amd64,linux/arm64   -t ghcr.io/florita1/ingestion-service:latest   -t ghcr.io/florita1/ingestion-service:v0.2.4   --push .
```

Published images:
- `ghcr.io/florita1/ingestion-service:latest`
- `ghcr.io/florita1/ingestion-service:v0.2.4`

---

## 👤 Author

[Florita Nichols](https://www.linkedin.com/in/floritanichols)
