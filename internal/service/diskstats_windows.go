//go:build windows

package service

func collectDiskStats() DiskStats {
	return DiskStats{}
}
