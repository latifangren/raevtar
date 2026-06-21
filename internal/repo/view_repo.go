package repo

import (
	"github.com/jmoiron/sqlx"
)

type ViewRepo struct{ db *sqlx.DB }

// RecordPostView inserts a view event for a post by IP hash.
func (r *ViewRepo) RecordPostView(postID int64, ipHash string) error {
	_, err := r.db.Exec(`
		INSERT INTO post_views (post_id, ip_hash) VALUES (?, ?)`,
		postID, ipHash,
	)
	return err
}

// CountPostViews returns total view count for a post.
func (r *ViewRepo) CountPostViews(postID int64) (int, error) {
	var count int
	err := r.db.Get(&count, `SELECT COUNT(*) FROM post_views WHERE post_id = ?`, postID)
	return count, err
}

// CountAllPostViews returns total views for each post, keyed by post ID.
func (r *ViewRepo) CountAllPostViews() (map[int64]int, error) {
	type row struct {
		PostID int64 `db:"post_id"`
		Count  int   `db:"c"`
	}
	var rows []row
	if err := r.db.Select(&rows, `SELECT post_id, COUNT(*) as c FROM post_views GROUP BY post_id`); err != nil {
		return nil, err
	}
	result := make(map[int64]int, len(rows))
	for _, r := range rows {
		result[r.PostID] = r.Count
	}
	return result, nil
}
