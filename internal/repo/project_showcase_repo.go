package repo

import (
	"raevtar/internal/model"

	"github.com/jmoiron/sqlx"
)

type ProjectShowcaseRepo struct{ db *sqlx.DB }

func (r *ProjectShowcaseRepo) ListByProjectID(projectID int64, publishedOnly bool) ([]model.ProjectShowcaseItem, error) {
	query := `SELECT * FROM project_showcase_items WHERE project_id = ?`
	args := []interface{}{projectID}
	if publishedOnly {
		query += ` AND published = 1`
	}
	query += ` ORDER BY sort_order ASC, id ASC`
	var items []model.ProjectShowcaseItem
	err := r.db.Select(&items, query, args...)
	return items, err
}

func (r *ProjectShowcaseRepo) GetByID(id int64) (*model.ProjectShowcaseItem, error) {
	var item model.ProjectShowcaseItem
	if err := r.db.Get(&item, `SELECT * FROM project_showcase_items WHERE id = ?`, id); err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *ProjectShowcaseRepo) Create(item *model.ProjectShowcaseItem) error {
	result, err := r.db.Exec(`
		INSERT INTO project_showcase_items (project_id, kind, title, body_md, asset_url, external_url, embed_provider, embed_ref, published, sort_order)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		item.ProjectID, item.Kind, item.Title, item.BodyMD, item.AssetURL, item.ExternalURL, item.EmbedProvider, item.EmbedRef, item.Published, item.SortOrder,
	)
	if err != nil {
		return err
	}
	id, _ := result.LastInsertId()
	item.ID = id
	return nil
}

func (r *ProjectShowcaseRepo) Update(item *model.ProjectShowcaseItem) error {
	_, err := r.db.Exec(`
		UPDATE project_showcase_items
		SET kind = ?, title = ?, body_md = ?, asset_url = ?, external_url = ?, embed_provider = ?, embed_ref = ?, published = ?, sort_order = ?, updated_at = ?
		WHERE id = ?`,
		item.Kind, item.Title, item.BodyMD, item.AssetURL, item.ExternalURL, item.EmbedProvider, item.EmbedRef, item.Published, item.SortOrder, item.UpdatedAt, item.ID,
	)
	return err
}

func (r *ProjectShowcaseRepo) Delete(id int64) error {
	_, err := r.db.Exec(`DELETE FROM project_showcase_items WHERE id = ?`, id)
	return err
}
