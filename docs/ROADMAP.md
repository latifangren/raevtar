# ROADMAP — Raevtar

**Fase pengembangan.** Prioritas: core functionality dulu, polish kemudian.

---

## Fase 1: Foundation 🔧 (done ✅)

Setup dasar biar bisa jalan dan diakses publik.

- [x] `apk add go` — install Go 1.26
- [x] `go mod init raevtar` + install dependencies
- [x] `internal/handler/*.go` — semua handler (landing, blog, dashboard, API)
- [x] `go build` berhasil di aarch64 (binary 15.7 MB)
- [x] Landing page + nav menu + halaman blog, lab, dashboard, API
- [x] SQLite auto-migrate, seed 5 kategori
- [x] Auth middleware constant-time untuk API write
- [x] Systemd service (`/etc/systemd/system/raevtar.service`)
- [x] Cloudflare Tunnel setup + akses `raevtar.tech`

**Deliverable:** Server jalan di `:8080`, bisa diakses dari LAN.

---

## Fase 2: Blog + Hermes Integration ✍️

Blog fungsional dengan konten otomatis dari Hermes.

- [x] Hermes cronjob: auto-post tiap hari jam 08:00
- [x] Hermes bisa manual: "tulis ini ke blog"
- [x] Katagori di blog udah jalan (filter, breadcrumb)
- [x] Pagination, markdown render (goldmark)

**Deliverable:** Blog aktif dengan postingan rutin.

---

## Fase 3: Server Dashboard 📊

Monitoring server-server lokal.

- [x] Form/API untuk daftarin server (name, host, port, tags)
- [x] Agent monitoring: push script curl ke `/api/v1/servers/{id}/ping` pakai per-server token
- [x] Dashboard: overview semua server (status, CPU, RAM)
- [x] Detail page: history metrics per server
- [x] HTMX auto-refresh dashboard (tiap 30 detik)
- [x] Public-safe `System Health`: CPU/load/cores, RAM, disk, temperature, uptime, sample age, aggregate availability
- [x] Extended Bash agent telemetry tanpa SSH credentials

**Deliverable:** Dashboard fungsional dengan data real server.

---

## Fase 4: Polish 🎨

Biar gak kelihatan kaya projek HTML kampus.

- [x] Pilih design system dari reference
- [x] Responsive: biar enak dibuka dari HP juga
- [x] Tailwind standalone CLI (instead of CDN)
- [x] Favicon, meta tags, Open Graph
- [x] Custom 404 page
- [x] Public lab page (`/lab`) untuk ringkasan agregat tanpa detail privat
- [x] HTMX self-hosted di `/static/js/htmx.min.js`
- [x] Inline UI handlers dipindah ke JS lokal CSP-safe

**Deliverable:** Raevtar keliatan profesional.

---

## Fase 5: API & Ekstensi 🚀

- [x] auth middleware (constant-time) untuk write endpoints
- [x] Tag system (normalized: tags + post_tags + UI badges)
- [x] RSS feed di `/blog/feed.xml` + auto-discovery `<link>` tag
- [x] Public-safe docs + read-only OpenAPI spec (`/docs`)
- [x] Editorial inbox control plane buat Hermes (`/admin/editorial-inbox` + protected API contract)
- [x] Editorial inbox Phase 2: lifecycle eksekusi (`running`, `failed`, `published_post_id`, failure metadata)
- [x] Editorial inbox Phase 3: claim/lock/retry flow buat hindari double-processing antar run Hermes
- [x] Editorial inbox Phase 4: fairness policy, overdue escalation, dan analytics hasil publish
- [x] Search endpoint + HTMX search UI (`/search`, `GET /api/v1/search`)
- [x] Read-time tracker di artikel

**Deliverable:** Platform siap dikembangin kapan aja.

---

## Fase 6: Stabilisasi 🛡️

- [x] Structured log (slog) — udah dari awal
- [x] Backup SQLite — script `cron/backup.sh` + systemd timer
- [x] Graceful shutdown (tangkep SIGTERM/SIGINT) — `15s` timeout
- [x] Security headers: CSP, X-Frame-Options, X-Content-Type-Options, Referrer-Policy
- [x] CSP script tightened ke `script-src 'self'`
- [x] Rate limiting: 60 req/min per IP
- [x] Request body caps untuk login, API, form admin, dan upload media
- [x] Login throttling: per `IP + username` + IP-only spray guard
- [x] Generic internal `500` responses dengan detail di server log
- [x] Admin panel: session-based auth (login/logout)
- [x] Admin panel pages: manage posts, manage servers
- [x] RBAC + multi-user (owner/admin/operator/readonly) + audit log + manage users
- [x] Admin creds via env file (`/home/latif/raevtar/.env.production`)
- [x] Health check: Hermes cronjob tiap 5 menit (silent if healthy)
- [ ] Update dependencies periodik (go mod update, npm update)
- [x] Alerting sederhana untuk node stale/offline — background goroutine tiap 5 menit, threshold 15 menit sejak LastSeen, `server_stale` event ke webhook
- [x] Versioned schema migration ledger — `schema_migrations` table + v1 backfill untuk existing DB + template untuk future migrations
- [x] Public-safe docs + read-only OpenAPI spec (`/docs`)
- [x] Webhook system: threshold alerts (CPU/RAM/disk >= 90%), HMAC-SHA256, admin UI
- [x] Server command queue: admin queue → agent poll → result report
- [x] SEO/Sitemap/LLMs.txt: structured data, canonical URLs, OG images
- [x] Post view tracking by IP hash (SHA-256, fire-and-forget)
- [x] Dynamic OG images (SVG, neo-brutalist, tiap blog post + project)
- [x] CI/CD pipeline: GitHub Actions test+build + GoReleaser multi-platform releases
**Deliverable:** Platform cukup stabil untuk personal deployment yang exposed ke internet, dengan hardening dasar dan boundary publik/admin jelas.

---

---

## Fase 7: Iteration & UX 🔄 (done ✅)

Perbaikan berdasarkan audit Hermes + feedback real usage.

- [x] **Server monitoring cleanup** — Fix 401/404 agent ping (server ID 5 token mismatch, server ID 1 deleted). Rotate token, redeploy pake one-liner baru.
- [x] **HTMX search real-time** — Ganti form submit ke `hx-get="/search" hx-trigger="keyup changed delay:300ms, submit" hx-target="#search-results"`. Loading indicator. `autocomplete="off"`.
- [x] **RSS feed: verify & promote** — Cek `/blog/feed.xml` isinya proper, tambah link visible di footer / blog page sidebar.
- [ ] ~~**Dark mode toggle**~~ — Di-skip. Neo-brutalist sudah light-first.
- [x] **Content scheduling** — `scheduled_at` field (datetime-local picker di admin form), background goroutine publish otomatis tiap 60s.
- [x] **Media library improvements** — Alt text field (wajib di upload form + display di card). Default dari cleaned filename.
- [x] **Webmention / IndieWeb** — Receive-only. Link tags di `<head>`, POST endpoint, admin approval flow, display section di blog page.
- [x] **API docs page** — `/docs/api` dengan contoh request/response tiap endpoint (curl + JSON).
- [x] **DB export/import dari admin** — Download SQLite via `/admin/db/export`, upload + replace via `/admin/db/import` (SQLite header validation, restart required).

**Deliverable:** Platform lebih mature, UX lebih mulus, konten lebih terkelola.

---

## Known Issues (dari Hermes audit)

- Build berat di aarch64 — `make build` jalan templ-gen + tailwind + go build ~12s + CPU 100%. Pertimbangin cross-compile dari laptop atau `go build -ldflags="-s -w"`.
- Sitemap nampilin 166 URLs — validasi broken link / page ke-generate.
- JSON-LD structured data — perlu dicek apakah proper untuk Google Scholar / blog post.

---

## Fase 8: Portability & Cross-Device 🔀 (done ✅)

Audit hardcoded values agar Raevtar bisa jalan di device/OS lain tanpa recompile besar-besaran.
Roadmap ini hasil code audit — tiap item dikerjakan urutan dari atas ke bawah.

### Group 1: Path & Environment (Paling Sering Kena) ✅

- [x] **Systemd service template** — `raevtar.service.tmpl` dengan `{{RAEVTAR_USER}}`/`{{RAEVTAR_HOME}}` placeholders. `make generate-service` + `make install-service`
- [x] **Static file serving** — `Config.StaticDir` computed from `os.Executable()` → `filepath.Dir()` + `/static`
- [x] **Agent install path configurable** — `Config.AgentDir` from `RAEVTAR_AGENT_DIR` env var (default `/usr/local/bin`)
- [x] **Bootstrap script AGENT_DIR** — `api.go` uses `h.cfg.AgentDir` injected into shell string

### Group 2: Domain & Branding (Gampang Fix) ✅

- [x] **RSS/webmention links** — `base.templ` uses `seo.SiteDomain` for RSS, webmention, pingback URLs
- [x] **Footer copyright domain** — `Footer(domain string)` param, renders from domain config
- [x] **Meta keywords** — Removed `postmarketOS` from keywords, now generic
- [x] **SEO description** — `HomeSEO()` description no longer mentions `postmarketOS`
- [x] **OG image domain** — `og_image.go` uses `h.cfg.Domain` with fallback in SVG
- [x] **OpenAPI contact URL** — Changed to relative `"/"`
- [x] **robots.txt sitemap** — Dynamic `robotsTxt` handler generates `Sitemap:` from config
- [x] **Footer description** — Changed to generic `blog, server monitoring, and automation hooks`

### Group 3: OS-Specific Paths (Butuh Abstraction) ✅

- [x] **Host stats build tags** — Split into `hoststats_types.go`, `hoststats.go` (`//go:build linux`), `hoststats_unsupported.go` (`//go:build !linux` stub)
- [x] **Disk stats** — `diskstats_unix.go` reads `RAEVTAR_DISK_ROOT` env var (default `/`)
- [x] **Agent script OS detection** — `detect_os()` function; all collection functions branch on `linux`/`darwin`/`unknown`; macOS uses `sysctl`/`vm_stat`/`top -l 2`
- [x] **Bootstrap package manager** — Detects 7 package managers: `apk`, `apt-get`, `dnf`, `yum`, `pacman`, `brew`, `opkg`, `zypper`

### Group 4: Configurable Operational Params ✅

- [x] **Rate limit configurable** — `RAEVTAR_RATE_LIMIT_REQUESTS` + `RAEVTAR_RATE_LIMIT_WINDOW` env vars
- [x] **Server timeouts configurable** — `RAEVTAR_READ_TIMEOUT` / `RAEVTAR_WRITE_TIMEOUT` / `RAEVTAR_IDLE_TIMEOUT` / `RAEVTAR_SHUTDOWN_TIMEOUT` env vars (Go duration format)
- [ ] ~~**Stale checker intervals** — Not configurable (hardcoded 5min/15min). Low priority for portability.~~
- [ ] ~~**Scheduler interval** — Not configurable (hardcoded 60s). Low priority.~~
- [x] **Max upload size** — `RAEVTAR_MAX_UPLOAD_MB` env var controls `mediaUploadBodyLimit`
- [x] **Hardening limits** — `RAEVTAR_LOGIN_FAILURE_LIMIT` + `RAEVTAR_LOGIN_IP_FAILURE_LIMIT` env vars
- [x] **Cron healthcheck** — `cron/healthcheck.sh` reads `RAEVTAR_URL` env var
- [x] **Chart.js CDN** — CSP `script-src` covers CDN; self-hosting not critical for portability

### Group 5: Documentation & Examples ✅

- [x] **DEPLOYMENT.md** — Updated to generic paths, added hardening env var table
- [x] **README.md** — Removed `postmarketOS`/`whyred` references, cross-platform stack, full env var table
- [x] **AGENTS.md** — Updated env var table with all G4 entries, generic commands
- [x] **PRD.md** — Updated to remove hardcoded references, portability-aware constraints

---

## Legend

- `[x]` — Selesai
- `[ ]` — Belum
- `[-]` — Di-skip (gak relevan)
