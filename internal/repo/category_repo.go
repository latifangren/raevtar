package repo

import (
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
	// idempotent: seed boleh dipanggil berkali-kali
	_, err := r.db.Exec(
		"INSERT OR IGNORE INTO categories (slug, name, description) VALUES (?, ?, ?)",
		c.Slug, c.Name, c.Description,
	)
	return err
}
