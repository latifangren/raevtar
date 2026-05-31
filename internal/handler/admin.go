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

	stats := collectHostStats()

	// Calculate online/offline + build server rows
	onlineCount := 0
	var serverRows string
	for _, s := range servers {
		online := s.LastSeen != nil && time.Since(*s.LastSeen) < 10*time.Minute
		if online {
			onlineCount++
		}

		statusDot := "bg-rose-400"
		statusText := "Offline"
		if online {
			statusDot = "bg-emerald-400"
			statusText = "Online"
		}

		lastSeenStr := "-"
		if s.LastSeen != nil {
			lastSeenStr = s.LastSeen.Local().Format("15:04")
		}

		serverRows += `<div class="flex items-center justify-between p-4 bg-white border-2 border-black">
			<div class="flex items-center gap-3">
				<span class="inline-block w-3 h-3 border-2 border-black ` + statusDot + `"></span>
				<div>
					<p class="text-sm font-bold text-black">` + s.Name + `</p>
					<p class="text-xs font-mono text-neutral-500">` + s.Host + `:` + itoa(s.Port) + `</p>
				</div>
			</div>
			<div class="flex items-center gap-4">
				<span class="text-xs font-mono text-neutral-500">` + lastSeenStr + `</span>
				<span class="text-xs px-2 py-0.5 font-bold border-2 border-black ` + bgColorForStatus(statusText) + ` ` + textColorForStatus(statusText) + `">` + statusText + `</span>
			</div>
		</div>`
	}

	if serverRows == "" {
		serverRows = `<div class="p-6 text-center text-neutral-500 font-bold text-sm">No servers registered</div>`
	}

	heroStatusText := "All systems operational"
	heroStatusColor := "text-emerald-600"
	heroDotColor := "bg-emerald-400"
	heroPingColor := "bg-emerald-400"
	if onlineCount == 0 && serverCount > 0 {
		heroStatusText = "No servers online"
		heroStatusColor = "text-rose-600"
		heroDotColor = "bg-rose-400"
		heroPingColor = "bg-rose-400"
	} else if serverCount == 0 {
		heroStatusText = "No servers registered"
		heroStatusColor = "text-neutral-500"
		heroDotColor = "bg-neutral-300"
		heroPingColor = "bg-neutral-300"
	}

	content := `<div class="space-y-6">

		<!-- Hero Metric -->
		<div class="bg-white border-2 border-black p-6 md:p-8 shadow-[4px_4px_0px_0px_#000]">
			<div class="flex items-start justify-between mb-4">
				<div>
					<p class="text-xs font-black uppercase tracking-widest text-neutral-500">Online Servers</p>
				</div>
				<div class="flex items-center gap-2 px-3 py-1.5 border-2 border-black bg-white">
					<span class="relative flex h-3 w-3">
						<span class="animate-ping absolute inline-flex h-full w-full rounded-none ` + heroPingColor + ` opacity-75"></span>
						<span class="relative inline-flex rounded-none h-3 w-3 border-2 border-black ` + heroDotColor + `"></span>
					</span>
					<span class="text-xs font-mono font-bold text-black">` + itoa(onlineCount) + `/` + itoa(serverCount) + `</span>
				</div>
			</div>
			<div class="flex items-baseline gap-2">
				<span class="text-5xl md:text-6xl font-black font-mono tracking-tight text-black">` + itoa(onlineCount) + `</span>
				<span class="text-sm font-bold text-neutral-500 lowercase">/ ` + itoa(serverCount) + ` servers</span>
			</div>
			<p class="text-xs mt-2 font-bold ` + heroStatusColor + `">` + heroStatusText + `</p>
		</div>

		<!-- Metric Cards Row -->
		<div class="grid grid-cols-2 md:grid-cols-4 gap-4">
			<div class="bg-white border-2 border-black p-5 shadow-[4px_4px_0px_0px_#000]">
				<p class="text-xs font-black uppercase text-neutral-500 mb-1">Posts</p>
				<p class="text-3xl font-black font-mono text-black">` + itoa(postCount) + `</p>
				<p class="text-xs font-bold text-neutral-500 mt-0.5">total articles</p>
			</div>
			<div class="bg-white border-2 border-black p-5 shadow-[4px_4px_0px_0px_#000]">
				<p class="text-xs font-black uppercase text-neutral-500 mb-1">Servers</p>
				<p class="text-3xl font-black font-mono text-black">` + itoa(serverCount) + `</p>
				<p class="text-xs font-bold text-neutral-500 mt-0.5">registered</p>
			</div>
			<div class="bg-white border-2 border-black p-5 shadow-[4px_4px_0px_0px_#000]">
				<p class="text-xs font-black uppercase text-neutral-500 mb-1">Users</p>
				<p class="text-3xl font-black font-mono text-black">` + itoa(userCount) + `</p>
				<p class="text-xs font-bold text-neutral-500 mt-0.5">accounts</p>
			</div>
			<div class="bg-white border-2 border-black p-5 shadow-[4px_4px_0px_0px_#000]">
				<p class="text-xs font-black uppercase text-neutral-500 mb-1">CPU Load</p>
				<p class="text-3xl font-black font-mono ` + cpuLoadColor(stats.CPU.Load1, stats.CPU.Cores) + `">` + strconv.FormatFloat(stats.CPU.Load1, 'f', 2, 64) + `</p>
				<p class="text-xs font-bold text-neutral-500 mt-0.5">` + itoa(stats.CPU.Cores) + ` cores &middot; 5m ` + strconv.FormatFloat(stats.CPU.Load5, 'f', 2, 64) + `</p>
			</div>
		</div>

		<!-- Server List -->
		<div>
			<div class="flex items-center justify-between mb-3">
				<h2 class="text-sm font-black uppercase">Server Status</h2>
				<a href="/admin/servers" class="text-xs font-bold no-underline hover:bg-yellow-200">Manage &rarr;</a>
			</div>
			<div class="space-y-2">
				` + serverRows + `
			</div>
		</div>

		<!-- Quick Links -->
		<div>
			<h2 class="text-sm font-black uppercase mb-3">Quick Actions</h2>
			<div class="grid grid-cols-2 md:grid-cols-4 gap-4">
				<a href="/admin/posts" class="block bg-white border-2 border-black p-5 shadow-[3px_3px_0px_0px_#000] hover:shadow-[1px_1px_0px_0px_#000] hover:translate-x-[2px] hover:translate-y-[2px] transition-all no-underline group">
					<p class="text-sm font-black text-black">Posts</p>
					<p class="text-xs font-bold text-neutral-500 mt-0.5">Manage articles</p>
				</a>
				<a href="/admin/servers" class="block bg-white border-2 border-black p-5 shadow-[3px_3px_0px_0px_#000] hover:shadow-[1px_1px_0px_0px_#000] hover:translate-x-[2px] hover:translate-y-[2px] transition-all no-underline group">
					<p class="text-sm font-black text-black">Servers</p>
					<p class="text-xs font-bold text-neutral-500 mt-0.5">Add or remove</p>
				</a>
				<a href="/admin/manage-users" class="block bg-white border-2 border-black p-5 shadow-[3px_3px_0px_0px_#000] hover:shadow-[1px_1px_0px_0px_#000] hover:translate-x-[2px] hover:translate-y-[2px] transition-all no-underline group">
					<p class="text-sm font-black text-black">Users</p>
					<p class="text-xs font-bold text-neutral-500 mt-0.5">Manage accounts</p>
				</a>
				<a href="/admin/audit-log" class="block bg-white border-2 border-black p-5 shadow-[3px_3px_0px_0px_#000] hover:shadow-[1px_1px_0px_0px_#000] hover:translate-x-[2px] hover:translate-y-[2px] transition-all no-underline group">
					<p class="text-sm font-black text-black">Audit</p>
					<p class="text-xs font-bold text-neutral-500 mt-0.5">View activity log</p>
				</a>
			</div>
		</div>

		<!-- System Health -->
		<div>
			<h2 class="text-sm font-black uppercase mb-3">System Health</h2>
			<div class="grid grid-cols-1 md:grid-cols-3 gap-4">
				<div class="bg-white border-2 border-black p-5 shadow-[4px_4px_0px_0px_#000]">
					<div class="flex items-center justify-between mb-1.5">
						<p class="text-xs font-black uppercase text-neutral-500">RAM</p>
						<p class="text-xs font-mono font-bold text-neutral-600">` + formatBytes(stats.RAM.Used) + ` / ` + formatBytes(stats.RAM.Total) + `</p>
					</div>
					<div class="w-full bg-neutral-200 border-2 border-black h-4 mt-1">
						<div class="` + barColor(stats.RAM.Percent, 90, 75) + ` h-full transition-all" style="width: ` + strconv.FormatFloat(stats.RAM.Percent, 'f', 0, 64) + `%"></div>
					</div>
				</div>
				<div class="bg-white border-2 border-black p-5 shadow-[4px_4px_0px_0px_#000]">
					<div class="flex items-center justify-between mb-1.5">
						<p class="text-xs font-black uppercase text-neutral-500">Disk</p>
						<p class="text-xs font-mono font-bold text-neutral-600">` + formatBytes(stats.Disk.Used) + ` / ` + formatBytes(stats.Disk.Total) + `</p>
					</div>
					<div class="w-full bg-neutral-200 border-2 border-black h-4 mt-1">
						<div class="` + barColor(stats.Disk.Percent, 90, 75) + ` h-full transition-all" style="width: ` + strconv.FormatFloat(stats.Disk.Percent, 'f', 0, 64) + `%"></div>
					</div>
				</div>
				<div class="bg-white border-2 border-black p-5 shadow-[4px_4px_0px_0px_#000]">
					<div class="flex items-center justify-between">
						<p class="text-xs font-black uppercase text-neutral-500">Temperature</p>
						<p class="text-sm font-mono font-black ` + tempColor(stats.Temp) + `">` + strconv.FormatFloat(stats.Temp, 'f', 1, 64) + `&deg;C</p>
					</div>
				</div>
			</div>
		</div>

	</div>`

	html := adminLayout("Dashboard", content)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(html))
}

func cpuLoadColor(load float64, cores int) string {
	ratio := load / float64(cores)
	switch {
	case ratio > 0.9:
		return "text-rose-600"
	case ratio > 0.7:
		return "text-amber-600"
	default:
		return "text-emerald-600"
	}
}

func barColor(percent float64, high, mid float64) string {
	switch {
	case percent > high:
		return "bg-rose-400"
	case percent > mid:
		return "bg-amber-300"
	default:
		return "bg-emerald-400"
	}
}

func tempColor(temp float64) string {
	switch {
	case temp > 80:
		return "text-rose-600"
	case temp > 60:
		return "text-amber-600"
	default:
		return "text-emerald-600"
	}
}

func textColorForStatus(status string) string {
	if status == "Online" {
		return "text-black"
	}
	return "text-black"
}

func bgColorForStatus(status string) string {
	if status == "Online" {
		return "bg-emerald-400"
	}
	return "bg-rose-400"
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
			delBtn = `<a href="/admin/manage-users/delete/` + itoa64(u.ID) + `" class="text-xs font-bold text-neutral-500 hover:bg-rose-200 px-2 py-1 border-2 border-black no-underline" onclick="return confirm('Hapus user ` + u.Username + `?')">✕</a>`
		}

		rows += `<tr class="border-b-2 border-black">
			<td class="py-3 px-4 text-sm font-bold text-black">` + u.Username + `</td>
			<td class="py-3 px-4">` + roleBadge + `</td>
			<td class="py-3 px-4 text-sm font-mono font-bold text-neutral-500">` + u.CreatedAt.Format("2006-01-02") + `</td>
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
				<a href="/admin" class="text-xs font-bold no-underline hover:bg-yellow-200">&larr; Dashboard</a>
				<h1 class="text-xl font-black uppercase text-black mt-1">Manage Users</h1>
			</div>
			<span class="text-xs font-bold bg-white border-2 border-black px-2.5 py-1 font-mono">` + itoa(len(users)) + ` total</span>
		</div>

		<div class="bg-white border-2 border-black shadow-[4px_4px_0px_0px_#000] overflow-x-auto">
			<table class="w-full text-left">
				<thead>
					<tr class="border-b-2 border-black text-xs uppercase bg-black text-white">
						<th class="py-3 px-4 font-bold">Username</th>
						<th class="py-3 px-4 font-bold">Role</th>
						<th class="py-3 px-4 font-bold">Created</th>
						<th class="py-3 px-4"></th>
					</tr>
				</thead>
				<tbody>
					` + rows + `
				</tbody>
			</table>
		</div>

		<div class="bg-white border-2 border-black p-6 shadow-[4px_4px_0px_0px_#000]">
			<h2 class="text-sm font-black uppercase text-black mb-4">Add New User</h2>
			<form method="POST" action="/admin/manage-users" class="grid grid-cols-1 md:grid-cols-3 gap-3">
				<div>
					<input type="text" name="username" required placeholder="Username" class="w-full px-3 py-2 border-2 border-black bg-white text-sm font-bold text-black placeholder:text-neutral-400 focus:outline-none focus:ring-2 focus:ring-black">
				</div>
				<div>
					<input type="password" name="password" required placeholder="Password (min 8 chars)" class="w-full px-3 py-2 border-2 border-black bg-white text-sm font-bold text-black placeholder:text-neutral-400 focus:outline-none focus:ring-2 focus:ring-black">
				</div>
				<div>
					<select name="role" class="w-full px-3 py-2 border-2 border-black bg-white text-sm font-bold text-black focus:outline-none focus:ring-2 focus:ring-black">
						<option value="">Select role...</option>
						` + roleOptions + `
					</select>
				</div>
				<div class="md:col-span-3 flex justify-end">
					<button type="submit" class="px-5 py-2 border-2 border-black bg-emerald-400 text-black font-bold text-sm shadow-[3px_3px_0px_0px_#000] hover:shadow-[1px_1px_0px_0px_#000] hover:translate-x-[2px] hover:translate-y-[2px] transition-all cursor-pointer">Add User</button>
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

	if !model.CanManage(entry.role, targetUser.Role) {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	h.svc.Repos.Audit.Insert(entry.username, "DELETE_USER", "deleted user: "+targetUser.Username+" (role: "+targetUser.Role+")", r.RemoteAddr)
	h.svc.Repos.User.Delete(id)
	http.Redirect(w, r, "/admin/manage-users", http.StatusSeeOther)
}

// --- Admin: Audit Log ---

func (h *Handler) adminAuditLog(w http.ResponseWriter, r *http.Request) {
	logs, _ := h.svc.Repos.Audit.List(200, 0)

	var rows string
	for _, l := range logs {
		var actionBadge string
		switch {
		case strings.Contains(l.Action, "LOGIN"):
			actionBadge = `<span class="text-xs px-2 py-0.5 font-bold border-2 border-black bg-emerald-300 text-black font-mono">` + l.Action + `</span>`
		case strings.Contains(l.Action, "CREATE"):
			actionBadge = `<span class="text-xs px-2 py-0.5 font-bold border-2 border-black bg-blue-300 text-black font-mono">` + l.Action + `</span>`
		case strings.Contains(l.Action, "DELETE"):
			actionBadge = `<span class="text-xs px-2 py-0.5 font-bold border-2 border-black bg-rose-300 text-black font-mono">` + l.Action + `</span>`
		default:
			actionBadge = `<span class="text-xs px-2 py-0.5 font-bold border-2 border-black bg-neutral-200 text-black font-mono">` + l.Action + `</span>`
		}

		rows += `<tr class="border-b-2 border-black text-xs">
			<td class="py-2.5 px-4 text-neutral-500 font-mono font-bold whitespace-nowrap">` + l.CreatedAt.Format("02 Jan 15:04") + `</td>
			<td class="py-2.5 px-4">` + actionBadge + `</td>
			<td class="py-2.5 px-4 font-bold text-black">` + l.User + `</td>
			<td class="py-2.5 px-4 text-neutral-600 font-bold max-w-xs truncate">` + l.Details + `</td>
			<td class="py-2.5 px-4 text-neutral-500 font-mono">` + l.IP + `</td>
		</tr>`
	}

	if rows == "" {
		rows = `<tr><td colspan="5" class="py-8 text-center text-neutral-500 font-bold text-sm">No audit log entries yet</td></tr>`
	}

	content := `<div class="space-y-6">

		<div class="flex items-center justify-between">
			<div>
				<a href="/admin" class="text-xs font-bold no-underline hover:bg-yellow-200">&larr; Dashboard</a>
				<h1 class="text-xl font-black uppercase text-black mt-1">Audit Log</h1>
			</div>
			<span class="text-xs font-bold bg-white border-2 border-black px-2.5 py-1 font-mono">` + itoa(len(logs)) + ` entries</span>
		</div>

		<div class="bg-white border-2 border-black shadow-[4px_4px_0px_0px_#000] overflow-x-auto">
			<table class="w-full text-left">
				<thead>
					<tr class="border-b-2 border-black text-xs uppercase bg-black text-white">
						<th class="py-3 px-4 font-bold">Time</th>
						<th class="py-3 px-4 font-bold">Action</th>
						<th class="py-3 px-4 font-bold">User</th>
						<th class="py-3 px-4 font-bold">Details</th>
						<th class="py-3 px-4 font-bold">IP</th>
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
		catColor := "bg-cyan-300 text-black border-black"
		switch p.CategorySlug {
		case "ai-agent":
			catColor = "bg-purple-300 text-black border-black"
		case "security":
			catColor = "bg-rose-300 text-black border-black"
		case "kernel-embedded":
			catColor = "bg-orange-300 text-black border-black"
		case "devops":
			catColor = "bg-blue-300 text-black border-black"
		case "tools":
			catColor = "bg-emerald-300 text-black border-black"
		}

		rows += `<tr class="border-b-2 border-black">
			<td class="py-3 px-4 text-sm font-bold text-black max-w-[300px] truncate">` + p.Title + `</td>
			<td class="py-3 px-4"><span class="text-xs px-2 py-0.5 font-bold border-2 border-black font-mono ` + catColor + `">` + p.CategorySlug + `</span></td>
			<td class="py-3 px-4 text-sm font-mono font-bold text-neutral-500">` + p.CreatedAt.Format("2006-01-02") + `</td>
			<td class="py-3 px-4 text-right">
				<a href="/admin/posts/delete/` + itoa64(p.ID) + `" class="text-xs font-bold text-neutral-500 hover:bg-rose-200 px-2 py-1 border-2 border-black no-underline" onclick="return confirm('Hapus post ini?')">✕</a>
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
				<a href="/admin" class="text-xs font-bold no-underline hover:bg-yellow-200">&larr; Dashboard</a>
				<h1 class="text-xl font-black uppercase text-black mt-1">Manage Posts</h1>
			</div>
			<span class="text-xs font-bold bg-white border-2 border-black px-2.5 py-1 font-mono">` + itoa(len(posts)) + ` total</span>
		</div>

		<div class="bg-white border-2 border-black shadow-[4px_4px_0px_0px_#000] overflow-x-auto">
			<table class="w-full text-left">
				<thead>
					<tr class="border-b-2 border-black text-xs uppercase bg-black text-white">
						<th class="py-3 px-4 font-bold">Title</th>
						<th class="py-3 px-4 font-bold">Category</th>
						<th class="py-3 px-4 font-bold">Date</th>
						<th class="py-3 px-4"></th>
					</tr>
				</thead>
				<tbody>
					` + rows + `
				</tbody>
			</table>
		</div>

		<div class="bg-white border-2 border-black p-6 shadow-[4px_4px_0px_0px_#000]">
			<h2 class="text-sm font-black uppercase text-black mb-4">Create New Post</h2>
			<form method="POST" action="/admin/posts" class="space-y-4">
				<div>
					<label class="block text-sm font-bold text-neutral-600 mb-1">Title</label>
					<input type="text" name="title" required class="w-full px-3 py-2 border-2 border-black bg-white text-sm font-bold text-black placeholder:text-neutral-400 focus:outline-none focus:ring-2 focus:ring-black" placeholder="Post title">
				</div>
				<div>
					<label class="block text-sm font-bold text-neutral-600 mb-1">Category</label>
					<select name="category_slug" required class="w-full px-3 py-2 border-2 border-black bg-white text-sm font-bold text-black focus:outline-none focus:ring-2 focus:ring-black">
						<option value="">Select category</option>` + catOptions + `
					</select>
				</div>
				<div>
					<label class="block text-sm font-bold text-neutral-600 mb-1">Content (Markdown)</label>
					<textarea name="content" rows="15" required class="w-full px-3 py-2 border-2 border-black bg-white text-sm font-bold text-black font-mono placeholder:text-neutral-400 focus:outline-none focus:ring-2 focus:ring-black" placeholder="Write in Markdown..."></textarea>
				</div>
				<div class="flex justify-end">
					<button type="submit" class="px-5 py-2 border-2 border-black bg-emerald-400 text-black font-bold text-sm shadow-[3px_3px_0px_0px_#000] hover:shadow-[1px_1px_0px_0px_#000] hover:translate-x-[2px] hover:translate-y-[2px] transition-all cursor-pointer">Publish Post</button>
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

		statusDot := "bg-rose-400"
		statusPing := "bg-rose-400"
		statusText := "Offline"
		statusBadgeColor := "text-black"
		if online {
			statusDot = "bg-emerald-400"
			statusPing = "bg-emerald-400"
			statusText = "Online"
			statusBadgeColor = "text-black"
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
					tagBadges += `<span class="text-xs px-2 py-0.5 font-bold border-2 border-black bg-neutral-200 text-black font-mono">` + t + `</span> `
				}
			}
		}

		inlineStatus := statusText
		cards += `<div class="bg-white border-2 border-black p-5 shadow-[4px_4px_0px_0px_#000] hover:shadow-[2px_2px_0px_0px_#000] hover:translate-x-[2px] hover:translate-y-[2px] transition-all">
			<div class="flex items-start justify-between mb-4">
				<div class="flex items-center gap-3">
					<div class="w-10 h-10 border-2 border-black bg-black flex items-center justify-center">
						<span class="text-sm font-black text-white font-mono">` + s.Name[:min(len(s.Name), 2)] + `</span>
					</div>
					<div>
						<p class="text-sm font-black text-black">` + s.Name + `</p>
						<p class="text-xs font-mono font-bold text-neutral-500">` + s.Host + `:` + itoa(s.Port) + `</p>
					</div>
				</div>
				<div class="flex items-center gap-2">
					<span class="relative flex h-3 w-3">
						<span class="animate-ping absolute inline-flex h-full w-full rounded-none ` + statusPing + ` opacity-75"></span>
						<span class="relative inline-flex rounded-none h-3 w-3 border-2 border-black ` + statusDot + `"></span>
					</span>
					<span class="text-xs font-bold ` + statusBadgeColor + `">` + inlineStatus + `</span>
				</div>
			</div>
			<div class="flex items-center justify-between">
				<div class="flex flex-wrap gap-1">
					` + tagBadges + `
				</div>
				<div class="flex items-center gap-3 text-xs font-mono font-bold text-neutral-500">
					<span>Last seen: ` + lastSeenStr + `</span>
					<a href="/admin/servers/delete/` + itoa64(s.ID) + `" class="font-bold text-neutral-500 hover:bg-rose-200 px-2 py-0.5 border-2 border-black no-underline" onclick="return confirm('Hapus server ` + s.Name + `?')">✕</a>
				</div>
			</div>
		</div>`
	}

	if cards == "" {
		cards = `<div class="bg-white border-2 border-black p-8 text-center shadow-[4px_4px_0px_0px_#000]">
			<p class="text-neutral-500 font-bold text-sm">No servers registered.</p>
		</div>`
	}

	content := `<div class="space-y-6">

		<div class="flex items-center justify-between">
			<div>
				<a href="/admin" class="text-xs font-bold no-underline hover:bg-yellow-200">&larr; Dashboard</a>
				<h1 class="text-xl font-black uppercase text-black mt-1">Manage Servers</h1>
			</div>
			<span class="text-xs font-bold bg-white border-2 border-black px-2.5 py-1 font-mono">` + itoa(len(servers)) + ` nodes</span>
		</div>

		<div class="space-y-4">
			` + cards + `
		</div>

		<!-- Register Server -->
		<div class="bg-white border-2 border-black p-6 shadow-[4px_4px_0px_0px_#000]">
			<h2 class="text-sm font-black uppercase text-black mb-4">Register New Server</h2>
			<form method="POST" action="/admin/servers" class="grid grid-cols-1 md:grid-cols-2 gap-3">
				<div>
					<input type="text" name="name" placeholder="Server name" required class="w-full px-3 py-2 border-2 border-black bg-white text-sm font-bold placeholder:text-neutral-400 focus:outline-none focus:ring-2 focus:ring-black">
				</div>
				<div>
					<input type="text" name="host" placeholder="Host or IP" required class="w-full px-3 py-2 border-2 border-black bg-white text-sm font-bold placeholder:text-neutral-400 focus:outline-none focus:ring-2 focus:ring-black">
				</div>
				<div>
					<input type="number" name="port" placeholder="Port" value="9100" required class="w-full px-3 py-2 border-2 border-black bg-white text-sm font-bold placeholder:text-neutral-400 focus:outline-none focus:ring-2 focus:ring-black">
				</div>
				<div>
					<input type="text" name="tags" placeholder="Tags (comma separated)" class="w-full px-3 py-2 border-2 border-black bg-white text-sm font-bold placeholder:text-neutral-400 focus:outline-none focus:ring-2 focus:ring-black">
				</div>
				<div class="md:col-span-2 flex justify-end">
					<button type="submit" class="px-5 py-2 border-2 border-black bg-emerald-400 text-black font-bold text-sm shadow-[3px_3px_0px_0px_#000] hover:shadow-[1px_1px_0px_0px_#000] hover:translate-x-[2px] hover:translate-y-[2px] transition-all cursor-pointer">Register Server</button>
				</div>
			</form>
		</div>

	</div>`

	html := adminLayout("Manage Servers", content)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(html))
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

	h.svc.Repos.Audit.Insert(entry.username, "CREATE_SERVER", "created server: "+name+" ("+host+":"+portStr+")", r.RemoteAddr)
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
		return `<span class="text-xs px-1.5 py-0.5 font-bold border-2 border-black bg-rose-300 text-black">owner</span>`
	case model.RoleAdmin:
		return `<span class="text-xs px-1.5 py-0.5 font-bold border-2 border-black bg-amber-300 text-black">admin</span>`
	case model.RoleOperator:
		return `<span class="text-xs px-1.5 py-0.5 font-bold border-2 border-black bg-blue-300 text-black">operator</span>`
	case model.RoleReadonly:
		return `<span class="text-xs px-1.5 py-0.5 font-bold border-2 border-black bg-neutral-300 text-black">readonly</span>`
	default:
		return `<span class="text-xs px-1.5 py-0.5 font-bold border-2 border-black bg-neutral-300 text-black">` + role + `</span>`
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
	<link href="https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700;900&family=JetBrains+Mono:wght@400;500;700&display=swap" rel="stylesheet">
	<link rel="stylesheet" href="/static/css/style.css">
	<link rel="icon" type="image/svg+xml" href="/static/favicon.svg">
</head>
<body class="bg-neutral-100 text-black font-sans min-h-screen">
	<div class="min-h-screen flex">
		<!-- Sidebar (desktop) -->
		<aside class="w-64 bg-black text-white hidden md:flex flex-col border-r-2 border-black">
			<div class="p-6 flex items-center gap-3 border-b-2 border-white/20">
				<div class="w-8 h-8 bg-emerald-400 flex items-center justify-center border-2 border-white">
					<span class="text-black font-black">R</span>
				</div>
				<span class="text-xl font-black tracking-tight text-white">Raevtar</span>
			</div>
			<nav class="flex-1 px-4 space-y-2 mt-4">
				<a href="/admin" class="flex items-center gap-3 px-3 py-2.5 text-white hover:bg-white/10 font-bold transition-colors">Dashboard</a>
				<a href="/admin/posts" class="flex items-center gap-3 px-3 py-2.5 text-neutral-400 hover:text-white hover:bg-white/10 font-bold transition-colors">Posts</a>
				<a href="/admin/servers" class="flex items-center gap-3 px-3 py-2.5 text-neutral-400 hover:text-white hover:bg-white/10 font-bold transition-colors">Servers</a>
				<a href="/admin/manage-users" class="flex items-center gap-3 px-3 py-2.5 text-neutral-400 hover:text-white hover:bg-white/10 font-bold transition-colors">Users</a>
				<a href="/admin/audit-log" class="flex items-center gap-3 px-3 py-2.5 text-neutral-400 hover:text-white hover:bg-white/10 font-bold transition-colors">Audit</a>
			</nav>
			<div class="p-4 border-t-2 border-white/20">
				<a href="/admin/logout" class="block text-center px-3 py-2 border-2 border-white text-white font-bold hover:bg-white hover:text-black transition-colors">Logout</a>
			</div>
		</aside>

		<!-- Main -->
		<main class="flex-1 flex flex-col h-screen overflow-hidden">
			<header class="h-16 border-b-2 border-black flex items-center justify-between px-4 md:px-6 bg-black text-white sticky top-0 z-10">
				<div class="flex items-center gap-3">
					<div class="md:hidden w-7 h-7 bg-emerald-400 flex items-center justify-center border-2 border-white">
						<span class="text-black font-black text-sm">R</span>
					</div>
					<div class="text-base md:text-lg font-black text-white">` + title + `</div>
				</div>
				<div class="flex items-center gap-3">
					<div class="flex items-center gap-2 px-3 py-1.5 bg-white border-2 border-black text-xs md:text-sm">
						<span class="relative flex h-3 w-3">
							<span class="animate-ping absolute inline-flex h-full w-full bg-emerald-400 opacity-75"></span>
							<span class="relative inline-flex h-3 w-3 border-2 border-black bg-emerald-400"></span>
						</span>
						<span class="text-black font-bold">Connected</span>
					</div>
				</div>
			</header>

			<div class="p-4 md:p-8 space-y-6 flex-1 overflow-y-auto bg-neutral-100">
				` + content + `
			</div>
		</main>
	</div>
</body>
</html>`
}
