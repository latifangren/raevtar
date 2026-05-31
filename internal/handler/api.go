package handler

import (
	"crypto/subtle"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"raevtar/internal/model"
)

func (h *Handler) apiListPosts(w http.ResponseWriter, r *http.Request) {
	cat := r.URL.Query().Get("category")
	posts, _, _ := h.svc.Blog.ListPosts(cat, 1, 100)

	writeJSON(w, http.StatusOK, posts)
}

func (h *Handler) apiCreatePost(w http.ResponseWriter, r *http.Request) {
	var input model.PostCreate
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
		return
	}

	if input.Title == "" || input.CategorySlug == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "title and category_slug required"})
		return
	}

	post, err := h.svc.Blog.CreatePost(input)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusCreated, post)
}

func (h *Handler) apiListCategories(w http.ResponseWriter, r *http.Request) {
	cats, err := h.svc.Repos.Category.List()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, cats)
}

func (h *Handler) apiCreateServer(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name string `json:"name"`
		Host string `json:"host"`
		Port int    `json:"port"`
		Tags string `json:"tags"`
	}
	// Accept both JSON and form-encoded (HTMX)
	ct := r.Header.Get("Content-Type")
	if len(ct) >= 16 && ct[:16] == "application/json" {
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
			return
		}
	} else {
		input.Name = r.FormValue("name")
		input.Host = r.FormValue("host")
		if p, err := strconv.Atoi(r.FormValue("port")); err == nil {
			input.Port = p
		}
		input.Tags = r.FormValue("tags")
	}
	if input.Name == "" || input.Host == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "name and host required"})
		return
	}
	if input.Port == 0 {
		input.Port = 22
	}

	s, err := h.svc.Monitor.CreateServer(input.Name, input.Host, input.Port, input.Tags)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusCreated, s)
}

func (h *Handler) apiListServers(w http.ResponseWriter, r *http.Request) {
	servers, _ := h.svc.Monitor.ListServers()
	writeJSON(w, http.StatusOK, servers)
}

func (h *Handler) apiGetServer(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("serverID")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}
	server, err := h.svc.Monitor.GetServer(id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
		return
	}
	writeJSON(w, http.StatusOK, server)
}

func (h *Handler) apiRecordMetrics(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("serverID")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}

	var m model.ServerMetric
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
		return
	}

	m.ServerID = id
	m.RecordedAt = time.Now()
	if err := h.svc.Repos.Metric.Insert(&m); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	// update last_seen on server
	_ = h.svc.Repos.Server.UpdateLastSeen(id)

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// adminAuth is middleware that validates Authorization: Bearer <key>
// using constant-time comparison to prevent timing attacks.
func (h *Handler) adminAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if h.cfg.AdminKey == "" {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "admin key not configured"})
			return
		}

		auth := r.Header.Get("Authorization")
		if len(auth) < 8 || auth[:7] != "Bearer " {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "missing or invalid Authorization header"})
			return
		}

		provided := auth[7:]
		if subtle.ConstantTimeCompare([]byte(provided), []byte(h.cfg.AdminKey)) != 1 {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid admin key"})
			return
		}

		next.ServeHTTP(w, r)
	})
}
