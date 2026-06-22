# Modules Codemap

**Last Updated:** 2026-06-22

## Module Overview

```
raevtar/
├── cmd/server/        # Entry point
├── internal/
│   ├── config/        # Configuration management
│   ├── handler/       # HTTP handlers (27 files)
│   ├── service/       # Business logic (16 files)
│   ├── repo/          # Database access (19 files)
│   ├── model/         # Data structures (17 files)
│   └── view/          # Templ templates
│       ├── pages/     # Public page templates
│       ├── admin/     # Admin panel templates
│       ├── layouts/   # Base layout
│       └── components/# Reusable UI components
├── static/            # Static assets
├── cron/              # Script automation
├── docs/              # Documentation
└── migrations/        # SQL migration files
```

---

## Config Module

**Purpose:** Load and validate environment configuration

**Location:** `internal/config/`

**Key Files:**
- `config.go` — Config struct with env var loading, validation
- `cidr.go` — CIDR parsing utilities for trusted proxy config

**Exports:**
- `Config` struct — all app configuration fields
- `Load()` — load config from environment
- `ParseCIDRs()` — parse CIDR string to net.IPNet slice
- `IsTrustedIP()` — check if IP is in trusted CIDR list

**Dependencies:** None (stdlib only)

---

## Handler Module

**Purpose:** HTTP request/response handling, routing, middleware

**Location:** `internal/handler/`

**Key Files:**
- `routes.go` — All route definitions, Chi router setup
- `handlers.go` — Public page handlers (landing, blog, projects, dashboard)
- `admin.go` — Admin panel handlers (CRUD for posts, projects, servers, etc.)
- `api.go` — JSON API v1 handlers (read + admin-protected writes)
- `auth.go` — In-memory session store, login/logout, CSRF
- `security.go` — Security headers middleware, rate limiter, trusted proxies
- `hardening.go` — Request body limits, login throttling, error helpers
- `editorial.go` — Editorial inbox admin handlers
- `hoststats.go` — Local host stats collection (/proc/sysfs)
- `rss.go` — RSS 2.0 feed generation
- `discovery.go` — sitemap.xml and llms.txt generation
- `og_image.go` — SVG-based Open Graph image generation
- `render.go` — HTML render helper (Templ)
- `middleware_test.go` — Middleware tests

**Dependencies:**
- Service layer (`internal/service/`)
- View layer (`internal/view/`)
- Config (`internal/config/`)
- Chi router (`go-chi/chi/v5`)

**Public Routes:**

| Route | Handler | Description |
|-------|---------|-------------|
| `GET /` | `landingIndex` | Landing page with recent posts, servers |
| `GET /about` | `aboutPage` | About page |
| `GET /blog` | `blogList` | Blog list (paginated, filterable) |
| `GET /blog/{slug}` | `blogDetail` | Single blog post |
| `GET /contact` | `contactPage` | Contact page |
| `GET /lab` | `labPage` | Lab landing page |
| `GET /docs` | `docsPage` | Documentation page |
| `GET /projects` | `projectsPage` | Project portfolio |
| `GET /projects/{slug}` | `projectDetail` | Project detail with timeline |
| `GET /projects/{slug}/changelog` | `projectChangelogPage` | Full changelog |
| `GET /search` | `searchPage` | Site search (HTMX partials) |
| `GET /topics` | `topicsPage` | Browse by category |
| `GET /dashboard` | `dashboardIndex` | Server monitoring dashboard |
| `GET /dashboard/{id}` | `dashboardDetail` | Server detail view |
| `GET /dashboard/{id}/live` | `dashboardDetailLive` | Live server view |

**Admin Routes** (under `/admin`, session-protected):
- CRUD for posts, projects, categories, pages, servers, media, webhooks
- Editorial inbox management
- Audit log, user management
- Server command queuing

**API v1 Routes** (under `/api/v1`):
- `GET /posts`, `GET /projects` — public read
- `POST /posts`, `POST /projects` — admin auth (Bearer)
- Full CRUD for projects, updates, relations, showcase
- Editorial inbox API (admin auth)
- Server ping/command endpoints (agent auth)
- Host stats endpoint (admin auth)

---

## Service Module

**Purpose:** Business logic, validation, data transformation

**Location:** `internal/service/`

**Aggregate Service:**
- `service.go` — `Service` struct bundles all sub-services
- `seed.go` — Initial data seeding (categories, admin user, default pages)

### BlogService

**Key Files:** `blog.go`

**Purpose:** Blog post CRUD, markdown rendering, view tracking

**Exports:**
- `ListPosts(category, page, pageSize)` — paginated published posts
- `ListPostsWithOptions(BlogListOptions)` — with query/filter
- `ListAllPosts(page, pageSize)` — all posts (including drafts)
- `GetPublishedPost(slug)` — single published post
- `CreatePost(PostCreate)` — create new post
- `UpdatePost(id, PostUpdate)` — update existing post
- `DeletePost(id)` — delete post
- `RecordPostView(postID, ipHash)` — async view tracking
- `ListCategories()` — all categories
- `CreateCategory()`, `UpdateCategory()`, `DeleteCategory()`
- `RenderMarkdown(md)` — goldmark GFM rendering

### ProjectService

**Key Files:** `project.go`

**Purpose:** Project portfolio CRUD, timeline, changelog, content relations

**Exports:**
- `ListProjects(page, pageSize, opts)` — paginated project list
- `GetPublishedProject(slug)` — single published project
- `CreateProject()` / `UpdateProject()` / `DeleteProject()`
- Project updates CRUD (timeline/build_log/changelog)
- Project content relations CRUD
- Project showcase items CRUD
- `ListProjectTimeline()` / `ListProjectChangelog()`

### MonitorService

**Key Files:** `monitor.go`

**Purpose:** Server monitoring, agent token management, health tracking

**Exports:**
- `CreateServer()`, `CreateServerWithAgentToken()`
- `ListServers()`, `GetServer()`, `GetServerByName()`
- `UpdateServer()`, `DeleteServer()`, `RotateAgentToken()`
- `RecordMetrics()` — ingest agent telemetry
- `GetRecentMetrics()` — recent metric history
- `VerifyAgentToken()` — constant-time token validation
- Integrates with WebhookService for alerting

### SearchService

**Key Files:** `search.go`

**Purpose:** Unified search across posts, projects, and pages

**Exports:**
- `SearchPublic(opts)` — search with scope (all/posts/projects/pages)

**Dependencies:** BlogService, ProjectService, PageContentService

### SiteMetaService

**Key Files:** `site_meta.go`

**Purpose:** SEO metadata, sitemap, llms.txt generation

**Exports:**
- `CanonicalURL(path)` — full URL builder
- `DefaultSEO(path)` / `HomeSEO()` / `BlogPostSEO()` / `ProjectSEO()`
- `SitemapEntries()` — URL list for sitemap.xml
- `LLMSText()` — content for /llms.txt

### PageContentService

**Key Files:** `page_content.go`

**Purpose:** Manage static page content (About, Contact)

**Exports:**
- `ListPages()`, `GetPage(key)`, `UpsertPage()`
- `RenderMarkdown()` — goldmark rendering

### EditorialInboxService

**Key Files:** `editorial_inbox.go`

**Purpose:** Scheduled/automated content publishing queue

**Exports:**
- `CreateInboxItem()`, `ListInboxItems()`, `GetInboxItem()`
- `UpdateInboxItem()`, `DeleteInboxItem()`
- `ClaimInboxItem()` — claim with lease, `CompleteClaim()` / `FailClaim()`
- `CountInboxStatuses()`, `GetInboxSummary()`
- Fairness-aware claiming with autonomous gap detection

### MediaService

**Key Files:** `media.go`

**Purpose:** File upload management (max 5MB per upload)

**Exports:**
- `Upload(file, reader)` — store file, create DB record
- `ListAssets()`, `GetAsset(id)`
- `DeleteAsset(id)` — delete file + DB record
- `FilePath(filename)` — resolve stored path

### AdminService

**Key Files:** `admin.go`

**Purpose:** User management, authentication, auditing

**Exports:**
- `Authenticate()` — verify password + bcrypt
- `ListUsers()`, `CreateUser()`, `DeleteUser()`
- `LogAudit()` — insert audit log entry
- `LogLogout()`, `GetAuditLogs()`

### CommandQueueService

**Key Files:** `command_queue.go`

**Purpose:** Queue server commands for agent execution

**Exports:**
- `QueueCommand(serverID, command, payload)`
- `PendingCommands(serverID)`
- `CompleteCommand(id, result)`, `FailCommand(id, result)`
- `TakeAndRun(id)`, `CommandHistory(serverID, limit)`

### WebhookService

**Key Files:** `webhook.go`

**Purpose:** Outgoing webhook notifications for events

**Exports:**
- `CreateConfig()` / `ListConfigs()` / `UpdateConfig()` / `DeleteConfig()`
- `FireEvent(eventType, serverID, payload)` — HMAC-signed POST
- `ListEvents()` — event history

---

## Repo Module

**Purpose:** Database queries only. No business logic.

**Location:** `internal/repo/`

**Key Files:**
- `db.go` — DB initialization, auto-migration, `Repositories` aggregate
- `post_repo.go` — Post CRUD
- `project_repo.go` — Project CRUD
- `project_update_repo.go` — Project updates CRUD
- `category_repo.go` — Category CRUD
- `server_repo.go` — Server CRUD
- `metric_repo.go` — Server metrics CRUD
- `tag_repo.go` — Tag CRUD + post/project tag associations
- `media_repo.go` — Media asset CRUD
- `user_repo.go` — User CRUD + password hashing
- `audit_repo.go` — Audit log CRUD
- `view_repo.go` — Post view tracking
- `command_repo.go` — Server command CRUD
- `webhook_repo.go` — Webhook config + event CRUD
- `page_content_repo.go` — Page content CRUD
- `editorial_inbox_repo.go` — Editorial inbox CRUD
- `content_relation_repo.go` — Content relations CRUD
- `project_showcase_repo.go` — Showcase items CRUD

**Tables:** categories, posts, projects, project_updates, content_relations, project_showcase_items, servers, server_metrics, tags, post_tags, project_tags, post_views, page_contents, media_assets, users, audit_logs, editorial_inbox, editorial_policy_state, server_commands, webhook_configs, webhook_events (20 tables)

---

## Model Module

**Purpose:** Data structures (structs). No business methods.

**Location:** `internal/model/`

**Key Files:**
- `post.go` — Post, PostCreate, PostUpdate structs
- `project.go` — Project, ProjectCreate, ProjectUpdate, state/kind constants
- `project_update_entry.go` — ProjectUpdateEntry structs
- `project_showcase_item.go` — Showcase item structs
- `content_relation.go` — ContentRelation, ContentRelationView structs
- `category.go` — Category struct
- `server.go` — Server struct
- `server_metric.go` — ServerMetric struct (CPU, RAM, disk, temp, uptime)
- `server_command.go` — ServerCommand struct + status constants
- `tag.go` — Tag struct
- `media.go` — MediaAsset struct
- `user.go` — User, AuditLog structs + role system (owner/admin/operator/readonly)
- `editorial_inbox.go` — EditorialInboxItem struct, mode/status constants
- `page_content.go` — PageContent struct
- `webhook.go` — WebhookConfig, WebhookEvent structs
- `seo.go` — SEOData struct
- `validators_test.go` — Input validation tests

---

## View Module

**Purpose:** Server-rendered HTML templates (Templ components)

**Location:** `internal/view/`

### Pages (`internal/view/pages/`)

| Template | Data Type | Purpose |
|----------|-----------|---------|
| `index.templ` | IndexData | Landing page |
| `blog_list.templ` | BlogListData | Blog post list |
| `blog_post.templ` | BlogPostData | Single post |
| `projects.templ` | ProjectsData | Project portfolio |
| `project_detail.templ` | ProjectDetailData | Single project |
| `project_changelog.templ` | ProjectChangelogData | Full changelog |
| `about.templ` | AboutData | About page |
| `contact.templ` | ContactData | Contact page |
| `lab.templ` | LabData | Lab landing |
| `docs.templ` | DocsData | Documentation hub |
| `dashboard.templ` | DashboardData | Server monitoring |
| `server_detail.templ` | ServerDetailData | Server detail |
| `server_detail_charts.templ` | — | Charts partial |
| `server_detail_extras.templ` | — | Extras partial |
| `search.templ` | SearchData | Site search |
| `search_partial.templ` | — | HTMX search results |
| `topics.templ` | TopicsData | Category browser |
| `page.templ` | — | Generic content page |
| `not_found.templ` | NotFoundData | 404 page |
| `blog_list_partial.templ` | — | HTMX blog list partial |

### Admin (`internal/view/admin/`)

| Template | Purpose |
|----------|---------|
| `layout.templ` | Admin layout wrapper |
| `editorial.templ` | Editorial inbox UI |
| `server_detail.templ` | Server detail (admin) |
| `pages.templ` | Page content editor |
| `webhooks.templ` | Webhook config UI |

### Components (`internal/view/components/`)

| Template | Purpose |
|----------|---------|
| `nav.templ` | Site navigation |
| `footer.templ` | Site footer |
| `post_card.templ` | Blog post card |
| `project_card.templ` | Project card |
| `server_card.templ` | Server status card |
| `pagination.templ` | Pagination UI |

### Layouts (`internal/view/layouts/`)

- `base.templ` — Base HTML layout with head, meta tags, GA, HTMX
- `helpers.go` — RawHTML component helper

---

## Related Areas

- [Architecture Codemap](ARCHITECTURE.md) — system overview
- [Files Codemap](FILES.md) — file-by-file reference
