//go:build linux

package handler

import "raevtar/internal/service"

// collectHostStats reads /proc + sysfs for live system metrics (Linux only).
func collectHostStats() HostStats {
	return service.CollectHostStats()
}
