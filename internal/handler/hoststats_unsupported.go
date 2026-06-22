//go:build !linux

package handler

// collectHostStats is a stub for non-Linux platforms.
func collectHostStats() HostStats {
	return HostStats{}
}
