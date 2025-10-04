package ingestion

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"os"
	"strings"
	"time"

	"github.com/segmentio/kafka-go"
	"go.opentelemetry.io/otel"

	"github.com/florita1/ingestion-service/internal/logging"
	"github.com/florita1/ingestion-service/internal/model"
)

type staticResolver struct{ ip string }

func (r staticResolver) LookupHost(ctx context.Context, host string) ([]string, error) {
	return []string{r.ip}, nil
}

func RunCDC(ctx context.Context, cfg Config) error {
	tr := otel.Tracer("ingestion-service")

	var dialer *kafka.Dialer
	if os.Getenv("KAFKA_FORCE_LOCAL") == "1" {
		dialer = &kafka.Dialer{
			Timeout:  10 * time.Second,
			Resolver: staticResolver{ip: "127.0.0.1"},
		}
	}

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  cfg.Brokers,
		Topic:    cfg.Topic,
		GroupID:  cfg.GroupID,
		Dialer:   dialer,
		MinBytes: 1,
		MaxBytes: 10 << 20,
	})
	defer reader.Close()

	log.Printf("[cdc] brokers=%v topic=%s group=%s", cfg.Brokers, cfg.Topic, cfg.GroupID)

	for {
		msg, err := reader.ReadMessage(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return nil
			}
			return err
		}

		// New span per message
		mCtx, span := tr.Start(ctx, "cdc.message")
		span.SetAttributes()
		logging.WithTrace(mCtx, "WAL event received")

		var env model.DBZEnvelope
		if !tryUnmarshalEnvelope(msg.Value, &env) {
			logging.WithTrace(mCtx, "[cdc] bad payload at offset=%d", msg.Offset)
			span.End()
			continue
		}

		row := translateEnvelopeToRow(mCtx, env, msg.Key)
		if row == nil {
			span.End()
			continue
		}
		logging.WithTrace(mCtx, "WAL envelope translated to row op=%d id=%d", row.Op, row.ID)

		if err := InsertCDCUser(mCtx, cfg, *row); err != nil {
			logging.WithTrace(mCtx, "[cdc] insert error: %v", err)
			span.End()
			continue
		}

		logging.WithTrace(mCtx, "[cdc] ok op=%d id=%d", row.Op, row.ID)
		span.End()
	}
}

func tryUnmarshalEnvelope(b []byte, out *model.DBZEnvelope) bool {
	// direct
	if json.Unmarshal(b, out) == nil {
		return true
	}
	// sometimes JSON is stringified: "{"before":...}"
	var s string
	if json.Unmarshal(b, &s) == nil {
		return json.Unmarshal([]byte(s), out) == nil
	}
	return false
}

func translateEnvelopeToRow(ctx context.Context, env model.DBZEnvelope, key []byte) *model.CDCUserRow {
	var lsn uint64
	if env.Source.LSN != nil {
		lsn = *env.Source.LSN
	}

	ts := time.Unix(0, 0).UTC()
	if env.TsUS != nil {
		ts = time.UnixMicro(*env.TsUS).UTC()
	}

	row := model.CDCUserRow{LSN: lsn, Ts: ts}

	switch env.Op {
	case "c", "u":
		if env.After == nil {
			logging.WithTrace(ctx, "[cdc] missing 'after' for op=%s", env.Op)
			return nil
		}
		row.ID = env.After.ID
		row.Name = env.After.Name
		row.Email = env.After.Email
		row.Op = opToEnum(env.Op)
		return &row

	case "d":
		// Prefer payload.before; else parse key
		if env.Before != nil && env.Before.ID != 0 {
			row.ID = env.Before.ID
		} else {
			var k model.DBZKey
			if err := json.Unmarshal(key, &k); err == nil {
				row.ID = k.ID
			}
		}
		row.IsDeleted = 1
		row.Op = 3
		return &row

	default:
		logging.WithTrace(ctx, "[cdc] unknown op=%s", env.Op)
		return nil
	}
}

func opToEnum(op string) uint8 {
	switch strings.ToLower(op) {
	case "c":
		return 1
	case "u":
		return 2
	case "d":
		return 3
	default:
		return 0
	}
}
