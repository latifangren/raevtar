package service

import (
	"path/filepath"
	"testing"
	"time"

	"raevtar/internal/config"
	"raevtar/internal/model"
	"raevtar/internal/repo"
)

type testServices struct {
	svc   *Service
	repos *repo.Repositories
}

func newTestServices(t *testing.T) *testServices {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "raevtar_test.db")
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

	repos := repo.New(db)
	svc := New(repos, cfg)
	if err := svc.SeedData(); err != nil {
		t.Fatalf("seed data: %v", err)
	}

	return &testServices{svc: svc, repos: repos}
}

func TestBlogServiceListPostsFiltersByCategory(t *testing.T) {
	state := newTestServices(t)

	_, err := state.svc.Blog.CreatePost(model.PostCreate{
		CategorySlug: "devops",
		Title:        "DevOps Post",
		ContentMD:    "# DevOps Post",
		Excerpt:      "DevOps excerpt",
		Published:    true,
	})
	if err != nil {
		t.Fatalf("create devops post: %v", err)
	}

	_, err = state.svc.Blog.CreatePost(model.PostCreate{
		CategorySlug: "tools",
		Title:        "Tools Draft",
		ContentMD:    "# Tools Draft",
		Excerpt:      "Tools excerpt",
		Published:    false,
	})
	if err != nil {
		t.Fatalf("create tools post: %v", err)
	}

	posts, total, err := state.svc.Blog.ListPosts("", 1, 10)
	if err != nil {
		t.Fatalf("list all posts: %v", err)
	}
	if total != 1 {
		t.Fatalf("total posts = %d, want 1", total)
	}
	if len(posts) != 1 {
		t.Fatalf("posts len = %d, want 1", len(posts))
	}
	if posts[0].Slug != "devops-post" {
		t.Fatalf("post slug = %q, want devops-post", posts[0].Slug)
	}
	if posts[0].CategorySlug != "devops" {
		t.Fatalf("category slug = %q, want devops", posts[0].CategorySlug)
	}

	filtered, filteredTotal, err := state.svc.Blog.ListPosts("devops", 1, 10)
	if err != nil {
		t.Fatalf("list filtered posts: %v", err)
	}
	if filteredTotal != 1 {
		t.Fatalf("filtered total = %d, want 1", filteredTotal)
	}
	if len(filtered) != 1 {
		t.Fatalf("filtered len = %d, want 1", len(filtered))
	}
	if filtered[0].CategorySlug != "devops" {
		t.Fatalf("filtered category slug = %q, want devops", filtered[0].CategorySlug)
	}
	if filtered[0].Published != true {
		t.Fatalf("filtered post should be published")
	}
}

func TestBlogServicePublishedPostLookupHidesDrafts(t *testing.T) {
	state := newTestServices(t)

	draft, err := state.svc.Blog.CreatePost(model.PostCreate{
		CategorySlug: "tools",
		Title:        "Hidden Draft",
		ContentMD:    "# Hidden Draft",
		Excerpt:      "Draft excerpt",
		Published:    false,
	})
	if err != nil {
		t.Fatalf("create draft: %v", err)
	}

	if _, err := state.svc.Blog.GetPost(draft.Slug); err != nil {
		t.Fatalf("internal get draft: %v", err)
	}
	if _, err := state.svc.Blog.GetPublishedPost(draft.Slug); err == nil {
		t.Fatalf("published get draft should fail")
	}
}

func TestBlogServiceCreatePostGeneratesUniqueSlugs(t *testing.T) {
	state := newTestServices(t)

	first, err := state.svc.Blog.CreatePost(model.PostCreate{
		CategorySlug: "devops",
		Title:        "Duplicate Title",
		ContentMD:    "# First",
		Published:    true,
	})
	if err != nil {
		t.Fatalf("create first post: %v", err)
	}
	second, err := state.svc.Blog.CreatePost(model.PostCreate{
		CategorySlug: "devops",
		Title:        "Duplicate Title",
		ContentMD:    "# Second",
		Published:    true,
	})
	if err != nil {
		t.Fatalf("create second post: %v", err)
	}

	if first.Slug != "duplicate-title" {
		t.Fatalf("first slug = %q, want duplicate-title", first.Slug)
	}
	if second.Slug != "duplicate-title-2" {
		t.Fatalf("second slug = %q, want duplicate-title-2", second.Slug)
	}
}

func TestBlogServiceUpdatePostPreservesSlugAndReplacesTags(t *testing.T) {
	state := newTestServices(t)

	post, err := state.svc.Blog.CreatePost(model.PostCreate{
		CategorySlug: "devops",
		Title:        "Original Title",
		ContentMD:    "# Original",
		Published:    true,
		Tags:         []string{"old"},
	})
	if err != nil {
		t.Fatalf("create post: %v", err)
	}
	updated, err := state.svc.Blog.UpdatePost(post.ID, model.PostUpdate{
		CategorySlug: "tools",
		Title:        "Updated Title",
		ContentMD:    "# Updated",
		Excerpt:      "New excerpt",
		Published:    false,
		Tags:         []string{"new", " commissioned "},
	})
	if err != nil {
		t.Fatalf("update post: %v", err)
	}

	if updated.Slug != post.Slug {
		t.Fatalf("updated slug = %q, want preserved %q", updated.Slug, post.Slug)
	}
	if updated.Title != "Updated Title" || updated.CategorySlug != "tools" || updated.ContentMD != "# Updated" || updated.Excerpt != "New excerpt" || updated.Published {
		t.Fatalf("updated post mismatch: %+v", updated)
	}
	if len(updated.Tags) != 2 || updated.Tags[0].Name != "commissioned" || updated.Tags[1].Name != "new" {
		t.Fatalf("updated tags = %+v, want commissioned,new", updated.Tags)
	}
}

func TestBlogServiceListAllPostsIncludesDrafts(t *testing.T) {
	state := newTestServices(t)

	if _, err := state.svc.Blog.CreatePost(model.PostCreate{CategorySlug: "devops", Title: "Published", ContentMD: "# Published", Published: true}); err != nil {
		t.Fatalf("create published post: %v", err)
	}
	if _, err := state.svc.Blog.CreatePost(model.PostCreate{CategorySlug: "tools", Title: "Draft", ContentMD: "# Draft", Published: false}); err != nil {
		t.Fatalf("create draft post: %v", err)
	}
	posts, total, err := state.svc.Blog.ListAllPosts(1, 10)
	if err != nil {
		t.Fatalf("list all posts: %v", err)
	}
	if total != 2 || len(posts) != 2 {
		t.Fatalf("posts total=%d len=%d, want 2/2", total, len(posts))
	}
}

func TestMonitorServiceMetricsAndServerAccess(t *testing.T) {
	state := newTestServices(t)

	server, err := state.svc.Monitor.CreateServer("whyred", "127.0.0.1", 9100, "local")
	if err != nil {
		t.Fatalf("create server: %v", err)
	}
	if server.ID == 0 {
		t.Fatalf("expected created server id")
	}

	servers, err := state.svc.Monitor.ListServers()
	if err != nil {
		t.Fatalf("list servers: %v", err)
	}
	if len(servers) != 1 {
		t.Fatalf("servers len = %d, want 1", len(servers))
	}
	if servers[0].Name != "whyred" {
		t.Fatalf("server name = %q, want whyred", servers[0].Name)
	}

	fetched, err := state.svc.Monitor.GetServer(server.ID)
	if err != nil {
		t.Fatalf("get server: %v", err)
	}
	if fetched.Host != "127.0.0.1" || fetched.Port != 9100 {
		t.Fatalf("fetched server = %+v, want host 127.0.0.1 port 9100", fetched)
	}

	if err := state.svc.Monitor.RecordMetrics(server.ID, model.ServerMetric{
		CPUPercent:    12.5,
		RAMUsedMB:     256,
		RAMTotalMB:    1024,
		DiskUsedGB:    8.5,
		UptimeSeconds: 3600,
		Online:        true,
	}); err != nil {
		t.Fatalf("record metrics: %v", err)
	}

	metrics, err := state.svc.Monitor.GetRecentMetrics(server.ID, 20)
	if err != nil {
		t.Fatalf("get recent metrics: %v", err)
	}
	if len(metrics) != 1 {
		t.Fatalf("metrics len = %d, want 1", len(metrics))
	}
	if metrics[0].ServerID != server.ID {
		t.Fatalf("metric server id = %d, want %d", metrics[0].ServerID, server.ID)
	}
	if !metrics[0].Online {
		t.Fatalf("metric should be online")
	}
	if metrics[0].CPUPercent != 12.5 || metrics[0].RAMUsedMB != 256 || metrics[0].RAMTotalMB != 1024 {
		t.Fatalf("metric payload mismatch: %+v", metrics[0])
	}
	updated, err := state.svc.Monitor.GetServer(server.ID)
	if err != nil {
		t.Fatalf("get updated server: %v", err)
	}
	if updated.LastSeen == nil {
		t.Fatalf("last seen should be set")
	}
	if !updated.LastSeen.Equal(metrics[0].RecordedAt) {
		t.Fatalf("last seen = %s, recorded at = %s; want same timestamp", updated.LastSeen.Format(time.RFC3339Nano), metrics[0].RecordedAt.Format(time.RFC3339Nano))
	}
	if metrics[0].RecordedAt.Location() != time.UTC {
		t.Fatalf("recorded at location = %s, want UTC", metrics[0].RecordedAt.Location())
	}
	if updated.LastSeen.Location() != time.UTC {
		t.Fatalf("last seen location = %s, want UTC", updated.LastSeen.Location())
	}
}

func TestMonitorServiceRecordMetricsKeepsLastSeenCloseToLatestMetric(t *testing.T) {
	state := newTestServices(t)

	server, err := state.svc.Monitor.CreateServer("fresh", "127.0.0.1", 9100, "local")
	if err != nil {
		t.Fatalf("create server: %v", err)
	}

	before := time.Now().UTC()
	if err := state.svc.Monitor.RecordMetrics(server.ID, model.ServerMetric{Online: true}); err != nil {
		t.Fatalf("record metrics: %v", err)
	}
	after := time.Now().UTC()

	metrics, err := state.svc.Monitor.GetRecentMetrics(server.ID, 1)
	if err != nil {
		t.Fatalf("get recent metrics: %v", err)
	}
	if len(metrics) != 1 {
		t.Fatalf("metrics len = %d, want 1", len(metrics))
	}

	updated, err := state.svc.Monitor.GetServer(server.ID)
	if err != nil {
		t.Fatalf("get server: %v", err)
	}
	if updated.LastSeen == nil {
		t.Fatalf("last seen should be set")
	}
	if updated.LastSeen.Before(before) || updated.LastSeen.After(after) {
		t.Fatalf("last seen = %s, want between %s and %s", updated.LastSeen.Format(time.RFC3339Nano), before.Format(time.RFC3339Nano), after.Format(time.RFC3339Nano))
	}
	if !updated.LastSeen.Equal(metrics[0].RecordedAt) {
		t.Fatalf("last seen = %s, latest metric recorded at = %s; want same timestamp", updated.LastSeen.Format(time.RFC3339Nano), metrics[0].RecordedAt.Format(time.RFC3339Nano))
	}
}

func TestMonitorServiceRotatesAgentToken(t *testing.T) {
	state := newTestServices(t)

	server, err := state.svc.Monitor.CreateServer("agent", "127.0.0.1", 9100, "test")
	if err != nil {
		t.Fatalf("create server: %v", err)
	}

	first, err := state.svc.Monitor.RotateAgentToken(server.ID)
	if err != nil {
		t.Fatalf("rotate first token: %v", err)
	}
	if first == "" {
		t.Fatalf("expected token")
	}
	if !state.svc.Monitor.VerifyAgentToken(server.ID, first) {
		t.Fatalf("first token should verify")
	}

	second, err := state.svc.Monitor.RotateAgentToken(server.ID)
	if err != nil {
		t.Fatalf("rotate second token: %v", err)
	}
	if second == "" || second == first {
		t.Fatalf("second token = %q, first = %q", second, first)
	}
	if state.svc.Monitor.VerifyAgentToken(server.ID, first) {
		t.Fatalf("old token should be invalid")
	}
	if !state.svc.Monitor.VerifyAgentToken(server.ID, second) {
		t.Fatalf("second token should verify")
	}
}

func TestMonitorServiceRotateAgentTokenRequiresExistingServer(t *testing.T) {
	state := newTestServices(t)

	if token, err := state.svc.Monitor.RotateAgentToken(999); err == nil || token != "" {
		t.Fatalf("rotate missing server token = %q, err = %v; want error and empty token", token, err)
	}
}

func TestMonitorServiceRecordMetricsRequiresExistingServer(t *testing.T) {
	state := newTestServices(t)

	err := state.svc.Monitor.RecordMetrics(999, model.ServerMetric{Online: true})
	if err == nil {
		t.Fatalf("record metrics for missing server should fail")
	}

	metrics, err := state.svc.Monitor.GetRecentMetrics(999, 1)
	if err != nil {
		t.Fatalf("get recent metrics: %v", err)
	}
	if len(metrics) != 0 {
		t.Fatalf("metrics len = %d, want 0", len(metrics))
	}
}
