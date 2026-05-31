package model

import "time"

type ServerMetric struct {
	ID            int64     `db:"id" json:"id"`
	ServerID      int64     `db:"server_id" json:"server_id"`
	CPUPercent    float64   `db:"cpu_percent" json:"cpu_percent"`
	RAMUsedMB     float64   `db:"ram_used_mb" json:"ram_used_mb"`
	RAMTotalMB    float64   `db:"ram_total_mb" json:"ram_total_mb"`
	DiskUsedGB    float64   `db:"disk_used_gb" json:"disk_used_gb"`
	UptimeSeconds int64     `db:"uptime_seconds" json:"uptime_seconds"`
	Online        bool      `db:"online" json:"online"`
	RecordedAt    time.Time `db:"recorded_at" json:"recorded_at"`
}
