package handler

import (
	"fmt"
	"html"
	"net/http"
)

func (h *Handler) serveBlogOGImage(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	post, err := h.svc.Blog.GetPublishedPost(slug)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	h.serveOGImage(w, post.Title, "Blog dispatch from Raevtar")
}

func (h *Handler) serveProjectOGImage(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	project, err := h.svc.Projects.GetPublishedProject(slug)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	h.serveOGImage(w, project.Title, "Project from small-machine lab")
}

func (h *Handler) serveOGImage(w http.ResponseWriter, title, subtitle string) {
	w.Header().Set("Content-Type", "image/svg+xml")
	domain := h.cfg.Domain
	if domain == "" {
		domain = "raevtar.tech"
	}
	// Bold Neo-Brutalist SVG OG Image
	fmt.Fprintf(w, `<svg width="1200" height="630" viewBox="0 0 1200 630" xmlns="http://www.w3.org/2000/svg">
		<rect width="1200" height="630" fill="#F5F2ED"/>
		<!-- Layered borders -->
		<rect x="20" y="20" width="1160" height="590" fill="none" stroke="#2D3748" stroke-width="8"/>
		<rect x="40" y="40" width="1160" height="590" fill="none" stroke="#2D3748" stroke-width="8"/>
		<!-- Background highlight -->
		<rect x="30" y="30" width="1140" height="570" fill="#FACC15" stroke="#2D3748" stroke-width="4"/>
		
		<!-- Subtitle -->
		<text x="60" y="100" font-family="sans-serif" font-size="24" font-weight="900" fill="#2D3748" text-transform="uppercase">%s</text>
		
		<!-- Main Title (wrapped manually for simplicity in SVG) -->
		<text x="60" y="300" font-family="sans-serif" font-size="72" font-weight="900" fill="#2D3748">%s</text>
		
		<!-- Branding -->
		<rect x="60" y="500" width="200" height="60" fill="#2D3748"/>
		<text x="80" y="540" font-family="sans-serif" font-size="32" font-weight="bold" fill="#F5F2ED">`+domain+`</text>
	</svg>`, html.EscapeString(subtitle), html.EscapeString(title))
}
