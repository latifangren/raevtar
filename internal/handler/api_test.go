package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"raevtar/internal/model"
)

func TestAPIGetPosts(t *testing.T) {
	app := newPublicTestApp(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/posts", nil)
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

	var posts []model.Post
	if err := json.NewDecoder(rr.Body).Decode(&posts); err != nil {
		t.Fatalf("decode posts: %v", err)
	}
	if len(posts) == 0 {
		t.Fatalf("expected at least 1 post, got 0")
	}
	var found bool
	for _, p := range posts {
		if p.Title == "Hello Raevtar" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("seeded post \"Hello Raevtar\" not found in response")
	}
}

func TestAPIGetPostsByCategory(t *testing.T) {
	app := newPublicTestApp(t)

	t.Run("existing category", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/posts?category=devops", nil)
		testRequestCounter++
		req.RemoteAddr = fmt.Sprintf("203.0.113.%d:1234", testRequestCounter%250+1)
		rr := httptest.NewRecorder()
		app.handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
		}
		var posts []model.Post
		if err := json.NewDecoder(rr.Body).Decode(&posts); err != nil {
			t.Fatalf("decode posts: %v", err)
		}
		if len(posts) != 1 {
			t.Fatalf("expected 1 post, got %d", len(posts))
		}
		if posts[0].Title != "Hello Raevtar" {
			t.Fatalf("post title = %q, want %q", posts[0].Title, "Hello Raevtar")
		}
	})

	t.Run("nonexistent category", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/posts?category=nonexistent", nil)
		testRequestCounter++
		req.RemoteAddr = fmt.Sprintf("203.0.113.%d:1234", testRequestCounter%250+1)
		rr := httptest.NewRecorder()
		app.handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
		}
		var posts []model.Post
		if err := json.NewDecoder(rr.Body).Decode(&posts); err != nil {
			t.Fatalf("decode posts: %v", err)
		}
		if len(posts) != 0 {
			t.Fatalf("expected 0 posts, got %d", len(posts))
		}
	})
}

func TestAPIGetServers(t *testing.T) {
	app := newPublicTestApp(t)

	t.Run("unauthorized", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/servers", nil)
		testRequestCounter++
		req.RemoteAddr = fmt.Sprintf("203.0.113.%d:1234", testRequestCounter%250+1)
		rr := httptest.NewRecorder()
		app.handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusUnauthorized, rr.Body.String())
		}
	})

	t.Run("authorized", func(t *testing.T) {
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
			t.Fatalf("expected at least 1 server")
		}
		var found bool
		for _, s := range servers {
			if s.Name == "whyred" {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("seeded server \"whyred\" not found in response")
		}
		// agent_token_hash is json:"-", should never appear in JSON body
		if strings.Contains(rr.Body.String(), "agent_token_hash") {
			t.Fatalf("response leaked agent_token_hash: %s", rr.Body.String())
		}
	})
}

func TestAPIGetProjects(t *testing.T) {
	app := newPublicTestApp(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/projects", nil)
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

	var projects []model.Project
	if err := json.NewDecoder(rr.Body).Decode(&projects); err != nil {
		t.Fatalf("decode projects: %v", err)
	}
	if len(projects) == 0 {
		t.Fatalf("expected at least 1 project, got 0")
	}
	var found bool
	for _, p := range projects {
		if p.Title == "Whyred Watchtower" {
			found = true
			if !p.Published {
				t.Fatalf("project Whyred Watchtower has published=false in API response")
			}
			break
		}
	}
	if !found {
		t.Fatalf("seeded project \"Whyred Watchtower\" not found in response")
	}
}

func TestAPICreatePost(t *testing.T) {
	app := newPublicTestApp(t)

	payload := `{"category_slug":"devops","title":"API Created Post","content_md":"# API Created Post","excerpt":"Created via API test","published":true}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/posts", strings.NewReader(payload))
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

	var post model.Post
	if err := json.NewDecoder(rr.Body).Decode(&post); err != nil {
		t.Fatalf("decode post: %v", err)
	}
	if post.ID == 0 {
		t.Fatalf("expected non-zero post ID")
	}
	if post.Title != "API Created Post" {
		t.Fatalf("post title = %q, want %q", post.Title, "API Created Post")
	}
	if post.Slug != "api-created-post" {
		t.Fatalf("post slug = %q, want %q", post.Slug, "api-created-post")
	}
	if post.Excerpt != "Created via API test" {
		t.Fatalf("post excerpt = %q, want %q", post.Excerpt, "Created via API test")
	}
	if !post.Published {
		t.Fatalf("expected published post")
	}
}

func TestAPICreatePostUnauthorized(t *testing.T) {
	app := newPublicTestApp(t)

	payload := `{"category_slug":"devops","title":"No Auth","content_md":"# No Auth"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/posts", strings.NewReader(payload))
	testRequestCounter++
	req.RemoteAddr = fmt.Sprintf("203.0.113.%d:1234", testRequestCounter%250+1)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	app.handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusUnauthorized, rr.Body.String())
	}
	if ct := rr.Header().Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
		t.Fatalf("content type = %q, want application/json", ct)
	}
	assertContains(t, rr.Body.String(), "error")
}

func TestAPICreatePostInvalidBody(t *testing.T) {
	app := newPublicTestApp(t)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/posts", strings.NewReader(`{invalid json`))
	testRequestCounter++
	req.RemoteAddr = fmt.Sprintf("203.0.113.%d:1234", testRequestCounter%250+1)
	req.Header.Set("Authorization", "Bearer admin-key")
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	app.handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusBadRequest, rr.Body.String())
	}
	if ct := rr.Header().Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
		t.Fatalf("content type = %q, want application/json", ct)
	}
	assertContains(t, rr.Body.String(), "invalid JSON")
}

func TestAPIGetPendingCommands(t *testing.T) {
	app := newPublicTestApp(t)

	// rotate agent token so we have a known token
	agentToken, err := app.svc.Monitor.RotateAgentToken(app.serverID)
	if err != nil {
		t.Fatalf("rotate agent token: %v", err)
	}

	// queue a command via service layer directly to have something pending
	queued, err := app.svc.CommandQ.QueueCommand(app.serverID, "restart", `{"reason":"test"}`)
	if err != nil {
		t.Fatalf("queue command: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/servers/"+strconv.FormatInt(app.serverID, 10)+"/commands", nil)
	testRequestCounter++
	req.RemoteAddr = fmt.Sprintf("203.0.113.%d:1234", testRequestCounter%250+1)
	req.Header.Set("Authorization", "Bearer "+agentToken)
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
	if len(cmds) == 0 {
		t.Fatalf("expected at least 1 pending command")
	}
	var found bool
	for _, c := range cmds {
		if c.ID == queued.ID && c.Command == "restart" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("queued command not found in pending list")
	}
}

func TestAPIGetPendingCommandsUnauthorized(t *testing.T) {
	app := newPublicTestApp(t)

	t.Run("no auth header", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/servers/"+strconv.FormatInt(app.serverID, 10)+"/commands", nil)
		testRequestCounter++
		req.RemoteAddr = fmt.Sprintf("203.0.113.%d:1234", testRequestCounter%250+1)
		rr := httptest.NewRecorder()
		app.handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusUnauthorized, rr.Body.String())
		}
		assertContains(t, rr.Body.String(), "missing or invalid")
	})

	t.Run("wrong token", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/servers/"+strconv.FormatInt(app.serverID, 10)+"/commands", nil)
		testRequestCounter++
		req.RemoteAddr = fmt.Sprintf("203.0.113.%d:1234", testRequestCounter%250+1)
		req.Header.Set("Authorization", "Bearer wrong-token")
		rr := httptest.NewRecorder()
		app.handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusUnauthorized, rr.Body.String())
		}
		assertContains(t, rr.Body.String(), "missing or invalid")
	})
}

func TestAPIReportCommandResult(t *testing.T) {
	app := newPublicTestApp(t)

	// rotate agent token
	agentToken, err := app.svc.Monitor.RotateAgentToken(app.serverID)
	if err != nil {
		t.Fatalf("rotate agent token: %v", err)
	}

	// queue a command
	queued, err := app.svc.CommandQ.QueueCommand(app.serverID, "restart", `{"reason":"test"}`)
	if err != nil {
		t.Fatalf("queue command: %v", err)
	}

	serverPath := "/api/v1/servers/" + strconv.FormatInt(app.serverID, 10)

	t.Run("report success", func(t *testing.T) {
		payload := fmt.Sprintf(`{"command_id":%d,"result":"restarted OK","failed":false}`, queued.ID)
		req := httptest.NewRequest(http.MethodPost, serverPath+"/commands/result", strings.NewReader(payload))
		testRequestCounter++
		req.RemoteAddr = fmt.Sprintf("203.0.113.%d:1234", testRequestCounter%250+1)
		req.Header.Set("Authorization", "Bearer "+agentToken)
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()
		app.handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
		}
		assertContains(t, rr.Body.String(), `"status":"ok"`)
	})

	t.Run("command no longer pending", func(t *testing.T) {
		// after success, the pending list should be empty for this command
		getReq := httptest.NewRequest(http.MethodGet, serverPath+"/commands", nil)
		testRequestCounter++
		getReq.RemoteAddr = fmt.Sprintf("203.0.113.%d:1234", testRequestCounter%250+1)
		getReq.Header.Set("Authorization", "Bearer "+agentToken)
		getRR := httptest.NewRecorder()
		app.handler.ServeHTTP(getRR, getReq)

		if getRR.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d; body: %s", getRR.Code, http.StatusOK, getRR.Body.String())
		}
		var cmds []model.ServerCommand
		if err := json.NewDecoder(getRR.Body).Decode(&cmds); err != nil {
			t.Fatalf("decode commands: %v", err)
		}
		for _, c := range cmds {
			if c.ID == queued.ID {
				t.Fatalf("completed command still pending: %+v", c)
			}
		}
	})

	t.Run("report failure", func(t *testing.T) {
		// queue another command
		failCmd, err := app.svc.CommandQ.QueueCommand(app.serverID, "update", `{"version":"2.0"}`)
		if err != nil {
			t.Fatalf("queue fail command: %v", err)
		}

		payload := fmt.Sprintf(`{"command_id":%d,"result":"update failed","failed":true}`, failCmd.ID)
		req := httptest.NewRequest(http.MethodPost, serverPath+"/commands/result", strings.NewReader(payload))
		testRequestCounter++
		req.RemoteAddr = fmt.Sprintf("203.0.113.%d:1234", testRequestCounter%250+1)
		req.Header.Set("Authorization", "Bearer "+agentToken)
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()
		app.handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
		}
		assertContains(t, rr.Body.String(), `"status":"ok"`)
	})

	t.Run("missing command_id", func(t *testing.T) {
		payload := `{"result":"no id","failed":false}`
		req := httptest.NewRequest(http.MethodPost, serverPath+"/commands/result", strings.NewReader(payload))
		testRequestCounter++
		req.RemoteAddr = fmt.Sprintf("203.0.113.%d:1234", testRequestCounter%250+1)
		req.Header.Set("Authorization", "Bearer "+agentToken)
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()
		app.handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusBadRequest, rr.Body.String())
		}
		assertContains(t, rr.Body.String(), "command_id required")
	})
}

func TestAPIGetPostsPagination(t *testing.T) {
	app := newPublicTestApp(t)

	// Create a second post so we have more than one
	_, err := app.svc.Blog.CreatePost(model.PostCreate{
		CategorySlug: "devops",
		Title:        "Second Post",
		ContentMD:    "# Second Post\n\nPagination fixture.",
		Excerpt:      "Second test post.",
		Published:    true,
	})
	if err != nil {
		t.Fatalf("create second post: %v", err)
	}

	t.Run("limit=1 returns all posts (handler ignores limit)", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/posts?limit=1", nil)
		testRequestCounter++
		req.RemoteAddr = fmt.Sprintf("203.0.113.%d:1234", testRequestCounter%250+1)
		rr := httptest.NewRecorder()
		app.handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
		}
		var posts []model.Post
		if err := json.NewDecoder(rr.Body).Decode(&posts); err != nil {
			t.Fatalf("decode posts: %v", err)
		}
		if len(posts) < 2 {
			t.Fatalf("expected at least 2 posts (handler ignores limit), got %d", len(posts))
		}
	})
}

func TestAPIGetPostsInvalidLimit(t *testing.T) {
	app := newPublicTestApp(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/posts?limit=-1", nil)
	testRequestCounter++
	req.RemoteAddr = fmt.Sprintf("203.0.113.%d:1234", testRequestCounter%250+1)
	rr := httptest.NewRecorder()
	app.handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}
	var posts []model.Post
	if err := json.NewDecoder(rr.Body).Decode(&posts); err != nil {
		t.Fatalf("decode posts: %v", err)
	}
	if len(posts) == 0 {
		t.Fatalf("expected posts even with invalid limit param")
	}
}

func TestAPICreatePostWithTags(t *testing.T) {
	app := newPublicTestApp(t)

	payload := `{"category_slug":"devops","title":"Tagged API Post","content_md":"# Tagged","excerpt":"Post with tags","published":true,"tags":["golang","api","testing"]}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/posts", strings.NewReader(payload))
	testRequestCounter++
	req.RemoteAddr = fmt.Sprintf("203.0.113.%d:1234", testRequestCounter%250+1)
	req.Header.Set("Authorization", "Bearer admin-key")
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	app.handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusCreated, rr.Body.String())
	}

	var post model.Post
	if err := json.NewDecoder(rr.Body).Decode(&post); err != nil {
		t.Fatalf("decode post: %v", err)
	}
	if len(post.Tags) != 3 {
		t.Fatalf("expected 3 tags, got %d: %+v", len(post.Tags), post.Tags)
	}

	tagNames := make(map[string]bool)
	for _, tag := range post.Tags {
		tagNames[tag.Name] = true
	}
	for _, want := range []string{"golang", "api", "testing"} {
		if !tagNames[want] {
			t.Fatalf("missing tag %q in response tags: %+v", want, post.Tags)
		}
	}
}
