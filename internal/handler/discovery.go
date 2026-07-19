package handler

import (
	"fmt"
	"net/http"
)

func (h *Handler) sitemapXML(w http.ResponseWriter, r *http.Request) {
	xml, err := h.svc.SiteMeta.SitemapXML()
	if err != nil {
		internalServerError(w, r, err)
		return
	}
	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	_, _ = w.Write([]byte(xml))
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
	body := h.svc.SiteMeta.RobotsTxt()
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = fmt.Fprint(w, body)
}
