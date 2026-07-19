//go:build !windows

package handler

import "raevtar/internal/service"

// DiskRootPath is the filesystem path to check for disk stats.
var DiskRootPath = service.DiskRootPath
