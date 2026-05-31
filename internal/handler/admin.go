package handler

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"raevtar/internal/model"
	"raevtar/internal/repo"
)

// --- Admin: Dashboard (from existing admin.go) ---

func (h *Handler) adminIndex(w http.ResponseWriter, r *http.Request) {
	allPosts, _, _ := h.svc.Blog.ListPosts("", 1, 9999)
	postCount := 0
	if allPosts != nil {
		postCount = len(allPosts)
	}

	servers, _ := h.svc.Monitor.ListServers()
	serverCount := len(servers)

	users, _ := h.svc.Repos.User.List()
	userCount := len(users)

	// Calculate online/offline + build server rows
	onlineCount := 0
	var serverRows string
	for _, s := range servers {
		online := s.LastSeen != nil && time.Since(*s.LastSeen) < 10*time.Minute
		if online {
			onlineCount++
		}

		statusDot := "bg-red-500"
		statusText := "Offline"
		if online {
			statusDot = "bg-green-500"
			statusText = "Online"
		}

		lastSeenStr := "-"
		if s.LastSeen != nil {
			lastSeenStr = s.LastSeen.Local().Format("15:04")
		}

		serverRows += `<div class="flex items-center justify-between p-4 bg-zinc-900 border border-zinc-800 rounded-lg">
			<div class="flex items-center gap-3">
				<span class="relative flex h-2.5 w-2.5">
					<span class="` + statusDot + ` inline-flex h-2.5 w-2.5 rounded-full"></span>
				</span>
				<div>
					<p class="text-sm font-medium text-white">` + s.Name + `</p>
					<p class="text-xs text-zinc-500 font-mono">` + s.Host + `:` + itoa(s.Port) + `</p>
				</div>
			</div>
			<div class="flex items-center gap-4">
				<span class="text-xs text-zinc-500 font-mono">` + lastSeenStr + `</span>
				<span class="text-xs px-2 py-0.5 rounded-full font-medium ` + textColorForStatus(statusText) + ` ` + bgColorForStatus(statusText) + `">` + statusText + `</span>
			</div>
		</div>`
	}

	if serverRows == "" {
		serverRows = `<div class="p-6 text-center text-zinc-500 text-sm">No servers registered</div>`
	}

	heroStatusText := "All systems operational"
	heroStatusColor := "text-green-400"
	heroDotColor := "bg-green-500"
	heroPingColor := "bg-green-400"
	if onlineCount == 0 && serverCount > 0 {
		heroStatusText = "No servers online"
		heroStatusColor = "text-red-400"
		heroDotColor = "bg-red-500"
		heroPingColor = "bg-red-400"
	} else if serverCount == 0 {
		heroStatusText = "No servers registered"
		heroStatusColor = "text-zinc-500"
		heroDotColor = "bg-zinc-500"
		heroPingColor = "bg-zinc-400"
	}

	content := `<div class="space-y-6">

		<!-- Hero Metric -->
		<div class="bg-zinc-900 border border-zinc-800 rounded-xl p-6 md:p-8">
			<div class="flex items-start justify-between mb-4">
				<div>
					<p class="text-xs font-medium text-zinc-500 uppercase tracking-widest">Online Servers</p>
				</div>
				<div class="flex items-center gap-2 px-3 py-1.5 bg-zinc-950 border border-zinc-800 rounded-full">
					<span class="relative flex h-2 w-2">
						<span class="animate-ping absolute inline-flex h-full w-full rounded-full ` + heroPingColor + ` opacity-75"></span>
						<span class="relative inline-flex rounded-full h-2 w-2 ` + heroDotColor + `"></span>
					</span>
					<span class="text-xs text-zinc-400 font-mono">` + itoa(onlineCount) + `/` + itoa(serverCount) + `</span>
				</div>
			</div>
			<div class="flex items-baseline gap-2">
				<span class="text-5xl md:text-6xl font-bold font-mono tracking-tight text-white">` + itoa(onlineCount) + `</span>
				<span class="text-sm text-zinc-500 lowercase">/ ` + itoa(serverCount) + ` servers</span>
			</div>
			<p class="text-xs mt-2 ` + heroStatusColor + `">` + heroStatusText + `</p>
		</div>

		<!-- Metric Cards Row -->
		<div class="grid grid-cols-2 md:grid-cols-4 gap-3 md:gap-4">
			<div class="bg-zinc-900 border border-zinc-800 rounded-xl p-4 md:p-5">
				<p class="text-xs text-zinc-500 font-medium mb-1">Posts</p>
				<p class="text-2xl md:text-3xl font-bold font-mono text-white">` + itoa(postCount) + `</p>
				<p class="text-xs text-zinc-600 mt-0.5">total articles</p>
			</div>
			<div class="bg-zinc-900 border border-zinc-800 rounded-xl p-4 md:p-5">
				<p class="text-xs text-zinc-500 font-medium mb-1">Servers</p>
				<p class="text-2xl md:text-3xl font-bold font-mono text-white">` + itoa(serverCount) + `</p>
				<p class="text-xs text-zinc-600 mt-0.5">registered</p>
			</div>
			<div class="bg-zinc-900 border border-zinc-800 rounded-xl p-4 md:p-5">
				<p class="text-xs text-zinc-500 font-medium mb-1">Users</p>
				<p class="text-2xl md:text-3xl font-bold font-mono text-white">` + itoa(userCount) + `</p>
				<p class="text-xs text-zinc-600 mt-0.5">accounts</p>
			</div>
			<div class="bg-zinc-900 border border-zinc-800 rounded-xl p-4 md:p-5">
				<p class="text-xs text-zinc-500 font-medium mb-1">System</p>
				<p class="text-2xl md:text-3xl font-bold font-mono text-green-400">OK</p>
				<p class="text-xs text-zinc-600 mt-0.5">binary running</p>
			</div>
		</div>

		<!-- Server List -->
		<div>
			<div class="flex items-center justify-between mb-3">
				<h2 class="text-sm font-semibold text-zinc-200">Server Status</h2>
				<a href="/admin/servers" class="text-xs text-zinc-500 hover:text-zinc-300 transition-colors no-underline">Manage →</a>
			</div>
			<div class="space-y-2">
				` + serverRows + `
			</div>
		</div>

		<!-- Quick Links -->
		<div>
			<h2 class="text-sm font-semibold text-zinc-200 mb-3">Quick Actions</h2>
			<div class="grid grid-cols-2 md:grid-cols-4 gap-3">
				<a href="/admin/posts" class="block bg-zinc-900 border border-zinc-800 rounded-xl p-4 hover:border-green-500/30 transition-colors no-underline group">
					<p class="text-sm text-zinc-400 group-hover:text-white transition-colors">Posts</p>
					<p class="text-xs text-zinc-600 mt-0.5">Manage articles</p>
				</a>
				<a href="/admin/servers" class="block bg-zinc-900 border border-zinc-800 rounded-xl p-4 hover:border-green-500/30 transition-colors no-underline group">
					<p class="text-sm text-zinc-400 group-hover:text-white transition-colors">Servers</p>
					<p class="text-xs text-zinc-600 mt-0.5">Add or remove</p>
				</a>
				<a href="/admin/manage-users" class="block bg-zinc-900 border border-zinc-800 rounded-xl p-4 hover:border-green-500/30 transition-colors no-underline group">
					<p class="text-sm text-zinc-400 group-hover:text-white transition-colors">Users</p>
					<p class="text-xs text-zinc-600 mt-0.5">Manage accounts</p>
				</a>
				<a href="/admin/audit-log" class="block bg-zinc-900 border border-zinc-800 rounded-xl p-4 hover:border-green-500/30 transition-colors no-underline group">
					<p class="text-sm text-zinc-400 group-hover:text-white transition-colors">Audit</p>
					<p class="text-xs text-zinc-600 mt-0.5">View activity log</p>
				</a>
			</div>
		</div>

	</div>`

	html := adminLayout("Dashboard", content)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(html))
}

func textColorForStatus(status string) string {
	if status == "Online" {
		return "text-green-400"
	}
	return "text-red-400"
}

func bgColorForStatus(status string) string {
	if status == "Online" {
		return "bg-green-500/10"
	}
	return "bg-red-500/10"
}

// --- Admin: Manage Users ---

func (h *Handler) adminUsers(w http.ResponseWriter, r *http.Request) {
	users, _ := h.svc.Repos.User.List()
	entry, _ := getSessionEntry(r)

	var rows string
	for _, u := range users {
		roleBadge := roleBadgeHTML(u.Role)

		canManage := model.CanManage(entry.role, u.Role)
		delBtn := ""
		if canManage && u.Username != entry.username {
			delBtn = `<a href="/admin/manage-users/delete/` + itoa64(u.ID) + `" class="text-xs text-zinc-600 hover:text-red-400 transition-colors no-underline" onclick="return confirm('Hapus user ` + u.Username + `?')">✕</a>`
		}

		rows += `<tr class="border-b border-zinc-800 hover:bg-zinc-800/30">
			<td class="py-3 px-4 text-sm text-white font-medium">` + u.Username + `</td>
			<td class="py-3 px-4">` + roleBadge + `</td>
			<td class="py-3 px-4 text-sm text-zinc-500 font-mono">` + u.CreatedAt.Format("2006-01-02") + `</td>
			<td class="py-3 px-4 text-right text-xs">
				` + delBtn + `
			</td>
		</tr>`
	}

	// Build role options for create form (only roles this user can create)
	var roleOptions string
	for _, r := range model.ValidRoles() {
		if model.CanManage(entry.role, r) {
			selected := ""
			if r == model.RoleOperator {
				selected = " selected"
			}
			roleOptions += `<option value="` + r + `"` + selected + `>` + r + `</option>`
		}
	}

	content := `<div class="space-y-6">

		<div class="flex items-center justify-between">
			<div>
				<a href="/admin" class="text-xs text-zinc-500 hover:text-zinc-300 transition-colors no-underline">&larr; Dashboard</a>
				<h1 class="text-xl font-bold text-white mt-1">Manage Users</h1>
			</div>
			<span class="text-xs text-zinc-500 bg-zinc-900 px-2.5 py-1 rounded-full border border-zinc-800 font-mono">` + itoa(len(users)) + ` total</span>
		</div>

		<div class="bg-zinc-900 border border-zinc-800 rounded-xl overflow-x-auto">
			<table class="w-full text-left">
				<thead>
					<tr class="border-b border-zinc-800 text-xs uppercase tracking-wider text-zinc-500">
						<th class="py-3 px-4 font-medium">Username</th>
						<th class="py-3 px-4 font-medium">Role</th>
						<th class="py-3 px-4 font-medium">Created</th>
						<th class="py-3 px-4"></th>
					</tr>
				</thead>
				<tbody>
					` + rows + `
				</tbody>
			</table>
		</div>

		<div class="bg-zinc-900 border border-zinc-800 rounded-xl p-6">
			<h2 class="text-sm font-semibold text-zinc-200 mb-4">Add New User</h2>
			<form method="POST" action="/admin/manage-users" class="grid grid-cols-1 md:grid-cols-3 gap-3">
				<div>
					<input type="text" name="username" required placeholder="Username" class="w-full px-3 py-2 rounded-lg bg-zinc-950 border border-zinc-800 text-sm text-white placeholder-zinc-600 focus:outline-none focus:border-green-500/50 transition-colors">
				</div>
				<div>
					<input type="password" name="password" required placeholder="Password (min 8 chars)" class="w-full px-3 py-2 rounded-lg bg-zinc-950 border border-zinc-800 text-sm text-white placeholder-zinc-600 focus:outline-none focus:border-green-500/50 transition-colors">
				</div>
				<div>
					<select name="role" class="w-full px-3 py-2 rounded-lg bg-zinc-950 border border-zinc-800 text-sm text-white focus:outline-none focus:border-green-500/50 transition-colors">
						<option value="">Select role...</option>
						` + roleOptions + `
					</select>
				</div>
				<div class="md:col-span-3 flex justify-end">
					<button type="submit" class="px-5 py-2 rounded-lg bg-green-600 hover:bg-green-500 text-white text-sm font-medium transition-colors">Add User</button>
				</div>
			</form>
		</div>

	</div>`

	html := adminLayout("Manage Users", content)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(html))
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

	if !model.IsValidRole(role) {
		role = model.RoleOperator
	}

	// Check permission: can't assign a role higher than yourself
	if !model.CanManage(entry.role, role) {
		http.Error(w, "you cannot create users with this role", http.StatusForbidden)
		return
	}

	hash, err := repo.HashPassword(password)
	if err != nil {
		http.Error(w, "failed to hash password", http.StatusInternalServerError)
		return
	}

	u, err := h.svc.Repos.User.Create(username, hash, role, username)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.svc.Repos.Audit.Insert(entry.username, "CREATE_USER", "created user "+u.Username+" with role "+u.Role, r.RemoteAddr)
	http.Redirect(w, r, "/admin/manage-users", http.StatusSeeOther)
}

func (h *Handler) adminDeleteUser(w http.ResponseWriter, r *http.Request) {
	entry, _ := getSessionEntry(r)
	idStr := r.PathValue("userID")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	targetUser, err := h.svc.Repos.User.GetByID(id)
	if err != nil {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	// Can't delete yourself
	if targetUser.Username == entry.username {
		http.Error(w, "cannot delete yourself", http.StatusForbidden)
		return
	}

	// Can't delete users of higher privilege
	if !model.CanManage(entry.role, targetUser.Role) {
		http.Error(w, "you cannot delete this user", http.StatusForbidden)
		return
	}

	if err := h.svc.Repos.User.Delete(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.svc.Repos.Audit.Insert(entry.username, "DELETE_USER", "deleted user "+targetUser.Username, r.RemoteAddr)
	http.Redirect(w, r, "/admin/manage-users", http.StatusSeeOther)
}

// --- Admin: Audit Log ---

func (h *Handler) adminAuditLog(w http.ResponseWriter, r *http.Request) {
	logs, _ := h.svc.Repos.Audit.List(200, 0)

	var rows string
	for _, l := range logs {
		actionBadge := ""
		switch l.Action {
		case "LOGIN", "LOGIN_SUCCESS":
			actionBadge = `<span class="text-xs px-2 py-0.5 rounded-md bg-green-500/10 text-green-400 border border-green-500/20 font-mono">` + l.Action + `</span>`
		case "LOGIN_FAILED":
			actionBadge = `<span class="text-xs px-2 py-0.5 rounded-md bg-red-500/10 text-red-400 border border-red-500/20 font-mono">` + l.Action + `</span>`
		case "LOGOUT":
			actionBadge = `<span class="text-xs px-2 py-0.5 rounded-md bg-zinc-500/10 text-zinc-400 border border-zinc-500/20 font-mono">` + l.Action + `</span>`
		case "CREATE_POST", "CREATE_USER", "CREATE_SERVER":
			actionBadge = `<span class="text-xs px-2 py-0.5 rounded-md bg-blue-500/10 text-blue-400 border border-blue-500/20 font-mono">` + l.Action + `</span>`
		case "DELETE_POST", "DELETE_USER", "DELETE_SERVER":
			actionBadge = `<span class="text-xs px-2 py-0.5 rounded-md bg-red-500/10 text-red-400 border border-red-500/20 font-mono">` + l.Action + `</span>`
		default:
			actionBadge = `<span class="text-xs px-2 py-0.5 rounded-md bg-zinc-500/10 text-zinc-400 border border-zinc-500/20 font-mono">` + l.Action + `</span>`
		}

		rows += `<tr class="border-b border-zinc-800 hover:bg-zinc-800/30 text-xs">
			<td class="py-2.5 px-4 text-zinc-500 font-mono whitespace-nowrap">` + l.CreatedAt.Format("02 Jan 15:04") + `</td>
			<td class="py-2.5 px-4">` + actionBadge + `</td>
			<td class="py-2.5 px-4 text-zinc-300">` + l.User + `</td>
			<td class="py-2.5 px-4 text-zinc-500 max-w-xs truncate">` + l.Details + `</td>
			<td class="py-2.5 px-4 text-zinc-600 font-mono">` + l.IP + `</td>
		</tr>`
	}

	if rows == "" {
		rows = `<tr><td colspan="5" class="py-8 text-center text-zinc-600 text-sm">No audit log entries yet</td></tr>`
	}

	content := `<div class="space-y-6">

		<div class="flex items-center justify-between">
			<div>
				<a href="/admin" class="text-xs text-zinc-500 hover:text-zinc-300 transition-colors no-underline">&larr; Dashboard</a>
				<h1 class="text-xl font-bold text-white mt-1">Audit Log</h1>
			</div>
			<span class="text-xs text-zinc-500 bg-zinc-900 px-2.5 py-1 rounded-full border border-zinc-800 font-mono">` + itoa(len(logs)) + ` entries</span>
		</div>

		<div class="bg-zinc-900 border border-zinc-800 rounded-xl overflow-x-auto">
			<table class="w-full text-left">
				<thead>
					<tr class="border-b border-zinc-800 text-xs uppercase tracking-wider text-zinc-500">
						<th class="py-3 px-4 font-medium">Time</th>
						<th class="py-3 px-4 font-medium">Action</th>
						<th class="py-3 px-4 font-medium">User</th>
						<th class="py-3 px-4 font-medium">Details</th>
						<th class="py-3 px-4 font-medium">IP</th>
					</tr>
				</thead>
				<tbody>
					` + rows + `
				</tbody>
			</table>
		</div>

	</div>`

	html := adminLayout("Audit Log", content)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(html))
}

// --- Admin: Manage posts (from existing admin.go, same) ---

func (h *Handler) adminPosts(w http.ResponseWriter, r *http.Request) {
	posts, _, _ := h.svc.Blog.ListPosts("", 1, 9999)
	categories, _ := h.svc.Repos.Category.List()

	var rows string
	for _, p := range posts {
		catColor := "bg-cyan-500/10 text-cyan-400 border-cyan-500/20"
		switch p.CategorySlug {
		case "ai-agent":
			catColor = "bg-purple-500/10 text-purple-400 border-purple-500/20"
		case "security":
			catColor = "bg-red-500/10 text-red-400 border-red-500/20"
		case "kernel-embedded":
			catColor = "bg-orange-500/10 text-orange-400 border-orange-500/20"
		case "devops":
			catColor = "bg-blue-500/10 text-blue-400 border-blue-500/20"
		case "tools":
			catColor = "bg-green-500/10 text-green-400 border-green-500/20"
		}

		rows += `<tr class="border-b border-zinc-800 hover:bg-zinc-800/30">
			<td class="py-3 px-4 text-sm text-white max-w-[300px] truncate">` + p.Title + `</td>
			<td class="py-3 px-4"><span class="text-xs px-2 py-0.5 rounded-md font-mono border ` + catColor + `">` + p.CategorySlug + `</span></td>
			<td class="py-3 px-4 text-sm text-zinc-500 font-mono">` + p.CreatedAt.Format("2006-01-02") + `</td>
			<td class="py-3 px-4 text-right">
				<a href="/admin/posts/delete/` + itoa64(p.ID) + `" class="text-xs text-zinc-600 hover:text-red-400 transition-colors no-underline" onclick="return confirm('Hapus post ini?')">✕</a>
			</td>
		</tr>`
	}

	var catOptions string
	for _, c := range categories {
		catOptions += `<option value="` + c.Slug + `">` + c.Name + `</option>`
	}

	content := `<div class="space-y-6">

		<div class="flex items-center justify-between">
			<div>
				<a href="/admin" class="text-xs text-zinc-500 hover:text-zinc-300 transition-colors no-underline">&larr; Dashboard</a>
				<h1 class="text-xl font-bold text-white mt-1">Manage Posts</h1>
			</div>
			<span class="text-xs text-zinc-500 bg-zinc-900 px-2.5 py-1 rounded-full border border-zinc-800 font-mono">` + itoa(len(posts)) + ` total</span>
		</div>

		<div class="bg-zinc-900 border border-zinc-800 rounded-xl overflow-x-auto">
			<table class="w-full text-left">
				<thead>
					<tr class="border-b border-zinc-800 text-xs uppercase tracking-wider text-zinc-500">
						<th class="py-3 px-4 font-medium">Title</th>
						<th class="py-3 px-4 font-medium">Category</th>
						<th class="py-3 px-4 font-medium">Date</th>
						<th class="py-3 px-4"></th>
					</tr>
				</thead>
				<tbody>
					` + rows + `
				</tbody>
			</table>
		</div>

		<div class="bg-zinc-900 border border-zinc-800 rounded-xl p-6">
			<h2 class="text-sm font-semibold text-zinc-200 mb-4">Create New Post</h2>
			<form method="POST" action="/admin/posts" class="space-y-4">
				<div>
					<label class="block text-sm text-zinc-500 mb-1">Title</label>
					<input type="text" name="title" required class="w-full px-3 py-2 rounded-lg bg-zinc-950 border border-zinc-800 text-sm text-white placeholder-zinc-600 focus:outline-none focus:border-green-500/50 transition-colors" placeholder="Post title">
				</div>
				<div>
					<label class="block text-sm text-zinc-500 mb-1">Category</label>
					<select name="category_slug" required class="w-full px-3 py-2 rounded-lg bg-zinc-950 border border-zinc-800 text-sm text-white focus:outline-none focus:border-green-500/50 transition-colors">
						<option value="">Select category</option>` + catOptions + `
					</select>
				</div>
				<div>
					<label class="block text-sm text-zinc-500 mb-1">Content (Markdown)</label>
					<textarea name="content" rows="15" required class="w-full px-3 py-2 rounded-lg bg-zinc-950 border border-zinc-800 text-sm text-white font-mono placeholder-zinc-600 focus:outline-none focus:border-green-500/50 transition-colors" placeholder="Write in Markdown..."></textarea>
				</div>
				<div class="flex justify-end">
					<button type="submit" class="px-5 py-2 rounded-lg bg-green-600 hover:bg-green-500 text-white text-sm font-medium transition-colors">Publish Post</button>
				</div>
			</form>
		</div>

	</div>`

	html := adminLayout("Manage Posts", content)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(html))
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

	h.svc.Repos.Audit.Insert(entry.username, "CREATE_POST", "created post: "+post.Title, r.RemoteAddr)
	http.Redirect(w, r, "/admin/posts", http.StatusSeeOther)
}

func (h *Handler) adminDeletePost(w http.ResponseWriter, r *http.Request) {
	entry, _ := getSessionEntry(r)
	idStr := r.PathValue("postID")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	post, err := h.svc.Repos.Post.GetByID(id)
	if err == nil {
		h.svc.Repos.Audit.Insert(entry.username, "DELETE_POST", "deleted post: "+post.Title, r.RemoteAddr)
	}

	h.svc.Repos.Post.Delete(id)
	http.Redirect(w, r, "/admin/posts", http.StatusSeeOther)
}

// --- Admin: Manage servers ---

func (h *Handler) adminServers(w http.ResponseWriter, r *http.Request) {
	servers, _ := h.svc.Monitor.ListServers()

	var cards string
	for _, s := range servers {
		online := s.LastSeen != nil && time.Since(*s.LastSeen) < 10*time.Minute

		statusDot := "bg-red-500"
		statusPing := "bg-red-400"
		statusText := "Offline"
		statusBadgeColor := "text-red-400"
		if online {
			statusDot = "bg-green-500"
			statusPing = "bg-green-400"
			statusText = "Online"
			statusBadgeColor = "text-green-400"
		}

		lastSeenStr := "Never"
		if s.LastSeen != nil {
			lastSeenStr = s.LastSeen.Local().Format("02 Jan 2006 15:04")
		}

		// Tags badges
		var tagBadges string
		if s.Tags != "" {
			tags := strings.Split(s.Tags, ",")
			for _, t := range tags {
				t = strings.TrimSpace(t)
				if t != "" {
					tagBadges += `<span class="text-xs px-2 py-0.5 rounded-full bg-zinc-800 text-zinc-400 border border-zinc-700 font-mono">` + t + `</span>`
				}
			}
		}

		cards += `<div class="bg-zinc-900 border border-zinc-800 rounded-xl p-5 hover:border-zinc-700 transition-colors">
			<div class="flex items-start justify-between mb-4">
				<div class="flex items-center gap-3">
					<div class="w-9 h-9 rounded-lg bg-zinc-950 border border-zinc-700 flex items-center justify-center">
						<span class="text-sm font-bold text-zinc-400 font-mono">` + s.Name[:min(len(s.Name), 2)] + `</span>
					</div>
					<div>
						<p class="text-sm font-semibold text-white">` + s.Name + `</p>
						<p class="text-xs text-zinc-500 font-mono">` + s.Host + `:` + itoa(s.Port) + `</p>
					</div>
				</div>
				<div class="flex items-center gap-2">
					<span class="relative flex h-2.5 w-2.5">
						<span class="animate-ping absolute inline-flex h-full w-full rounded-full ` + statusPing + ` opacity-75"></span>
						<span class="relative inline-flex rounded-full h-2.5 w-2.5 ` + statusDot + `"></span>
					</span>
					<span class="text-xs font-medium ` + statusBadgeColor + `">` + statusText + `</span>
				</div>
			</div>
			<div class="flex items-center justify-between">
				<div class="flex flex-wrap gap-1.5">
					` + tagBadges + `
				</div>
				<div class="flex items-center gap-3">
					<span class="text-xs text-zinc-600">Last seen: ` + lastSeenStr + `</span>
					<a href="/admin/servers/delete/` + itoa64(s.ID) + `" class="text-xs text-zinc-600 hover:text-red-400 transition-colors no-underline" onclick="return confirm('Hapus server ` + s.Name + `?')">✕</a>
				</div>
			</div>
		</div>`
	}

	if cards == "" {
		cards = `<div class="p-10 text-center text-zinc-500 text-sm border border-dashed border-zinc-800 rounded-xl">No servers registered yet</div>`
	}

	content := `<div class="space-y-6">

		<!-- Header -->
		<div class="flex items-center justify-between">
			<div>
				<a href="/admin" class="text-xs text-zinc-500 hover:text-zinc-300 transition-colors no-underline">&larr; Dashboard</a>
				<h1 class="text-xl font-bold text-white mt-1">Manage Servers</h1>
			</div>
			<span class="text-xs text-zinc-500 bg-zinc-900 px-2.5 py-1 rounded-full border border-zinc-800 font-mono">` + itoa(len(servers)) + ` total</span>
		</div>

		<!-- Server Cards -->
		<div class="space-y-3">
			` + cards + `
		</div>

		<!-- Register Form -->
		<div class="bg-zinc-900 border border-zinc-800 rounded-xl p-6">
			<h2 class="text-sm font-semibold text-zinc-200 mb-4">Register New Server</h2>
			<form method="POST" action="/admin/servers" class="grid grid-cols-1 md:grid-cols-4 gap-3">
				<div>
					<input type="text" name="name" required placeholder="Name" class="w-full px-3 py-2 rounded-lg bg-zinc-950 border border-zinc-800 text-sm text-white placeholder-zinc-600 focus:outline-none focus:border-green-500/50 transition-colors">
				</div>
				<div>
					<input type="text" name="host" required placeholder="IP or hostname" class="w-full px-3 py-2 rounded-lg bg-zinc-950 border border-zinc-800 text-sm text-white placeholder-zinc-600 focus:outline-none focus:border-green-500/50 transition-colors">
				</div>
				<div>
					<input type="number" name="port" value="22" placeholder="Port" class="w-full px-3 py-2 rounded-lg bg-zinc-950 border border-zinc-800 text-sm text-white placeholder-zinc-600 focus:outline-none focus:border-green-500/50 transition-colors font-mono">
				</div>
				<div>
					<input type="text" name="tags" placeholder="tags (comma,separated)" class="w-full px-3 py-2 rounded-lg bg-zinc-950 border border-zinc-800 text-sm text-white placeholder-zinc-600 focus:outline-none focus:border-green-500/50 transition-colors">
				</div>
				<div class="md:col-span-4 flex justify-end">
					<button type="submit" class="px-5 py-2 rounded-lg bg-green-600 hover:bg-green-500 text-white text-sm font-medium transition-colors">Add Server</button>
				</div>
			</form>
		</div>

	</div>`

	html := adminLayout("Manage Servers", content)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(html))
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (h *Handler) adminCreateServer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST required", http.StatusMethodNotAllowed)
		return
	}

	entry, _ := getSessionEntry(r)
	name := r.FormValue("name")
	host := r.FormValue("host")
	port := 22
	if p, err := strconv.Atoi(r.FormValue("port")); err == nil {
		port = p
	}
	tags := r.FormValue("tags")

	if name == "" || host == "" {
		http.Error(w, "name and host required", http.StatusBadRequest)
		return
	}

	if _, err := h.svc.Monitor.CreateServer(name, host, port, tags); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.svc.Repos.Audit.Insert(entry.username, "CREATE_SERVER", "registered server: "+name, r.RemoteAddr)
	http.Redirect(w, r, "/admin/servers", http.StatusSeeOther)
}

func (h *Handler) adminDeleteServer(w http.ResponseWriter, r *http.Request) {
	entry, _ := getSessionEntry(r)
	idStr := r.PathValue("serverID")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	h.svc.Repos.Audit.Insert(entry.username, "DELETE_SERVER", "deleted server id: "+idStr, r.RemoteAddr)
	h.svc.Repos.Server.Delete(id)
	http.Redirect(w, r, "/admin/servers", http.StatusSeeOther)
}

// --- Helpers ---

func roleBadgeHTML(role string) string {
	switch role {
	case model.RoleOwner:
		return `<span class="text-xs px-1.5 py-0.5 rounded bg-red-500/10 text-red-400 border border-red-500/30 font-medium">owner</span>`
	case model.RoleAdmin:
		return `<span class="text-xs px-1.5 py-0.5 rounded bg-amber-500/10 text-amber-400 border border-amber-500/30 font-medium">admin</span>`
	case model.RoleOperator:
		return `<span class="text-xs px-1.5 py-0.5 rounded bg-blue-500/10 text-blue-400 border border-blue-500/30 font-medium">operator</span>`
	case model.RoleReadonly:
		return `<span class="text-xs px-1.5 py-0.5 rounded bg-slate-500/10 text-slate-400 border border-slate-500/30 font-medium">readonly</span>`
	default:
		return `<span class="text-xs px-1.5 py-0.5 rounded bg-slate-500/10 text-slate-400 border border-slate-500/30 font-medium">` + role + `</span>`
	}
}

func adminLayout(title, content string) string {
	return `<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<title>` + title + ` — Raevtar Admin</title>
	<link rel="preconnect" href="https://fonts.googleapis.com">
	<link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
	<link href="https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700&family=JetBrains+Mono:wght@400;500&display=swap" rel="stylesheet">
	<link rel="stylesheet" href="/static/css/style.css">
	<link rel="icon" type="image/svg+xml" href="/static/favicon.svg">
</head>
<body class="bg-zinc-950 text-zinc-100 font-sans min-h-screen">
	<div class="min-h-screen flex">
		<!-- Sidebar (desktop) -->
		<aside class="w-64 bg-zinc-900 border-r border-zinc-800 hidden md:flex flex-col">
			<div class="p-6 flex items-center gap-3">
				<div class="w-8 h-8 bg-green-500 rounded-lg flex items-center justify-center">
					<span class="text-zinc-950 font-black">R</span>
				</div>
				<span class="text-xl font-bold tracking-tight">Raevtar</span>
			</div>
			<nav class="flex-1 px-4 space-y-2 mt-2">
				<a href="/admin" class="flex items-center gap-3 px-3 py-2.5 text-zinc-200 hover:bg-zinc-800/50 rounded-lg transition-colors font-medium">Dashboard</a>
				<a href="/admin/posts" class="flex items-center gap-3 px-3 py-2.5 text-zinc-400 hover:text-zinc-100 hover:bg-zinc-800/50 rounded-lg transition-colors font-medium">Posts</a>
				<a href="/admin/servers" class="flex items-center gap-3 px-3 py-2.5 text-zinc-400 hover:text-zinc-100 hover:bg-zinc-800/50 rounded-lg transition-colors font-medium">Servers</a>
				<a href="/admin/manage-users" class="flex items-center gap-3 px-3 py-2.5 text-zinc-400 hover:text-zinc-100 hover:bg-zinc-800/50 rounded-lg transition-colors font-medium">Users</a>
				<a href="/admin/audit-log" class="flex items-center gap-3 px-3 py-2.5 text-zinc-400 hover:text-zinc-100 hover:bg-zinc-800/50 rounded-lg transition-colors font-medium">Audit</a>
			</nav>
			<div class="p-4 border-t border-zinc-800">
				<a href="/admin/logout" class="block text-center px-3 py-2 rounded-lg bg-zinc-950 border border-zinc-800 text-zinc-300 hover:text-red-300 hover:border-red-500/30 transition-colors">Logout</a>
			</div>
		</aside>

		<!-- Main -->
		<main class="flex-1 flex flex-col h-screen overflow-hidden">
			<header class="h-16 border-b border-zinc-800 flex items-center justify-between px-4 md:px-6 bg-zinc-900/50 backdrop-blur-md sticky top-0 z-10">
				<div class="flex items-center gap-3">
					<div class="md:hidden w-7 h-7 bg-green-500 rounded-md flex items-center justify-center">
						<span class="text-zinc-950 font-black text-sm">R</span>
					</div>
					<div class="text-base md:text-lg font-semibold text-zinc-200">` + title + `</div>
				</div>
				<div class="flex items-center gap-3">
					<div class="flex items-center gap-2 px-3 py-1.5 bg-zinc-900 border border-zinc-800 rounded-full text-xs md:text-sm">
						<span class="relative flex h-2.5 w-2.5">
							<span class="animate-ping absolute inline-flex h-full w-full rounded-full bg-green-400 opacity-75"></span>
							<span class="relative inline-flex rounded-full h-2.5 w-2.5 bg-green-500"></span>
						</span>
						<span class="text-zinc-300 font-medium">Connected</span>
					</div>
				</div>
			</header>

			<div class="p-4 md:p-8 space-y-6 flex-1 overflow-y-auto">
				` + content + `
			</div>
		</main>
	</div>
</body>
</html>`
}
