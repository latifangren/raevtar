package handler

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strings"
	"testing"

	"raevtar/internal/config"
	"raevtar/internal/model"
	"raevtar/internal/repo"
	"raevtar/internal/service"
)

type publicTestApp struct {
	handler http.Handler
}

func newPublicTestApp(t *testing.T) *publicTestApp {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "raevtar_test.db")
	cfg := &config.Config{
		DatabasePath: dbPath,
		Domain:       "raevtar.test",
		AdminUser:    "admin",
		AdminPass:    "demo-pass-123",
	}

	db := repo.InitSQLite(cfg.DatabasePath)
	t.Cleanup(func() {
		_ = db.Close()
	})
	repo.AutoMigrate(db)

	repos := repo.New(db)
	svc := service.New(repos, cfg)
	if err := svc.SeedData(); err != nil {
		t.Fatalf("seed data: %v", err)
	}

	post, err := svc.Blog.CreatePost(model.PostCreate{
		CategorySlug: "devops",
		Title:        "Hello Raevtar",
		ContentMD:    "# Hello Raevtar\n\nBaseline route test.",
		Excerpt:      "First test post.",
		Published:    true,
	})
	if err != nil {
		t.Fatalf("create post: %v", err)
	}

	if _, err := svc.Monitor.CreateServer("whyred", "127.0.0.1", 9100, "local"); err != nil {
		t.Fatalf("create server: %v", err)
	}

	if post.Slug == "" {
		t.Fatalf("expected generated slug")
	}

	return &publicTestApp{handler: New(svc, cfg)}
}

func TestAdminLoginRedirectsToDashboard(t *testing.T) {
	app := newPublicTestApp(t)

	form := url.Values{
		"username": {"admin"},
		"password": {"demo-pass-123"},
	}
	req := httptest.NewRequest(http.MethodPost, "/admin/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()

	app.handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusSeeOther, rr.Body.String())
	}
	if got := rr.Header().Get("Location"); got != "/admin" {
		t.Fatalf("Location = %q, want /admin", got)
	}
	if strings.Contains(rr.Body.String(), `"redirect":"/admin"`) {
		t.Fatalf("login returned JSON redirect body: %s", rr.Body.String())
	}
}

func TestPublicRoutes(t *testing.T) {
	app := newPublicTestApp(t)

	tests := []struct {
		name           string
		path           string
		wantStatus     int
		wantContentTyp string
		wantContains   []string
	}{
		{
			name:           "landing",
			path:           "/",
			wantStatus:     http.StatusOK,
			wantContentTyp: "text/html",
			wantContains: []string{
				"raevtar",
				"Read the Blog",
				"Open Dashboard",
				"Hello Raevtar",
				`href="/blog"`,
				`href="/dashboard"`,
			},
		},
		{
			name:           "blog index",
			path:           "/blog",
			wantStatus:     http.StatusOK,
			wantContentTyp: "text/html",
			wantContains: []string{
				"Blog",
				"Hello Raevtar",
				`href="/blog?category=devops"`,
				`href="/blog/hello-raevtar"`,
			},
		},
		{
			name:           "dashboard index",
			path:           "/dashboard",
			wantStatus:     http.StatusOK,
			wantContentTyp: "text/html",
			wantContains: []string{
				"Server Dashboard",
				"whyred",
				"127.0.0.1:9100",
				`href="/dashboard/1"`,
			},
		},
		{
			name:           "missing route",
			path:           "/missing-route",
			wantStatus:     http.StatusOK,
			wantContentTyp: "text/html",
			wantContains: []string{
				"Halaman gak ketemu",
				"Balik ke Beranda",
			},
		},
		{
			name:           "rss feed",
			path:           "/blog/feed.xml",
			wantStatus:     http.StatusOK,
			wantContentTyp: "application/rss+xml",
			wantContains: []string{
				"<title>Hello Raevtar</title>",
				"https://raevtar.test/blog/hello-raevtar",
				"<category>devops</category>",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			rr := httptest.NewRecorder()

			app.handler.ServeHTTP(rr, req)

			if rr.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", rr.Code, tt.wantStatus)
			}
			if got := rr.Header().Get("Content-Type"); !strings.HasPrefix(got, tt.wantContentTyp) {
				t.Fatalf("content type = %q, want prefix %q", got, tt.wantContentTyp)
			}

			body := rr.Body.String()
			for _, want := range tt.wantContains {
				if !strings.Contains(body, want) {
					t.Fatalf("body missing %q\nbody: %s", want, body)
				}
			}
		})
	}
}

func TestBlogDetailHidesDraftPosts(t *testing.T) {
	app := newPublicTestApp(t)

	req := httptest.NewRequest(http.MethodGet, "/blog/tools-draft", nil)
	rr := httptest.NewRecorder()

	app.handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusNotFound)
	}
	if strings.Contains(rr.Body.String(), "Tools Draft") {
		t.Fatalf("draft content leaked in response: %s", rr.Body.String())
	}
}

func TestRSSFeedAllowsNoPosts(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "raevtar_empty_feed.db")
	cfg := &config.Config{
		DatabasePath: dbPath,
		Domain:       "raevtar.test",
		AdminUser:    "admin",
	}

	db := repo.InitSQLite(cfg.DatabasePath)
	t.Cleanup(func() {
		_ = db.Close()
	})
	repo.AutoMigrate(db)

	svc := service.New(repo.New(db), cfg)
	if err := svc.SeedData(); err != nil {
		t.Fatalf("seed data: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/blog/feed.xml", nil)
	rr := httptest.NewRecorder()
	New(svc, cfg).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}
	if got := rr.Header().Get("Content-Type"); !strings.HasPrefix(got, "application/rss+xml") {
		t.Fatalf("content type = %q, want application/rss+xml", got)
	}
	body := rr.Body.String()
	for _, want := range []string{"<channel>", "<title>Raevtar</title>", "</rss>"} {
		if !strings.Contains(body, want) {
			t.Fatalf("body missing %q\nbody: %s", want, body)
		}
	}
}

func TestAdminMutationsRequireOwnerOrAdmin(t *testing.T) {
	app := newPublicTestApp(t)
	token := sessions.create(42, "reader", model.RoleReadonly)

	form := url.Values{
		"title":         {"Blocked Post"},
		"category_slug": {"devops"},
		"content":       {"# Blocked"},
	}
	req := httptest.NewRequest(http.MethodPost, "/admin/posts", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: token})
	rr := httptest.NewRecorder()

	app.handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusForbidden)
	}
}

func TestAdminManagementPagesRequireOwnerOrAdmin(t *testing.T) {
	app := newPublicTestApp(t)
	token := sessions.create(44, "reader", model.RoleReadonly)

	for _, path := range []string{
		"/admin/posts",
		"/admin/servers",
		"/admin/audit-log",
		"/admin/manage-users",
	} {
		t.Run(path, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, path, nil)
			req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: token})
			rr := httptest.NewRecorder()

			app.handler.ServeHTTP(rr, req)

			if rr.Code != http.StatusForbidden {
				t.Fatalf("status = %d, want %d", rr.Code, http.StatusForbidden)
			}
		})
	}
}

func TestAdminDeleteRoutesRejectGET(t *testing.T) {
	app := newPublicTestApp(t)
	token := sessions.create(43, "owner", model.RoleOwner)

	for _, path := range []string{
		"/admin/posts/delete/1",
		"/admin/servers/delete/1",
		"/admin/manage-users/delete/1",
	} {
		t.Run(path, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, path, nil)
			req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: token})
			rr := httptest.NewRecorder()

			app.handler.ServeHTTP(rr, req)

			if rr.Code != http.StatusMethodNotAllowed {
				t.Fatalf("status = %d, want %d", rr.Code, http.StatusMethodNotAllowed)
			}
		})
	}
}
