package repo

import (
	"time"

	"raevtar/internal/model"

	"github.com/jmoiron/sqlx"
)

type AuditRepo struct{ db *sqlx.DB }

func (r *AuditRepo) Insert(user, action, details, ip string) error {
	_, err := r.db.Exec(
		"INSERT INTO audit_logs (user, action, details, ip, created_at) VALUES (?, ?, ?, ?, ?)",
		user, action, details, ip, time.Now(),
	)
	return err
}

func (r *AuditRepo) List(limit, offset int) ([]model.AuditLog, error) {
	if limit <= 0 {
		limit = 50
	}
	var logs []model.AuditLog
	err := r.db.Select(&logs, "SELECT * FROM audit_logs ORDER BY created_at DESC LIMIT ? OFFSET ?", limit, offset)
	return logs, err
}

func (r *AuditRepo) Count() (int, error) {
	var count int
	err := r.db.Get(&count, "SELECT COUNT(*) FROM audit_logs")
	return count, err
}
