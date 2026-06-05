package handler

import (
	"database/sql"
	"errors"
	"math"
	"net/http"
	"strconv"
	"strings"

	"raevtar/internal/model"
	"raevtar/internal/service"
	adminview "raevtar/internal/view/admin"
)

const (
	adminPostIntentDraft   = "draft"
	adminPostIntentPublish = "publish"
	adminPostIntentUpdate  = "update"
)

func (h *Handler) adminIndex(w http.ResponseWriter, r *http.Request) {
	allPosts, _, err := h.svc.Blog.ListAllPosts(1, 9999)
	if err != nil {
		internalServerError(w, r, err)
		return
	}
	servers, err := h.svc.Monitor.ListServers()
	if err != nil {
		internalServerError(w, r, err)
		return
	}
	users, err := h.svc.Admin.ListUsers()
	if err != nil {
		internalServerError(w, r, err)
		return
	}
	stats := collectHostStats()

	onlineCount := 0
	for _, server := range servers {
		if adminview.IsOnline(server.LastSeen) {
			onlineCount++
		}
	}

	renderHTML(w, r, adminview.Dashboard(adminview.DashboardData{
		CurrentPath: r.URL.Path,
		CSRFToken:   csrfTokenForRequest(r),
		PostCount:   len(allPosts),
		ServerCount: len(servers),
		UserCount:   len(users),
		OnlineCount: onlineCount,
		Servers:     servers,
		Stats:       adminHostStats(stats),
	}))
}

func (h *Handler) adminUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.svc.Admin.ListUsers()
	if err != nil {
		internalServerError(w, r, err)
		return
	}
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
		CSRFToken:   csrfTokenForRequest(r),
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

	_, err := h.svc.Admin.CreateUser(entry.role, entry.username, username, password, role, clientIP(r))
	if err != nil {
		internalServerError(w, r, err)
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

	if err := h.svc.Admin.DeleteUser(entry.role, entry.username, id, clientIP(r)); err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	http.Redirect(w, r, "/admin/manage-users", http.StatusSeeOther)
}

func (h *Handler) adminAuditLog(w http.ResponseWriter, r *http.Request) {
	logs, err := h.svc.Admin.ListAuditLogs(200, 0)
	if err != nil {
		internalServerError(w, r, err)
		return
	}
	renderHTML(w, r, adminview.Audit(adminview.AuditData{CurrentPath: r.URL.Path, CSRFToken: csrfTokenForRequest(r), Logs: logs}))
}

func (h *Handler) adminPosts(w http.ResponseWriter, r *http.Request) {
	posts, _, err := h.svc.Blog.ListAllPosts(1, 9999)
	if err != nil {
		internalServerError(w, r, err)
		return
	}
	categories, err := h.svc.Blog.ListCategories()
	if err != nil {
		internalServerError(w, r, err)
		return
	}
	mediaAssets, err := h.svc.Media.ListAssets()
	if err != nil {
		internalServerError(w, r, err)
		return
	}
	renderHTML(w, r, adminview.Posts(adminview.PostsData{
		CurrentPath: r.URL.Path,
		CSRFToken:   csrfTokenForRequest(r),
		Posts:       posts,
		Categories:  categories,
		MediaAssets: mediaAssets,
	}))
}

func (h *Handler) adminMedia(w http.ResponseWriter, r *http.Request) {
	assets, err := h.svc.Media.ListAssets()
	if err != nil {
		internalServerError(w, r, err)
		return
	}
	renderHTML(w, r, adminview.Media(adminview.MediaData{CurrentPath: r.URL.Path, CSRFToken: csrfTokenForRequest(r), Assets: assets}))
}

func (h *Handler) adminUploadMedia(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST required", http.StatusMethodNotAllowed)
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "file required", http.StatusBadRequest)
		return
	}
	defer file.Close()
	_, err = h.svc.Media.Upload(service.MediaUpload{OriginalName: header.Filename, Reader: file})
	if err != nil {
		if errors.Is(err, service.ErrInvalidMediaInput) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		internalServerError(w, r, err)
		return
	}
	http.Redirect(w, r, "/admin/media", http.StatusSeeOther)
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
	excerpt := r.FormValue("excerpt")
	coverImageURL := r.FormValue("cover_image_url")
	tags := splitTags(r.FormValue("tags"))
	intent := r.FormValue("intent")

	if title == "" || catSlug == "" || content == "" {
		http.Error(w, "title, category_slug, and content required", http.StatusBadRequest)
		return
	}

	post, err := h.svc.Blog.CreatePost(model.PostCreate{
		Title:         title,
		ContentMD:     content,
		Excerpt:       excerpt,
		CoverImageURL: coverImageURL,
		CategorySlug:  catSlug,
		Published:     intent == adminPostIntentPublish,
		Tags:          tags,
	})
	if err != nil {
		internalServerError(w, r, err)
		return
	}

	warnAfterMutation(r, "create_post_audit", h.svc.Admin.LogPostCreated(entry.username, post.Title, clientIP(r)))
	http.Redirect(w, r, "/admin/posts", http.StatusSeeOther)
}

func (h *Handler) adminPreviewPost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST required", http.StatusMethodNotAllowed)
		return
	}

	content := r.FormValue("content")
	if content == "" {
		content = r.FormValue("content_md")
	}
	if strings.TrimSpace(content) == "" {
		renderHTML(w, r, adminview.PostMarkdownPreview("", true))
		return
	}

	contentHTML, err := h.svc.Blog.RenderMarkdown(content)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	renderHTML(w, r, adminview.PostMarkdownPreview(contentHTML, false))
}

func (h *Handler) adminEditPost(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("postID"), 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	post, err := h.svc.Blog.GetPostByID(id)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			internalServerError(w, r, err)
			return
		}
		http.Error(w, "post not found", http.StatusNotFound)
		return
	}
	categories, err := h.svc.Blog.ListCategories()
	if err != nil {
		internalServerError(w, r, err)
		return
	}
	mediaAssets, err := h.svc.Media.ListAssets()
	if err != nil {
		internalServerError(w, r, err)
		return
	}
	renderHTML(w, r, adminview.PostEdit(adminview.PostEditData{
		CurrentPath: r.URL.Path,
		CSRFToken:   csrfTokenForRequest(r),
		Post:        post,
		Categories:  categories,
		MediaAssets: mediaAssets,
	}))
}

func (h *Handler) adminUpdatePost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST required", http.StatusMethodNotAllowed)
		return
	}
	entry, _ := getSessionEntry(r)
	id, err := strconv.ParseInt(r.PathValue("postID"), 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	currentPost, err := h.svc.Blog.GetPostByID(id)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			internalServerError(w, r, err)
			return
		}
		http.Error(w, "post not found", http.StatusNotFound)
		return
	}
	published := currentPost.Published
	switch r.FormValue("intent") {
	case adminPostIntentDraft:
		published = false
	case adminPostIntentPublish:
		published = true
	case adminPostIntentUpdate:
	}
	post, err := h.svc.Blog.UpdatePost(id, model.PostUpdate{
		Title:         r.FormValue("title"),
		ContentMD:     r.FormValue("content"),
		Excerpt:       r.FormValue("excerpt"),
		CoverImageURL: r.FormValue("cover_image_url"),
		CategorySlug:  r.FormValue("category_slug"),
		Published:     published,
		Tags:          splitTags(r.FormValue("tags")),
	})
	if err != nil {
		if errors.Is(err, service.ErrPostNotFound) {
			http.Error(w, "post not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	warnAfterMutation(r, "update_post_audit", h.svc.Admin.LogPostUpdated(entry.username, post.Title, clientIP(r)))
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

	if err := h.svc.Admin.DeletePost(entry.username, id, clientIP(r)); err != nil {
		if errors.Is(err, service.ErrPostNotFound) {
			http.Error(w, "post not found", http.StatusNotFound)
			return
		}
		internalServerError(w, r, err)
		return
	}
	http.Redirect(w, r, "/admin/posts", http.StatusSeeOther)
}

func (h *Handler) adminProjects(w http.ResponseWriter, r *http.Request) {
	projects, _, err := h.svc.Projects.ListAllProjects(1, 9999, service.ProjectListOptions{})
	if err != nil {
		internalServerError(w, r, err)
		return
	}
	mediaAssets, err := h.svc.Media.ListAssets()
	if err != nil {
		internalServerError(w, r, err)
		return
	}
	renderHTML(w, r, adminview.Projects(adminview.ProjectsData{
		CurrentPath: r.URL.Path,
		CSRFToken:   csrfTokenForRequest(r),
		Projects:    projects,
		MediaAssets: mediaAssets,
	}))
}

func (h *Handler) adminCreateProject(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST required", http.StatusMethodNotAllowed)
		return
	}
	entry, _ := getSessionEntry(r)
	intent := r.FormValue("intent")
	project, err := h.svc.Projects.CreateProject(model.ProjectCreate{
		Title:         r.FormValue("title"),
		ContentMD:     r.FormValue("content"),
		Excerpt:       r.FormValue("excerpt"),
		CoverImageURL: r.FormValue("cover_image_url"),
		Published:     intent == adminPostIntentPublish,
		Featured:      r.FormValue("featured") == "on",
		SortOrder:     parseSortOrder(r.FormValue("sort_order")),
		Tags:          splitTags(r.FormValue("tags")),
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	warnAfterMutation(r, "create_project_audit", h.svc.Admin.LogProjectCreated(entry.username, project.Title, clientIP(r)))
	http.Redirect(w, r, "/admin/projects", http.StatusSeeOther)
}

func (h *Handler) adminPreviewProject(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST required", http.StatusMethodNotAllowed)
		return
	}
	content := r.FormValue("content")
	if strings.TrimSpace(content) == "" {
		renderHTML(w, r, adminview.PostMarkdownPreview("", true))
		return
	}
	contentHTML, err := h.svc.Projects.RenderMarkdown(content)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	renderHTML(w, r, adminview.PostMarkdownPreview(contentHTML, false))
}

func (h *Handler) adminEditProject(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("projectID"), 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	project, err := h.svc.Projects.GetProjectByID(id)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			internalServerError(w, r, err)
			return
		}
		http.Error(w, "project not found", http.StatusNotFound)
		return
	}
	mediaAssets, err := h.svc.Media.ListAssets()
	if err != nil {
		internalServerError(w, r, err)
		return
	}
	renderHTML(w, r, adminview.ProjectEdit(adminview.ProjectEditData{
		CurrentPath: r.URL.Path,
		CSRFToken:   csrfTokenForRequest(r),
		Project:     project,
		MediaAssets: mediaAssets,
	}))
}

func (h *Handler) adminUpdateProject(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST required", http.StatusMethodNotAllowed)
		return
	}
	entry, _ := getSessionEntry(r)
	id, err := strconv.ParseInt(r.PathValue("projectID"), 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	currentProject, err := h.svc.Projects.GetProjectByID(id)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			internalServerError(w, r, err)
			return
		}
		http.Error(w, "project not found", http.StatusNotFound)
		return
	}
	published := currentProject.Published
	switch r.FormValue("intent") {
	case adminPostIntentDraft:
		published = false
	case adminPostIntentPublish:
		published = true
	case adminPostIntentUpdate:
	}
	project, err := h.svc.Projects.UpdateProject(id, model.ProjectUpdate{
		Title:         r.FormValue("title"),
		ContentMD:     r.FormValue("content"),
		Excerpt:       r.FormValue("excerpt"),
		CoverImageURL: r.FormValue("cover_image_url"),
		Published:     published,
		Featured:      r.FormValue("featured") == "on",
		SortOrder:     parseSortOrder(r.FormValue("sort_order")),
		Tags:          splitTags(r.FormValue("tags")),
	})
	if err != nil {
		if errors.Is(err, service.ErrProjectNotFound) {
			http.Error(w, "project not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	warnAfterMutation(r, "update_project_audit", h.svc.Admin.LogProjectUpdated(entry.username, project.Title, clientIP(r)))
	http.Redirect(w, r, "/admin/projects", http.StatusSeeOther)
}

func (h *Handler) adminDeleteProject(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST required", http.StatusMethodNotAllowed)
		return
	}
	entry, _ := getSessionEntry(r)
	id, err := strconv.ParseInt(r.PathValue("projectID"), 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	if err := h.svc.Admin.DeleteProject(entry.username, id, clientIP(r)); err != nil {
		if errors.Is(err, service.ErrProjectNotFound) {
			http.Error(w, "project not found", http.StatusNotFound)
			return
		}
		internalServerError(w, r, err)
		return
	}
	http.Redirect(w, r, "/admin/projects", http.StatusSeeOther)
}

func (h *Handler) adminPages(w http.ResponseWriter, r *http.Request) {
	pagesData, err := h.svc.Pages.ListPages()
	if err != nil {
		internalServerError(w, r, err)
		return
	}
	renderHTML(w, r, adminview.Pages(adminview.PagesData{
		CurrentPath: r.URL.Path,
		CSRFToken:   csrfTokenForRequest(r),
		Pages:       pagesData,
	}))
}

func (h *Handler) adminEditPage(w http.ResponseWriter, r *http.Request) {
	page, err := h.svc.Pages.GetPage(r.PathValue("pageKey"))
	if err != nil {
		internalServerError(w, r, err)
		return
	}
	renderHTML(w, r, adminview.PageEdit(adminview.PageEditData{
		CurrentPath: r.URL.Path,
		CSRFToken:   csrfTokenForRequest(r),
		Page:        page,
	}))
}

func (h *Handler) adminUpdatePage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST required", http.StatusMethodNotAllowed)
		return
	}
	entry, _ := getSessionEntry(r)
	page, err := h.svc.Pages.UpdatePage(model.PageContent{
		Key:       r.PathValue("pageKey"),
		Title:     r.FormValue("title"),
		Summary:   r.FormValue("summary"),
		ContentMD: r.FormValue("content"),
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	warnAfterMutation(r, "update_page_audit", h.svc.Admin.LogPageUpdated(entry.username, page, clientIP(r)))
	http.Redirect(w, r, "/admin/pages", http.StatusSeeOther)
}

func (h *Handler) adminServers(w http.ResponseWriter, r *http.Request) {
	servers, err := h.svc.Monitor.ListServers()
	if err != nil {
		internalServerError(w, r, err)
		return
	}
	serverID, token := popAgentTokenFlash(r)
	h.renderAdminServers(w, r, servers, serverID, token)
}

func (h *Handler) adminServerDetail(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("serverID"), 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	server, err := h.svc.Monitor.GetServer(id)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			internalServerError(w, r, err)
			return
		}
		http.Error(w, "server not found", http.StatusNotFound)
		return
	}
	metrics, err := h.svc.Monitor.GetRecentMetrics(id, 50)
	if err != nil {
		internalServerError(w, r, err)
		return
	}
	logs, err := h.svc.Admin.ListServerAuditLogs(id, 20)
	if err != nil {
		internalServerError(w, r, err)
		return
	}
	renderHTML(w, r, adminview.ServerDetail(adminview.ServerDetailData{
		CurrentPath:     r.URL.Path,
		CSRFToken:       csrfTokenForRequest(r),
		Server:          server,
		Metrics:         metrics,
		Logs:            logs,
		AgentURLExample: agentURLExample(h.cfg.Domain),
	}))
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

	server, token, err := h.svc.Monitor.CreateServerWithAgentToken(name, host, port, tags)
	if err != nil {
		internalServerError(w, r, err)
		return
	}

	warnAfterMutation(r, "create_server_audit", h.svc.Admin.LogServerCreated(entry.username, name, host, portStr, clientIP(r)))
	setAgentTokenFlash(r, server.ID, token)
	http.Redirect(w, r, "/admin/servers", http.StatusSeeOther)
}

func (h *Handler) adminUpdateServer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST required", http.StatusMethodNotAllowed)
		return
	}

	entry, _ := getSessionEntry(r)
	id, err := strconv.ParseInt(r.PathValue("serverID"), 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	port, err := strconv.Atoi(r.FormValue("port"))
	if err != nil {
		http.Error(w, "invalid port", http.StatusBadRequest)
		return
	}
	server, err := h.svc.Monitor.UpdateServer(id, r.FormValue("name"), r.FormValue("host"), port, r.FormValue("tags"))
	if err != nil {
		if errors.Is(err, service.ErrServerNotFound) {
			http.Error(w, "server not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	warnAfterMutation(r, "update_server_audit", h.svc.Admin.LogServerUpdated(entry.username, id, server.Name, server.Host, strconv.Itoa(server.Port), clientIP(r)))
	http.Redirect(w, r, "/admin/servers/"+strconv.FormatInt(id, 10), http.StatusSeeOther)
}

func (h *Handler) adminRotateServerToken(w http.ResponseWriter, r *http.Request) {
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
	token, err := h.svc.Monitor.RotateAgentToken(id)
	if err != nil {
		if errors.Is(err, service.ErrServerNotFound) {
			http.Error(w, "server not found", http.StatusNotFound)
			return
		}
		internalServerError(w, r, err)
		return
	}
	warnAfterMutation(r, "rotate_agent_token_audit", h.svc.Admin.LogAgentTokenRotated(entry.username, idStr, clientIP(r)))
	setAgentTokenFlash(r, id, token)
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

	if err := h.svc.Admin.DeleteServer(entry.username, id, idStr, clientIP(r)); err != nil {
		if errors.Is(err, service.ErrServerNotFound) {
			http.Error(w, "server not found", http.StatusNotFound)
			return
		}
		internalServerError(w, r, err)
		return
	}
	http.Redirect(w, r, "/admin/servers", http.StatusSeeOther)
}

func (h *Handler) renderAdminServers(w http.ResponseWriter, r *http.Request, servers []model.Server, tokenServerID int64, token string) {
	renderHTML(w, r, adminview.Servers(adminview.ServersData{
		CurrentPath:            r.URL.Path,
		CSRFToken:              csrfTokenForRequest(r),
		Servers:                servers,
		AgentURLExample:        agentURLExample(h.cfg.Domain),
		GeneratedTokenServerID: tokenServerID,
		GeneratedAgentToken:    token,
	}))
}

func agentURLExample(domain string) string {
	domain = strings.TrimSpace(domain)
	if domain == "" {
		return "http://192.168.100.5:8080"
	}
	if strings.HasPrefix(domain, "http://") || strings.HasPrefix(domain, "https://") {
		return domain
	}
	return "https://" + domain
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

func splitTags(tags string) []string {
	parts := strings.Split(tags, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}

func parseSortOrder(value string) int {
	n, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil || n < 0 {
		return 0
	}
	return n
}
