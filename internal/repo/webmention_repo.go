package repo

import (
	"raevtar/internal/model"

	"github.com/jmoiron/sqlx"
)

type WebmentionRepo struct{ db *sqlx.DB }

func (r *WebmentionRepo) Create(w *model.Webmention) error {
	result, err := r.db.Exec(`
		INSERT INTO webmentions (source_url, target_url, post_id, title, author, author_url, content, approved)
		VALUES (?, ?, ?, ?, ?, ?, ?, 0)`,
		w.SourceURL, w.TargetURL, w.PostID, w.Title, w.Author, w.AuthorURL, w.Content,
	)
	if err != nil {
		return err
	}
	id, _ := result.LastInsertId()
	w.ID = id
	return nil
}

func (r *WebmentionRepo) ListByPost(postID int64, approvedOnly bool) ([]model.Webmention, error) {
	query := "SELECT * FROM webmentions WHERE post_id = ?"
	if approvedOnly {
		query += " AND approved = 1"
	}
	query += " ORDER BY created_at DESC"

	var mentions []model.Webmention
	if err := r.db.Select(&mentions, query, postID); err != nil {
		return nil, err
	}
	return mentions, nil
}

func (r *WebmentionRepo) ListAll(limit int) ([]model.Webmention, error) {
	var mentions []model.Webmention
	if err := r.db.Select(&mentions, "SELECT * FROM webmentions ORDER BY created_at DESC LIMIT ?", limit); err != nil {
		return nil, err
	}
	return mentions, nil
}

func (r *WebmentionRepo) Approve(id int64) error {
	_, err := r.db.Exec("UPDATE webmentions SET approved = 1 WHERE id = ?", id)
	return err
}

func (r *WebmentionRepo) Delete(id int64) error {
	_, err := r.db.Exec("DELETE FROM webmentions WHERE id = ?", id)
	return err
}
