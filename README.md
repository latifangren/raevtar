# Raevtar

Raevtar adalah platform personal yang jalan sebagai satu binary Go: blog, public lab, dashboard monitoring server lokal, admin panel, dan REST API kecil. Target utamanya tetap sederhana: bisa hidup di postmarketOS/aarch64, hemat resource, dan cukup aman untuk dibuka lewat domain pribadi.

## Current Scope

| Area | Status |
|------|--------|
| Blog | Markdown posts, kategori, tags, RSS, cover image, media upload |
| Monitoring | Push-based agent telemetry, public-safe dashboard, admin diagnostics |
| Admin | Session login, RBAC, posts/media/servers/users/audit log |
| API | Public read endpoints + protected write/monitoring endpoints |
| Hardening | Request caps, login throttling, CSP, generic 500s, production secret checks |

Raevtar bukan multi-tenant SaaS. Ini personal app untuk `raevtar.tech`, dengan boundary jelas antara tampilan publik dan data operator.

## Features

### Blog

- Artikel Markdown disimpan di SQLite dan dirender pakai Goldmark.
- Kategori bawaan: AI Agent, Security, Kernel & Embedded, DevOps, Tools.
- Tags normalized (`tags` + `post_tags`) dan badge di UI.
- Admin Content Studio: draft/publish, Markdown preview, media upload, cover image.
- RSS feed di `/blog/feed.xml`.
- API `POST /api/v1/posts` untuk integrasi agent seperti Hermes.

### Public Monitoring

- `/dashboard` menampilkan status semua node tanpa membuka topology privat.
- `Platform System Health` menampilkan snapshot host Raevtar: CPU load, RAM, disk, temperature jika ada.
- Tiap node punya `System Health` public-safe: CPU %, load, cores, RAM, disk, temperature, uptime, latest sample age, sample count, history window, dan availability aggregate.
- HTMX refresh: grid tiap 30 detik, detail node tiap 15 detik.
- Public view tidak menampilkan host/IP, port, tags privat, token, install command, audit log, atau raw setup detail.

### Admin Panel

- Login session di `/admin/login`.
- RBAC role: `owner`, `admin`, `operator`, `readonly`.
- Manage posts, media, servers, users, dan audit log.
- Server diagnostics di `/admin/servers/{id}` berisi endpoint, metric history, setup command, token rotation, dan activity log admin-only.

### REST API

| Method | Path | Scope |
|--------|------|-------|
| `GET` | `/api/v1/posts` | Public list posts |
| `POST` | `/api/v1/posts` | Admin key required |
| `GET` | `/api/v1/categories` | Public categories |
| `GET` | `/api/v1/servers` | Admin key required |
| `GET` | `/api/v1/servers/{id}` | Admin key required |
| `POST` | `/api/v1/servers` | Admin key required; returns one-time `agent_token` |
| `POST` | `/api/v1/servers/{id}/ping` | Agent token or admin key |
| `GET` | `/api/v1/hoststats` | Admin key required |
| `GET` | `/docs`, `/lab/docs` | Public-safe docs |

Public docs sengaja hanya menjelaskan read-only surface dan privacy boundary. Endpoint admin/server setup tetap operator-only.

## Agent Monitoring

Monitoring pakai push model, bukan SSH pull. Tiap perangkat jalanin agent ringan dan kirim metrics ke Raevtar:

```bash
RAEVTAR_URL=http://192.168.100.5:8080 \
RAEVTAR_SERVER_ID=2 \
RAEVTAR_AGENT_TOKEN=token-per-server \
/usr/local/bin/raevtar-agent.sh
```

Agent mengirim CPU percent, CPU load 1/5/15, core count, RAM used/total, disk used/total, uptime, online flag, dan temperature jika sensor tersedia. Token diambil dari `/admin/servers` atau response `POST /api/v1/servers`; token hanya ditampilkan saat create/rotate.

## Hardening Notes

- Production mode menolak start kalau `RAEVTAR_ADMIN_KEY` atau `RAEVTAR_ADMIN_PASS` kosong.
- API auth pakai bearer token dengan constant-time validation.
- Global rate limit in-memory: 60 request/menit per IP.
- Admin login throttling in-memory: per `IP + username` dan IP-only spray guard.
- Request body dibatasi untuk login, API payload, admin forms, dan upload media.
- Internal server errors dikembalikan sebagai pesan generik; detail masuk log server.
- CSP memakai `script-src 'self'`; HTMX dan UI helper disajikan dari `/static/js/`.
- `RAEVTAR_TRUSTED_PROXY_CIDRS` opsional untuk membaca `CF-Connecting-IP` dari proxy tepercaya saja.

## Stack

| Lapisan | Teknologi |
|---------|-----------|
| Backend | Go 1.26.3, `github.com/go-chi/chi/v5` |
| Templates | `github.com/a-h/templ` |
| Frontend | SSR + self-hosted HTMX + Tailwind CSS |
| Database | SQLite via `modernc.org/sqlite` + `github.com/jmoiron/sqlx` |
| Markdown | `github.com/yuin/goldmark` |
| Runtime target | postmarketOS/aarch64, Cloudflare Tunnel |

## Struktur

```text
raevtar/
├── cmd/server/        # Entry point: config, DB, router, HTTP server
├── internal/
│   ├── config/        # Env-based config loader
│   ├── model/         # Data structs
│   ├── repo/          # SQL queries + migration helpers
│   ├── service/       # Business logic and validation
│   ├── handler/       # HTTP handlers + routing
│   └── view/          # templ layouts, pages, admin views, components
├── cron/              # Backup/automation scripts
├── migrations/        # Fresh SQLite schema
├── static/            # CSS, JS, agent script, uploads/static assets
└── docs/              # Notes and historical debugging docs
```

## Arsitektur Layer

```text
Handler -> Service -> Repo -> SQLite
```

- Handler parse HTTP request, set response, dan render Templ.
- Service berisi business logic dan validasi; tidak tahu `http.Request`/`http.ResponseWriter`.
- Repo hanya query SQL dan mapping data.
- Model hanya struct.

## Konfigurasi

| Variable | Default | Keterangan |
|----------|---------|------------|
| `RAEVTAR_ADDR` | `:8080` | Listen address |
| `RAEVTAR_DB` | `~/.raevtar/data.db` | Path SQLite database |
| `RAEVTAR_MEDIA_DIR` | `~/.raevtar/uploads` | Direktori upload media publik |
| `RAEVTAR_DOMAIN` | `raevtar.tech` | Domain public |
| `RAEVTAR_LOG_LEVEL` | `info` | `debug`, `info`, `warn`, `error` |
| `RAEVTAR_ADMIN_KEY` | `""` | Wajib untuk API protected |
| `RAEVTAR_ADMIN_USER` | `admin` | Seed admin username |
| `RAEVTAR_ADMIN_PASS` | `""` | Wajib untuk admin login |
| `RAEVTAR_ENV` | `""` | Set `production` untuk strict secret check |
| `RAEVTAR_TRUSTED_PROXY_CIDRS` | `""` | CIDR proxy tepercaya untuk forwarded client IP |

## Build & Run

```bash
# Build production binary
make build

# Run local binary
RAEVTAR_ADMIN_KEY=dev-key RAEVTAR_ADMIN_PASS=dev-pass ./raevtar

# Generate templ manually if needed
go run github.com/a-h/templ/cmd/templ@v0.3.906 generate

# Regenerate Tailwind manually if needed
npx --yes tailwindcss@3.4.19 -i static/css/tailwind.src.css -o static/css/style.css --minify
```

`make build` menjalankan templ generate, Tailwind build, lalu `go build`.

## Deploy Singkat

```bash
# systemd service setup, jalankan di host operator
sudo cp raevtar.service /etc/systemd/system/
sudo systemctl enable --now raevtar

# backup harian via systemd timer atau cron
# 0 3 * * * /home/latif/raevtar/cron/backup.sh
```

Runbook lengkap ada di `DEPLOYMENT.md`. Jangan restart/deploy service kecuali memang sedang melakukan operasi deploy.

## Prinsip

1. Satu binary, satu runtime utama: Go.
2. SSR dulu; HTMX hanya untuk progressive enhancement ringan.
3. Public view boleh menampilkan ringkasan health, tapi topology dan setup tetap admin-only.
4. SQLite cukup untuk single-user personal platform.
5. Tambah fitur lewat layer yang benar, bukan shortcut handler langsung ke repo.
