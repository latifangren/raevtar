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
	footer := `<footer class="mt-12 py-8 border-t-2 border-black">
		<div class="flex flex-wrap items-center justify-center gap-4 mb-4">
			<span class="text-xs px-2.5 py-1 font-bold border-2 border-black bg-white">Go</span>
			<span class="text-xs px-2.5 py-1 font-bold border-2 border-black bg-white">HTMX</span>
			<span class="text-xs px-2.5 py-1 font-bold border-2 border-black bg-white">SQLite</span>
			<span class="text-xs px-2.5 py-1 font-bold border-2 border-black bg-white">Cloudflare</span>
		</div>
		<p class="text-center text-sm text-neutral-600"><a href="https://raevtar.tech" class="font-bold no-underline hover:bg-yellow-200">raevtar.tech</a> &middot; single binary on postmarketOS (whyred)</p>
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
	<link href="https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700;900&family=JetBrains+Mono:wght@400;500;700&display=swap" rel="stylesheet">
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
<body class="min-h-screen bg-neutral-100 text-black antialiased">
	<div class="max-w-5xl mx-auto px-4 pt-4">
	` + content + `
	</div>
</body>
</html>`
}

func navBar(cats any, currentPath string) string {
	activeClass := "text-sm font-bold text-white no-underline"
	inactiveClass := "text-sm font-medium text-neutral-400 hover:text-white no-underline transition-colors"
	active := func(path string) string {
		if currentPath == path {
			return activeClass
		}
		return inactiveClass
	}
	return `<nav class="sticky top-0 z-50 flex items-center gap-4 sm:gap-6 py-3 px-4 sm:px-6 mb-8 bg-black border-b-2 border-black shadow-[0_4px_0px_0px_#000] max-w-5xl mx-auto" style="margin-left:0;margin-right:0">
		<a href="/" class="text-lg sm:text-xl font-black text-emerald-400 hover:text-emerald-300 no-underline">raevtar</a>
		<a href="/blog" class="` + active("/blog") + `">blog</a>
		<a href="/dashboard" class="` + active("/dashboard") + `">dashboard</a>
	</nav>`
}

func landingPageHTML(data map[string]any) string {
	posts, _ := data["Posts"].([]model.Post)
	cats, _ := data["Categories"].([]model.Category)

	h := ``

	// ===== HERO — full-width breakout =====
	h += `<section class="full-width bg-emerald-400 border-b-2 border-black mb-12">`
	h += `<div class="max-w-5xl mx-auto px-4 py-24 md:py-32 text-center">`
	h += `<h1 class="text-6xl md:text-7xl font-black mb-4 tracking-tight text-black uppercase">raevtar</h1>`
	h += `<p class="text-lg text-black font-bold mb-2 max-w-xl mx-auto opacity-80">Personal platform for project notes, server monitoring, and automation.</p>`
	h += `<p class="text-sm text-black font-medium mb-10 opacity-60 max-w-lg mx-auto">Single Go binary running on postmarketOS (whyred).</p>`
	h += `<div class="flex flex-wrap gap-4 justify-center">`
	h += `<a href="/blog" class="inline-flex items-center gap-1.5 px-6 py-3 border-2 border-black bg-white text-black font-bold text-sm no-underline shadow-[4px_4px_0px_0px_#000] hover:shadow-[2px_2px_0px_0px_#000] hover:translate-x-[2px] hover:translate-y-[2px] transition-all">Read the Blog &rarr;</a>`
	h += `<a href="/dashboard" class="inline-flex items-center gap-1.5 px-6 py-3 border-2 border-black bg-black text-white font-bold text-sm no-underline shadow-[4px_4px_0px_0px_#000] hover:shadow-[2px_2px_0px_0px_#000] hover:translate-x-[2px] hover:translate-y-[2px] transition-all">Open Dashboard</a>`
	h += `</div>`
	h += `</div>`
	h += `</section>`

	// ===== STATS BAR =====
	totalPosts := len(posts)
	totalCats := len(cats)
	if totalPosts > 0 || totalCats > 0 {
		h += `<section class="flex flex-wrap gap-4 justify-center mb-12 py-4 px-6 bg-white border-2 border-black shadow-[4px_4px_0px_0px_#000]">`
		if totalPosts > 0 {
			h += `<div class="text-center px-4"><span class="text-3xl font-black text-black">` + itoa(totalPosts) + `</span><span class="block text-xs uppercase font-bold text-neutral-600 mt-0.5">posts</span></div>`
		}
		if totalCats > 0 {
			h += `<div class="text-center px-4"><span class="text-3xl font-black text-black">` + itoa(totalCats) + `</span><span class="block text-xs uppercase font-bold text-neutral-600 mt-0.5">categories</span></div>`
		}
		h += `<div class="text-center px-4"><span class="text-3xl font-black text-emerald-600">1</span><span class="block text-xs uppercase font-bold text-neutral-600 mt-0.5">binary</span></div>`
		h += `</section>`
	}

	// ===== LATEST POSTS =====
	if len(posts) > 0 {
		h += `<section class="mb-12">`
		h += `<div class="flex items-center justify-between mb-4">`
		h += `<h2 class="text-lg font-black uppercase">Latest Posts</h2>`
		h += `<a href="/blog" class="text-sm font-bold no-underline hover:bg-yellow-200">View all &rarr;</a>`
		h += `</div>`
		for _, p := range posts {
			h += postCardHTML(p)
		}
		h += `</section>`
	}

	// ===== EMPTY STATE =====
	if len(posts) == 0 {
		h += `<section class="mb-12 text-center py-12">`
		h += `<p class="text-neutral-500 font-bold">No posts yet.</p>`
		h += `<p class="text-sm text-neutral-600 mt-2">Check the <a href="/dashboard" class="font-bold no-underline hover:bg-yellow-200">dashboard</a> or start publishing via the API.</p>`
		h += `</section>`
	}

	return h
}

func postCardHTML(p model.Post) string {
	tagHTML := tagsInline(p.Tags)
	return `<article class="bg-white border-2 border-black p-5 mb-4 shadow-[4px_4px_0px_0px_#000] hover:shadow-[2px_2px_0px_0px_#000] hover:translate-x-[2px] hover:translate-y-[2px] transition-all">
		<div class="flex items-center gap-2 mb-2 flex-wrap">
			<span class="text-xs px-2 py-0.5 font-bold border-2 border-black bg-emerald-400">` + p.CategoryName + `</span>
			<span class="text-xs font-bold text-neutral-500">` + p.CreatedAt.Format("Jan 2, 2006") + `</span>
			` + tagHTML + `
		</div>
		<a href="/blog/` + p.Slug + `" class="text-lg font-bold text-black hover:bg-yellow-200 no-underline">` + p.Title + `</a>
		<p class="text-sm text-neutral-600 mt-1">` + p.Excerpt + `</p>
	</article>`
}

func serverCardHTML(s model.Server) string {
	statusDot := "bg-neutral-300"
	statusText := "Unknown"
	if s.LastSeen != nil {
		statusDot = "bg-emerald-400"
		statusText = "Online"
	}

	return `<a href="/dashboard/` + itoa64(s.ID) + `" class="block bg-white border-2 border-black p-5 mb-4 shadow-[4px_4px_0px_0px_#000] hover:shadow-[2px_2px_0px_0px_#000] hover:translate-x-[2px] hover:translate-y-[2px] transition-all no-underline">
		<div class="flex items-center justify-between">
			<div>
				<h3 class="font-bold text-black">` + s.Name + `</h3>
				<p class="text-sm font-mono text-neutral-600">` + s.Host + `:` + itoa(s.Port) + `</p>
			</div>
			<div class="flex items-center gap-1.5 text-sm">
				<span class="inline-block w-3 h-3 border-2 border-black ` + statusDot + `"></span>
				<span class="font-bold text-neutral-700">` + statusText + `</span>
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

	html := `<h1 class="text-3xl font-black uppercase mb-6">Blog</h1>`

	html += `<div class="flex flex-wrap gap-2 mb-6">`
	activeStyle := "px-3 py-1.5 border-2 border-black bg-black text-white font-bold text-sm no-underline shadow-[2px_2px_0px_0px_#000]"
	inactiveStyle := "px-3 py-1.5 border-2 border-black bg-white text-black font-bold text-sm no-underline hover:shadow-[2px_2px_0px_0px_#000] transition-all"
	if curCat == "" {
		html += `<a href="/blog" class="` + activeStyle + `">Semua</a>`
	} else {
		html += `<a href="/blog" class="` + inactiveStyle + `">Semua</a>`
	}
	for _, c := range cats {
		if curCat == c.Slug {
			html += `<a href="/blog?category=` + c.Slug + `" class="` + activeStyle + `">` + c.Name + `</a>`
		} else {
			html += `<a href="/blog?category=` + c.Slug + `" class="` + inactiveStyle + `">` + c.Name + `</a>`
		}
	}
	html += `</div>`

	if len(posts) == 0 {
		html += `<p class="text-neutral-500 font-bold">Belum ada postingan.</p>`
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
				html += `<span class="px-3 py-1.5 border-2 border-black bg-black text-white font-bold text-sm">` + itoa(i) + `</span>`
			} else {
				html += `<a href="/blog?page=` + itoa(i) + q + `" class="px-3 py-1.5 border-2 border-black bg-white text-black font-bold text-sm no-underline hover:shadow-[2px_2px_0px_0px_#000] transition-all">` + itoa(i) + `</a>`
			}
		}
		html += `</div>`
	}

	return html
}

func blogPostHTML(data map[string]any) string {
	post, ok := data["Post"].(*model.Post)
	if !ok {
		return `<h1 class="text-6xl font-black">404</h1><p class="font-bold">Post not found.</p>`
	}

	tagHTML := tagsInline(post.Tags)

	readMin := estimateReadTime(post.ContentMD)

	return `<article>
		<div class="mb-6 bg-white border-2 border-black p-5 shadow-[4px_4px_0px_0px_#000]">
			<div class="flex items-center gap-2 mb-2 flex-wrap">
				<span class="text-xs px-2 py-0.5 font-bold border-2 border-black bg-emerald-400">` + post.CategoryName + `</span>
				<span class="text-xs font-bold text-neutral-500">` + post.CreatedAt.Format("Jan 2, 2006") + `</span>
				<span class="text-xs font-bold text-neutral-500">&middot; ` + readMin + ` min read</span>
				` + tagHTML + `
			</div>
			<h1 class="text-3xl font-black uppercase">` + post.Title + `</h1>
		</div>
		<div class="prose text-black">` + post.ContentHTML + `</div>
		<div class="mt-8 pt-4 border-t-2 border-black">
			<a href="/blog" class="font-bold no-underline hover:bg-yellow-200">&larr; Kembali ke blog</a>
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
			html += `<span class="text-xs px-2 py-0.5 font-bold border-2 border-black bg-violet-300 text-black">auto</span>`
		case "commissioned":
			html += `<span class="text-xs px-2 py-0.5 font-bold border-2 border-black bg-amber-300 text-black">commissioned</span>`
		default:
			html += `<span class="text-xs px-2 py-0.5 font-bold border-2 border-black bg-neutral-200 text-black">` + t.Name + `</span>`
		}
	}
	return html
}

func dashboardHTML(data map[string]any) string {
	servers, _ := data["Servers"].([]model.Server)

	html := `<div class="flex items-center justify-between mb-6">
		<h1 class="text-3xl font-black uppercase">Server Dashboard</h1>
		<button onclick="toggleForm()" class="px-4 py-2 border-2 border-black bg-black text-white font-bold text-sm shadow-[3px_3px_0px_0px_#000] hover:shadow-[1px_1px_0px_0px_#000] hover:translate-x-[2px] hover:translate-y-[2px] transition-all cursor-pointer">+ Register Server</button>
	</div>`

	// Register form (hidden by default)
	html += `<div id="register-form" class="hidden mb-6 bg-white border-2 border-black p-5 shadow-[4px_4px_0px_0px_#000]">
		<h2 class="text-sm font-black uppercase mb-3">Register New Server</h2>
		<form hx-post="/api/v1/servers" hx-trigger="submit" hx-target="#server-list" hx-swap="outerHTML" hx-on::after-request="if(event.detail.successful){this.reset();document.getElementById('register-form').classList.add('hidden')}" class="space-y-3">
			<div class="grid grid-cols-1 sm:grid-cols-3 gap-3">
				<input type="text" name="name" placeholder="Server name" required class="w-full px-3 py-2 border-2 border-black bg-white text-sm font-bold placeholder:text-neutral-400 focus:outline-none focus:ring-2 focus:ring-black">
				<input type="text" name="host" placeholder="host or IP" required class="w-full px-3 py-2 border-2 border-black bg-white text-sm font-bold placeholder:text-neutral-400 focus:outline-none focus:ring-2 focus:ring-black">
				<input type="number" name="port" placeholder="port" value="9100" required class="w-full px-3 py-2 border-2 border-black bg-white text-sm font-bold placeholder:text-neutral-400 focus:outline-none focus:ring-2 focus:ring-black">
			</div>
			<input type="text" name="tags" placeholder="tags (comma separated)" class="w-full px-3 py-2 border-2 border-black bg-white text-sm font-bold placeholder:text-neutral-400 focus:outline-none focus:ring-2 focus:ring-black">
			<div class="flex justify-end">
				<button type="submit" class="px-4 py-2 border-2 border-black bg-emerald-400 text-black font-bold text-sm shadow-[3px_3px_0px_0px_#000] hover:shadow-[1px_1px_0px_0px_#000] hover:translate-x-[2px] hover:translate-y-[2px] transition-all cursor-pointer">Register</button>
			</div>
		</form>
	</div>`

	// Server list
	if len(servers) == 0 {
		html += `<div class="bg-white border-2 border-black p-8 text-center shadow-[4px_4px_0px_0px_#000]">
			<p class="text-neutral-500 font-bold">Belum ada server terdaftar.</p>
			<p class="text-sm text-neutral-600 mt-2">Klik <button onclick="toggleForm()" class="font-bold underline hover:bg-yellow-200 cursor-pointer">+ Register Server</button> untuk mulai.</p>
		</div>`
	} else {
		html += `<div id="server-list" class="space-y-4">`
		for _, s := range servers {
			statusDot := "bg-neutral-300"
			statusText := "Unknown"
			if s.LastSeen != nil {
				statusDot = "bg-emerald-400"
				statusText = "Online — " + s.LastSeen.Format("Jan 2 15:04")
			}
			html += `<a href="/dashboard/` + itoa64(s.ID) + `" class="block bg-white border-2 border-black p-5 shadow-[4px_4px_0px_0px_#000] hover:shadow-[2px_2px_0px_0px_#000] hover:translate-x-[2px] hover:translate-y-[2px] transition-all no-underline">
				<div class="flex items-center justify-between">
					<div>
						<h3 class="font-bold text-black">` + s.Name + `</h3>
						<p class="text-sm font-mono text-neutral-600">` + s.Host + `:` + itoa(s.Port) + `</p>
					</div>
					<div class="flex items-center gap-1.5 text-sm">
						<span class="inline-block w-3 h-3 border-2 border-black ` + statusDot + `"></span>
						<span class="font-bold text-neutral-700">` + statusText + `</span>
					</div>
				</div>
			</a>`
		}
		html += `</div>`
	}

	// HTMX auto-refresh + JS toggle
	html += `<div class="mt-4 text-xs font-bold text-neutral-500" hx-get="/dashboard" hx-trigger="every 30s" hx-select="#server-list" hx-swap="outerHTML" hx-target="#server-list">Auto-refresh every 30s</div>
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
		return `<h1 class="text-6xl font-black">404</h1><p class="font-bold">Server not found.</p>`
	}
	metrics, _ := data["Metrics"].([]model.ServerMetric)

	html := `<a href="/dashboard" class="font-bold no-underline hover:bg-yellow-200">&larr; Kembali</a>
	<h1 class="text-3xl font-black uppercase mt-2 mb-1">` + server.Name + `</h1>
	<p class="text-sm font-mono text-neutral-600 mb-6">` + server.Host + `:` + itoa(server.Port) + `</p>`

	if len(metrics) == 0 {
		html += `<div class="bg-white border-2 border-black p-6 shadow-[4px_4px_0px_0px_#000]">
			<p class="text-neutral-600 font-bold text-sm">Belum ada data metrics. Kirim data lewat POST /api/v1/servers/:id/ping</p>
		</div>`
	} else {
		html += `<div class="bg-white border-2 border-black shadow-[4px_4px_0px_0px_#000] overflow-hidden">
			<div class="border-b-2 border-black bg-black text-white px-4 py-2">
				<h2 class="text-sm font-bold uppercase tracking-wider">Metrics Terakhir</h2>
			</div>`
		for _, m := range metrics {
			statusDot := "bg-rose-400"
			if m.Online {
				statusDot = "bg-emerald-400"
			}
			html += `<div class="flex items-center justify-between py-3 px-4 border-b-2 border-black last:border-b-0">
				<div class="flex items-center gap-2">
					<span class="inline-block w-3 h-3 border-2 border-black ` + statusDot + `"></span>
					<span class="text-sm font-bold text-black">` + m.RecordedAt.Format("Jan 2 15:04") + `</span>
				</div>
				<div class="flex gap-4 text-sm font-mono font-bold text-neutral-600">
					<span>CPU: <span class="text-black">` + ftoa(m.CPUPercent) + `%</span></span>
					<span>RAM: <span class="text-black">` + ftoa(m.RAMUsedMB) + `/` + ftoa(m.RAMTotalMB) + `</span> MB</span>
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
	if len(md) == 0 {
		return "<1"
	}
	words := len(strings.Fields(md))
	if words == 0 {
		return "<1"
	}
	min := words / 200
	if min < 1 {
		return "<1"
	}
	return strconv.Itoa(min)
}

func page404HTML(data map[string]any) string {
	_ = data // unused, just for interface
	return `<div class="flex flex-col items-center justify-center py-24 text-center">
		<h1 class="text-9xl font-black text-black border-2 border-black px-8 py-4 inline-block shadow-[8px_8px_0px_0px_#000]">404</h1>
		<h2 class="text-2xl font-black uppercase text-black mt-6 mb-2">Halaman gak ketemu</h2>
		<p class="text-neutral-600 font-bold mb-8 max-w-md">Entah udah dipindah, dihapus, atau emang gak pernah ada.</p>
		<a href="/" class="inline-flex items-center gap-1.5 px-6 py-3 border-2 border-black bg-black text-white font-bold text-sm no-underline shadow-[4px_4px_0px_0px_#000] hover:shadow-[2px_2px_0px_0px_#000] hover:translate-x-[2px] hover:translate-y-[2px] transition-all">
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
