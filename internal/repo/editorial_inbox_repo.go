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
	ItemID         int64
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

type EditorialModeCountRow struct {
	Mode  string `db:"mode"`
	Count int    `db:"count"`
}

type EditorialAnalyticsRow struct {
	DoneCount               int   `db:"done_count"`
	FailedCount             int   `db:"failed_count"`
	CompletedWithPostCount  int   `db:"completed_with_post_count"`
	AverageQueueWaitSeconds int64 `db:"average_queue_wait_seconds"`
	AverageReadyWaitSeconds int64 `db:"average_ready_wait_seconds"`
	OverdueCompletedCount   int   `db:"overdue_completed_count"`
}

func (r *EditorialInboxRepo) Create(item *model.EditorialInboxItem) error {
	result, err := r.db.Exec(`
		INSERT INTO editorial_inbox (
			source_type, source_value, category_hint, priority, not_before, deadline, note, mode, status, published_post_id, failure_note, failure_meta, claimed_by, claim_token_hash, claimed_at, lease_expires_at, attempt_count, completed_at, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
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
		item.CompletedAt,
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
		SET source_type = ?, source_value = ?, category_hint = ?, priority = ?, not_before = ?, deadline = ?, note = ?, mode = ?, status = ?, published_post_id = ?, failure_note = ?, failure_meta = ?, claimed_by = ?, claim_token_hash = ?, claimed_at = ?, lease_expires_at = ?, attempt_count = ?, completed_at = ?, updated_at = ?
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
		item.CompletedAt,
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
		coalesceClaimItemID(params.ItemID, item.ID),
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
	if err := tx.Get(&item, `SELECT * FROM editorial_inbox WHERE id = ?`, coalesceClaimItemID(params.ItemID, item.ID)); err != nil {
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
		SET status = ?, published_post_id = ?, claimed_by = '', claim_token_hash = '', claimed_at = NULL, lease_expires_at = NULL, completed_at = ?, updated_at = ?
		WHERE id = ? AND status = ? AND claim_token_hash = ?`,
		model.EditorialStatusDone,
		params.PublishedPostID,
		params.Now.UTC(),
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

func coalesceClaimItemID(explicitID, fallbackID int64) int64 {
	if explicitID > 0 {
		return explicitID
	}
	return fallbackID
}

func (r *EditorialInboxRepo) FailClaim(params EditorialInboxFailureParams) (bool, error) {
	result, err := r.db.Exec(`
		UPDATE editorial_inbox
		SET status = ?, not_before = ?, failure_note = ?, failure_meta = ?, published_post_id = NULL, claimed_by = '', claim_token_hash = '', claimed_at = NULL, lease_expires_at = NULL, completed_at = NULL, updated_at = ?
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

func (r *EditorialInboxRepo) GetPolicyState(name string) (int, error) {
	var value int
	err := r.db.Get(&value, `SELECT value FROM editorial_policy_state WHERE name = ?`, name)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		return 0, err
	}
	return value, nil
}

func (r *EditorialInboxRepo) SetPolicyState(name string, value int, now time.Time) error {
	_, err := r.db.Exec(`
		INSERT INTO editorial_policy_state (name, value, updated_at)
		VALUES (?, ?, ?)
		ON CONFLICT(name) DO UPDATE SET value = excluded.value, updated_at = excluded.updated_at`,
		name,
		value,
		now.UTC(),
	)
	return err
}

func (r *EditorialInboxRepo) CountOverdue(now time.Time) (model.EditorialOverdueSummary, error) {
	var summary model.EditorialOverdueSummary
	if err := r.db.Get(&summary, `
		SELECT
			COALESCE(SUM(CASE WHEN status = 'approved' AND deadline IS NOT NULL AND deadline < ? THEN 1 ELSE 0 END), 0) AS approved_count,
			COALESCE(SUM(CASE WHEN status = 'running' AND deadline IS NOT NULL AND deadline < ? THEN 1 ELSE 0 END), 0) AS running_count,
			COALESCE(SUM(CASE WHEN status = 'done' AND deadline IS NOT NULL AND completed_at IS NOT NULL AND completed_at > deadline THEN 1 ELSE 0 END), 0) AS completed_count
		FROM editorial_inbox`, now.UTC(), now.UTC()); err != nil {
		return model.EditorialOverdueSummary{}, err
	}
	return summary, nil
}

func (r *EditorialInboxRepo) GetAnalytics() (EditorialAnalyticsRow, error) {
	var row EditorialAnalyticsRow
	err := r.db.Get(&row, `
		SELECT
			COALESCE(SUM(CASE WHEN status = 'done' THEN 1 ELSE 0 END), 0) AS done_count,
			COALESCE(SUM(CASE WHEN status = 'failed' THEN 1 ELSE 0 END), 0) AS failed_count,
			COALESCE(SUM(CASE WHEN status = 'done' AND published_post_id IS NOT NULL THEN 1 ELSE 0 END), 0) AS completed_with_post_count,
			COALESCE(CAST(AVG(CASE WHEN status = 'done' AND completed_at IS NOT NULL THEN (julianday(completed_at) - julianday(created_at)) * 86400 END) AS INTEGER), 0) AS average_queue_wait_seconds,
			COALESCE(CAST(AVG(CASE WHEN status = 'done' AND completed_at IS NOT NULL THEN (julianday(completed_at) - julianday(not_before)) * 86400 END) AS INTEGER), 0) AS average_ready_wait_seconds,
			COALESCE(SUM(CASE WHEN status = 'done' AND deadline IS NOT NULL AND completed_at IS NOT NULL AND completed_at > deadline THEN 1 ELSE 0 END), 0) AS overdue_completed_count
		FROM editorial_inbox`)
	return row, err
}

func (r *EditorialInboxRepo) CountDoneByMode() ([]EditorialModeCountRow, error) {
	rows := []EditorialModeCountRow{}
	if err := r.db.Select(&rows, `SELECT mode, COUNT(*) AS count FROM editorial_inbox WHERE status = ? GROUP BY mode ORDER BY mode ASC`, model.EditorialStatusDone); err != nil {
		return nil, err
	}
	return rows, nil
}
