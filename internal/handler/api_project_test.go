package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"raevtar/internal/model"
)

func TestAPIListProjectUpdates(t *testing.T) {
	app := newPublicTestApp(t)

	proj, err := app.svc.Projects.GetPublishedProject("whyred-watchtower")
	if err != nil {
		t.Fatalf("get project: %v", err)
	}

	update, err := app.svc.Projects.CreateProjectUpdate(proj.ID, model.ProjectUpdateEntryCreate{
		Kind:      "timeline",
		Title:     "Initial release",
		ContentMD: "# v1.0\n\nFirst stable release.",
		Published: true,
		EventAt:   time.Now(),
	})
	if err != nil {
		t.Fatalf("create project update: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/projects/whyred-watchtower/updates", nil)
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

	var updates []model.ProjectUpdateEntry
	if err := json.NewDecoder(rr.Body).Decode(&updates); err != nil {
		t.Fatalf("decode updates: %v", err)
	}
	var found bool
	for _, u := range updates {
		if u.ID == update.ID && u.Title == "Initial release" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("created update not found in response: %+v", updates)
	}
}

func TestAPIListProjectChangelog(t *testing.T) {
	app := newPublicTestApp(t)

	proj, err := app.svc.Projects.GetPublishedProject("whyred-watchtower")
	if err != nil {
		t.Fatalf("get project: %v", err)
	}

	// changelog kind is filtered separately from timeline
	changelog, err := app.svc.Projects.CreateProjectUpdate(proj.ID, model.ProjectUpdateEntryCreate{
		Kind:      "changelog",
		Title:     "Changelog entry",
		ContentMD: "## Fixed\n\nBug resolved.",
		Published: true,
		EventAt:   time.Now(),
	})
	if err != nil {
		t.Fatalf("create changelog: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/projects/whyred-watchtower/changelog", nil)
	testRequestCounter++
	req.RemoteAddr = fmt.Sprintf("203.0.113.%d:1234", testRequestCounter%250+1)
	rr := httptest.NewRecorder()
	app.handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}

	var entries []model.ProjectUpdateEntry
	if err := json.NewDecoder(rr.Body).Decode(&entries); err != nil {
		t.Fatalf("decode changelog: %v", err)
	}
	var found bool
	for _, e := range entries {
		if e.ID == changelog.ID && e.Kind == "changelog" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("changelog entry not found in response: %+v", entries)
	}
}

func TestAPIListProjectRelations(t *testing.T) {
	app := newPublicTestApp(t)

	proj, err := app.svc.Projects.GetPublishedProject("whyred-watchtower")
	if err != nil {
		t.Fatalf("get project: %v", err)
	}

	// Get the seeded post so we can link to it
	posts, _, err := app.svc.Blog.ListPosts("", 1, 10)
	if err != nil {
		t.Fatalf("list posts: %v", err)
	}
	if len(posts) == 0 {
		t.Fatalf("no seeded posts")
	}

	rel, err := app.svc.Projects.CreateProjectRelation(proj.ID, model.ContentRelationCreate{
		TargetType:   "post",
		TargetID:     posts[0].ID,
		RelationKind: "related",
	})
	if err != nil {
		t.Fatalf("create relation: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/projects/whyred-watchtower/relations", nil)
	testRequestCounter++
	req.RemoteAddr = fmt.Sprintf("203.0.113.%d:1234", testRequestCounter%250+1)
	rr := httptest.NewRecorder()
	app.handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}

	var relations []model.ContentRelationView
	if err := json.NewDecoder(rr.Body).Decode(&relations); err != nil {
		t.Fatalf("decode relations: %v", err)
	}
	var found bool
	for _, r := range relations {
		if r.ID == rel.ID {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("created relation not found in response: %+v", relations)
	}
}

func TestAPIListProjectShowcase(t *testing.T) {
	app := newPublicTestApp(t)

	proj, err := app.svc.Projects.GetPublishedProject("whyred-watchtower")
	if err != nil {
		t.Fatalf("get project: %v", err)
	}

	item, err := app.svc.Projects.CreateProjectShowcase(proj.ID, model.ProjectShowcaseItemCreate{
		Kind:        "link",
		Title:       "GitHub Repo",
		BodyMD:      "Source code for the project.",
		ExternalURL: "https://github.com/example/watchtower",
		Published:   true,
	})
	if err != nil {
		t.Fatalf("create showcase: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/projects/whyred-watchtower/showcase", nil)
	testRequestCounter++
	req.RemoteAddr = fmt.Sprintf("203.0.113.%d:1234", testRequestCounter%250+1)
	rr := httptest.NewRecorder()
	app.handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}

	var items []model.ProjectShowcaseItem
	if err := json.NewDecoder(rr.Body).Decode(&items); err != nil {
		t.Fatalf("decode showcase: %v", err)
	}
	var found bool
	for _, s := range items {
		if s.ID == item.ID && s.Title == "GitHub Repo" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("created showcase not found in response: %+v", items)
	}
}

func TestAPICreateProjectUpdate(t *testing.T) {
	app := newPublicTestApp(t)

	proj, err := app.svc.Projects.GetPublishedProject("whyred-watchtower")
	if err != nil {
		t.Fatalf("get project: %v", err)
	}
	projectID := strconv.FormatInt(proj.ID, 10)

	payload := `{"kind":"timeline","title":"API Update","content_md":"# API Update\n\nCreated via API.","published":true}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/projects/"+projectID+"/updates", strings.NewReader(payload))
	testRequestCounter++
	req.RemoteAddr = fmt.Sprintf("203.0.113.%d:1234", testRequestCounter%250+1)
	req.Header.Set("Authorization", "Bearer admin-key")
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	app.handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusCreated, rr.Body.String())
	}
	if ct := rr.Header().Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
		t.Fatalf("content type = %q, want application/json", ct)
	}

	var update model.ProjectUpdateEntry
	if err := json.NewDecoder(rr.Body).Decode(&update); err != nil {
		t.Fatalf("decode update: %v", err)
	}
	if update.ID == 0 {
		t.Fatalf("expected non-zero update ID")
	}
	if update.Title != "API Update" {
		t.Fatalf("update title = %q, want %q", update.Title, "API Update")
	}
	if update.Kind != "timeline" {
		t.Fatalf("update kind = %q, want %q", update.Kind, "timeline")
	}
}

func TestAPIUpdateProjectUpdate(t *testing.T) {
	app := newPublicTestApp(t)

	proj, err := app.svc.Projects.GetPublishedProject("whyred-watchtower")
	if err != nil {
		t.Fatalf("get project: %v", err)
	}

	created, err := app.svc.Projects.CreateProjectUpdate(proj.ID, model.ProjectUpdateEntryCreate{
		Kind:      "build_log",
		Title:     "Original Title",
		ContentMD: "# Original content",
		Published: true,
		EventAt:   time.Now(),
	})
	if err != nil {
		t.Fatalf("create update: %v", err)
	}

	payload := `{"kind":"build_log","title":"Updated Title","content_md":"# Updated content","published":true}`
	req := httptest.NewRequest(http.MethodPut, "/api/v1/projects/updates/"+strconv.FormatInt(created.ID, 10), strings.NewReader(payload))
	testRequestCounter++
	req.RemoteAddr = fmt.Sprintf("203.0.113.%d:1234", testRequestCounter%250+1)
	req.Header.Set("Authorization", "Bearer admin-key")
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	app.handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}

	var update model.ProjectUpdateEntry
	if err := json.NewDecoder(rr.Body).Decode(&update); err != nil {
		t.Fatalf("decode update: %v", err)
	}
	if update.Title != "Updated Title" {
		t.Fatalf("update title = %q, want %q", update.Title, "Updated Title")
	}
	if update.ID != created.ID {
		t.Fatalf("update ID = %d, want %d", update.ID, created.ID)
	}
}

func TestAPIDeleteProjectUpdate(t *testing.T) {
	app := newPublicTestApp(t)

	proj, err := app.svc.Projects.GetPublishedProject("whyred-watchtower")
	if err != nil {
		t.Fatalf("get project: %v", err)
	}

	created, err := app.svc.Projects.CreateProjectUpdate(proj.ID, model.ProjectUpdateEntryCreate{
		Kind:      "build_log",
		Title:     "To Delete",
		ContentMD: "# To be deleted",
		Published: true,
		EventAt:   time.Now(),
	})
	if err != nil {
		t.Fatalf("create update: %v", err)
	}

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/projects/updates/"+strconv.FormatInt(created.ID, 10), nil)
	testRequestCounter++
	req.RemoteAddr = fmt.Sprintf("203.0.113.%d:1234", testRequestCounter%250+1)
	req.Header.Set("Authorization", "Bearer admin-key")
	rr := httptest.NewRecorder()
	app.handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}
	assertContains(t, rr.Body.String(), `"status":"ok"`)
}

func TestAPICreateProjectRelation(t *testing.T) {
	app := newPublicTestApp(t)

	proj, err := app.svc.Projects.GetPublishedProject("whyred-watchtower")
	if err != nil {
		t.Fatalf("get project: %v", err)
	}

	// Get or create a target post
	posts, _, err := app.svc.Blog.ListPosts("", 1, 10)
	if err != nil {
		t.Fatalf("list posts: %v", err)
	}
	if len(posts) == 0 {
		t.Fatalf("no seeded posts")
	}

	payload := fmt.Sprintf(`{"target_type":"post","target_id":%d,"relation_kind":"related"}`, posts[0].ID)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/projects/"+strconv.FormatInt(proj.ID, 10)+"/relations", strings.NewReader(payload))
	testRequestCounter++
	req.RemoteAddr = fmt.Sprintf("203.0.113.%d:1234", testRequestCounter%250+1)
	req.Header.Set("Authorization", "Bearer admin-key")
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	app.handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusCreated, rr.Body.String())
	}

	var rel model.ContentRelation
	if err := json.NewDecoder(rr.Body).Decode(&rel); err != nil {
		t.Fatalf("decode relation: %v", err)
	}
	if rel.ID == 0 {
		t.Fatalf("expected non-zero relation ID")
	}
	if rel.RelationKind != "related" {
		t.Fatalf("relation kind = %q, want %q", rel.RelationKind, "related")
	}
}

func TestAPIDeleteProjectRelation(t *testing.T) {
	app := newPublicTestApp(t)

	proj, err := app.svc.Projects.GetPublishedProject("whyred-watchtower")
	if err != nil {
		t.Fatalf("get project: %v", err)
	}

	posts, _, err := app.svc.Blog.ListPosts("", 1, 10)
	if err != nil {
		t.Fatalf("list posts: %v", err)
	}

	rel, err := app.svc.Projects.CreateProjectRelation(proj.ID, model.ContentRelationCreate{
		TargetType:   "post",
		TargetID:     posts[0].ID,
		RelationKind: "related",
	})
	if err != nil {
		t.Fatalf("create relation: %v", err)
	}

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/projects/relations/"+strconv.FormatInt(rel.ID, 10), nil)
	testRequestCounter++
	req.RemoteAddr = fmt.Sprintf("203.0.113.%d:1234", testRequestCounter%250+1)
	req.Header.Set("Authorization", "Bearer admin-key")
	rr := httptest.NewRecorder()
	app.handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}
	assertContains(t, rr.Body.String(), `"status":"ok"`)
}

func TestAPICreateProjectShowcase(t *testing.T) {
	app := newPublicTestApp(t)

	proj, err := app.svc.Projects.GetPublishedProject("whyred-watchtower")
	if err != nil {
		t.Fatalf("get project: %v", err)
	}

	payload := `{"kind":"image","title":"Screenshot","body_md":"## Screenshot\n\nMain dashboard view.","published":true}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/projects/"+strconv.FormatInt(proj.ID, 10)+"/showcase", strings.NewReader(payload))
	testRequestCounter++
	req.RemoteAddr = fmt.Sprintf("203.0.113.%d:1234", testRequestCounter%250+1)
	req.Header.Set("Authorization", "Bearer admin-key")
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	app.handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusCreated, rr.Body.String())
	}

	var item model.ProjectShowcaseItem
	if err := json.NewDecoder(rr.Body).Decode(&item); err != nil {
		t.Fatalf("decode showcase: %v", err)
	}
	if item.ID == 0 {
		t.Fatalf("expected non-zero showcase item ID")
	}
	if item.Title != "Screenshot" {
		t.Fatalf("showcase title = %q, want %q", item.Title, "Screenshot")
	}
	if item.Kind != "image" {
		t.Fatalf("showcase kind = %q, want %q", item.Kind, "image")
	}
}

func TestAPIUpdateProjectShowcase(t *testing.T) {
	app := newPublicTestApp(t)

	proj, err := app.svc.Projects.GetPublishedProject("whyred-watchtower")
	if err != nil {
		t.Fatalf("get project: %v", err)
	}

	created, err := app.svc.Projects.CreateProjectShowcase(proj.ID, model.ProjectShowcaseItemCreate{
		Kind:      "image",
		Title:     "Original Screenshot",
		BodyMD:    "# Original",
		Published: true,
	})
	if err != nil {
		t.Fatalf("create showcase: %v", err)
	}

	payload := `{"kind":"image","title":"Updated Screenshot","body_md":"# Updated","published":true}`
	req := httptest.NewRequest(http.MethodPut, "/api/v1/projects/showcase/"+strconv.FormatInt(created.ID, 10), strings.NewReader(payload))
	testRequestCounter++
	req.RemoteAddr = fmt.Sprintf("203.0.113.%d:1234", testRequestCounter%250+1)
	req.Header.Set("Authorization", "Bearer admin-key")
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	app.handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}

	var item model.ProjectShowcaseItem
	if err := json.NewDecoder(rr.Body).Decode(&item); err != nil {
		t.Fatalf("decode showcase: %v", err)
	}
	if item.Title != "Updated Screenshot" {
		t.Fatalf("showcase title = %q, want %q", item.Title, "Updated Screenshot")
	}
	if item.ID != created.ID {
		t.Fatalf("showcase ID = %d, want %d", item.ID, created.ID)
	}
}

func TestAPIDeleteProjectShowcase(t *testing.T) {
	app := newPublicTestApp(t)

	proj, err := app.svc.Projects.GetPublishedProject("whyred-watchtower")
	if err != nil {
		t.Fatalf("get project: %v", err)
	}

	created, err := app.svc.Projects.CreateProjectShowcase(proj.ID, model.ProjectShowcaseItemCreate{
		Kind:      "link",
		Title:     "To Delete",
		BodyMD:    "# Delete me",
		Published: true,
	})
	if err != nil {
		t.Fatalf("create showcase: %v", err)
	}

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/projects/showcase/"+strconv.FormatInt(created.ID, 10), nil)
	testRequestCounter++
	req.RemoteAddr = fmt.Sprintf("203.0.113.%d:1234", testRequestCounter%250+1)
	req.Header.Set("Authorization", "Bearer admin-key")
	rr := httptest.NewRecorder()
	app.handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}
	assertContains(t, rr.Body.String(), `"status":"ok"`)
}

func TestAPIGetServer(t *testing.T) {
	app := newPublicTestApp(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/servers/"+strconv.FormatInt(app.serverID, 10), nil)
	testRequestCounter++
	req.RemoteAddr = fmt.Sprintf("203.0.113.%d:1234", testRequestCounter%250+1)
	req.Header.Set("Authorization", "Bearer admin-key")
	rr := httptest.NewRecorder()
	app.handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}
	if ct := rr.Header().Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
		t.Fatalf("content type = %q, want application/json", ct)
	}

	var server model.Server
	if err := json.NewDecoder(rr.Body).Decode(&server); err != nil {
		t.Fatalf("decode server: %v", err)
	}
	if server.ID != app.serverID {
		t.Fatalf("server ID = %d, want %d", server.ID, app.serverID)
	}
	if server.Name != "whyred" {
		t.Fatalf("server name = %q, want %q", server.Name, "whyred")
	}
	if strings.Contains(rr.Body.String(), "agent_token_hash") {
		t.Fatalf("response leaked agent_token_hash: %s", rr.Body.String())
	}
}

func TestAPIListCategories(t *testing.T) {
	app := newPublicTestApp(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/categories", nil)
	testRequestCounter++
	req.RemoteAddr = fmt.Sprintf("203.0.113.%d:1234", testRequestCounter%250+1)
	rr := httptest.NewRecorder()
	app.handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}

	var categories []model.Category
	if err := json.NewDecoder(rr.Body).Decode(&categories); err != nil {
		t.Fatalf("decode categories: %v", err)
	}
	if len(categories) == 0 {
		t.Fatalf("expected at least 1 category, got 0")
	}
	var found bool
	for _, c := range categories {
		if c.Slug == "devops" && c.Name == "DevOps" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("seeded category \"devops\" not found in response: %+v", categories)
	}
}

func TestAPIListServers(t *testing.T) {
	app := newPublicTestApp(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/servers", nil)
	testRequestCounter++
	req.RemoteAddr = fmt.Sprintf("203.0.113.%d:1234", testRequestCounter%250+1)
	req.Header.Set("Authorization", "Bearer admin-key")
	rr := httptest.NewRecorder()
	app.handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}

	var servers []model.Server
	if err := json.NewDecoder(rr.Body).Decode(&servers); err != nil {
		t.Fatalf("decode servers: %v", err)
	}
	if len(servers) == 0 {
		t.Fatalf("expected at least 1 server, got 0")
	}
	var found bool
	for _, s := range servers {
		if s.ID == app.serverID && s.Name == "whyred" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("seeded server \"whyred\" not found in response: %+v", servers)
	}
	if strings.Contains(rr.Body.String(), "agent_token_hash") {
		t.Fatalf("response leaked agent_token_hash: %s", rr.Body.String())
	}
}

func TestAPIProjectEndpointsRequireAuth(t *testing.T) {
	app := newPublicTestApp(t)

	proj, err := app.svc.Projects.GetPublishedProject("whyred-watchtower")
	if err != nil {
		t.Fatalf("get project: %v", err)
	}
	projectID := strconv.FormatInt(proj.ID, 10)

	update, err := app.svc.Projects.CreateProjectUpdate(proj.ID, model.ProjectUpdateEntryCreate{
		Kind:      "build_log",
		Title:     "Auth Test",
		ContentMD: "# Auth test",
		Published: true,
		EventAt:   time.Now(),
	})
	if err != nil {
		t.Fatalf("create update fixture: %v", err)
	}

	showcase, err := app.svc.Projects.CreateProjectShowcase(proj.ID, model.ProjectShowcaseItemCreate{
		Kind:      "image",
		Title:     "Auth Showcase",
		BodyMD:    "# Auth showcase",
		Published: true,
	})
	if err != nil {
		t.Fatalf("create showcase fixture: %v", err)
	}

	posts, _, err := app.svc.Blog.ListPosts("", 1, 10)
	if err != nil {
		t.Fatalf("list posts: %v", err)
	}

	var relationID int64
	if len(posts) > 0 {
		rel, err := app.svc.Projects.CreateProjectRelation(proj.ID, model.ContentRelationCreate{
			TargetType:   "post",
			TargetID:     posts[0].ID,
			RelationKind: "related",
		})
		if err == nil {
			relationID = rel.ID
		}
	}

	for _, tt := range []struct {
		name   string
		method string
		path   string
	}{
		{"create update", http.MethodPost, "/api/v1/projects/" + projectID + "/updates"},
		{"update update", http.MethodPut, "/api/v1/projects/updates/" + strconv.FormatInt(update.ID, 10)},
		{"delete update", http.MethodDelete, "/api/v1/projects/updates/" + strconv.FormatInt(update.ID, 10)},
		{"create relation", http.MethodPost, "/api/v1/projects/" + projectID + "/relations"},
		{"delete relation", http.MethodDelete, "/api/v1/projects/relations/" + strconv.FormatInt(relationID, 10)},
		{"create showcase", http.MethodPost, "/api/v1/projects/" + projectID + "/showcase"},
		{"update showcase", http.MethodPut, "/api/v1/projects/showcase/" + strconv.FormatInt(showcase.ID, 10)},
		{"delete showcase", http.MethodDelete, "/api/v1/projects/showcase/" + strconv.FormatInt(showcase.ID, 10)},
	} {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(tt.method, tt.path, strings.NewReader(`{}`))
			testRequestCounter++
			r.RemoteAddr = fmt.Sprintf("203.0.113.%d:1234", testRequestCounter%250+1)
			r.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()
			app.handler.ServeHTTP(rr, r)
			if rr.Code != http.StatusUnauthorized {
				t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusUnauthorized, rr.Body.String())
			}
		})
	}
}
