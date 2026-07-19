//go:build !linux

package service

// CollectHostStats is a stub for non-Linux platforms.
func CollectHostStats() HostStats {
	return HostStats{}
}
