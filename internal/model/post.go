package model

import "time"

type Post struct {
	ID            int64      `db:"id" json:"id"`
	CategoryID    int64      `db:"category_id" json:"category_id"`
	Title         string     `db:"title" json:"title"`
	Slug          string     `db:"slug" json:"slug"`
	ContentMD     string     `db:"content_md" json:"content_md"`
	ContentHTML   string     `db:"-" json:"content_html"`
	Excerpt       string     `db:"excerpt" json:"excerpt"`
	CoverImageURL string     `db:"cover_image_url" json:"cover_image_url,omitempty"`
	Published     bool       `db:"published" json:"published"`
	CreatedAt     time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt     time.Time  `db:"updated_at" json:"updated_at"`
	ScheduledAt   *time.Time `db:"scheduled_at" json:"scheduled_at,omitempty"`

	CategoryName string `db:"category_name" json:"category_name,omitempty"`
	CategorySlug string `db:"category_slug" json:"category_slug,omitempty"`

	Tags []Tag `db:"-" json:"tags,omitempty"`
}

// PostCreate adalah input buat bikin post baru (via API atau cron)
type PostCreate struct {
	CategorySlug  string     `json:"category_slug"`
	Title         string     `json:"title"`
	ContentMD     string     `json:"content_md"`
	Excerpt       string     `json:"excerpt"`
	CoverImageURL string     `json:"cover_image_url"`
	Published     bool       `json:"published"`
	ScheduledAt   *time.Time `json:"scheduled_at,omitempty"`
	Tags          []string   `json:"tags"` // tag names, created if not exist
}

type PostUpdate struct {
	CategorySlug  string
	Title         string
	ContentMD     string
	Excerpt       string
	CoverImageURL string
	Published     bool
	ScheduledAt   *time.Time
	Tags          []string
}
