package handler

import (
	"encoding/json"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
)

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

// collectHostStats reads /proc + sysfs for live system metrics.
func collectHostStats() HostStats {
	var s HostStats

	// --- CPU Load ---
	if data, err := os.ReadFile("/proc/loadavg"); err == nil {
		parts := strings.Fields(string(data))
		if len(parts) >= 3 {
			s.CPU.Load1, _ = strconv.ParseFloat(parts[0], 64)
			s.CPU.Load5, _ = strconv.ParseFloat(parts[1], 64)
			s.CPU.Load15, _ = strconv.ParseFloat(parts[2], 64)
		}
	}
	s.CPU.Cores = cpuCount()

	// --- RAM (from /proc/meminfo, values in kB) ---
	if data, err := os.ReadFile("/proc/meminfo"); err == nil {
		for _, line := range strings.Split(string(data), "\n") {
			if strings.HasPrefix(line, "MemTotal:") {
				f := strings.Fields(line)
				if len(f) >= 2 {
					s.RAM.Total, _ = strconv.ParseUint(f[1], 10, 64)
				}
			}
			if strings.HasPrefix(line, "MemAvailable:") {
				f := strings.Fields(line)
				if len(f) >= 2 {
					s.RAM.Available, _ = strconv.ParseUint(f[1], 10, 64)
				}
			}
		}
		if s.RAM.Total > 0 {
			s.RAM.Used = s.RAM.Total - s.RAM.Available
			s.RAM.Percent = math.Round(float64(s.RAM.Used) / float64(s.RAM.Total) * 100)
		}
	}

	s.Disk = collectDiskStats()

	// --- Temperature ---
	s.Temp, s.TempAvailable = readTemp()

	return s
}

func cpuCount() int {
	data, err := os.ReadFile("/proc/cpuinfo")
	if err != nil {
		return 0
	}
	n := 0
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, "processor") {
			n++
		}
	}
	if n == 0 {
		return 0
	}
	return n
}

func readTemp() (float64, bool) {
	entries, err := os.ReadDir("/sys/class/thermal")
	if err != nil {
		return 0, false
	}
	for _, e := range entries {
		name := e.Name()
		if !strings.HasPrefix(name, "thermal_zone") {
			continue
		}
		data, err := os.ReadFile("/sys/class/thermal/" + name + "/temp")
		if err != nil {
			continue
		}
		val, err := strconv.ParseFloat(strings.TrimSpace(string(data)), 64)
		if err != nil {
			continue
		}
		// Temperature is in millidegrees Celsius
		return math.Round(val/100) / 10, true
	}
	return 0, false
}

// apiHostStats returns live system stats as JSON (for HTMX polling).
func (h *Handler) apiHostStats(w http.ResponseWriter, r *http.Request) {
	stats := collectHostStats()
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(stats); err != nil {
		logHandlerError(r, err)
	}
}

// formatBytes converts kB to human-readable string (e.g. "7.2 GB").
func formatBytes(kb uint64) string {
	const unit = 1024
	if kb < unit {
		return strconv.FormatUint(kb, 10) + " KB"
	}
	div, exp := uint64(unit), 0
	for n := kb / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	val := float64(kb) / float64(div)
	return strconv.FormatFloat(math.Round(val*10)/10, 'f', 1, 64) + " " + []string{"MB", "GB", "TB"}[exp]
}
