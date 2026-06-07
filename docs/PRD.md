# PRD — Raevtar

**Product Requirements Document**
Domain: raevtar.tech
Runtime: postmarketOS (aarch64) — Redmi Note 5 (whyred)

---

## 1. Tujuan

Platform pribadi all-in-one yang jalan di HP (postmarketOS) dengan akses publik via Cloudflare Tunnel. Isinya blog rekomendasi projek GitHub, dashboard monitoring server lokal, public lab, landing page profil, admin panel, dan API kecil untuk integrasi agent.

Observed di deployment whyred saat idle: Raevtar berjalan di sekitar **24 MB RSS**, CPU time nyaris tidak bergerak, binary sekitar **17.5 MB**, dan SQLite database masih sub-1 MB pada pemakaian sekarang. Ini memperkuat constraint bahwa stack harus tetap hemat resource di device target.

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
- `GET /api/v1/categories` — daftar kategori
- `GET /api/v1/servers` — daftar server + status (auth)
- `POST /api/v1/servers` — register server (auth, return one-time agent token)
- `POST /api/v1/servers/{id}/ping` — report metrics (agent token atau admin key)
- `GET /api/v1/servers/{id}` — detail server (auth)
- `GET /api/v1/hoststats` — host resource snapshot (auth)
- `/docs` untuk public-safe docs dan read-only OpenAPI surface; endpoint admin/server/agent tidak dipublikasikan di docs publik

### 3.5 Admin Panel
- Session login di `/admin/login`
- Manage posts, media, servers, users
- RBAC role: `owner`, `admin`, `operator`, `readonly`
- Audit log untuk login/logout dan aksi admin

### 3.6 Hardening
- Production mode wajib punya `RAEVTAR_ADMIN_KEY` dan `RAEVTAR_ADMIN_PASS`
- Rate limiting in-memory per IP
- Login throttling in-memory per `IP + username` dan IP-only spray guard
- Request body caps untuk login, API payload, admin forms, dan media upload
- Generic `internal server error` untuk 500; detail hanya di log server
- CSP `script-src 'self'`; HTMX disajikan dari `/static/js/htmx.min.js`

### 3.7 Integrasi Hermes
- Cronjob harian: riset projek GitHub → nulis artikel → POST ke API
- Hermes bisa manual: "tulis ini ke blog" → curl endpoint
- Server monitoring: agent push metrics dari tiap perangkat

## 4. Non-Goals (sengaja gak dilakukan)

- ❌ Public account registration / multi-tenant SaaS
- ❌ Comments / diskusi (ini bukan platform sosial)
- ❌ Database server terpisah (SQLite cukup)
- ❌ SPA / React / Next.js (RAM hp gak cukup)
- ❌ Docker / Kubernetes (gak perlu untuk satu binary)
- ❌ CI/CD pipeline (cukup `make build && ./raevtar`)
- ❌ Search engine (bisa nanti, gak wajib awal)

## 5. Constraints

| Constraint | Detail |
|------------|--------|
| RAM | 3.6GB total (~2GB available). Harus hemat. |
| Storage | 50GB (32GB free). SQLite dan binary kecil. |
| CPU | aarch64 (SDM660). Build harus cepet. |
| Network | Gak ada IP publik. Harus Cloudflare Tunnel. |
| Arsitektur | aarch64 — semua binary harus ARM64 native. |
| OS | Alpine-based (postmarketOS). Pakai apk. |

Catatan observasional saat ini: virtual memory Go process bisa terlihat besar karena mapping runtime, tapi footprint yang lebih relevan untuk operasi harian adalah RSS nyata. Build cache Go juga bisa ratusan MB, namun itu overhead toolchain/dev, bukan biaya runtime aplikasi.

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
