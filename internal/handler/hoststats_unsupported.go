//go:build !linux

package handler

import "raevtar/internal/service"

// collectHostStats is a stub for non-Linux platforms.
func collectHostStats() HostStats {
	return service.CollectHostStats()
}
