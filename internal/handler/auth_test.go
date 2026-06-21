package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"raevtar/internal/model"
)

// loginSession logs in as the seed admin user and returns the session cookie
// and CSRF token from the in-memory session store.
func loginSession(t *testing.T, app *publicTestApp) (*http.Cookie, string) {
	t.Helper()
	form := url.Values{"username": {"admin"}, "password": {"demo-pass-123"}}
	req := httptest.NewRequest(http.MethodPost, "/admin/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	testRequestCounter++
	req.RemoteAddr = fmt.Sprintf("203.0.113.%d:1234", testRequestCounter%250+1)
	rr := httptest.NewRecorder()
	app.handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Fatalf("login status = %d, want %d; body: %s", rr.Code, http.StatusSeeOther, rr.Body.String())
	}

	cookies := rr.Result().Cookies()
	var sessionCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == sessionCookieName {
			sessionCookie = c
			break
		}
	}
	if sessionCookie == nil {
		t.Fatalf("no session cookie set after login")
	}

	entry, ok := sessions.get(sessionCookie.Value)
	if !ok {
		t.Fatalf("session not found in store after login")
	}
	return sessionCookie, entry.csrfToken
}

func TestAdminLoginPageUnauthenticated(t *testing.T) {
	app := newPublicTestApp(t)

	req := httptest.NewRequest(http.MethodGet, "/admin/login", nil)
	rr := httptest.NewRecorder()
	app.handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}
	body := rr.Body.String()
	if !strings.Contains(body, "Login") && !strings.Contains(body, "Sign in") {
		t.Fatalf("response body missing login form elements: %s", body)
	}
	assertContains(t, body, `<form method="POST" action="/admin/login"`)
	assertContains(t, body, `name="username"`)
	assertContains(t, body, `name="password"`)
}

func TestAdminLoginPageAuthenticated(t *testing.T) {
	app := newPublicTestApp(t)
	sessionCookie, _ := loginSession(t, app)

	req := httptest.NewRequest(http.MethodGet, "/admin/login", nil)
	req.AddCookie(sessionCookie)
	rr := httptest.NewRecorder()
	app.handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusSeeOther, rr.Body.String())
	}
	if got := rr.Header().Get("Location"); got != "/admin" {
		t.Fatalf("Location = %q, want /admin", got)
	}
}

func TestAdminLoginValidCredentials(t *testing.T) {
	app := newPublicTestApp(t)

	form := url.Values{"username": {"admin"}, "password": {"demo-pass-123"}}
	req := httptest.NewRequest(http.MethodPost, "/admin/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	testRequestCounter++
	req.RemoteAddr = fmt.Sprintf("203.0.113.%d:1234", testRequestCounter%250+1)
	rr := httptest.NewRecorder()
	app.handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusSeeOther, rr.Body.String())
	}
	if got := rr.Header().Get("Location"); got != "/admin" {
		t.Fatalf("Location = %q, want /admin", got)
	}

	// Verify session cookie is set
	cookies := rr.Result().Cookies()
	var sessionCookieFound bool
	for _, c := range cookies {
		if c.Name == sessionCookieName && c.Value != "" && c.HttpOnly {
			sessionCookieFound = true
			break
		}
	}
	if !sessionCookieFound {
		t.Fatalf("no valid session cookie in response")
	}
}

func TestAdminLoginInvalidPassword(t *testing.T) {
	app := newPublicTestApp(t)

	form := url.Values{"username": {"admin"}, "password": {"wrong-password"}}
	req := httptest.NewRequest(http.MethodPost, "/admin/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	testRequestCounter++
	req.RemoteAddr = fmt.Sprintf("192.0.2.%d:5678", testRequestCounter%250+1)
	rr := httptest.NewRecorder()
	app.handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusUnauthorized, rr.Body.String())
	}
	var payload map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload["error"] != "invalid credentials" {
		t.Fatalf("error = %q, want %q", payload["error"], "invalid credentials")
	}
}

func TestAdminLogout(t *testing.T) {
	app := newPublicTestApp(t)
	sessionCookie, csrfToken := loginSession(t, app)

	// Verify session exists before logout
	if _, ok := sessions.get(sessionCookie.Value); !ok {
		t.Fatalf("session should exist before logout")
	}

	form := url.Values{"_csrf": {csrfToken}}
	req := httptest.NewRequest(http.MethodPost, "/admin/logout", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(sessionCookie)
	rr := httptest.NewRecorder()
	app.handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusSeeOther, rr.Body.String())
	}
	if got := rr.Header().Get("Location"); got != "/admin/login" {
		t.Fatalf("Location = %q, want /admin/login", got)
	}

	// Verify session is invalidated
	if _, ok := sessions.get(sessionCookie.Value); ok {
		t.Fatalf("session should be deleted after logout")
	}
}

func TestAdminAPIAuthValidKey(t *testing.T) {
	app := newPublicTestApp(t)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/posts", strings.NewReader(`{"title":"Test","content_md":"# Test","category_slug":"devops"}`))
	req.Header.Set("Authorization", "Bearer admin-key")
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	app.handler.ServeHTTP(rr, req)

	// Should reach the handler (may 400 due to validation, but not 401)
	if rr.Code == http.StatusUnauthorized {
		t.Fatalf("got 401 with valid admin key; body: %s", rr.Body.String())
	}
}

func TestAdminAPIAuthMissingKey(t *testing.T) {
	app := newPublicTestApp(t)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/posts", nil)
	rr := httptest.NewRecorder()
	app.handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusUnauthorized, rr.Body.String())
	}
	var payload map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload["error"] == "" {
		t.Fatalf("expected error message in response body: %s", rr.Body.String())
	}
}

func TestAdminAPIAuthInvalidKey(t *testing.T) {
	app := newPublicTestApp(t)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/posts", nil)
	req.Header.Set("Authorization", "Bearer wrong-key")
	rr := httptest.NewRecorder()
	app.handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusUnauthorized, rr.Body.String())
	}
	var payload map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload["error"] == "" {
		t.Fatalf("expected error message in response body: %s", rr.Body.String())
	}
}

func TestAdminRequiredWithoutSession(t *testing.T) {
	app := newPublicTestApp(t)

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	rr := httptest.NewRecorder()
	app.handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusSeeOther, rr.Body.String())
	}
	if got := rr.Header().Get("Location"); got != "/admin/login" {
		t.Fatalf("Location = %q, want /admin/login", got)
	}
}

func TestAdminRequiredWithValidSession(t *testing.T) {
	app := newPublicTestApp(t)
	sessionCookie, _ := loginSession(t, app)

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	req.AddCookie(sessionCookie)
	rr := httptest.NewRecorder()
	app.handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}
}

func TestOwnerRequiredWithOperatorRole(t *testing.T) {
	app := newPublicTestApp(t)

	// Create an operator user via service
	operator, err := app.svc.Admin.CreateUser(model.RoleOwner, "admin", "operator-user", "operator-pass", model.RoleOperator, "127.0.0.1")
	if err != nil {
		t.Fatalf("create operator user: %v", err)
	}

	// Log in as operator to get a session
	form := url.Values{"username": {"operator-user"}, "password": {"operator-pass"}}
	loginReq := httptest.NewRequest(http.MethodPost, "/admin/login", strings.NewReader(form.Encode()))
	loginReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	testRequestCounter++
	loginReq.RemoteAddr = fmt.Sprintf("198.51.100.%d:9999", testRequestCounter%250+1)
	loginRR := httptest.NewRecorder()
	app.handler.ServeHTTP(loginRR, loginReq)

	if loginRR.Code != http.StatusSeeOther {
		t.Fatalf("operator login status = %d, want %d; body: %s", loginRR.Code, http.StatusSeeOther, loginRR.Body.String())
	}

	cookies := loginRR.Result().Cookies()
	var operatorCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == sessionCookieName {
			operatorCookie = c
			break
		}
	}
	if operatorCookie == nil {
		t.Fatalf("no session cookie for operator")
	}

	// Verify role in session
	entry, ok := sessions.get(operatorCookie.Value)
	if !ok || entry.role != model.RoleOperator {
		t.Fatalf("operator session role = %q, want %q", entry.role, model.RoleOperator)
	}

	_ = operator // created user unused

	// Build the ownerRequired middleware directly since no route uses it
	h := &Handler{svc: app.svc, cfg: app.svc.Cfg}
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("owner authorized"))
	})
	wrapped := h.ownerRequired(testHandler)

	req := httptest.NewRequest(http.MethodGet, "/admin/owner-only-test", nil)
	req.AddCookie(operatorCookie)
	rr := httptest.NewRecorder()
	wrapped.ServeHTTP(rr, req)

	// Operator should be rejected with 403
	if rr.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusForbidden, rr.Body.String())
	}
}

func TestOwnerRequiredWithOwnerRole(t *testing.T) {
	app := newPublicTestApp(t)

	// Create a session with owner role (seed admin user is owner)
	ownerCookie := &http.Cookie{Name: sessionCookieName, Value: sessions.create(201, "admin-owner", model.RoleOwner)}

	h := &Handler{svc: app.svc, cfg: app.svc.Cfg}
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("owner authorized"))
	})
	wrapped := h.ownerRequired(testHandler)

	req := httptest.NewRequest(http.MethodGet, "/admin/owner-only-test", nil)
	req.AddCookie(ownerCookie)
	rr := httptest.NewRecorder()
	wrapped.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}
}

func TestOwnerRequiredWithoutSession(t *testing.T) {
	app := newPublicTestApp(t)

	h := &Handler{svc: app.svc, cfg: app.svc.Cfg}
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("owner authorized"))
	})
	wrapped := h.ownerRequired(testHandler)

	req := httptest.NewRequest(http.MethodGet, "/admin/owner-only-test", nil)
	rr := httptest.NewRecorder()
	wrapped.ServeHTTP(rr, req)

	// No session → redirect to /admin/login
	if rr.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusSeeOther, rr.Body.String())
	}
	if got := rr.Header().Get("Location"); got != "/admin/login" {
		t.Fatalf("Location = %q, want /admin/login", got)
	}
}

func TestOwnerOrAdminRequiredWithOperatorRole(t *testing.T) {
	app := newPublicTestApp(t)

	// Create an operator user via service
	_, err := app.svc.Admin.CreateUser(model.RoleOwner, "admin", "operator-user", "operator-pass", model.RoleOperator, "127.0.0.1")
	if err != nil {
		t.Fatalf("create operator user: %v", err)
	}

	// Log in as operator
	form := url.Values{"username": {"operator-user"}, "password": {"operator-pass"}}
	loginReq := httptest.NewRequest(http.MethodPost, "/admin/login", strings.NewReader(form.Encode()))
	loginReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	testRequestCounter++
	loginReq.RemoteAddr = fmt.Sprintf("198.51.100.%d:9999", testRequestCounter%250+1)
	loginRR := httptest.NewRecorder()
	app.handler.ServeHTTP(loginRR, loginReq)

	if loginRR.Code != http.StatusSeeOther {
		t.Fatalf("operator login status = %d, want %d; body: %s", loginRR.Code, http.StatusSeeOther, loginRR.Body.String())
	}

	cookies := loginRR.Result().Cookies()
	var operatorCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == sessionCookieName {
			operatorCookie = c
			break
		}
	}
	if operatorCookie == nil {
		t.Fatalf("no session cookie for operator")
	}

	// Access /admin/manage-users which is protected by ownerOrAdminRequired
	req := httptest.NewRequest(http.MethodGet, "/admin/manage-users", nil)
	req.AddCookie(operatorCookie)
	rr := httptest.NewRecorder()
	app.handler.ServeHTTP(rr, req)

	// Operator should be rejected with 403
	if rr.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusForbidden, rr.Body.String())
	}
}

func TestOwnerOrAdminRequiredWithOwnerRole(t *testing.T) {
	app := newPublicTestApp(t)

	// Login as seed admin (which has owner role)
	sessionCookie, _ := loginSession(t, app)

	// Access /admin/manage-users which is protected by ownerOrAdminRequired
	req := httptest.NewRequest(http.MethodGet, "/admin/manage-users", nil)
	req.AddCookie(sessionCookie)
	rr := httptest.NewRecorder()
	app.handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}
}

func TestOwnerOrAdminRequiredWithReadonlyRole(t *testing.T) {
	app := newPublicTestApp(t)

	// Create session with readonly role
	readonlyCookie := &http.Cookie{Name: sessionCookieName, Value: sessions.create(202, "readonly-user", model.RoleReadonly)}

	req := httptest.NewRequest(http.MethodGet, "/admin/manage-users", nil)
	req.AddCookie(readonlyCookie)
	rr := httptest.NewRecorder()
	app.handler.ServeHTTP(rr, req)

	// Readonly should be rejected with 403
	if rr.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusForbidden, rr.Body.String())
	}
}

func TestAdminCSRFRejectsMissingToken(t *testing.T) {
	app := newPublicTestApp(t)
	sessionCookie, _ := loginSession(t, app)

	// POST /admin/logout without _csrf token
	req := httptest.NewRequest(http.MethodPost, "/admin/logout", nil)
	req.AddCookie(sessionCookie)
	rr := httptest.NewRecorder()
	app.handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusForbidden, rr.Body.String())
	}
}

func TestAdminCSRFRejectsInvalidToken(t *testing.T) {
	app := newPublicTestApp(t)
	sessionCookie, _ := loginSession(t, app)

	// POST /admin/logout with invalid CSRF token
	form := url.Values{"_csrf": {"invalid-csrf-token"}}
	req := httptest.NewRequest(http.MethodPost, "/admin/logout", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(sessionCookie)
	rr := httptest.NewRecorder()
	app.handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusForbidden, rr.Body.String())
	}
}

func TestAdminCSRFPassesWithValidToken(t *testing.T) {
	app := newPublicTestApp(t)
	sessionCookie, csrfToken := loginSession(t, app)

	// POST /admin/logout with valid CSRF token should succeed (303 redirect)
	form := url.Values{"_csrf": {csrfToken}}
	req := httptest.NewRequest(http.MethodPost, "/admin/logout", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(sessionCookie)
	rr := httptest.NewRecorder()
	app.handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusSeeOther, rr.Body.String())
	}
	if got := rr.Header().Get("Location"); got != "/admin/login" {
		t.Fatalf("Location = %q, want /admin/login", got)
	}
}

func TestAdminAPIAuthSessionNoSession(t *testing.T) {
	app := newPublicTestApp(t)

	// adminAPIAuth returns JSON 401 for unauthenticated requests (no redirect)
	h := &Handler{svc: app.svc, cfg: app.svc.Cfg}
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	wrapped := h.adminAPIAuth(testHandler)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/protected", nil)
	rr := httptest.NewRecorder()
	wrapped.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusUnauthorized, rr.Body.String())
	}
	var payload map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload["error"] != "not authenticated" {
		t.Fatalf("error = %q, want %q", payload["error"], "not authenticated")
	}
}

func TestCanManageServer(t *testing.T) {

	tests := []struct {
		name    string
		role    string
		wantCan bool
	}{
		{name: "owner", role: model.RoleOwner, wantCan: true},
		{name: "admin", role: model.RoleAdmin, wantCan: true},
		{name: "operator", role: model.RoleOperator, wantCan: false},
		{name: "readonly", role: model.RoleReadonly, wantCan: false},
		{name: "no session", role: "", wantCan: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var cookie *http.Cookie
			if tt.role != "" {
				cookie = &http.Cookie{Name: sessionCookieName, Value: sessions.create(300, tt.name, tt.role)}
			}

			req := httptest.NewRequest(http.MethodGet, "/admin/servers", nil)
			if cookie != nil {
				req.AddCookie(cookie)
			}
			got := canManageServer(req)
			if got != tt.wantCan {
				t.Fatalf("canManageServer(role=%q) = %v, want %v", tt.role, got, tt.wantCan)
			}
		})
	}
}

func TestCSRFTokenForRequest(t *testing.T) {

	token := sessions.create(301, "test-user", model.RoleOwner)
	cookie := &http.Cookie{Name: sessionCookieName, Value: token}

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	req.AddCookie(cookie)

	got := csrfTokenForRequest(req)
	if got == "" {
		t.Fatalf("csrfTokenForRequest returned empty string for valid session")
	}

	entry, ok := sessions.get(token)
	if !ok {
		t.Fatalf("session not found")
	}
	if got != entry.csrfToken {
		t.Fatalf("csrfTokenForRequest = %q, want %q", got, entry.csrfToken)
	}
}

func TestCSRFTokenForRequestNoSession(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	got := csrfTokenForRequest(req)
	if got != "" {
		t.Fatalf("csrfTokenForRequest = %q, want empty for no session", got)
	}
}
