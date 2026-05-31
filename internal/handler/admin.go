package handler

import (
	"math"
	"net/http"
	"strconv"
	"time"

	"raevtar/internal/model"
	adminview "raevtar/internal/view/admin"
)

func (h *Handler) adminIndex(w http.ResponseWriter, r *http.Request) {
	allPosts, _, _ := h.svc.Blog.ListPosts("", 1, 9999)
	servers, _ := h.svc.Monitor.ListServers()
	users, _ := h.svc.Admin.ListUsers()
	stats := collectHostStats()

	onlineCount := 0
	for _, server := range servers {
		if server.LastSeen != nil && time.Since(*server.LastSeen) < 10*time.Minute {
			onlineCount++
		}
	}

	renderHTML(w, r, adminview.Dashboard(adminview.DashboardData{
		CurrentPath: r.URL.Path,
		PostCount:   len(allPosts),
		ServerCount: len(servers),
		UserCount:   len(users),
		OnlineCount: onlineCount,
		Servers:     servers,
		Stats:       adminHostStats(stats),
	}))
}

func (h *Handler) adminUsers(w http.ResponseWriter, r *http.Request) {
	users, _ := h.svc.Admin.ListUsers()
	entry, _ := getSessionEntry(r)

	rows := make([]adminview.UserRow, 0, len(users))
	for _, user := range users {
		rows = append(rows, adminview.UserRow{
			User:      user,
			CanDelete: model.CanManage(entry.role, user.Role) && user.Username != entry.username,
		})
	}

	roleOptions := make([]adminview.RoleOption, 0, len(model.ValidRoles()))
	for _, role := range model.ValidRoles() {
		if model.CanManage(entry.role, role) {
			roleOptions = append(roleOptions, adminview.RoleOption{
				Value:    role,
				Selected: role == model.RoleOperator,
			})
		}
	}

	renderHTML(w, r, adminview.Users(adminview.UsersData{
		CurrentPath: r.URL.Path,
		Users:       rows,
		RoleOptions: roleOptions,
	}))
}

func (h *Handler) adminCreateUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST required", http.StatusMethodNotAllowed)
		return
	}

	entry, _ := getSessionEntry(r)
	username := r.FormValue("username")
	password := r.FormValue("password")
	role := r.FormValue("role")

	if username == "" || password == "" {
		http.Error(w, "username and password required", http.StatusBadRequest)
		return
	}

	if len(password) < 8 {
		http.Error(w, "password must be at least 8 characters", http.StatusBadRequest)
		return
	}

	if model.IsValidRole(role) && !model.CanManage(entry.role, role) {
		http.Error(w, "you cannot create users with this role", http.StatusForbidden)
		return
	}

	_, err := h.svc.Admin.CreateUser(entry.role, entry.username, username, password, role, r.RemoteAddr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin/manage-users", http.StatusSeeOther)
}

func (h *Handler) adminDeleteUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST required", http.StatusMethodNotAllowed)
		return
	}

	entry, _ := getSessionEntry(r)
	idStr := r.PathValue("userID")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	if err := h.svc.Admin.DeleteUser(entry.role, entry.username, id, r.RemoteAddr); err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	http.Redirect(w, r, "/admin/manage-users", http.StatusSeeOther)
}

func (h *Handler) adminAuditLog(w http.ResponseWriter, r *http.Request) {
	logs, _ := h.svc.Admin.ListAuditLogs(200, 0)
	renderHTML(w, r, adminview.Audit(adminview.AuditData{CurrentPath: r.URL.Path, Logs: logs}))
}

func (h *Handler) adminPosts(w http.ResponseWriter, r *http.Request) {
	posts, _, _ := h.svc.Blog.ListPosts("", 1, 9999)
	categories, _ := h.svc.Blog.ListCategories()
	renderHTML(w, r, adminview.Posts(adminview.PostsData{
		CurrentPath: r.URL.Path,
		Posts:       posts,
		Categories:  categories,
	}))
}

func (h *Handler) adminCreatePost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST required", http.StatusMethodNotAllowed)
		return
	}

	entry, _ := getSessionEntry(r)
	title := r.FormValue("title")
	catSlug := r.FormValue("category_slug")
	content := r.FormValue("content")

	if title == "" || catSlug == "" || content == "" {
		http.Error(w, "title, category_slug, and content required", http.StatusBadRequest)
		return
	}

	post, err := h.svc.Blog.CreatePost(model.PostCreate{
		Title:        title,
		ContentMD:    content,
		CategorySlug: catSlug,
		Published:    true,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_ = h.svc.Admin.LogPostCreated(entry.username, post.Title, r.RemoteAddr)
	http.Redirect(w, r, "/admin/posts", http.StatusSeeOther)
}

func (h *Handler) adminDeletePost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST required", http.StatusMethodNotAllowed)
		return
	}

	entry, _ := getSessionEntry(r)
	idStr := r.PathValue("postID")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	_ = h.svc.Admin.DeletePost(entry.username, id, r.RemoteAddr)
	http.Redirect(w, r, "/admin/posts", http.StatusSeeOther)
}

func (h *Handler) adminServers(w http.ResponseWriter, r *http.Request) {
	servers, _ := h.svc.Monitor.ListServers()
	renderHTML(w, r, adminview.Servers(adminview.ServersData{CurrentPath: r.URL.Path, Servers: servers}))
}

func (h *Handler) adminCreateServer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST required", http.StatusMethodNotAllowed)
		return
	}

	entry, _ := getSessionEntry(r)
	name := r.FormValue("name")
	host := r.FormValue("host")
	portStr := r.FormValue("port")
	tags := r.FormValue("tags")

	port, err := strconv.Atoi(portStr)
	if err != nil {
		http.Error(w, "invalid port", http.StatusBadRequest)
		return
	}

	_, err = h.svc.Monitor.CreateServer(name, host, port, tags)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_ = h.svc.Admin.LogServerCreated(entry.username, name, host, portStr, r.RemoteAddr)
	http.Redirect(w, r, "/admin/servers", http.StatusSeeOther)
}

func (h *Handler) adminDeleteServer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST required", http.StatusMethodNotAllowed)
		return
	}

	entry, _ := getSessionEntry(r)
	idStr := r.PathValue("serverID")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	_ = h.svc.Admin.DeleteServer(entry.username, id, idStr, r.RemoteAddr)
	http.Redirect(w, r, "/admin/servers", http.StatusSeeOther)
}

func adminHostStats(stats HostStats) adminview.HostStatsData {
	loadPercent := 0.0
	if stats.CPU.Cores > 0 {
		loadPercent = (stats.CPU.Load1 / float64(stats.CPU.Cores)) * 100
	}

	return adminview.HostStatsData{
		CPULoad1:       strconv.FormatFloat(stats.CPU.Load1, 'f', 2, 64),
		CPULoad5:       strconv.FormatFloat(stats.CPU.Load5, 'f', 2, 64),
		CPUCores:       stats.CPU.Cores,
		RAMUsed:        formatBytes(stats.RAM.Used),
		RAMTotal:       formatBytes(stats.RAM.Total),
		RAMPercent:     stats.RAM.Percent,
		DiskUsed:       formatBytes(stats.Disk.Used),
		DiskTotal:      formatBytes(stats.Disk.Total),
		DiskPercent:    stats.Disk.Percent,
		Temperature:    strconv.FormatFloat(stats.Temp, 'f', 1, 64),
		TempValue:      stats.Temp,
		CPULoadPercent: math.Round(loadPercent),
	}
}
