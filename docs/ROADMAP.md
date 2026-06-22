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

## Legend

- `[x]` — Selesai
- `[ ]` — Belum
- `[-]` — Di-skip (gak relevan)
