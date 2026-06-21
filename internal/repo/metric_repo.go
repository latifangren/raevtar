package repo

import (
	"raevtar/internal/model"

	"github.com/jmoiron/sqlx"
)

type MetricRepo struct{ db *sqlx.DB }

func (r *MetricRepo) Insert(m *model.ServerMetric) error {
	result, err := r.db.Exec(`
			INSERT INTO server_metrics (server_id, cpu_percent, cpu_load_1, cpu_load_5,
				cpu_load_15, cpu_cores, ram_used_mb, ram_total_mb, disk_used_gb,
				disk_total_gb, temperature_c, temperature_available, uptime_seconds,
				online, top_processes, logs, recorded_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		m.ServerID, m.CPUPercent, m.CPULoad1, m.CPULoad5,
		m.CPULoad15, m.CPUCores, m.RAMUsedMB, m.RAMTotalMB,
		m.DiskUsedGB, m.DiskTotalGB, m.TemperatureC,
		m.TemperatureAvailable, m.UptimeSeconds, m.Online, m.TopProcesses, m.Logs, m.RecordedAt,
	)
	if err != nil {
		return err
	}
	id, _ := result.LastInsertId()
	m.ID = id
	return nil
}

func (r *MetricRepo) GetByServerID(serverID int64, limit int) ([]model.ServerMetric, error) {
	var metrics []model.ServerMetric
	return metrics, r.db.Select(&metrics,
		"SELECT * FROM server_metrics WHERE server_id = ? ORDER BY recorded_at DESC LIMIT ?",
		serverID, limit,
	)
}
