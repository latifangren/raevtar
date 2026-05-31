package repo

import (
	"fmt"
	"strings"

	"raevtar/internal/model"

	"github.com/jmoiron/sqlx"
)

type TagRepo struct{ db *sqlx.DB }

// GetByPostID returns tags for a given post.
func (r *TagRepo) GetByPostID(postID int64) ([]model.Tag, error) {
	var tags []model.Tag
	err := r.db.Select(&tags, `
		SELECT t.* FROM tags t
		JOIN post_tags pt ON pt.tag_id = t.id
		WHERE pt.post_id = ?
		ORDER BY t.name`, postID)
	return tags, err
}

// GetByPostIDs returns tags for multiple posts, keyed by post ID.
func (r *TagRepo) GetByPostIDs(postIDs []int64) (map[int64][]model.Tag, error) {
	if len(postIDs) == 0 {
		return map[int64][]model.Tag{}, nil
	}

	// Build placeholder string
	placeholders := make([]string, len(postIDs))
	args := make([]interface{}, len(postIDs))
	for i, id := range postIDs {
		placeholders[i] = "?"
		args[i] = id
	}

	rows, err := r.db.Query(`
		SELECT pt.post_id, t.id, t.name, t.slug, t.created_at
		FROM post_tags pt
		JOIN tags t ON t.id = pt.tag_id
		WHERE pt.post_id IN (`+strings.Join(placeholders, ",")+`)
		ORDER BY t.name`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[int64][]model.Tag)
	for rows.Next() {
		var postID int64
		var t model.Tag
		if err := rows.Scan(&postID, &t.ID, &t.Name, &t.Slug, &t.CreatedAt); err != nil {
			return nil, err
		}
		result[postID] = append(result[postID], t)
	}
	return result, rows.Err()
}

// Ensure creates a tag if it doesn't exist, returns existing one if it does.
func (r *TagRepo) Ensure(name string) (*model.Tag, error) {
	slug := makeSlug(name)
	if slug == "" {
		return nil, fmt.Errorf("invalid tag name: %q", name)
	}

	// Try insert, ignore conflict
	_, err := r.db.Exec(
		"INSERT INTO tags (name, slug) VALUES (?, ?) ON CONFLICT(slug) DO NOTHING",
		name, slug,
	)
	if err != nil {
		return nil, err
	}

	var t model.Tag
	err = r.db.Get(&t, "SELECT * FROM tags WHERE slug = ?", slug)
	return &t, err
}

// SetTags replaces all tags for a post. Creates tags if needed.
func (r *TagRepo) SetTags(postID int64, tagNames []string) error {
	tx, err := r.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Remove existing
	if _, err := tx.Exec("DELETE FROM post_tags WHERE post_id = ?", postID); err != nil {
		return err
	}

	// Add new
	for _, name := range tagNames {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		slug := makeSlug(name)
		if slug == "" {
			continue
		}

		// Ensure tag exists
		_, err := tx.Exec(
			"INSERT INTO tags (name, slug) VALUES (?, ?) ON CONFLICT(slug) DO NOTHING",
			name, slug,
		)
		if err != nil {
			return err
		}

		// Get tag ID
		var tagID int64
		if err := tx.Get(&tagID, "SELECT id FROM tags WHERE slug = ?", slug); err != nil {
			return err
		}

		// Link
		if _, err := tx.Exec(
			"INSERT INTO post_tags (post_id, tag_id) VALUES (?, ?) ON CONFLICT DO NOTHING",
			postID, tagID,
		); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func makeSlug(s string) string {
	slug := strings.ToLower(s)
	slug = strings.ReplaceAll(slug, " ", "-")
	// Remove non-alphanumeric except hyphens
	var result strings.Builder
	for _, r := range slug {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			result.WriteRune(r)
		}
	}
	return strings.Trim(result.String(), "-")
}
