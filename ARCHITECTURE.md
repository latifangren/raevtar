# RAEVTAR — Arsitektur

```
raevtar.tech — blog, dashboard, lab, API, landing page
Single binary. Go + Templ + HTMX + SQLite.
```

---

## Prinsip

1. **Satu bahasa, satu runtime** — Go. Gak ada Node, gak ada Python, gak ada PHP.
2. **Single binary** — `make build` → `./raevtar` → langsung jalan. Templ/Tailwind cuma build-time.
3. **Separation of concerns** — handler jangan ngotak-atik DB langsung. Model jangan tau soal HTTP.
4. **Backend process stateless** — persistent state di SQLite. Restart server = gak ada data ilang.
5. **Progressive enhancement** — HTML dikirim dari server (SSR). HTMX untuk interaktivitas tanpa JS berat. API untuk akses dari luar.

---

## Layer Architecture

```
                    Request
                       │
                       ▼
              ┌─────────────────┐
              │   HTTP Router    │  net/http + Chi router
              │   /cmd/server/   │
              └────────┬────────┘
                       │
          ┌────────────┼────────────┐
          ▼            ▼            ▼
   ┌──────────┐ ┌──────────┐ ┌──────────┐
   │  Handler  │ │  Handler  │ │  Handler  │  internal/handler/
   │  /blog    │ │ /dashboard│ │ /api/v1   │  → parse request, call service, render response
   └─────┬─────┘ └─────┬─────┘ └─────┬─────┘
          │             │             │
          ▼             ▼             ▼
   ┌────────────────────────────────────┐
   │           Service Layer            │  internal/service/
   │  → business logic, validasi, auth  │  → gak tau soal HTTP
   └──────────────┬─────────────────────┘
                  │
                  ▼
   ┌────────────────────────────────────┐
   │         Repository Layer           │  internal/repo/
   │  → database operations (SQLite)    │  → cuma CRUD, gak ada logic
   └──────────────┬─────────────────────┘
                  │
                  ▼
   ┌────────────────────────────────────┐
   │         SQLite Database            │  ~/.raevtar/data.db
   │  (via modernc.org/sqlite — no CGO) │
   └────────────────────────────────────┘
```

### Alur request (contoh: buka blog post)

```
Browser → GET /blog/rekomendasi-ai-agent-2026
              │
              ▼
         chi router → handler.BlogDetail(c, "rekomendasi-ai-agent-2026")
              │
              ▼
         service.GetPostBySlug("rekomendasi-ai-agent-2026")
              │
              ▼
         repo.GetPostBySlug("rekomendasi-ai-agent-2026")
              │
              ▼
         SQLite query → struct Post
              │
              ▼
         Templ render → HTML
              │
              ▼
         Response ke browser (200 OK, Content-Type: text/html)
```

---

## Struktur Folder

```
raevtar/
├── cmd/
│   └── server/
│       └── main.go              # Entry point: init config, DB, router, start HTTP
│
├── internal/
│   ├── config/
│   │   └── config.go            # Struct + loader dari env
│   │
│   ├── model/
│   │   ├── post.go              # Post struct — blog article
│   │   ├── category.go          # Category struct
│   │   ├── server.go            # Server struct — monitoring target
│   │   ├── server_metric.go     # Metrics history (CPU, RAM, uptime)
│   │   ├── tag.go               # Normalized blog tags
│   │   ├── user.go              # Admin users + RBAC roles
│   │   └── audit.go             # Admin audit log
│   │
│   ├── repo/
│   │   ├── db.go                # Init SQLite connection + migrations auto-run
│   │   ├── post_repo.go         # CRUD posts
│   │   ├── category_repo.go     # CRUD categories
│   │   ├── server_repo.go       # CRUD servers
│   │   ├── metric_repo.go       # Insert/query server metrics
│   │   ├── tag_repo.go          # Tags + post_tags join table
│   │   ├── user_repo.go         # Admin users
│   │   └── audit_repo.go        # Audit log queries
│   │
│   ├── service/
│   │   ├── blog.go              # Blog logic: slug generation, markdown render, pagination
│   │   ├── monitor.go           # Server monitoring: agent tokens + metrics recording
│   │   ├── admin.go             # Admin auth/users/audit boundary
│   │   └── seed.go              # Seed initial data (default categories, admin user)
│   │
│   ├── handler/
│   │   ├── routes.go            # Route mounting (Chi router)
│   │   ├── handlers.go          # Public page handlers render templ pages
│   │   ├── render.go            # templ.Component HTML response helper
│   │   ├── admin.go             # Admin panel handlers render templ views
│   │   ├── auth.go              # API key + admin session auth
│   │   ├── api.go               # REST API handlers (JSON)
│   │   └── rss.go               # RSS feed
│   │
│   └── view/
│       ├── layouts/
│       │   └── base.templ       # HTML shell: <head>, nav, footer, CSS, HTMX CDN
│       ├── pages/
│       │   ├── index.templ       # Landing page
│       │   ├── blog_list.templ   # Blog listing dengan filter kategori
│       │   ├── blog_post.templ   # Single post (render markdown → HTML)
│       │   ├── lab.templ         # Public-safe aggregate lab page
│       │   ├── dashboard.templ   # Dashboard overview (HTMX auto-refresh)
│       │   ├── server_detail.templ # Detail satu server
│       │   └── not_found.templ   # Custom 404 page
│       ├── admin/
│       │   ├── layout.templ      # Admin shell + login page
│       │   ├── pages.templ       # Admin dashboard/posts/servers/users/audit pages
│       │   └── data.go           # Admin view data + presentation helpers
│       └── components/
│           ├── nav.templ         # Navigasi bar
│           ├── post_card.templ   # Card ringkasan post (reusable)
│           ├── server_card.templ # Card status server (reusable)
│           └── pagination.templ  # Pagination component
│
├── static/
│   └── css/
│       └── style.css            # Tailwind output (kalo pake standalone CLI)
│
├── migrations/
│   └── 001_init.sql             # SQL init: create tables
│
├── cron/
│   └── auto_post.py             # Script cron Hermes buat nulis artikel
│
├── go.mod                       # Go module definition
├── go.sum
├── Makefile                     # Build, run, migrate, seed commands
└── .env.example                 # Contoh konfigurasi
```

---

## Data Model (ERD)

```
┌────────────┐     ┌──────────────────┐
│  categories │     │      posts       │
├────────────┤     ├──────────────────┤
│ id (PK)    │◄────│ id (PK)          │
│ slug       │     │ category_id (FK) │
│ name       │     │ title            │
│ description│     │ slug (UNIQUE)    │
│ created_at │     │ content_md       │  ←  markdown
│ updated_at │     │ excerpt          │  ←  ringkasan
└────────────┘     │ published        │  ←  boolean
                   │ created_at       │
                   │ updated_at       │
                   └──────────────────┘

┌────────────┐     ┌──────────────────┐
│  servers   │     │ server_metrics   │
├────────────┤     ├──────────────────┤
│ id (PK)    │◄────│ id (PK)          │
│ name       │     │ server_id (FK)   │
│ host       │     │ cpu_percent      │
│ port       │     │ ram_used_mb      │
│ tags       │     │ ram_total_mb     │
│ token_hash │     │ disk_used_gb     │
│ last_seen  │     │ uptime_seconds   │
│ created_at │     │ online           │
└────────────┘     │ recorded_at      │
                   └──────────────────┘

┌────────────┐     ┌──────────────────┐
│    tags    │     │    post_tags     │
├────────────┤     ├──────────────────┤
│ id (PK)    │◄────│ tag_id (FK)      │
│ name       │     │ post_id (FK)     │────► posts.id
│ slug       │     └──────────────────┘
│ created_at │
└────────────┘

┌────────────┐     ┌──────────────────┐
│   users    │     │    audit_logs    │
├────────────┤     ├──────────────────┤
│ id (PK)    │     │ id (PK)          │
│ username   │     │ user             │
│ role       │     │ action           │
│ display_name     │ details          │
│ created_at │     │ ip               │
│ updated_at │     │ created_at       │
└────────────┘     └──────────────────┘
```

---

## Routing Map

| Method | Path | Handler | Keterangan |
|--------|------|---------|------------|
| GET | `/` | landing.Index | Landing page |
| GET | `/blog` | blog.List | Blog list (semua) |
| GET | `/blog?category=ai-agent` | blog.List | Filter by kategori |
| GET | `/blog/{slug}` | blog.Detail | Single post |
| GET | `/blog/feed.xml` | rss.Feed | RSS feed |
| GET | `/lab` | lab.Page | Public-safe aggregate lab page |
| GET | `/dashboard` | dashboard.Index | Server monitoring |
| GET | `/dashboard/{serverID}` | dashboard.Detail | Detail server |
| GET | `/dashboard/{serverID}/live` | dashboard.DetailLive | HTMX fragment detail server |
| GET | `/admin/*` | admin.* | Admin panel session auth |
| GET | `/api/v1/posts` | api.ListPosts | JSON posts |
| POST | `/api/v1/posts` | api.CreatePost | JSON create (cron) |
| GET | `/api/v1/categories` | api.ListCategories | JSON categories |
| GET | `/api/v1/hoststats` | api.HostStats | Host CPU/RAM/disk/temp (Bearer auth) |
| GET | `/api/v1/servers` | api.ListServers | JSON server status (Bearer auth) |
| POST | `/api/v1/servers` | api.CreateServer | Register server + return one-time agent token |
| GET | `/api/v1/servers/{id}` | api.GetServer | JSON detail server (Bearer auth) |
| POST | `/api/v1/servers/{id}/ping` | api.RecordMetrics | Record server metrics (agent token atau admin key) |
| GET | `/docs` | public docs page | Public-safe docs untuk read-only API, route map, dan privacy boundary |
| GET | `/lab/docs` | public docs page | Alias docs dari area lab |

Public docs dirender via Templ. `static/openapi.json` tetap tersedia sebagai public read-only spec dan sengaja tidak mendokumentasikan endpoint admin/server/agent.

---

## Bagaimana Hermes (gw) Interact

### Hermes Cronjob — Auto Blog Post

1. Lo jalanin: `hermes cron create --schedule "0 8 * * *" --prompt "..." --skills blog-autopost`
2. Tiap jam 8 pagi, gw riset, nulis artikel markdown, POST ke `http://localhost:8080/api/v1/posts`
3. Artikel masuk DB → muncul di blog

### Hermes Manual — Lo suruh gw nulis

```
> [Latifan🐾] nih gw nemu projek keren, tulis aja di blog: <link>
> [Hermes] *curl POST /api/v1/posts* — jadi, langsung muncul
```

### Server Monitoring — Agent push

- Setiap mesin target jalanin **script kecil** dari `/static/agent/raevtar-agent.sh`
- Agent push metrics ke `${RAEVTAR_URL}/api/v1/servers/{id}/ping` tiap 1 menit
- `RAEVTAR_URL` bisa domain publik, LAN IP, hostname lokal, atau tunnel
- Auth pakai token per server; token didapat saat register via admin/API, dan bisa di-rotate dari `/admin/servers`
- Raevtar tidak perlu SSH user/password ke perangkat target

---

## Tech Detail

| Komponen | Pilihan | Alasan |
|----------|---------|--------|
| **HTTP Router** | `github.com/go-chi/chi/v5` | Ringan, idiomatic Go, middleware built-in |
| **Templating** | `github.com/a-h/templ` | Type-safe, compile-time checked |
| **SQLite** | `modernc.org/sqlite` | Pure Go, gak perlu CGO, gak perlu gcc |
| **ORM/Query** | `github.com/jmoiron/sqlx` | Ringan, tetap SQL mentah tanpa abstraction layer gede |
| **Markdown** | `github.com/yuin/goldmark` | Standar, extensible |
| **Tailwind** | Standalone CLI via `npx tailwindcss` | Scan templ files, view helper Go, and handler Go |
| **Config** | Environment variables + `.env` | Standard 12-factor |

---

## Keunggulan Arsitektur Ini

- **Testable** — handler > service > repo, tiap layer bisa di-test sendiri
- **Expandable** — mau tambah fitur baru (comments, newsletter, webhook)? Tinggal tambah handler + service, gak perlu rombak struktuk
- **Portable** — binary bisa di-copy ke server lain, laptop, VPS, jalan sama persis
- **Lo-fi** — gak perlu Docker, gak perlu k8s, gak perlu CI/CD pipeline. Cuma `make build && ./raevtar`
