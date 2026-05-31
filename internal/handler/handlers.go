package handler

import (
	"net/http"
	"strconv"
	"strings"

	"raevtar/internal/model"
)

func (h *Handler) landingIndex(w http.ResponseWriter, r *http.Request) {
	posts, _, _ := h.svc.Blog.ListPosts("", 1, 3)
	servers, _ := h.svc.Monitor.ListServers()
	categories, _ := h.svc.Repos.Category.List()

	render(w, r, "pages/index.templ", map[string]any{
		"Title":      "Raevtar",
		"Posts":      posts,
		"Servers":    servers,
		"Categories": categories,
		"Domain":     h.cfg.Domain,
	})
}

func (h *Handler) blogList(w http.ResponseWriter, r *http.Request) {
	cat := r.URL.Query().Get("category")
	pageStr := r.URL.Query().Get("page")
	page, _ := strconv.Atoi(pageStr)
	if page < 1 {
		page = 1
	}

	posts, total, err := h.svc.Blog.ListPosts(cat, page, 10)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	categories, _ := h.svc.Repos.Category.List()
	totalPages := (total + 9) / 10

	render(w, r, "pages/blog_list.templ", map[string]any{
		"Title":       "Blog — Raevtar",
		"Posts":       posts,
		"Categories":  categories,
		"CurrentCat":  cat,
		"Page":        page,
		"TotalPages":  totalPages,
		"CurrentPage": page,
	})
}

func (h *Handler) blogDetail(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	post, err := h.svc.Blog.GetPost(slug)
	if err != nil {
		http.Error(w, "Post not found", http.StatusNotFound)
		return
	}

	categories, _ := h.svc.Repos.Category.List()

	render(w, r, "pages/blog_post.templ", map[string]any{
		"Title":      post.Title + " — Raevtar",
		"Post":       post,
		"Categories": categories,
	})
}

func (h *Handler) dashboardIndex(w http.ResponseWriter, r *http.Request) {
	servers, err := h.svc.Monitor.ListServers()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	categories, _ := h.svc.Repos.Category.List()

	render(w, r, "pages/dashboard.templ", map[string]any{
		"Title":      "Dashboard — Raevtar",
		"Servers":    servers,
		"Categories": categories,
	})
}

func (h *Handler) dashboardDetail(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("serverID")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid server ID", http.StatusBadRequest)
		return
	}

	server, err := h.svc.Monitor.GetServer(id)
	if err != nil {
		http.Error(w, "Server not found", http.StatusNotFound)
		return
	}

	metrics, _ := h.svc.Repos.Metric.GetByServerID(id, 50)
	categories, _ := h.svc.Repos.Category.List()

	render(w, r, "pages/server_detail.templ", map[string]any{
		"Title":      server.Name + " — Raevtar",
		"Server":     server,
		"Metrics":    metrics,
		"Categories": categories,
	})
}

// --- HTML renderers ---

func render(w http.ResponseWriter, r *http.Request, tpl string, data map[string]any) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	nav := navBar(data["Categories"], r.URL.Path)
	footer := `<footer class="mt-12 py-8 border-t border-slate-700/50">
		<div class="flex flex-wrap items-center justify-center gap-4 mb-4">
			<span class="text-xs px-2.5 py-1 rounded-md font-mono border bg-cyan-500/10 text-cyan-400 border-cyan-500/20">Go</span>
			<span class="text-xs px-2.5 py-1 rounded-md font-mono border bg-sky-500/10 text-sky-400 border-sky-500/20">Templ</span>
			<span class="text-xs px-2.5 py-1 rounded-md font-mono border bg-blue-500/10 text-blue-400 border-blue-500/20">HTMX</span>
			<span class="text-xs px-2.5 py-1 rounded-md font-mono border bg-indigo-500/10 text-indigo-400 border-indigo-500/20">SQLite</span>
			<span class="text-xs px-2.5 py-1 rounded-md font-mono border bg-orange-500/10 text-orange-400 border-orange-500/20">Cloudflare</span>
		</div>
		<p class="text-center text-sm text-slate-500"><a href="https://raevtar.tech" class="text-green-400 hover:text-green-300">raevtar.tech</a> &middot; single binary on postmarketOS (whyred)</p>
	</footer>`

	var body string
	switch tpl {
	case "pages/index.templ":
		body = landingPageHTML(data)
	case "pages/blog_list.templ":
		body = blogListHTML(data)
	case "pages/blog_post.templ":
		body = blogPostHTML(data)
	case "pages/dashboard.templ":
		body = dashboardHTML(data)
	case "pages/server_detail.templ":
		body = serverDetailHTML(data)
	default:
		body = page404HTML(data)
	}

	full := pageHTML(data["Title"].(string), nav+body+footer)
	w.Write([]byte(full))
}

func pageHTML(title, content string) string {
	return `<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<title>` + title + `</title>
	<link rel="preconnect" href="https://fonts.googleapis.com">
	<link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
	<link href="https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700&family=JetBrains+Mono:wght@400;500&display=swap" rel="stylesheet">
	<link rel="stylesheet" href="/static/css/style.css">
	<link rel="icon" type="image/svg+xml" href="/static/favicon.svg">
	<link rel="alternate" type="application/rss+xml" title="Raevtar" href="https://raevtar.tech/blog/feed.xml">
	<script src="https://unpkg.com/htmx.org@2.0.4"></script>
	<meta name="description" content="Raevtar — Personal blog, server dashboard, and automation platform running on postmarketOS">
	<meta name="keywords" content="raevtar, blog, server monitoring, postmarketOS, linux">
	<meta name="author" content="latifangren">
	<meta property="og:type" content="website">
	<meta property="og:site_name" content="Raevtar">
	<meta property="og:title" content="` + title + `">
	<meta property="og:description" content="Personal platform for project notes, server monitoring, and automation.">
	<meta property="og:url" content="https://raevtar.tech">
	<meta name="twitter:card" content="summary">
	<meta name="twitter:title" content="Raevtar">
	<meta name="twitter:description" content="Personal platform for project notes, server monitoring, and automation.">
</head>
<body class="min-h-screen bg-slate-900 text-slate-100 antialiased">
	<div class="max-w-5xl mx-auto px-4 pt-4">
	` + content + `
	</div>
</body>
</html>`
}

func navBar(cats any, currentPath string) string {
	activeClass := "text-sm text-white font-medium no-underline"
	inactiveClass := "text-sm text-slate-400 hover:text-white transition-colors no-underline"
	active := func(path string) string {
		if currentPath == path {
			return activeClass
		}
		return inactiveClass
	}
	return `<nav class="sticky top-0 z-50 flex items-center gap-4 sm:gap-6 py-3 mb-8 border-b border-slate-700/50 bg-slate-900/80 backdrop-blur-sm">
		<a href="/" class="text-lg sm:text-xl font-bold text-green-400 hover:text-green-300 no-underline">raevtar</a>
		<a href="/blog" class="` + active("/blog") + `">blog</a>
		<a href="/dashboard" class="` + active("/dashboard") + `">dashboard</a>
	</nav>`
}

func landingPageHTML(data map[string]any) string {
	posts, _ := data["Posts"].([]model.Post)
	cats, _ := data["Categories"].([]model.Category)

	h := ``

	// ===== HERO — full-width breakout =====
	h += `<section class="full-width bg-gradient-to-b from-green-900/10 to-transparent border-b border-slate-700/30 mb-12">`
	h += `<div class="max-w-5xl mx-auto px-4 py-20 md:py-28 text-center">`
	h += `<h1 class="text-5xl md:text-6xl font-bold mb-4 tracking-tight text-white">raevtar</h1>`
	h += `<p class="text-lg text-slate-300 mb-2 max-w-xl mx-auto">Personal platform for project notes, server monitoring, and automation.</p>`
	h += `<p class="text-sm text-slate-500 mb-8 max-w-lg mx-auto">Single Go binary running on postmarketOS (whyred).</p>`
	h += `<div class="flex flex-wrap gap-3 justify-center">`
	h += `<a href="/blog" class="inline-flex items-center gap-1.5 px-5 py-2.5 rounded-lg bg-green-500 text-white text-sm font-medium hover:bg-green-400 transition-colors no-underline">Read the Blog &rarr;</a>`
	h += `<a href="/dashboard" class="inline-flex items-center gap-1.5 px-5 py-2.5 rounded-lg bg-slate-800 text-slate-300 text-sm font-medium border border-slate-700/30 hover:border-slate-600/50 hover:text-white transition-colors no-underline">Open Dashboard</a>`
	h += `</div>`
	h += `</div>`
	h += `</section>`

	// ===== STATS BAR (only if meaningful data exists) =====
	totalPosts := len(posts)
	totalCats := len(cats)
	if totalPosts > 0 || totalCats > 0 {
		h += `<section class="flex flex-wrap gap-6 justify-center mb-12 py-4 px-6 bg-slate-800/30 rounded-xl border border-slate-700/30">`
		if totalPosts > 0 {
			h += `<div class="text-center"><span class="text-2xl font-bold text-white">` + itoa(totalPosts) + `</span><span class="block text-xs text-slate-500 mt-0.5">posts</span></div>`
		}
		if totalCats > 0 {
			h += `<div class="text-center"><span class="text-2xl font-bold text-white">` + itoa(totalCats) + `</span><span class="block text-xs text-slate-500 mt-0.5">categories</span></div>`
		}
		h += `<div class="text-center"><span class="text-2xl font-bold text-green-400">1</span><span class="block text-xs text-slate-500 mt-0.5">binary</span></div>`
		h += `</section>`
	}

	// ===== LATEST POSTS =====
	if len(posts) > 0 {
		h += `<section class="mb-12">`
		h += `<div class="flex items-center justify-between mb-4">`
		h += `<h2 class="text-lg font-semibold">Latest Posts</h2>`
		h += `<a href="/blog" class="text-sm text-green-400 hover:text-green-300 no-underline">View all &rarr;</a>`
		h += `</div>`
		for _, p := range posts {
			h += postCardHTML(p)
		}
		h += `</section>`
	}

	// ===== EMPTY STATE =====
	if len(posts) == 0 {
		h += `<section class="mb-12 text-center py-12">`
		h += `<p class="text-slate-500">No posts yet.</p>`
		h += `<p class="text-sm text-slate-600 mt-2">Check the <a href="/dashboard" class="text-green-400 hover:text-green-300">dashboard</a> or start publishing via the API.</p>`
		h += `</section>`
	}

	return h
}

func postCardHTML(p model.Post) string {
	tagHTML := tagsInline(p.Tags)
	return `<article class="bg-slate-800/50 rounded-lg p-4 mb-3 border border-slate-700/30 hover:border-green-500/30 transition-colors">
		<div class="flex items-center gap-2 mb-2">
			<span class="text-xs px-2 py-0.5 rounded bg-slate-700 text-slate-400">` + p.CategoryName + `</span>
			<span class="text-xs text-slate-500">` + p.CreatedAt.Format("Jan 2, 2006") + `</span>
			` + tagHTML + `
		</div>
		<a href="/blog/` + p.Slug + `" class="text-lg font-semibold text-white hover:text-green-400 transition-colors no-underline">` + p.Title + `</a>
		<p class="text-sm text-slate-400 mt-1">` + p.Excerpt + `</p>
	</article>`
}

func serverCardHTML(s model.Server) string {
	statusDot := "bg-slate-500"
	statusText := "Unknown"
	if s.LastSeen != nil {
		statusDot = "bg-green-500"
		statusText = "Online"
	}

	return `<a href="/dashboard/` + itoa64(s.ID) + `" class="block bg-slate-800/50 rounded-lg p-4 mb-3 border border-slate-700/30 hover:border-green-500/30 transition-colors no-underline">
		<div class="flex items-center justify-between">
			<div>
				<h3 class="font-semibold text-white">` + s.Name + `</h3>
				<p class="text-sm text-slate-400">` + s.Host + `:` + itoa(s.Port) + `</p>
			</div>
			<div class="flex items-center gap-1.5 text-sm">
				<span class="inline-block w-2 h-2 rounded-full ` + statusDot + `"></span>
				<span class="text-slate-400">` + statusText + `</span>
			</div>
		</div>
	</a>`
}

func blogListHTML(data map[string]any) string {
	posts, _ := data["Posts"].([]model.Post)
	cats, _ := data["Categories"].([]model.Category)
	curCat, _ := data["CurrentCat"].(string)
	page, _ := data["Page"].(int)
	totalPages, _ := data["TotalPages"].(int)

	html := `<h1 class="text-2xl font-bold mb-6">Blog</h1>`

	html += `<div class="flex flex-wrap gap-2 mb-6">`
	active := "px-3 py-1 rounded-md bg-green-500/20 text-green-400 text-sm no-underline"
	inactive := "px-3 py-1 rounded-md bg-slate-800 text-slate-300 text-sm hover:bg-slate-700 no-underline"
	if curCat == "" {
		html += `<a href="/blog" class="` + active + `">Semua</a>`
	} else {
		html += `<a href="/blog" class="` + inactive + `">Semua</a>`
	}
	for _, c := range cats {
		if curCat == c.Slug {
			html += `<a href="/blog?category=` + c.Slug + `" class="` + active + `">` + c.Name + `</a>`
		} else {
			html += `<a href="/blog?category=` + c.Slug + `" class="` + inactive + `">` + c.Name + `</a>`
		}
	}
	html += `</div>`

	if len(posts) == 0 {
		html += `<p class="text-slate-500">Belum ada postingan.</p>`
	}
	for _, p := range posts {
		html += postCardHTML(p)
	}

	// pagination
	if totalPages > 1 {
		html += `<div class="flex gap-2 mt-6 justify-center">`
		q := ""
		if curCat != "" {
			q = "&category=" + curCat
		}
		for i := 1; i <= totalPages; i++ {
			if i == page {
				html += `<span class="px-3 py-1 rounded bg-green-500/20 text-green-400 text-sm">` + itoa(i) + `</span>`
			} else {
				html += `<a href="/blog?page=` + itoa(i) + q + `" class="px-3 py-1 rounded bg-slate-800 text-slate-300 text-sm hover:bg-slate-700 no-underline">` + itoa(i) + `</a>`
			}
		}
		html += `</div>`
	}

	return html
}

func blogPostHTML(data map[string]any) string {
	post, ok := data["Post"].(*model.Post)
	if !ok {
		return `<h1>404</h1><p>Post not found.</p>`
	}

	tagHTML := tagsInline(post.Tags)

	readMin := estimateReadTime(post.ContentMD)

	return `<article>
		<div class="mb-6">
			<div class="flex items-center gap-2 mb-2">
				<span class="text-xs px-2 py-0.5 rounded bg-slate-700 text-slate-400">` + post.CategoryName + `</span>
				<span class="text-xs text-slate-500">` + post.CreatedAt.Format("Jan 2, 2006") + `</span>
				<span class="text-xs text-slate-600">&middot; ` + readMin + ` min read</span>
				` + tagHTML + `
			</div>
			<h1 class="text-3xl font-bold">` + post.Title + `</h1>
		</div>
		<div class="prose text-slate-300">` + post.ContentHTML + `</div>
		<div class="mt-8 pt-4 border-t border-slate-700/50">
			<a href="/blog" class="text-sm text-green-400 hover:text-green-300 no-underline">&larr; Kembali ke blog</a>
		</div>
	</article>`
}

// tagsInline renders tag badges, special-casing "auto" tag
func tagsInline(tags []model.Tag) string {
	if len(tags) == 0 {
		return ""
	}
	var html string
	for _, t := range tags {
		switch t.Name {
		case "auto":
			html += `<span class="text-xs px-2 py-0.5 rounded bg-purple-700/50 text-purple-300 border border-purple-600/30">auto</span>`
		case "commissioned":
			html += `<span class="text-xs px-2 py-0.5 rounded bg-amber-700/50 text-amber-300 border border-amber-600/30">commissioned</span>`
		default:
			html += `<span class="text-xs px-2 py-0.5 rounded bg-slate-700 text-slate-400">` + t.Name + `</span>`
		}
	}
	return html
}

func dashboardHTML(data map[string]any) string {
	servers, _ := data["Servers"].([]model.Server)

	html := `<div class="flex items-center justify-between mb-6">
		<h1 class="text-2xl font-bold">Server Dashboard</h1>
		<button onclick="toggleForm()" class="text-xs px-3 py-1.5 rounded bg-green-600 hover:bg-green-500 text-white transition-colors">+ Register Server</button>
	</div>`

	// Register form (hidden by default)
	html += `<div id="register-form" class="hidden mb-6 bg-slate-800/50 rounded-lg p-4 border border-slate-700/30">
		<h2 class="text-sm font-semibold text-slate-300 mb-3">Register New Server</h2>
		<form hx-post="/api/v1/servers" hx-trigger="submit" hx-target="#server-list" hx-swap="outerHTML" hx-on::after-request="this.reset(); toggleForm()">
			<input type="hidden" name="_method" value="POST">
			<div class="flex flex-wrap gap-3">
				<input type="text" name="name" placeholder="Name (e.g. whyred)" required
					class="flex-1 min-w-[120px] px-3 py-2 rounded bg-slate-700 border border-slate-600 text-sm text-white placeholder-slate-400 focus:outline-none focus:border-green-500">
				<input type="text" name="host" placeholder="Host/IP" required
					class="flex-1 min-w-[120px] px-3 py-2 rounded bg-slate-700 border border-slate-600 text-sm text-white placeholder-slate-400 focus:outline-none focus:border-green-500">
				<input type="number" name="port" placeholder="Port" value="22"
					class="w-20 px-3 py-2 rounded bg-slate-700 border border-slate-600 text-sm text-white placeholder-slate-400 focus:outline-none focus:border-green-500">
				<input type="text" name="tags" placeholder="Tags (comma)"
					class="flex-1 min-w-[100px] px-3 py-2 rounded bg-slate-700 border border-slate-600 text-sm text-white placeholder-slate-400 focus:outline-none focus:border-green-500">
				<button type="submit" class="px-4 py-2 rounded bg-green-600 hover:bg-green-500 text-white text-sm transition-colors">Add</button>
			</div>
		</form>
	</div>`

	html += `<div id="server-list">`

	if len(servers) == 0 {
		html += `<div class="bg-slate-800/50 rounded-lg p-8 text-center border border-slate-700/30">
			<p class="text-slate-400 mb-2">Belum ada server terdaftar.</p>
			<p class="text-sm text-slate-500">Klik "Register Server" di atas untuk daftarin server pertama.</p>
		</div>`
	}

	for _, s := range servers {
		statusDot := "bg-slate-500"
		statusText := "Unknown"
		if s.LastSeen != nil {
			statusDot = "bg-green-500"
			statusText = "Online — " + s.LastSeen.Format("Jan 2 15:04")
		}
		html += `<a href="/dashboard/` + itoa64(s.ID) + `" class="block bg-slate-800/50 rounded-lg p-4 mb-3 border border-slate-700/30 hover:border-green-500/30 transition-colors no-underline">
			<div class="flex items-center justify-between">
				<div>
					<h3 class="font-semibold text-white">` + s.Name + `</h3>
					<p class="text-sm text-slate-400">` + s.Host + `:` + itoa(s.Port) + `</p>
				</div>
				<div class="flex items-center gap-1.5 text-sm">
					<span class="inline-block w-2 h-2 rounded-full ` + statusDot + `"></span>
					<span class="text-slate-400">` + statusText + `</span>
				</div>
			</div>
		</a>`
	}

	html += `</div>`

	// HTMX auto-refresh + JS toggle
	html += `<div class="mt-4 text-xs text-slate-500" hx-get="/dashboard" hx-trigger="every 30s" hx-select="#server-list" hx-swap="outerHTML" hx-target="#server-list">Auto-refresh every 30s</div>
	<script>
	function toggleForm() {
		var f = document.getElementById('register-form');
		f.classList.toggle('hidden');
	}
	</script>`

	return html
}

func serverDetailHTML(data map[string]any) string {
	server, ok := data["Server"].(*model.Server)
	if !ok {
		return `<h1>404</h1><p>Server not found.</p>`
	}
	metrics, _ := data["Metrics"].([]model.ServerMetric)

	html := `<a href="/dashboard" class="text-sm text-green-400 hover:text-green-300 no-underline">&larr; Kembali</a>
	<h1 class="text-2xl font-bold mt-2 mb-1">` + server.Name + `</h1>
	<p class="text-sm text-slate-400 mb-6">` + server.Host + `:` + itoa(server.Port) + `</p>`

	if len(metrics) == 0 {
		html += `<div class="bg-slate-800/50 rounded-lg p-6 border border-slate-700/30">
			<p class="text-slate-400 text-sm">Belum ada data metrics. Kirim data lewat POST /api/v1/servers/:id/ping</p>
		</div>`
	} else {
		html += `<div class="bg-slate-800/50 rounded-lg p-4 border border-slate-700/30">
			<h2 class="text-sm font-semibold uppercase tracking-wider text-slate-400 mb-3">Metrics Terakhir</h2>`
		for _, m := range metrics {
			statusDot := "bg-red-500"
			if m.Online {
				statusDot = "bg-green-500"
			}
			html += `<div class="flex items-center justify-between py-2 border-b border-slate-700/30 last:border-0">
				<div class="flex items-center gap-2">
					<span class="inline-block w-2 h-2 rounded-full ` + statusDot + `"></span>
					<span class="text-sm text-slate-300">` + m.RecordedAt.Format("Jan 2 15:04") + `</span>
				</div>
				<div class="flex gap-4 text-sm text-slate-400">
					<span>CPU: <span class="text-white">` + ftoa(m.CPUPercent) + `%</span></span>
					<span>RAM: <span class="text-white">` + ftoa(m.RAMUsedMB) + `/` + ftoa(m.RAMTotalMB) + ` MB</span></span>
				</div>
			</div>`
		}
		html += `</div>`
	}

	return html
}

// helpers
func itoa(n int) string    { return strconv.Itoa(n) }
func itoa64(n int64) string { return strconv.FormatInt(n, 10) }
func ftoa(f float64) string { return strconv.FormatFloat(f, 'f', 1, 64) }

// estimateReadTime returns "X min read" for a markdown string.
func estimateReadTime(md string) string {
	// Count words: split on whitespace
	if len(md) == 0 {
		return "<1"
	}
	words := len(strings.Fields(md))
	if words == 0 {
		return "<1"
	}
	// Average reading speed: 200 words per minute
	min := words / 200
	if min < 1 {
		return "<1"
	}
	return strconv.Itoa(min)
}

func page404HTML(data map[string]any) string {
	_ = data // unused, just for interface
	return `<div class="flex flex-col items-center justify-center py-24 text-center">
		<h1 class="text-9xl font-bold text-slate-800">404</h1>
		<h2 class="text-2xl font-semibold text-slate-400 mt-4 mb-2">Halaman gak ketemu</h2>
		<p class="text-slate-500 mb-8 max-w-md">Entah udah dipindah, dihapus, atau emang gak pernah ada.</p>
		<a href="/" class="inline-flex items-center gap-1.5 px-5 py-2.5 rounded-lg bg-green-500 text-white text-sm font-medium hover:bg-green-400 transition-colors no-underline">
			&larr; Balik ke Beranda
		</a>
	</div>`
}

// serveStatic serves a file from the ./static directory
func (h *Handler) serveStatic(filename string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/"+filename)
	}
}

// page404Handler renders the custom 404 page for unmatched routes
func page404Handler(w http.ResponseWriter, r *http.Request) {
	render(w, r, "", map[string]any{
		"Title": "404 — Raevtar",
	})
}
