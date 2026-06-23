package handler

import (
	"net/http"
	"raevtar/internal/config"
	"raevtar/internal/service"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Handler struct {
	svc *service.Service
	cfg *config.Config
	Mux *chi.Mux // expose chi mux for debugging
}

func New(svc *service.Service, cfg *config.Config) http.Handler {
	h := &Handler{svc: svc, cfg: cfg}
	configureTrustedProxies(cfg)
	initRateLimiter(cfg)
	initHardening(cfg)
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(RateLimit)

	// static files (resolve relative to binary, not CWD)
	fs := http.FileServer(http.Dir(h.cfg.StaticDir))
	r.Handle("/static/*", http.StripPrefix("/static/", fs))
	r.Get("/robots.txt", h.robotsTxt)
	r.Get("/favicon.svg", h.serveStatic("favicon.svg"))
	r.Get("/llms.txt", h.llmsTxt)
	r.Get("/sitemap.xml", h.sitemapXML)
	r.Get("/uploads/{filename}", h.serveUpload)
	r.Get("/og-image/blog/{slug}", h.serveBlogOGImage)
	r.Get("/og-image/project/{slug}", h.serveProjectOGImage)

	// pages
	r.Get("/", h.landingIndex)
	r.Get("/about", h.aboutPage)
	r.Get("/blog", h.blogList)
	r.Get("/blog/{slug}", h.blogDetail)
	r.Get("/contact", h.contactPage)
	r.Get("/lab", h.labPage)
	r.Get("/docs", h.docsPage)
	r.Get("/docs/api", h.apiDocsPage)
	r.Get("/lab/docs", h.docsPage)
	r.Get("/projects", h.projectsPage)
	r.Get("/search", h.searchPage)
	r.Get("/projects/{slug}/changelog", h.projectChangelogPage)
	r.Get("/projects/{slug}", h.projectDetail)
	r.Get("/lab/node-status/{name}", h.nodeStatusShortcode)
	r.Get("/topics", h.topicsPage)
	r.Get("/dashboard", h.dashboardIndex)
	r.Get("/dashboard/{serverID}/live", h.dashboardDetailLive)
	r.Get("/dashboard/{serverID}", h.dashboardDetail)

	// RSS feed
	r.Get("/blog/feed.xml", h.rssFeed)

	// Webmention
	r.Post("/webmention", h.handleWebmention)

	// 404
	r.NotFound(h.page404)

	// Admin routes (session-based auth)
	r.Route("/admin", func(r chi.Router) {
		// Public (no auth needed)
		r.Get("/login", h.adminLoginPage)
		r.Post("/login", h.adminLogin)

		// Protected — require valid session
		r.Group(func(r chi.Router) {
			r.Use(h.adminRequired)
			r.Use(h.adminCSRF)
			r.Get("/", h.adminIndex)
			r.Post("/logout", h.adminLogout)

			r.Group(func(r chi.Router) {
				r.Use(h.ownerOrAdminRequired)
				r.Get("/editorial-inbox", h.adminEditorialInbox)
				r.Get("/posts", h.adminPosts)
				r.Get("/topics", h.adminTopics)
				r.Get("/projects", h.adminProjects)
				r.Get("/pages", h.adminPages)
				r.Post("/editorial-inbox", h.adminCreateEditorialInbox)
				r.Post("/editorial-inbox/update/{itemID}", h.adminUpdateEditorialInbox)
				r.Post("/editorial-inbox/delete/{itemID}", h.adminDeleteEditorialInbox)
				r.Post("/posts/preview", h.adminPreviewPost)
				r.Post("/posts", h.adminCreatePost)
				r.Get("/posts/edit/{postID}", h.adminEditPost)
				r.Post("/posts/update/{postID}", h.adminUpdatePost)
				r.Post("/posts/delete/{postID}", h.adminDeletePost)
				r.Post("/topics", h.adminCreateTopic)
				r.Get("/topics/edit/{categoryID}", h.adminEditTopic)
				r.Post("/topics/update/{categoryID}", h.adminUpdateTopic)
				r.Post("/topics/delete/{categoryID}", h.adminDeleteTopic)
				r.Post("/projects/preview", h.adminPreviewProject)
				r.Post("/projects", h.adminCreateProject)
				r.Get("/projects/edit/{projectID}", h.adminEditProject)
				r.Post("/projects/update/{projectID}", h.adminUpdateProject)
				r.Post("/projects/delete/{projectID}", h.adminDeleteProject)
				r.Post("/projects/{projectID}/updates", h.adminCreateProjectUpdate)
				r.Post("/projects/updates/update/{updateID}", h.adminUpdateProjectUpdate)
				r.Post("/projects/updates/delete/{updateID}", h.adminDeleteProjectUpdate)
				r.Post("/projects/{projectID}/relations", h.adminCreateProjectRelation)
				r.Post("/projects/relations/delete/{relationID}", h.adminDeleteProjectRelation)
				r.Post("/projects/{projectID}/showcase", h.adminCreateProjectShowcase)
				r.Post("/projects/showcase/update/{itemID}", h.adminUpdateProjectShowcase)
				r.Post("/projects/showcase/delete/{itemID}", h.adminDeleteProjectShowcase)
				r.Get("/pages/edit/{pageKey}", h.adminEditPage)
				r.Post("/pages/update/{pageKey}", h.adminUpdatePage)
				r.Get("/media", h.adminMedia)
				r.Post("/media", h.adminUploadMedia)
				r.Get("/servers", h.adminServers)
				r.Get("/servers/{serverID}", h.adminServerDetail)
				r.Post("/servers", h.adminCreateServer)
				r.Post("/servers/update/{serverID}", h.adminUpdateServer)
				r.Post("/servers/rotate-token/{serverID}", h.adminRotateServerToken)
				r.Post("/servers/delete/{serverID}", h.adminDeleteServer)
				r.Post("/servers/command/{serverID}", h.adminServerCommand)
				r.Get("/webhooks", h.adminWebhooks)
				r.Post("/webhooks", h.adminCreateWebhook)
				r.Post("/webhooks/update/{webhookID}", h.adminUpdateWebhook)
				r.Post("/webhooks/delete/{webhookID}", h.adminDeleteWebhook)
				r.Get("/audit-log", h.adminAuditLog)
				r.Get("/manage-users", h.adminUsers)
				r.Post("/manage-users", h.adminCreateUser)
				r.Post("/manage-users/delete/{userID}", h.adminDeleteUser)
				r.Get("/db", h.adminDBPage)
				r.Get("/db/export", h.adminDBExport)
				r.Post("/db/import", h.adminDBImport)
			})
		})
	})

	// API v1
	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/posts", h.apiListPosts)
		r.Post("/posts/{id}/read-time", h.apiRecordPostReadTime)
		r.Get("/search", h.apiSearch)
		r.Get("/projects", h.apiListProjects)
		r.Get("/projects/{slug}/updates", h.apiListProjectUpdates)
		r.Get("/projects/{slug}/changelog", h.apiListProjectChangelog)
		r.Get("/projects/{slug}/relations", h.apiListProjectRelations)
		r.Get("/projects/{slug}/showcase", h.apiListProjectShowcase)
		r.With(h.adminAuth).Post("/posts", h.apiCreatePost)
		r.With(h.adminAuth).Post("/projects", h.apiCreateProject)
		r.With(h.adminAuth).Put("/projects/{projectID}", h.apiUpdateProject)
		r.With(h.adminAuth).Delete("/projects/{projectID}", h.apiDeleteProject)
		r.With(h.adminAuth).Post("/projects/{projectID}/updates", h.apiCreateProjectUpdate)
		r.With(h.adminAuth).Put("/projects/updates/{updateID}", h.apiUpdateProjectUpdate)
		r.With(h.adminAuth).Delete("/projects/updates/{updateID}", h.apiDeleteProjectUpdate)
		r.With(h.adminAuth).Post("/projects/{projectID}/relations", h.apiCreateProjectRelation)
		r.With(h.adminAuth).Delete("/projects/relations/{relationID}", h.apiDeleteProjectRelation)
		r.With(h.adminAuth).Post("/projects/{projectID}/showcase", h.apiCreateProjectShowcase)
		r.With(h.adminAuth).Put("/projects/showcase/{itemID}", h.apiUpdateProjectShowcase)
		r.With(h.adminAuth).Delete("/projects/showcase/{itemID}", h.apiDeleteProjectShowcase)
		r.With(h.adminAuth).Get("/editorial-inbox/contract", h.apiEditorialInboxContract)
		r.With(h.adminAuth).Get("/editorial-inbox/summary", h.apiEditorialInboxSummary)
		r.With(h.adminAuth).Get("/editorial-inbox", h.apiListEditorialInbox)
		r.With(h.adminAuth).Post("/editorial-inbox", h.apiCreateEditorialInbox)
		r.With(h.adminAuth).Post("/editorial-inbox/claim", h.apiClaimEditorialInbox)
		r.With(h.adminAuth).Get("/editorial-inbox/{itemID}", h.apiGetEditorialInbox)
		r.With(h.adminAuth).Post("/editorial-inbox/{itemID}", h.apiUpdateEditorialInbox)
		r.With(h.adminAuth).Post("/editorial-inbox/{itemID}/complete", h.apiCompleteEditorialInboxClaim)
		r.With(h.adminAuth).Post("/editorial-inbox/{itemID}/fail", h.apiFailEditorialInboxClaim)

		r.Get("/categories", h.apiListCategories)

		r.With(h.adminAuth).Get("/hoststats", h.apiHostStats)

		r.With(h.adminAuth).Get("/servers", h.apiListServers)
		r.With(h.adminAuth).Get("/servers/{serverID}", h.apiGetServer)
		r.With(h.adminAuth).Post("/servers", h.apiCreateServer)
		r.Post("/servers/{serverID}/ping", h.apiRecordMetrics)
		r.Get("/servers/{serverID}/commands", h.apiGetPendingCommands)
		r.Post("/servers/{serverID}/commands/result", h.apiReportCommandResult)
		r.Get("/bootstrap/{serverID}/{token}", h.apiBootstrap)
	})

	// Store mux for debugging
	h.Mux = r

	return r
}
