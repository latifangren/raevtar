package model

import "time"

type Post struct {
	ID          int64     `db:"id" json:"id"`
	CategoryID  int64     `db:"category_id" json:"category_id"`
	Title       string    `db:"title" json:"title"`
	Slug        string    `db:"slug" json:"slug"`
	ContentMD   string    `db:"content_md" json:"content_md"`
	ContentHTML string    `db:"-" json:"content_html"`
	Excerpt     string    `db:"excerpt" json:"excerpt"`
	Published   bool      `db:"published" json:"published"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time `db:"updated_at" json:"updated_at"`

	CategoryName string `db:"category_name" json:"category_name,omitempty"`
	CategorySlug string `db:"category_slug" json:"category_slug,omitempty"`

	Tags []Tag `db:"-" json:"tags,omitempty"`
}

// PostCreate adalah input buat bikin post baru (via API atau cron)
type PostCreate struct {
	CategorySlug string   `json:"category_slug"`
	Title        string   `json:"title"`
	ContentMD    string   `json:"content_md"`
	Excerpt      string   `json:"excerpt"`
	Published    bool     `json:"published"`
	Tags         []string `json:"tags"` // tag names, created if not exist
}

type PostUpdate struct {
	CategorySlug string
	Title        string
	ContentMD    string
	Excerpt      string
	Published    bool
	Tags         []string
}
