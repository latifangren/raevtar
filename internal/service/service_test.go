package service

import (
	"math"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"raevtar/internal/config"
	"raevtar/internal/model"
	"raevtar/internal/repo"
)

func setServerMetricField(t *testing.T, metric *model.ServerMetric, name string, value any) {
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

func assertServerMetricField(t *testing.T, metric model.ServerMetric, name string, want any) {
	t.Helper()
	field := reflect.ValueOf(metric).FieldByName(name)
	if !field.IsValid() {
		t.Fatalf("ServerMetric missing authorized field %s", name)
	}
	expected := reflect.ValueOf(want).Convert(field.Type()).Interface()
	if got := field.Interface(); !reflect.DeepEqual(got, expected) {
		t.Fatalf("ServerMetric.%s = %v, want %v", name, got, expected)
	}
}

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
		MediaDir:     filepath.Join(t.TempDir(), "uploads"),
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

func TestEditorialInboxServiceReadyOrderingAndValidation(t *testing.T) {
	state := newTestServices(t)
	now := time.Now().UTC().Truncate(time.Minute)
	if _, err := state.svc.Editorial.CreateInboxItem(model.EditorialInboxCreate{
		SourceType:  "repo",
		SourceValue: "https://github.com/example/one",
		Priority:    80,
		NotBefore:   now.Add(2 * time.Hour),
		Mode:        model.EditorialModeScheduled,
		Status:      model.EditorialStatusApproved,
	}); err != nil {
		t.Fatalf("create future approved: %v", err)
	}
	first, err := state.svc.Editorial.CreateInboxItem(model.EditorialInboxCreate{
		SourceType:   "repo",
		SourceValue:  "https://github.com/example/two",
		CategoryHint: "devops",
		Priority:     40,
		NotBefore:    now.Add(-1 * time.Hour),
		Mode:         model.EditorialModeOpportunistic,
		Status:       model.EditorialStatusApproved,
	})
	if err != nil {
		t.Fatalf("create ready item: %v", err)
	}
	_, err = state.svc.Editorial.CreateInboxItem(model.EditorialInboxCreate{
		SourceType:  "repo",
		SourceValue: "https://github.com/example/three",
		Priority:    90,
		NotBefore:   now.Add(-1 * time.Hour),
		Mode:        model.EditorialModeScheduled,
		Status:      model.EditorialStatusPaused,
	})
	if err != nil {
		t.Fatalf("create paused item: %v", err)
	}
	second, err := state.svc.Editorial.CreateInboxItem(model.EditorialInboxCreate{
		SourceType:  "topic",
		SourceValue: "agent telemetry",
		Priority:    90,
		NotBefore:   now.Add(-30 * time.Minute),
		Mode:        model.EditorialModeCampaign,
		Status:      model.EditorialStatusApproved,
	})
	if err != nil {
		t.Fatalf("create high priority ready item: %v", err)
	}
	ready, err := state.svc.Editorial.ListReadyInboxItems(now, 10)
	if err != nil {
		t.Fatalf("list ready: %v", err)
	}
	if len(ready) != 2 {
		t.Fatalf("ready len = %d, want 2", len(ready))
	}
	if ready[0].ID != second.ID || ready[1].ID != first.ID {
		t.Fatalf("ready ordering ids = [%d %d], want [%d %d]", ready[0].ID, ready[1].ID, second.ID, first.ID)
	}
	if _, err := state.svc.Editorial.CreateInboxItem(model.EditorialInboxCreate{}); err == nil {
		t.Fatalf("expected invalid editorial input error")
	}
	postID := int64(42)
	updated, err := state.svc.Editorial.UpdateInboxItem(second.ID, model.EditorialInboxUpdate{
		SourceType:      second.SourceType,
		SourceValue:     second.SourceValue,
		CategoryHint:    second.CategoryHint,
		Priority:        second.Priority,
		NotBefore:       second.NotBefore,
		Deadline:        second.Deadline,
		Note:            second.Note,
		Mode:            second.Mode,
		Status:          model.EditorialStatusDone,
		PublishedPostID: &postID,
	})
	if err != nil {
		t.Fatalf("update done item: %v", err)
	}
	if updated.PublishedPostID == nil || *updated.PublishedPostID != postID {
		t.Fatalf("published post id = %v, want %d", updated.PublishedPostID, postID)
	}
	if _, err := state.svc.Editorial.UpdateInboxItem(first.ID, model.EditorialInboxUpdate{
		SourceType:   first.SourceType,
		SourceValue:  first.SourceValue,
		CategoryHint: first.CategoryHint,
		Priority:     first.Priority,
		NotBefore:    first.NotBefore,
		Deadline:     first.Deadline,
		Note:         first.Note,
		Mode:         first.Mode,
		Status:       model.EditorialStatusDone,
	}); err == nil {
		t.Fatalf("expected done item without published_post_id to fail")
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

func TestMonitorServiceExtendedMetricsRoundTrip(t *testing.T) {
	state := newTestServices(t)
	server, err := state.svc.Monitor.CreateServer("health-node", "127.0.0.1", 9100, "local")
	if err != nil {
		t.Fatalf("create server: %v", err)
	}

	metric := model.ServerMetric{
		CPUPercent:    25.5,
		RAMUsedMB:     512,
		RAMTotalMB:    2048,
		DiskUsedGB:    12.5,
		UptimeSeconds: 3600,
		Online:        true,
	}
	for name, value := range map[string]any{"CPULoad1": 0.4, "CPULoad5": 0.3, "CPULoad15": 0.2, "CPUCores": int64(4), "DiskTotalGB": 32.0, "TemperatureC": 48.5, "TemperatureAvailable": true} {
		setServerMetricField(t, &metric, name, value)
	}

	if err := state.svc.Monitor.RecordMetrics(server.ID, metric); err != nil {
		t.Fatalf("record metrics: %v", err)
	}
	metrics, err := state.svc.Monitor.GetRecentMetrics(server.ID, 1)
	if err != nil {
		t.Fatalf("get metrics: %v", err)
	}
	if len(metrics) != 1 {
		t.Fatalf("metrics len = %d, want 1", len(metrics))
	}
	for name, want := range map[string]any{"CPULoad1": 0.4, "CPULoad5": 0.3, "CPULoad15": 0.2, "CPUCores": int64(4), "DiskTotalGB": 32.0, "TemperatureC": 48.5, "TemperatureAvailable": true} {
		assertServerMetricField(t, metrics[0], name, want)
	}
}

func TestMonitorServiceRejectsInvalidMetrics(t *testing.T) {
	state := newTestServices(t)
	server, err := state.svc.Monitor.CreateServer("invalid-health", "127.0.0.1", 9100, "local")
	if err != nil {
		t.Fatalf("create server: %v", err)
	}

	tests := []struct {
		name   string
		metric model.ServerMetric
	}{
		{name: "negative CPU", metric: model.ServerMetric{CPUPercent: -0.1, Online: true}},
		{name: "CPU over 100", metric: model.ServerMetric{CPUPercent: 100.1, Online: true}},
		{name: "NaN CPU", metric: model.ServerMetric{CPUPercent: math.NaN(), Online: true}},
		{name: "infinite RAM", metric: model.ServerMetric{RAMUsedMB: math.Inf(1), Online: true}},
		{name: "RAM used exceeds total", metric: model.ServerMetric{RAMUsedMB: 1025, RAMTotalMB: 1024, Online: true}},
		{name: "negative RAM total", metric: model.ServerMetric{RAMTotalMB: -0.1, Online: true}},
		{name: "negative disk", metric: model.ServerMetric{DiskUsedGB: -0.1, Online: true}},
		{name: "disk used exceeds total", metric: model.ServerMetric{DiskUsedGB: 10.1, DiskTotalGB: 10, Online: true}},
		{name: "available temperature too low", metric: model.ServerMetric{TemperatureC: -50.1, TemperatureAvailable: true, Online: true}},
		{name: "available temperature too high", metric: model.ServerMetric{TemperatureC: 150.1, TemperatureAvailable: true, Online: true}},
		{name: "negative uptime", metric: model.ServerMetric{UptimeSeconds: -1, Online: true}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := state.svc.Monitor.RecordMetrics(server.ID, tt.metric); err == nil {
				t.Fatalf("RecordMetrics(%+v) succeeded, want validation error", tt.metric)
			}
		})
	}
}

func TestMonitorServiceRejectsInvalidExtendedMetrics(t *testing.T) {
	state := newTestServices(t)
	server, err := state.svc.Monitor.CreateServer("invalid-extended-health", "127.0.0.1", 9100, "local")
	if err != nil {
		t.Fatalf("create server: %v", err)
	}

	tests := []struct {
		name  string
		field string
		value any
	}{
		{name: "negative load", field: "CPULoad1", value: -0.1},
		{name: "NaN load", field: "CPULoad5", value: math.NaN()},
		{name: "infinite load", field: "CPULoad15", value: math.Inf(1)},
		{name: "negative cores", field: "CPUCores", value: int64(-1)},
		{name: "negative disk total", field: "DiskTotalGB", value: -0.1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metric := model.ServerMetric{Online: true}
			setServerMetricField(t, &metric, tt.field, tt.value)
			if err := state.svc.Monitor.RecordMetrics(server.ID, metric); err == nil {
				t.Fatalf("RecordMetrics(%s=%v) succeeded, want validation error", tt.field, tt.value)
			}
		})
	}
}
