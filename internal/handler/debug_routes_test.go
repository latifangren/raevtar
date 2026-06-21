package handler

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"raevtar/internal/config"
	"raevtar/internal/model"
	"raevtar/internal/repo"
	"raevtar/internal/service"

	"github.com/jmoiron/sqlx"
)

var testRequestCounter int

type publicTestApp struct {
	handler  http.Handler
	svc      *service.Service
	db       *sqlx.DB
	serverID int64
}

func newPublicTestApp(t *testing.T) *publicTestApp {
	t.Helper()

	limiter.mu.Lock()
	limiter.requests = make(map[string]*bucketEntry)
	limiter.mu.Unlock()

	dbPath := filepath.Join(t.TempDir(), "raevtar_test.db")
	cfg := &config.Config{
		DatabasePath: dbPath,
		Domain:       "raevtar.test",
		AdminKey:     "admin-key",
		AdminUser:    "admin",
		AdminPass:    "demo-pass-123",
		MediaDir:     filepath.Join(t.TempDir(), "uploads"),
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

	project, err := svc.Projects.CreateProject(model.ProjectCreate{
		Title:     "Whyred Watchtower",
		ContentMD: "# Whyred Watchtower\n\nProject page baseline route test.",
		Excerpt:   "Project baseline route test.",
		Published: true,
		Featured:  true,
		SortOrder: 1,
		Tags:      []string{"oss"},
	})
	if err != nil {
		t.Fatalf("create project: %v", err)
	}

	server, err := svc.Monitor.CreateServer("whyred", "127.0.0.1", 9100, "local")
	if err != nil {
		t.Fatalf("create server: %v", err)
	}

	if post.Slug == "" {
		t.Fatalf("expected generated slug")
	}
	if project.Slug == "" {
		t.Fatalf("expected generated project slug")
	}

	return &publicTestApp{handler: New(svc, cfg), svc: svc, db: db, serverID: server.ID}
}

func assertContains(t *testing.T, body, want string) {
	t.Helper()
	if !strings.Contains(body, want) {
		t.Fatalf("body missing %q\nbody: %s", want, body)
	}
}

func assertNotContains(t *testing.T, body, want string) {
	t.Helper()
	if strings.Contains(body, want) {
		t.Fatalf("body leaked %q\nbody: %s", want, body)
	}
}

func setMetricField(t *testing.T, metric *model.ServerMetric, name string, value any) {
	t.Helper()
	field := reflect.ValueOf(metric).Elem().FieldByName(name)
	if !field.IsValid() {
		t.Fatalf("ServerMetric missing authorized field %s", name)
	}
	input := reflect.ValueOf(value)
	if !input.Type().ConvertibleTo(field.Type()) {
		t.Fatalf("ServerMetric.%s type = %s, cannot set %s", name, field.Type(), input.Type())
	}
	field.Set(input.Convert(field.Type()))
}

func getBody(t *testing.T, app *publicTestApp, path string, cookie *http.Cookie) (int, string) {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	testRequestCounter++
	req.RemoteAddr = fmt.Sprintf("203.0.113.%d:1234", testRequestCounter%250+1)
	if cookie != nil {
		req.AddCookie(cookie)
	}
	rr := httptest.NewRecorder()

	app.handler.ServeHTTP(rr, req)

	return rr.Code, rr.Body.String()
}

func testPNG(t *testing.T) []byte {
	t.Helper()
	data, err := base64.StdEncoding.DecodeString("iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAwMCAO+/p9sAAAAASUVORK5CYII=")
	if err != nil {
		t.Fatalf("decode png: %v", err)
	}
	return data
}

func mediaUploadRequest(t *testing.T, path, token, csrf, filename string, data []byte) *http.Request {
	t.Helper()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	_ = writer.WriteField("_csrf", csrf)
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}
	if _, err := part.Write(data); err != nil {
		t.Fatalf("write form file: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, path, &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: token})
	return req
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

func TestAdminLoginIgnoresForwardedForHeader(t *testing.T) {
	app := newPublicTestApp(t)

	form := url.Values{
		"username": {"admin"},
		"password": {"demo-pass-123"},
	}
	req := httptest.NewRequest(http.MethodPost, "/admin/login", strings.NewReader(form.Encode()))
	req.RemoteAddr = "192.0.2.10:4567"
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("X-Forwarded-For", "203.0.113.99")
	rr := httptest.NewRecorder()

	app.handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusSeeOther, rr.Body.String())
	}
	logs, err := app.svc.Admin.ListAuditLogs(5, 0)
	if err != nil {
		t.Fatalf("list audit logs: %v", err)
	}
	if len(logs) == 0 {
		t.Fatalf("expected login audit log")
	}
	if logs[0].IP != "192.0.2.10" {
		t.Fatalf("audit IP = %q, want RemoteAddr host", logs[0].IP)
	}
}

func TestAdminLoginUsesCFConnectingIPFromTrustedProxy(t *testing.T) {
	app := newPublicTestApp(t)
	app.svc.Cfg.TrustedProxyCIDRs = []string{"192.0.2.10/32"}
	app.handler = New(app.svc, app.svc.Cfg)

	form := url.Values{
		"username": {"admin"},
		"password": {"demo-pass-123"},
	}
	req := httptest.NewRequest(http.MethodPost, "/admin/login", strings.NewReader(form.Encode()))
	req.RemoteAddr = "192.0.2.10:4567"
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("CF-Connecting-IP", "203.0.113.99")
	rr := httptest.NewRecorder()

	app.handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusSeeOther, rr.Body.String())
	}
	logs, err := app.svc.Admin.ListAuditLogs(5, 0)
	if err != nil {
		t.Fatalf("list audit logs: %v", err)
	}
	if len(logs) == 0 {
		t.Fatalf("expected login audit log")
	}
	if logs[0].IP != "203.0.113.99" {
		t.Fatalf("audit IP = %q, want trusted CF header", logs[0].IP)
	}
}

func TestAdminLoginIgnoresForwardedHeaderFromUntrustedProxy(t *testing.T) {
	app := newPublicTestApp(t)
	app.svc.Cfg.TrustedProxyCIDRs = []string{"198.51.100.10/32"}
	app.handler = New(app.svc, app.svc.Cfg)

	form := url.Values{
		"username": {"admin"},
		"password": {"demo-pass-123"},
	}
	req := httptest.NewRequest(http.MethodPost, "/admin/login", strings.NewReader(form.Encode()))
	req.RemoteAddr = "192.0.2.10:4567"
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("CF-Connecting-IP", "203.0.113.99")
	rr := httptest.NewRecorder()

	app.handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusSeeOther, rr.Body.String())
	}
	logs, err := app.svc.Admin.ListAuditLogs(5, 0)
	if err != nil {
		t.Fatalf("list audit logs: %v", err)
	}
	if len(logs) == 0 {
		t.Fatalf("expected login audit log")
	}
	if logs[0].IP != "192.0.2.10" {
		t.Fatalf("audit IP = %q, want RemoteAddr host", logs[0].IP)
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
				"Browse Notes",
				"Live Status",
				"Platform Showcase",
				"about",
				"projects",
				"topics",
				"contact",
				"status",
				"lab",
				`href="/about"`,
				`href="/lab"`,
				`href="/projects"`,
				`href="/topics"`,
				`href="/contact"`,
				"Hello Raevtar",
				`href="/blog"`,
				`href="/dashboard"`,
				`href="/docs"`,
				`rel="canonical" href="https://raevtar.test/"`,
				`application/ld+json`,
				`"@type":"WebSite"`,
			},
		},
		{
			name:           "about page",
			path:           "/about",
			wantStatus:     http.StatusOK,
			wantContentTyp: "text/html",
			wantContains: []string{
				"About Raevtar",
				"Single binary, public-safe by design.",
				"Raevtar is personal platform for writing notes",
				"raevtar.test",
				"topics",
				"signals",
				"Built like small lab notebook, not SaaS brochure.",
				`href="/projects"`,
			},
		},
		{
			name:           "projects page",
			path:           "/projects",
			wantStatus:     http.StatusOK,
			wantContentTyp: "text/html",
			wantContains: []string{
				"Projects",
				"Three public lanes, one small machine.",
				"Build log archive",
				"Sort archive",
				"Featured only",
				"Reset filters",
				"featured lane",
				"Whyred Watchtower",
				`href="/projects/whyred-watchtower"`,
				"Publishing System",
				"Signal Board",
				"Automation Surface",
				"Move sideways, not only downward.",
				`href="/about"`,
				`href="/topics"`,
			},
		},
		{
			name:           "project detail",
			path:           "/projects/whyred-watchtower",
			wantStatus:     http.StatusOK,
			wantContentTyp: "text/html",
			wantContains: []string{
				"Project entry",
				"Whyred Watchtower",
				"Project page baseline route test.",
				"Featured build",
				"Back to projects",
			},
		},
		{
			name:           "topics page",
			path:           "/topics",
			wantStatus:     http.StatusOK,
			wantContentTyp: "text/html",
			wantContains: []string{
				"Topics",
				"Browse the archive by category.",
				"DevOps",
				`href="/blog?category=devops"`,
				"next stops",
				`href="/contact"`,
			},
		},
		{
			name:           "contact page",
			path:           "/contact",
			wantStatus:     http.StatusOK,
			wantContentTyp: "text/html",
			wantContains: []string{
				"Contact",
				"Raevtar keeps contact intentionally lightweight.",
				"Best first contact here is context, not a form.",
				"raevtar.test",
				"Read the docs",
				"Best route depends on what you need.",
				`href="/about"`,
			},
		},
		{
			name:           "lab page",
			path:           "/lab",
			wantStatus:     http.StatusOK,
			wantContentTyp: "text/html",
			wantContains: []string{
				"Public lab bench",
				"Signal Lab",
				"Content Lab",
				"Automation Lab",
			},
		},
		{
			name:           "docs page",
			path:           "/docs",
			wantStatus:     http.StatusOK,
			wantContentTyp: "text/html",
			wantContains: []string{
				"Public docs",
				"Public front, admin back room",
				"GET /api/v1/posts",
				"GET /api/v1/projects",
				"GET /api/v1/categories",
				"related public pages",
				`href="/projects"`,
				`rel="canonical" href="https://raevtar.test/docs"`,
			},
		},
		{
			name:           "lab docs page",
			path:           "/lab/docs",
			wantStatus:     http.StatusOK,
			wantContentTyp: "text/html",
			wantContains: []string{
				"Public docs",
				"Public front, admin back room",
				"related public pages",
				`rel="canonical" href="https://raevtar.test/docs"`,
			},
		},
		{
			name:           "stale swagger shell removed",
			path:           "/static/docs.html",
			wantStatus:     http.StatusNotFound,
			wantContentTyp: "text/plain",
			wantContains: []string{
				"404 page not found",
			},
		},
		{
			name:           "blog index",
			path:           "/blog",
			wantStatus:     http.StatusOK,
			wantContentTyp: "text/html",
			wantContains: []string{
				"Field Notes",
				"Latest dispatches",
				"Blog",
				"Hello Raevtar",
				"Read Article",
				`href="/blog?category=devops"`,
				`href="/blog/hello-raevtar"`,
				`rel="canonical" href="https://raevtar.test/blog"`,
			},
		},
		{
			name:           "blog detail",
			path:           "/blog/hello-raevtar",
			wantStatus:     http.StatusOK,
			wantContentTyp: "text/html",
			wantContains: []string{
				"Signal brief",
				"Filed under",
				"Back to blog",
				"Hello Raevtar",
				"Baseline route test.",
				`rel="canonical" href="https://raevtar.test/blog/hello-raevtar"`,
				`"@type":"BlogPosting"`,
				`"headline":"Hello Raevtar"`,
			},
		},
		{
			name:           "dashboard index",
			path:           "/dashboard",
			wantStatus:     http.StatusOK,
			wantContentTyp: "text/html",
			wantContains: []string{
				"Server Status",
				"Monitor Grid",
				"status",
				"Online",
				"Stale",
				"Offline",
				"last agent signal",
				"grid refresh every 30s",
				"whyred",
				"Hidden on public view",
				"redacted",
				"related public routes",
				`href="/projects"`,
				`href="/dashboard/1"`,
				`id="server-list"`,
				`hx-get="/dashboard"`,
				`hx-trigger="every 30s"`,
				`hx-select="#server-list"`,
				`hx-target="#server-list"`,
				`hx-swap="outerHTML"`,
			},
		},
		{
			name:           "missing route",
			path:           "/missing-route",
			wantStatus:     http.StatusNotFound,
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
		{
			name:           "llms txt",
			path:           "/llms.txt",
			wantStatus:     http.StatusOK,
			wantContentTyp: "text/plain",
			wantContains: []string{
				"# Raevtar",
				"https://raevtar.test/blog",
				"https://raevtar.test/projects",
				"Hello Raevtar",
				"Whyred Watchtower",
			},
		},
		{
			name:           "sitemap xml",
			path:           "/sitemap.xml",
			wantStatus:     http.StatusOK,
			wantContentTyp: "application/xml",
			wantContains: []string{
				`<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">`,
				`<loc>https://raevtar.test/</loc>`,
				`<loc>https://raevtar.test/blog/hello-raevtar</loc>`,
				`<loc>https://raevtar.test/projects/whyred-watchtower</loc>`,
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
				assertContains(t, body, want)
			}
		})
	}
}

func TestCanonicalNormalization(t *testing.T) {
	app := newPublicTestApp(t)

	status, body := getBody(t, app, "/blog?category=devops&page=2", nil)
	if status != http.StatusOK {
		t.Fatalf("blog status = %d, want %d; body: %s", status, http.StatusOK, body)
	}
	assertContains(t, body, `rel="canonical" href="https://raevtar.test/blog"`)
	assertNotContains(t, body, `https://raevtar.test/blog?category=devops&page=2`)

	status, body = getBody(t, app, "/projects?featured=true&sort=oldest", nil)
	if status != http.StatusOK {
		t.Fatalf("projects status = %d, want %d; body: %s", status, http.StatusOK, body)
	}
	assertContains(t, body, `rel="canonical" href="https://raevtar.test/projects"`)
	assertNotContains(t, body, `https://raevtar.test/projects?featured=true&sort=oldest`)

	status, body = getBody(t, app, "/search?q=hello&scope=posts&page=2", nil)
	if status != http.StatusOK {
		t.Fatalf("search status = %d, want %d; body: %s", status, http.StatusOK, body)
	}
	assertContains(t, body, `rel="canonical" href="https://raevtar.test/search"`)
	assertNotContains(t, body, `https://raevtar.test/search?q=hello&scope=posts&page=2`)
}

func TestBlogPageSupportsSearchQuery(t *testing.T) {
	app := newPublicTestApp(t)

	_, err := app.svc.Blog.CreatePost(model.PostCreate{
		CategorySlug: "tools",
		Title:        "Terminal Search",
		ContentMD:    "# Terminal Search\n\nFinding notes fast.",
		Excerpt:      "Search fixture",
		Published:    true,
	})
	if err != nil {
		t.Fatalf("create search post: %v", err)
	}

	status, body := getBody(t, app, "/blog?q=terminal", nil)
	if status != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", status, http.StatusOK, body)
	}
	assertContains(t, body, "<mark")
	assertContains(t, body, `href="/blog/terminal-search"`)
	assertNotContains(t, body, "Hello Raevtar")
	assertContains(t, body, `name="q" value="terminal"`)

	status, body = getBody(t, app, "/blog?category=tools&q=terminal", nil)
	if status != http.StatusOK {
		t.Fatalf("filtered status = %d, want %d; body: %s", status, http.StatusOK, body)
	}
	assertContains(t, body, "<mark")
	assertContains(t, body, `href="/blog/terminal-search"`)
	assertContains(t, body, `/blog?category=tools&amp;q=terminal`)
}

func TestAPIListProjectsReturnsPublishedProjectsOnly(t *testing.T) {
	app := newPublicTestApp(t)

	_, err := app.svc.Projects.CreateProject(model.ProjectCreate{
		Title:     "Draft Project",
		ContentMD: "# Draft",
		Excerpt:   "hidden draft",
		Published: false,
		Featured:  true,
		SortOrder: 0,
	})
	if err != nil {
		t.Fatalf("create draft project: %v", err)
	}
	_, err = app.svc.Projects.CreateProject(model.ProjectCreate{
		Title:     "Later Project",
		ContentMD: "# Later",
		Excerpt:   "later project",
		Published: true,
		Featured:  false,
		SortOrder: 9,
	})
	if err != nil {
		t.Fatalf("create later project: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/projects", nil)
	rr := httptest.NewRecorder()
	app.handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}
	var projects []model.Project
	if err := json.NewDecoder(rr.Body).Decode(&projects); err != nil {
		t.Fatalf("decode projects: %v", err)
	}
	if len(projects) != 2 {
		t.Fatalf("projects len = %d, want 2", len(projects))
	}
	if projects[0].Title != "Whyred Watchtower" || !projects[0].Featured {
		t.Fatalf("first project = %+v, want featured whyred", projects[0])
	}
	if projects[1].Title != "Later Project" {
		t.Fatalf("second project = %+v, want Later Project", projects[1])
	}
	for _, project := range projects {
		if project.Title == "Draft Project" {
			t.Fatalf("draft project leaked in api response")
		}
	}
}

func TestAPIListProjectsSupportsFeaturedFilterAndRejectsInvalidSort(t *testing.T) {
	app := newPublicTestApp(t)

	_, err := app.svc.Projects.CreateProject(model.ProjectCreate{
		Title:     "Normal Project",
		ContentMD: "# Normal",
		Excerpt:   "normal project",
		Published: true,
		Featured:  false,
		SortOrder: 3,
	})
	if err != nil {
		t.Fatalf("create normal project: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/projects?featured=true", nil)
	rr := httptest.NewRecorder()
	app.handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}
	var projects []model.Project
	if err := json.NewDecoder(rr.Body).Decode(&projects); err != nil {
		t.Fatalf("decode featured projects: %v", err)
	}
	if len(projects) != 1 || projects[0].Title != "Whyred Watchtower" {
		t.Fatalf("featured projects = %+v, want only Whyred Watchtower", projects)
	}

	badReq := httptest.NewRequest(http.MethodGet, "/api/v1/projects?sort=sideways", nil)
	badRR := httptest.NewRecorder()
	app.handler.ServeHTTP(badRR, badReq)
	if badRR.Code != http.StatusBadRequest {
		t.Fatalf("bad sort status = %d, want %d; body: %s", badRR.Code, http.StatusBadRequest, badRR.Body.String())
	}
	assertContains(t, badRR.Body.String(), "invalid project sort")
}

func TestAPIProjectWriteEndpointsRequireAdminKey(t *testing.T) {
	app := newPublicTestApp(t)

	for _, tt := range []struct {
		method string
		path   string
	}{
		{method: http.MethodPost, path: "/api/v1/projects"},
		{method: http.MethodPut, path: "/api/v1/projects/1"},
		{method: http.MethodDelete, path: "/api/v1/projects/1"},
	} {
		t.Run(tt.method+" "+tt.path, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, strings.NewReader(`{}`))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()
			app.handler.ServeHTTP(rr, req)
			if rr.Code != http.StatusUnauthorized {
				t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusUnauthorized, rr.Body.String())
			}
		})
	}
}

func TestAPIProjectCreateUpdateDeleteFlow(t *testing.T) {
	app := newPublicTestApp(t)

	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/projects", strings.NewReader(`{"title":"API Project","content_md":"# API Project","excerpt":"Created by API","published":true,"featured":true,"sort_order":4,"tags":["api"," automation "]}`))
	createReq.Header.Set("Authorization", "Bearer admin-key")
	createReq.Header.Set("Content-Type", "application/json")
	createRR := httptest.NewRecorder()
	app.handler.ServeHTTP(createRR, createReq)
	if createRR.Code != http.StatusCreated {
		t.Fatalf("create status = %d, want %d; body: %s", createRR.Code, http.StatusCreated, createRR.Body.String())
	}
	var created model.Project
	if err := json.NewDecoder(createRR.Body).Decode(&created); err != nil {
		t.Fatalf("decode created project: %v", err)
	}
	if created.ID == 0 || created.Slug != "api-project" || !created.Featured || created.SortOrder != 4 {
		t.Fatalf("created project mismatch: %+v", created)
	}
	if len(created.Tags) != 2 {
		t.Fatalf("created tags = %+v, want 2", created.Tags)
	}

	updateReq := httptest.NewRequest(http.MethodPut, "/api/v1/projects/"+strconv.FormatInt(created.ID, 10), strings.NewReader(`{"title":"API Project Updated","content_md":"# API Updated","excerpt":"Updated by API","published":false,"featured":false,"sort_order":-5,"tags":["updated"]}`))
	updateReq.Header.Set("Authorization", "Bearer admin-key")
	updateReq.Header.Set("Content-Type", "application/json")
	updateRR := httptest.NewRecorder()
	app.handler.ServeHTTP(updateRR, updateReq)
	if updateRR.Code != http.StatusOK {
		t.Fatalf("update status = %d, want %d; body: %s", updateRR.Code, http.StatusOK, updateRR.Body.String())
	}
	var updated model.Project
	if err := json.NewDecoder(updateRR.Body).Decode(&updated); err != nil {
		t.Fatalf("decode updated project: %v", err)
	}
	if updated.Slug != created.Slug || updated.Title != "API Project Updated" || updated.Published || updated.Featured || updated.SortOrder != 0 {
		t.Fatalf("updated project mismatch: %+v", updated)
	}

	deleteReq := httptest.NewRequest(http.MethodDelete, "/api/v1/projects/"+strconv.FormatInt(created.ID, 10), nil)
	deleteReq.Header.Set("Authorization", "Bearer admin-key")
	deleteRR := httptest.NewRecorder()
	app.handler.ServeHTTP(deleteRR, deleteReq)
	if deleteRR.Code != http.StatusOK {
		t.Fatalf("delete status = %d, want %d; body: %s", deleteRR.Code, http.StatusOK, deleteRR.Body.String())
	}
	assertContains(t, deleteRR.Body.String(), `"status":"ok"`)

	listReq := httptest.NewRequest(http.MethodGet, "/api/v1/projects", nil)
	listRR := httptest.NewRecorder()
	app.handler.ServeHTTP(listRR, listReq)
	if listRR.Code != http.StatusOK {
		t.Fatalf("list status = %d, want %d", listRR.Code, http.StatusOK)
	}
	var projects []model.Project
	if err := json.NewDecoder(listRR.Body).Decode(&projects); err != nil {
		t.Fatalf("decode projects: %v", err)
	}
	for _, project := range projects {
		if project.ID == created.ID {
			t.Fatalf("deleted project still visible: %+v", project)
		}
	}
}

func TestAPIProjectWriteValidationAndNotFound(t *testing.T) {
	app := newPublicTestApp(t)

	badCreateReq := httptest.NewRequest(http.MethodPost, "/api/v1/projects", strings.NewReader(`{"title":"No Body"}`))
	badCreateReq.Header.Set("Authorization", "Bearer admin-key")
	badCreateReq.Header.Set("Content-Type", "application/json")
	badCreateRR := httptest.NewRecorder()
	app.handler.ServeHTTP(badCreateRR, badCreateReq)
	if badCreateRR.Code != http.StatusBadRequest {
		t.Fatalf("bad create status = %d, want %d; body: %s", badCreateRR.Code, http.StatusBadRequest, badCreateRR.Body.String())
	}

	badIDReq := httptest.NewRequest(http.MethodPut, "/api/v1/projects/not-a-number", strings.NewReader(`{"title":"x","content_md":"# x"}`))
	badIDReq.Header.Set("Authorization", "Bearer admin-key")
	badIDReq.Header.Set("Content-Type", "application/json")
	badIDRR := httptest.NewRecorder()
	app.handler.ServeHTTP(badIDRR, badIDReq)
	if badIDRR.Code != http.StatusBadRequest {
		t.Fatalf("bad id status = %d, want %d; body: %s", badIDRR.Code, http.StatusBadRequest, badIDRR.Body.String())
	}

	notFoundReq := httptest.NewRequest(http.MethodDelete, "/api/v1/projects/9999", nil)
	notFoundReq.Header.Set("Authorization", "Bearer admin-key")
	notFoundRR := httptest.NewRecorder()
	app.handler.ServeHTTP(notFoundRR, notFoundReq)
	if notFoundRR.Code != http.StatusNotFound {
		t.Fatalf("delete missing status = %d, want %d; body: %s", notFoundRR.Code, http.StatusNotFound, notFoundRR.Body.String())
	}
}

func TestProjectsPageSupportsFeaturedFilterAndSortQuery(t *testing.T) {
	app := newPublicTestApp(t)

	_, err := app.svc.Projects.CreateProject(model.ProjectCreate{
		Title:     "Old Featured",
		ContentMD: "# Old Featured",
		Excerpt:   "old featured project",
		Published: true,
		Featured:  true,
		SortOrder: 5,
	})
	if err != nil {
		t.Fatalf("create old featured project: %v", err)
	}
	_, err = app.svc.Projects.CreateProject(model.ProjectCreate{
		Title:     "Normal Project",
		ContentMD: "# Normal Project",
		Excerpt:   "normal project",
		Published: true,
		Featured:  false,
		SortOrder: 8,
	})
	if err != nil {
		t.Fatalf("create normal project: %v", err)
	}

	status, body := getBody(t, app, "/projects?featured=true&sort=oldest", nil)
	if status != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", status, http.StatusOK, body)
	}
	assertContains(t, body, `name="featured" value="true" checked`)
	assertContains(t, body, `<option value="oldest" selected>Oldest first</option>`)
	assertContains(t, body, "Old Featured")
	assertContains(t, body, "Whyred Watchtower")
	assertNotContains(t, body, "Normal Project")
	assertContains(t, body, `name="q" value=""`)

	searchStatus, searchBody := getBody(t, app, "/projects?q=watchtower", nil)
	if searchStatus != http.StatusOK {
		t.Fatalf("search status = %d, want %d; body: %s", searchStatus, http.StatusOK, searchBody)
	}
	assertContains(t, searchBody, `name="q" value="watchtower"`)
	assertContains(t, searchBody, "<mark")
	assertContains(t, searchBody, `href="/projects/whyred-watchtower"`)
	assertNotContains(t, searchBody, "Old Featured")

	comboStatus, comboBody := getBody(t, app, "/projects?featured=true&sort=oldest&q=watchtower", nil)
	if comboStatus != http.StatusOK {
		t.Fatalf("combo status = %d, want %d; body: %s", comboStatus, http.StatusOK, comboBody)
	}
	assertContains(t, comboBody, `name="q" value="watchtower"`)
	assertContains(t, comboBody, `name="featured" value="true" checked`)
	assertContains(t, comboBody, `<option value="oldest" selected>Oldest first</option>`)

	badStatus, badBody := getBody(t, app, "/projects?sort=sideways", nil)
	if badStatus != http.StatusBadRequest {
		t.Fatalf("bad sort status = %d, want %d; body: %s", badStatus, http.StatusBadRequest, badBody)
	}
	assertContains(t, badBody, "invalid project sort")
}

func TestUnifiedSearchPageSupportsScopes(t *testing.T) {
	app := newPublicTestApp(t)

	status, body := getBody(t, app, "/search?q=hello&scope=posts", nil)
	if status != http.StatusOK {
		t.Fatalf("posts scope status = %d, want %d; body: %s", status, http.StatusOK, body)
	}
	assertContains(t, body, "hello-raevtar")
	assertContains(t, body, "<mark")
	assertContains(t, body, "Posts")
	assertNotContains(t, body, "Whyred Watchtower")

	status, body = getBody(t, app, "/search?q=watchtower&scope=projects", nil)
	if status != http.StatusOK {
		t.Fatalf("projects scope status = %d, want %d; body: %s", status, http.StatusOK, body)
	}
	assertContains(t, body, "whyred-watchtower")
	assertContains(t, body, "<mark")
	assertContains(t, body, `value="projects" selected`)

	status, body = getBody(t, app, "/search?q=lightweight&scope=pages", nil)
	if status != http.StatusOK {
		t.Fatalf("pages scope status = %d, want %d; body: %s", status, http.StatusOK, body)
	}
	assertContains(t, body, `href="/contact"`)
	assertContains(t, body, "<mark")
	assertContains(t, body, `href="/contact"`)

	status, body = getBody(t, app, "/search?q=hello", nil)
	if status != http.StatusOK {
		t.Fatalf("all scope status = %d, want %d; body: %s", status, http.StatusOK, body)
	}
	assertContains(t, body, "total hits")
	assertContains(t, body, "hello-raevtar")

	badStatus, badBody := getBody(t, app, "/search?q=test&scope=servers", nil)
	if badStatus != http.StatusBadRequest {
		t.Fatalf("bad scope status = %d, want %d; body: %s", badStatus, http.StatusBadRequest, badBody)
	}
	assertContains(t, badBody, "invalid search scope")
}

func TestAPISearchReturnsGroupedPublicResults(t *testing.T) {
	app := newPublicTestApp(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/search?q=hello&scope=posts&page=1&page_size=5", nil)
	rr := httptest.NewRecorder()
	app.handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}
	var payload struct {
		Query    string              `json:"query"`
		Scope    string              `json:"scope"`
		Page     int                 `json:"page"`
		PageSize int                 `json:"page_size"`
		Counts   map[string]int      `json:"counts"`
		Posts    []model.Post        `json:"posts"`
		Projects []model.Project     `json:"projects"`
		Pages    []model.PageContent `json:"pages"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&payload); err != nil {
		t.Fatalf("decode payload: %v", err)
	}
	if payload.Query != "hello" || payload.Scope != "posts" || payload.Page != 1 || payload.PageSize != 5 {
		t.Fatalf("unexpected search payload meta: %+v", payload)
	}
	if payload.Counts["posts"] == 0 || len(payload.Posts) == 0 {
		t.Fatalf("expected post results, got %+v", payload)
	}
	if len(payload.Projects) != 0 || len(payload.Pages) != 0 {
		t.Fatalf("posts scope should not include other groups: %+v", payload)
	}

	badReq := httptest.NewRequest(http.MethodGet, "/api/v1/search?q=test&scope=servers", nil)
	badRR := httptest.NewRecorder()
	app.handler.ServeHTTP(badRR, badReq)
	if badRR.Code != http.StatusBadRequest {
		t.Fatalf("bad scope status = %d, want %d; body: %s", badRR.Code, http.StatusBadRequest, badRR.Body.String())
	}
	assertContains(t, badRR.Body.String(), "invalid search scope")

	badPageReq := httptest.NewRequest(http.MethodGet, "/api/v1/search?page=oops", nil)
	badPageRR := httptest.NewRecorder()
	app.handler.ServeHTTP(badPageRR, badPageReq)
	if badPageRR.Code != http.StatusBadRequest {
		t.Fatalf("bad page status = %d, want %d; body: %s", badPageRR.Code, http.StatusBadRequest, badPageRR.Body.String())
	}
	assertContains(t, badPageRR.Body.String(), "invalid page")
}

func TestUIPolishSourceHooks(t *testing.T) {
	files := map[string][]string{
		filepath.Join("..", "..", "internal", "view", "components", "post_card.templ"): {
			`nb-card group`,
		},
		filepath.Join("..", "..", "internal", "view", "components", "server_card.templ"): {
			`rv-card-lift`,
		},
		filepath.Join("..", "..", "internal", "view", "pages", "index.templ"): {
			`-rotate-1`,
			`marquee-container`,
			`marquee-content`,
		},
		filepath.Join("..", "..", "internal", "view", "pages", "lab.templ"): {
			`rv-hero-ambient`,
			`-rotate-1`,
		},
		filepath.Join("..", "..", "internal", "view", "pages", "dashboard.templ"): {
			`id="server-list"`,
			`hx-get="/dashboard"`,
			`hx-trigger="every 30s"`,
			`hx-select="#server-list"`,
			`hx-target="#server-list"`,
			`hx-indicator="#server-list"`,
		},
		filepath.Join("..", "..", "internal", "view", "pages", "server_detail.templ"): {
			`id="server-detail-live"`,
			`hx-get={ templ.URL("/dashboard/" + IDText(data.Server.ID) + "/live") }`,
			`hx-trigger="every 15s"`,
			`hx-swap="outerHTML"`,
			`hx-indicator="#server-detail-live"`,
		},
		filepath.Join("..", "..", "static", "css", "tailwind.src.css"): {
			`@tailwind base`,
			`@tailwind components`,
			`@tailwind utilities`,
			`@keyframes marquee`,
			`.animate-marquee`,
			`.marquee-container`,
			`.marquee-content`,
			`.nb-border`,
			`.nb-shadow`,
		},
	}

	for path, wants := range files {
		t.Run(path, func(t *testing.T) {
			content, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("read %s: %v", path, err)
			}
			body := string(content)
			for _, want := range wants {
				assertContains(t, body, want)
			}
		})
	}
}

func TestLandingUsesTotalPublishedPostCount(t *testing.T) {
	app := newPublicTestApp(t)

	for i := 2; i <= 5; i++ {
		_, err := app.svc.Blog.CreatePost(model.PostCreate{
			CategorySlug: "devops",
			Title:        fmt.Sprintf("Counted Dispatch %d", i),
			ContentMD:    fmt.Sprintf("# Counted Dispatch %d\n\nPublished count regression fixture.", i),
			Excerpt:      "Published count regression fixture.",
			Published:    true,
		})
		if err != nil {
			t.Fatalf("create post %d: %v", i, err)
		}
	}

	status, body := getBody(t, app, "/", nil)
	if status != http.StatusOK {
		t.Fatalf("status = %d, want %d", status, http.StatusOK)
	}
	assertContains(t, body, `>5</span>`)
	assertContains(t, body, `<span class="text-xs font-black uppercase tracking-widest">Posts</span>`)
}

func TestPublicDashboardRedactsServerTopology(t *testing.T) {
	app := newPublicTestApp(t)

	status, body := getBody(t, app, "/dashboard", nil)
	if status != http.StatusOK {
		t.Fatalf("status = %d, want %d", status, http.StatusOK)
	}

	assertContains(t, body, "whyred")
	assertContains(t, body, "Hidden on public view")
	assertContains(t, body, "redacted")
	assertContains(t, body, "tagged node")
	assertContains(t, body, "last agent signal")
	assertContains(t, body, "grid refresh every 30s")
	assertContains(t, body, "Online")
	assertContains(t, body, "Stale")
	assertContains(t, body, "Offline")
	assertContains(t, body, "Platform System Health")
	assertContains(t, body, "N/A")
	assertNotContains(t, body, "127.0.0.1")
	assertNotContains(t, body, "127.0.0.1:9100")
	assertNotContains(t, body, ">9100<")
	assertNotContains(t, body, ">local<")
	assertNotContains(t, body, "agent token")
	assertNotContains(t, body, "raevtar-agent.sh")
	assertNotContains(t, body, "POST /api/v1/servers")
	assertNotContains(t, body, "Register Server")
	assertNotContains(t, body, "host or IP")
	assertNotContains(t, body, "tags (comma separated)")
	assertNotContains(t, body, `value="9100"`)
}

func TestPublicLabRedactsPrivateTopologyAndGuidance(t *testing.T) {
	app := newPublicTestApp(t)

	status, body := getBody(t, app, "/lab", nil)
	if status != http.StatusOK {
		t.Fatalf("status = %d, want %d", status, http.StatusOK)
	}

	for _, want := range []string{"Public lab bench", "Signal Lab", "Content Lab", "Automation Lab"} {
		assertContains(t, body, want)
	}
	assertNotContains(t, body, "127.0.0.1")
	assertNotContains(t, body, "127.0.0.1:9100")
	assertNotContains(t, body, ">9100<")
	assertNotContains(t, body, ">local<")
	assertNotContains(t, body, "agent token")
	assertNotContains(t, body, "raevtar-agent.sh")
	assertNotContains(t, body, "POST /api/v1/servers")
	assertNotContains(t, body, "host or IP")
	assertNotContains(t, body, "tags (comma separated)")
	assertNotContains(t, body, `value="9100"`)
}

func TestPublicDocsRedactPrivateTopologyAndGuidance(t *testing.T) {
	app := newPublicTestApp(t)

	for _, path := range []string{"/docs", "/lab/docs"} {
		t.Run(path, func(t *testing.T) {
			status, body := getBody(t, app, path, nil)
			if status != http.StatusOK {
				t.Fatalf("status = %d, want %d", status, http.StatusOK)
			}
			for _, want := range []string{"Public docs", "GET /api/v1/posts", "GET /api/v1/categories"} {
				assertContains(t, body, want)
			}
			for _, leak := range []string{"127.0.0.1", "127.0.0.1:9100", ">9100<", ">local<", "agent token", "raevtar-agent.sh", "POST /api/v1/servers", "POST /api/v1/servers/1/ping", "host or IP", "tags (comma separated)", `value="9100"`} {
				assertNotContains(t, body, leak)
			}
		})
	}
}

func TestLimitedRolesDashboardRedactsServerTopology(t *testing.T) {
	app := newPublicTestApp(t)

	for _, role := range []string{model.RoleOperator, model.RoleReadonly} {
		t.Run(role, func(t *testing.T) {
			cookie := &http.Cookie{Name: sessionCookieName, Value: sessions.create(103, role, role)}
			status, body := getBody(t, app, "/dashboard", cookie)
			if status != http.StatusOK {
				t.Fatalf("status = %d, want %d", status, http.StatusOK)
			}

			assertContains(t, body, "Hidden on public view")
			assertContains(t, body, "redacted")
			assertContains(t, body, "last agent signal")
			assertContains(t, body, "grid refresh every 30s")
			assertContains(t, body, "Online")
			assertContains(t, body, "Stale")
			assertContains(t, body, "Offline")

			assertNotContains(t, body, "127.0.0.1")
			assertNotContains(t, body, "127.0.0.1:9100")
			assertNotContains(t, body, ">9100<")
			assertNotContains(t, body, ">local<")
			assertNotContains(t, body, "agent token")
			assertNotContains(t, body, "raevtar-agent.sh")
			assertNotContains(t, body, "POST /api/v1/servers")
			assertNotContains(t, body, "Register Server")
			assertNotContains(t, body, "host or IP")
			assertNotContains(t, body, "tags (comma separated)")
			assertNotContains(t, body, `value="9100"`)
		})
	}
}

func TestOwnerDashboardKeepsServerTopologyInAdminOnly(t *testing.T) {
	app := newPublicTestApp(t)
	cookie := &http.Cookie{Name: sessionCookieName, Value: sessions.create(101, "owner", model.RoleOwner)}

	status, body := getBody(t, app, "/dashboard", cookie)
	if status != http.StatusOK {
		t.Fatalf("status = %d, want %d", status, http.StatusOK)
	}

	assertContains(t, body, "Hidden on public view")
	assertContains(t, body, "redacted")
	assertNotContains(t, body, "127.0.0.1")
	assertNotContains(t, body, "127.0.0.1:9100")
	assertNotContains(t, body, ">9100<")
	assertNotContains(t, body, ">local<")
}

func TestDashboardRegisterControlsAreRoleGated(t *testing.T) {
	app := newPublicTestApp(t)
	roles := []struct {
		name         string
		role         string
		wantRegister bool
	}{
		{name: "public", wantRegister: false},
		{name: "operator", role: model.RoleOperator, wantRegister: false},
		{name: "readonly", role: model.RoleReadonly, wantRegister: false},
		{name: "owner", role: model.RoleOwner, wantRegister: false},
		{name: "admin", role: model.RoleAdmin, wantRegister: false},
	}

	for _, tt := range roles {
		t.Run(tt.name, func(t *testing.T) {
			var cookie *http.Cookie
			if tt.role != "" {
				cookie = &http.Cookie{Name: sessionCookieName, Value: sessions.create(100, tt.name, tt.role)}
			}

			status, body := getBody(t, app, "/dashboard", cookie)
			if status != http.StatusOK {
				t.Fatalf("status = %d, want %d", status, http.StatusOK)
			}

			if tt.wantRegister {
				assertContains(t, body, "Register Server")
				assertContains(t, body, `action="/admin/servers"`)
				assertNotContains(t, body, `hx-post="/api/v1/servers"`)
			} else {
				assertNotContains(t, body, "Register Server")
				assertNotContains(t, body, `hx-post="/api/v1/servers"`)
				assertNotContains(t, body, `action="/admin/servers"`)
				assertNotContains(t, body, "host or IP")
				assertNotContains(t, body, "tags (comma separated)")
				assertNotContains(t, body, `value="9100"`)
			}
		})
	}
}

func TestPublicServerDetailHidesPingGuidance(t *testing.T) {
	app := newPublicTestApp(t)

	status, body := getBody(t, app, "/dashboard/1", nil)
	if status != http.StatusOK {
		t.Fatalf("status = %d, want %d", status, http.StatusOK)
	}
	assertNotContains(t, body, "POST /api/v1/servers/1/ping")
	assertNotContains(t, body, "POST /api/v1/servers/{id}/ping")
	assertNotContains(t, body, "POST /api/v1/servers/")
}

func TestPublicServerDetailRedactsServerTopology(t *testing.T) {
	app := newPublicTestApp(t)

	status, body := getBody(t, app, "/dashboard/1", nil)
	if status != http.StatusOK {
		t.Fatalf("status = %d, want %d", status, http.StatusOK)
	}

	assertContains(t, body, "whyred")
	assertContains(t, body, "Endpoint hidden on public view.")
	assertContains(t, body, "Use admin diagnostics for network details.")
	assertNotContains(t, body, "127.0.0.1")
	assertNotContains(t, body, "127.0.0.1:9100")
	assertNotContains(t, body, "port 9100")
	assertNotContains(t, body, ">local<")
}

func TestLimitedRolesServerDetailRedactsServerTopology(t *testing.T) {
	app := newPublicTestApp(t)

	for _, role := range []string{model.RoleOperator, model.RoleReadonly} {
		t.Run(role, func(t *testing.T) {
			cookie := &http.Cookie{Name: sessionCookieName, Value: sessions.create(104, role, role)}
			status, body := getBody(t, app, "/dashboard/1", cookie)
			if status != http.StatusOK {
				t.Fatalf("status = %d, want %d", status, http.StatusOK)
			}

			assertContains(t, body, "Endpoint hidden on public view.")
			assertContains(t, body, "Use admin diagnostics for network details.")

			assertNotContains(t, body, "127.0.0.1")
			assertNotContains(t, body, "127.0.0.1:9100")
			assertNotContains(t, body, "port 9100")
			assertNotContains(t, body, ">local<")
			assertNotContains(t, body, "agent token")
			assertNotContains(t, body, "raevtar-agent.sh")
			assertNotContains(t, body, "POST /api/v1/servers")
		})
	}
}

func TestOwnerServerDetailKeepsTopologyInAdminOnly(t *testing.T) {
	app := newPublicTestApp(t)
	cookie := &http.Cookie{Name: sessionCookieName, Value: sessions.create(102, "owner", model.RoleOwner)}

	status, body := getBody(t, app, "/dashboard/1", cookie)
	if status != http.StatusOK {
		t.Fatalf("status = %d, want %d", status, http.StatusOK)
	}

	assertContains(t, body, "Endpoint hidden on public view.")
	assertNotContains(t, body, "127.0.0.1")
	assertNotContains(t, body, "127.0.0.1:9100")
	assertNotContains(t, body, "port 9100")
	assertNotContains(t, body, ">local<")
}

func TestPublicServerDetailLiveRendersFragmentWithPollingAndRedaction(t *testing.T) {
	app := newPublicTestApp(t)
	if err := app.svc.Monitor.RecordMetrics(app.serverID, model.ServerMetric{
		CPUPercent:    12.5,
		RAMUsedMB:     256,
		RAMTotalMB:    1024,
		DiskUsedGB:    8.5,
		UptimeSeconds: 3600,
		Online:        true,
	}); err != nil {
		t.Fatalf("record metrics: %v", err)
	}

	status, body := getBody(t, app, "/dashboard/1/live", nil)
	if status != http.StatusOK {
		t.Fatalf("status = %d, want %d", status, http.StatusOK)
	}

	assertContains(t, body, `id="server-detail-live"`)
	assertContains(t, body, `hx-get="/dashboard/1/live"`)
	assertContains(t, body, `hx-trigger="every 15s"`)
	assertContains(t, body, `hx-swap="outerHTML"`)
	assertContains(t, body, "freshness diagnosis")
	assertContains(t, body, "last signal age")
	assertContains(t, body, "System Health")
	assertContains(t, body, "Panel refreshed")
	assertContains(t, body, "Telemetry is fresh. Latest agent signal arrived recently.")
	assertContains(t, body, "Endpoint hidden on public view.")
	assertContains(t, body, "Use admin diagnostics for network details.")
	for _, want := range []string{"CPU", "RAM", "Disk", "Uptime", "12.5%", "256.0 / 1024.0 MB", "8.5 GB", "1h 0m"} {
		assertContains(t, body, want)
	}
	assertNotContains(t, body, "<!DOCTYPE html>")
	assertNotContains(t, body, "<html")
	assertNotContains(t, body, "127.0.0.1")
	assertNotContains(t, body, "127.0.0.1:9100")
	assertNotContains(t, body, "port 9100")
	assertNotContains(t, body, ">local<")
	assertNotContains(t, body, "POST /api/v1/servers/")
}

func TestLimitedRolesServerDetailLiveRedactsServerTopology(t *testing.T) {
	app := newPublicTestApp(t)

	for _, role := range []string{model.RoleOperator, model.RoleReadonly} {
		t.Run(role, func(t *testing.T) {
			cookie := &http.Cookie{Name: sessionCookieName, Value: sessions.create(104, role, role)}
			status, body := getBody(t, app, "/dashboard/1/live", cookie)
			if status != http.StatusOK {
				t.Fatalf("status = %d, want %d", status, http.StatusOK)
			}

			assertContains(t, body, "Endpoint hidden on public view.")
			assertContains(t, body, "No telemetry received yet. Waiting for first agent signal.")
			assertContains(t, body, "System Health")
			assertContains(t, body, "N/A")
			assertNotContains(t, body, "127.0.0.1")
			assertNotContains(t, body, "127.0.0.1:9100")
			assertNotContains(t, body, "port 9100")
			assertNotContains(t, body, ">local<")
			assertNotContains(t, body, `href="/admin/servers"`)
			assertNotContains(t, body, "agent token")
			assertNotContains(t, body, "raevtar-agent.sh")
			assertNotContains(t, body, "POST /api/v1/servers")
		})
	}
}

func TestOwnerAndAdminServerDetailLiveKeepsTopologyInAdminOnly(t *testing.T) {
	app := newPublicTestApp(t)

	for _, role := range []string{model.RoleOwner, model.RoleAdmin} {
		t.Run(role, func(t *testing.T) {
			cookie := &http.Cookie{Name: sessionCookieName, Value: sessions.create(102, role, role)}
			status, body := getBody(t, app, "/dashboard/1/live", cookie)
			if status != http.StatusOK {
				t.Fatalf("status = %d, want %d", status, http.StatusOK)
			}

			assertContains(t, body, "Endpoint hidden on public view.")
			assertContains(t, body, "System Health")
			assertContains(t, body, "N/A")
			assertNotContains(t, body, "127.0.0.1")
			assertNotContains(t, body, "127.0.0.1:9100")
			assertNotContains(t, body, "port 9100")
			assertNotContains(t, body, ">local<")
			assertNotContains(t, body, `href="/admin/servers"`)
		})
	}
}

func TestPublicServerDetailShowsSafeMetricValues(t *testing.T) {
	app := newPublicTestApp(t)
	if err := app.svc.Monitor.RecordMetrics(app.serverID, model.ServerMetric{
		CPUPercent:    12.5,
		RAMUsedMB:     256,
		RAMTotalMB:    1024,
		DiskUsedGB:    8.5,
		UptimeSeconds: 3600,
		Online:        true,
	}); err != nil {
		t.Fatalf("record metrics: %v", err)
	}

	status, body := getBody(t, app, "/dashboard/1", nil)
	if status != http.StatusOK {
		t.Fatalf("status = %d, want %d", status, http.StatusOK)
	}
	assertContains(t, body, "System Health")
	for _, want := range []string{"CPU", "RAM", "Disk", "Uptime", "12.5%", "256.0 / 1024.0 MB", "8.5 GB", "1h 0m"} {
		assertContains(t, body, want)
	}
	for _, leak := range []string{"127.0.0.1", "127.0.0.1:9100", "port 9100", ">local<", "agent token", "raevtar-agent.sh", "POST /api/v1/servers", "install", "cron", "audit"} {
		assertNotContains(t, body, leak)
	}
}

func TestOwnerServerDetailShowsSamePublicSafeMetricValues(t *testing.T) {
	app := newPublicTestApp(t)
	if err := app.svc.Monitor.RecordMetrics(app.serverID, model.ServerMetric{
		CPUPercent:    12.5,
		RAMUsedMB:     256,
		RAMTotalMB:    1024,
		DiskUsedGB:    8.5,
		UptimeSeconds: 3600,
		Online:        true,
	}); err != nil {
		t.Fatalf("record metrics: %v", err)
	}

	cookie := &http.Cookie{Name: sessionCookieName, Value: sessions.create(106, "owner", model.RoleOwner)}
	status, body := getBody(t, app, "/dashboard/1", cookie)
	if status != http.StatusOK {
		t.Fatalf("status = %d, want %d", status, http.StatusOK)
	}
	for _, want := range []string{"System Health", "CPU", "RAM", "Disk", "Uptime", "12.5%", "256.0 / 1024.0 MB", "8.5 GB", "1h 0m"} {
		assertContains(t, body, want)
	}
}

func TestAgentTokenAuthorizesMetrics(t *testing.T) {
	app := newPublicTestApp(t)
	token, err := app.svc.Monitor.RotateAgentToken(app.serverID)
	if err != nil {
		t.Fatalf("rotate token: %v", err)
	}

	payload := []byte(`{"cpu_percent":14.5,"ram_used_mb":512,"ram_total_mb":2048,"disk_used_gb":12,"uptime_seconds":90,"online":true}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/servers/1/ping", bytes.NewReader(payload))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	app.handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}
	metrics, err := app.svc.Monitor.GetRecentMetrics(app.serverID, 1)
	if err != nil {
		t.Fatalf("metrics: %v", err)
	}
	if len(metrics) != 1 || metrics[0].CPUPercent != 14.5 {
		t.Fatalf("metrics = %+v, want recorded payload", metrics)
	}
}

func TestInvalidAgentTokenRejectsMetrics(t *testing.T) {
	app := newPublicTestApp(t)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/servers/1/ping", strings.NewReader(`{"online":true}`))
	req.Header.Set("Authorization", "Bearer wrong-token")
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	app.handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusUnauthorized)
	}
}

func TestAdminKeyStillAuthorizesMetrics(t *testing.T) {
	app := newPublicTestApp(t)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/servers/1/ping", strings.NewReader(`{"online":true}`))
	req.Header.Set("Authorization", "Bearer admin-key")
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	app.handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}
}

func TestMissingServerPingReturnsNotFound(t *testing.T) {
	app := newPublicTestApp(t)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/servers/999/ping", strings.NewReader(`{"online":true}`))
	req.Header.Set("Authorization", "Bearer admin-key")
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	app.handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusNotFound, rr.Body.String())
	}
}

func TestAPICreateServerReturnsAgentToken(t *testing.T) {
	app := newPublicTestApp(t)

	payload := []byte(`{"name":"api-agent","host":"192.168.100.77","port":9100,"tags":"api"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/servers", bytes.NewReader(payload))
	req.Header.Set("Authorization", "Bearer admin-key")
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	app.handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusCreated, rr.Body.String())
	}
	var response struct {
		Server     model.Server `json:"server"`
		AgentToken string       `json:"agent_token"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.Server.ID == 0 || response.Server.Name != "api-agent" {
		t.Fatalf("server response = %+v, want created api-agent", response.Server)
	}
	if response.Server.AgentTokenHash != "" {
		t.Fatalf("server JSON leaked token hash: %+v", response.Server)
	}
	if response.AgentToken == "" {
		t.Fatalf("expected one-time agent token")
	}
	if !app.svc.Monitor.VerifyAgentToken(response.Server.ID, response.AgentToken) {
		t.Fatalf("returned agent token should verify")
	}
}

func TestAdminRotateMissingServerReturnsNotFound(t *testing.T) {
	app := newPublicTestApp(t)
	token := sessions.create(50, "owner", model.RoleOwner)
	entry, ok := sessions.get(token)
	if !ok {
		t.Fatalf("expected session")
	}

	form := url.Values{"_csrf": {entry.csrfToken}}
	req := httptest.NewRequest(http.MethodPost, "/admin/servers/rotate-token/999", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: token})
	rr := httptest.NewRecorder()

	app.handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusNotFound, rr.Body.String())
	}
}

func TestAdminCreateServerShowsAgentInstallToken(t *testing.T) {
	app := newPublicTestApp(t)
	token := sessions.create(49, "owner", model.RoleOwner)
	entry, ok := sessions.get(token)
	if !ok {
		t.Fatalf("expected session")
	}

	form := url.Values{
		"name":  {"agent-node"},
		"host":  {"192.168.100.50"},
		"port":  {"9100"},
		"tags":  {"lan,agent"},
		"_csrf": {entry.csrfToken},
	}
	req := httptest.NewRequest(http.MethodPost, "/admin/servers", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: token})
	rr := httptest.NewRecorder()

	app.handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusSeeOther, rr.Body.String())
	}
	status, body := getBody(t, app, "/admin/servers", &http.Cookie{Name: sessionCookieName, Value: token})
	if status != http.StatusOK {
		t.Fatalf("status = %d, want %d", status, http.StatusOK)
	}
	assertContains(t, body, "One-time agent token")
	assertContains(t, body, "Copy before leaving or reloading")
	assertContains(t, body, "Copy token")
	assertContains(t, body, `data-copy-target="server-2-agent-token"`)
	assertContains(t, body, `id="server-2-agent-token"`)
	assertContains(t, body, "Copy install command")
	assertContains(t, body, "Copy run command")
	assertContains(t, body, "Copy cron line")
	assertContains(t, body, `data-copy-target="server-2-install-command"`)
	assertContains(t, body, `data-copy-target="server-2-run-command"`)
	assertContains(t, body, `data-copy-target="server-2-cron-command"`)
	assertContains(t, body, `data-confirm="Rotate agent token?`)
	assertContains(t, body, `data-confirm="Delete this server?`)
	assertContains(t, body, `data-confirm="Register this server and issue a one-time agent token?"`)
	assertContains(t, body, "raevtar-agent.sh")
	assertContains(t, body, "RAEVTAR_URL=https://raevtar.test")
	assertContains(t, body, "RAEVTAR_SERVER_ID=2")
	assertContains(t, body, "RAEVTAR_AGENT_TOKEN=")
	assertNotContains(t, body, `data-copy-target="`+app.svc.Cfg.AdminKey)

	_, secondBody := getBody(t, app, "/admin/servers", &http.Cookie{Name: sessionCookieName, Value: token})
	assertNotContains(t, secondBody, "One-time agent token")
}

func TestAdminServersEmptyStateShowsSetupAffordances(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "raevtar_empty_servers.db")
	cfg := &config.Config{
		DatabasePath: dbPath,
		Domain:       "raevtar.test",
		AdminKey:     "admin-key",
		AdminUser:    "admin",
		AdminPass:    "demo-pass-123",
	}
	db := repo.InitSQLite(cfg.DatabasePath)
	t.Cleanup(func() { _ = db.Close() })
	repo.AutoMigrate(db)
	svc := service.New(repo.New(db), cfg)
	if err := svc.SeedData(); err != nil {
		t.Fatalf("seed data: %v", err)
	}

	app := &publicTestApp{handler: New(svc, cfg), svc: svc}
	token := sessions.create(56, "owner", model.RoleOwner)
	status, body := getBody(t, app, "/admin/servers", &http.Cookie{Name: sessionCookieName, Value: token})
	if status != http.StatusOK {
		t.Fatalf("status = %d, want %d", status, http.StatusOK)
	}

	for _, want := range []string{
		"No servers registered yet.",
		"Register the first node below",
		"Jump to registration form",
		`href="#register-server"`,
		`id="register-server"`,
		`data-confirm="Register this server and issue a one-time agent token?"`,
		"Register New Server",
	} {
		assertContains(t, body, want)
	}
	for _, leak := range []string{"One-time agent token", "Copy token", `data-copy-target="server-`, "Copy install command", "Copy run command", "Copy cron line", "RAEVTAR_AGENT_TOKEN=", "paste-token-here", "raevtar-agent.sh"} {
		assertNotContains(t, body, leak)
	}
}

func TestPublicPagesDoNotLeakAdminAgentSetupAffordances(t *testing.T) {
	app := newPublicTestApp(t)

	for _, path := range []string{"/dashboard", "/dashboard/1", "/lab", "/docs", "/lab/docs"} {
		t.Run(path, func(t *testing.T) {
			status, body := getBody(t, app, path, nil)
			if status != http.StatusOK {
				t.Fatalf("status = %d, want %d", status, http.StatusOK)
			}

			for _, leak := range []string{
				"RAEVTAR_AGENT_TOKEN=",
				"paste-token-here",
				"raevtar-agent.sh",
				"Copy install command",
				"Copy run command",
				"Copy cron line",
				"One-time agent token",
				"data-copy-target",
			} {
				assertNotContains(t, body, leak)
			}
		})
	}
}

func TestAdminServerDiagnosticsAreOwnerOnlyAndShowPrivateDetails(t *testing.T) {
	app := newPublicTestApp(t)
	ownerToken := sessions.create(51, "owner", model.RoleOwner)
	readonlyToken := sessions.create(52, "reader", model.RoleReadonly)
	if err := app.svc.Monitor.RecordMetrics(app.serverID, model.ServerMetric{
		CPUPercent:    12.5,
		RAMUsedMB:     256,
		RAMTotalMB:    1024,
		DiskUsedGB:    8.5,
		UptimeSeconds: 3600,
		Online:        true,
	}); err != nil {
		t.Fatalf("record metrics: %v", err)
	}

	status, body := getBody(t, app, "/admin/servers/1", &http.Cookie{Name: sessionCookieName, Value: ownerToken})
	if status != http.StatusOK {
		t.Fatalf("owner status = %d, want %d", status, http.StatusOK)
	}
	for _, want := range []string{"whyred", "127.0.0.1", "9100", "local", "Remote Command Center", "Node Metadata", "Audit Logs", "Agent Config", "Restart Agent", "Clear Cache", "Reboot Node", "Update Agent", "Rotate Token", "Ping Endpoint", "Agent Secret Token", "token missing"} {
		assertContains(t, body, want)
	}
	for _, want := range []string{
		`data-confirm="Rotate token?`,
		"Server Detail: whyred",
	} {
		assertContains(t, body, want)
	}
	assertNotContains(t, body, "One-time agent token")
	assertNotContains(t, body, "agent_token_hash")

	status, body = getBody(t, app, "/admin/servers/1", &http.Cookie{Name: sessionCookieName, Value: readonlyToken})
	if status != http.StatusForbidden {
		t.Fatalf("readonly status = %d, want %d", status, http.StatusForbidden)
	}
	assertNotContains(t, body, "127.0.0.1")
	assertNotContains(t, body, "raevtar-agent.sh")
}

func TestAdminPostEditUpdatesExistingPost(t *testing.T) {
	app := newPublicTestApp(t)
	token := sessions.create(54, "owner", model.RoleOwner)
	entry, ok := sessions.get(token)
	if !ok {
		t.Fatalf("expected session")
	}
	post, err := app.svc.Blog.GetPost("hello-raevtar")
	if err != nil {
		t.Fatalf("get post: %v", err)
	}

	status, body := getBody(t, app, "/admin/posts/edit/"+strconv.FormatInt(post.ID, 10), &http.Cookie{Name: sessionCookieName, Value: token})
	if status != http.StatusOK {
		t.Fatalf("edit status = %d, want %d", status, http.StatusOK)
	}
	assertContains(t, body, "Edit post")
	assertContains(t, body, "Hello Raevtar")
	assertContains(t, body, `data-confirm="Save post changes?"`)
	assertContains(t, body, "Current status")
	assertContains(t, body, `name="intent" value="draft"`)
	assertContains(t, body, `name="intent" value="update"`)
	assertContains(t, body, `name="intent" value="publish"`)
	assertContains(t, body, `hx-post="/admin/posts/preview"`)
	assertContains(t, body, `hx-target="#post-markdown-preview"`)
	assertContains(t, body, `id="post-markdown-preview"`)
	assertNotContains(t, body, `name="published"`)

	form := url.Values{
		"_csrf":         {entry.csrfToken},
		"title":         {"Updated Dispatch"},
		"category_slug": {"tools"},
		"content":       {"# Updated Dispatch"},
		"excerpt":       {"Updated excerpt"},
		"tags":          {"new, commissioned"},
		"intent":        {"draft"},
	}
	req := httptest.NewRequest(http.MethodPost, "/admin/posts/update/"+strconv.FormatInt(post.ID, 10), strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: token})
	rr := httptest.NewRecorder()
	app.handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusSeeOther {
		t.Fatalf("update status = %d, want %d; body: %s", rr.Code, http.StatusSeeOther, rr.Body.String())
	}
	updated, err := app.svc.Blog.GetPost("hello-raevtar")
	if err != nil {
		t.Fatalf("get updated post by preserved slug: %v", err)
	}
	if updated.Title != "Updated Dispatch" || updated.CategorySlug != "tools" || updated.Excerpt != "Updated excerpt" || updated.Published {
		t.Fatalf("updated post = %+v", updated)
	}
	if len(updated.Tags) != 2 {
		t.Fatalf("updated tags = %+v, want 2", updated.Tags)
	}
	status, body = getBody(t, app, "/admin/posts", &http.Cookie{Name: sessionCookieName, Value: token})
	if status != http.StatusOK {
		t.Fatalf("posts status = %d, want %d", status, http.StatusOK)
	}
	assertContains(t, body, "Updated Dispatch")
	assertContains(t, body, "draft")
}

func TestAdminPostCreateIntentControlsPublishedState(t *testing.T) {
	app := newPublicTestApp(t)
	token := sessions.create(64, "owner", model.RoleOwner)
	entry, ok := sessions.get(token)
	if !ok {
		t.Fatalf("expected session")
	}

	status, body := getBody(t, app, "/admin/posts", &http.Cookie{Name: sessionCookieName, Value: token})
	if status != http.StatusOK {
		t.Fatalf("posts status = %d, want %d", status, http.StatusOK)
	}
	assertContains(t, body, `name="intent" value="draft"`)
	assertContains(t, body, `name="intent" value="publish"`)
	assertContains(t, body, `hx-post="/admin/posts/preview"`)
	assertContains(t, body, `hx-target="#post-markdown-preview"`)
	assertContains(t, body, `id="post-markdown-preview"`)
	assertNotContains(t, body, `name="published"`)

	for _, tt := range []struct {
		name      string
		title     string
		intent    string
		published bool
	}{
		{name: "draft", title: "Admin Draft Dispatch", intent: "draft", published: false},
		{name: "published", title: "Admin Published Dispatch", intent: "publish", published: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			form := url.Values{
				"_csrf":         {entry.csrfToken},
				"title":         {tt.title},
				"category_slug": {"devops"},
				"content":       {"# " + tt.title},
				"excerpt":       {tt.title + " excerpt"},
				"intent":        {tt.intent},
			}
			req := httptest.NewRequest(http.MethodPost, "/admin/posts", strings.NewReader(form.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: token})
			rr := httptest.NewRecorder()
			app.handler.ServeHTTP(rr, req)
			if rr.Code != http.StatusSeeOther {
				t.Fatalf("create status = %d, want %d; body: %s", rr.Code, http.StatusSeeOther, rr.Body.String())
			}
			post, err := app.svc.Blog.GetPost(strings.ToLower(strings.ReplaceAll(tt.title, " ", "-")))
			if err != nil {
				t.Fatalf("get post: %v", err)
			}
			if post.Published != tt.published {
				t.Fatalf("published = %v, want %v", post.Published, tt.published)
			}
		})
	}
}

func TestAdminPostPreviewRequiresSessionOwnerAdminAndCSRF(t *testing.T) {
	app := newPublicTestApp(t)

	form := url.Values{"content": {"# Preview"}}
	unauthReq := httptest.NewRequest(http.MethodPost, "/admin/posts/preview", strings.NewReader(form.Encode()))
	unauthReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	unauthRR := httptest.NewRecorder()
	app.handler.ServeHTTP(unauthRR, unauthReq)
	if unauthRR.Code != http.StatusSeeOther {
		t.Fatalf("unauthenticated status = %d, want %d", unauthRR.Code, http.StatusSeeOther)
	}
	if got := unauthRR.Header().Get("Location"); got != "/admin/login" {
		t.Fatalf("unauthenticated Location = %q, want /admin/login", got)
	}

	readerToken := sessions.create(66, "reader", model.RoleReadonly)
	readerEntry, ok := sessions.get(readerToken)
	if !ok {
		t.Fatalf("expected reader session")
	}
	readerForm := url.Values{"_csrf": {readerEntry.csrfToken}, "content": {"# Preview"}}
	readerReq := httptest.NewRequest(http.MethodPost, "/admin/posts/preview", strings.NewReader(readerForm.Encode()))
	readerReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	readerReq.AddCookie(&http.Cookie{Name: sessionCookieName, Value: readerToken})
	readerRR := httptest.NewRecorder()
	app.handler.ServeHTTP(readerRR, readerReq)
	if readerRR.Code != http.StatusForbidden {
		t.Fatalf("readonly status = %d, want %d", readerRR.Code, http.StatusForbidden)
	}

	ownerToken := sessions.create(67, "owner", model.RoleOwner)
	missingCSRFReq := httptest.NewRequest(http.MethodPost, "/admin/posts/preview", strings.NewReader(form.Encode()))
	missingCSRFReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	missingCSRFReq.AddCookie(&http.Cookie{Name: sessionCookieName, Value: ownerToken})
	missingCSRFRR := httptest.NewRecorder()
	app.handler.ServeHTTP(missingCSRFRR, missingCSRFReq)
	if missingCSRFRR.Code != http.StatusForbidden {
		t.Fatalf("missing CSRF status = %d, want %d", missingCSRFRR.Code, http.StatusForbidden)
	}
	assertContains(t, missingCSRFRR.Body.String(), "invalid CSRF token")
}

func TestAdminPostPreviewRendersMarkdownHTML(t *testing.T) {
	app := newPublicTestApp(t)
	token := sessions.create(68, "owner", model.RoleOwner)
	entry, ok := sessions.get(token)
	if !ok {
		t.Fatalf("expected session")
	}

	form := url.Values{
		"_csrf":   {entry.csrfToken},
		"content": {"# Preview Title\n\n- **one**\n- two"},
	}
	req := httptest.NewRequest(http.MethodPost, "/admin/posts/preview", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: token})
	rr := httptest.NewRecorder()
	app.handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("preview status = %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}
	body := rr.Body.String()
	assertContains(t, body, `id="post-markdown-preview"`)
	assertContains(t, body, `<h1>Preview Title</h1>`)
	assertContains(t, body, `<strong>one</strong>`)
	assertContains(t, body, `<ul>`)
	assertNotContains(t, body, "Write Markdown, then preview without saving.")

	form.Set("content", "   ")
	req = httptest.NewRequest(http.MethodPost, "/admin/posts/preview", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: token})
	rr = httptest.NewRecorder()
	app.handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("blank preview status = %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}
	assertContains(t, rr.Body.String(), "Write Markdown, then preview without saving.")
}

func TestAdminMediaUploadStoresServesAndCanCoverPost(t *testing.T) {
	app := newPublicTestApp(t)
	token := sessions.create(69, "owner", model.RoleOwner)
	entry, ok := sessions.get(token)
	if !ok {
		t.Fatalf("expected session")
	}

	status, body := getBody(t, app, "/admin/media", &http.Cookie{Name: sessionCookieName, Value: token})
	if status != http.StatusOK {
		t.Fatalf("media status = %d, want %d", status, http.StatusOK)
	}
	assertContains(t, body, "Media Library")
	assertContains(t, body, "Upload image")

	req := mediaUploadRequest(t, "/admin/media", token, entry.csrfToken, "cover.png", testPNG(t))
	rr := httptest.NewRecorder()
	app.handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusSeeOther {
		t.Fatalf("upload status = %d, want %d; body: %s", rr.Code, http.StatusSeeOther, rr.Body.String())
	}
	if got := rr.Header().Get("Location"); got != "/admin/media" {
		t.Fatalf("upload Location = %q, want /admin/media", got)
	}

	assets, err := app.svc.Media.ListAssets()
	if err != nil {
		t.Fatalf("list assets: %v", err)
	}
	if len(assets) != 1 {
		t.Fatalf("assets len = %d, want 1", len(assets))
	}
	asset := assets[0]
	if asset.OriginalName != "cover.png" || !strings.HasPrefix(asset.URL, "/uploads/") || asset.MimeType != "image/png" {
		t.Fatalf("asset = %+v", asset)
	}

	status, body = getBody(t, app, "/admin/media", &http.Cookie{Name: sessionCookieName, Value: token})
	if status != http.StatusOK {
		t.Fatalf("media after upload status = %d, want %d", status, http.StatusOK)
	}
	assertContains(t, body, "cover.png")
	assertContains(t, body, asset.URL)
	assertContains(t, body, "![cover.png]("+asset.URL+")")

	status, body = getBody(t, app, asset.URL, nil)
	if status != http.StatusOK {
		t.Fatalf("upload serve status = %d, want %d", status, http.StatusOK)
	}
	if !strings.Contains(body, "PNG") {
		t.Fatalf("served upload body missing PNG signature marker")
	}

	form := url.Values{
		"_csrf":           {entry.csrfToken},
		"title":           {"Covered Dispatch"},
		"category_slug":   {"devops"},
		"content":         {"# Covered Dispatch"},
		"excerpt":         {"Covered excerpt"},
		"cover_image_url": {asset.URL},
		"intent":          {"publish"},
	}
	postReq := httptest.NewRequest(http.MethodPost, "/admin/posts", strings.NewReader(form.Encode()))
	postReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	postReq.AddCookie(&http.Cookie{Name: sessionCookieName, Value: token})
	rr = httptest.NewRecorder()
	app.handler.ServeHTTP(rr, postReq)
	if rr.Code != http.StatusSeeOther {
		t.Fatalf("create covered post status = %d, want %d; body: %s", rr.Code, http.StatusSeeOther, rr.Body.String())
	}
	post, err := app.svc.Blog.GetPost("covered-dispatch")
	if err != nil {
		t.Fatalf("get covered post: %v", err)
	}
	if post.CoverImageURL != asset.URL {
		t.Fatalf("cover url = %q, want %q", post.CoverImageURL, asset.URL)
	}
	status, body = getBody(t, app, "/blog/covered-dispatch", nil)
	if status != http.StatusOK {
		t.Fatalf("covered post status = %d, want %d", status, http.StatusOK)
	}
	assertContains(t, body, `src="`+asset.URL+`"`)
}

func TestAdminMediaUploadRejectsNonImages(t *testing.T) {
	app := newPublicTestApp(t)
	token := sessions.create(70, "owner", model.RoleOwner)
	entry, ok := sessions.get(token)
	if !ok {
		t.Fatalf("expected session")
	}

	req := mediaUploadRequest(t, "/admin/media", token, entry.csrfToken, "not-image.png", []byte("plain text"))
	rr := httptest.NewRecorder()
	app.handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("upload status = %d, want %d; body: %s", rr.Code, http.StatusBadRequest, rr.Body.String())
	}
	assertContains(t, rr.Body.String(), "unsupported image content")
}

func TestAdminPostUpdateIntentTransitionsPublishedState(t *testing.T) {
	app := newPublicTestApp(t)
	token := sessions.create(65, "owner", model.RoleOwner)
	entry, ok := sessions.get(token)
	if !ok {
		t.Fatalf("expected session")
	}

	post, err := app.svc.Blog.CreatePost(model.PostCreate{
		CategorySlug: "devops",
		Title:        "Workflow Draft",
		ContentMD:    "# Workflow Draft",
		Published:    false,
	})
	if err != nil {
		t.Fatalf("create draft: %v", err)
	}

	form := url.Values{
		"_csrf":         {entry.csrfToken},
		"title":         {"Workflow Draft"},
		"category_slug": {"devops"},
		"content":       {"# Workflow Draft Published"},
		"intent":        {"publish"},
	}
	req := httptest.NewRequest(http.MethodPost, "/admin/posts/update/"+strconv.FormatInt(post.ID, 10), strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: token})
	rr := httptest.NewRecorder()
	app.handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusSeeOther {
		t.Fatalf("publish status = %d, want %d; body: %s", rr.Code, http.StatusSeeOther, rr.Body.String())
	}
	updated, err := app.svc.Blog.GetPostByID(post.ID)
	if err != nil {
		t.Fatalf("get published post: %v", err)
	}
	if !updated.Published {
		t.Fatalf("post should be published after publish intent: %+v", updated)
	}

	form.Set("content", "# Workflow Draft Edited")
	form.Set("intent", "update")
	req = httptest.NewRequest(http.MethodPost, "/admin/posts/update/"+strconv.FormatInt(post.ID, 10), strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: token})
	rr = httptest.NewRecorder()
	app.handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusSeeOther {
		t.Fatalf("update status = %d, want %d; body: %s", rr.Code, http.StatusSeeOther, rr.Body.String())
	}
	updated, err = app.svc.Blog.GetPostByID(post.ID)
	if err != nil {
		t.Fatalf("get updated post: %v", err)
	}
	if !updated.Published {
		t.Fatalf("post should stay published after update intent: %+v", updated)
	}

	form.Set("intent", "draft")
	req = httptest.NewRequest(http.MethodPost, "/admin/posts/update/"+strconv.FormatInt(post.ID, 10), strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: token})
	rr = httptest.NewRecorder()
	app.handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusSeeOther {
		t.Fatalf("draft status = %d, want %d; body: %s", rr.Code, http.StatusSeeOther, rr.Body.String())
	}
	updated, err = app.svc.Blog.GetPostByID(post.ID)
	if err != nil {
		t.Fatalf("get draft post: %v", err)
	}
	if updated.Published {
		t.Fatalf("post should be draft after draft intent: %+v", updated)
	}
}

func TestAdminPostUpdateMissingPostReturnsNotFound(t *testing.T) {
	app := newPublicTestApp(t)
	token := sessions.create(55, "owner", model.RoleOwner)
	entry, ok := sessions.get(token)
	if !ok {
		t.Fatalf("expected session")
	}
	form := url.Values{
		"_csrf":         {entry.csrfToken},
		"title":         {"Missing"},
		"category_slug": {"devops"},
		"content":       {"# Missing"},
	}
	req := httptest.NewRequest(http.MethodPost, "/admin/posts/update/999", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: token})
	rr := httptest.NewRecorder()
	app.handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusNotFound, rr.Body.String())
	}
}

func TestAdminServerDiagnosticsDoNotMixAuditLogIDs(t *testing.T) {
	app := newPublicTestApp(t)
	ownerToken := sessions.create(53, "owner", model.RoleOwner)
	entry, ok := sessions.get(ownerToken)
	if !ok {
		t.Fatalf("expected session")
	}

	server, err := app.svc.Monitor.CreateServer("ten", "192.0.2.10", 9100, "admin")
	if err != nil {
		t.Fatalf("create server 10-ish: %v", err)
	}
	if err := app.svc.Admin.LogServerUpdated("owner", server.ID, "ten", "192.0.2.10", "9100", "198.51.100.10"); err != nil {
		t.Fatalf("log server update: %v", err)
	}
	if err := app.svc.Admin.LogServerUpdated("owner", app.serverID, "whyred", "127.0.0.1", "9100", "198.51.100.1"); err != nil {
		t.Fatalf("log server update 1: %v", err)
	}

	form := url.Values{
		"_csrf": {entry.csrfToken},
		"name":  {"node-10"},
		"host":  {"192.0.2.10"},
		"port":  {"9100"},
		"tags":  {"admin"},
	}
	req := httptest.NewRequest(http.MethodPost, "/admin/servers/update/"+strconv.FormatInt(server.ID, 10), strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: ownerToken})
	rr := httptest.NewRecorder()
	app.handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusSeeOther {
		t.Fatalf("update status = %d, want %d; body: %s", rr.Code, http.StatusSeeOther, rr.Body.String())
	}

	status, body := getBody(t, app, "/admin/servers/"+strconv.FormatInt(app.serverID, 10), &http.Cookie{Name: sessionCookieName, Value: ownerToken})
	if status != http.StatusOK {
		t.Fatalf("status = %d, want %d", status, http.StatusOK)
	}
	assertContains(t, body, "updated server id: 1")
	assertNotContains(t, body, "updated server id: 2")
	assertNotContains(t, body, "node-10")
}

func TestHermesPostPayloadRequiresContentAndPreservesTags(t *testing.T) {
	app := newPublicTestApp(t)

	missingContent := []byte(`{"category_slug":"devops","title":"No Body"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/posts", bytes.NewReader(missingContent))
	req.Header.Set("Authorization", "Bearer admin-key")
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	app.handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("missing content status = %d, want %d; body: %s", rr.Code, http.StatusBadRequest, rr.Body.String())
	}

	payload := []byte(`{"category_slug":"devops","title":"Hermes Dispatch","content_md":"# Hermes Dispatch","excerpt":"Automation note","published":true,"tags":["auto"," commissioned "]}`)
	req = httptest.NewRequest(http.MethodPost, "/api/v1/posts", bytes.NewReader(payload))
	req.Header.Set("Authorization", "Bearer admin-key")
	req.Header.Set("Content-Type", "application/json")
	rr = httptest.NewRecorder()
	app.handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("create status = %d, want %d; body: %s", rr.Code, http.StatusCreated, rr.Body.String())
	}
	var post model.Post
	if err := json.NewDecoder(rr.Body).Decode(&post); err != nil {
		t.Fatalf("decode post: %v", err)
	}
	if post.Excerpt != "Automation note" || len(post.Tags) != 2 {
		t.Fatalf("post excerpt/tags = %q/%+v", post.Excerpt, post.Tags)
	}
}

func TestHostStatsRequireAdminKey(t *testing.T) {
	app := newPublicTestApp(t)

	for _, path := range []string{"/api/v1/hoststats", "/api/v1/servers", "/api/v1/servers/1"} {
		t.Run(path, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, path, nil)
			rr := httptest.NewRecorder()

			app.handler.ServeHTTP(rr, req)

			if rr.Code != http.StatusUnauthorized {
				t.Fatalf("status = %d, want %d", rr.Code, http.StatusUnauthorized)
			}
			assertNotContains(t, rr.Body.String(), "127.0.0.1")
			assertNotContains(t, rr.Body.String(), "whyred")
			assertNotContains(t, rr.Body.String(), "load1")
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
	entry, ok := sessions.get(token)
	if !ok {
		t.Fatalf("expected session")
	}

	form := url.Values{
		"title":         {"Blocked Post"},
		"category_slug": {"devops"},
		"content":       {"# Blocked"},
		"_csrf":         {entry.csrfToken},
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

func TestAdminMutationsRequireCSRF(t *testing.T) {
	app := newPublicTestApp(t)
	token := sessions.create(45, "owner", model.RoleOwner)

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
	assertContains(t, rr.Body.String(), "invalid CSRF token")
}

func TestAdminMutationsAcceptValidCSRF(t *testing.T) {
	app := newPublicTestApp(t)
	token := sessions.create(46, "owner", model.RoleOwner)
	entry, ok := sessions.get(token)
	if !ok {
		t.Fatalf("expected session")
	}

	form := url.Values{
		"title":         {"Allowed Post"},
		"category_slug": {"devops"},
		"content":       {"# Allowed"},
		"_csrf":         {entry.csrfToken},
	}
	req := httptest.NewRequest(http.MethodPost, "/admin/posts", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: token})
	rr := httptest.NewRecorder()

	app.handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusSeeOther, rr.Body.String())
	}
}

func TestAdminPagesRenderCSRFTokens(t *testing.T) {
	app := newPublicTestApp(t)
	token := sessions.create(47, "owner", model.RoleOwner)

	for _, path := range []string{"/admin", "/admin/posts", "/admin/topics", "/admin/servers", "/admin/manage-users", "/admin/editorial-inbox"} {
		t.Run(path, func(t *testing.T) {
			status, body := getBody(t, app, path, &http.Cookie{Name: sessionCookieName, Value: token})
			if status != http.StatusOK {
				t.Fatalf("status = %d, want %d", status, http.StatusOK)
			}
			assertContains(t, body, `name="_csrf"`)
		})
	}
}

func TestAdminLogoutRequiresPOSTAndCSRF(t *testing.T) {
	app := newPublicTestApp(t)
	token := sessions.create(48, "owner", model.RoleOwner)

	getReq := httptest.NewRequest(http.MethodGet, "/admin/logout", nil)
	getReq.AddCookie(&http.Cookie{Name: sessionCookieName, Value: token})
	getRR := httptest.NewRecorder()
	app.handler.ServeHTTP(getRR, getReq)
	if getRR.Code != http.StatusMethodNotAllowed {
		t.Fatalf("GET status = %d, want %d", getRR.Code, http.StatusMethodNotAllowed)
	}
	if _, ok := sessions.get(token); !ok {
		t.Fatalf("GET logout deleted session")
	}

	postReq := httptest.NewRequest(http.MethodPost, "/admin/logout", strings.NewReader(url.Values{}.Encode()))
	postReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	postReq.AddCookie(&http.Cookie{Name: sessionCookieName, Value: token})
	postRR := httptest.NewRecorder()
	app.handler.ServeHTTP(postRR, postReq)
	if postRR.Code != http.StatusForbidden {
		t.Fatalf("POST without CSRF status = %d, want %d", postRR.Code, http.StatusForbidden)
	}
	if _, ok := sessions.get(token); !ok {
		t.Fatalf("CSRF-blocked logout deleted session")
	}

	entry, ok := sessions.get(token)
	if !ok {
		t.Fatalf("expected session")
	}
	form := url.Values{"_csrf": {entry.csrfToken}}
	validReq := httptest.NewRequest(http.MethodPost, "/admin/logout", strings.NewReader(form.Encode()))
	validReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	validReq.AddCookie(&http.Cookie{Name: sessionCookieName, Value: token})
	validRR := httptest.NewRecorder()
	app.handler.ServeHTTP(validRR, validReq)
	if validRR.Code != http.StatusSeeOther {
		t.Fatalf("valid logout status = %d, want %d", validRR.Code, http.StatusSeeOther)
	}
	if _, ok := sessions.get(token); ok {
		t.Fatalf("valid logout kept session")
	}
}

func TestAdminManagementPagesRequireOwnerOrAdmin(t *testing.T) {
	app := newPublicTestApp(t)
	token := sessions.create(44, "reader", model.RoleReadonly)

	for _, path := range []string{
		"/admin/editorial-inbox",
		"/admin/posts",
		"/admin/topics",
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

	for i, path := range []string{
		"/admin/posts/delete/1",
		"/admin/topics/delete/1",
		"/admin/servers/rotate-token/1",
		"/admin/servers/delete/1",
		"/admin/manage-users/delete/1",
	} {
		t.Run(path, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, path, nil)
			req.RemoteAddr = "198.51.100." + strconv.Itoa(i+1) + ":1234"
			req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: token})
			rr := httptest.NewRecorder()

			app.handler.ServeHTTP(rr, req)

			if rr.Code != http.StatusMethodNotAllowed {
				t.Fatalf("status = %d, want %d", rr.Code, http.StatusMethodNotAllowed)
			}
		})
	}
}

func TestAdminTopicRoutesRequireOwnerOrAdminAndExposeCRUDSurface(t *testing.T) {
	app := newPublicTestApp(t)
	readonlyToken := sessions.create(49, "reader", model.RoleReadonly)
	ownerToken := sessions.create(50, "owner", model.RoleOwner)
	ownerEntry, ok := sessions.get(ownerToken)
	if !ok {
		t.Fatalf("expected owner session")
	}

	for _, path := range []string{"/admin/topics", "/admin/topics/edit/1"} {
		t.Run("redirect guest "+path, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, path, nil)
			rr := httptest.NewRecorder()
			app.handler.ServeHTTP(rr, req)
			if rr.Code != http.StatusSeeOther {
				t.Fatalf("guest status = %d, want %d", rr.Code, http.StatusSeeOther)
			}
			if got := rr.Header().Get("Location"); got != "/admin/login" {
				t.Fatalf("guest location = %q, want /admin/login", got)
			}
		})

		t.Run("forbid readonly "+path, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, path, nil)
			req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: readonlyToken})
			rr := httptest.NewRecorder()
			app.handler.ServeHTTP(rr, req)
			if rr.Code != http.StatusForbidden {
				t.Fatalf("readonly status = %d, want %d", rr.Code, http.StatusForbidden)
			}
		})

		t.Run("owner can reach "+path, func(t *testing.T) {
			status, body := getBody(t, app, path, &http.Cookie{Name: sessionCookieName, Value: ownerToken})
			if status != http.StatusOK {
				t.Fatalf("owner status = %d, want %d; body: %s", status, http.StatusOK, body)
			}
			assertContains(t, body, `name="_csrf"`)
		})
	}

	for _, tc := range []struct {
		name string
		path string
	}{
		{name: "create topic", path: "/admin/topics"},
		{name: "update topic", path: "/admin/topics/update/1"},
		{name: "delete topic", path: "/admin/topics/delete/1"},
	} {
		t.Run(tc.name+" requires csrf", func(t *testing.T) {
			form := url.Values{"name": {"Topic"}, "slug": {"topic-slug"}}
			req := httptest.NewRequest(http.MethodPost, tc.path, strings.NewReader(form.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: ownerToken})
			rr := httptest.NewRecorder()
			app.handler.ServeHTTP(rr, req)
			if rr.Code != http.StatusForbidden {
				t.Fatalf("status = %d, want %d", rr.Code, http.StatusForbidden)
			}
		})

		t.Run(tc.name+" accepts valid csrf", func(t *testing.T) {
			form := url.Values{
				"name":  {"Topic"},
				"slug":  {"topic-slug"},
				"_csrf": {ownerEntry.csrfToken},
			}
			req := httptest.NewRequest(http.MethodPost, tc.path, strings.NewReader(form.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: ownerToken})
			rr := httptest.NewRecorder()
			app.handler.ServeHTTP(rr, req)
			if rr.Code == http.StatusForbidden {
				t.Fatalf("valid csrf should pass middleware for %s; body: %s", tc.path, rr.Body.String())
			}
			if rr.Code == http.StatusNotFound {
				t.Fatalf("missing admin topic route: %s", tc.path)
			}
		})
	}
}

func TestPublicDashboardPlatformSystemHealthFallsBackToNA(t *testing.T) {
	app := newPublicTestApp(t)

	status, body := getBody(t, app, "/dashboard", nil)
	if status != http.StatusOK {
		t.Fatalf("status = %d, want %d", status, http.StatusOK)
	}
	assertContains(t, body, "Platform System Health")
	assertContains(t, body, "N/A")
}

func TestAPIRequestBodiesAreCapped(t *testing.T) {
	app := newPublicTestApp(t)
	largeValue := strings.Repeat("x", 2<<20)

	tests := []struct {
		name    string
		path    string
		payload string
	}{
		{name: "create post", path: "/api/v1/posts", payload: `{"category_slug":"devops","title":"large","content_md":"` + largeValue + `"}`},
		{name: "create project", path: "/api/v1/projects", payload: `{"title":"large","content_md":"` + largeValue + `"}`},
		{name: "update project", path: "/api/v1/projects/1", payload: `{"title":"large","content_md":"` + largeValue + `"}`},
		{name: "create server", path: "/api/v1/servers", payload: `{"name":"large","host":"127.0.0.1","padding":"` + largeValue + `"}`},
		{name: "record metric", path: "/api/v1/servers/1/ping", payload: `{"online":true,"padding":"` + largeValue + `"}`},
	}

	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			method := http.MethodPost
			if tt.name == "update project" {
				method = http.MethodPut
			}
			req := httptest.NewRequest(method, tt.path, strings.NewReader(tt.payload))
			req.RemoteAddr = fmt.Sprintf("198.51.100.%d:1234", i+10)
			req.Header.Set("Authorization", "Bearer admin-key")
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()

			app.handler.ServeHTTP(rr, req)

			if rr.Code != http.StatusRequestEntityTooLarge {
				t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusRequestEntityTooLarge, rr.Body.String())
			}
		})
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/editorial-inbox", strings.NewReader(`{"source_type":"repo","source_value":"`+largeValue+`","priority":50,"not_before":"2026-06-05T08:00:00Z","mode":"scheduled_assignment","status":"queued"}`))
	req.RemoteAddr = "198.51.100.99:1234"
	req.Header.Set("Authorization", "Bearer admin-key")
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	app.handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("editorial api body cap status = %d, want %d; body: %s", rr.Code, http.StatusRequestEntityTooLarge, rr.Body.String())
	}
}

func TestEditorialInboxAdminAndAPIFlow(t *testing.T) {
	app := newPublicTestApp(t)
	token := sessions.create(88, "owner", model.RoleOwner)
	entry, ok := sessions.get(token)
	if !ok {
		t.Fatalf("expected session")
	}
	form := url.Values{
		"source_value":  {"https://github.com/example/repo"},
		"category_hint": {"devops"},
		"note":          {"queue this repo first"},
		"_csrf":         {entry.csrfToken},
	}
	req := httptest.NewRequest(http.MethodPost, "/admin/editorial-inbox", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: token})
	rr := httptest.NewRecorder()
	app.handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusSeeOther {
		t.Fatalf("create editorial status = %d, want %d; body: %s", rr.Code, http.StatusSeeOther, rr.Body.String())
	}
	status, body := getBody(t, app, "/admin/editorial-inbox", &http.Cookie{Name: sessionCookieName, Value: token})
	if status != http.StatusOK {
		t.Fatalf("editorial page status = %d, want %d", status, http.StatusOK)
	}
	assertContains(t, body, "Editorial Inbox")
	assertContains(t, body, "https://github.com/example/repo")

	apiReq := httptest.NewRequest(http.MethodGet, "/api/v1/editorial-inbox/contract", nil)
	apiReq.Header.Set("Authorization", "Bearer admin-key")
	apiRR := httptest.NewRecorder()
	app.handler.ServeHTTP(apiRR, apiReq)
	if apiRR.Code != http.StatusOK {
		t.Fatalf("editorial contract status = %d, want %d; body: %s", apiRR.Code, http.StatusOK, apiRR.Body.String())
	}
	assertContains(t, apiRR.Body.String(), `"source_of_truth":"raevtar"`)
	assertContains(t, apiRR.Body.String(), `"published_post_id"`)
	assertNotContains(t, body, "/api/v1/editorial-inbox/contract")

	listReq := httptest.NewRequest(http.MethodGet, "/api/v1/editorial-inbox", nil)
	listReq.Header.Set("Authorization", "Bearer admin-key")
	listRR := httptest.NewRecorder()
	app.handler.ServeHTTP(listRR, listReq)
	if listRR.Code != http.StatusOK {
		t.Fatalf("editorial list status = %d, want %d; body: %s", listRR.Code, http.StatusOK, listRR.Body.String())
	}
	assertContains(t, listRR.Body.String(), `"status":"approved"`)
}

func TestEditorialInboxClaimCompleteAndFailAPIFlow(t *testing.T) {
	app := newPublicTestApp(t)
	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/editorial-inbox", strings.NewReader(`{"source_type":"repo","source_value":"https://github.com/example/claim","priority":90,"not_before":"2025-06-05T08:00:00Z","mode":"scheduled_assignment","status":"approved"}`))
	createReq.RemoteAddr = "198.51.100.90:1234"
	createReq.Header.Set("Authorization", "Bearer admin-key")
	createReq.Header.Set("Content-Type", "application/json")
	createRR := httptest.NewRecorder()
	app.handler.ServeHTTP(createRR, createReq)
	if createRR.Code != http.StatusCreated {
		t.Fatalf("create status = %d, want %d; body: %s", createRR.Code, http.StatusCreated, createRR.Body.String())
	}
	var created model.EditorialInboxItem
	if err := json.Unmarshal(createRR.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode created item: %v", err)
	}
	claimReq := httptest.NewRequest(http.MethodPost, "/api/v1/editorial-inbox/claim", strings.NewReader(`{"worker":"hermes-cron"}`))
	claimReq.RemoteAddr = "198.51.100.91:1234"
	claimReq.Header.Set("Authorization", "Bearer admin-key")
	claimReq.Header.Set("Content-Type", "application/json")
	claimRR := httptest.NewRecorder()
	app.handler.ServeHTTP(claimRR, claimReq)
	if claimRR.Code != http.StatusOK {
		t.Fatalf("claim status = %d, want %d; body: %s", claimRR.Code, http.StatusOK, claimRR.Body.String())
	}
	var claim model.EditorialInboxClaimResult
	if err := json.Unmarshal(claimRR.Body.Bytes(), &claim); err != nil {
		t.Fatalf("decode claim result: %v", err)
	}
	if claim.Item == nil || claim.Item.ID != created.ID {
		t.Fatalf("claimed item = %+v, want id %d", claim.Item, created.ID)
	}
	if claim.ClaimToken == "" {
		t.Fatalf("expected claim token")
	}
	secondClaimReq := httptest.NewRequest(http.MethodPost, "/api/v1/editorial-inbox/claim", strings.NewReader(`{"worker":"hermes-cron"}`))
	secondClaimReq.RemoteAddr = "198.51.100.92:1234"
	secondClaimReq.Header.Set("Authorization", "Bearer admin-key")
	secondClaimReq.Header.Set("Content-Type", "application/json")
	secondClaimRR := httptest.NewRecorder()
	app.handler.ServeHTTP(secondClaimRR, secondClaimReq)
	if secondClaimRR.Code != http.StatusNoContent {
		t.Fatalf("second claim status = %d, want %d; body: %s", secondClaimRR.Code, http.StatusNoContent, secondClaimRR.Body.String())
	}
	badCompleteReq := httptest.NewRequest(http.MethodPost, "/api/v1/editorial-inbox/"+strconv.FormatInt(created.ID, 10)+"/complete", strings.NewReader(`{"claim_token":"wrong","published_post_id":11}`))
	badCompleteReq.RemoteAddr = "198.51.100.93:1234"
	badCompleteReq.Header.Set("Authorization", "Bearer admin-key")
	badCompleteReq.Header.Set("Content-Type", "application/json")
	badCompleteRR := httptest.NewRecorder()
	app.handler.ServeHTTP(badCompleteRR, badCompleteReq)
	if badCompleteRR.Code != http.StatusConflict {
		t.Fatalf("bad complete status = %d, want %d; body: %s", badCompleteRR.Code, http.StatusConflict, badCompleteRR.Body.String())
	}
	goodFailReq := httptest.NewRequest(http.MethodPost, "/api/v1/editorial-inbox/"+strconv.FormatInt(created.ID, 10)+"/fail", strings.NewReader(`{"claim_token":"`+claim.ClaimToken+`","failure_note":"publish timeout","failure_meta":"{\"status\":504}","retryable":true}`))
	goodFailReq.RemoteAddr = "198.51.100.94:1234"
	goodFailReq.Header.Set("Authorization", "Bearer admin-key")
	goodFailReq.Header.Set("Content-Type", "application/json")
	goodFailRR := httptest.NewRecorder()
	app.handler.ServeHTTP(goodFailRR, goodFailReq)
	if goodFailRR.Code != http.StatusOK {
		t.Fatalf("fail status = %d, want %d; body: %s", goodFailRR.Code, http.StatusOK, goodFailRR.Body.String())
	}
	assertContains(t, goodFailRR.Body.String(), `"status":"approved"`)
	assertContains(t, claimRR.Body.String(), `"claim_token":"`)
	assertContains(t, claimRR.Body.String(), `"attempt_count":1`)
	contractReq := httptest.NewRequest(http.MethodGet, "/api/v1/editorial-inbox/contract", nil)
	contractReq.RemoteAddr = "198.51.100.95:1234"
	contractReq.Header.Set("Authorization", "Bearer admin-key")
	contractRR := httptest.NewRecorder()
	app.handler.ServeHTTP(contractRR, contractReq)
	if contractRR.Code != http.StatusOK {
		t.Fatalf("contract status = %d, want %d; body: %s", contractRR.Code, http.StatusOK, contractRR.Body.String())
	}
	assertContains(t, contractRR.Body.String(), `"claim_endpoint":"/api/v1/editorial-inbox/claim"`)
	assertContains(t, contractRR.Body.String(), `"lease_ttl":"30m"`)
	assertContains(t, contractRR.Body.String(), `"single_running":"new claim blocked while any running item still has active lease"`)
}

func TestEditorialInboxPhase4SummaryAndAdminRendering(t *testing.T) {
	app := newPublicTestApp(t)
	token := sessions.create(89, "owner", model.RoleOwner)
	entry, ok := sessions.get(token)
	if !ok {
		t.Fatalf("expected session")
	}
	form := url.Values{
		"source_type":   {"repo"},
		"source_value":  {"https://github.com/example/overdue-phase4"},
		"category_hint": {"devops"},
		"priority":      {"40"},
		"not_before":    {"2025-06-05T08:00"},
		"deadline":      {"2025-06-05T09:00"},
		"mode":          {model.EditorialModeScheduled},
		"status":        {model.EditorialStatusApproved},
		"note":          {"phase 4 overdue test"},
		"_csrf":         {entry.csrfToken},
	}
	req := httptest.NewRequest(http.MethodPost, "/admin/editorial-inbox", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: token})
	rr := httptest.NewRecorder()
	app.handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusSeeOther {
		t.Fatalf("create overdue item status = %d, want %d; body: %s", rr.Code, http.StatusSeeOther, rr.Body.String())
	}
	summaryReq := httptest.NewRequest(http.MethodGet, "/api/v1/editorial-inbox/summary", nil)
	summaryReq.Header.Set("Authorization", "Bearer admin-key")
	summaryRR := httptest.NewRecorder()
	app.handler.ServeHTTP(summaryRR, summaryReq)
	if summaryRR.Code != http.StatusOK {
		t.Fatalf("summary status = %d, want %d; body: %s", summaryRR.Code, http.StatusOK, summaryRR.Body.String())
	}
	assertContains(t, summaryRR.Body.String(), `"fairness"`)
	assertContains(t, summaryRR.Body.String(), `"overdue"`)
	assertContains(t, summaryRR.Body.String(), `"analytics"`)
	status, body := getBody(t, app, "/admin/editorial-inbox", &http.Cookie{Name: sessionCookieName, Value: token})
	if status != http.StatusOK {
		t.Fatalf("editorial page status = %d, want %d", status, http.StatusOK)
	}
	assertContains(t, body, "Overdue")
	assertContains(t, body, "Fairness")
	assertContains(t, body, "Publish analytics")
}

func TestAdminEditorialInboxDeleteAndLockedRendering(t *testing.T) {
	app := newPublicTestApp(t)
	token := sessions.create(90, "owner", model.RoleOwner)
	entry, ok := sessions.get(token)
	if !ok {
		t.Fatalf("expected session")
	}
	createForm := url.Values{
		"source_value":  {"https://github.com/example/delete-me"},
		"category_hint": {"devops"},
		"note":          {"delete before first run"},
		"_csrf":         {entry.csrfToken},
	}
	createReq := httptest.NewRequest(http.MethodPost, "/admin/editorial-inbox", strings.NewReader(createForm.Encode()))
	createReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	createReq.AddCookie(&http.Cookie{Name: sessionCookieName, Value: token})
	createRR := httptest.NewRecorder()
	app.handler.ServeHTTP(createRR, createReq)
	if createRR.Code != http.StatusSeeOther {
		t.Fatalf("create mutable status = %d, want %d; body: %s", createRR.Code, http.StatusSeeOther, createRR.Body.String())
	}
	status, body := getBody(t, app, "/admin/editorial-inbox", &http.Cookie{Name: sessionCookieName, Value: token})
	if status != http.StatusOK {
		t.Fatalf("editorial page status = %d, want %d", status, http.StatusOK)
	}
	assertContains(t, body, "/admin/editorial-inbox/delete/")
	items, err := app.svc.Editorial.ListInboxItems(service.EditorialInboxListFilter{})
	if err != nil {
		t.Fatalf("list items: %v", err)
	}
	var mutableID int64
	for _, item := range items {
		if item.SourceValue == "https://github.com/example/delete-me" {
			mutableID = item.ID
			break
		}
	}
	if mutableID == 0 {
		t.Fatalf("expected mutable item id")
	}
	deleteForm := url.Values{"_csrf": {entry.csrfToken}}
	deleteReq := httptest.NewRequest(http.MethodPost, "/admin/editorial-inbox/delete/"+strconv.FormatInt(mutableID, 10), strings.NewReader(deleteForm.Encode()))
	deleteReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	deleteReq.AddCookie(&http.Cookie{Name: sessionCookieName, Value: token})
	deleteRR := httptest.NewRecorder()
	app.handler.ServeHTTP(deleteRR, deleteReq)
	if deleteRR.Code != http.StatusSeeOther {
		t.Fatalf("delete status = %d, want %d; body: %s", deleteRR.Code, http.StatusSeeOther, deleteRR.Body.String())
	}
	lockedCreateReq := httptest.NewRequest(http.MethodPost, "/api/v1/editorial-inbox", strings.NewReader(`{"source_type":"repo","source_value":"https://github.com/example/locked-ui","priority":80,"not_before":"2025-06-05T08:00:00Z","mode":"scheduled_assignment","status":"approved"}`))
	lockedCreateReq.Header.Set("Authorization", "Bearer admin-key")
	lockedCreateReq.Header.Set("Content-Type", "application/json")
	lockedCreateRR := httptest.NewRecorder()
	app.handler.ServeHTTP(lockedCreateRR, lockedCreateReq)
	if lockedCreateRR.Code != http.StatusCreated {
		t.Fatalf("create locked status = %d, want %d; body: %s", lockedCreateRR.Code, http.StatusCreated, lockedCreateRR.Body.String())
	}
	claimReq := httptest.NewRequest(http.MethodPost, "/api/v1/editorial-inbox/claim", strings.NewReader(`{"worker":"hermes-cron"}`))
	claimReq.Header.Set("Authorization", "Bearer admin-key")
	claimReq.Header.Set("Content-Type", "application/json")
	claimRR := httptest.NewRecorder()
	app.handler.ServeHTTP(claimRR, claimReq)
	if claimRR.Code != http.StatusOK {
		t.Fatalf("claim status = %d, want %d; body: %s", claimRR.Code, http.StatusOK, claimRR.Body.String())
	}
	status, body = getBody(t, app, "/admin/editorial-inbox", &http.Cookie{Name: sessionCookieName, Value: token})
	if status != http.StatusOK {
		t.Fatalf("editorial page status = %d, want %d", status, http.StatusOK)
	}
	assertContains(t, body, "Locked after first execution attempt")
}

func TestAdminLoginRequestBodyIsCapped(t *testing.T) {
	app := newPublicTestApp(t)
	form := url.Values{"username": {strings.Repeat("x", 2<<20)}, "password": {"wrong-password"}}
	req := httptest.NewRequest(http.MethodPost, "/admin/login", strings.NewReader(form.Encode()))
	req.RemoteAddr = "198.51.100.30:1234"
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()

	app.handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusRequestEntityTooLarge, rr.Body.String())
	}
}

func TestAdminLoginThrottlesRepeatedFailuresPerIP(t *testing.T) {
	app := newPublicTestApp(t)
	form := url.Values{"username": {"admin"}, "password": {"wrong-password"}}

	for attempt := 1; attempt <= 6; attempt++ {
		req := httptest.NewRequest(http.MethodPost, "/admin/login", strings.NewReader(form.Encode()))
		req.RemoteAddr = "198.51.100.40:1234"
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()

		app.handler.ServeHTTP(rr, req)

		if attempt <= 5 && rr.Code != http.StatusUnauthorized {
			t.Fatalf("attempt %d status = %d, want %d; body: %s", attempt, rr.Code, http.StatusUnauthorized, rr.Body.String())
		}
		if attempt == 6 {
			if rr.Code != http.StatusTooManyRequests {
				t.Fatalf("attempt %d status = %d, want %d; body: %s", attempt, rr.Code, http.StatusTooManyRequests, rr.Body.String())
			}
			if rr.Header().Get("Retry-After") == "" {
				t.Fatalf("attempt %d missing Retry-After header", attempt)
			}
		}
	}
}

func TestAdminLoginThrottlesRotatingUsernamesPerIP(t *testing.T) {
	app := newPublicTestApp(t)

	for attempt := 1; attempt <= loginIPFailureLimit+1; attempt++ {
		if attempt == loginIPFailureLimit/2+1 {
			form := url.Values{"username": {"admin"}, "password": {"demo-pass-123"}}
			req := httptest.NewRequest(http.MethodPost, "/admin/login", strings.NewReader(form.Encode()))
			req.RemoteAddr = "198.51.100.41:1234"
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			rr := httptest.NewRecorder()

			app.handler.ServeHTTP(rr, req)

			if rr.Code != http.StatusSeeOther {
				t.Fatalf("successful login status = %d, want %d; body: %s", rr.Code, http.StatusSeeOther, rr.Body.String())
			}
		}

		form := url.Values{"username": {fmt.Sprintf("spray-%d", attempt)}, "password": {"wrong-password"}}
		req := httptest.NewRequest(http.MethodPost, "/admin/login", strings.NewReader(form.Encode()))
		req.RemoteAddr = "198.51.100.41:1234"
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()

		app.handler.ServeHTTP(rr, req)

		if attempt <= loginIPFailureLimit && rr.Code != http.StatusUnauthorized {
			t.Fatalf("attempt %d status = %d, want %d; body: %s", attempt, rr.Code, http.StatusUnauthorized, rr.Body.String())
		}
		if attempt == loginIPFailureLimit+1 {
			if rr.Code != http.StatusTooManyRequests {
				t.Fatalf("attempt %d status = %d, want %d; body: %s", attempt, rr.Code, http.StatusTooManyRequests, rr.Body.String())
			}
			if rr.Header().Get("Retry-After") == "" {
				t.Fatalf("attempt %d missing Retry-After header", attempt)
			}
		}
	}
}

func TestAdminDeleteMissingResourcesDoNotSilentlyRedirect(t *testing.T) {
	app := newPublicTestApp(t)
	token := sessions.create(71, "owner", model.RoleOwner)
	entry, ok := sessions.get(token)
	if !ok {
		t.Fatalf("expected session")
	}

	for i, path := range []string{"/admin/posts/delete/999", "/admin/servers/delete/999"} {
		t.Run(path, func(t *testing.T) {
			form := url.Values{"_csrf": {entry.csrfToken}}
			req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(form.Encode()))
			req.RemoteAddr = fmt.Sprintf("198.51.100.%d:1234", i+50)
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: token})
			rr := httptest.NewRecorder()

			app.handler.ServeHTTP(rr, req)

			if rr.Code != http.StatusNotFound {
				t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusNotFound, rr.Body.String())
			}
		})
	}
}

func TestInternalErrorsReturnGenericResponses(t *testing.T) {
	t.Run("HTML", func(t *testing.T) {
		app := newPublicTestApp(t)
		if err := app.db.Close(); err != nil {
			t.Fatalf("close database: %v", err)
		}
		status, body := getBody(t, app, "/", nil)
		if status != http.StatusInternalServerError {
			t.Fatalf("status = %d, want %d", status, http.StatusInternalServerError)
		}
		assertContains(t, body, "internal server error")
		assertNotContains(t, body, "sql: database is closed")
	})

	t.Run("JSON", func(t *testing.T) {
		app := newPublicTestApp(t)
		if err := app.db.Close(); err != nil {
			t.Fatalf("close database: %v", err)
		}
		req := httptest.NewRequest(http.MethodGet, "/api/v1/categories", nil)
		req.RemoteAddr = "198.51.100.61:1234"
		rr := httptest.NewRecorder()
		app.handler.ServeHTTP(rr, req)
		if rr.Code != http.StatusInternalServerError {
			t.Fatalf("status = %d, want %d", rr.Code, http.StatusInternalServerError)
		}
		if got := strings.TrimSpace(rr.Body.String()); got != `{"error":"internal server error"}` {
			t.Fatalf("body = %q, want generic JSON error", got)
		}
	})
}

func TestBaseLayoutUsesSelfHostedHTMX(t *testing.T) {
	content, err := os.ReadFile(filepath.Join("..", "..", "internal", "view", "layouts", "base.templ"))
	if err != nil {
		t.Fatalf("read base layout: %v", err)
	}
	body := string(content)
	assertContains(t, body, `<script src="/static/js/htmx.min.js"></script>`)
	assertNotContains(t, body, "unpkg.com")
	if _, err := os.Stat(filepath.Join("..", "..", "static", "js", "htmx.min.js")); err != nil {
		t.Fatalf("self-hosted HTMX asset missing: %v", err)
	}
}

func TestSecurityHeadersAllowOnlySelfHostedScripts(t *testing.T) {
	handler := WithSecurityHeaders(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	csp := rr.Header().Get("Content-Security-Policy")
	assertContains(t, csp, "script-src 'self'")
	assertNotContains(t, csp, "unpkg.com")
	assertNotContains(t, csp, "script-src 'self' 'unsafe-inline'")
}

func TestAgentMetricPayloadAcceptsExtendedNodeHealthTelemetry(t *testing.T) {
	app := newPublicTestApp(t)
	token, err := app.svc.Monitor.RotateAgentToken(app.serverID)
	if err != nil {
		t.Fatalf("rotate token: %v", err)
	}
	payload := []byte(`{"cpu_percent":14.5,"cpu_load_1":0.4,"cpu_load_5":0.3,"cpu_load_15":0.2,"cpu_cores":4,"ram_used_mb":512,"ram_total_mb":2048,"disk_used_gb":12,"disk_total_gb":32,"temperature_c":48.5,"temperature_available":true,"uptime_seconds":90,"online":true}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/servers/1/ping", bytes.NewReader(payload))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	app.handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}
	metrics, err := app.svc.Monitor.GetRecentMetrics(app.serverID, 1)
	if err != nil {
		t.Fatalf("metrics: %v", err)
	}
	if len(metrics) != 1 {
		t.Fatalf("metrics len = %d, want 1", len(metrics))
	}
	for name, want := range map[string]any{"CPULoad1": 0.4, "CPULoad5": 0.3, "CPULoad15": 0.2, "CPUCores": int64(4), "DiskTotalGB": 32.0, "TemperatureC": 48.5, "TemperatureAvailable": true} {
		field := reflect.ValueOf(metrics[0]).FieldByName(name)
		if !field.IsValid() {
			t.Fatalf("ServerMetric missing authorized field %s", name)
		}
		if got := field.Interface(); !reflect.DeepEqual(got, reflect.ValueOf(want).Convert(field.Type()).Interface()) {
			t.Fatalf("ServerMetric.%s = %v, want %v", name, got, want)
		}
	}
}
