package handler

import (
	"fmt"
	"net/http"
	"strings"
)

func (h *Handler) sitemapXML(w http.ResponseWriter, r *http.Request) {
	entries, err := h.svc.SiteMeta.SitemapEntries()
	if err != nil {
		internalServerError(w, r, err)
		return
	}
	var b strings.Builder
	b.WriteString("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n")
	b.WriteString("<urlset xmlns=\"http://www.sitemaps.org/schemas/sitemap/0.9\">\n")
	for _, entry := range entries {
		b.WriteString("  <url>\n")
		b.WriteString("    <loc>" + xmlEscape(entry.URL) + "</loc>\n")
		if !entry.LastMod.IsZero() {
			b.WriteString("    <lastmod>" + entry.LastMod.UTC().Format("2006-01-02T15:04:05Z") + "</lastmod>\n")
		}
		b.WriteString("  </url>\n")
	}
	b.WriteString("</urlset>")
	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	_, _ = w.Write([]byte(b.String()))
}

func (h *Handler) llmsTxt(w http.ResponseWriter, r *http.Request) {
	body, err := h.svc.SiteMeta.LLMSText()
	if err != nil {
		internalServerError(w, r, err)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = fmt.Fprint(w, body)
}

func (h *Handler) robotsTxt(w http.ResponseWriter, r *http.Request) {
	domain := h.cfg.Domain
	if domain == "" {
		domain = "example.com"
	}
	body := fmt.Sprintf("User-agent: *\nAllow: /\nSitemap: https://%s/sitemap.xml\n", domain)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = fmt.Fprint(w, body)
}
