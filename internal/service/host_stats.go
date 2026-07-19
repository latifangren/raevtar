package service

// HostStats holds live system metrics for the local host.
type HostStats struct {
	CPU           CPUStats  `json:"cpu"`
	RAM           RAMStats  `json:"ram"`
	Disk          DiskStats `json:"disk"`
	Temp          float64   `json:"temp"`
	TempAvailable bool      `json:"temp_available"`
}

type CPUStats struct {
	Load1  float64 `json:"load1"`
	Load5  float64 `json:"load5"`
	Load15 float64 `json:"load15"`
	Cores  int     `json:"cores"`
}

type RAMStats struct {
	Total     uint64  `json:"total"`
	Available uint64  `json:"available"`
	Used      uint64  `json:"used"`
	Percent   float64 `json:"percent"`
}

type DiskStats struct {
	Total   uint64  `json:"total"`
	Free    uint64  `json:"free"`
	Used    uint64  `json:"used"`
	Percent float64 `json:"percent"`
}

func (s *MonitorService) GetHostSnapshot() HostStats {
	return CollectHostStats()
}
