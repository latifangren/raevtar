package model

import "time"

type Category struct {
	ID          int64     `db:"id" json:"id"`
	Slug        string    `db:"slug" json:"slug"`               // "ai-agent"
	Name        string    `db:"name" json:"name"`               // "AI Agent"
	Description string    `db:"description" json:"description"` // optional
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time `db:"updated_at" json:"updated_at"`
}
