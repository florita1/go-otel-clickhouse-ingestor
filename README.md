# 🧪 Go Ingestion Service

A synthetic event generator written in Go that streams JSON-formatted events into ClickHouse. Fully instrumented with
OpenTelemetry and Prometheus for observability.

---

## 🚀 Features

- Modular, Cobra-based CLI
- Streams events to ClickHouse (via HTTP interface)
- Emits Prometheus metrics at `/metrics`
- Traces event generation and ingestion using OpenTelemetry
- Sends OTLP/HTTP traces to Grafana Alloy → Tempo

---

## 🏃‍♀️ Running the Service

Start the service using:

```bash
go run main.go \
  --clickhouse-url=http://localhost:8123 \
  --rate=5 \
  --duration=30
```

**Flags:**

| Flag               | Description                  | Default                 |
|--------------------|------------------------------|-------------------------|
| `--clickhouse-url` | ClickHouse HTTP endpoint     | `http://localhost:8123` |
| `--rate`           | Events per second            | `5`                     |
| `--duration`       | Duration to run (in seconds) | `30`                    |

---

## 📊 Metrics (Prometheus)

The service exposes a `/metrics` endpoint at port `8080`:

```bash
http://localhost:8080/metrics
```

**Metric exposed:**

- `ingestion_generated_events_total` – counter of generated events

Prometheus scrape config example:

```yaml
scrape_configs:
  - job_name: 'ingestion-service'
    static_configs:
      - targets: [ 'host.docker.internal:8080' ]
```

---

## 📈 Traces (OpenTelemetry)

Traces are sent via OTLP/HTTP to `localhost:4318`, typically routed through Grafana Alloy.

**Spans include:**

- `generateEvent`
- `InsertEvent`

Searchable in Grafana Tempo using:

```
service.name = "ingestion-service"
```

---

## 📁 Directory Structure

```
.
├── cmd
│   └── root.go
├── internal
│   ├── generator
│   │   └── generator.go
│   ├── ingestion
│   │   └── clickhouse.go
│   ├── metrics
│   │   └── metrics.go
│   ├── model
│   │   └── event.go
│   └── tracing
│       └── tracing.go
├── go.mod
├── go.sum
├── main.go
```

---

## 🛠️ Development

Install dependencies:

```bash
go mod tidy
```

Run locally with trace + metric output:

```bash
go run main.go --clickhouse-url=http://localhost:8123
```

---

## 👤 Author

[Florita Nichols](https://www.linkedin.com/in/floritanichols)
