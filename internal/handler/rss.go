package handler

import (
	"net/http"

	"raevtar/internal/service"
)

// rssFeed generates an RSS 2.0 feed of recent blog posts.
func (h *Handler) rssFeed(w http.ResponseWriter, r *http.Request) {
	feed, err := h.svc.SiteMeta.RSSFeed()
	if err != nil {
		internalServerError(w, r, err)
		return
	}
	w.Header().Set("Content-Type", "application/rss+xml; charset=utf-8")
	_, _ = w.Write([]byte(feed))
}

func xmlEscape(s string) string {
	return service.XMLEscape(s)
}
