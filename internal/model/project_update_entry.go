package model

import "time"

type ProjectUpdateEntry struct {
	ID          int64     `db:"id" json:"id"`
	ProjectID   int64     `db:"project_id" json:"project_id"`
	Kind        string    `db:"kind" json:"kind"`
	Title       string    `db:"title" json:"title"`
	ContentMD   string    `db:"content_md" json:"content_md"`
	ContentHTML string    `db:"-" json:"content_html"`
	Published   bool      `db:"published" json:"published"`
	Pinned      bool      `db:"pinned" json:"pinned"`
	SortOrder   int       `db:"sort_order" json:"sort_order"`
	EventAt     time.Time `db:"event_at" json:"event_at"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time `db:"updated_at" json:"updated_at"`
}

type ProjectUpdateEntryCreate struct {
	Kind      string    `json:"kind"`
	Title     string    `json:"title"`
	ContentMD string    `json:"content_md"`
	Published bool      `json:"published"`
	Pinned    bool      `json:"pinned"`
	SortOrder int       `json:"sort_order"`
	EventAt   time.Time `json:"event_at"`
}

type ProjectUpdateEntryUpdate struct {
	Kind      string    `json:"kind"`
	Title     string    `json:"title"`
	ContentMD string    `json:"content_md"`
	Published bool      `json:"published"`
	Pinned    bool      `json:"pinned"`
	SortOrder int       `json:"sort_order"`
	EventAt   time.Time `json:"event_at"`
}
