package repo

import (
	"raevtar/internal/model"

	"github.com/jmoiron/sqlx"
)

type PageContentRepo struct{ db *sqlx.DB }

func (r *PageContentRepo) GetByKey(key string) (*model.PageContent, error) {
	var page model.PageContent
	if err := r.db.Get(&page, `SELECT * FROM page_contents WHERE key = ?`, key); err != nil {
		return nil, err
	}
	return &page, nil
}

func (r *PageContentRepo) List() ([]model.PageContent, error) {
	var pages []model.PageContent
	if err := r.db.Select(&pages, `SELECT * FROM page_contents ORDER BY key ASC`); err != nil {
		return nil, err
	}
	return pages, nil
}

func (r *PageContentRepo) Upsert(page *model.PageContent) error {
	_, err := r.db.Exec(`
		INSERT INTO page_contents (key, title, summary, content_md, updated_at)
		VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(key) DO UPDATE SET
			title = excluded.title,
			summary = excluded.summary,
			content_md = excluded.content_md,
			updated_at = CURRENT_TIMESTAMP`,
		page.Key, page.Title, page.Summary, page.ContentMD,
	)
	return err
}
