package repo

import (
	"database/sql"
	"time"

	"raevtar/internal/model"

	"github.com/jmoiron/sqlx"
)

type CategoryRepo struct{ db *sqlx.DB }

func (r *CategoryRepo) List() ([]model.Category, error) {
	var cats []model.Category
	return cats, r.db.Select(&cats, "SELECT * FROM categories ORDER BY name")
}

func (r *CategoryRepo) GetBySlug(slug string) (*model.Category, error) {
	var cat model.Category
	err := r.db.Get(&cat, "SELECT * FROM categories WHERE slug = ?", slug)
	if err != nil {
		return nil, err
	}
	return &cat, nil
}

func (r *CategoryRepo) Create(c *model.Category) error {
	result, err := r.db.Exec(
		"INSERT INTO categories (slug, name, description, created_at, updated_at) VALUES (?, ?, ?, ?, ?)",
		c.Slug, c.Name, c.Description, c.CreatedAt, c.UpdatedAt,
	)
	if err != nil {
		return err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	c.ID = id
	return nil
}

func (r *CategoryRepo) Seed(c *model.Category) error {
	if c.CreatedAt.IsZero() {
		c.CreatedAt = time.Now()
	}
	if c.UpdatedAt.IsZero() {
		c.UpdatedAt = c.CreatedAt
	}
	_, err := r.db.Exec(
		"INSERT OR IGNORE INTO categories (slug, name, description, created_at, updated_at) VALUES (?, ?, ?, ?, ?)",
		c.Slug, c.Name, c.Description, c.CreatedAt, c.UpdatedAt,
	)
	if err != nil {
		return err
	}
	stored, err := r.GetBySlug(c.Slug)
	if err != nil {
		return err
	}
	*c = *stored
	return nil
}

func (r *CategoryRepo) GetByID(id int64) (*model.Category, error) {
	var cat model.Category
	if err := r.db.Get(&cat, "SELECT * FROM categories WHERE id = ?", id); err != nil {
		return nil, err
	}
	return &cat, nil
}

func (r *CategoryRepo) SlugExists(slug string) (bool, error) {
	var count int
	if err := r.db.Get(&count, "SELECT COUNT(*) FROM categories WHERE slug = ?", slug); err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *CategoryRepo) SlugExistsExcludingID(slug string, id int64) (bool, error) {
	var count int
	if err := r.db.Get(&count, "SELECT COUNT(*) FROM categories WHERE slug = ? AND id != ?", slug, id); err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *CategoryRepo) Update(c *model.Category) error {
	result, err := r.db.Exec(
		"UPDATE categories SET slug = ?, name = ?, description = ?, updated_at = ? WHERE id = ?",
		c.Slug, c.Name, c.Description, c.UpdatedAt, c.ID,
	)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *CategoryRepo) Delete(id int64) error {
	result, err := r.db.Exec("DELETE FROM categories WHERE id = ?", id)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}
