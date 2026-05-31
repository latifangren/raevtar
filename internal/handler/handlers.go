package handler

import (
	"net/http"
	"strconv"
	"time"

	"raevtar/internal/view/pages"
)

func (h *Handler) landingIndex(w http.ResponseWriter, r *http.Request) {
	posts, _, err := h.svc.Blog.ListPosts("", 1, 3)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	servers, err := h.svc.Monitor.ListServers()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	categories, err := h.svc.Blog.ListCategories()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	renderHTML(w, r, pages.Index(pages.IndexData{
		CurrentPath: r.URL.Path,
		Posts:       posts,
		Servers:     servers,
		Categories:  categories,
		Domain:      h.cfg.Domain,
	}))
}

func (h *Handler) blogList(w http.ResponseWriter, r *http.Request) {
	cat := r.URL.Query().Get("category")
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	posts, total, err := h.svc.Blog.ListPosts(cat, page, 10)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	categories, err := h.svc.Blog.ListCategories()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	renderHTML(w, r, pages.BlogList(pages.BlogListData{
		CurrentPath: r.URL.Path,
		Posts:       posts,
		Categories:  categories,
		CurrentCat:  cat,
		Page:        page,
		TotalPages:  (total + 9) / 10,
	}))
}

func (h *Handler) blogDetail(w http.ResponseWriter, r *http.Request) {
	post, err := h.svc.Blog.GetPublishedPost(r.PathValue("slug"))
	if err != nil {
		http.Error(w, "Post not found", http.StatusNotFound)
		return
	}
	categories, err := h.svc.Blog.ListCategories()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	renderHTML(w, r, pages.BlogPost(pages.BlogPostData{
		CurrentPath: r.URL.Path,
		Post:        post,
		Categories:  categories,
	}))
}

func (h *Handler) dashboardIndex(w http.ResponseWriter, r *http.Request) {
	servers, err := h.svc.Monitor.ListServers()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	categories, err := h.svc.Blog.ListCategories()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	canManage := canManageServer(r)
	renderHTML(w, r, pages.Dashboard(pages.DashboardData{
		CurrentPath:       r.URL.Path,
		Servers:           servers,
		Categories:        categories,
		CanRegisterServer: canManage,
		CanViewServerInfo: canManage,
		CSRFToken:         csrfTokenForRequest(r),
	}))
}

func (h *Handler) dashboardDetail(w http.ResponseWriter, r *http.Request) {
	data, ok := h.loadServerDetailData(w, r)
	if !ok {
		return
	}

	renderHTML(w, r, pages.ServerDetail(data))
}

func (h *Handler) dashboardDetailLive(w http.ResponseWriter, r *http.Request) {
	data, ok := h.loadServerDetailData(w, r)
	if !ok {
		return
	}

	renderHTML(w, r, pages.ServerDetailLive(data))
}

func (h *Handler) loadServerDetailData(w http.ResponseWriter, r *http.Request) (pages.ServerDetailData, bool) {
	id, err := strconv.ParseInt(r.PathValue("serverID"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid server ID", http.StatusBadRequest)
		return pages.ServerDetailData{}, false
	}
	server, err := h.svc.Monitor.GetServer(id)
	if err != nil {
		http.Error(w, "Server not found", http.StatusNotFound)
		return pages.ServerDetailData{}, false
	}
	metrics, err := h.svc.Monitor.GetRecentMetrics(id, 50)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return pages.ServerDetailData{}, false
	}
	categories, err := h.svc.Blog.ListCategories()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return pages.ServerDetailData{}, false
	}

	canManage := canManageServer(r)
	return pages.ServerDetailData{
		CurrentPath:       r.URL.Path,
		Server:            server,
		Metrics:           metrics,
		Categories:        categories,
		CanManageServer:   canManage,
		CanViewServerInfo: canManage,
		RefreshedAt:       time.Now().UTC(),
	}, true
}

func itoa(n int) string { return strconv.Itoa(n) }

func itoa64(n int64) string { return strconv.FormatInt(n, 10) }

func (h *Handler) serveStatic(filename string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/"+filename)
	}
}

func (h *Handler) page404(w http.ResponseWriter, r *http.Request) {
	categories, _ := h.svc.Blog.ListCategories()
	renderHTML(w, r, pages.NotFound(pages.NotFoundData{CurrentPath: r.URL.Path, Categories: categories}))
}
