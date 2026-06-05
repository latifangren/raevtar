package model

import "time"

const (
	PageKeyAbout   = "about"
	PageKeyContact = "contact"
)

type PageContent struct {
	Key         string    `db:"key" json:"key"`
	Title       string    `db:"title" json:"title"`
	Summary     string    `db:"summary" json:"summary"`
	ContentMD   string    `db:"content_md" json:"content_md"`
	ContentHTML string    `db:"-" json:"content_html"`
	UpdatedAt   time.Time `db:"updated_at" json:"updated_at"`
}
