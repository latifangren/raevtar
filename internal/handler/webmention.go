package handler

import (
	"encoding/json"
	"net/http"

	"raevtar/internal/service"
)

func (h *Handler) handleWebmention(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST required", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	source := r.FormValue("source")
	target := r.FormValue("target")

	mention, err := h.svc.Webmention.Receive(service.WebmentionInput{
		Source: source,
		Target: target,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"result": "accepted",
		"id":     mention.ID,
	})
}
