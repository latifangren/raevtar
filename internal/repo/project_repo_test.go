package repo

import (
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	"raevtar/internal/model"
)

func TestProjectRepoCreateAndGetByID(t *testing.T) {
	db := InitSQLite(filepath.Join(t.TempDir(), "test.db"))
	t.Cleanup(func() { _ = db.Close() })
	AutoMigrate(db)
	repos := New(db)

	p := &model.Project{
		Title:         "My Project",
		Slug:          "my-project",
		ContentMD:     "# Project\n\nDetailed description.",
		Excerpt:       "A short excerpt about the project",
		CoverImageURL: "https://example.com/cover.jpg",
		Published:     true,
		State:         model.ProjectStateActive,
		Featured:      false,
		SortOrder:     0,
	}
	if err := repos.Project.Create(p); err != nil {
		t.Fatalf("create project: %v", err)
	}
	if p.ID == 0 {
		t.Fatal("expected project.ID to be set after Create")
	}

	loaded, err := repos.Project.GetByID(p.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if loaded == nil {
		t.Fatal("GetByID returned nil")
	}
	if loaded.ID != p.ID {
		t.Errorf("ID: got %d, want %d", loaded.ID, p.ID)
	}
	if loaded.Title != p.Title {
		t.Errorf("Title: got %q, want %q", loaded.Title, p.Title)
	}
	if loaded.Slug != p.Slug {
		t.Errorf("Slug: got %q, want %q", loaded.Slug, p.Slug)
	}
	if loaded.ContentMD != p.ContentMD {
		t.Errorf("ContentMD: got %q, want %q", loaded.ContentMD, p.ContentMD)
	}
	if loaded.Excerpt != p.Excerpt {
		t.Errorf("Excerpt: got %q, want %q", loaded.Excerpt, p.Excerpt)
	}
	if loaded.CoverImageURL != p.CoverImageURL {
		t.Errorf("CoverImageURL: got %q, want %q", loaded.CoverImageURL, p.CoverImageURL)
	}
	if loaded.Published != p.Published {
		t.Errorf("Published: got %v, want %v", loaded.Published, p.Published)
	}
	if loaded.State != p.State {
		t.Errorf("State: got %q, want %q", loaded.State, p.State)
	}
	if loaded.Featured != p.Featured {
		t.Errorf("Featured: got %v, want %v", loaded.Featured, p.Featured)
	}
	if loaded.SortOrder != p.SortOrder {
		t.Errorf("SortOrder: got %d, want %d", loaded.SortOrder, p.SortOrder)
	}
	if loaded.CreatedAt.IsZero() {
		t.Error("CreatedAt should not be zero (set by SQLite DEFAULT)")
	}
	if loaded.UpdatedAt.IsZero() {
		t.Error("UpdatedAt should not be zero (set by SQLite DEFAULT)")
	}
}

func TestProjectRepoGetBySlug(t *testing.T) {
	db := InitSQLite(filepath.Join(t.TempDir(), "test.db"))
	t.Cleanup(func() { _ = db.Close() })
	AutoMigrate(db)
	repos := New(db)

	p := &model.Project{
		Title:     "Slug-Based Project",
		Slug:      "slug-based-project",
		ContentMD: "content via slug lookup",
		Excerpt:   "excerpt via slug lookup",
		Published: true,
		State:     model.ProjectStateActive,
	}
	if err := repos.Project.Create(p); err != nil {
		t.Fatalf("create project: %v", err)
	}

	loaded, err := repos.Project.GetBySlug("slug-based-project")
	if err != nil {
		t.Fatalf("GetBySlug: %v", err)
	}
	if loaded == nil {
		t.Fatal("GetBySlug returned nil")
	}
	if loaded.ID != p.ID {
		t.Errorf("ID: got %d, want %d", loaded.ID, p.ID)
	}
	if loaded.Slug != "slug-based-project" {
		t.Errorf("Slug: got %q, want %q", loaded.Slug, "slug-based-project")
	}
	if loaded.Title != "Slug-Based Project" {
		t.Errorf("Title: got %q, want %q", loaded.Title, "Slug-Based Project")
	}
}

func TestProjectRepoGetBySlugNotFound(t *testing.T) {
	db := InitSQLite(filepath.Join(t.TempDir(), "test.db"))
	t.Cleanup(func() { _ = db.Close() })
	AutoMigrate(db)
	repos := New(db)

	loaded, err := repos.Project.GetBySlug("non-existent-project-slug")
	if err == nil {
		t.Fatal("expected error for non-existent slug")
	}
	if err != sql.ErrNoRows {
		t.Fatalf("expected sql.ErrNoRows, got %v", err)
	}
	if loaded != nil {
		t.Fatal("expected nil project for non-existent slug")
	}
}

func TestProjectRepoListAndCount(t *testing.T) {
	db := InitSQLite(filepath.Join(t.TempDir(), "test.db"))
	t.Cleanup(func() { _ = db.Close() })
	AutoMigrate(db)
	repos := New(db)

	now := time.Now().Truncate(time.Second)
	slugOrder := []string{"project-a", "project-b", "project-c"}
	for _, slug := range slugOrder {
		p := &model.Project{
			Title:     "Project " + slug,
			Slug:      slug,
			ContentMD: "content for " + slug,
			Excerpt:   "excerpt for " + slug,
			Published: true,
			State:     model.ProjectStateActive,
		}
		if err := repos.Project.Create(p); err != nil {
			t.Fatalf("create project %q: %v", slug, err)
		}
	}

	// Set distinct created_at values for deterministic ordering
	for i, slug := range slugOrder {
		if _, err := db.Exec("UPDATE projects SET created_at = ? WHERE slug = ?", now.Add(-time.Duration(i)*time.Minute), slug); err != nil {
			t.Fatalf("set created_at for %q: %v", slug, err)
		}
	}

	// List with pagination: page 1
	page1, err := repos.Project.List(ProjectListOptions{Limit: 2, Offset: 0})
	if err != nil {
		t.Fatalf("List with limit=2 offset=0: %v", err)
	}
	if len(page1) != 2 {
		t.Fatalf("expected 2 projects on page 1, got %d", len(page1))
	}

	// List with pagination: page 2
	page2, err := repos.Project.List(ProjectListOptions{Limit: 2, Offset: 2})
	if err != nil {
		t.Fatalf("List with limit=2 offset=2: %v", err)
	}
	if len(page2) != 1 {
		t.Fatalf("expected 1 project on page 2, got %d", len(page2))
	}

	// Pages should not overlap
	if page1[0].ID == page2[0].ID || page1[1].ID == page2[0].ID {
		t.Error("pages should not share projects")
	}

	// Count all (no filters)
	count, err := repos.Project.Count(false, false, "")
	if err != nil {
		t.Fatalf("Count: %v", err)
	}
	if count != 3 {
		t.Errorf("expected count 3, got %d", count)
	}
}

func TestProjectRepoListFiltersByPublished(t *testing.T) {
	db := InitSQLite(filepath.Join(t.TempDir(), "test.db"))
	t.Cleanup(func() { _ = db.Close() })
	AutoMigrate(db)
	repos := New(db)

	published1 := &model.Project{
		Title: "Published One", Slug: "published-one",
		ContentMD: "content", Excerpt: "excerpt",
		Published: true, State: model.ProjectStateActive,
	}
	published2 := &model.Project{
		Title: "Published Two", Slug: "published-two",
		ContentMD: "content", Excerpt: "excerpt",
		Published: true, State: model.ProjectStateActive,
	}
	draft := &model.Project{
		Title: "Draft Project", Slug: "draft-project",
		ContentMD: "draft content", Excerpt: "draft excerpt",
		Published: false, State: model.ProjectStatePlanning,
	}
	for _, p := range []*model.Project{published1, published2, draft} {
		if err := repos.Project.Create(p); err != nil {
			t.Fatalf("create project %q: %v", p.Slug, err)
		}
	}

	result, err := repos.Project.List(ProjectListOptions{PublishedOnly: true, Limit: 10, Offset: 0})
	if err != nil {
		t.Fatalf("List with PublishedOnly: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 published projects, got %d", len(result))
	}
	for _, p := range result {
		if !p.Published {
			t.Errorf("project %q has Published=false, expected only published", p.Slug)
		}
	}

	// Count with published filter
	count, err := repos.Project.Count(true, false, "")
	if err != nil {
		t.Fatalf("Count with publishedOnly=true: %v", err)
	}
	if count != 2 {
		t.Errorf("expected published count 2, got %d", count)
	}
}

func TestProjectRepoListFiltersByFeatured(t *testing.T) {
	db := InitSQLite(filepath.Join(t.TempDir(), "test.db"))
	t.Cleanup(func() { _ = db.Close() })
	AutoMigrate(db)
	repos := New(db)

	featured1 := &model.Project{
		Title: "Featured One", Slug: "featured-one",
		ContentMD: "content", Excerpt: "excerpt",
		Published: true, State: model.ProjectStateActive,
		Featured: true,
	}
	featured2 := &model.Project{
		Title: "Featured Two", Slug: "featured-two",
		ContentMD: "content", Excerpt: "excerpt",
		Published: true, State: model.ProjectStateActive,
		Featured: true,
	}
	normal := &model.Project{
		Title: "Normal Project", Slug: "normal-project",
		ContentMD: "content", Excerpt: "excerpt",
		Published: true, State: model.ProjectStateActive,
		Featured: false,
	}
	for _, p := range []*model.Project{featured1, featured2, normal} {
		if err := repos.Project.Create(p); err != nil {
			t.Fatalf("create project %q: %v", p.Slug, err)
		}
	}

	result, err := repos.Project.List(ProjectListOptions{FeaturedOnly: true, Limit: 10, Offset: 0})
	if err != nil {
		t.Fatalf("List with FeaturedOnly: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 featured projects, got %d", len(result))
	}
	for _, p := range result {
		if !p.Featured {
			t.Errorf("project %q has Featured=false, expected only featured", p.Slug)
		}
	}

	// Count with featured filter
	count, err := repos.Project.CountWithOptions(ProjectListOptions{FeaturedOnly: true})
	if err != nil {
		t.Fatalf("CountWithOptions with FeaturedOnly: %v", err)
	}
	if count != 2 {
		t.Errorf("expected featured count 2, got %d", count)
	}
}

func TestProjectRepoListAllNoPagination(t *testing.T) {
	db := InitSQLite(filepath.Join(t.TempDir(), "test.db"))
	t.Cleanup(func() { _ = db.Close() })
	AutoMigrate(db)
	repos := New(db)

	for i := 0; i < 5; i++ {
		p := &model.Project{
			Title:     "Project " + string(rune('A'+i)),
			Slug:      "project-" + string(rune('a'+i)),
			ContentMD: "content",
			Excerpt:   "excerpt",
			Published: true,
			State:     model.ProjectStateActive,
		}
		if err := repos.Project.Create(p); err != nil {
			t.Fatalf("create project %d: %v", i, err)
		}
	}

	result, err := repos.Project.List(ProjectListOptions{Limit: 100, Offset: 0})
	if err != nil {
		t.Fatalf("List with large limit: %v", err)
	}
	if len(result) != 5 {
		t.Fatalf("expected 5 projects, got %d", len(result))
	}
}

func TestProjectRepoUpdate(t *testing.T) {
	db := InitSQLite(filepath.Join(t.TempDir(), "test.db"))
	t.Cleanup(func() { _ = db.Close() })
	AutoMigrate(db)
	repos := New(db)

	p := &model.Project{
		Title:         "Original Title",
		Slug:          "original-title",
		ContentMD:     "Original content",
		Excerpt:       "Original excerpt",
		CoverImageURL: "https://example.com/original.jpg",
		Published:     true,
		State:         model.ProjectStateActive,
		Featured:      false,
		SortOrder:     0,
	}
	if err := repos.Project.Create(p); err != nil {
		t.Fatalf("create project: %v", err)
	}

	updatedAt := time.Now().Truncate(time.Second)
	p.Title = "Updated Title"
	p.ContentMD = "Updated content"
	p.Excerpt = "Updated excerpt"
	p.CoverImageURL = "https://example.com/updated.jpg"
	p.Published = false
	p.State = model.ProjectStatePaused
	p.Featured = true
	p.SortOrder = 5
	p.UpdatedAt = updatedAt

	if err := repos.Project.Update(p); err != nil {
		t.Fatalf("Update: %v", err)
	}

	loaded, err := repos.Project.GetByID(p.ID)
	if err != nil {
		t.Fatalf("GetByID after update: %v", err)
	}

	if loaded.Title != "Updated Title" {
		t.Errorf("Title: got %q, want %q", loaded.Title, "Updated Title")
	}
	if loaded.ContentMD != "Updated content" {
		t.Errorf("ContentMD: got %q, want %q", loaded.ContentMD, "Updated content")
	}
	if loaded.Excerpt != "Updated excerpt" {
		t.Errorf("Excerpt: got %q, want %q", loaded.Excerpt, "Updated excerpt")
	}
	if loaded.CoverImageURL != "https://example.com/updated.jpg" {
		t.Errorf("CoverImageURL: got %q, want %q", loaded.CoverImageURL, "https://example.com/updated.jpg")
	}
	if loaded.Published {
		t.Error("Published should be false after update")
	}
	if loaded.State != model.ProjectStatePaused {
		t.Errorf("State: got %q, want %q", loaded.State, model.ProjectStatePaused)
	}
	if !loaded.Featured {
		t.Error("Featured should be true after update")
	}
	if loaded.SortOrder != 5 {
		t.Errorf("SortOrder: got %d, want %d", loaded.SortOrder, 5)
	}
	if loaded.Slug != "original-title" {
		t.Errorf("Slug changed after update: got %q, want %q", loaded.Slug, "original-title")
	}
	if !loaded.UpdatedAt.Equal(updatedAt) {
		t.Errorf("UpdatedAt: got %v, want %v", loaded.UpdatedAt, updatedAt)
	}
	if loaded.CreatedAt.IsZero() {
		t.Error("CreatedAt should not be zero after update")
	}
}

func TestProjectRepoDelete(t *testing.T) {
	db := InitSQLite(filepath.Join(t.TempDir(), "test.db"))
	t.Cleanup(func() { _ = db.Close() })
	AutoMigrate(db)
	repos := New(db)

	p := &model.Project{
		Title:     "Delete Me",
		Slug:      "delete-me",
		ContentMD: "to be deleted",
		Excerpt:   "bye",
		Published: true,
		State:     model.ProjectStateActive,
	}
	if err := repos.Project.Create(p); err != nil {
		t.Fatalf("create project: %v", err)
	}

	if err := repos.Project.Delete(p.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	loaded, err := repos.Project.GetByID(p.ID)
	if err == nil {
		t.Fatal("expected error after deleting project")
	}
	if err != sql.ErrNoRows {
		t.Fatalf("expected sql.ErrNoRows, got %v", err)
	}
	if loaded != nil {
		t.Fatal("expected nil project after deletion")
	}
}

func TestProjectRepoSlugExists(t *testing.T) {
	db := InitSQLite(filepath.Join(t.TempDir(), "test.db"))
	t.Cleanup(func() { _ = db.Close() })
	AutoMigrate(db)
	repos := New(db)

	p := &model.Project{
		Title:     "Slug Check",
		Slug:      "slug-check",
		ContentMD: "content",
		Excerpt:   "excerpt",
		Published: true,
		State:     model.ProjectStateActive,
	}
	if err := repos.Project.Create(p); err != nil {
		t.Fatalf("create project: %v", err)
	}

	exists, err := repos.Project.SlugExists("slug-check")
	if err != nil {
		t.Fatalf("SlugExists for existing slug: %v", err)
	}
	if !exists {
		t.Error("SlugExists returned false for existing slug 'slug-check'")
	}

	notExists, err := repos.Project.SlugExists("non-existent-project-slug")
	if err != nil {
		t.Fatalf("SlugExists for non-existent slug: %v", err)
	}
	if notExists {
		t.Error("SlugExists returned true for non-existent slug")
	}
}
