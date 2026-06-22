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
)

func TestAdminEditProjectPage(t *testing.T) {
	app := newPublicTestApp(t)
	sessionCookie, _ := loginSession(t, app)
	status, body := getBody(t, app, "/admin/projects/edit/1", sessionCookie)
	if status != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", status, http.StatusOK, body)
	}
	assertContains(t, body, "Edit Project")
	assertContains(t, body, "whyred-watchtower")
	assertContains(t, body, "name=\"_csrf\"")
}
func TestAdminEditProjectNotFound(t *testing.T) {
	app := newPublicTestApp(t)
	sessionCookie, _ := loginSession(t, app)
	status, body := getBody(t, app, "/admin/projects/edit/999", sessionCookie)
	if status != http.StatusNotFound {
		t.Fatalf("status = %d, want %d; body: %s", status, http.StatusNotFound, body)
	}
	assertContains(t, body, "project not found")
}
func TestAdminPreviewProject(t *testing.T) {
	app := newPublicTestApp(t)
	sessionCookie, csrf := loginSession(t, app)
	form := url.Values{
		"_csrf":   {csrf},
		"content": {"# Preview Title\n\n- **one**\n- two"},
	}
	req := httptest.NewRequest(http.MethodPost, "/admin/projects/preview", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(sessionCookie)
	rr := httptest.NewRecorder()
	app.handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}
	body := rr.Body.String()
	assertContains(t, body, "<h1>Preview Title</h1>")
	assertContains(t, body, "<strong>one</strong>")
	form.Set("content", "   ")
	req = httptest.NewRequest(http.MethodPost, "/admin/projects/preview", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(sessionCookie)
	rr = httptest.NewRecorder()
	app.handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("blank preview status = %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}
	assertContains(t, rr.Body.String(), "Write Markdown, then preview without saving.")
}
func TestAdminCreateProjectUpdate(t *testing.T) {
	app := newPublicTestApp(t)
	sessionCookie, csrf := loginSession(t, app)
	form := url.Values{
		"_csrf":      {csrf},
		"kind":       {"changelog"},
		"title":      {"Fixed the watcher loop"},
		"content_md": {"# Watcher fix\n\nReduced polling interval."},
	}
	req := httptest.NewRequest(http.MethodPost, "/admin/projects/1/updates", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(sessionCookie)
	rr := httptest.NewRecorder()
	app.handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusSeeOther, rr.Body.String())
	}
	loc := rr.Header().Get("Location")
	if loc != "/admin/projects/edit/1" {
		t.Fatalf("Location = %q, want /admin/projects/edit/1", loc)
	}
}
func TestAdminUpdateProjectUpdate(t *testing.T) {
	app := newPublicTestApp(t)
	sessionCookie, csrf := loginSession(t, app)
	update, err := app.svc.Projects.CreateProjectUpdate(1, model.ProjectUpdateEntryCreate{
		Kind:      "changelog",
		Title:     "Original title",
		ContentMD: "# Original",
		Published: true,
	})
	if err != nil {
		t.Fatalf("create update: %v", err)
	}
	form := url.Values{
		"_csrf":      {csrf},
		"project_id": {"1"},
		"kind":       {"changelog"},
		"title":      {"Updated title"},
		"content_md": {"# Updated"},
	}
	req := httptest.NewRequest(http.MethodPost, "/admin/projects/updates/update/"+strconv.FormatInt(update.ID, 10), strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(sessionCookie)
	rr := httptest.NewRecorder()
	app.handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusSeeOther, rr.Body.String())
	}
}
func TestAdminDeleteProjectUpdate(t *testing.T) {
	app := newPublicTestApp(t)
	sessionCookie, csrf := loginSession(t, app)
	update, err := app.svc.Projects.CreateProjectUpdate(1, model.ProjectUpdateEntryCreate{
		Kind:      "changelog",
		Title:     "To delete",
		ContentMD: "# Delete me",
		Published: true,
	})
	if err != nil {
		t.Fatalf("create update: %v", err)
	}
	form := url.Values{
		"_csrf":      {csrf},
		"project_id": {"1"},
	}
	req := httptest.NewRequest(http.MethodPost, "/admin/projects/updates/delete/"+strconv.FormatInt(update.ID, 10), strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(sessionCookie)
	rr := httptest.NewRecorder()
	app.handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusSeeOther, rr.Body.String())
	}
	loc := rr.Header().Get("Location")
	if loc != "/admin/projects/edit/1" {
		t.Fatalf("Location = %q, want /admin/projects/edit/1", loc)
	}
}
func TestAdminCreateProjectRelation(t *testing.T) {
	app := newPublicTestApp(t)
	sessionCookie, csrf := loginSession(t, app)
	// Create a second project to relate to
	related, err := app.svc.Projects.CreateProject(model.ProjectCreate{
		Title:     "Sister Project",
		ContentMD: "# Sister Project",
		Excerpt:   "related sister",
		Published: true,
		SortOrder: 2,
	})
	if err != nil {
		t.Fatalf("create sister project: %v", err)
	}
	form := url.Values{
		"_csrf":         {csrf},
		"target_id":     {strconv.FormatInt(related.ID, 10)},
		"target_type":   {"project"},
		"relation_kind": {"related"},
	}
	req := httptest.NewRequest(http.MethodPost, "/admin/projects/1/relations", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(sessionCookie)
	rr := httptest.NewRecorder()
	app.handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusSeeOther, rr.Body.String())
	}
	loc := rr.Header().Get("Location")
	if loc != "/admin/projects/edit/1" {
		t.Fatalf("Location = %q, want /admin/projects/edit/1", loc)
	}
}
func TestAdminDeleteProjectRelation(t *testing.T) {
	app := newPublicTestApp(t)
	sessionCookie, csrf := loginSession(t, app)
	_, err := app.svc.Projects.CreateProject(model.ProjectCreate{Title: "Sibling Project", ContentMD: "# Sibling", State: "active"})
	if err != nil {
		t.Fatalf("create sibling project: %v", err)
	}
	rel, err := app.svc.Projects.CreateProjectRelation(1, model.ContentRelationCreate{
		TargetType:   "project",
		TargetID:     2,
		RelationKind: "related",
	})
	if err != nil {
		t.Fatalf("create relation: %v", err)
	}
	form := url.Values{
		"_csrf":      {csrf},
		"project_id": {"1"},
	}
	req := httptest.NewRequest(http.MethodPost, "/admin/projects/relations/delete/"+strconv.FormatInt(rel.ID, 10), strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(sessionCookie)
	rr := httptest.NewRecorder()
	app.handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusSeeOther, rr.Body.String())
	}
}
func TestAdminCreateProjectShowcase(t *testing.T) {
	app := newPublicTestApp(t)
	sessionCookie, csrf := loginSession(t, app)
	form := url.Values{
		"_csrf":   {csrf},
		"kind":    {"image"},
		"title":   {"Project Screenshot"},
		"body_md": {"A screenshot of the main dashboard."},
	}
	req := httptest.NewRequest(http.MethodPost, "/admin/projects/1/showcase", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(sessionCookie)
	rr := httptest.NewRecorder()
	app.handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusSeeOther, rr.Body.String())
	}
	loc := rr.Header().Get("Location")
	if loc != "/admin/projects/edit/1" {
		t.Fatalf("Location = %q, want /admin/projects/edit/1", loc)
	}
}
func TestAdminUpdateProjectShowcase(t *testing.T) {
	app := newPublicTestApp(t)
	sessionCookie, csrf := loginSession(t, app)
	item, err := app.svc.Projects.CreateProjectShowcase(1, model.ProjectShowcaseItemCreate{
		Kind:      "image",
		Title:     "Original",
		BodyMD:    "# Original",
		Published: true,
	})
	if err != nil {
		t.Fatalf("create showcase: %v", err)
	}
	form := url.Values{
		"_csrf":      {csrf},
		"project_id": {"1"},
		"kind":       {"image"},
		"title":      {"Updated Showcase"},
		"body_md":    {"# Updated"},
	}
	req := httptest.NewRequest(http.MethodPost, "/admin/projects/showcase/update/"+strconv.FormatInt(item.ID, 10), strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(sessionCookie)
	rr := httptest.NewRecorder()
	app.handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusSeeOther, rr.Body.String())
	}
}
func TestAdminDeleteProjectShowcase(t *testing.T) {
	app := newPublicTestApp(t)
	sessionCookie, csrf := loginSession(t, app)
	item, err := app.svc.Projects.CreateProjectShowcase(1, model.ProjectShowcaseItemCreate{
		Kind:      "image",
		Title:     "To delete",
		BodyMD:    "# Delete me",
		Published: true,
	})
	if err != nil {
		t.Fatalf("create showcase: %v", err)
	}
	form := url.Values{
		"_csrf":      {csrf},
		"project_id": {"1"},
	}
	req := httptest.NewRequest(http.MethodPost, "/admin/projects/showcase/delete/"+strconv.FormatInt(item.ID, 10), strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(sessionCookie)
	rr := httptest.NewRecorder()
	app.handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusSeeOther, rr.Body.String())
	}
}
func TestParseDateTimeValue(t *testing.T) {
	t.Run("valid time", func(t *testing.T) {
		got := parseDateTimeValue("2026-06-22T15:04")
		if got.IsZero() {
			t.Fatalf("expected non-zero time for valid input")
		}
		if got.Year() != 2026 || got.Month() != 6 || got.Day() != 22 || got.Hour() != 15 || got.Minute() != 4 {
			t.Fatalf("parsed time = %v, want 2026-06-22 15:04", got)
		}
	})
	t.Run("empty string", func(t *testing.T) {
		got := parseDateTimeValue("")
		if !got.IsZero() {
			t.Fatalf("expected zero time for empty string, got %v", got)
		}
	})
	t.Run("whitespace", func(t *testing.T) {
		got := parseDateTimeValue("   ")
		if !got.IsZero() {
			t.Fatalf("expected zero time for whitespace, got %v", got)
		}
	})
	t.Run("invalid format", func(t *testing.T) {
		got := parseDateTimeValue("not-a-date")
		if !got.IsZero() {
			t.Fatalf("expected zero time for invalid format, got %v", got)
		}
	})
}
func TestAdminEditPage(t *testing.T) {
	app := newPublicTestApp(t)
	sessionCookie, _ := loginSession(t, app)
	status, body := getBody(t, app, "/admin/pages/edit/about", sessionCookie)
	if status != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", status, http.StatusOK, body)
	}
	assertContains(t, body, "Edit")
	assertContains(t, body, `name="_csrf"`)
}
func TestAdminEditPageContact(t *testing.T) {
	app := newPublicTestApp(t)
	sessionCookie, _ := loginSession(t, app)
	status, body := getBody(t, app, "/admin/pages/edit/contact", sessionCookie)
	if status != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", status, http.StatusOK, body)
	}
	assertContains(t, body, "Edit")
	assertContains(t, body, `name="_csrf"`)
}
func TestAdminServerCommand(t *testing.T) {
	app := newPublicTestApp(t)
	sessionCookie, csrf := loginSession(t, app)
	form := url.Values{
		"_csrf":   {csrf},
		"command": {"CLEAR_CACHE"},
	}
	req := httptest.NewRequest(http.MethodPost, "/admin/servers/command/"+strconv.FormatInt(app.serverID, 10), strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(sessionCookie)
	rr := httptest.NewRecorder()
	app.handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusSeeOther, rr.Body.String())
	}
	loc := rr.Header().Get("Location")
	wantLoc := "/admin/servers/" + strconv.FormatInt(app.serverID, 10)
	if loc != wantLoc {
		t.Fatalf("Location = %q, want %q", loc, wantLoc)
	}
}
func TestAdminServerCommandEmptyCommand(t *testing.T) {
	app := newPublicTestApp(t)
	sessionCookie, csrf := loginSession(t, app)
	form := url.Values{"_csrf": {csrf}}
	req := httptest.NewRequest(http.MethodPost, "/admin/servers/command/"+strconv.FormatInt(app.serverID, 10), strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(sessionCookie)
	rr := httptest.NewRecorder()
	app.handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusBadRequest, rr.Body.String())
	}
	assertContains(t, rr.Body.String(), "command required")
}
func TestAdminServerCommandInvalidCommand(t *testing.T) {
	app := newPublicTestApp(t)
	sessionCookie, csrf := loginSession(t, app)
	form := url.Values{
		"_csrf":   {csrf},
		"command": {"INVALID_CMD"},
	}
	req := httptest.NewRequest(http.MethodPost, "/admin/servers/command/"+strconv.FormatInt(app.serverID, 10), strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(sessionCookie)
	rr := httptest.NewRecorder()
	app.handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusBadRequest, rr.Body.String())
	}
	assertContains(t, rr.Body.String(), "invalid command")
}
func TestAdminServerCommandGET(t *testing.T) {
	app := newPublicTestApp(t)
	sessionCookie, _ := loginSession(t, app)
	req := httptest.NewRequest(http.MethodGet, "/admin/servers/command/"+strconv.FormatInt(app.serverID, 10), nil)
	req.AddCookie(sessionCookie)
	rr := httptest.NewRecorder()
	app.handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusMethodNotAllowed && rr.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d or %d; body: %s", rr.Code, http.StatusMethodNotAllowed, http.StatusNotFound, rr.Body.String())
	}
}
func TestAPIHostStats(t *testing.T) {
	app := newPublicTestApp(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/hoststats", nil)
	req.Header.Set("Authorization", "Bearer admin-key")
	rr := httptest.NewRecorder()
	app.handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}
	var stats HostStats
	if err := json.NewDecoder(rr.Body).Decode(&stats); err != nil {
		t.Fatalf("decode hoststats: %v", err)
	}
	if stats.CPU.Cores < 0 {
		t.Fatalf("unexpected CPU cores: %d", stats.CPU.Cores)
	}
}
func TestAPIHostStatsRequiresAuth(t *testing.T) {
	app := newPublicTestApp(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/hoststats", nil)
	rr := httptest.NewRecorder()
	app.handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusUnauthorized, rr.Body.String())
	}
}
func TestAPIGetEditorialInbox(t *testing.T) {
	app := newPublicTestApp(t)
	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/editorial-inbox", strings.NewReader(`{"source_type":"repo","source_value":"https://github.com/example/test-get","priority":50,"not_before":"2025-06-05T08:00:00Z","mode":"scheduled_assignment","status":"approved"}`))
	createReq.Header.Set("Authorization", "Bearer admin-key")
	createReq.Header.Set("Content-Type", "application/json")
	createRR := httptest.NewRecorder()
	app.handler.ServeHTTP(createRR, createReq)
	if createRR.Code != http.StatusCreated {
		t.Fatalf("create status = %d, want %d; body: %s", createRR.Code, http.StatusCreated, createRR.Body.String())
	}
	var created model.EditorialInboxItem
	if err := json.Unmarshal(createRR.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode created: %v", err)
	}
	getReq := httptest.NewRequest(http.MethodGet, "/api/v1/editorial-inbox/"+strconv.FormatInt(created.ID, 10), nil)
	getReq.Header.Set("Authorization", "Bearer admin-key")
	getRR := httptest.NewRecorder()
	app.handler.ServeHTTP(getRR, getReq)
	if getRR.Code != http.StatusOK {
		t.Fatalf("get status = %d, want %d; body: %s", getRR.Code, http.StatusOK, getRR.Body.String())
	}
	var fetched model.EditorialInboxItem
	if err := json.Unmarshal(getRR.Body.Bytes(), &fetched); err != nil {
		t.Fatalf("decode fetched: %v", err)
	}
	if fetched.ID != created.ID || fetched.SourceValue != "https://github.com/example/test-get" {
		t.Fatalf("fetched = %+v, want id=%d source=%s", fetched, created.ID, "https://github.com/example/test-get")
	}
}
func TestAPIGetEditorialInboxNotFound(t *testing.T) {
	app := newPublicTestApp(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/editorial-inbox/99999", nil)
	req.Header.Set("Authorization", "Bearer admin-key")
	rr := httptest.NewRecorder()
	app.handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusNotFound, rr.Body.String())
	}
	assertContains(t, rr.Body.String(), "not found")
}
func TestAPIUpdateEditorialInbox(t *testing.T) {
	app := newPublicTestApp(t)
	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/editorial-inbox", strings.NewReader(`{"source_type":"repo","source_value":"https://github.com/example/test-update","priority":50,"not_before":"2025-06-05T08:00:00Z","mode":"scheduled_assignment","status":"approved"}`))
	createReq.Header.Set("Authorization", "Bearer admin-key")
	createReq.Header.Set("Content-Type", "application/json")
	createRR := httptest.NewRecorder()
	app.handler.ServeHTTP(createRR, createReq)
	if createRR.Code != http.StatusCreated {
		t.Fatalf("create status = %d, want %d; body: %s", createRR.Code, http.StatusCreated, createRR.Body.String())
	}
	var created model.EditorialInboxItem
	if err := json.Unmarshal(createRR.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode created: %v", err)
	}
	updatePayload := `{"source_type":"repo","source_value":"https://github.com/example/test-update","note":"updated note","priority":80,"not_before":"2025-06-05T08:00:00Z","mode":"scheduled_assignment","status":"approved"}`
	updateReq := httptest.NewRequest(http.MethodPost, "/api/v1/editorial-inbox/"+strconv.FormatInt(created.ID, 10), strings.NewReader(updatePayload))
	updateReq.Header.Set("Authorization", "Bearer admin-key")
	updateReq.Header.Set("Content-Type", "application/json")
	updateRR := httptest.NewRecorder()
	app.handler.ServeHTTP(updateRR, updateReq)
	if updateRR.Code != http.StatusOK {
		t.Fatalf("update status = %d, want %d; body: %s", updateRR.Code, http.StatusOK, updateRR.Body.String())
	}
	var updated model.EditorialInboxItem
	if err := json.Unmarshal(updateRR.Body.Bytes(), &updated); err != nil {
		t.Fatalf("decode updated: %v", err)
	}
	if updated.Note != "updated note" {
		t.Fatalf("updated note = %q, want %q", updated.Note, "updated note")
	}
}
func TestAPIUpdateEditorialInboxNotFound(t *testing.T) {
	app := newPublicTestApp(t)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/editorial-inbox/99999", strings.NewReader(`{"note":"nope"}`))
	req.Header.Set("Authorization", "Bearer admin-key")
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	app.handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusNotFound, rr.Body.String())
	}
	assertContains(t, rr.Body.String(), "not found")
}
func TestAPIFailEditorialInboxClaim(t *testing.T) {
	app := newPublicTestApp(t)
	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/editorial-inbox", strings.NewReader(`{"source_type":"repo","source_value":"https://github.com/example/test-fail","priority":90,"not_before":"2025-06-05T08:00:00Z","mode":"scheduled_assignment","status":"approved"}`))
	createReq.Header.Set("Authorization", "Bearer admin-key")
	createReq.Header.Set("Content-Type", "application/json")
	createRR := httptest.NewRecorder()
	app.handler.ServeHTTP(createRR, createReq)
	if createRR.Code != http.StatusCreated {
		t.Fatalf("create status = %d, want %d; body: %s", createRR.Code, http.StatusCreated, createRR.Body.String())
	}
	var created model.EditorialInboxItem
	if err := json.Unmarshal(createRR.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode created: %v", err)
	}
	claimReq := httptest.NewRequest(http.MethodPost, "/api/v1/editorial-inbox/claim", strings.NewReader(`{"worker":"test-worker"}`))
	claimReq.Header.Set("Authorization", "Bearer admin-key")
	claimReq.Header.Set("Content-Type", "application/json")
	claimRR := httptest.NewRecorder()
	app.handler.ServeHTTP(claimRR, claimReq)
	if claimRR.Code != http.StatusOK {
		t.Fatalf("claim status = %d, want %d; body: %s", claimRR.Code, http.StatusOK, claimRR.Body.String())
	}
	var claim model.EditorialInboxClaimResult
	if err := json.Unmarshal(claimRR.Body.Bytes(), &claim); err != nil {
		t.Fatalf("decode claim: %v", err)
	}
	failPayload := fmt.Sprintf(`{"claim_token":"%s","failure_note":"test failure","failure_meta":"{\"code\":500}","retryable":true}`, claim.ClaimToken)
	failReq := httptest.NewRequest(http.MethodPost, "/api/v1/editorial-inbox/"+strconv.FormatInt(created.ID, 10)+"/fail", strings.NewReader(failPayload))
	failReq.Header.Set("Authorization", "Bearer admin-key")
	failReq.Header.Set("Content-Type", "application/json")
	failRR := httptest.NewRecorder()
	app.handler.ServeHTTP(failRR, failReq)
	if failRR.Code != http.StatusOK {
		t.Fatalf("fail status = %d, want %d; body: %s", failRR.Code, http.StatusOK, failRR.Body.String())
	}
	assertContains(t, failRR.Body.String(), `"failure_note":"test failure"`)
}
func TestAPIFailEditorialInboxClaimBadToken(t *testing.T) {
	app := newPublicTestApp(t)
	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/editorial-inbox", strings.NewReader(`{"source_type":"repo","source_value":"https://github.com/example/test-bad-token","priority":90,"not_before":"2025-06-05T08:00:00Z","mode":"scheduled_assignment","status":"approved"}`))
	createReq.Header.Set("Authorization", "Bearer admin-key")
	createReq.Header.Set("Content-Type", "application/json")
	createRR := httptest.NewRecorder()
	app.handler.ServeHTTP(createRR, createReq)
	if createRR.Code != http.StatusCreated {
		t.Fatalf("create status = %d, want %d; body: %s", createRR.Code, http.StatusCreated, createRR.Body.String())
	}
	var created model.EditorialInboxItem
	if err := json.Unmarshal(createRR.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode created: %v", err)
	}
	failReq := httptest.NewRequest(http.MethodPost, "/api/v1/editorial-inbox/"+strconv.FormatInt(created.ID, 10)+"/fail", strings.NewReader(`{"claim_token":"wrong-token","failure_note":"fail","retryable":false}`))
	failReq.Header.Set("Authorization", "Bearer admin-key")
	failReq.Header.Set("Content-Type", "application/json")
	failRR := httptest.NewRecorder()
	app.handler.ServeHTTP(failRR, failReq)
	if failRR.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d; body: %s", failRR.Code, http.StatusConflict, failRR.Body.String())
	}
	assertContains(t, failRR.Body.String(), "invalid claim")
}
func TestItoa(t *testing.T) {
	tests := []struct {
		input int
		want  string
	}{
		{input: 0, want: "0"},
		{input: 42, want: "42"},
		{input: -1, want: "-1"},
		{input: 999, want: "999"},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("itoa(%d)", tt.input), func(t *testing.T) {
			got := itoa(tt.input)
			if got != tt.want {
				t.Fatalf("itoa(%d) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
func TestItoa64(t *testing.T) {
	tests := []struct {
		input int64
		want  string
	}{
		{input: 0, want: "0"},
		{input: 42, want: "42"},
		{input: -1, want: "-1"},
		{input: 9223372036854775807, want: "9223372036854775807"},
		{input: -9223372036854775808, want: "-9223372036854775808"},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("itoa64(%d)", tt.input), func(t *testing.T) {
			got := itoa64(tt.input)
			if got != tt.want {
				t.Fatalf("itoa64(%d) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
func TestAdminUpdateProjectShowsEditPageAfterUpdate(t *testing.T) {
	app := newPublicTestApp(t)
	sessionCookie, csrf := loginSession(t, app)
	form := url.Values{
		"_csrf":   {csrf},
		"title":   {"Whyred Watchtower Updated"},
		"content": {"# Updated content"},
		"excerpt": {"Updated excerpt"},
		"state":   {"active"},
		"intent":  {"update"},
	}
	req := httptest.NewRequest(http.MethodPost, "/admin/projects/update/1", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(sessionCookie)
	rr := httptest.NewRecorder()
	app.handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusSeeOther {
		t.Fatalf("update status = %d, want %d; body: %s", rr.Code, http.StatusSeeOther, rr.Body.String())
	}
	if got := rr.Header().Get("Location"); got != "/admin/projects" {
		t.Fatalf("Location = %q, want /admin/projects", got)
	}
	status, body := getBody(t, app, "/admin/projects/edit/1", sessionCookie)
	if status != http.StatusOK {
		t.Fatalf("edit status = %d, want %d; body: %s", status, http.StatusOK, body)
	}
	assertContains(t, body, "Whyred Watchtower Updated")
}
func TestAdminCreateProjectUpdateWithEventAt(t *testing.T) {
	app := newPublicTestApp(t)
	sessionCookie, csrf := loginSession(t, app)
	form := url.Values{
		"_csrf":      {csrf},
		"kind":       {"changelog"},
		"title":      {"v1.0 released"},
		"content_md": {"Released version 1.0."},
		"event_at":   {"2026-06-22T10:00"},
		"published":  {"on"},
	}
	req := httptest.NewRequest(http.MethodPost, "/admin/projects/1/updates", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(sessionCookie)
	rr := httptest.NewRecorder()
	app.handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusSeeOther, rr.Body.String())
	}
}
func TestAdminCreateProjectUpdateRequiresCSRF(t *testing.T) {
	app := newPublicTestApp(t)
	sessionCookie, _ := loginSession(t, app)
	form := url.Values{
		"kind":       {"changelog"},
		"title":      {"No CSRF"},
		"content_md": {"# No CSRF"},
	}
	req := httptest.NewRequest(http.MethodPost, "/admin/projects/1/updates", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(sessionCookie)
	rr := httptest.NewRecorder()
	app.handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusForbidden, rr.Body.String())
	}
}

// =============================================================================
// Topic/User/Category form tests
// =============================================================================

func TestAdminUpdateTopicForm(t *testing.T) {
	app := newPublicTestApp(t)
	sessionCookie, csrf := loginSession(t, app)

	cats, err := app.svc.Blog.ListCategories()
	if err != nil {
		t.Fatalf("list categories: %v", err)
	}
	var catID int64
	for _, c := range cats {
		if c.Slug == "devops" {
			catID = c.ID
			break
		}
	}
	if catID == 0 {
		t.Fatalf("seed category not found")
	}

	// Keep slug unchanged — the seed category has posts so slug cannot change.
	form := url.Values{
		"_csrf":       {csrf},
		"name":        {"DevOps Updated"},
		"slug":        {"devops"},
		"description": {"Updated description"},
	}
	req := httptest.NewRequest(http.MethodPost, "/admin/topics/update/"+strconv.FormatInt(catID, 10), strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(sessionCookie)
	rr := httptest.NewRecorder()
	app.handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusSeeOther, rr.Body.String())
	}
	if got := rr.Header().Get("Location"); got != "/admin/topics" {
		t.Fatalf("Location = %q, want /admin/topics", got)
	}

	cat, _, err := app.svc.Blog.GetCategoryByID(catID)
	if err != nil {
		t.Fatalf("get category: %v", err)
	}
	if cat.Name != "DevOps Updated" {
		t.Fatalf("category name = %q, want %q", cat.Name, "DevOps Updated")
	}
	if cat.Slug != "devops" {
		t.Fatalf("category slug = %q, want %q", cat.Slug, "devops")
	}
}

func TestAdminDeleteTopicForm(t *testing.T) {
	app := newPublicTestApp(t)
	sessionCookie, csrf := loginSession(t, app)

	cat, err := app.svc.Blog.CreateCategory(model.Category{
		Slug:        "delete-me",
		Name:        "Delete Me",
		Description: "Will be deleted.",
	})
	if err != nil {
		t.Fatalf("create category: %v", err)
	}

	form := url.Values{"_csrf": {csrf}}
	req := httptest.NewRequest(http.MethodPost, "/admin/topics/delete/"+strconv.FormatInt(cat.ID, 10), strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(sessionCookie)
	rr := httptest.NewRecorder()
	app.handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusSeeOther, rr.Body.String())
	}
	if got := rr.Header().Get("Location"); got != "/admin/topics" {
		t.Fatalf("Location = %q, want /admin/topics", got)
	}

	_, _, err = app.svc.Blog.GetCategoryByID(cat.ID)
	if err == nil {
		t.Fatalf("category should have been deleted")
	}
}

// =============================================================================
// Post deletion and intent tests
// =============================================================================

func TestAdminDeletePostForm(t *testing.T) {
	app := newPublicTestApp(t)
	sessionCookie, csrf := loginSession(t, app)

	posts, _, err := app.svc.Blog.ListAllPosts(1, 9999)
	if err != nil {
		t.Fatalf("list posts: %v", err)
	}
	var postID int64
	for _, p := range posts {
		if p.Title == "Hello Raevtar" {
			postID = p.ID
			break
		}
	}
	if postID == 0 {
		t.Fatalf("seed post not found")
	}

	form := url.Values{"_csrf": {csrf}}
	req := httptest.NewRequest(http.MethodPost, "/admin/posts/delete/"+strconv.FormatInt(postID, 10), strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(sessionCookie)
	rr := httptest.NewRecorder()
	app.handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusSeeOther, rr.Body.String())
	}
	if got := rr.Header().Get("Location"); got != "/admin/posts" {
		t.Fatalf("Location = %q, want /admin/posts", got)
	}

	_, err = app.svc.Blog.GetPostByID(postID)
	if err == nil {
		t.Fatalf("post should have been deleted")
	}
}

func TestAdminUpdatePostFormWithPublishIntent(t *testing.T) {
	app := newPublicTestApp(t)
	sessionCookie, csrf := loginSession(t, app)

	// Create a draft post first
	draftPost, err := app.svc.Blog.CreatePost(model.PostCreate{
		Title:        "Draft To Publish",
		ContentMD:    "# Draft\n\nContent to be published.",
		Excerpt:      "Draft post for publish intent test.",
		CategorySlug: "devops",
		Published:    false,
	})
	if err != nil {
		t.Fatalf("create draft post: %v", err)
	}

	form := url.Values{
		"_csrf":         {csrf},
		"title":         {"Draft To Publish"},
		"category_slug": {"devops"},
		"content":       {"# Draft\n\nUpdated content."},
		"excerpt":       {"Updated via publish intent."},
		"intent":        {"publish"},
	}
	req := httptest.NewRequest(http.MethodPost, "/admin/posts/update/"+strconv.FormatInt(draftPost.ID, 10), strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(sessionCookie)
	rr := httptest.NewRecorder()
	app.handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusSeeOther, rr.Body.String())
	}
	if got := rr.Header().Get("Location"); got != "/admin/posts" {
		t.Fatalf("Location = %q, want /admin/posts", got)
	}

	updated, err := app.svc.Blog.GetPostByID(draftPost.ID)
	if err != nil {
		t.Fatalf("get post: %v", err)
	}
	if !updated.Published {
		t.Fatalf("post should be published after publish intent")
	}
	if updated.Title != "Draft To Publish" {
		t.Fatalf("post title = %q, want %q", updated.Title, "Draft To Publish")
	}
}

func TestAdminUpdatePostFormWithDraftIntent(t *testing.T) {
	app := newPublicTestApp(t)
	sessionCookie, csrf := loginSession(t, app)

	// The seed post "Hello Raevtar" is published = true
	posts, _, err := app.svc.Blog.ListAllPosts(1, 9999)
	if err != nil {
		t.Fatalf("list posts: %v", err)
	}
	var postID int64
	for _, p := range posts {
		if p.Title == "Hello Raevtar" {
			postID = p.ID
			break
		}
	}
	if postID == 0 {
		t.Fatalf("seed post not found")
	}

	form := url.Values{
		"_csrf":         {csrf},
		"title":         {"Hello Raevtar"},
		"category_slug": {"devops"},
		"content":       {"# Hello Raevtar\n\nUnpublished via draft intent."},
		"excerpt":       {"Unpublished via draft intent."},
		"intent":        {"draft"},
	}
	req := httptest.NewRequest(http.MethodPost, "/admin/posts/update/"+strconv.FormatInt(postID, 10), strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(sessionCookie)
	rr := httptest.NewRecorder()
	app.handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusSeeOther, rr.Body.String())
	}
	if got := rr.Header().Get("Location"); got != "/admin/posts" {
		t.Fatalf("Location = %q, want /admin/posts", got)
	}

	updated, err := app.svc.Blog.GetPostByID(postID)
	if err != nil {
		t.Fatalf("get post: %v", err)
	}
	if updated.Published {
		t.Fatalf("post should be unpublished after draft intent")
	}
}

func TestAdminUpdatePostFormWithInvalidIntent(t *testing.T) {
	app := newPublicTestApp(t)
	sessionCookie, csrf := loginSession(t, app)

	// The seed post "Hello Raevtar" is published = true.
	// An invalid intent should preserve the current published state.
	posts, _, err := app.svc.Blog.ListAllPosts(1, 9999)
	if err != nil {
		t.Fatalf("list posts: %v", err)
	}
	var postID int64
	for _, p := range posts {
		if p.Title == "Hello Raevtar" {
			postID = p.ID
			break
		}
	}
	if postID == 0 {
		t.Fatalf("seed post not found")
	}

	form := url.Values{
		"_csrf":         {csrf},
		"title":         {"Hello Raevtar"},
		"category_slug": {"devops"},
		"content":       {"# Hello Raevtar\n\nInvalid intent preserves publish state."},
		"excerpt":       {"Invalid intent test."},
		"intent":        {"invalid-intent-value"},
	}
	req := httptest.NewRequest(http.MethodPost, "/admin/posts/update/"+strconv.FormatInt(postID, 10), strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(sessionCookie)
	rr := httptest.NewRecorder()
	app.handler.ServeHTTP(rr, req)
	// Invalid intent does not produce an error — it preserves current published state.
	if rr.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusSeeOther, rr.Body.String())
	}
	if got := rr.Header().Get("Location"); got != "/admin/posts" {
		t.Fatalf("Location = %q, want /admin/posts", got)
	}

	updated, err := app.svc.Blog.GetPostByID(postID)
	if err != nil {
		t.Fatalf("get post: %v", err)
	}
	if !updated.Published {
		t.Fatalf("post should remain published after invalid intent")
	}
}

// =============================================================================
// Server form tests
// =============================================================================

func TestAdminCreateServerForm(t *testing.T) {
	app := newPublicTestApp(t)
	sessionCookie, csrf := loginSession(t, app)

	form := url.Values{
		"_csrf": {csrf},
		"name":  {"test-server"},
		"host":  {"192.168.1.100"},
		"port":  {"9090"},
		"tags":  {"testing,monitoring"},
	}
	req := httptest.NewRequest(http.MethodPost, "/admin/servers", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(sessionCookie)
	rr := httptest.NewRecorder()
	app.handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusSeeOther, rr.Body.String())
	}
	if got := rr.Header().Get("Location"); got != "/admin/servers" {
		t.Fatalf("Location = %q, want /admin/servers", got)
	}

	servers, err := app.svc.Monitor.ListServers()
	if err != nil {
		t.Fatalf("list servers: %v", err)
	}
	var found bool
	for _, s := range servers {
		if s.Name == "test-server" && s.Host == "192.168.1.100" && s.Port == 9090 {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("created server not found in listing")
	}
}

func TestAdminDeleteServerForm(t *testing.T) {
	app := newPublicTestApp(t)
	sessionCookie, csrf := loginSession(t, app)

	server, err := app.svc.Monitor.CreateServer("to-delete", "10.0.0.1", 8080, "delete-me")
	if err != nil {
		t.Fatalf("create server: %v", err)
	}

	form := url.Values{"_csrf": {csrf}}
	req := httptest.NewRequest(http.MethodPost, "/admin/servers/delete/"+strconv.FormatInt(server.ID, 10), strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(sessionCookie)
	rr := httptest.NewRecorder()
	app.handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusSeeOther, rr.Body.String())
	}
	if got := rr.Header().Get("Location"); got != "/admin/servers" {
		t.Fatalf("Location = %q, want /admin/servers", got)
	}

	_, err = app.svc.Monitor.GetServer(server.ID)
	if err == nil {
		t.Fatalf("server should have been deleted")
	}
}

// =============================================================================
// Preview tests
// =============================================================================

func TestAdminPreviewPost(t *testing.T) {
	app := newPublicTestApp(t)
	sessionCookie, csrf := loginSession(t, app)

	form := url.Values{
		"_csrf":   {csrf},
		"content": {"# Post Preview\n\n- **bold**\n- _italic_"},
	}
	req := httptest.NewRequest(http.MethodPost, "/admin/posts/preview", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(sessionCookie)
	rr := httptest.NewRecorder()
	app.handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}
	body := rr.Body.String()
	assertContains(t, body, "<h1>Post Preview</h1>")
	assertContains(t, body, "<strong>bold</strong>")
	assertContains(t, body, "<em>italic</em>")

	// Blank content should return empty preview message
	form.Set("content", "   ")
	req = httptest.NewRequest(http.MethodPost, "/admin/posts/preview", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(sessionCookie)
	rr = httptest.NewRecorder()
	app.handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("blank preview status = %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}
	assertContains(t, rr.Body.String(), "Write Markdown, then preview without saving.")
}

func TestAdminPreviewPostWithContentMD(t *testing.T) {
	app := newPublicTestApp(t)
	sessionCookie, csrf := loginSession(t, app)

	// adminPreviewPost tries content first, then content_md.
	form := url.Values{
		"_csrf":      {csrf},
		"content_md": {"# Fallback Preview\n\nContent from content_md field."},
	}
	req := httptest.NewRequest(http.MethodPost, "/admin/posts/preview", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(sessionCookie)
	rr := httptest.NewRecorder()
	app.handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}
	body := rr.Body.String()
	assertContains(t, body, "<h1>Fallback Preview</h1>")
	assertContains(t, body, "Content from content_md field.")
}
