package repo

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"raevtar/internal/model"

	"github.com/jmoiron/sqlx"
)

type EditorialInboxRepo struct{ db *sqlx.DB }

type EditorialInboxClaimParams struct {
	Worker         string
	ClaimTokenHash string
	Now            time.Time
	LeaseExpiresAt time.Time
}

type EditorialInboxCompletionParams struct {
	ID              int64
	ClaimTokenHash  string
	PublishedPostID int64
	Now             time.Time
}

type EditorialInboxFailureParams struct {
	ID             int64
	ClaimTokenHash string
	Status         string
	NotBefore      time.Time
	FailureNote    string
	FailureMeta    string
	Now            time.Time
}

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
			source_type, source_value, category_hint, priority, not_before, deadline, note, mode, status, published_post_id, failure_note, failure_meta, claimed_by, claim_token_hash, claimed_at, lease_expires_at, attempt_count, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		item.SourceType,
		item.SourceValue,
		item.CategoryHint,
		item.Priority,
		item.NotBefore.UTC(),
		item.Deadline,
		item.Note,
		item.Mode,
		item.Status,
		item.PublishedPostID,
		item.FailureNote,
		item.FailureMeta,
		item.ClaimedBy,
		item.ClaimTokenHash,
		item.ClaimedAt,
		item.LeaseExpiresAt,
		item.AttemptCount,
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
		SET source_type = ?, source_value = ?, category_hint = ?, priority = ?, not_before = ?, deadline = ?, note = ?, mode = ?, status = ?, published_post_id = ?, failure_note = ?, failure_meta = ?, claimed_by = ?, claim_token_hash = ?, claimed_at = ?, lease_expires_at = ?, attempt_count = ?, updated_at = ?
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
		item.PublishedPostID,
		item.FailureNote,
		item.FailureMeta,
		item.ClaimedBy,
		item.ClaimTokenHash,
		item.ClaimedAt,
		item.LeaseExpiresAt,
		item.AttemptCount,
		item.UpdatedAt.UTC(),
		item.ID,
	)
	return err
}

func (r *EditorialInboxRepo) ClaimNextReady(params EditorialInboxClaimParams) (*model.EditorialInboxItem, error) {
	tx, err := r.db.Beginx()
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	var item model.EditorialInboxItem
	err = tx.Get(&item, `
		SELECT * FROM editorial_inbox
		WHERE (
			(status = ? AND not_before <= ?)
			OR (status = ? AND lease_expires_at IS NOT NULL AND lease_expires_at <= ?)
		)
		ORDER BY priority DESC, CASE WHEN deadline IS NULL THEN 1 ELSE 0 END ASC, deadline ASC, created_at ASC
		LIMIT 1`, model.EditorialStatusApproved, params.Now.UTC(), model.EditorialStatusRunning, params.Now.UTC())
	if err != nil {
		if err == sql.ErrNoRows {
			if commitErr := tx.Commit(); commitErr != nil {
				return nil, commitErr
			}
			return nil, nil
		}
		return nil, err
	}

	result, err := tx.Exec(`
		UPDATE editorial_inbox
		SET status = ?, claimed_by = ?, claim_token_hash = ?, claimed_at = ?, lease_expires_at = ?, attempt_count = attempt_count + 1, updated_at = ?
		WHERE id = ?
		  AND (
			(status = ? AND not_before <= ?)
			OR (status = ? AND lease_expires_at IS NOT NULL AND lease_expires_at <= ?)
		  )`,
		model.EditorialStatusRunning,
		params.Worker,
		params.ClaimTokenHash,
		params.Now.UTC(),
		params.LeaseExpiresAt.UTC(),
		params.Now.UTC(),
		item.ID,
		model.EditorialStatusApproved,
		params.Now.UTC(),
		model.EditorialStatusRunning,
		params.Now.UTC(),
	)
	if err != nil {
		return nil, err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return nil, err
	}
	if rows == 0 {
		if commitErr := tx.Commit(); commitErr != nil {
			return nil, commitErr
		}
		return nil, nil
	}
	if err := tx.Get(&item, `SELECT * FROM editorial_inbox WHERE id = ?`, item.ID); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *EditorialInboxRepo) CompleteClaim(params EditorialInboxCompletionParams) (bool, error) {
	result, err := r.db.Exec(`
		UPDATE editorial_inbox
		SET status = ?, published_post_id = ?, claimed_by = '', claim_token_hash = '', claimed_at = NULL, lease_expires_at = NULL, updated_at = ?
		WHERE id = ? AND status = ? AND claim_token_hash = ?`,
		model.EditorialStatusDone,
		params.PublishedPostID,
		params.Now.UTC(),
		params.ID,
		model.EditorialStatusRunning,
		params.ClaimTokenHash,
	)
	if err != nil {
		return false, err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	return rows > 0, nil
}

func (r *EditorialInboxRepo) FailClaim(params EditorialInboxFailureParams) (bool, error) {
	result, err := r.db.Exec(`
		UPDATE editorial_inbox
		SET status = ?, not_before = ?, failure_note = ?, failure_meta = ?, published_post_id = NULL, claimed_by = '', claim_token_hash = '', claimed_at = NULL, lease_expires_at = NULL, updated_at = ?
		WHERE id = ? AND status = ? AND claim_token_hash = ?`,
		params.Status,
		params.NotBefore.UTC(),
		params.FailureNote,
		params.FailureMeta,
		params.Now.UTC(),
		params.ID,
		model.EditorialStatusRunning,
		params.ClaimTokenHash,
	)
	if err != nil {
		return false, err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	return rows > 0, nil
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
