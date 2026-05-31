# ROADMAP — Raevtar

**Fase pengembangan.** Prioritas: core functionality dulu, polish kemudian.

---

## Fase 1: Foundation 🔧 (done ✅)

Setup dasar biar bisa jalan dan diakses publik.

- [x] `apk add go` — install Go 1.26
- [x] `go mod init raevtar` + install dependencies
- [x] `internal/handler/*.go` — semua handler (landing, blog, dashboard, API)
- [x] `go build` berhasil di aarch64 (binary 15.7 MB)
- [x] Landing page + nav menu + halaman blog, dashboard, API
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

**Deliverable:** Dashboard fungsional dengan data real server.

---

## Fase 4: Polish 🎨

Biar gak kelihatan kaya projek HTML kampus.

- [x] Pilih design system dari reference
- [x] Responsive: biar enak dibuka dari HP juga
- [x] Tailwind standalone CLI (instead of CDN)
- [x] Favicon, meta tags, Open Graph
- [x] Custom 404 page

**Deliverable:** Raevtar keliatan profesional.

---

## Fase 5: API & Ekstensi 🚀

- [x] auth middleware (constant-time) untuk write endpoints
- [x] Tag system (normalized: tags + post_tags + UI badges)
- [x] RSS feed di `/blog/feed.xml` + auto-discovery `<link>` tag
- [x] OpenAPI spec + Swagger UI (`/docs`)
- [ ] Read-time tracker di artikel

**Deliverable:** Platform siap dikembangin kapan aja.

---

## Fase 6: Stabilisasi 🛡️

- [x] Structured log (slog) — udah dari awal
- [x] Backup SQLite — script `cron/backup.sh` + systemd timer
- [x] Graceful shutdown (tangkep SIGTERM/SIGINT) — `15s` timeout
- [x] Security headers: CSP, X-Frame-Options, X-Content-Type-Options, Referrer-Policy
- [x] Rate limiting: 60 req/min per IP
- [x] Admin panel: session-based auth (login/logout)
- [x] Admin panel pages: manage posts, manage servers
- [x] RBAC + multi-user (owner/admin/operator/readonly) + audit log + manage users
- [x] Admin creds via env file (`/home/latif/raevtar/.env.production`)
- [x] Health check: Hermes cronjob tiap 5 menit (silent if healthy)
- [ ] Update dependencies periodik (go mod update, npm update)
- [x] OpenAPI spec + Swagger UI (`/docs`)

**Deliverable:** Platform stabil, aman exposed ke internet.

---

## Legend

- `[x]` — Selesai
- `[ ]` — Belum
- `[-]` — Di-skip (gak relevan)
