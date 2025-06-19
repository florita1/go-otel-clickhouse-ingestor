package model

import "time"

type Event struct {
	Timestamp time.Time `json:"timestamp"`
	UserID    string    `json:"user_id"`
	Action    string    `json:"action"`
	Payload   string    `json:"payload"`
}
