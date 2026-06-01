# AGENTS.md — Raevtar

**Panduan untuk AI agents** (Hermes dan lainnya) yang bekerja di proyek ini.

---

## Stack

| Komponen | Teknologi | Catatan |
|----------|-----------|---------|
| Bahasa | **Go 1.26** | Gak ada Node, gak ada PHP, gak ada Ruby. Python hanya di automation opsional (Hermes cron), bukan runtime utama |
| HTTP Router | `github.com/go-chi/chi/v5` | Bukan gin, bukan echo |
| Template | `github.com/a-h/templ` | **Bukan** html/template standar. Compile dulu pake `make templ-gen` |
| Database | `modernc.org/sqlite` + `github.com/jmoiron/sqlx` | Pure Go, no CGO |
| Markdown | `github.com/yuin/goldmark` | GFM enabled |
| CSS | Tailwind | Standalone CLI via `npx tailwindcss`; scan `internal/view/**/*.templ`, `internal/view/**/*.go`, dan handler Go |
| Interactivity | HTMX via CDN | Gak ada JavaScript framework |

## Arsitektur Layer (WAJIB dipatuhi)

```
Handler → Service → Repo → SQLite
```

1. **Handler** (`internal/handler/`) — parse HTTP request, panggil service, render Templ
2. **Service** (`internal/service/`) — business logic. **Gak boleh tau soal HTTP, request, response**
3. **Repo** (`internal/repo/`) — database queries. **Cuma SQL. Gak ada logic**
4. **Model** (`internal/model/`) — struct definitions. Gak ada method bisnis

### Aturan layer:
- ❌ Handler gak boleh panggil repo langsung
- ❌ Service gak boleh akses `r *http.Request` atau `w http.ResponseWriter`
- ❌ Repo gak boleh ada logic (if/else) — cuma query
- ✅ Handler → Service → Repo (one direction)

## Convention & Style

### File naming
- Go files: `snake_case.go`
- Templ files: `snake_case.templ`
- Semua file handler di `internal/handler/`
- Semua file service di `internal/service/`

### Error handling
- Error dari service: `fmt.Errorf("context: %w", err)` — wrap with context
- Error handler: return HTTP 4xx/5xx via helper function
- Jangan panic kecuali di `main.go` untuk fatal startup error

### Routing
- Semua route didefinisikan di `internal/handler/routes.go`
- Route grouping via Chi: `r.Route("/api/v1", ...)`
- Nama path pake `{param}` (Chi format)

### Database
- Migrasi otomatis di `repo.AutoMigrate()` — jalan tiap startup
- Jangan drop table di migration — cuma CREATE IF NOT EXISTS
- SQLite write lock: `db.SetMaxOpenConns(1)` — jangan diubah

### Templating
- Template di-compile: `make templ-gen` atau `go run github.com/a-h/templ/cmd/templ@v0.3.906 generate` sebelum build
- Komponen reusable di `internal/view/components/`
- Layout base di `internal/view/layouts/base.templ`
- Halaman di `internal/view/pages/`

## Membuat Post Baru via API

Endpoint: `POST /api/v1/posts`

```json
{
  "category_slug": "ai-agent",
  "title": "Judul Artikel",
  "content_md": "# Markdown content...",
  "excerpt": "Ringkasan 1-2 kalimat",
  "published": true
}
```

Header: `Authorization: Bearer <RAEVTAR_ADMIN_KEY>`

## Hermes Integration Notes

- Cronjob auto-post: Hermes langsung `curl` ke API localhost. Atau pake `cron/auto_post.sh` kalo mau standalone
- Untuk nambah server monitoring: register server di admin, rotate/copy agent token, lalu jalankan `/static/agent/raevtar-agent.sh` di perangkat target dengan `RAEVTAR_URL`, `RAEVTAR_SERVER_ID`, dan `RAEVTAR_AGENT_TOKEN`
- Gw bisa manual nulis artikel: kasih gw link/topik → gw riset → gw POST ke API

## Build & Run

```bash
# Development (butuh entr)
make dev

# Production build
make build

# Run
./raevtar

# Generate templ
make templ-gen

# Reset database
make db-reset
```

## Environment Variables

| Variable | Default | Wajib |
|----------|---------|-------|
| `RAEVTAR_ADDR` | `:8080` | Tidak |
| `RAEVTAR_DB` | `~/.raevtar/data.db` | Tidak |
| `RAEVTAR_DOMAIN` | `raevtar.tech` | Tidak |
| `RAEVTAR_LOG_LEVEL` | `info` | Tidak |
| `RAEVTAR_ADMIN_KEY` | `""` | **Ya** (untuk endpoint API auth-protected, constant-time validated) |
| `RAEVTAR_ADMIN_USER` | `admin` | Tidak |
| `RAEVTAR_ADMIN_PASS` | `""` | **Ya** (untuk admin panel login) |
| `RAEVTAR_ENV` | `""` | Set `production` untuk strict secret check |
| `RAEVTAR_TRUSTED_PROXY_CIDRS` | `""` | Opsional, CIDR proxy tepercaya untuk `CF-Connecting-IP` |

## Ketika Lo Bingung

1. Cek ARCHITECTURE.md — arsitektur lengkap
2. Cek PRD.md — fitur apa aja yg harus ada
3. Cek ROADMAP.md — apa yg belum dikerjain
4. Jangan nembus layer: Handler → Service → Repo
5. Jangan nambah dependency tanpa alasan jelas
