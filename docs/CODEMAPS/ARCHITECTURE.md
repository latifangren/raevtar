# Architecture Codemap

**Last Updated:** 2026-06-22

## System Overview

Raevtar is a single Go binary serving blog content, server monitoring dashboards, project portfolios, and automation features. It uses SQLite as its database, Chi router for HTTP routing, and Templ for server-rendered HTML templates.

```
┌─────────────────────────────────────────────────────────┐
│                    HTTP Server (Chi)                     │
│  :8080                                                   │
├─────────────────────────────────────────────────────────┤
│                                                          │
│  ┌────────────┐  ┌──────────────┐  ┌──────────────────┐ │
│  │  Public    │  │   Admin      │  │   API v1         │ │
│  │  Pages    │  │   Panel      │  │   (JSON)         │ │
│  └─────┬──────┘  └──────┬───────┘  └───────┬──────────┘ │
│        │                │                  │            │
│        └────────┬───────┴──────────┬───────┘            │
│                 │                  │                     │
│        ┌────────▼──────────────────▼────────┐            │
│        │         Handler Layer               │           │
│        │  (HTTP parsing, auth, rendering)    │           │
│        └────────────────┬───────────────────┘            │
│                         │                                │
│        ┌────────────────▼───────────────────┐            │
│        │         Service Layer              │            │
│        │  (Business logic, validation)      │            │
│        └────────────────┬───────────────────┘            │
│                         │                                │
│        ┌────────────────▼───────────────────┐            │
│        │          Repo Layer                │            │
│        │   (SQL queries, no business logic) │            │
│        └────────────────┬───────────────────┘            │
│                         │                                │
│        ┌────────────────▼───────────────────┐            │
│        │          SQLite Database            │            │
│        │   (modernc.org/sqlite, no CGO)      │            │
│        └────────────────────────────────────┘            │
│                                                          │
└─────────────────────────────────────────────────────────┘
```

## Layer Architecture (Handler → Service → Repo)

### Handler Layer (`internal/handler/`)
- Parses HTTP requests, calls services, renders Templ components
- Handles auth (session-based admin, API key admin auth)
- Rate limiting, security headers, body size limits
- Route definitions in `routes.go`

### Service Layer (`internal/service/`)
- Business logic only — no HTTP awareness
- Markdown rendering (goldmark with GFM)
- Validation, slug generation, data transformations
- Webhook firing, command queue processing

### Repo Layer (`internal/repo/`)
- SQL queries only — no conditional business logic
- All repos aggregated in `Repositories` struct
- Auto-migration on startup (CREATE IF NOT EXISTS)

## Data Flow

```
User Request → Chi Router → Middleware (RateLimit, Logger, Recoverer)
  → Handler method → Service method → Repo method → SQLite
  ← JSON/HTML response ← Templ component ← Service result
```

## Key Architecture Decisions

- **Single binary** — no external runtime dependencies beyond SQLite
- **CGO-free SQLite** via `modernc.org/sqlite`
- **Server-rendered HTML** via Templ + HTMX for interactivity
- **Session auth** for admin panel (in-memory map)
- **Bearer token auth** for API write endpoints
- **Live server monitoring** via agent poll/ping model
- **Auto-migration** — schema applied at startup, no migration tooling
- **Webhook system** for server event notifications
- **Editorial inbox** for scheduled/automated content publishing

## Route Groups

| Group | Prefix | Auth | Purpose |
|-------|--------|------|---------|
| Public Pages | `/`, `/blog`, `/projects`, etc. | None | Content delivery |
| Dashboard | `/dashboard` | None | Public server monitoring |
| Admin Panel | `/admin` | Session | Content & system management |
| API v1 | `/api/v1` | Bearer (write) | Programmatic access |
| RSS Feed | `/blog/feed.xml` | None | Blog syndication |
| Static | `/static/*` | None | CSS, JS, assets |

## External Dependencies

| Package | Purpose | Version |
|---------|---------|---------|
| `github.com/go-chi/chi/v5` | HTTP router | v5.3.0 |
| `github.com/a-h/templ` | HTML templating | v0.3.906 |
| `github.com/jmoiron/sqlx` | SQL extensions | v1.4.0 |
| `github.com/yuin/goldmark` | Markdown rendering | v1.8.2 |
| `modernc.org/sqlite` | CGO-free SQLite | v1.51.0 |
| `golang.org/x/crypto` | bcrypt password hashing | v0.52.0 |

## Related Areas

- [Modules Codemap](MODULES.md) — detailed module descriptions
- [Files Codemap](FILES.md) — directory structure and file purposes
