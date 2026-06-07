package model

import "time"

type ProjectShowcaseItem struct {
	ID            int64     `db:"id" json:"id"`
	ProjectID     int64     `db:"project_id" json:"project_id"`
	Kind          string    `db:"kind" json:"kind"`
	Title         string    `db:"title" json:"title"`
	BodyMD        string    `db:"body_md" json:"body_md"`
	BodyHTML      string    `db:"-" json:"body_html"`
	AssetURL      string    `db:"asset_url" json:"asset_url"`
	ExternalURL   string    `db:"external_url" json:"external_url"`
	EmbedProvider string    `db:"embed_provider" json:"embed_provider"`
	EmbedRef      string    `db:"embed_ref" json:"embed_ref"`
	Published     bool      `db:"published" json:"published"`
	SortOrder     int       `db:"sort_order" json:"sort_order"`
	CreatedAt     time.Time `db:"created_at" json:"created_at"`
	UpdatedAt     time.Time `db:"updated_at" json:"updated_at"`
}

type ProjectShowcaseItemCreate struct {
	Kind          string `json:"kind"`
	Title         string `json:"title"`
	BodyMD        string `json:"body_md"`
	AssetURL      string `json:"asset_url"`
	ExternalURL   string `json:"external_url"`
	EmbedProvider string `json:"embed_provider"`
	EmbedRef      string `json:"embed_ref"`
	Published     bool   `json:"published"`
	SortOrder     int    `json:"sort_order"`
}

type ProjectShowcaseItemUpdate struct {
	Kind          string `json:"kind"`
	Title         string `json:"title"`
	BodyMD        string `json:"body_md"`
	AssetURL      string `json:"asset_url"`
	ExternalURL   string `json:"external_url"`
	EmbedProvider string `json:"embed_provider"`
	EmbedRef      string `json:"embed_ref"`
	Published     bool   `json:"published"`
	SortOrder     int    `json:"sort_order"`
}
