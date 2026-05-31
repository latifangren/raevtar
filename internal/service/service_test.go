package service

import (
	"path/filepath"
	"testing"

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
}
