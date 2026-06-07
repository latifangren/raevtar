package repo

import (
	"raevtar/internal/model"

	"github.com/jmoiron/sqlx"
)

type ProjectUpdateRepo struct{ db *sqlx.DB }

type ProjectUpdateListOptions struct {
	ProjectID     int64
	PublishedOnly bool
	Kinds         []string
	Limit         int
}

func (r *ProjectUpdateRepo) List(opts ProjectUpdateListOptions) ([]model.ProjectUpdateEntry, error) {
	query := `SELECT * FROM project_updates WHERE project_id = ?`
	args := make([]interface{}, 0, 4)
	args = append(args, opts.ProjectID)
	if opts.PublishedOnly {
		query += ` AND published = 1`
	}
	if len(opts.Kinds) > 0 {
		query += ` AND kind IN (`
		for i, kind := range opts.Kinds {
			if i > 0 {
				query += `,`
			}
			query += `?`
			args = append(args, kind)
		}
		query += `)`
	}
	query += ` ORDER BY pinned DESC, event_at DESC, sort_order ASC, id DESC`
	if opts.Limit > 0 {
		query += ` LIMIT ?`
		args = append(args, opts.Limit)
	}
	var items []model.ProjectUpdateEntry
	if err := r.db.Select(&items, query, args...); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *ProjectUpdateRepo) GetByID(id int64) (*model.ProjectUpdateEntry, error) {
	var item model.ProjectUpdateEntry
	if err := r.db.Get(&item, `SELECT * FROM project_updates WHERE id = ?`, id); err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *ProjectUpdateRepo) Create(item *model.ProjectUpdateEntry) error {
	result, err := r.db.Exec(`
		INSERT INTO project_updates (project_id, kind, title, content_md, published, pinned, sort_order, event_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		item.ProjectID, item.Kind, item.Title, item.ContentMD, item.Published, item.Pinned, item.SortOrder, item.EventAt,
	)
	if err != nil {
		return err
	}
	id, _ := result.LastInsertId()
	item.ID = id
	return nil
}

func (r *ProjectUpdateRepo) Update(item *model.ProjectUpdateEntry) error {
	_, err := r.db.Exec(`
		UPDATE project_updates
		SET kind = ?, title = ?, content_md = ?, published = ?, pinned = ?, sort_order = ?, event_at = ?, updated_at = ?
		WHERE id = ?`,
		item.Kind, item.Title, item.ContentMD, item.Published, item.Pinned, item.SortOrder, item.EventAt, item.UpdatedAt, item.ID,
	)
	return err
}

func (r *ProjectUpdateRepo) Delete(id int64) error {
	_, err := r.db.Exec(`DELETE FROM project_updates WHERE id = ?`, id)
	return err
}
