package service

import (
	"errors"
	"math"
	"path/filepath"
	"reflect"
	"strings"
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

func requireBlogMethod(t *testing.T, state *testServices, name string) reflect.Value {
	t.Helper()
	method := reflect.ValueOf(state.svc.Blog).MethodByName(name)
	if !method.IsValid() {
		t.Fatalf("BlogService missing %s; topic/category CRUD support not wired yet", name)
	}
	return method
}

func callCreateCategory(t *testing.T, state *testServices, input model.Category) (*model.Category, error) {
	t.Helper()
	results := requireBlogMethod(t, state, "CreateCategory").Call([]reflect.Value{reflect.ValueOf(input)})
	if len(results) != 2 {
		t.Fatalf("CreateCategory result count = %d, want 2", len(results))
	}
	if !results[0].IsNil() {
		cat, ok := results[0].Interface().(*model.Category)
		if !ok {
			t.Fatalf("CreateCategory first result type = %T, want *model.Category", results[0].Interface())
		}
		if results[1].IsNil() {
			return cat, nil
		}
		return cat, results[1].Interface().(error)
	}
	if results[1].IsNil() {
		return nil, nil
	}
	return nil, results[1].Interface().(error)
}

func callUpdateCategory(t *testing.T, state *testServices, id int64, input model.Category) (*model.Category, error) {
	t.Helper()
	results := requireBlogMethod(t, state, "UpdateCategory").Call([]reflect.Value{reflect.ValueOf(id), reflect.ValueOf(input)})
	if len(results) != 2 {
		t.Fatalf("UpdateCategory result count = %d, want 2", len(results))
	}
	if !results[0].IsNil() {
		cat, ok := results[0].Interface().(*model.Category)
		if !ok {
			t.Fatalf("UpdateCategory first result type = %T, want *model.Category", results[0].Interface())
		}
		if results[1].IsNil() {
			return cat, nil
		}
		return cat, results[1].Interface().(error)
	}
	if results[1].IsNil() {
		return nil, nil
	}
	return nil, results[1].Interface().(error)
}

func callDeleteCategory(t *testing.T, state *testServices, id int64) error {
	t.Helper()
	results := requireBlogMethod(t, state, "DeleteCategory").Call([]reflect.Value{reflect.ValueOf(id)})
	if len(results) != 1 {
		t.Fatalf("DeleteCategory result count = %d, want 1", len(results))
	}
	if results[0].IsNil() {
		return nil
	}
	return results[0].Interface().(error)
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

func TestBlogServiceTopicCRUDSupportsSafeCategoryLifecycle(t *testing.T) {
	state := newTestServices(t)

	created, err := callCreateCategory(t, state, model.Category{
		Slug:        "systems",
		Name:        "Systems",
		Description: "Low-level systems notes",
	})
	if err != nil {
		t.Fatalf("create category: %v", err)
	}
	if created == nil {
		t.Fatalf("create category returned nil category")
	}
	if created.ID == 0 {
		t.Fatalf("created category id = 0, want persisted category")
	}
	if created.Slug != "systems" || created.Name != "Systems" {
		t.Fatalf("created category mismatch: %+v", created)
	}

	stored, err := state.repos.Category.GetBySlug("systems")
	if err != nil {
		t.Fatalf("get created category by slug: %v", err)
	}
	if stored.ID != created.ID {
		t.Fatalf("stored category id = %d, want %d", stored.ID, created.ID)
	}

	updated, err := callUpdateCategory(t, state, created.ID, model.Category{
		Slug:        "systems-and-ops",
		Name:        "Systems and Ops",
		Description: "Expanded operating notes",
	})
	if err != nil {
		t.Fatalf("update category: %v", err)
	}
	if updated == nil {
		t.Fatalf("update category returned nil category")
	}
	if updated.ID != created.ID {
		t.Fatalf("updated category id = %d, want %d", updated.ID, created.ID)
	}
	if updated.Slug != "systems-and-ops" || updated.Name != "Systems and Ops" || updated.Description != "Expanded operating notes" {
		t.Fatalf("updated category mismatch: %+v", updated)
	}

	if _, err := state.repos.Category.GetBySlug("systems"); err == nil {
		t.Fatalf("old category slug should no longer resolve")
	}
	stored, err = state.repos.Category.GetBySlug("systems-and-ops")
	if err != nil {
		t.Fatalf("get updated category by slug: %v", err)
	}
	if stored.Name != "Systems and Ops" || stored.Description != "Expanded operating notes" {
		t.Fatalf("stored updated category mismatch: %+v", stored)
	}

	if err := callDeleteCategory(t, state, created.ID); err != nil {
		t.Fatalf("delete category: %v", err)
	}
	if _, err := state.repos.Category.GetBySlug("systems-and-ops"); err == nil {
		t.Fatalf("deleted category should not resolve by slug")
	}
}

func TestBlogServiceTopicSlugChangeBlockedWhenPostsExist(t *testing.T) {
	state := newTestServices(t)

	category, err := state.repos.Category.GetBySlug("devops")
	if err != nil {
		t.Fatalf("get seeded category: %v", err)
	}
	post, err := state.svc.Blog.CreatePost(model.PostCreate{
		CategorySlug: category.Slug,
		Title:        "Pinned Topic Post",
		ContentMD:    "# Keep slug stable",
		Published:    true,
	})
	if err != nil {
		t.Fatalf("create post in category: %v", err)
	}

	updated, err := callUpdateCategory(t, state, category.ID, model.Category{
		Slug:        "platform-ops",
		Name:        "Platform Ops",
		Description: "Attempted rename while posts exist",
	})
	if err == nil {
		t.Fatalf("expected slug change for category with posts to fail; got %+v", updated)
	}
	if updated != nil {
		t.Fatalf("blocked slug change should not return updated category: %+v", updated)
	}
	if !strings.Contains(strings.ToLower(err.Error()), "slug") {
		t.Fatalf("err = %v, want slug-related failure", err)
	}

	reloaded, err := state.repos.Post.GetByID(post.ID)
	if err != nil {
		t.Fatalf("reload post after blocked slug change: %v", err)
	}
	if reloaded.CategorySlug != category.Slug {
		t.Fatalf("post category slug = %q, want preserved %q", reloaded.CategorySlug, category.Slug)
	}
	unchanged, err := state.repos.Category.GetBySlug(category.Slug)
	if err != nil {
		t.Fatalf("reload original category slug: %v", err)
	}
	if unchanged.ID != category.ID {
		t.Fatalf("category id after blocked slug change = %d, want %d", unchanged.ID, category.ID)
	}
}

func TestBlogServiceTopicDeleteBlockedWhenPostsExist(t *testing.T) {
	state := newTestServices(t)

	category, err := callCreateCategory(t, state, model.Category{
		Slug:        "field-notes",
		Name:        "Field Notes",
		Description: "Observations from field work",
	})
	if err != nil {
		t.Fatalf("create category: %v", err)
	}
	if _, err := state.svc.Blog.CreatePost(model.PostCreate{
		CategorySlug: category.Slug,
		Title:        "Field Note #1",
		ContentMD:    "# Field Note #1",
		Published:    true,
	}); err != nil {
		t.Fatalf("create post in created category: %v", err)
	}

	err = callDeleteCategory(t, state, category.ID)
	if err == nil {
		t.Fatalf("expected delete for category with posts to fail")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "post") {
		t.Fatalf("err = %v, want post-related safety failure", err)
	}
	if _, err := state.repos.Category.GetBySlug(category.Slug); err != nil {
		t.Fatalf("blocked delete should keep category addressable: %v", err)
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

func TestBlogServiceListPostsWithQuery(t *testing.T) {
	state := newTestServices(t)

	if _, err := state.svc.Blog.CreatePost(model.PostCreate{CategorySlug: "devops", Title: "Go Search Notes", Excerpt: "SQLite search field notes", ContentMD: "# Search\n\nquery tuning", Published: true}); err != nil {
		t.Fatalf("create search post: %v", err)
	}
	if _, err := state.svc.Blog.CreatePost(model.PostCreate{CategorySlug: "tools", Title: "Hidden Search Draft", Excerpt: "should never leak", ContentMD: "# draft", Published: false}); err != nil {
		t.Fatalf("create hidden draft: %v", err)
	}
	if _, err := state.svc.Blog.CreatePost(model.PostCreate{CategorySlug: "tools", Title: "Terminal Tricks", Excerpt: "CLI grep tricks", ContentMD: "# Terminal", Published: true}); err != nil {
		t.Fatalf("create terminal post: %v", err)
	}

	posts, total, err := state.svc.Blog.ListPostsWithOptions(BlogListOptions{Query: "search", Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("search posts: %v", err)
	}
	if total != 1 || len(posts) != 1 {
		t.Fatalf("search total/len = %d/%d, want 1/1", total, len(posts))
	}
	if posts[0].Title != "Go Search Notes" {
		t.Fatalf("search result title = %q, want Go Search Notes", posts[0].Title)
	}

	filtered, filteredTotal, err := state.svc.Blog.ListPostsWithOptions(BlogListOptions{CategorySlug: "tools", Query: "grep", Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("filtered search: %v", err)
	}
	if filteredTotal != 1 || len(filtered) != 1 {
		t.Fatalf("filtered total/len = %d/%d, want 1/1", filteredTotal, len(filtered))
	}
	if filtered[0].Title != "Terminal Tricks" {
		t.Fatalf("filtered title = %q, want Terminal Tricks", filtered[0].Title)
	}
}

func TestProjectServiceCreateAndUpdateProject(t *testing.T) {
	state := newTestServices(t)

	project, err := state.svc.Projects.CreateProject(model.ProjectCreate{
		Title:     "Raevtar Project",
		ContentMD: "# Raevtar Project\n\nBuild log.",
		Excerpt:   "Initial excerpt",
		Published: true,
		Featured:  true,
		SortOrder: 3,
		Tags:      []string{"oss", " infra "},
	})
	if err != nil {
		t.Fatalf("create project: %v", err)
	}
	if project.Slug != "raevtar-project" {
		t.Fatalf("project slug = %q, want raevtar-project", project.Slug)
	}
	if len(project.Tags) != 2 {
		t.Fatalf("project tags len = %d, want 2", len(project.Tags))
	}
	if !project.Featured || project.SortOrder != 3 {
		t.Fatalf("project featured/sort = %v/%d, want true/3", project.Featured, project.SortOrder)
	}

	updated, err := state.svc.Projects.UpdateProject(project.ID, model.ProjectUpdate{
		Title:     "Raevtar Project Updated",
		ContentMD: "# Updated",
		Excerpt:   "Updated excerpt",
		Published: false,
		Featured:  false,
		SortOrder: -9,
		Tags:      []string{"lab"},
	})
	if err != nil {
		t.Fatalf("update project: %v", err)
	}
	if updated.Slug != project.Slug {
		t.Fatalf("updated slug = %q, want preserved %q", updated.Slug, project.Slug)
	}
	if updated.Title != "Raevtar Project Updated" || updated.Excerpt != "Updated excerpt" || updated.Published {
		t.Fatalf("updated project mismatch: %+v", updated)
	}
	if updated.Featured || updated.SortOrder != 0 {
		t.Fatalf("updated featured/sort = %v/%d, want false/0", updated.Featured, updated.SortOrder)
	}
	if len(updated.Tags) != 1 || updated.Tags[0].Name != "lab" {
		t.Fatalf("updated tags = %+v, want lab", updated.Tags)
	}
	if _, err := state.svc.Projects.GetPublishedProject(project.Slug); err == nil {
		t.Fatalf("published lookup should hide draft project")
	}
}

func TestProjectServiceListProjectsOrdersFeaturedThenSortOrder(t *testing.T) {
	state := newTestServices(t)

	create := func(title string, featured bool, sortOrder int, published bool) {
		t.Helper()
		_, err := state.svc.Projects.CreateProject(model.ProjectCreate{
			Title:     title,
			ContentMD: "# " + title,
			Excerpt:   title + " excerpt",
			Published: published,
			Featured:  featured,
			SortOrder: sortOrder,
		})
		if err != nil {
			t.Fatalf("create project %s: %v", title, err)
		}
	}

	create("Later Normal", false, 9, true)
	create("Featured Second", true, 2, true)
	create("Featured First", true, 1, true)
	create("Draft Hidden", true, 0, false)

	projects, total, err := state.svc.Projects.ListProjects(1, 10, ProjectListOptions{})
	if err != nil {
		t.Fatalf("list projects: %v", err)
	}
	if total != 3 {
		t.Fatalf("total = %d, want 3", total)
	}
	if len(projects) != 3 {
		t.Fatalf("len = %d, want 3", len(projects))
	}
	if projects[0].Title != "Featured First" || projects[1].Title != "Featured Second" || projects[2].Title != "Later Normal" {
		t.Fatalf("unexpected order: %#v", []string{projects[0].Title, projects[1].Title, projects[2].Title})
	}

	featuredOnly, featuredTotal, err := state.svc.Projects.ListProjects(1, 10, ProjectListOptions{FeaturedOnly: true})
	if err != nil {
		t.Fatalf("list featured projects: %v", err)
	}
	if featuredTotal != 2 || len(featuredOnly) != 2 {
		t.Fatalf("featured total/len = %d/%d, want 2/2", featuredTotal, len(featuredOnly))
	}
	if featuredOnly[0].Title != "Featured First" || featuredOnly[1].Title != "Featured Second" {
		t.Fatalf("unexpected featured order: %#v", []string{featuredOnly[0].Title, featuredOnly[1].Title})
	}

	adminProjects, adminTotal, err := state.svc.Projects.ListAllProjects(1, 10, ProjectListOptions{})
	if err != nil {
		t.Fatalf("list all projects: %v", err)
	}
	if adminTotal != 4 || len(adminProjects) != 4 {
		t.Fatalf("admin total/len = %d/%d, want 4/4", adminTotal, len(adminProjects))
	}
}

func TestProjectServiceRejectsInvalidSort(t *testing.T) {
	state := newTestServices(t)

	_, _, err := state.svc.Projects.ListProjects(1, 10, ProjectListOptions{Sort: "random"})
	if err == nil {
		t.Fatalf("expected invalid sort error")
	}
	if !errors.Is(err, ErrInvalidProjectSort) {
		t.Fatalf("err = %v, want ErrInvalidProjectSort", err)
	}
}

func TestProjectServiceListProjectsWithQuery(t *testing.T) {
	state := newTestServices(t)

	if _, err := state.svc.Projects.CreateProject(model.ProjectCreate{Title: "Search Relay", ContentMD: "# Search Relay\n\nIndexing public notes.", Excerpt: "Search aggregator project", Published: true, State: model.ProjectStateActive, Featured: true}); err != nil {
		t.Fatalf("create search project: %v", err)
	}
	if _, err := state.svc.Projects.CreateProject(model.ProjectCreate{Title: "Paused Bot", ContentMD: "# Paused", Excerpt: "Background worker", Published: true, State: model.ProjectStatePaused}); err != nil {
		t.Fatalf("create paused project: %v", err)
	}
	if _, err := state.svc.Projects.CreateProject(model.ProjectCreate{Title: "Hidden Search Project", ContentMD: "# Hidden", Excerpt: "Should not leak", Published: false, State: model.ProjectStateActive}); err != nil {
		t.Fatalf("create hidden project: %v", err)
	}

	projects, total, err := state.svc.Projects.ListProjects(1, 10, ProjectListOptions{Query: "search"})
	if err != nil {
		t.Fatalf("search projects: %v", err)
	}
	if total != 1 || len(projects) != 1 {
		t.Fatalf("search total/len = %d/%d, want 1/1", total, len(projects))
	}
	if projects[0].Title != "Search Relay" {
		t.Fatalf("project title = %q, want Search Relay", projects[0].Title)
	}

	paused, pausedTotal, err := state.svc.Projects.ListProjects(1, 10, ProjectListOptions{Query: "worker", State: model.ProjectStatePaused})
	if err != nil {
		t.Fatalf("paused search: %v", err)
	}
	if pausedTotal != 1 || len(paused) != 1 {
		t.Fatalf("paused total/len = %d/%d, want 1/1", pausedTotal, len(paused))
	}
	if paused[0].Title != "Paused Bot" {
		t.Fatalf("paused title = %q, want Paused Bot", paused[0].Title)
	}
}

func TestProjectServiceDeleteProjectAndMissingUpdate(t *testing.T) {
	state := newTestServices(t)

	project, err := state.svc.Projects.CreateProject(model.ProjectCreate{
		Title:     "Delete Me",
		ContentMD: "# Delete Me",
		Published: true,
	})
	if err != nil {
		t.Fatalf("create project: %v", err)
	}

	if err := state.svc.Projects.DeleteProject(project.ID); err != nil {
		t.Fatalf("delete project: %v", err)
	}
	if _, err := state.svc.Projects.GetProjectByID(project.ID); err == nil {
		t.Fatalf("expected deleted project lookup to fail")
	}

	_, err = state.svc.Projects.UpdateProject(9999, model.ProjectUpdate{
		Title:     "Missing",
		ContentMD: "# Missing",
	})
	if err == nil {
		t.Fatalf("expected missing project update to fail")
	}
	if !errors.Is(err, ErrProjectNotFound) {
		t.Fatalf("err = %v, want ErrProjectNotFound", err)
	}
}

func TestPageContentServiceUpdateAndRender(t *testing.T) {
	state := newTestServices(t)

	about, err := state.svc.Pages.GetPage(model.PageKeyAbout)
	if err != nil {
		t.Fatalf("get about page: %v", err)
	}
	if about.Title == "" || about.ContentHTML == "" {
		t.Fatalf("expected seeded about page content, got %+v", about)
	}

	updated, err := state.svc.Pages.UpdatePage(model.PageContent{
		Key:       model.PageKeyContact,
		Title:     "Reach out carefully",
		Summary:   "Summary first",
		ContentMD: "# Contact\n\nUpdated body.",
	})
	if err != nil {
		t.Fatalf("update contact page: %v", err)
	}
	if updated.Title != "Reach out carefully" || updated.Summary != "Summary first" {
		t.Fatalf("updated page mismatch: %+v", updated)
	}
	if updated.ContentHTML == "" || !strings.Contains(updated.ContentHTML, "Updated body") {
		t.Fatalf("expected rendered html, got %q", updated.ContentHTML)
	}
}

func TestSearchServiceUnifiedPublicScopes(t *testing.T) {
	state := newTestServices(t)

	if _, err := state.svc.Blog.CreatePost(model.PostCreate{CategorySlug: "devops", Title: "Search Story", Excerpt: "Public search article", ContentMD: "# Story", Published: true}); err != nil {
		t.Fatalf("create post: %v", err)
	}
	if _, err := state.svc.Projects.CreateProject(model.ProjectCreate{Title: "Search Engine Board", ContentMD: "# Board", Excerpt: "Public search project", Published: true, State: model.ProjectStateActive}); err != nil {
		t.Fatalf("create project: %v", err)
	}
	if _, err := state.svc.Pages.UpdatePage(model.PageContent{Key: model.PageKeyAbout, Title: "About search", Summary: "Search summary", ContentMD: "# About\n\nSearch appears here too."}); err != nil {
		t.Fatalf("update page: %v", err)
	}

	all, err := state.svc.Search.SearchPublic(SearchOptions{Query: "search", Scope: SearchScopeAll, Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("search all: %v", err)
	}
	if all.PostCount == 0 || all.ProjectCount == 0 || all.PageCount == 0 {
		t.Fatalf("expected all scopes populated, got posts=%d projects=%d pages=%d", all.PostCount, all.ProjectCount, all.PageCount)
	}

	postsOnly, err := state.svc.Search.SearchPublic(SearchOptions{Query: "search", Scope: SearchScopePosts, Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("search posts scope: %v", err)
	}
	if postsOnly.ProjectCount != 0 || postsOnly.PageCount != 0 || postsOnly.PostCount == 0 {
		t.Fatalf("unexpected posts scope counts: %+v", postsOnly)
	}

	pagesOnly, err := state.svc.Search.SearchPublic(SearchOptions{Query: "search", Scope: SearchScopePages, Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("search pages scope: %v", err)
	}
	if pagesOnly.PageCount == 0 || len(pagesOnly.Posts) != 0 || len(pagesOnly.Projects) != 0 {
		t.Fatalf("unexpected pages scope results: %+v", pagesOnly)
	}

	empty, err := state.svc.Search.SearchPublic(SearchOptions{Query: "   ", Scope: SearchScopeAll, Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("empty search: %v", err)
	}
	if empty.Total != 0 || len(empty.Posts) != 0 || len(empty.Projects) != 0 || len(empty.Pages) != 0 {
		t.Fatalf("empty search should not dump content: %+v", empty)
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

func TestEditorialInboxServiceClaimCompleteAndRetryFlow(t *testing.T) {
	state := newTestServices(t)
	now := time.Now().UTC().Truncate(time.Minute)
	item, err := state.svc.Editorial.CreateInboxItem(model.EditorialInboxCreate{
		SourceType:  "repo",
		SourceValue: "https://github.com/example/phase3",
		Priority:    70,
		NotBefore:   now.Add(-10 * time.Minute),
		Mode:        model.EditorialModeScheduled,
		Status:      model.EditorialStatusApproved,
	})
	if err != nil {
		t.Fatalf("create item: %v", err)
	}
	claim, err := state.svc.Editorial.ClaimNextInboxItem("hermes-cron", now)
	if err != nil {
		t.Fatalf("claim item: %v", err)
	}
	if claim.Item.ID != item.ID {
		t.Fatalf("claimed item id = %d, want %d", claim.Item.ID, item.ID)
	}
	if claim.Item.Status != model.EditorialStatusRunning {
		t.Fatalf("claimed status = %q, want running", claim.Item.Status)
	}
	if claim.Item.AttemptCount != 1 {
		t.Fatalf("attempt_count = %d, want 1", claim.Item.AttemptCount)
	}
	if _, err := state.svc.Editorial.ClaimNextInboxItem("hermes-cron", now); !errors.Is(err, ErrEditorialInboxNoClaimableItem) {
		t.Fatalf("second claim err = %v, want ErrEditorialInboxNoClaimableItem", err)
	}
	failed, err := state.svc.Editorial.FailInboxItemClaim(item.ID, claim.ClaimToken, "publish timeout", `{"status":504}`, true, now)
	if err != nil {
		t.Fatalf("retryable fail: %v", err)
	}
	if failed.Status != model.EditorialStatusApproved {
		t.Fatalf("retryable status = %q, want approved", failed.Status)
	}
	if !failed.NotBefore.After(now) {
		t.Fatalf("retryable not_before = %s, want future", failed.NotBefore)
	}
	claim2, err := state.svc.Editorial.ClaimNextInboxItem("hermes-cron", failed.NotBefore)
	if err != nil {
		t.Fatalf("reclaim item: %v", err)
	}
	postID := int64(77)
	done, err := state.svc.Editorial.CompleteInboxItemClaim(item.ID, claim2.ClaimToken, postID)
	if err != nil {
		t.Fatalf("complete claim: %v", err)
	}
	if done.Status != model.EditorialStatusDone {
		t.Fatalf("done status = %q, want done", done.Status)
	}
	if done.PublishedPostID == nil || *done.PublishedPostID != postID {
		t.Fatalf("published_post_id = %v, want %d", done.PublishedPostID, postID)
	}
	if done.ClaimedBy != "" || done.LeaseExpiresAt != nil {
		t.Fatalf("done claim metadata not cleared: %+v", done)
	}
	if _, err := state.svc.Editorial.CompleteInboxItemClaim(item.ID, claim2.ClaimToken, postID); !errors.Is(err, ErrEditorialInboxInvalidClaim) {
		t.Fatalf("repeat complete err = %v, want ErrEditorialInboxInvalidClaim", err)
	}
	stale, err := state.svc.Editorial.CreateInboxItem(model.EditorialInboxCreate{
		SourceType:  "topic",
		SourceValue: "stale lock",
		Priority:    80,
		NotBefore:   now.Add(-5 * time.Minute),
		Mode:        model.EditorialModeCampaign,
		Status:      model.EditorialStatusApproved,
	})
	if err != nil {
		t.Fatalf("create stale item: %v", err)
	}
	firstClaim, err := state.svc.Editorial.ClaimNextInboxItem("worker-a", now)
	if err != nil {
		t.Fatalf("claim stale item: %v", err)
	}
	if firstClaim.Item.ID != stale.ID {
		t.Fatalf("stale claim id = %d, want %d", firstClaim.Item.ID, stale.ID)
	}
	reclaimed, err := state.svc.Editorial.ClaimNextInboxItem("worker-b", now.Add(editorialInboxLeaseTTL+time.Minute))
	if err != nil {
		t.Fatalf("reclaim after lease expiry: %v", err)
	}
	if reclaimed.Item.ID != stale.ID {
		t.Fatalf("reclaimed id = %d, want %d", reclaimed.Item.ID, stale.ID)
	}
	terminal, err := state.svc.Editorial.FailInboxItemClaim(stale.ID, reclaimed.ClaimToken, "bad source", `{"retryable":false}`, false, now.Add(editorialInboxLeaseTTL+2*time.Minute))
	if err != nil {
		t.Fatalf("terminal fail: %v", err)
	}
	if terminal.Status != model.EditorialStatusFailed {
		t.Fatalf("terminal status = %q, want failed", terminal.Status)
	}
}

func TestEditorialInboxServiceMutabilityAndDeleteRules(t *testing.T) {
	state := newTestServices(t)
	now := time.Now().UTC().Truncate(time.Minute)
	item, err := state.svc.Editorial.CreateInboxItem(model.EditorialInboxCreate{
		SourceType:  "repo",
		SourceValue: "https://github.com/example/mutable",
		Priority:    50,
		NotBefore:   now,
		Mode:        model.EditorialModeScheduled,
		Status:      model.EditorialStatusApproved,
	})
	if err != nil {
		t.Fatalf("create mutable item: %v", err)
	}
	updated, err := state.svc.Editorial.UpdateInboxItem(item.ID, model.EditorialInboxUpdate{
		SourceType:   item.SourceType,
		SourceValue:  item.SourceValue,
		CategoryHint: item.CategoryHint,
		Priority:     item.Priority,
		NotBefore:    item.NotBefore,
		Deadline:     item.Deadline,
		Note:         "updated before first execution",
		Mode:         item.Mode,
		Status:       model.EditorialStatusPaused,
	})
	if err != nil {
		t.Fatalf("update mutable item: %v", err)
	}
	if updated.Status != model.EditorialStatusPaused {
		t.Fatalf("updated status = %q, want paused", updated.Status)
	}
	deleted, err := state.svc.Editorial.DeleteInboxItem(item.ID)
	if err != nil {
		t.Fatalf("delete mutable item: %v", err)
	}
	if deleted.ID != item.ID {
		t.Fatalf("deleted id = %d, want %d", deleted.ID, item.ID)
	}
	locked, err := state.svc.Editorial.CreateInboxItem(model.EditorialInboxCreate{
		SourceType:  "repo",
		SourceValue: "https://github.com/example/locked",
		Priority:    50,
		NotBefore:   now.Add(-5 * time.Minute),
		Mode:        model.EditorialModeScheduled,
		Status:      model.EditorialStatusApproved,
	})
	if err != nil {
		t.Fatalf("create locked item: %v", err)
	}
	claim, err := state.svc.Editorial.ClaimNextInboxItem("worker-lock", now)
	if err != nil {
		t.Fatalf("claim locked item: %v", err)
	}
	if claim.Item.ID != locked.ID {
		t.Fatalf("claimed id = %d, want %d", claim.Item.ID, locked.ID)
	}
	if _, err := state.svc.Editorial.UpdateInboxItem(locked.ID, model.EditorialInboxUpdate{
		SourceType:   locked.SourceType,
		SourceValue:  locked.SourceValue,
		CategoryHint: locked.CategoryHint,
		Priority:     locked.Priority,
		NotBefore:    locked.NotBefore,
		Deadline:     locked.Deadline,
		Note:         locked.Note,
		Mode:         locked.Mode,
		Status:       model.EditorialStatusRunning,
	}); !errors.Is(err, ErrEditorialInboxImmutable) {
		t.Fatalf("update locked err = %v, want ErrEditorialInboxImmutable", err)
	}
	if _, err := state.svc.Editorial.DeleteInboxItem(locked.ID); !errors.Is(err, ErrEditorialInboxImmutable) {
		t.Fatalf("delete locked err = %v, want ErrEditorialInboxImmutable", err)
	}
	if _, err := state.svc.Editorial.FailInboxItemClaim(locked.ID, claim.ClaimToken, "retry later", `{}`, true, now); err != nil {
		t.Fatalf("retry locked item: %v", err)
	}
	if _, err := state.svc.Editorial.DeleteInboxItem(locked.ID); !errors.Is(err, ErrEditorialInboxImmutable) {
		t.Fatalf("delete retried locked err = %v, want ErrEditorialInboxImmutable", err)
	}
}

func TestEditorialInboxServicePhase4FairnessOverdueAndSummary(t *testing.T) {
	state := newTestServices(t)
	now := time.Now().UTC().Truncate(time.Minute)
	overdueDeadline := now.Add(-30 * time.Minute)
	overdue, err := state.svc.Editorial.CreateInboxItem(model.EditorialInboxCreate{
		SourceType:  "repo",
		SourceValue: "https://github.com/example/overdue",
		Priority:    10,
		NotBefore:   now.Add(-2 * time.Hour),
		Deadline:    &overdueDeadline,
		Mode:        model.EditorialModeScheduled,
		Status:      model.EditorialStatusApproved,
	})
	if err != nil {
		t.Fatalf("create overdue item: %v", err)
	}
	highPriority, err := state.svc.Editorial.CreateInboxItem(model.EditorialInboxCreate{
		SourceType:  "topic",
		SourceValue: "non overdue high priority",
		Priority:    95,
		NotBefore:   now.Add(-1 * time.Hour),
		Mode:        model.EditorialModeOpportunistic,
		Status:      model.EditorialStatusApproved,
	})
	if err != nil {
		t.Fatalf("create high priority item: %v", err)
	}
	claim, err := state.svc.Editorial.ClaimNextInboxItem("worker-overdue", now)
	if err != nil {
		t.Fatalf("claim overdue item: %v", err)
	}
	if claim.Item.ID != overdue.ID {
		t.Fatalf("claimed id = %d, want overdue id %d over high-priority id %d", claim.Item.ID, overdue.ID, highPriority.ID)
	}
	completed, err := state.svc.Editorial.CompleteInboxItemClaim(overdue.ID, claim.ClaimToken, 101)
	if err != nil {
		t.Fatalf("complete overdue item: %v", err)
	}
	if completed.CompletedAt == nil {
		t.Fatalf("completed_at should be set")
	}
	for i := 0; i < 4; i++ {
		_, err := state.svc.Editorial.CreateInboxItem(model.EditorialInboxCreate{
			SourceType:  "topic",
			SourceValue: "fairness-seed-" + time.Now().UTC().Add(time.Duration(i)*time.Second).Format(time.RFC3339Nano),
			Priority:    20,
			NotBefore:   now.Add(-10 * time.Minute),
			Mode:        model.EditorialModeSeed,
			Status:      model.EditorialStatusApproved,
		})
		if err != nil {
			t.Fatalf("create fairness item %d: %v", i, err)
		}
		fairClaim, err := state.svc.Editorial.ClaimNextInboxItem("fairness-worker", now.Add(time.Duration(i+1)*time.Minute))
		if i == 3 {
			if !errors.Is(err, ErrEditorialInboxNoClaimableItem) {
				t.Fatalf("fairness gate err = %v, want ErrEditorialInboxNoClaimableItem", err)
			}
			break
		}
		if err != nil {
			t.Fatalf("claim fairness item %d: %v", i, err)
		}
		if _, err := state.svc.Editorial.FailInboxItemClaim(fairClaim.Item.ID, fairClaim.ClaimToken, "skip to keep queue moving", `{"retryable":false}`, false, now.Add(time.Duration(i+1)*time.Minute)); err != nil {
			t.Fatalf("terminal fail fairness item %d: %v", i, err)
		}
	}
	summary, err := state.svc.Editorial.GetInboxSummary(now.Add(10 * time.Minute))
	if err != nil {
		t.Fatalf("get summary: %v", err)
	}
	if summary.Fairness.NonUrgentClaimStreak != 0 {
		t.Fatalf("fairness streak = %d, want reset to 0 after autonomous gap", summary.Fairness.NonUrgentClaimStreak)
	}
	if !summary.Fairness.AutonomousGapOpened {
		t.Fatalf("expected autonomous gap opened")
	}
	if summary.Analytics.DoneCount < 1 {
		t.Fatalf("done count = %d, want at least 1", summary.Analytics.DoneCount)
	}
	if summary.Analytics.FailedCount < 1 {
		t.Fatalf("failed count = %d, want at least 1", summary.Analytics.FailedCount)
	}
	if summary.Analytics.CompletedWithPostCount != 1 {
		t.Fatalf("completed with post count = %d, want 1", summary.Analytics.CompletedWithPostCount)
	}
	if summary.Overdue.ApprovedCount != 0 {
		t.Fatalf("approved overdue count = %d, want 0 after completion", summary.Overdue.ApprovedCount)
	}
	if summary.Overdue.CompletedCount < 1 {
		t.Fatalf("completed overdue count = %d, want at least 1", summary.Overdue.CompletedCount)
	}
	if len(summary.Analytics.ByMode) == 0 {
		t.Fatalf("expected analytics by mode")
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
