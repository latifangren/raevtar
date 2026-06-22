# RAEVTAR — Arsitektur

```
raevtar.tech — blog, dashboard, lab, API, landing page
Single binary. Go + Templ + self-hosted HTMX + SQLite.
```

---

## Prinsip

1. **Satu bahasa, satu runtime** — Go. Gak ada Node, gak ada Python, gak ada PHP.
2. **Single binary** — `make build` → `./raevtar` → langsung jalan. Templ/Tailwind cuma build-time.
3. **Separation of concerns** — handler jangan ngotak-atik DB langsung. Model jangan tau soal HTTP.
4. **Backend process stateless** — persistent state di SQLite. Restart server = gak ada data ilang.
5. **Progressive enhancement** — HTML dikirim dari server (SSR). HTMX untuk interaktivitas tanpa JS berat. API untuk akses dari luar.

Observed di deployment whyred saat idle, keputusan arsitektur ini memang menghasilkan footprint kecil: sekitar **24 MB RSS**, CPU idle nyaris nol, binary sekitar **17.5 MB**, dan SQLite database masih sub-1 MB pada pemakaian saat ini. Virtual memory process Go bisa tampak jauh lebih besar karena mapping runtime, tapi itu bukan angka resident memory yang benar-benar dipakai sehari-hari.

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
│   │   ├── project.go           # Project struct + updates/relations/showcase
│   │   ├── server.go            # Server struct — monitoring target
│   │   ├── server_metric.go     # Metrics history (CPU/load/RAM/disk/temp/uptime)
│   │   ├── server_command.go    # Server command queue (pending/running/completed/failed)
│   │   ├── tag.go               # Normalized blog tags
│   │   ├── user.go              # Admin users + RBAC roles
│   │   ├── audit.go             # Admin audit log
│   │   ├── webhook.go           # Webhook config + event log
│   │   ├── seo.go               # SEO data struct (description, canonical, JSON-LD)
│   │   ├── page_content.go      # Managed page content (about, contact)
│   │   ├── editorial_inbox.go   # Editorial inbox item + lifecycle
│   │   └── media.go             # Media asset struct
│   │
│   ├── repo/
│   │   ├── db.go                # Init SQLite connection + migrations auto-run
│   │   ├── post_repo.go         # CRUD posts
│   │   ├── category_repo.go     # CRUD categories
│   │   ├── project_repo.go      # CRUD projects + updates/relations/showcase
│   │   ├── server_repo.go       # CRUD servers
│   │   ├── metric_repo.go       # Insert/query server metrics
│   │   ├── command_repo.go      # Server command queue CRUD
│   │   ├── tag_repo.go          # Tags + post_tags join table
│   │   ├── user_repo.go         # Admin users
│   │   ├── audit_repo.go        # Audit log queries
│   │   ├── webhook_repo.go      # Webhook configs + events
│   │   ├── view_repo.go         # Post view tracking by IP hash
│   │   ├── page_content_repo.go # Managed page content
│   │   ├── editorial_inbox_repo.go # Editorial inbox CRUD
│   │   └── media_repo.go        # Media asset metadata
│   │
│   ├── service/
│   │   ├── blog.go              # Blog logic: slug, markdown render, pagination, view recording
│   │   ├── project.go           # Project logic: lifecycle, timeline, relations, showcase
│   │   ├── search.go            # Full-text search across posts/projects/pages
│   │   ├── site_meta.go         # SEO: canonical URLs, JSON-LD, sitemap, LLMs.txt, OG images
│   │   ├── page_content.go      # Managed page content (about, contact)
│   │   ├── editorial_inbox.go   # Editorial inbox lifecycle + claim + fairness policy
│   │   ├── monitor.go           # Server monitoring: agent tokens, metrics recording, webhook alerts
│   │   ├── admin.go             # Admin auth/users/audit boundary
│   │   ├── command_queue.go     # Server command queue lifecycle
│   │   ├── webhook.go           # Webhook config + threshold evaluation + fire
│   │   ├── media.go             # Media upload/storage
│   │   └── seed.go              # Seed initial data (default categories, admin user)
│   │
│   ├── handler/
│   │   ├── routes.go            # Route mounting (Chi router)
│   │   ├── handlers.go          # Public page handlers render templ pages
│   │   ├── render.go            # templ.Component HTML response helper
│   │   ├── admin.go             # Admin panel handlers render templ views
│   │   ├── auth.go              # API key + admin session auth + CSRF
│   │   ├── api.go               # REST API handlers (JSON)
│   │   ├── hardening.go         # Request caps, generic 500s, login throttle
│   │   ├── security.go          # Security headers + CSP + trusted proxy CIDR
│   │   ├── rss.go               # RSS feed
│   │   └── og_image.go          # Dynamic SVG OG images for blog posts + projects
│   │
│   └── view/
│       ├── layouts/
│       │   └── base.templ       # HTML shell: <head>, nav, footer, CSS, self-hosted HTMX, SEO meta
│       ├── pages/
│       │   ├── index.templ       # Landing page (featured projects, recent posts, server health)
│       │   ├── blog_list.templ   # Blog listing dengan filter kategori + HTMX partial
│       │   ├── blog_post.templ   # Single post (render markdown → HTML)
│       │   ├── projects.templ    # Project archive with filters
│       │   ├── project_detail.templ # Project detail + timeline/changelog/showcase
│       │   ├── search.templ      # Search page (HTMX-powered partial results)
│       │   ├── lab.templ         # Public-safe aggregate lab page
│       │   ├── dashboard.templ   # Dashboard overview (HTMX auto-refresh)
│       │   ├── server_detail.templ # Detail satu server + chart extras
│       │   ├── topics.templ      # Topic/index switchboard
│       │   ├── about.templ       # About page (managed content)
│       │   ├── contact.templ     # Contact page (managed content)
│       │   ├── docs.templ        # Public-safe docs page
│       │   └── not_found.templ   # Custom 404 page
│       ├── admin/
│       │   ├── layout.templ      # Admin shell + login page
│       │   ├── pages.templ       # Admin dashboard/posts/servers/users/audit pages
│       │   ├── server_detail.templ # Admin server detail + commands + metrics
│       │   ├── webhooks.templ    # Admin webhook config management
│       │   └── data.go           # Admin view data + presentation helpers (722 lines)
│       └── components/
│           ├── nav.templ         # Navigasi bar
│           ├── footer.templ      # Footer
│           ├── post_card.templ   # Card ringkasan post (reusable)
│           ├── project_card.templ # Card ringkasan project (reusable)
│           ├── server_card.templ # Card status server (reusable)
│           └── pagination.templ  # Pagination component
│
├── static/
│   ├── agent/
│   │   └── raevtar-agent.sh     # Lightweight push telemetry agent
│   ├── css/
│   │   └── style.css            # Tailwind output
│   └── js/
│       ├── htmx.min.js          # Vendored HTMX
│       └── raevtar-ui.js        # Local UI behavior, CSP-safe
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
                   │ cover_image_url  │
                   │ created_at       │
                   │ updated_at       │
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
│ display_name│     │ details          │
│ created_at │     │ ip               │
│ updated_at │     │ created_at       │
└────────────┘     └──────────────────┘

┌────────────┐     ┌──────────────────┐
│  servers   │     │ server_metrics   │
├────────────┤     ├──────────────────┤
│ id (PK)    │◄────│ id (PK)          │
│ name       │     │ server_id (FK)   │
│ host       │     │ cpu_percent      │
│ port       │     │ cpu_load_1/5/15  │
│ tags       │     │ cpu_cores        │
│ token_hash │     │ ram_used_mb      │
│ last_seen  │     │ ram_total_mb     │
│ created_at │     │ disk_used_gb     │
└────────────┘     │ disk_total_gb    │
                   │ temperature_c    │
       │           │ uptime_seconds   │
       │           │ online           │
       ▼           │ recorded_at      │
┌────────────┐     └──────────────────┘
│server_cmds │
├────────────┤     ┌──────────────────┐
│ id (PK)    │     │   projects       │
│ server_id  │     ├──────────────────┤
│ command    │     │ id (PK)          │
│ status     │     │ slug (UNIQUE)    │
│ payload    │     │ title            │
│ result     │     │ excerpt          │
│ queued_at  │     │ content_md       │
│ started_at │     │ state            │
│ completed  │     │ featured         │
└────────────┘     │ cover_image_url  │
                   │ sort_order       │
                   │ published        │
                   └──────────────────┘

┌──────────────┐       │
│webhook_cfgs  │       │
├──────────────┤       ▼
│ id (PK)      │  ┌──────────────────┐
│ name         │  │  webhook_events  │
│ url          │  ├──────────────────┤
│ secret       │  │ id (PK)          │
│ enabled      │  │ webhook_id (FK)  │
│ created_at   │  │ event_type       │
└──────────────┘  │ server_id        │
                  │ payload          │
                  │ response_code    │
                  │ response_body    │
                  │ fired_at         │
                  └──────────────────┘

┌──────────────┐  ┌──────────────────┐
│   post_views  │  │ page_content     │
├──────────────┤  ├──────────────────┤
│ id (PK)      │  │ page_key (PK)    │
│ post_id (FK) │  │ title            │
│ ip_hash      │  │ content_md       │
│ viewed_at    │  │ updated_at       │
└──────────────┘  └──────────────────┘
```

---

## Routing Map

| Method | Path | Handler | Keterangan |
|--------|------|---------|------------|
| GET | `/` | landing.Index | Landing page (featured projects, recent posts, server health) |
| GET | `/about` | about.Page | About page (managed content) |
| GET | `/blog` | blog.List | Blog list (semua) |
| GET | `/blog?category=ai-agent` | blog.List | Filter by kategori/topic |
| GET | `/blog/{slug}` | blog.Detail | Single post |
| GET | `/contact` | contact.Page | Contact page (managed content) |
| GET | `/topics` | topics.Page | Public topic index / switchboard |
| GET | `/blog/feed.xml` | rss.Feed | RSS feed |
| GET | `/search` | search.Page | Public search page (HTMX partial) |
| GET | `/lab` | lab.Page | Public-safe aggregate lab page |
| GET | `/lab/node-status/{name}` | lab.NodeStatus | Node status shortcode (inline HTML embed) |
| GET | `/projects` | projects.Page | Project archive with featured/state/sort filters |
| GET | `/projects/{slug}` | projects.Detail | Project detail + timeline/changelog/relations/showcase |
| GET | `/projects/{slug}/changelog` | projects.Changelog | Project changelog page |
| GET | `/dashboard` | dashboard.Index | Public-safe server monitoring + Platform System Health |
| GET | `/dashboard/{serverID}` | dashboard.Detail | Public-safe detail server |
| GET | `/dashboard/{serverID}/live` | dashboard.DetailLive | HTMX fragment detail server, refresh 15s |
| GET | `/docs` | public docs page | Public-safe docs untuk read-only API, route map, dan privacy boundary |
| GET | `/lab/docs` | public docs page | Alias docs dari area lab |
| GET | `/sitemap.xml` | meta.Sitemap | XML sitemap |
| GET | `/llms.txt` | meta.LLMsTxt | LLM discovery text |
| GET | `/robots.txt` | meta.Robots | Allow all robots |
| GET | `/favicon.svg` | meta.Favicon | SVG favicon |
| GET | `/uploads/{filename}` | media.Serve | Serve uploaded media |
| GET | `/og-image/blog/{slug}` | og.Blog | Dynamic SVG OG image for blog post |
| GET | `/og-image/project/{slug}` | og.Project | Dynamic SVG OG image for project |
| | | | |
| **Admin panel** (session auth) | | | |
| GET/POST | `/admin/login` | admin.Login | Login page + POST login |
| GET | `/admin/` | admin.Dashboard | Admin dashboard (stats) |
| GET/POST | `/admin/editorial-inbox` | admin.Editorial | Manage editorial inbox |
| GET/POST | `/admin/posts` | admin.Posts | Manage posts + preview |
| GET/POST | `/admin/topics` | admin.Topics | Manage blog topics/categories |
| GET/POST | `/admin/projects` | admin.Projects | Manage projects + updates/relations/showcase |
| GET/POST | `/admin/pages` | admin.Pages | Manage static pages (about, contact) |
| GET/POST | `/admin/media` | admin.Media | Manage media uploads |
| GET/POST | `/admin/servers` | admin.Servers | Manage servers + detail + commands + token rotate |
| GET/POST | `/admin/webhooks` | admin.Webhooks | Manage webhook configurations |
| GET | `/admin/audit-log` | admin.Audit | Audit log |
| GET/POST | `/admin/manage-users` | admin.Users | Manage admin users |
| | | | |
| **API v1** (JSON) | | | |
| GET | `/api/v1/posts` | api.ListPosts | Public list posts |
| POST | `/api/v1/posts` | api.CreatePost | Create post (admin key) |
| GET | `/api/v1/search` | api.Search | Public search (q, scope, page, page_size) |
| GET | `/api/v1/projects` | api.ListProjects | Public list projects (featured, state, sort) |
| GET | `/api/v1/projects/{slug}/updates` | api.ListProjectUpdates | Public timeline |
| GET | `/api/v1/projects/{slug}/changelog` | api.ListProjectChangelog | Public changelog |
| GET | `/api/v1/projects/{slug}/relations` | api.ListProjectRelations | Public related content |
| GET | `/api/v1/projects/{slug}/showcase` | api.ListProjectShowcase | Public showcase |
| POST/PUT/DELETE | `/api/v1/projects[/...]` | api.ProjectCRUD | Project child resource CRUD (admin key) |
| GET | `/api/v1/categories` | api.ListCategories | Public categories |
| GET | `/api/v1/servers` | api.ListServers | Server list (admin key) |
| POST | `/api/v1/servers` | api.CreateServer | Register + return one-time token (admin key) |
| GET | `/api/v1/servers/{id}` | api.GetServer | Server detail (admin key) |
| POST | `/api/v1/servers/{id}/ping` | api.RecordMetrics | Agent report metrics (agent token or admin key) |
| GET | `/api/v1/servers/{id}/commands` | api.PendingCommands | Agent poll pending commands |
| POST | `/api/v1/servers/{id}/commands/result` | api.ReportCommandResult | Agent report command result |
| GET | `/api/v1/hoststats` | api.HostStats | Host resource snapshot (admin key) |
| GET/POST | `/api/v1/editorial-inbox[/...]` | api.Editorial | Editorial inbox CRUD/claim/complete/fail (admin key) |

Public docs dirender via Templ. `static/openapi.json` tetap tersedia sebagai public read-only spec dan sengaja tidak mendokumentasikan endpoint admin/server/agent setup.

`categories` adalah source of truth untuk topic blog. Public `/blog` dan `/topics` membaca entity yang sama dengan admin topic management di `/admin/topics`. Karena `posts` menyimpan `category_id`, service layer memblok slug change dan delete saat category masih dipakai post, supaya link/filter publik tidak rusak.

Public dashboard/detail boleh menampilkan resource summary yang sudah dibulatkan dan aman: CPU, load, RAM, disk, temperature, uptime, latest sample age, sample count, history window, dan availability aggregate. Host/IP, port, tags privat, agent token, install command, setup command, dan audit log tetap admin-only.

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
- Payload mencakup CPU %, load 1/5/15, cores, RAM used/total, disk used/total, uptime, online flag, dan temperature jika sensor tersedia
- `RAEVTAR_URL` bisa domain publik, LAN IP, hostname lokal, atau tunnel
- Auth pakai token per server; token didapat saat register via admin/API, dan bisa di-rotate dari `/admin/servers`
- Raevtar tidak perlu SSH user/password ke perangkat target

---

## Hardening Boundary

- `RAEVTAR_ENV=production` menolak startup kalau admin key/password kosong.
- API key comparison constant-time.
- Rate limiting global in-memory: 60 request/menit per IP.
- Admin login throttling in-memory: per `IP + username` dan IP-only spray guard.
- Request body cap dipasang untuk login, JSON API, form admin, dan media upload.
- Internal `500` dikembalikan sebagai `internal server error`; detail masuk log server.
- CSP memakai `script-src 'self'`; HTMX disajikan dari `/static/js/htmx.min.js`, bukan CDN.
- `RAEVTAR_TRUSTED_PROXY_CIDRS` hanya untuk proxy tepercaya; default mengabaikan forwarded IP header.

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
| **HTMX** | Self-hosted `/static/js/htmx.min.js` | Interaktivitas ringan tanpa CDN runtime |
| **Config** | Environment variables + `.env` | Standard 12-factor |

---

## Keunggulan Arsitektur Ini

- **Testable** — handler > service > repo, tiap layer bisa di-test sendiri
- **Expandable** — mau tambah fitur baru (comments, newsletter, webhook)? Tinggal tambah handler + service, gak perlu rombak struktuk
- **Portable** — binary bisa di-copy ke server lain, laptop, VPS, jalan sama persis
- **Lo-fi** — gak perlu Docker, gak perlu k8s, gak perlu CI/CD pipeline. Cuma `make build && ./raevtar`
