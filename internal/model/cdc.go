package model

import "time"

// --- Debezium payload-only envelope for app.users ---

type DBZSource struct {
	LSN    *uint64 `json:"lsn"`
	TsUS   *int64  `json:"ts_us"`
	Schema string  `json:"schema"`
	Table  string  `json:"table"`
}

type DBZUser struct {
	ID    uint64 `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type DBZEnvelope struct {
	Before *DBZUser `json:"before"`
	After  *DBZUser `json:"after"`
	Source DBZSource `json:"source"`
	Op     string    `json:"op"`    // "c","u","d"
	TsUS   *int64    `json:"ts_us"` // microseconds
}

type DBZKey struct {
	ID uint64 `json:"id"`
}

// Row we insert into ClickHouse app.users_cur
type CDCUserRow struct {
	ID        uint64
	Name      string
	Email     string
	IsDeleted uint8
	Op        uint8      // 1=c,2=u,3=d  (maps to _op)
	LSN       uint64     // _lsn
	Ts        time.Time  // _ts (UTC)
}
