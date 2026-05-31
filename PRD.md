# PRD — Raevtar

**Product Requirements Document**
Domain: raevtar.tech
Runtime: postmarketOS (aarch64) — Redmi Note 5 (whyred)

---

## 1. Tujuan

Platform pribadi all-in-one yang jalan di HP (postmarketOS) dengan akses publik via Cloudflare Tunnel. Isinya blog rekomendasi projek GitHub, dashboard monitoring server lokal, landing page profil, dan API untuk integrasi masa depan.

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

### 3.2 Server Dashboard
- Daftar server lokal yang dimonitor
- Tiap server punya: nama, host, port, tags
- Metrics: CPU, RAM, disk, uptime, online/offline
- History metrics (tabel server_metrics)
- Agent (Hermes cron atau script di tiap server) ngirim data via API

### 3.3 Landing Page
- Halaman index — profil singkat, link ke blog & dashboard
- Static tapi dirender via Templ (gak perlu HTML manual)

### 3.4 REST API (v1)
- `GET/POST /api/v1/posts` — CRUD blog posts
- `GET /api/v1/categories` — daftar kategori
- `GET /api/v1/servers` — daftar server + status
- `POST /api/v1/servers/:id/ping` — report metrics
- `GET /api/v1/servers/:id` — detail server + history
- Auto-generated OpenAPI docs via `/docs`

### 3.5 Integrasi Hermes
- Cronjob harian: riset projek GitHub → nulis artikel → POST ke API
- Hermes bisa manual: "tulis ini ke blog" → curl endpoint
- Server monitoring: polling dari sini atau agent ngirim data

## 4. Non-Goals (sengaja gak dilakukan)

- ❌ Multi-user / auth system (single-user, tunnel udah cukup)
- ❌ Comments / diskusi (ini bukan platform sosial)
- ❌ Database server terpisah (SQLite cukup)
- ❌ SPA / React / Next.js (RAM hp gak cukup)
- ❌ Docker / Kubernetes (gak perlu untuk satu binary)
- ❌ CI/CD pipeline (cukup `go build && ./raevtar`)
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

## 6. Stack Decision

| Layer | Pilihan | Alternatif | Alasan |
|-------|---------|------------|--------|
| Backend | Go + Chi | Python/FastAPI, Node/Express | Single binary, rendah RAM, build cepat |
| Template | Templ | Jinja2, Pug, html/template | Type-safe, compile-time check |
| CSS | Tailwind | Bootstrap, vanilla CSS | Utility-first, responsive out of box |
| Interaksi | HTMX | Alpine.js, React, Vue | Gak perlu JS build, ringan |
| Database | SQLite + modernc | PostgreSQL, MySQL | Gak perlu server, pure Go |
| Tunnel | cloudflared | ngrok, frp | DNS sendiri (raevtar.tech), OpenRC support |

## 7. Success Metrics

- Blog bisa diakses publik via `raevtar.tech`
- Ada artikel baru tiap hari (dari Hermes)
- Dashboard nunjukin status semua server lokal
- API bisa dipanggil dari luar
- Binary bisa jalan di background via OpenRC
- Restart hp → service jalan otomatis
