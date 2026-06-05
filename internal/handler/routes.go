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
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(RateLimit)

	// static files
	fs := http.FileServer(http.Dir("./static"))
	r.Handle("/static/*", http.StripPrefix("/static/", fs))
	r.Get("/robots.txt", h.serveStatic("robots.txt"))
	r.Get("/favicon.svg", h.serveStatic("favicon.svg"))
	r.Get("/uploads/{filename}", h.serveUpload)

	// pages
	r.Get("/", h.landingIndex)
	r.Get("/about", h.aboutPage)
	r.Get("/blog", h.blogList)
	r.Get("/blog/{slug}", h.blogDetail)
	r.Get("/contact", h.contactPage)
	r.Get("/lab", h.labPage)
	r.Get("/docs", h.docsPage)
	r.Get("/lab/docs", h.docsPage)
	r.Get("/projects", h.projectsPage)
	r.Get("/projects/{slug}", h.projectDetail)
	r.Get("/topics", h.topicsPage)
	r.Get("/dashboard", h.dashboardIndex)
	r.Get("/dashboard/{serverID}/live", h.dashboardDetailLive)
	r.Get("/dashboard/{serverID}", h.dashboardDetail)

	// RSS feed
	r.Get("/blog/feed.xml", h.rssFeed)

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
				r.Post("/projects/preview", h.adminPreviewProject)
				r.Post("/projects", h.adminCreateProject)
				r.Get("/projects/edit/{projectID}", h.adminEditProject)
				r.Post("/projects/update/{projectID}", h.adminUpdateProject)
				r.Post("/projects/delete/{projectID}", h.adminDeleteProject)
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
				r.Get("/audit-log", h.adminAuditLog)
				r.Get("/manage-users", h.adminUsers)
				r.Post("/manage-users", h.adminCreateUser)
				r.Post("/manage-users/delete/{userID}", h.adminDeleteUser)
			})
		})
	})

	// API v1
	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/posts", h.apiListPosts)
		r.Get("/projects", h.apiListProjects)
		r.With(h.adminAuth).Post("/posts", h.apiCreatePost)
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
	})

	// Store mux for debugging
	h.Mux = r

	return r
}
