package repo

import (
	"raevtar/internal/model"

	"github.com/jmoiron/sqlx"
)

type ContentRelationRepo struct{ db *sqlx.DB }

func (r *ContentRelationRepo) ListBySource(sourceType string, sourceID int64) ([]model.ContentRelation, error) {
	var items []model.ContentRelation
	err := r.db.Select(&items, `
		SELECT * FROM content_relations
		WHERE source_type = ? AND source_id = ?
		ORDER BY sort_order ASC, id ASC`, sourceType, sourceID)
	return items, err
}

func (r *ContentRelationRepo) Create(item *model.ContentRelation) error {
	result, err := r.db.Exec(`
		INSERT INTO content_relations (source_type, source_id, target_type, target_id, relation_kind, sort_order)
		VALUES (?, ?, ?, ?, ?, ?)`,
		item.SourceType, item.SourceID, item.TargetType, item.TargetID, item.RelationKind, item.SortOrder,
	)
	if err != nil {
		return err
	}
	id, _ := result.LastInsertId()
	item.ID = id
	return nil
}

func (r *ContentRelationRepo) Delete(id int64) error {
	_, err := r.db.Exec(`DELETE FROM content_relations WHERE id = ?`, id)
	return err
}
