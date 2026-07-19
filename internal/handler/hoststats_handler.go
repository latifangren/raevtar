package handler

import (
	"encoding/json"
	"math"
	"net/http"
	"strconv"
)

func (h *Handler) apiHostStats(w http.ResponseWriter, r *http.Request) {
	stats := h.svc.Monitor.GetHostSnapshot()
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
