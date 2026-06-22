//go:build !windows

package handler

import (
	"math"
	"os"
	"syscall"
)

// DiskRootPath is the filesystem path to check for disk stats.
// Override via RAEVTAR_DISK_ROOT env var.
var DiskRootPath string

func init() {
	DiskRootPath = os.Getenv("RAEVTAR_DISK_ROOT")
	if DiskRootPath == "" {
		DiskRootPath = "/"
	}
}

func collectDiskStats() DiskStats {
	var s DiskStats
	var stat syscall.Statfs_t
	if err := syscall.Statfs(DiskRootPath, &stat); err != nil {
		return s
	}
	totalKB := stat.Blocks * uint64(stat.Bsize) / 1024
	freeKB := stat.Bfree * uint64(stat.Bsize) / 1024
	s.Total = totalKB
	s.Free = freeKB
	s.Used = totalKB - freeKB
	if totalKB > 0 {
		s.Percent = math.Round(float64(s.Used) / float64(totalKB) * 100)
	}
	return s
}
