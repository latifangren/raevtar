package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"raevtar/internal/config"
	"raevtar/internal/model"
	"raevtar/internal/repo"
	"raevtar/internal/service"
)

var testRequestCounter int

type publicTestApp struct {
	handler  http.Handler
	svc      *service.Service
	serverID int64
}

func newPublicTestApp(t *testing.T) *publicTestApp {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "raevtar_test.db")
	cfg := &config.Config{
		DatabasePath: dbPath,
		Domain:       "raevtar.test",
		AdminKey:     "admin-key",
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

	server, err := svc.Monitor.CreateServer("whyred", "127.0.0.1", 9100, "local")
	if err != nil {
		t.Fatalf("create server: %v", err)
	}

	if post.Slug == "" {
		t.Fatalf("expected generated slug")
	}

	return &publicTestApp{handler: New(svc, cfg), svc: svc, serverID: server.ID}
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
				"Platform Showcase",
				"Lab",
				`href="/lab"`,
				`id="lab"`,
				"Hello Raevtar",
				`href="/blog"`,
				`href="/dashboard"`,
				`href="/docs"`,
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
				"GET /api/v1/categories",
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
				"Read dispatch",
				`href="/blog?category=devops"`,
				`href="/blog/hello-raevtar"`,
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
			},
		},
		{
			name:           "dashboard index",
			path:           "/dashboard",
			wantStatus:     http.StatusOK,
			wantContentTyp: "text/html",
			wantContains: []string{
				"Server Dashboard",
				"Monitor Grid",
				"Online",
				"Stale",
				"Offline",
				"last agent signal",
				"grid refresh every 30s",
				"whyred",
				"Hidden on public view",
				"redacted",
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
				assertContains(t, body, want)
			}
		})
	}
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

func TestOwnerDashboardShowsServerTopology(t *testing.T) {
	app := newPublicTestApp(t)
	cookie := &http.Cookie{Name: sessionCookieName, Value: sessions.create(101, "owner", model.RoleOwner)}

	status, body := getBody(t, app, "/dashboard", cookie)
	if status != http.StatusOK {
		t.Fatalf("status = %d, want %d", status, http.StatusOK)
	}

	assertContains(t, body, "127.0.0.1:9100")
	assertContains(t, body, ">9100<")
	assertContains(t, body, ">local<")
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
		{name: "owner", role: model.RoleOwner, wantRegister: true},
		{name: "admin", role: model.RoleAdmin, wantRegister: true},
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
	assertContains(t, body, "Login as owner/admin for network details.")
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
			assertContains(t, body, "Login as owner/admin for network details.")

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

func TestOwnerServerDetailShowsServerTopology(t *testing.T) {
	app := newPublicTestApp(t)
	cookie := &http.Cookie{Name: sessionCookieName, Value: sessions.create(102, "owner", model.RoleOwner)}

	status, body := getBody(t, app, "/dashboard/1", cookie)
	if status != http.StatusOK {
		t.Fatalf("status = %d, want %d", status, http.StatusOK)
	}

	assertContains(t, body, "127.0.0.1:9100")
	assertContains(t, body, "port 9100")
	assertContains(t, body, ">local<")
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
	assertContains(t, body, "metric detail")
	assertContains(t, body, "Metric detail hidden on public view.")
	assertContains(t, body, "Panel refreshed")
	assertContains(t, body, "Telemetry is fresh. Latest agent signal arrived recently.")
	assertContains(t, body, "Endpoint hidden on public view.")
	assertContains(t, body, "Login as owner/admin for network details.")
	for _, leak := range []string{"CPU", "RAM", "Disk", "Uptime", "12.5%", "256.0 / 1024.0 MB", "8.5 GB", "sample count", "history window", "availability"} {
		assertNotContains(t, body, leak)
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
			assertContains(t, body, "Metric detail hidden on public view.")

			for _, leak := range []string{"CPU", "RAM", "Disk", "Uptime", "sample count", "history window", "availability"} {
				assertNotContains(t, body, leak)
			}
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

func TestOwnerAndAdminServerDetailLiveShowTopologyAndManageLink(t *testing.T) {
	app := newPublicTestApp(t)

	for _, role := range []string{model.RoleOwner, model.RoleAdmin} {
		t.Run(role, func(t *testing.T) {
			cookie := &http.Cookie{Name: sessionCookieName, Value: sessions.create(102, role, role)}
			status, body := getBody(t, app, "/dashboard/1/live", cookie)
			if status != http.StatusOK {
				t.Fatalf("status = %d, want %d", status, http.StatusOK)
			}

			assertContains(t, body, "127.0.0.1:9100")
			assertContains(t, body, "port 9100")
			assertContains(t, body, ">local<")
			assertContains(t, body, `href="/admin/servers"`)
			assertContains(t, body, "Manage in admin")
			assertContains(t, body, "latest metric")
			assertContains(t, body, "sample count")
			assertContains(t, body, "history window")
			assertContains(t, body, "availability")
		})
	}
}

func TestPublicServerDetailRedactsMetricValues(t *testing.T) {
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
	assertContains(t, body, "Metric detail hidden on public view.")
	for _, leak := range []string{"CPU", "RAM", "Disk", "Uptime", "12.5%", "256.0 / 1024.0 MB", "8.5 GB", "sample count", "history window", "availability"} {
		assertNotContains(t, body, leak)
	}
}

func TestOwnerServerDetailRendersMetricMarkers(t *testing.T) {
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
	for _, want := range []string{"CPU", "RAM", "Disk", "Uptime", "12.5%", "256.0 / 1024.0 MB", "8.5 GB"} {
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
	assertContains(t, body, "Copy this token now")
	assertContains(t, body, "raevtar-agent.sh")
	assertContains(t, body, "RAEVTAR_URL=https://raevtar.test")
	assertContains(t, body, "RAEVTAR_SERVER_ID=2")
	assertContains(t, body, "RAEVTAR_AGENT_TOKEN=")

	_, secondBody := getBody(t, app, "/admin/servers", &http.Cookie{Name: sessionCookieName, Value: token})
	assertNotContains(t, secondBody, "Copy this token now")
}

func TestAdminServerDiagnosticsAreOwnerOnlyAndShowPrivateDetails(t *testing.T) {
	app := newPublicTestApp(t)
	ownerToken := sessions.create(51, "owner", model.RoleOwner)
	readonlyToken := sessions.create(52, "reader", model.RoleReadonly)

	status, body := getBody(t, app, "/admin/servers/1", &http.Cookie{Name: sessionCookieName, Value: ownerToken})
	if status != http.StatusOK {
		t.Fatalf("owner status = %d, want %d", status, http.StatusOK)
	}
	for _, want := range []string{"whyred diagnostics", "127.0.0.1", "9100", "local", "Agent setup", "raevtar-agent.sh", "Admin activity"} {
		assertContains(t, body, want)
	}
	assertNotContains(t, body, "agent_token_hash")

	status, body = getBody(t, app, "/admin/servers/1", &http.Cookie{Name: sessionCookieName, Value: readonlyToken})
	if status != http.StatusForbidden {
		t.Fatalf("readonly status = %d, want %d", status, http.StatusForbidden)
	}
	assertNotContains(t, body, "127.0.0.1")
	assertNotContains(t, body, "raevtar-agent.sh")
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

	for _, path := range []string{"/admin", "/admin/posts", "/admin/servers", "/admin/manage-users", "/dashboard"} {
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

	for i, path := range []string{
		"/admin/posts/delete/1",
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
