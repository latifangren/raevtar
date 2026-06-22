package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"raevtar/internal/model"
	"raevtar/internal/service"
)

func mustMarshal(t *testing.T, v any) []byte {
	t.Helper()
	data, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	return data
}

// ---------------------------------------------------------------------------
// Section 4: API Server Endpoints — truly uncovered server tests
// ---------------------------------------------------------------------------

func TestAPICreateServer(t *testing.T) {
	app := newPublicTestApp(t)

	payload := mustMarshal(t, map[string]any{
		"name": "new-node",
		"host": "10.0.0.50",
		"port": 2200,
		"tags": "dmz,monitor",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/servers", bytes.NewReader(payload))
	req.Header.Set("Authorization", "Bearer admin-key")
	req.Header.Set("Content-Type", "application/json")
	testRequestCounter++
	req.RemoteAddr = fmt.Sprintf("203.0.113.%d:1234", testRequestCounter%250+1)
	rr := httptest.NewRecorder()
	app.handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusCreated, rr.Body.String())
	}
	if ct := rr.Header().Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
		t.Fatalf("content type = %q, want application/json", ct)
	}

	var resp struct {
		Server     model.Server `json:"server"`
		AgentToken string       `json:"agent_token"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Server.ID == 0 {
		t.Fatalf("expected non-zero server id")
	}
	if resp.Server.Name != "new-node" {
		t.Fatalf("server name = %q, want %q", resp.Server.Name, "new-node")
	}
	if resp.AgentToken == "" {
		t.Fatalf("expected agent token in response")
	}
	if resp.Server.AgentTokenHash != "" {
		t.Fatalf("server JSON leaked agent_token_hash")
	}
}

func TestAPIGetPendingCommandsViaAgentToken(t *testing.T) {
	app := newPublicTestApp(t)

	// Create a pending command via the service layer
	cmd, err := app.svc.CommandQ.QueueCommand(app.serverID, "PING", "")
	if err != nil {
		t.Fatalf("enqueue command: %v", err)
	}

	// Verify agent token works
	token, err := app.svc.Monitor.RotateAgentToken(app.serverID)
	if err != nil {
		t.Fatalf("rotate token: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/servers/"+strconv.FormatInt(app.serverID, 10)+"/commands", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	testRequestCounter++
	req.RemoteAddr = fmt.Sprintf("203.0.113.%d:1234", testRequestCounter%250+1)
	rr := httptest.NewRecorder()
	app.handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}
	if ct := rr.Header().Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
		t.Fatalf("content type = %q, want application/json", ct)
	}

	var cmds []model.ServerCommand
	if err := json.NewDecoder(rr.Body).Decode(&cmds); err != nil {
		t.Fatalf("decode commands: %v", err)
	}
	var found bool
	for _, c := range cmds {
		if c.ID == cmd.ID && c.Command == "PING" && c.Status == model.CommandPending {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("pending command not found in response: %+v", cmds)
	}
}

func TestAPIGetPendingCommandsRequiresAgentToken(t *testing.T) {
	app := newPublicTestApp(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/servers/"+strconv.FormatInt(app.serverID, 10)+"/commands", nil)
	testRequestCounter++
	req.RemoteAddr = fmt.Sprintf("203.0.113.%d:1234", testRequestCounter%250+1)
	rr := httptest.NewRecorder()
	app.handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusUnauthorized, rr.Body.String())
	}
	assertContains(t, rr.Body.String(), "missing or invalid agent token")
}

func TestAPIGetPendingCommandsAdminKeyWorks(t *testing.T) {
	app := newPublicTestApp(t)

	_, err := app.svc.CommandQ.QueueCommand(app.serverID, "REBOOT", "")
	if err != nil {
		t.Fatalf("enqueue command: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/servers/"+strconv.FormatInt(app.serverID, 10)+"/commands", nil)
	req.Header.Set("Authorization", "Bearer admin-key")
	testRequestCounter++
	req.RemoteAddr = fmt.Sprintf("203.0.113.%d:1234", testRequestCounter%250+1)
	rr := httptest.NewRecorder()
	app.handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}
	var cmds []model.ServerCommand
	if err := json.NewDecoder(rr.Body).Decode(&cmds); err != nil {
		t.Fatalf("decode commands: %v", err)
	}
	if len(cmds) == 0 {
		t.Fatalf("expected at least 1 pending command")
	}
}

func TestAPIGetPendingCommandsReturnsEmptyArrayWhenNone(t *testing.T) {
	app := newPublicTestApp(t)

	token, err := app.svc.Monitor.RotateAgentToken(app.serverID)
	if err != nil {
		t.Fatalf("rotate token: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/servers/"+strconv.FormatInt(app.serverID, 10)+"/commands", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	testRequestCounter++
	req.RemoteAddr = fmt.Sprintf("203.0.113.%d:1234", testRequestCounter%250+1)
	rr := httptest.NewRecorder()
	app.handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}
	// Should return empty array, not null
	assertContains(t, rr.Body.String(), "[")
	assertContains(t, rr.Body.String(), "]")
}

func TestAPIReportCommandResultViaAgentToken(t *testing.T) {
	app := newPublicTestApp(t)

	token, err := app.svc.Monitor.RotateAgentToken(app.serverID)
	if err != nil {
		t.Fatalf("rotate token: %v", err)
	}

	cmd, err := app.svc.CommandQ.QueueCommand(app.serverID, "PING", "")
	if err != nil {
		t.Fatalf("enqueue command: %v", err)
	}

	payload := mustMarshal(t, map[string]any{
		"command_id": cmd.ID,
		"result":     `{"latency_ms":42}`,
		"failed":     false,
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/servers/"+strconv.FormatInt(app.serverID, 10)+"/commands/result", bytes.NewReader(payload))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	testRequestCounter++
	req.RemoteAddr = fmt.Sprintf("203.0.113.%d:1234", testRequestCounter%250+1)
	rr := httptest.NewRecorder()
	app.handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}
	assertContains(t, rr.Body.String(), `"status":"ok"`)

	// Verify command is now completed
	cmds, err := app.svc.CommandQ.PendingCommands(app.serverID)
	if err != nil {
		t.Fatalf("pending commands: %v", err)
	}
	for _, c := range cmds {
		if c.ID == cmd.ID {
			t.Fatalf("command should no longer be pending: %+v", c)
		}
	}
}

func TestAPIReportCommandResultFailed(t *testing.T) {
	app := newPublicTestApp(t)

	token, err := app.svc.Monitor.RotateAgentToken(app.serverID)
	if err != nil {
		t.Fatalf("rotate token: %v", err)
	}

	cmd, err := app.svc.CommandQ.QueueCommand(app.serverID, "CLEAR_CACHE", "")
	if err != nil {
		t.Fatalf("enqueue command: %v", err)
	}

	payload := mustMarshal(t, map[string]any{
		"command_id": cmd.ID,
		"result":     "cache clear failed: timeout",
		"failed":     true,
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/servers/"+strconv.FormatInt(app.serverID, 10)+"/commands/result", bytes.NewReader(payload))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	testRequestCounter++
	req.RemoteAddr = fmt.Sprintf("203.0.113.%d:1234", testRequestCounter%250+1)
	rr := httptest.NewRecorder()
	app.handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}
	assertContains(t, rr.Body.String(), `"status":"ok"`)
}

func TestAPIReportCommandResultMissingCommandID(t *testing.T) {
	app := newPublicTestApp(t)

	token, err := app.svc.Monitor.RotateAgentToken(app.serverID)
	if err != nil {
		t.Fatalf("rotate token: %v", err)
	}

	payload := mustMarshal(t, map[string]any{
		"result": "no command id",
		"failed": false,
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/servers/"+strconv.FormatInt(app.serverID, 10)+"/commands/result", bytes.NewReader(payload))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	testRequestCounter++
	req.RemoteAddr = fmt.Sprintf("203.0.113.%d:1234", testRequestCounter%250+1)
	rr := httptest.NewRecorder()
	app.handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusBadRequest, rr.Body.String())
	}
	assertContains(t, rr.Body.String(), "command_id required")
}

// ---------------------------------------------------------------------------
// Section 6: Admin Editorial Form — adminDeleteEditorialInbox
// ---------------------------------------------------------------------------

func TestAdminDeleteEditorialInbox(t *testing.T) {
	app := newPublicTestApp(t)
	sessionCookie, csrfToken := loginAsAdmin(t, app)

	// Create an item first via the form
	createForm := url.Values{
		"source_value":  {"https://github.com/example/admin-delete-test"},
		"category_hint": {"devops"},
		"note":          {"Item to be deleted via admin form."},
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

	// Find the item ID
	items, err := app.svc.Editorial.ListInboxItems(service.EditorialInboxListFilter{})
	if err != nil {
		t.Fatalf("list items: %v", err)
	}
	var itemID int64
	for _, item := range items {
		if item.SourceValue == "https://github.com/example/admin-delete-test" {
			itemID = item.ID
			break
		}
	}
	if itemID == 0 {
		t.Fatalf("expected to find created item")
	}

	// Delete via admin form
	deleteForm := url.Values{"_csrf": {csrfToken}}
	deleteReq := httptest.NewRequest(http.MethodPost, "/admin/editorial-inbox/delete/"+strconv.FormatInt(itemID, 10), strings.NewReader(deleteForm.Encode()))
	deleteReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	deleteReq.AddCookie(sessionCookie)
	deleteRR := httptest.NewRecorder()
	app.handler.ServeHTTP(deleteRR, deleteReq)

	if deleteRR.Code != http.StatusSeeOther {
		t.Fatalf("delete status = %d, want %d; body: %s", deleteRR.Code, http.StatusSeeOther, deleteRR.Body.String())
	}
	if got := deleteRR.Header().Get("Location"); got != "/admin/editorial-inbox" {
		t.Fatalf("Location = %q, want /admin/editorial-inbox", got)
	}

	// Verify deleted
	items, err = app.svc.Editorial.ListInboxItems(service.EditorialInboxListFilter{})
	if err != nil {
		t.Fatalf("list items after delete: %v", err)
	}
	for _, item := range items {
		if item.ID == itemID {
			t.Fatalf("deleted item still present: %+v", item)
		}
	}
}

func TestAdminDeleteEditorialInboxNotFound(t *testing.T) {
	app := newPublicTestApp(t)
	sessionCookie, csrfToken := loginAsAdmin(t, app)

	form := url.Values{"_csrf": {csrfToken}}
	req := httptest.NewRequest(http.MethodPost, "/admin/editorial-inbox/delete/99999", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(sessionCookie)
	rr := httptest.NewRecorder()
	app.handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusNotFound, rr.Body.String())
	}
	assertContains(t, rr.Body.String(), "not found")
}

func TestAdminDeleteEditorialInboxRequiresCSRF(t *testing.T) {
	app := newPublicTestApp(t)
	sessionCookie, _ := loginAsAdmin(t, app)

	// Create item to delete
	item, err := app.svc.Editorial.CreateInboxItem(model.EditorialInboxCreate{
		SourceType:  "repo",
		SourceValue: "https://github.com/example/csrf-delete-test",
		Priority:    50,
		NotBefore:   time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
		Mode:        model.EditorialModeScheduled,
		Status:      model.EditorialStatusApproved,
	})
	if err != nil {
		t.Fatalf("create inbox item: %v", err)
	}

	// Delete without CSRF
	req := httptest.NewRequest(http.MethodPost, "/admin/editorial-inbox/delete/"+strconv.FormatInt(item.ID, 10), strings.NewReader(url.Values{}.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(sessionCookie)
	rr := httptest.NewRecorder()
	app.handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusForbidden, rr.Body.String())
	}
	assertContains(t, rr.Body.String(), "invalid CSRF token")
}

func TestAdminDeleteEditorialInboxGETRejected(t *testing.T) {
	app := newPublicTestApp(t)
	sessionCookie, _ := loginAsAdmin(t, app)

	req := httptest.NewRequest(http.MethodGet, "/admin/editorial-inbox/delete/1", nil)
	req.AddCookie(sessionCookie)
	rr := httptest.NewRecorder()
	app.handler.ServeHTTP(rr, req)

	// Chi returns 405 for POST-only routes when GET is used; handler body is not reached
	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusMethodNotAllowed, rr.Body.String())
	}
}

// ---------------------------------------------------------------------------
// Project updates: additional edge cases not yet covered
// ---------------------------------------------------------------------------

func TestAPIProjectUpdatesNotFound(t *testing.T) {
	app := newPublicTestApp(t)

	// GET updates for non-existent project
	req := httptest.NewRequest(http.MethodGet, "/api/v1/projects/non-existent-slug/updates", nil)
	testRequestCounter++
	req.RemoteAddr = fmt.Sprintf("203.0.113.%d:1234", testRequestCounter%250+1)
	rr := httptest.NewRecorder()
	app.handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusNotFound, rr.Body.String())
	}
	assertContains(t, rr.Body.String(), "project not found")
}

func TestAPIProjectChangelogNotFound(t *testing.T) {
	app := newPublicTestApp(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/projects/non-existent-slug/changelog", nil)
	testRequestCounter++
	req.RemoteAddr = fmt.Sprintf("203.0.113.%d:1234", testRequestCounter%250+1)
	rr := httptest.NewRecorder()
	app.handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusNotFound, rr.Body.String())
	}
	assertContains(t, rr.Body.String(), "project not found")
}

func TestAPIProjectRelationsNotFound(t *testing.T) {
	app := newPublicTestApp(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/projects/non-existent-slug/relations", nil)
	testRequestCounter++
	req.RemoteAddr = fmt.Sprintf("203.0.113.%d:1234", testRequestCounter%250+1)
	rr := httptest.NewRecorder()
	app.handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusNotFound, rr.Body.String())
	}
	assertContains(t, rr.Body.String(), "project not found")
}

func TestAPIProjectShowcaseNotFound(t *testing.T) {
	app := newPublicTestApp(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/projects/non-existent-slug/showcase", nil)
	testRequestCounter++
	req.RemoteAddr = fmt.Sprintf("203.0.113.%d:1234", testRequestCounter%250+1)
	rr := httptest.NewRecorder()
	app.handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusNotFound, rr.Body.String())
	}
	assertContains(t, rr.Body.String(), "project not found")
}

func TestAPICreateProjectUpdateInvalidJSON(t *testing.T) {
	app := newPublicTestApp(t)

	proj, err := app.svc.Projects.GetPublishedProject("whyred-watchtower")
	if err != nil {
		t.Fatalf("get project: %v", err)
	}
	projectID := strconv.FormatInt(proj.ID, 10)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/projects/"+projectID+"/updates", strings.NewReader(`not json`))
	req.Header.Set("Authorization", "Bearer admin-key")
	req.Header.Set("Content-Type", "application/json")
	testRequestCounter++
	req.RemoteAddr = fmt.Sprintf("203.0.113.%d:1234", testRequestCounter%250+1)
	rr := httptest.NewRecorder()
	app.handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusBadRequest, rr.Body.String())
	}
	assertContains(t, rr.Body.String(), "invalid JSON")
}

func TestAPIProjectUpdateInvalidID(t *testing.T) {
	app := newPublicTestApp(t)

	// PUT with non-numeric updateID
	req := httptest.NewRequest(http.MethodPut, "/api/v1/projects/updates/abc", strings.NewReader(`{}`))
	req.Header.Set("Authorization", "Bearer admin-key")
	req.Header.Set("Content-Type", "application/json")
	testRequestCounter++
	req.RemoteAddr = fmt.Sprintf("203.0.113.%d:1234", testRequestCounter%250+1)
	rr := httptest.NewRecorder()
	app.handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusBadRequest, rr.Body.String())
	}
	assertContains(t, rr.Body.String(), "invalid id")
}

func TestAPIDeleteProjectUpdateNotFound(t *testing.T) {
	app := newPublicTestApp(t)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/projects/updates/99999", nil)
	req.Header.Set("Authorization", "Bearer admin-key")
	testRequestCounter++
	req.RemoteAddr = fmt.Sprintf("203.0.113.%d:1234", testRequestCounter%250+1)
	rr := httptest.NewRecorder()
	app.handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusNotFound, rr.Body.String())
	}
}

func TestAPIServerNotFound(t *testing.T) {
	app := newPublicTestApp(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/servers/99999", nil)
	req.Header.Set("Authorization", "Bearer admin-key")
	testRequestCounter++
	req.RemoteAddr = fmt.Sprintf("203.0.113.%d:1234", testRequestCounter%250+1)
	rr := httptest.NewRecorder()
	app.handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusNotFound, rr.Body.String())
	}
}

// ---------------------------------------------------------------------------
// Editorial inbox additional edge cases
// ---------------------------------------------------------------------------

func TestAPIEditorialInboxAuthRequired(t *testing.T) {
	app := newPublicTestApp(t)

	for _, path := range []string{
		"/api/v1/editorial-inbox/summary",
		"/api/v1/editorial-inbox",
		"/api/v1/editorial-inbox/claim",
	} {
		t.Run(path, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, path, nil)
			testRequestCounter++
			req.RemoteAddr = fmt.Sprintf("203.0.113.%d:1234", testRequestCounter%250+1)
			rr := httptest.NewRecorder()
			app.handler.ServeHTTP(rr, req)

			if rr.Code != http.StatusUnauthorized {
				t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusUnauthorized, rr.Body.String())
			}
		})
	}
}

func TestAPICreateEditorialInboxInvalidJSON(t *testing.T) {
	app := newPublicTestApp(t)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/editorial-inbox", strings.NewReader(`not json`))
	req.Header.Set("Authorization", "Bearer admin-key")
	req.Header.Set("Content-Type", "application/json")
	testRequestCounter++
	req.RemoteAddr = fmt.Sprintf("203.0.113.%d:1234", testRequestCounter%250+1)
	rr := httptest.NewRecorder()
	app.handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusBadRequest, rr.Body.String())
	}
	assertContains(t, rr.Body.String(), "invalid JSON")
}

func TestAPIClaimEditorialInboxNoContent(t *testing.T) {
	app := newPublicTestApp(t)

	// Try claiming when no items exist
	req := httptest.NewRequest(http.MethodPost, "/api/v1/editorial-inbox/claim", strings.NewReader(`{"worker":"test-worker"}`))
	req.Header.Set("Authorization", "Bearer admin-key")
	req.Header.Set("Content-Type", "application/json")
	testRequestCounter++
	req.RemoteAddr = fmt.Sprintf("203.0.113.%d:1234", testRequestCounter%250+1)
	rr := httptest.NewRecorder()
	app.handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusNoContent, rr.Body.String())
	}
}

func TestAPICompleteEditorialInboxInvalidClaim(t *testing.T) {
	app := newPublicTestApp(t)

	// Create item
	created := createEditorialInboxItemViaAPI(t, app, "https://github.com/example/bad-claim-complete")

	// Try completing without claiming
	payload := mustMarshal(t, model.EditorialInboxCompleteRequest{
		ClaimToken:      "wrong-token",
		PublishedPostID: 1,
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/editorial-inbox/"+strconv.FormatInt(created.ID, 10)+"/complete", bytes.NewReader(payload))
	req.Header.Set("Authorization", "Bearer admin-key")
	req.Header.Set("Content-Type", "application/json")
	testRequestCounter++
	req.RemoteAddr = fmt.Sprintf("203.0.113.%d:1234", testRequestCounter%250+1)
	rr := httptest.NewRecorder()
	app.handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusConflict, rr.Body.String())
	}
}

func TestAPIFailEditorialInboxClaimNotFound(t *testing.T) {
	app := newPublicTestApp(t)

	payload := mustMarshal(t, model.EditorialInboxFailRequest{
		ClaimToken:  "some-token",
		FailureNote: "not found test",
		Retryable:   false,
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/editorial-inbox/99999/fail", bytes.NewReader(payload))
	req.Header.Set("Authorization", "Bearer admin-key")
	req.Header.Set("Content-Type", "application/json")
	testRequestCounter++
	req.RemoteAddr = fmt.Sprintf("203.0.113.%d:1234", testRequestCounter%250+1)
	rr := httptest.NewRecorder()
	app.handler.ServeHTTP(rr, req)

	// Repo returns ok=false for non-existent items, mapped to ErrEditorialInboxInvalidClaim → 409
	if rr.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusConflict, rr.Body.String())
	}
	assertContains(t, rr.Body.String(), "invalid claim")
}

// createEditorialInboxItemViaAPI creates an inbox item via the API and returns it.
func createEditorialInboxItemViaAPI(t *testing.T, app *publicTestApp, sourceValue string) *model.EditorialInboxItem {
	t.Helper()
	payload := mustMarshal(t, model.EditorialInboxCreate{
		SourceType:  "repo",
		SourceValue: sourceValue,
		Priority:    50,
		NotBefore:   time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
		Mode:        model.EditorialModeScheduled,
		Status:      model.EditorialStatusApproved,
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/editorial-inbox", bytes.NewReader(payload))
	req.Header.Set("Authorization", "Bearer admin-key")
	req.Header.Set("Content-Type", "application/json")
	testRequestCounter++
	req.RemoteAddr = fmt.Sprintf("203.0.113.%d:1234", testRequestCounter%250+1)
	rr := httptest.NewRecorder()
	app.handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("create editorial via API status = %d, want %d; body: %s", rr.Code, http.StatusCreated, rr.Body.String())
	}
	var item model.EditorialInboxItem
	if err := json.NewDecoder(rr.Body).Decode(&item); err != nil {
		t.Fatalf("decode created item: %v", err)
	}
	if item.ID == 0 {
		t.Fatalf("expected non-zero item id")
	}
	return &item
}
