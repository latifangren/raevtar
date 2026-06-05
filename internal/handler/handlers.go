package handler

import (
	"database/sql"
	"errors"
	"math"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	"raevtar/internal/model"
	"raevtar/internal/view/pages"
)

func (h *Handler) landingIndex(w http.ResponseWriter, r *http.Request) {
	posts, postCount, err := h.svc.Blog.ListPosts("", 1, 3)
	if err != nil {
		internalServerError(w, r, err)
		return
	}
	servers, err := h.svc.Monitor.ListServers()
	if err != nil {
		internalServerError(w, r, err)
		return
	}
	categories, err := h.svc.Blog.ListCategories()
	if err != nil {
		internalServerError(w, r, err)
		return
	}

	renderHTML(w, r, pages.Index(pages.IndexData{
		CurrentPath: r.URL.Path,
		Posts:       posts,
		PostCount:   postCount,
		Servers:     servers,
		Categories:  categories,
		Domain:      h.cfg.Domain,
	}))
}

func (h *Handler) aboutPage(w http.ResponseWriter, r *http.Request) {
	categories, err := h.svc.Blog.ListCategories()
	if err != nil {
		internalServerError(w, r, err)
		return
	}
	_, postCount, err := h.svc.Blog.ListPosts("", 1, 1)
	if err != nil {
		internalServerError(w, r, err)
		return
	}
	servers, err := h.svc.Monitor.ListServers()
	if err != nil {
		internalServerError(w, r, err)
		return
	}
	page, err := h.svc.Pages.GetPage(model.PageKeyAbout)
	if err != nil {
		internalServerError(w, r, err)
		return
	}

	renderHTML(w, r, pages.About(pages.AboutData{
		CurrentPath:   r.URL.Path,
		Categories:    categories,
		PostCount:     postCount,
		CategoryCount: len(categories),
		ServerCount:   len(servers),
		Domain:        h.cfg.Domain,
		Page:          page,
	}))
}

func (h *Handler) labPage(w http.ResponseWriter, r *http.Request) {
	categories, err := h.svc.Blog.ListCategories()
	if err != nil {
		internalServerError(w, r, err)
		return
	}
	_, postCount, err := h.svc.Blog.ListPosts("", 1, 1)
	if err != nil {
		internalServerError(w, r, err)
		return
	}
	servers, err := h.svc.Monitor.ListServers()
	if err != nil {
		internalServerError(w, r, err)
		return
	}

	renderHTML(w, r, pages.Lab(pages.LabData{
		CurrentPath:   r.URL.Path,
		Categories:    categories,
		PostCount:     postCount,
		CategoryCount: len(categories),
		ServerCount:   len(servers),
	}))
}

func (h *Handler) docsPage(w http.ResponseWriter, r *http.Request) {
	categories, err := h.svc.Blog.ListCategories()
	if err != nil {
		internalServerError(w, r, err)
		return
	}
	_, postCount, err := h.svc.Blog.ListPosts("", 1, 1)
	if err != nil {
		internalServerError(w, r, err)
		return
	}

	renderHTML(w, r, pages.Docs(pages.DocsData{
		CurrentPath: r.URL.Path,
		Categories:  categories,
		PostCount:   postCount,
	}))
}

func (h *Handler) projectsPage(w http.ResponseWriter, r *http.Request) {
	categories, err := h.svc.Blog.ListCategories()
	if err != nil {
		internalServerError(w, r, err)
		return
	}
	_, postCount, err := h.svc.Blog.ListPosts("", 1, 1)
	if err != nil {
		internalServerError(w, r, err)
		return
	}
	servers, err := h.svc.Monitor.ListServers()
	if err != nil {
		internalServerError(w, r, err)
		return
	}
	projects, projectCount, err := h.svc.Projects.ListProjects(1, 12)
	if err != nil {
		internalServerError(w, r, err)
		return
	}

	renderHTML(w, r, pages.Projects(pages.ProjectsData{
		CurrentPath:   r.URL.Path,
		Categories:    categories,
		PostCount:     postCount,
		CategoryCount: len(categories),
		ServerCount:   len(servers),
		Projects:      projects,
		ProjectCount:  projectCount,
	}))
}

func (h *Handler) projectDetail(w http.ResponseWriter, r *http.Request) {
	project, err := h.svc.Projects.GetPublishedProject(r.PathValue("slug"))
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			internalServerError(w, r, err)
			return
		}
		http.Error(w, "Project not found", http.StatusNotFound)
		return
	}
	categories, err := h.svc.Blog.ListCategories()
	if err != nil {
		internalServerError(w, r, err)
		return
	}
	renderHTML(w, r, pages.ProjectDetail(pages.ProjectDetailData{
		CurrentPath: r.URL.Path,
		Project:     project,
		Categories:  categories,
	}))
}

func (h *Handler) contactPage(w http.ResponseWriter, r *http.Request) {
	categories, err := h.svc.Blog.ListCategories()
	if err != nil {
		internalServerError(w, r, err)
		return
	}
	page, err := h.svc.Pages.GetPage(model.PageKeyContact)
	if err != nil {
		internalServerError(w, r, err)
		return
	}

	renderHTML(w, r, pages.Contact(pages.ContactData{
		CurrentPath: r.URL.Path,
		Categories:  categories,
		Domain:      h.cfg.Domain,
		Page:        page,
	}))
}

func (h *Handler) topicsPage(w http.ResponseWriter, r *http.Request) {
	categories, err := h.svc.Blog.ListCategories()
	if err != nil {
		internalServerError(w, r, err)
		return
	}
	_, postCount, err := h.svc.Blog.ListPosts("", 1, 1)
	if err != nil {
		internalServerError(w, r, err)
		return
	}

	renderHTML(w, r, pages.Topics(pages.TopicsData{
		CurrentPath: r.URL.Path,
		Categories:  categories,
		PostCount:   postCount,
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
		internalServerError(w, r, err)
		return
	}
	categories, err := h.svc.Blog.ListCategories()
	if err != nil {
		internalServerError(w, r, err)
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
		if !errors.Is(err, sql.ErrNoRows) {
			internalServerError(w, r, err)
			return
		}
		http.Error(w, "Post not found", http.StatusNotFound)
		return
	}
	categories, err := h.svc.Blog.ListCategories()
	if err != nil {
		internalServerError(w, r, err)
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
		internalServerError(w, r, err)
		return
	}
	categories, err := h.svc.Blog.ListCategories()
	if err != nil {
		internalServerError(w, r, err)
		return
	}
	summaries := make([]pages.PublicServerSummary, 0, len(servers))
	for _, server := range servers {
		metrics, err := h.svc.Monitor.GetRecentMetrics(server.ID, 50)
		if err != nil {
			internalServerError(w, r, err)
			return
		}
		summaries = append(summaries, pages.PublicServerSummary{Server: server, Metrics: metrics})
	}
	stats := collectHostStats()

	renderHTML(w, r, pages.Dashboard(pages.DashboardData{
		CurrentPath:     r.URL.Path,
		Servers:         servers,
		ServerSummaries: summaries,
		Categories:      categories,
		PlatformHealth:  publicHostHealth(stats),
		RefreshedAt:     time.Now().UTC(),
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
		if !errors.Is(err, sql.ErrNoRows) {
			internalServerError(w, r, err)
			return pages.ServerDetailData{}, false
		}
		http.Error(w, "Server not found", http.StatusNotFound)
		return pages.ServerDetailData{}, false
	}
	metrics, err := h.svc.Monitor.GetRecentMetrics(id, 50)
	if err != nil {
		internalServerError(w, r, err)
		return pages.ServerDetailData{}, false
	}
	categories, err := h.svc.Blog.ListCategories()
	if err != nil {
		internalServerError(w, r, err)
		return pages.ServerDetailData{}, false
	}

	return pages.ServerDetailData{
		CurrentPath: r.URL.Path,
		Server:      server,
		Metrics:     metrics,
		Categories:  categories,
		RefreshedAt: time.Now().UTC(),
	}, true
}

func itoa(n int) string { return strconv.Itoa(n) }

func itoa64(n int64) string { return strconv.FormatInt(n, 10) }

func publicHostHealth(stats HostStats) pages.PublicHostHealthData {
	return pages.PublicHostHealthData{
		CPULoad:     publicCPULoadText(stats.CPU.Load1, stats.CPU.Load5, stats.CPU.Load15, int64(stats.CPU.Cores)),
		CPUCores:    publicCoresText(int64(stats.CPU.Cores)),
		RAMUsage:    publicCapacityText(kbToGB(stats.RAM.Used), kbToGB(stats.RAM.Total)),
		RAMPercent:  publicPercentText(stats.RAM.Percent, stats.RAM.Total > 0),
		DiskUsage:   publicCapacityText(kbToGB(stats.Disk.Used), kbToGB(stats.Disk.Total)),
		DiskPercent: publicPercentText(stats.Disk.Percent, stats.Disk.Total > 0),
		Temperature: publicTemperatureText(stats.Temp, stats.TempAvailable),
	}
}

func publicCPULoadText(load1, load5, load15 float64, cores int64) string {
	if cores <= 0 {
		return "N/A"
	}
	return strconv.FormatFloat(load1, 'f', 2, 64) + " / " + strconv.FormatFloat(load5, 'f', 2, 64) + " / " + strconv.FormatFloat(load15, 'f', 2, 64)
}

func publicCoresText(cores int64) string {
	if cores <= 0 {
		return "N/A"
	}
	return strconv.FormatInt(cores, 10)
}

func publicCapacityText(used, total float64) string {
	if total <= 0 {
		return "N/A"
	}
	return strconv.FormatFloat(round1(used), 'f', 1, 64) + " / " + strconv.FormatFloat(round1(total), 'f', 1, 64) + " GB"
}

func publicPercentText(percent float64, available bool) string {
	if !available {
		return "N/A"
	}
	return strconv.FormatFloat(math.Round(percent), 'f', 0, 64) + "%"
}

func publicTemperatureText(temp float64, available bool) string {
	if !available {
		return "N/A"
	}
	return strconv.FormatFloat(temp, 'f', 1, 64) + "°C"
}

func kbToGB(kb uint64) float64 {
	return float64(kb) / 1024 / 1024
}

func round1(value float64) float64 {
	return math.Round(value*10) / 10
}

func (h *Handler) serveStatic(filename string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/"+filename)
	}
}

func (h *Handler) serveUpload(w http.ResponseWriter, r *http.Request) {
	requested := r.PathValue("filename")
	name := filepath.Base(requested)
	if name == "." || name == "" || name != requested {
		http.NotFound(w, r)
		return
	}
	path, err := h.svc.Media.FilePath(name)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	http.ServeFile(w, r, path)
}

func (h *Handler) page404(w http.ResponseWriter, r *http.Request) {
	categories, err := h.svc.Blog.ListCategories()
	if err != nil {
		internalServerError(w, r, err)
		return
	}
	renderHTML(w, r, pages.NotFound(pages.NotFoundData{CurrentPath: r.URL.Path, Categories: categories}))
}
