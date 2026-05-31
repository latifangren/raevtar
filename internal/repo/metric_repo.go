package repo

import (
	"raevtar/internal/model"

	"github.com/jmoiron/sqlx"
)

type MetricRepo struct{ db *sqlx.DB }

func (r *MetricRepo) Insert(m *model.ServerMetric) error {
	_, err := r.db.Exec(`
		INSERT INTO server_metrics (server_id, cpu_percent, ram_used_mb, ram_total_mb,
			disk_used_gb, uptime_seconds, online, recorded_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		m.ServerID, m.CPUPercent, m.RAMUsedMB, m.RAMTotalMB,
		m.DiskUsedGB, m.UptimeSeconds, m.Online, m.RecordedAt,
	)
	return err
}

func (r *MetricRepo) GetByServerID(serverID int64, limit int) ([]model.ServerMetric, error) {
	var metrics []model.ServerMetric
	return metrics, r.db.Select(&metrics,
		"SELECT * FROM server_metrics WHERE server_id = ? ORDER BY recorded_at DESC LIMIT ?",
		serverID, limit,
	)
}
