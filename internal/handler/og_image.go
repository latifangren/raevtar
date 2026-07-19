package handler

import (
	"net/http"
)

func (h *Handler) serveBlogOGImage(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	svg, err := h.svc.SiteMeta.OGImageBlogSVG(slug)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "image/svg+xml")
	_, _ = w.Write([]byte(svg))
}

func (h *Handler) serveProjectOGImage(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	svg, err := h.svc.SiteMeta.OGImageProjectSVG(slug)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "image/svg+xml")
	_, _ = w.Write([]byte(svg))
}
