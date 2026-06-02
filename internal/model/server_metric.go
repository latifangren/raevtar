package model

import "time"

type ServerMetric struct {
	ID                   int64     `db:"id" json:"id"`
	ServerID             int64     `db:"server_id" json:"server_id"`
	CPUPercent           float64   `db:"cpu_percent" json:"cpu_percent"`
	CPULoad1             float64   `db:"cpu_load_1" json:"cpu_load_1"`
	CPULoad5             float64   `db:"cpu_load_5" json:"cpu_load_5"`
	CPULoad15            float64   `db:"cpu_load_15" json:"cpu_load_15"`
	CPUCores             int64     `db:"cpu_cores" json:"cpu_cores"`
	RAMUsedMB            float64   `db:"ram_used_mb" json:"ram_used_mb"`
	RAMTotalMB           float64   `db:"ram_total_mb" json:"ram_total_mb"`
	DiskUsedGB           float64   `db:"disk_used_gb" json:"disk_used_gb"`
	DiskTotalGB          float64   `db:"disk_total_gb" json:"disk_total_gb"`
	TemperatureC         float64   `db:"temperature_c" json:"temperature_c"`
	TemperatureAvailable bool      `db:"temperature_available" json:"temperature_available"`
	UptimeSeconds        int64     `db:"uptime_seconds" json:"uptime_seconds"`
	Online               bool      `db:"online" json:"online"`
	RecordedAt           time.Time `db:"recorded_at" json:"recorded_at"`
}
