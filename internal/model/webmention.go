package model

import "time"

type Webmention struct {
	ID        int64     `db:"id" json:"id"`
	SourceURL string    `db:"source_url" json:"source_url"`
	TargetURL string    `db:"target_url" json:"target_url"`
	PostID    int64     `db:"post_id" json:"post_id"`
	Title     string    `db:"title" json:"title,omitempty"`
	Author    string    `db:"author" json:"author,omitempty"`
	AuthorURL string    `db:"author_url" json:"author_url,omitempty"`
	Content   string    `db:"content" json:"content,omitempty"`
	Approved  bool      `db:"approved" json:"approved"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}
