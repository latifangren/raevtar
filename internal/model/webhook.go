package model

import "time"

type WebhookConfig struct {
	ID        int64     `db:"id" json:"id"`
	Name      string    `db:"name" json:"name"`
	URL       string    `db:"url" json:"url"`
	Secret    string    `db:"secret" json:"-"`
	Enabled   bool      `db:"enabled" json:"enabled"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

type WebhookEvent struct {
	ID           int64     `db:"id" json:"id"`
	WebhookID    int64     `db:"webhook_id" json:"webhook_id"`
	EventType    string    `db:"event_type" json:"event_type"`
	ServerID     int64     `db:"server_id" json:"server_id"`
	Payload      string    `db:"payload" json:"payload"`
	ResponseCode int       `db:"response_code" json:"response_code"`
	ResponseBody string    `db:"response_body" json:"response_body"`
	FiredAt      time.Time `db:"fired_at" json:"fired_at"`
}
