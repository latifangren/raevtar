package model

import "time"

type MediaAsset struct {
	ID           int64     `db:"id" json:"id"`
	OriginalName string    `db:"original_name" json:"original_name"`
	StoredName   string    `db:"stored_name" json:"stored_name"`
	URL          string    `db:"url" json:"url"`
	MimeType     string    `db:"mime_type" json:"mime_type"`
	SizeBytes    int64     `db:"size_bytes" json:"size_bytes"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
}
