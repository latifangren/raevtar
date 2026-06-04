package repo

import (
	"fmt"
	"strings"
	"time"

	"raevtar/internal/model"

	"github.com/jmoiron/sqlx"
)

type EditorialInboxRepo struct{ db *sqlx.DB }

type EditorialInboxFilter struct {
	Status string
	Mode   string
	Ready  bool
	Now    time.Time
	Limit  int
	Offset int
}

func (r *EditorialInboxRepo) Create(item *model.EditorialInboxItem) error {
	result, err := r.db.Exec(`
		INSERT INTO editorial_inbox (
			source_type, source_value, category_hint, priority, not_before, deadline, note, mode, status, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		item.SourceType,
		item.SourceValue,
		item.CategoryHint,
		item.Priority,
		item.NotBefore.UTC(),
		item.Deadline,
		item.Note,
		item.Mode,
		item.Status,
		item.CreatedAt.UTC(),
		item.UpdatedAt.UTC(),
	)
	if err != nil {
		return err
	}
	id, _ := result.LastInsertId()
	item.ID = id
	return nil
}

func (r *EditorialInboxRepo) GetByID(id int64) (*model.EditorialInboxItem, error) {
	var item model.EditorialInboxItem
	if err := r.db.Get(&item, `SELECT * FROM editorial_inbox WHERE id = ?`, id); err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *EditorialInboxRepo) Update(item *model.EditorialInboxItem) error {
	_, err := r.db.Exec(`
		UPDATE editorial_inbox
		SET source_type = ?, source_value = ?, category_hint = ?, priority = ?, not_before = ?, deadline = ?, note = ?, mode = ?, status = ?, updated_at = ?
		WHERE id = ?`,
		item.SourceType,
		item.SourceValue,
		item.CategoryHint,
		item.Priority,
		item.NotBefore.UTC(),
		item.Deadline,
		item.Note,
		item.Mode,
		item.Status,
		item.UpdatedAt.UTC(),
		item.ID,
	)
	return err
}

func (r *EditorialInboxRepo) List(filter EditorialInboxFilter) ([]model.EditorialInboxItem, error) {
	query := `SELECT * FROM editorial_inbox WHERE 1=1`
	args := make([]any, 0, 4)
	conditions := make([]string, 0, 4)
	if filter.Status != "" {
		conditions = append(conditions, "status = ?")
		args = append(args, filter.Status)
	}
	if filter.Mode != "" {
		conditions = append(conditions, "mode = ?")
		args = append(args, filter.Mode)
	}
	if filter.Ready {
		conditions = append(conditions, "status = ?")
		args = append(args, model.EditorialStatusApproved)
		conditions = append(conditions, "not_before <= ?")
		args = append(args, filter.Now.UTC())
	}
	if len(conditions) > 0 {
		query += " AND " + strings.Join(conditions, " AND ")
	}
	if filter.Ready {
		query += " ORDER BY priority DESC, CASE WHEN deadline IS NULL THEN 1 ELSE 0 END ASC, deadline ASC, created_at ASC"
	} else {
		query += " ORDER BY updated_at DESC, created_at DESC"
	}
	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", filter.Limit)
		if filter.Offset > 0 {
			query += fmt.Sprintf(" OFFSET %d", filter.Offset)
		}
	}
	items := []model.EditorialInboxItem{}
	if err := r.db.Select(&items, query, args...); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *EditorialInboxRepo) CountByStatus() (map[string]int, error) {
	rows, err := r.db.Queryx(`SELECT status, COUNT(*) FROM editorial_inbox GROUP BY status`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	counts := map[string]int{}
	for rows.Next() {
		var status string
		var count int
		if err := rows.Scan(&status, &count); err != nil {
			return nil, err
		}
		counts[status] = count
	}
	return counts, rows.Err()
}
