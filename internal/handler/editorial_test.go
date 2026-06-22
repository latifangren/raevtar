package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"raevtar/internal/model"
	"raevtar/internal/service"
)

// loginAsAdmin logs in as the seed admin user and returns the session cookie
// and CSRF token from the in-memory session store.
func loginAsAdmin(t *testing.T, app *publicTestApp) (*http.Cookie, string) {
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

// createEditorialItem posts a form to create an editorial inbox item via the
// admin web handler. It does NOT fail the test on non-303 responses so callers
// can check error status codes themselves.
func createEditorialItem(t *testing.T, app *publicTestApp, cookie *http.Cookie, csrf, sourceValue, categoryHint, note string) *http.Response {
	t.Helper()
	form := url.Values{
		"source_value":  {sourceValue},
		"category_hint": {categoryHint},
		"note":          {note},
		"_csrf":         {csrf},
	}
	req := httptest.NewRequest(http.MethodPost, "/admin/editorial-inbox", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(cookie)
	rr := httptest.NewRecorder()
	app.handler.ServeHTTP(rr, req)
	return rr.Result()
}

// TestAdminEditorialInboxPage verifies the editorial inbox list page renders
// for an authenticated owner/admin user.
func TestAdminEditorialInboxPage(t *testing.T) {
	app := newPublicTestApp(t)
	sessionCookie, _ := loginAsAdmin(t, app)

	status, body := getBody(t, app, "/admin/editorial-inbox", sessionCookie)
	if status != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", status, http.StatusOK, body)
	}
	assertContains(t, body, "Editorial Inbox")
}

// TestAdminEditorialInboxPageUnauthenticated verifies the editorial inbox page
// redirects to the login page when no session is present.
func TestAdminEditorialInboxPageUnauthenticated(t *testing.T) {
	app := newPublicTestApp(t)

	req := httptest.NewRequest(http.MethodGet, "/admin/editorial-inbox", nil)
	rr := httptest.NewRecorder()
	app.handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusSeeOther, rr.Body.String())
	}
	if got := rr.Header().Get("Location"); got != "/admin/login" {
		t.Fatalf("Location = %q, want /admin/login", got)
	}
}

// TestAdminCreateEditorialInbox verifies that POSTing valid form data to the
// create endpoint returns a 303 redirect to the inbox list.
func TestAdminCreateEditorialInbox(t *testing.T) {
	app := newPublicTestApp(t)
	sessionCookie, csrfToken := loginAsAdmin(t, app)

	form := url.Values{
		"source_value":  {"https://github.com/example/test-article"},
		"category_hint": {"devops"},
		"note":          {"Research this topic before the next sprint."},
		"_csrf":         {csrfToken},
	}
	req := httptest.NewRequest(http.MethodPost, "/admin/editorial-inbox", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(sessionCookie)
	rr := httptest.NewRecorder()
	app.handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Fatalf("create status = %d, want %d; body: %s", rr.Code, http.StatusSeeOther, rr.Body.String())
	}
	if got := rr.Header().Get("Location"); got != "/admin/editorial-inbox" {
		t.Fatalf("Location = %q, want /admin/editorial-inbox", got)
	}
}

// TestAdminCreateEditorialInboxNoCSRF verifies that a POST without a CSRF
// token is rejected with 403.
func TestAdminCreateEditorialInboxNoCSRF(t *testing.T) {
	app := newPublicTestApp(t)
	sessionCookie, _ := loginAsAdmin(t, app)

	form := url.Values{
		"source_value": {"https://github.com/example/no-csrf"},
		"note":         {"Missing CSRF token."},
	}
	req := httptest.NewRequest(http.MethodPost, "/admin/editorial-inbox", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(sessionCookie)
	rr := httptest.NewRecorder()
	app.handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusForbidden, rr.Body.String())
	}
	assertContains(t, rr.Body.String(), "invalid CSRF token")
}

// TestAdminCreateEditorialInboxInvalidCSRF verifies that a POST with a wrong
// CSRF token is rejected with 403.
func TestAdminCreateEditorialInboxInvalidCSRF(t *testing.T) {
	app := newPublicTestApp(t)
	sessionCookie, _ := loginAsAdmin(t, app)

	form := url.Values{
		"source_value": {"https://github.com/example/bad-csrf"},
		"note":         {"Wrong CSRF token."},
		"_csrf":        {"this-is-not-the-right-token"},
	}
	req := httptest.NewRequest(http.MethodPost, "/admin/editorial-inbox", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(sessionCookie)
	rr := httptest.NewRecorder()
	app.handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusForbidden, rr.Body.String())
	}
	assertContains(t, rr.Body.String(), "invalid CSRF token")
}

// TestAdminEditorialInboxPageShowsItems verifies that after creating an
// editorial inbox item, its title appears on the list page.
func TestAdminEditorialInboxPageShowsItems(t *testing.T) {
	app := newPublicTestApp(t)
	sessionCookie, csrfToken := loginAsAdmin(t, app)

	resp := createEditorialItem(t, app, sessionCookie, csrfToken,
		"https://github.com/example/visible-item",
		"tools",
		"This item should appear on the editorial inbox list.",
	)
	if resp.StatusCode != http.StatusSeeOther {
		t.Fatalf("create status = %d, want %d", resp.StatusCode, http.StatusSeeOther)
	}

	status, body := getBody(t, app, "/admin/editorial-inbox", sessionCookie)
	if status != http.StatusOK {
		t.Fatalf("list status = %d, want %d; body: %s", status, http.StatusOK, body)
	}
	assertContains(t, body, "https://github.com/example/visible-item")
	assertContains(t, body, "This item should appear on the editorial inbox list.")
}

// TestAdminUpdateEditorialInbox verifies the full create-then-update flow:
// create an item, update its fields via POST to the update endpoint, and
// verify the changes on the list page.
func TestAdminUpdateEditorialInbox(t *testing.T) {
	app := newPublicTestApp(t)
	sessionCookie, csrfToken := loginAsAdmin(t, app)

	// Step 1: Create an item
	createForm := url.Values{
		"source_value":  {"https://github.com/example/update-me"},
		"category_hint": {"devops"},
		"note":          {"This is the original note before update."},
		"priority":      {"30"},
		"_csrf":         {csrfToken},
	}
	createReq := httptest.NewRequest(http.MethodPost, "/admin/editorial-inbox", strings.NewReader(createForm.Encode()))
	createReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	createReq.AddCookie(sessionCookie)
	createRR := httptest.NewRecorder()
	app.handler.ServeHTTP(createRR, createReq)
	if createRR.Code != http.StatusSeeOther {
		t.Fatalf("create status = %d, want %d; body: %s", createRR.Code, http.StatusSeeOther, createRR.Body.String())
	}

	// Step 2: Find the item's ID from the service layer
	items, err := app.svc.Editorial.ListInboxItems(service.EditorialInboxListFilter{})
	if err != nil {
		t.Fatalf("list items: %v", err)
	}
	var itemID int64
	for _, item := range items {
		if item.SourceValue == "https://github.com/example/update-me" {
			itemID = item.ID
			break
		}
	}
	if itemID == 0 {
		t.Fatalf("expected to find created item with source_value %q", "https://github.com/example/update-me")
	}

	// Step 3: Update the item
	updateForm := url.Values{
		"source_type":   {"repo"},
		"source_value":  {"https://github.com/example/updated"},
		"category_hint": {"tools"},
		"note":          {"This note was updated after creation."},
		"priority":      {"75"},
		"not_before":    {"2026-06-25T10:00"},
		"mode":          {model.EditorialModeScheduled},
		"status":        {model.EditorialStatusApproved},
		"_csrf":         {csrfToken},
	}
	updateReq := httptest.NewRequest(http.MethodPost, "/admin/editorial-inbox/update/"+strconv.FormatInt(itemID, 10), strings.NewReader(updateForm.Encode()))
	updateReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	updateReq.AddCookie(sessionCookie)
	updateRR := httptest.NewRecorder()
	app.handler.ServeHTTP(updateRR, updateReq)
	if updateRR.Code != http.StatusSeeOther {
		t.Fatalf("update status = %d, want %d; body: %s", updateRR.Code, http.StatusSeeOther, updateRR.Body.String())
	}

	// Step 4: List page should show the updated values
	status, body := getBody(t, app, "/admin/editorial-inbox", sessionCookie)
	if status != http.StatusOK {
		t.Fatalf("list status = %d, want %d; body: %s", status, http.StatusOK, body)
	}
	assertContains(t, body, "https://github.com/example/updated")
	assertContains(t, body, "This note was updated after creation.")
}

// TestAdminUpdateEditorialInboxUnauthorized verifies that POSTing to the
// update endpoint without a session redirects to the login page.
func TestAdminUpdateEditorialInboxUnauthorized(t *testing.T) {
	app := newPublicTestApp(t)

	form := url.Values{
		"source_value": {"https://github.com/example/unauth-update"},
		"note":         {"No session should block this."},
	}
	req := httptest.NewRequest(http.MethodPost, "/admin/editorial-inbox/update/1", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	app.handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusSeeOther, rr.Body.String())
	}
	if got := rr.Header().Get("Location"); got != "/admin/login" {
		t.Fatalf("Location = %q, want /admin/login", got)
	}
}

// TestAdminCreateEditorialInboxWithAllFields verifies the create endpoint
// accepts all optional fields (mode, status, priority, dates) correctly.
func TestAdminCreateEditorialInboxWithAllFields(t *testing.T) {
	app := newPublicTestApp(t)
	sessionCookie, csrfToken := loginAsAdmin(t, app)

	form := url.Values{
		"source_type":   {"issue"},
		"source_value":  {"https://github.com/example/full-fields"},
		"category_hint": {"ai-agent"},
		"priority":      {"90"},
		"not_before":    {"2026-06-25T10:00"},
		"deadline":      {"2026-06-30T18:00"},
		"note":          {"Full field test with all options."},
		"mode":          {model.EditorialModeScheduled},
		"status":        {model.EditorialStatusQueued},
		"_csrf":         {csrfToken},
	}
	req := httptest.NewRequest(http.MethodPost, "/admin/editorial-inbox", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(sessionCookie)
	rr := httptest.NewRecorder()
	app.handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Fatalf("create status = %d, want %d; body: %s", rr.Code, http.StatusSeeOther, rr.Body.String())
	}

	// Verify the item was stored with correct field values
	items, err := app.svc.Editorial.ListInboxItems(service.EditorialInboxListFilter{})
	if err != nil {
		t.Fatalf("list items: %v", err)
	}
	var found bool
	for _, item := range items {
		if item.SourceValue == "https://github.com/example/full-fields" {
			found = true
			if item.SourceType != "issue" {
				t.Fatalf("SourceType = %q, want %q", item.SourceType, "issue")
			}
			if item.CategoryHint != "ai-agent" {
				t.Fatalf("CategoryHint = %q, want %q", item.CategoryHint, "ai-agent")
			}
			if item.Priority != 90 {
				t.Fatalf("Priority = %d, want %d", item.Priority, 90)
			}
			if item.Mode != model.EditorialModeScheduled {
				t.Fatalf("Mode = %q, want %q", item.Mode, model.EditorialModeScheduled)
			}
			if item.Status != model.EditorialStatusQueued {
				t.Fatalf("Status = %q, want %q", item.Status, model.EditorialStatusQueued)
			}
			break
		}
	}
	if !found {
		t.Fatalf("expected to find created item with source_value %q", "https://github.com/example/full-fields")
	}
}

// TestAdminEditorialInboxUpdateNotFound verifies that updating a non-existent
// item returns 404.
func TestAdminEditorialInboxUpdateNotFound(t *testing.T) {
	app := newPublicTestApp(t)
	sessionCookie, csrfToken := loginAsAdmin(t, app)

	form := url.Values{
		"source_type":  {"repo"},
		"source_value": {"https://github.com/example/ghost"},
		"not_before":   {"2026-06-25T10:00"},
		"mode":         {model.EditorialModeScheduled},
		"status":       {model.EditorialStatusApproved},
		"priority":     {"50"},
		"_csrf":        {csrfToken},
	}
	req := httptest.NewRequest(http.MethodPost, "/admin/editorial-inbox/update/99999", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(sessionCookie)
	rr := httptest.NewRecorder()
	app.handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusNotFound, rr.Body.String())
	}
	assertContains(t, rr.Body.String(), "not found")
}

// ---------- Editorial inbox API endpoint tests ----------

func TestAPIEditorialInboxSummary(t *testing.T) {
	app := newPublicTestApp(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/editorial-inbox/summary", nil)
	req.Header.Set("Authorization", "Bearer admin-key")
	testRequestCounter++
	req.RemoteAddr = fmt.Sprintf("203.0.113.%d:1234", testRequestCounter%250+1)
	rr := httptest.NewRecorder()
	app.handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}
	var summary model.EditorialInboxSummary
	if err := json.NewDecoder(rr.Body).Decode(&summary); err != nil {
		t.Fatalf("decode summary: %v", err)
	}
}

func TestAPICreateEditorialInbox(t *testing.T) {
	app := newPublicTestApp(t)

	payload := `{"source_type":"repo","source_value":"https://github.com/example/test","priority":50,"not_before":"2025-01-01T00:00:00Z","mode":"scheduled_assignment","status":"approved"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/editorial-inbox", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer admin-key")
	testRequestCounter++
	req.RemoteAddr = fmt.Sprintf("203.0.113.%d:1234", testRequestCounter%250+1)
	rr := httptest.NewRecorder()
	app.handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusCreated, rr.Body.String())
	}
	var item model.EditorialInboxItem
	if err := json.NewDecoder(rr.Body).Decode(&item); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if item.SourceValue != "https://github.com/example/test" {
		t.Fatalf("source_value = %q, want %q", item.SourceValue, "https://github.com/example/test")
	}
	if item.Status != "approved" {
		t.Fatalf("status = %q, want %q", item.Status, "approved")
	}
}

func TestAPIClaimEditorialInbox(t *testing.T) {
	app := newPublicTestApp(t)

	createPayload := "{\"source_type\":\"repo\",\"source_value\":\"https://github.com/example/claim-test\",\"priority\":90,\"not_before\":\"2025-01-01T00:00:00Z\",\"mode\":\"scheduled_assignment\",\"status\":\"approved\"}"
	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/editorial-inbox", strings.NewReader(createPayload))
	createReq.Header.Set("Content-Type", "application/json")
	createReq.Header.Set("Authorization", "Bearer admin-key")
	testRequestCounter++
	createReq.RemoteAddr = fmt.Sprintf("203.0.113.%d:1234", testRequestCounter%250+1)
	createRR := httptest.NewRecorder()
	app.handler.ServeHTTP(createRR, createReq)
	if createRR.Code != http.StatusCreated {
		t.Fatalf("create status = %d, want %d; body: %s", createRR.Code, http.StatusCreated, createRR.Body.String())
	}

	claimReq := httptest.NewRequest(http.MethodPost, "/api/v1/editorial-inbox/claim", strings.NewReader("{\"worker\":\"test-worker\"}"))
	claimReq.Header.Set("Content-Type", "application/json")
	claimReq.Header.Set("Authorization", "Bearer admin-key")
	testRequestCounter++
	claimReq.RemoteAddr = fmt.Sprintf("203.0.113.%d:1234", testRequestCounter%250+1)
	claimRR := httptest.NewRecorder()
	app.handler.ServeHTTP(claimRR, claimReq)

	if claimRR.Code != http.StatusOK {
		t.Fatalf("claim status = %d, want %d; body: %s", claimRR.Code, http.StatusOK, claimRR.Body.String())
	}
	var result model.EditorialInboxClaimResult
	if err := json.NewDecoder(claimRR.Body).Decode(&result); err != nil {
		t.Fatalf("decode claim result: %v", err)
	}
	if result.ClaimToken == "" {
		t.Fatal("claim_token is empty")
	}
	if result.Item == nil || result.Item.ID == 0 {
		t.Fatal("claim result item is nil or has zero ID")
	}
}

func TestAPICompleteEditorialInboxClaim(t *testing.T) {
	app := newPublicTestApp(t)

	post, err := app.svc.Blog.CreatePost(model.PostCreate{
		CategorySlug: "devops",
		Title:        "Editorial Complete Test Post",
		ContentMD:    "# Test",
		Published:    true,
	})
	if err != nil {
		t.Fatalf("create post: %v", err)
	}

	createPayload := "{\"source_type\":\"repo\",\"source_value\":\"https://github.com/example/complete-test\",\"priority\":90,\"not_before\":\"2025-01-01T00:00:00Z\",\"mode\":\"scheduled_assignment\",\"status\":\"approved\"}"
	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/editorial-inbox", strings.NewReader(createPayload))
	createReq.Header.Set("Content-Type", "application/json")
	createReq.Header.Set("Authorization", "Bearer admin-key")
	testRequestCounter++
	createReq.RemoteAddr = fmt.Sprintf("203.0.113.%d:1234", testRequestCounter%250+1)
	createRR := httptest.NewRecorder()
	app.handler.ServeHTTP(createRR, createReq)
	if createRR.Code != http.StatusCreated {
		t.Fatalf("create status = %d, want %d; body: %s", createRR.Code, http.StatusCreated, createRR.Body.String())
	}
	var createdItem model.EditorialInboxItem
	if err := json.NewDecoder(createRR.Body).Decode(&createdItem); err != nil {
		t.Fatalf("decode created item: %v", err)
	}

	claimReq := httptest.NewRequest(http.MethodPost, "/api/v1/editorial-inbox/claim", strings.NewReader("{\"worker\":\"test-worker\"}"))
	claimReq.Header.Set("Content-Type", "application/json")
	claimReq.Header.Set("Authorization", "Bearer admin-key")
	testRequestCounter++
	claimReq.RemoteAddr = fmt.Sprintf("203.0.113.%d:1234", testRequestCounter%250+1)
	claimRR := httptest.NewRecorder()
	app.handler.ServeHTTP(claimRR, claimReq)
	if claimRR.Code != http.StatusOK {
		t.Fatalf("claim status = %d, want %d; body: %s", claimRR.Code, http.StatusOK, claimRR.Body.String())
	}
	var claimResult model.EditorialInboxClaimResult
	if err := json.NewDecoder(claimRR.Body).Decode(&claimResult); err != nil {
		t.Fatalf("decode claim: %v", err)
	}

	completePayload := "{\"claim_token\":\"" + claimResult.ClaimToken + "\",\"published_post_id\":" + strconv.FormatInt(post.ID, 10) + "}"
	completeReq := httptest.NewRequest(http.MethodPost, "/api/v1/editorial-inbox/"+strconv.FormatInt(createdItem.ID, 10)+"/complete", strings.NewReader(completePayload))
	completeReq.Header.Set("Content-Type", "application/json")
	completeReq.Header.Set("Authorization", "Bearer admin-key")
	testRequestCounter++
	completeReq.RemoteAddr = fmt.Sprintf("203.0.113.%d:1234", testRequestCounter%250+1)
	completeRR := httptest.NewRecorder()
	app.handler.ServeHTTP(completeRR, completeReq)

	if completeRR.Code != http.StatusOK {
		t.Fatalf("complete status = %d, want %d; body: %s", completeRR.Code, http.StatusOK, completeRR.Body.String())
	}
	var completedItem model.EditorialInboxItem
	if err := json.NewDecoder(completeRR.Body).Decode(&completedItem); err != nil {
		t.Fatalf("decode completed item: %v", err)
	}
	if completedItem.Status != "done" {
		t.Fatalf("status = %q, want %q", completedItem.Status, "done")
	}
	if completedItem.PublishedPostID == nil || *completedItem.PublishedPostID != post.ID {
		t.Fatalf("published_post_id = %v, want %d", completedItem.PublishedPostID, post.ID)
	}
}

// ---------- API list editorial inbox ----------

func TestAPIListEditorialInbox(t *testing.T) {
	app := newPublicTestApp(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/editorial-inbox", nil)
	req.Header.Set("Authorization", "Bearer admin-key")
	testRequestCounter++
	req.RemoteAddr = fmt.Sprintf("203.0.113.%d:1234", testRequestCounter%250+1)
	rr := httptest.NewRecorder()
	app.handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}
	var items []model.EditorialInboxItem
	if err := json.NewDecoder(rr.Body).Decode(&items); err != nil {
		t.Fatalf("decode response: %v", err)
	}
}
