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
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(RateLimit)

	// static files
	fs := http.FileServer(http.Dir("./static"))
	r.Handle("/static/*", http.StripPrefix("/static/", fs))
	r.Get("/robots.txt", h.serveStatic("robots.txt"))
	r.Get("/favicon.svg", h.serveStatic("favicon.svg"))
	r.Get("/docs", h.serveStatic("docs.html"))

	// pages
	r.Get("/", h.landingIndex)
	r.Get("/blog", h.blogList)
	r.Get("/blog/{slug}", h.blogDetail)
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
				r.Get("/posts", h.adminPosts)
				r.Post("/posts", h.adminCreatePost)
				r.Post("/posts/delete/{postID}", h.adminDeletePost)
				r.Get("/servers", h.adminServers)
				r.Post("/servers", h.adminCreateServer)
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
		r.With(h.adminAuth).Post("/posts", h.apiCreatePost)

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
