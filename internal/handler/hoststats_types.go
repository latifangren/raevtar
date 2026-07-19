package handler

import "raevtar/internal/service"

// HostStats holds live system metrics for the local host.
type HostStats = service.HostStats
type CPUStats = service.CPUStats
type RAMStats = service.RAMStats
type DiskStats = service.DiskStats
