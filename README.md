# Raevtar

**Personal platform** — blog rekomendasi projek GitHub, dashboard monitoring server lokal, landing page, dan REST API. Satu binary, jalan di postmarketOS (aarch64, Redmi Note 5/whyred).

## Status

| Aspek | Status |
|-------|--------|
| Konsep | 8/10 |
| Arsitektur | 8/10 |
| Dokumentasi | 7/10 |
| Production readiness | 5.5/10 |

## Fitur

### Blog
- Artikel markdown, disimpan di SQLite
- 5 kategori: AI Agent, Security, Kernel & Embedded, DevOps, Tools
- Pagination + filter kategori
- Auto-post via Hermes cron setiap hari
- REST API untuk create/list posts

### Server Dashboard
- Daftar server lokal yang dimonitor
- Metrics: CPU, RAM, uptime, online/offline
- History metrics per server
- HTMX auto-refresh tiap 30 detik

### Landing Page
- Hero section + recent posts + kategori + server status
- Navigation menu ke semua halaman

### REST API
- `GET /api/v1/posts` — list posts
- `POST /api/v1/posts` — create post (auth required)
- `GET /api/v1/categories` — list categories
- `GET /api/v1/servers` — list servers
- `POST /api/v1/servers/:id/ping` — record metrics (auth required)

## Stack

| Lapisan | Teknologi |
|---------|-----------|
| Backend | Go 1.26, chi router |
| Templating | a-h/templ (type-safe) |
| Frontend | HTMX + Tailwind CSS |
| Database | SQLite (modernc.org/sqlite — no CGO) |
| Tunnel | Cloudflare Tunnel |
| Domain | raevtar.tech |

## Struktur

```
raevtar/
├── cmd/server/        # Entry point — init config, DB, router, start HTTP
├── internal/
│   ├── config/        # Env-based config loader
│   ├── model/         # Data structs (Post, Category, Server, Metric)
│   ├── repo/          # Database CRUD (SQLite + sqlx)
│   ├── service/       # Business logic (blog, monitor, seed)
│   └── handler/       # HTTP handlers + routing (Chi)
├── cron/
│   └── backup.sh      # SQLite backup (daily via systemd timer)
├── migrations/        # SQL init schema
└── static/            # CSS assets
```

## Arsitektur Layer

```
Handler → Service → Repo → SQLite
```

Handler gak boleh panggil repo langsung. Service gak tahu HTTP. Repo cuma query.

## Konfigurasi (Environment Variables)

| Variable | Default | Keterangan |
|----------|---------|------------|
| `RAEVTAR_ADDR` | `:8080` | Listen address |
| `RAEVTAR_DB` | `~/.raevtar/data.db` | Path SQLite database |
| `RAEVTAR_DOMAIN` | `raevtar.tech` | Domain public |
| `RAEVTAR_LOG_LEVEL` | `info` | debug / info / warn / error |
| `RAEVTAR_ADMIN_KEY` | `""` | **WAJIB** untuk API write. Constant-time validated |
| `RAEVTAR_ENV` | `""` | Set ke `production` untuk mode produksi |

## Quick Start

```bash
# Build
cd /home/latif/raevtar
go build -o raevtar ./cmd/server/

# Run (ganti admin key-nya)
RAEVTAR_ADMIN_KEY=your-secret-key ./raevtar

# Test
curl http://localhost:8080/
```

## Deploy (postmarketOS)

```bash
# systemd service
sudo cp raevtar.service /etc/systemd/system/
sudo systemctl enable --now raevtar

# Backup otomatis (harian)
# systemd-timer atau crontab:    0 3 * * * /home/latif/raevtar/cron/backup.sh
```

## Prinsip

1. **Handler → Service → Repo** — layer terpisah, gak ada campur aduk.
2. **Satu binary** — `go build` → `./raevtar` → langsung jalan. Gak perlu runtime.
3. **Backend process stateless** — persistent state di SQLite.
4. **SSR + HTMX** — halaman dirender server, interaksi ringan tanpa JS berat.
5. **Single domain** — blog, dashboard, API di satu tempat.
