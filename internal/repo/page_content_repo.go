package repo

import (
	"strings"

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

type PageContentSearchOptions struct {
	Keys   []string
	Query  string
	Limit  int
	Offset int
}

func (r *PageContentRepo) Search(opts PageContentSearchOptions) ([]model.PageContent, error) {
	if len(opts.Keys) == 0 {
		return nil, nil
	}
	query := `SELECT * FROM page_contents WHERE key IN (` + placeholders(len(opts.Keys)) + `)`
	args := make([]interface{}, 0, len(opts.Keys)+5)
	for _, key := range opts.Keys {
		args = append(args, key)
	}
	if pattern := likePattern(opts.Query); pattern != "" {
		query += ` AND (LOWER(title) LIKE LOWER(?) ESCAPE '\' OR LOWER(summary) LIKE LOWER(?) ESCAPE '\' OR LOWER(content_md) LIKE LOWER(?) ESCAPE '\' OR LOWER(key) LIKE LOWER(?) ESCAPE '\')`
		args = append(args, pattern, pattern, pattern, pattern)
	}
	query += ` ORDER BY key ASC LIMIT ? OFFSET ?`
	args = append(args, opts.Limit, opts.Offset)

	var pages []model.PageContent
	if err := r.db.Select(&pages, query, args...); err != nil {
		return nil, err
	}
	return pages, nil
}

func (r *PageContentRepo) CountSearch(keys []string, queryText string) (int, error) {
	if len(keys) == 0 {
		return 0, nil
	}
	query := `SELECT COUNT(*) FROM page_contents WHERE key IN (` + placeholders(len(keys)) + `)`
	args := make([]interface{}, 0, len(keys)+4)
	for _, key := range keys {
		args = append(args, key)
	}
	if pattern := likePattern(queryText); pattern != "" {
		query += ` AND (LOWER(title) LIKE LOWER(?) ESCAPE '\' OR LOWER(summary) LIKE LOWER(?) ESCAPE '\' OR LOWER(content_md) LIKE LOWER(?) ESCAPE '\' OR LOWER(key) LIKE LOWER(?) ESCAPE '\')`
		args = append(args, pattern, pattern, pattern, pattern)
	}
	var count int
	if err := r.db.Get(&count, query, args...); err != nil {
		return 0, err
	}
	return count, nil
}

func placeholders(count int) string {
	parts := make([]string, count)
	for i := range parts {
		parts[i] = "?"
	}
	return strings.Join(parts, ", ")
}
