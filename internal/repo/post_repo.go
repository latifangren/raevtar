package repo

import (
	"strings"

	"raevtar/internal/model"

	"github.com/jmoiron/sqlx"
)

type PostRepo struct{ db *sqlx.DB }

func (r *PostRepo) List(categorySlug string, publishedOnly bool, limit, offset int) ([]model.Post, error) {
	query := `
		SELECT p.*, c.name AS category_name, c.slug AS category_slug
		FROM posts p JOIN categories c ON p.category_id = c.id
		WHERE 1=1`
	var args []interface{}
	if categorySlug != "" {
		query += " AND c.slug = ?"
		args = append(args, categorySlug)
	}
	if publishedOnly {
		query += " AND p.published = 1"
	}
	query += " ORDER BY p.created_at DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	var posts []model.Post
	if err := r.db.Select(&posts, query, args...); err != nil {
		return nil, err
	}

	// load tags for all posts
	tagRepo := &TagRepo{db: r.db}
	ids := make([]int64, len(posts))
	for i, p := range posts {
		ids[i] = p.ID
	}
	tagMap, err := tagRepo.GetByPostIDs(ids)
	if err != nil {
		return nil, err
	}
	for i, p := range posts {
		posts[i].Tags = tagMap[p.ID]
	}

	return posts, nil
}

func (r *PostRepo) GetBySlug(slug string) (*model.Post, error) {
	query := `
		SELECT p.*, c.name AS category_name, c.slug AS category_slug
		FROM posts p JOIN categories c ON p.category_id = c.id
		WHERE p.slug = ?`
	var post model.Post
	if err := r.db.Get(&post, query, slug); err != nil {
		return nil, err
	}

	// load tags
	tagRepo := &TagRepo{db: r.db}
	tags, err := tagRepo.GetByPostID(post.ID)
	if err != nil {
		return nil, err
	}
	post.Tags = tags

	return &post, nil
}

func (r *PostRepo) GetByID(id int64) (*model.Post, error) {
	query := `
		SELECT p.*, c.name AS category_name, c.slug AS category_slug
		FROM posts p JOIN categories c ON p.category_id = c.id
		WHERE p.id = ?`
	var post model.Post
	if err := r.db.Get(&post, query, id); err != nil {
		return nil, err
	}

	tagRepo := &TagRepo{db: r.db}
	tags, err := tagRepo.GetByPostID(post.ID)
	if err != nil {
		return nil, err
	}
	post.Tags = tags

	return &post, nil
}

func (r *PostRepo) Create(p *model.Post) error {
	result, err := r.db.Exec(`
		INSERT INTO posts (category_id, title, slug, content_md, excerpt, cover_image_url, published)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		p.CategoryID, p.Title, p.Slug, p.ContentMD, p.Excerpt, p.CoverImageURL, p.Published,
	)
	if err != nil {
		return err
	}
	id, _ := result.LastInsertId()
	p.ID = id
	return nil
}

func (r *PostRepo) Update(p *model.Post) error {
	_, err := r.db.Exec(`
		UPDATE posts
		SET category_id = ?, title = ?, content_md = ?, excerpt = ?, cover_image_url = ?, published = ?, updated_at = ?
		WHERE id = ?`,
		p.CategoryID, p.Title, p.ContentMD, p.Excerpt, p.CoverImageURL, p.Published, p.UpdatedAt, p.ID,
	)
	return err
}

func (r *PostRepo) SlugExists(slug string) (bool, error) {
	var count int
	if err := r.db.Get(&count, "SELECT COUNT(*) FROM posts WHERE slug = ?", slug); err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *PostRepo) Delete(id int64) error {
	_, err := r.db.Exec("DELETE FROM posts WHERE id = ?", id)
	return err
}

func (r *PostRepo) Count(categorySlug string, publishedOnly bool) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM posts p JOIN categories c ON p.category_id = c.id`
	conditions := make([]string, 0, 2)
	args := make([]interface{}, 0, 1)
	if categorySlug != "" {
		conditions = append(conditions, "c.slug = ?")
		args = append(args, categorySlug)
	}
	if publishedOnly {
		conditions = append(conditions, "p.published = 1")
	}
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}
	return count, r.db.Get(&count, query, args...)
}
