# File Map

**Last Updated:** 2026-06-22

## Root

```
F:\GITHUB\raevtar\
├── cmd/
│   └── server/
│       └── main.go              # Entry point: config load, DB init, HTTP server
├── internal/
│   ├── config/
│   │   ├── config.go            # Config struct, Load(), Validate()
│   │   ├── config_test.go       # Config tests
│   │   └── cidr.go              # ParseCIDRs(), IsTrustedIP()
│   ├── handler/
│   │   ├── routes.go            # All Chi route definitions, Handler struct, New()
│   │   ├── handlers.go          # Public page handlers (index, about, blog, projects, dashboard, etc.)
│   │   ├── admin.go             # Admin panel CRUD handlers (1255 lines)
│   │   ├── api.go               # API v1 JSON handlers (public read + admin write)
│   │   ├── auth.go              # In-memory session store, login/logout, CSRF
│   │   ├── auth_test.go         # Auth handler tests
│   │   ├── security.go          # WithSecurityHeaders, RateLimit middleware, trusted proxies
│   │   ├── hardening.go         # Request body limits, login throttling, error helpers
│   │   ├── render.go            # renderHTML() Templ helper
│   │   ├── editorial.go         # Editorial inbox admin handlers
│   │   ├── hoststats.go         # collectHostStats() from /proc (CPU, RAM, disk, temp)
│   │   ├── discovery.go         # sitemap.xml + llms.txt generation
│   │   ├── rss.go               # RSS 2.0 feed for /blog/feed.xml
│   │   ├── og_image.go          # SVG OG image generation for blog posts + projects
│   │   ├── middleware_test.go   # Rate limit + security middleware tests
│   │   ├── diskstats_windows.go # Windows platform disk stats
│   │   ├── diskstats_unix.go    # Unix platform disk stats
│   │   ├── admin_test.go        # Admin handler tests
│   │   ├── admin_subfeature_test.go
│   │   ├── api_test.go          # API handler tests
│   │   ├── api_remaining_test.go
│   │   ├── api_project_test.go  # Project API handler tests
│   │   ├── editorial_test.go    # Editorial handler tests
│   │   ├── handler_remaining_test.go
│   │   ├── hoststats_public_test.go
│   │   └── public_monitoring_test.go
│   ├── service/
│   │   ├── service.go           # Service struct (bundles all sub-services), New()
│   │   ├── seed.go              # SeedData: categories, admin user, default pages
│   │   ├── blog.go              # BlogService: post CRUD, markdown, views (444 lines)
│   │   ├── blog_extra_test.go   # Blog service extra tests
│   │   ├── project.go           # ProjectService: project CRUD, updates, relations, showcase (625 lines)
│   │   ├── project_test.go      # Project service tests
│   │   ├── monitor.go           # MonitorService: server monitoring, agent tokens (238 lines)
│   │   ├── monitor_extra_test.go
│   │   ├── search.go            # SearchService: unified search across content types
│   │   ├── search_test.go       # Search service tests
│   │   ├── site_meta.go         # SiteMetaService: SEO, sitemap, llms.txt (230 lines)
│   │   ├── site_meta_test.go    # Site meta tests
│   │   ├── page_content.go      # PageContentService: about/contact pages
│   │   ├── page_content_test.go
│   │   ├── editorial_inbox.go   # EditorialInboxService: content queue with claiming (506 lines)
│   │   ├── editorial_inbox_extra_test.go
│   │   ├── admin.go             # AdminService: users, auth, audit (262 lines)
│   │   ├── admin_test.go        # Admin service tests
│   │   ├── admin_logging_test.go
│   │   ├── media.go             # MediaService: file upload management (144 lines)
│   │   ├── media_test.go
│   │   ├── command_queue.go     # CommandQueueService: server command queuing
│   │   ├── command_queue_test.go
│   │   ├── webhook.go           # WebhookService: outgoing webhooks (218 lines)
│   │   ├── webhook_test.go
│   │   └── service_test.go      # Service integration tests
│   ├── repo/
│   │   ├── db.go                # InitSQLite(), AutoMigrate(), Repositories aggregate, New()
│   │   ├── db_test.go           # DB tests
│   │   ├── post_repo.go         # PostRepo: posts CRUD + queries
│   │   ├── post_repo_test.go
│   │   ├── project_repo.go      # ProjectRepo: projects CRUD + queries
│   │   ├── project_repo_test.go
│   │   ├── project_update_repo.go    # ProjectUpdateRepo
│   │   ├── category_repo.go     # CategoryRepo
│   │   ├── category_repo_test.go
│   │   ├── server_repo.go       # ServerRepo
│   │   ├── server_repo_test.go
│   │   ├── metric_repo.go       # MetricRepo: server metrics CRUD
│   │   ├── metric_repo_test.go
│   │   ├── tag_repo.go          # TagRepo + post_tags/project_tags joins
│   │   ├── tag_repo_test.go
│   │   ├── media_repo.go        # MediaRepo
│   │   ├── media_repo_test.go
│   │   ├── user_repo.go         # UserRepo + HashPassword()/CheckPassword()
│   │   ├── user_repo_test.go
│   │   ├── audit_repo.go        # AuditRepo
│   │   ├── audit_repo_test.go
│   │   ├── view_repo.go         # ViewRepo: post view tracking
│   │   ├── view_repo_test.go
│   │   ├── command_repo.go      # CommandRepo
│   │   ├── command_repo_test.go
│   │   ├── webhook_repo.go      # WebhookRepo
│   │   ├── page_content_repo.go # PageContentRepo
│   │   ├── editorial_inbox_repo.go  # EditorialInboxRepo
│   │   ├── editorial_inbox_repo_test.go
│   │   ├── content_relation_repo.go # ContentRelationRepo
│   │   ├── project_showcase_repo.go # ProjectShowcaseRepo
│   │   └── all_repo_test.go     # Combined repo tests
│   ├── model/
│   │   ├── post.go              # Post, PostCreate, PostUpdate
│   │   ├── project.go           # Project, ProjectCreate, states, update kinds, relation/showcase constants
│   │   ├── project_update_entry.go  # ProjectUpdateEntry structs
│   │   ├── project_showcase_item.go # ProjectShowcaseItem structs
│   │   ├── content_relation.go  # ContentRelation, ContentRelationView
│   │   ├── category.go          # Category struct
│   │   ├── server.go            # Server struct
│   │   ├── server_metric.go     # ServerMetric struct (CPU, RAM, disk, temp, uptime)
│   │   ├── server_command.go    # ServerCommand struct, CommandStatus constants
│   │   ├── tag.go               # Tag struct
│   │   ├── media.go             # MediaAsset struct
│   │   ├── user.go              # User, AuditLog structs + role system
│   │   ├── editorial_inbox.go   # EditorialInboxItem, fairness summary
│   │   ├── page_content.go      # PageContent struct + page key constants
│   │   ├── webhook.go           # WebhookConfig, WebhookEvent structs
│   │   ├── seo.go               # SEOData struct, MustJSONLD helper
│   │   └── validators_test.go   # Input validation tests
│   └── view/
│       ├── pages/
│       │   ├── index.templ              # Landing page
│       │   ├── about.templ              # About page
│       │   ├── blog_list.templ          # Blog listing
│       │   ├── blog_list_partial.templ  # HTMX partial for blog list
│       │   ├── blog_post.templ          # Single blog post
│       │   ├── projects.templ           # Project portfolio
│       │   ├── project_detail.templ     # Single project detail
│       │   ├── project_changelog.templ  # Full changelog page
│       │   ├── contact.templ            # Contact page
│       │   ├── lab.templ                # Lab landing
│       │   ├── docs.templ               # Documentation hub
│       │   ├── dashboard.templ          # Server monitoring dashboard
│       │   ├── server_detail.templ      # Server detail view
│       │   ├── server_detail_charts.templ   # Charts partial
│       │   ├── server_detail_extras.templ   # Extras partial
│       │   ├── search.templ             # Site search
│       │   ├── search_partial.templ     # HTMX search partial
│       │   ├── topics.templ             # Category browser
│       │   ├── page.templ               # Generic content page
│       │   ├── not_found.templ          # 404 page
│       │   ├── data.go                  # Page data structs (612 lines)
│       │   └── data_test.go             # Page data tests
│       ├── admin/
│       │   ├── layout.templ             # Admin layout
│       │   ├── editorial.templ          # Editorial inbox UI
│       │   ├── server_detail.templ      # Server admin detail
│       │   ├── pages.templ              # Page content editor
│       │   ├── webhooks.templ           # Webhook config UI
│       │   ├── data.go                  # Admin data structs (722 lines)
│       │   └── data_test.go             # Admin data tests
│       ├── components/
│       │   ├── nav.templ                # Site navigation
│       │   ├── footer.templ             # Site footer
│       │   ├── post_card.templ          # Blog post card
│       │   ├── project_card.templ       # Project card
│       │   ├── server_card.templ        # Server status card
│       │   ├── server_card_test.go      # Server card tests
│       │   ├── pagination.templ         # Pagination UI
│       │   └── highlight.go             # Code highlighting utility
│       └── layouts/
│           ├── base.templ               # Base HTML layout (head, meta, scripts)
│           └── helpers.go               # RawHTML templ component helper
├── static/
│   ├── css/
│   │   ├── tailwind.src.css             # Tailwind source (input)
│   │   └── style.css                    # Compiled Tailwind CSS (output)
│   ├── js/
│   │   ├── htmx.min.js                  # HTMX library (self-hosted)
│   │   └── raevtar-ui.js                # Custom UI JS (defer)
│   ├── agent/
│   │   └── raevtar-agent.sh             # Monitoring agent script
│   ├── favicon.svg                      # Site favicon
│   ├── robots.txt                       # Robots exclusion
│   └── openapi.json                     # API specification
├── cron/
│   ├── agent-ping.sh                    # Cron-based agent ping
│   ├── backup.sh                        # DB backup script
│   └── healthcheck.sh                   # Health check script
├── docs/
│   ├── ARCHITECTURE.md                  # Architecture documentation
│   ├── CHI_ROUTING_BUG.md               # Chi routing issue notes
│   ├── DEPLOYMENT.md                    # Deployment guide
│   ├── DESIGN.md                        # Design decisions
│   ├── PRD.md                           # Product requirements
│   ├── ROADMAP.md                       # Development roadmap
│   ├── hermes-integration.md            # Hermes AI integration notes
│   ├── hermes-prompt-draft.md           # Hermes prompt drafts
│   ├── images/                          # Documentation images
│   └── CODEMAPS/                        # This directory
├── migrations/
│   └── 001_init.sql                     # Initial SQL schema (reference)
├── scripts/                             # Scripts directory (currently empty)
├── .env.example                         # Environment variable template
├── .goreleaser.yaml                     # GoReleaser config
├── Makefile                             # Build, test, dev commands
├── go.mod                               # Go module definition
├── go.sum                               # Go dependency checksums
├── package.json                         # Node deps (Tailwind CLI)
├── tailwind.config.js                   # Tailwind CSS configuration
├── raevtar.service                      # Systemd service file
├── AGENTS.md                            # Instructions for AI agents
└── README.md                            # Project README
```

## Key File Sizes

| File | Lines | Purpose |
|------|-------|---------|
| `handler/admin.go` | 1255 | Admin panel handlers (largest file) |
| `handler/handlers.go` | 692 | Public page handlers |
| `handler/api.go` | 657 | API v1 JSON handlers |
| `service/project.go` | 625 | Project business logic |
| `service/editorial_inbox.go` | 506 | Editorial inbox logic |
| `service/blog.go` | 444 | Blog business logic |
| `repo/db.go` | 411 | DB init, migration, repo aggregate |
| `view/pages/data.go` | 612 | Page data structs |
| `view/admin/data.go` | 722 | Admin data structs |

## Related Areas

- [Architecture Codemap](ARCHITECTURE.md) — system overview
- [Modules Codemap](MODULES.md) — detailed module descriptions
