# PRD — Raevtar

**Product Requirements Document**
Domain: raevtar.tech

---

## 1. Tujuan

Platform pribadi all-in-one dengan akses publik via Cloudflare Tunnel. Isinya blog rekomendasi projek GitHub, dashboard monitoring server multi-platform, public lab, landing page profil, admin panel, dan API untuk integrasi agent.

Observed di deployment referensi (Alpine Linux, aarch64) saat idle: Raevtar berjalan di sekitar **24 MB RSS**, CPU time nyaris tidak bergerak, binary sekitar **17.5 MB**, dan SQLite database masih sub-1 MB pada pemakaian sekarang. Ini memperkuat constraint bahwa stack harus tetap hemat resource di device target.

## 2. Target User

Hanya satu: **Latifan**. Bukan produk publik. Semua keputusan desain dibuat untuk single-user dengan kebutuhan spesifik.

## 3. Fitur

### 3.1 Blog
- Artikel berbasis Markdown, disimpan di SQLite
- Kategori: AI Agent, Security, Kernel & Embedded, DevOps, Tools
- Postingan otomatis tiap hari dari Hermes cronjob (jam 08:00)
- API endpoint untuk nulis artikel dari agent (Hermes)
- List per kategori, pagination
- Detail artikel dengan markdown render (goldmark)
- Tags normalized dan RSS feed `/blog/feed.xml`

### 3.2 Server Dashboard
- Daftar server lokal yang dimonitor
- Tiap server punya: nama, host, port, tags
- Metrics: CPU %, load 1/5/15, cores, RAM, disk, temperature jika tersedia, uptime, online/stale/offline
- History metrics (tabel `server_metrics`)
- Public dashboard menampilkan `System Health` public-safe tanpa host/IP, port, tags privat, token, install command, atau audit log
- Admin diagnostics tetap menyimpan endpoint, setup command, token rotation, dan metric history lengkap
- Lightweight Bash agent di tiap server ngirim data via API tanpa SSH credentials
- Per-server agent token bisa di-rotate dari admin panel

### 3.3 Landing Page
- Halaman index — profil singkat, link ke blog & dashboard
- Static tapi dirender via Templ (gak perlu HTML manual)

### 3.4 REST API (v1)
- `GET /api/v1/posts` — list blog posts
- `POST /api/v1/posts` — create blog post (auth)
- `GET /api/v1/search` — public search posts/projects/pages
- `GET /api/v1/projects` — public list projects dengan filter featured, state, sort
- `GET /api/v1/projects/{slug}/updates` — public project timeline
- `GET /api/v1/projects/{slug}/changelog` — public changelog entries
- `GET /api/v1/projects/{slug}/relations` — public related content
- `GET /api/v1/projects/{slug}/showcase` — public showcase items
- `POST /api/v1/projects` — create project (auth)
- `PUT /api/v1/projects/{id}` — update project (auth)
- `DELETE /api/v1/projects/{id}` — delete project (auth)
- Project child resource CRUD (updates, relations, showcase) — auth
- `GET /api/v1/categories` — daftar kategori
- `GET /api/v1/servers` — daftar server + status (auth)
- `POST /api/v1/servers` — register server (auth, return one-time agent token)
- `GET /api/v1/servers/{id}` — detail server (auth)
- `POST /api/v1/servers/{id}/ping` — report metrics (agent token atau admin key)
- `GET /api/v1/servers/{id}/commands` — agent polling pending commands
- `POST /api/v1/servers/{id}/commands/result` — agent report command result
- `GET /api/v1/hoststats` — host resource snapshot (auth)
- Editorial inbox API: contract, list, create, claim, detail, update, complete, fail (auth)
- `/docs` untuk public-safe docs dan read-only OpenAPI surface; endpoint admin/server/agent tidak dipublikasikan di docs publik

### 3.5 Admin Panel
- Session login di `/admin/login`
- Manage posts, topics, projects, media, servers, users, webhooks, dan static pages
- Editorial inbox untuk managing content queue Hermes
- RBAC role: `owner`, `admin`, `operator`, `readonly`
- Server diagnostics dengan metric history, command queue, token rotation
- Manage webhook alert endpoints dan event log
- Page editor untuk about dan contact pages
- Audit log untuk login/logout dan aksi admin

### 3.6 Hardening
- Production mode wajib punya `RAEVTAR_ADMIN_KEY` dan `RAEVTAR_ADMIN_PASS`
- Rate limiting in-memory per IP
- Login throttling in-memory per `IP + username` dan IP-only spray guard
- Request body caps untuk login, API payload, admin forms, dan media upload
- Generic `internal server error` untuk 500; detail hanya di log server
- CSP `script-src 'self'`; HTMX disajikan dari `/static/js/htmx.min.js`

### 3.7 Webhook Alerts
- Konfigurasi webhook URL dengan HMAC-SHA256 signature
- Threshold alerts: CPU >= 90%, RAM >= 90%, Disk >= 90%
- Webhook event log dengan response tracking
- Admin UI di `/admin/webhooks`

### 3.8 Server Command Queue
- Admin bisa mengantarkan perintah ke server via admin panel
- Agent polling pending commands via API
- Status lifecycle: pending → running → completed/failed
- History per server di admin panel

### 3.9 SEO & Discovery
- Sitemap XML di `/sitemap.xml` — semua pages, posts, projects
- LLMs.txt di `/llms.txt` — LLM-friendly site summary
- JSON-LD structured data (BlogPosting, CreativeWork, WebSite)
- Dynamic OG images per blog post dan project
- Canonical URLs dan robots.txt

### 3.10 Integrasi Hermes
- Cronjob harian: riset projek GitHub → nulis artikel → POST ke API
- Editorial inbox: Hermes cek queue Raevtar dulu sebelum tulis artikel
- Hermes bisa manual: "tulis ini ke blog" → curl endpoint
- Server monitoring: agent push metrics dari tiap perangkat

## 4. Non-Goals (sengaja gak dilakukan)

- ❌ Public account registration / multi-tenant SaaS
- ❌ Comments / diskusi (ini bukan platform sosial)
- ❌ Database server terpisah (SQLite cukup)
- ❌ SPA / React / Next.js (RAM hp gak cukup)
- ❌ Docker / Kubernetes (gak perlu untuk satu binary)
- ❌ CI/CD pipeline (cukup `make build && ./raevtar`)
- ~~❌ Search engine (bisa nanti, gak wajib awal)~~ ✅ Implemented via `/search` + `GET /api/v1/search`

## 5. Constraints

| Constraint | Detail |
|------------|--------|
| RAM | Target < 50 MB RSS. Harus hemat. |
| Storage | SQLite dan binary kecil (< 25 MB). |
| Network | Gak ada IP publik. Harus Cloudflare Tunnel. |
| OS | Cross-platform: Linux (Alpine/Debian/Arch), macOS, BSD. Static binary, no CGO. |

Catatan observasional: virtual memory Go process bisa terlihat besar karena mapping runtime, tapi footprint yang lebih relevan untuk operasi harian adalah RSS nyata. Build cache Go juga bisa ratusan MB, namun itu overhead toolchain/dev, bukan biaya runtime aplikasi.

## 6. Stack Decision

| Layer | Pilihan | Alternatif | Alasan |
|-------|---------|------------|--------|
| Backend | Go + Chi | Python/FastAPI, Node/Express | Single binary, rendah RAM, build cepat |
| Template | Templ | Jinja2, Pug, html/template | Type-safe, compile-time check |
| CSS | Tailwind | Bootstrap, vanilla CSS | Utility-first, responsive out of box |
| Interaksi | Self-hosted HTMX | Alpine.js, React, Vue | Gak perlu JS build, ringan, tidak bergantung CDN runtime |
| Database | SQLite + modernc | PostgreSQL, MySQL | Gak perlu server, pure Go |
| Tunnel | cloudflared | ngrok, frp | DNS sendiri (raevtar.tech), systemd setup aktif |

## 7. Success Metrics

- Blog bisa diakses publik via `raevtar.tech`
- Ada artikel baru tiap hari (dari Hermes)
- Dashboard nunjukin status semua server lokal
- Public dashboard jelas membedakan summary aman vs diagnostics admin-only
- API bisa dipanggil dari luar
- Binary bisa jalan di background via service manager
- Restart hp → service jalan otomatis
