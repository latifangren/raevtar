package model

import "time"

type Project struct {
	ID            int64     `db:"id" json:"id"`
	Title         string    `db:"title" json:"title"`
	Slug          string    `db:"slug" json:"slug"`
	ContentMD     string    `db:"content_md" json:"content_md"`
	ContentHTML   string    `db:"-" json:"content_html"`
	Excerpt       string    `db:"excerpt" json:"excerpt"`
	CoverImageURL string    `db:"cover_image_url" json:"cover_image_url,omitempty"`
	Published     bool      `db:"published" json:"published"`
	CreatedAt     time.Time `db:"created_at" json:"created_at"`
	UpdatedAt     time.Time `db:"updated_at" json:"updated_at"`

	Tags []Tag `db:"-" json:"tags,omitempty"`
}

type ProjectCreate struct {
	Title         string   `json:"title"`
	ContentMD     string   `json:"content_md"`
	Excerpt       string   `json:"excerpt"`
	CoverImageURL string   `json:"cover_image_url"`
	Published     bool     `json:"published"`
	Tags          []string `json:"tags"`
}

type ProjectUpdate struct {
	Title         string
	ContentMD     string
	Excerpt       string
	CoverImageURL string
	Published     bool
	Tags          []string
}
