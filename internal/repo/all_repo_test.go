package repo

import (
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	"raevtar/internal/model"
)

// ---------------------------------------------------------------------------
// page_content_repo.go — 7 tests
// ---------------------------------------------------------------------------

func TestPageContentRepoList(t *testing.T) {
	db := InitSQLite(filepath.Join(t.TempDir(), "test.db"))
	t.Cleanup(func() { _ = db.Close() })
	AutoMigrate(db)
	repos := New(db)

	pages := []*model.PageContent{
		{Key: "about", Title: "About Us", Summary: "About page", ContentMD: "# About"},
		{Key: "contact", Title: "Contact", Summary: "Contact page", ContentMD: "# Contact"},
		{Key: "privacy", Title: "Privacy Policy", Summary: "Privacy", ContentMD: "# Privacy"},
	}
	for _, p := range pages {
		if err := repos.PageContent.Upsert(p); err != nil {
			t.Fatalf("upsert page %q: %v", p.Key, err)
		}
	}

	result, err := repos.PageContent.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(result) != 3 {
		t.Fatalf("expected 3 pages, got %d", len(result))
	}

	// Ordered by key ASC: about, contact, privacy
	expected := []string{"about", "contact", "privacy"}
	for i, key := range expected {
		if result[i].Key != key {
			t.Errorf("position %d: expected key %q, got %q", i, key, result[i].Key)
		}
	}
}

func TestPageContentRepoListEmpty(t *testing.T) {
	db := InitSQLite(filepath.Join(t.TempDir(), "test.db"))
	t.Cleanup(func() { _ = db.Close() })
	AutoMigrate(db)
	repos := New(db)

	result, err := repos.PageContent.List()
	if err != nil {
		t.Fatalf("List on empty table: %v", err)
	}
	if len(result) != 0 {
		t.Fatalf("expected 0 pages, got %d", len(result))
	}
}

func TestPageContentRepoGetByKey(t *testing.T) {
	db := InitSQLite(filepath.Join(t.TempDir(), "test.db"))
	t.Cleanup(func() { _ = db.Close() })
	AutoMigrate(db)
	repos := New(db)

	page := &model.PageContent{Key: "about", Title: "About Us", Summary: "About page summary", ContentMD: "# About\nContent here"}
	if err := repos.PageContent.Upsert(page); err != nil {
		t.Fatalf("upsert page: %v", err)
	}

	loaded, err := repos.PageContent.GetByKey("about")
	if err != nil {
		t.Fatalf("GetByKey: %v", err)
	}
	if loaded == nil {
		t.Fatal("GetByKey returned nil")
	}
	if loaded.Key != "about" {
		t.Errorf("Key: got %q, want %q", loaded.Key, "about")
	}
	if loaded.Title != "About Us" {
		t.Errorf("Title: got %q, want %q", loaded.Title, "About Us")
	}
	if loaded.Summary != "About page summary" {
		t.Errorf("Summary: got %q, want %q", loaded.Summary, "About page summary")
	}
	if loaded.ContentMD != "# About\nContent here" {
		t.Errorf("ContentMD: got %q, want %q", loaded.ContentMD, "# About\nContent here")
	}
	if loaded.UpdatedAt.IsZero() {
		t.Error("UpdatedAt should not be zero")
	}
}

func TestPageContentRepoGetByKeyNotFound(t *testing.T) {
	db := InitSQLite(filepath.Join(t.TempDir(), "test.db"))
	t.Cleanup(func() { _ = db.Close() })
	AutoMigrate(db)
	repos := New(db)

	loaded, err := repos.PageContent.GetByKey("nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent key")
	}
	if err != sql.ErrNoRows {
		t.Fatalf("expected sql.ErrNoRows, got %v", err)
	}
	if loaded != nil {
		t.Fatal("expected nil page for nonexistent key")
	}
}

func TestPageContentRepoUpsertCreate(t *testing.T) {
	db := InitSQLite(filepath.Join(t.TempDir(), "test.db"))
	t.Cleanup(func() { _ = db.Close() })
	AutoMigrate(db)
	repos := New(db)

	page := &model.PageContent{Key: "new-page", Title: "New Page", Summary: "Freshly created", ContentMD: "# New"}
	if err := repos.PageContent.Upsert(page); err != nil {
		t.Fatalf("upsert new page: %v", err)
	}

	loaded, err := repos.PageContent.GetByKey("new-page")
	if err != nil {
		t.Fatalf("GetByKey after upsert create: %v", err)
	}
	if loaded.Title != "New Page" {
		t.Errorf("Title: got %q, want %q", loaded.Title, "New Page")
	}
}

func TestPageContentRepoUpsertUpdate(t *testing.T) {
	db := InitSQLite(filepath.Join(t.TempDir(), "test.db"))
	t.Cleanup(func() { _ = db.Close() })
	AutoMigrate(db)
	repos := New(db)

	page := &model.PageContent{Key: "updatable", Title: "Original Title", Summary: "Original summary", ContentMD: "# Original"}
	if err := repos.PageContent.Upsert(page); err != nil {
		t.Fatalf("upsert original: %v", err)
	}

	page.Title = "Updated Title"
	page.Summary = "Updated summary"
	page.ContentMD = "# Updated"
	if err := repos.PageContent.Upsert(page); err != nil {
		t.Fatalf("upsert update: %v", err)
	}

	loaded, err := repos.PageContent.GetByKey("updatable")
	if err != nil {
		t.Fatalf("GetByKey after upsert update: %v", err)
	}
	if loaded.Title != "Updated Title" {
		t.Errorf("Title: got %q, want %q", loaded.Title, "Updated Title")
	}
	if loaded.Summary != "Updated summary" {
		t.Errorf("Summary: got %q, want %q", loaded.Summary, "Updated summary")
	}
	if loaded.ContentMD != "# Updated" {
		t.Errorf("ContentMD: got %q, want %q", loaded.ContentMD, "# Updated")
	}
}

func TestPageContentRepoSearch(t *testing.T) {
	db := InitSQLite(filepath.Join(t.TempDir(), "test.db"))
	t.Cleanup(func() { _ = db.Close() })
	AutoMigrate(db)
	repos := New(db)

	pages := []*model.PageContent{
		{Key: "about", Title: "About Us", Summary: "Company about page", ContentMD: "# About the company"},
		{Key: "contact", Title: "Contact", Summary: "Get in touch", ContentMD: "# Contact us"},
		{Key: "privacy", Title: "Privacy Policy", Summary: "Data handling", ContentMD: "# Privacy details"},
	}
	for _, p := range pages {
		if err := repos.PageContent.Upsert(p); err != nil {
			t.Fatalf("upsert page %q: %v", p.Key, err)
		}
	}

	opts := PageContentSearchOptions{
		Keys:   []string{"about", "contact", "privacy"},
		Query:  "company",
		Limit:  10,
		Offset: 0,
	}
	result, err := repos.PageContent.Search(opts)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(result) == 0 {
		t.Fatal("expected at least one result for query 'company'")
	}
	for _, p := range result {
		if p.Key != "about" {
			t.Errorf("expected only 'about' page to match 'company', got %q", p.Key)
		}
	}
}

// ---------------------------------------------------------------------------
// content_relation_repo.go — 3 tests
// ---------------------------------------------------------------------------

func TestContentRelationCreate(t *testing.T) {
	db := InitSQLite(filepath.Join(t.TempDir(), "test.db"))
	t.Cleanup(func() { _ = db.Close() })
	AutoMigrate(db)
	repos := New(db)

	rel := &model.ContentRelation{
		SourceType:   "project",
		SourceID:     1,
		TargetType:   "project",
		TargetID:     2,
		RelationKind: "related",
		SortOrder:    0,
	}
	if err := repos.ContentRelation.Create(rel); err != nil {
		t.Fatalf("create content relation: %v", err)
	}
	if rel.ID == 0 {
		t.Fatal("expected relation ID to be set after Create")
	}
}

func TestContentRelationListBySource(t *testing.T) {
	db := InitSQLite(filepath.Join(t.TempDir(), "test.db"))
	t.Cleanup(func() { _ = db.Close() })
	AutoMigrate(db)
	repos := New(db)

	rels := []*model.ContentRelation{
		{SourceType: "project", SourceID: 10, TargetType: "project", TargetID: 20, RelationKind: "related", SortOrder: 0},
		{SourceType: "project", SourceID: 10, TargetType: "project", TargetID: 30, RelationKind: "builds_on", SortOrder: 1},
	}
	for _, r := range rels {
		if err := repos.ContentRelation.Create(r); err != nil {
			t.Fatalf("create relation: %v", err)
		}
	}

	result, err := repos.ContentRelation.ListBySource("project", 10)
	if err != nil {
		t.Fatalf("ListBySource: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 relations, got %d", len(result))
	}
	if result[0].TargetID != 20 {
		t.Errorf("first relation TargetID: got %d, want %d", result[0].TargetID, int64(20))
	}
	if result[1].TargetID != 30 {
		t.Errorf("second relation TargetID: got %d, want %d", result[1].TargetID, int64(30))
	}
}

func TestContentRelationDelete(t *testing.T) {
	db := InitSQLite(filepath.Join(t.TempDir(), "test.db"))
	t.Cleanup(func() { _ = db.Close() })
	AutoMigrate(db)
	repos := New(db)

	rel := &model.ContentRelation{
		SourceType:   "project",
		SourceID:     1,
		TargetType:   "project",
		TargetID:     2,
		RelationKind: "related",
		SortOrder:    0,
	}
	if err := repos.ContentRelation.Create(rel); err != nil {
		t.Fatalf("create relation: %v", err)
	}

	if err := repos.ContentRelation.Delete(rel.ID); err != nil {
		t.Fatalf("delete relation: %v", err)
	}

	result, err := repos.ContentRelation.ListBySource("project", 1)
	if err != nil {
		t.Fatalf("ListBySource after delete: %v", err)
	}
	if len(result) != 0 {
		t.Fatalf("expected 0 relations after delete, got %d", len(result))
	}
}

// ---------------------------------------------------------------------------
// project_update_repo.go — 6 tests
// ---------------------------------------------------------------------------

func TestProjectUpdateCreate(t *testing.T) {
	db := InitSQLite(filepath.Join(t.TempDir(), "test.db"))
	t.Cleanup(func() { _ = db.Close() })
	AutoMigrate(db)
	repos := New(db)

	now := time.Now().Truncate(time.Millisecond)
	item := &model.ProjectUpdateEntry{
		ProjectID: 1,
		Kind:      "timeline",
		Title:     "Initial Release",
		ContentMD: "Launched v1.0",
		Published: true,
		Pinned:    false,
		SortOrder: 0,
		EventAt:   now,
	}
	if err := repos.ProjectUpdate.Create(item); err != nil {
		t.Fatalf("create project update: %v", err)
	}
	if item.ID == 0 {
		t.Fatal("expected update ID to be set after Create")
	}
	if item.ProjectID != 1 {
		t.Errorf("ProjectID: got %d, want %d", item.ProjectID, int64(1))
	}
	if item.Title != "Initial Release" {
		t.Errorf("Title: got %q, want %q", item.Title, "Initial Release")
	}
}

func TestProjectUpdateGetByID(t *testing.T) {
	db := InitSQLite(filepath.Join(t.TempDir(), "test.db"))
	t.Cleanup(func() { _ = db.Close() })
	AutoMigrate(db)
	repos := New(db)

	now := time.Now().Truncate(time.Millisecond)
	item := &model.ProjectUpdateEntry{
		ProjectID: 42,
		Kind:      "changelog",
		Title:     "v2.0 Changelog",
		ContentMD: "Added many features",
		Published: true,
		Pinned:    true,
		SortOrder: 1,
		EventAt:   now,
	}
	if err := repos.ProjectUpdate.Create(item); err != nil {
		t.Fatalf("create project update: %v", err)
	}

	loaded, err := repos.ProjectUpdate.GetByID(item.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if loaded == nil {
		t.Fatal("GetByID returned nil")
	}
	if loaded.ID != item.ID {
		t.Errorf("ID: got %d, want %d", loaded.ID, item.ID)
	}
	if loaded.ProjectID != 42 {
		t.Errorf("ProjectID: got %d, want %d", loaded.ProjectID, int64(42))
	}
	if loaded.Kind != "changelog" {
		t.Errorf("Kind: got %q, want %q", loaded.Kind, "changelog")
	}
	if loaded.Title != "v2.0 Changelog" {
		t.Errorf("Title: got %q, want %q", loaded.Title, "v2.0 Changelog")
	}
	if loaded.ContentMD != "Added many features" {
		t.Errorf("ContentMD: got %q, want %q", loaded.ContentMD, "Added many features")
	}
	if !loaded.Published {
		t.Error("Published should be true")
	}
	if !loaded.Pinned {
		t.Error("Pinned should be true")
	}
}

func TestProjectUpdateGetByIDNotFound(t *testing.T) {
	db := InitSQLite(filepath.Join(t.TempDir(), "test.db"))
	t.Cleanup(func() { _ = db.Close() })
	AutoMigrate(db)
	repos := New(db)

	loaded, err := repos.ProjectUpdate.GetByID(99999)
	if err == nil {
		t.Fatal("expected error for nonexistent ID")
	}
	if err != sql.ErrNoRows {
		t.Fatalf("expected sql.ErrNoRows, got %v", err)
	}
	if loaded != nil {
		t.Fatal("expected nil for nonexistent ID")
	}
}

func TestProjectUpdateList(t *testing.T) {
	db := InitSQLite(filepath.Join(t.TempDir(), "test.db"))
	t.Cleanup(func() { _ = db.Close() })
	AutoMigrate(db)
	repos := New(db)

	now := time.Now().Truncate(time.Millisecond)
	items := []*model.ProjectUpdateEntry{
		{ProjectID: 7, Kind: "timeline", Title: "First", ContentMD: "Content A", Published: true, Pinned: false, SortOrder: 0, EventAt: now.Add(-2 * time.Hour)},
		{ProjectID: 7, Kind: "build_log", Title: "Second", ContentMD: "Content B", Published: true, Pinned: false, SortOrder: 0, EventAt: now.Add(-1 * time.Hour)},
	}
	for _, it := range items {
		if err := repos.ProjectUpdate.Create(it); err != nil {
			t.Fatalf("create update %q: %v", it.Title, err)
		}
	}

	result, err := repos.ProjectUpdate.List(ProjectUpdateListOptions{ProjectID: 7})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 updates, got %d", len(result))
	}
}

func TestProjectUpdateUpdate(t *testing.T) {
	db := InitSQLite(filepath.Join(t.TempDir(), "test.db"))
	t.Cleanup(func() { _ = db.Close() })
	AutoMigrate(db)
	repos := New(db)

	now := time.Now().Truncate(time.Millisecond)
	item := &model.ProjectUpdateEntry{
		ProjectID: 5,
		Kind:      "timeline",
		Title:     "Original Title",
		ContentMD: "Original content",
		Published: false,
		Pinned:    false,
		SortOrder: 0,
		EventAt:   now,
	}
	if err := repos.ProjectUpdate.Create(item); err != nil {
		t.Fatalf("create update: %v", err)
	}

	updatedAt := now.Add(1 * time.Hour)
	item.Title = "Updated Title"
	item.ContentMD = "Updated content"
	item.Published = true
	item.Pinned = true
	item.UpdatedAt = updatedAt

	if err := repos.ProjectUpdate.Update(item); err != nil {
		t.Fatalf("Update: %v", err)
	}

	loaded, err := repos.ProjectUpdate.GetByID(item.ID)
	if err != nil {
		t.Fatalf("GetByID after update: %v", err)
	}
	if loaded.Title != "Updated Title" {
		t.Errorf("Title: got %q, want %q", loaded.Title, "Updated Title")
	}
	if loaded.ContentMD != "Updated content" {
		t.Errorf("ContentMD: got %q, want %q", loaded.ContentMD, "Updated content")
	}
	if !loaded.Published {
		t.Error("Published should be true after update")
	}
	if !loaded.Pinned {
		t.Error("Pinned should be true after update")
	}
	if !loaded.UpdatedAt.Equal(updatedAt) {
		t.Errorf("UpdatedAt: got %v, want %v", loaded.UpdatedAt, updatedAt)
	}
}

func TestProjectUpdateDelete(t *testing.T) {
	db := InitSQLite(filepath.Join(t.TempDir(), "test.db"))
	t.Cleanup(func() { _ = db.Close() })
	AutoMigrate(db)
	repos := New(db)

	now := time.Now().Truncate(time.Millisecond)
	item := &model.ProjectUpdateEntry{
		ProjectID: 3,
		Kind:      "timeline",
		Title:     "Delete Me",
		ContentMD: "To be deleted",
		Published: true,
		Pinned:    false,
		SortOrder: 0,
		EventAt:   now,
	}
	if err := repos.ProjectUpdate.Create(item); err != nil {
		t.Fatalf("create update: %v", err)
	}

	if err := repos.ProjectUpdate.Delete(item.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	loaded, err := repos.ProjectUpdate.GetByID(item.ID)
	if err == nil {
		t.Fatal("expected error after deleting update")
	}
	if err != sql.ErrNoRows {
		t.Fatalf("expected sql.ErrNoRows, got %v", err)
	}
	if loaded != nil {
		t.Fatal("expected nil update after deletion")
	}
}

// ---------------------------------------------------------------------------
// project_showcase_repo.go — 6 tests
// ---------------------------------------------------------------------------

func TestProjectShowcaseCreate(t *testing.T) {
	db := InitSQLite(filepath.Join(t.TempDir(), "test.db"))
	t.Cleanup(func() { _ = db.Close() })
	AutoMigrate(db)
	repos := New(db)

	item := &model.ProjectShowcaseItem{
		ProjectID:     1,
		Kind:          "image",
		Title:         "Screenshot",
		BodyMD:        "A screenshot of the UI",
		AssetURL:      "https://example.com/img.png",
		ExternalURL:   "",
		EmbedProvider: "",
		EmbedRef:      "",
		Published:     true,
		SortOrder:     0,
	}
	if err := repos.ProjectShowcase.Create(item); err != nil {
		t.Fatalf("create showcase: %v", err)
	}
	if item.ID == 0 {
		t.Fatal("expected showcase ID to be set after Create")
	}
	if item.ProjectID != 1 {
		t.Errorf("ProjectID: got %d, want %d", item.ProjectID, int64(1))
	}
	if item.Title != "Screenshot" {
		t.Errorf("Title: got %q, want %q", item.Title, "Screenshot")
	}
}

func TestProjectShowcaseGetByID(t *testing.T) {
	db := InitSQLite(filepath.Join(t.TempDir(), "test.db"))
	t.Cleanup(func() { _ = db.Close() })
	AutoMigrate(db)
	repos := New(db)

	item := &model.ProjectShowcaseItem{
		ProjectID:     2,
		Kind:          "link",
		Title:         "Demo Link",
		BodyMD:        "Check out the demo",
		AssetURL:      "",
		ExternalURL:   "https://example.com/demo",
		EmbedProvider: "",
		EmbedRef:      "",
		Published:     true,
		SortOrder:     1,
	}
	if err := repos.ProjectShowcase.Create(item); err != nil {
		t.Fatalf("create showcase: %v", err)
	}

	loaded, err := repos.ProjectShowcase.GetByID(item.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if loaded == nil {
		t.Fatal("GetByID returned nil")
	}
	if loaded.ID != item.ID {
		t.Errorf("ID: got %d, want %d", loaded.ID, item.ID)
	}
	if loaded.ProjectID != 2 {
		t.Errorf("ProjectID: got %d, want %d", loaded.ProjectID, int64(2))
	}
	if loaded.Kind != "link" {
		t.Errorf("Kind: got %q, want %q", loaded.Kind, "link")
	}
	if loaded.Title != "Demo Link" {
		t.Errorf("Title: got %q, want %q", loaded.Title, "Demo Link")
	}
	if loaded.ExternalURL != "https://example.com/demo" {
		t.Errorf("ExternalURL: got %q, want %q", loaded.ExternalURL, "https://example.com/demo")
	}
	if !loaded.Published {
		t.Error("Published should be true")
	}
}

func TestProjectShowcaseGetByIDNotFound(t *testing.T) {
	db := InitSQLite(filepath.Join(t.TempDir(), "test.db"))
	t.Cleanup(func() { _ = db.Close() })
	AutoMigrate(db)
	repos := New(db)

	loaded, err := repos.ProjectShowcase.GetByID(99999)
	if err == nil {
		t.Fatal("expected error for nonexistent ID")
	}
	if err != sql.ErrNoRows {
		t.Fatalf("expected sql.ErrNoRows, got %v", err)
	}
	if loaded != nil {
		t.Fatal("expected nil for nonexistent ID")
	}
}

func TestProjectShowcaseListByProjectID(t *testing.T) {
	db := InitSQLite(filepath.Join(t.TempDir(), "test.db"))
	t.Cleanup(func() { _ = db.Close() })
	AutoMigrate(db)
	repos := New(db)

	items := []*model.ProjectShowcaseItem{
		{ProjectID: 10, Kind: "image", Title: "Image 1", BodyMD: "body", Published: true, SortOrder: 0},
		{ProjectID: 10, Kind: "link", Title: "Link 1", BodyMD: "body", ExternalURL: "https://x.com", Published: true, SortOrder: 1},
	}
	for _, it := range items {
		if err := repos.ProjectShowcase.Create(it); err != nil {
			t.Fatalf("create showcase %q: %v", it.Title, err)
		}
	}

	result, err := repos.ProjectShowcase.ListByProjectID(10, false)
	if err != nil {
		t.Fatalf("ListByProjectID: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 showcase items, got %d", len(result))
	}
	if result[0].Title != "Image 1" {
		t.Errorf("first item Title: got %q, want %q", result[0].Title, "Image 1")
	}
	if result[1].Title != "Link 1" {
		t.Errorf("second item Title: got %q, want %q", result[1].Title, "Link 1")
	}
}

func TestProjectShowcaseUpdate(t *testing.T) {
	db := InitSQLite(filepath.Join(t.TempDir(), "test.db"))
	t.Cleanup(func() { _ = db.Close() })
	AutoMigrate(db)
	repos := New(db)

	item := &model.ProjectShowcaseItem{
		ProjectID: 7,
		Kind:      "image",
		Title:     "Original Showcase",
		BodyMD:    "Original body",
		Published: false,
		SortOrder: 0,
	}
	if err := repos.ProjectShowcase.Create(item); err != nil {
		t.Fatalf("create showcase: %v", err)
	}

	updatedAt := time.Now().Truncate(time.Millisecond)
	item.Title = "Updated Showcase"
	item.BodyMD = "Updated body"
	item.Published = true
	item.SortOrder = 5
	item.UpdatedAt = updatedAt

	if err := repos.ProjectShowcase.Update(item); err != nil {
		t.Fatalf("Update: %v", err)
	}

	loaded, err := repos.ProjectShowcase.GetByID(item.ID)
	if err != nil {
		t.Fatalf("GetByID after update: %v", err)
	}
	if loaded.Title != "Updated Showcase" {
		t.Errorf("Title: got %q, want %q", loaded.Title, "Updated Showcase")
	}
	if loaded.BodyMD != "Updated body" {
		t.Errorf("BodyMD: got %q, want %q", loaded.BodyMD, "Updated body")
	}
	if !loaded.Published {
		t.Error("Published should be true after update")
	}
	if loaded.SortOrder != 5 {
		t.Errorf("SortOrder: got %d, want %d", loaded.SortOrder, 5)
	}
	if !loaded.UpdatedAt.Equal(updatedAt) {
		t.Errorf("UpdatedAt: got %v, want %v", loaded.UpdatedAt, updatedAt)
	}
}

func TestProjectShowcaseDelete(t *testing.T) {
	db := InitSQLite(filepath.Join(t.TempDir(), "test.db"))
	t.Cleanup(func() { _ = db.Close() })
	AutoMigrate(db)
	repos := New(db)

	item := &model.ProjectShowcaseItem{
		ProjectID: 4,
		Kind:      "video",
		Title:     "Delete Me",
		BodyMD:    "To be deleted",
		Published: true,
		SortOrder: 0,
	}
	if err := repos.ProjectShowcase.Create(item); err != nil {
		t.Fatalf("create showcase: %v", err)
	}

	if err := repos.ProjectShowcase.Delete(item.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	loaded, err := repos.ProjectShowcase.GetByID(item.ID)
	if err == nil {
		t.Fatal("expected error after deleting showcase")
	}
	if err != sql.ErrNoRows {
		t.Fatalf("expected sql.ErrNoRows, got %v", err)
	}
	if loaded != nil {
		t.Fatal("expected nil showcase after deletion")
	}
}

// ---------------------------------------------------------------------------
// category_repo.go — 1 test
// ---------------------------------------------------------------------------

func TestCategoryRepoSeed(t *testing.T) {
	db := InitSQLite(filepath.Join(t.TempDir(), "test.db"))
	t.Cleanup(func() { _ = db.Close() })
	AutoMigrate(db)
	repos := New(db)

	cat := &model.Category{
		Slug:        "test-seed",
		Name:        "Test Seed",
		Description: "A seeded category",
	}
	if err := repos.Category.Seed(cat); err != nil {
		t.Fatalf("Seed category: %v", err)
	}
	if cat.ID == 0 {
		t.Fatal("expected category ID to be set after Seed")
	}

	// Verify it exists via GetBySlug
	loaded, err := repos.Category.GetBySlug("test-seed")
	if err != nil {
		t.Fatalf("GetBySlug after seed: %v", err)
	}
	if loaded.Name != "Test Seed" {
		t.Errorf("Name: got %q, want %q", loaded.Name, "Test Seed")
	}
	if loaded.Description != "A seeded category" {
		t.Errorf("Description: got %q, want %q", loaded.Description, "A seeded category")
	}
	if loaded.CreatedAt.IsZero() {
		t.Error("CreatedAt should be set by Seed")
	}
	if loaded.UpdatedAt.IsZero() {
		t.Error("UpdatedAt should be set by Seed")
	}

	// Seed again with same slug (INSERT OR IGNORE) — should not error and should load existing
	cat2 := &model.Category{
		Slug: "test-seed",
		Name: "Should Not Overwrite",
	}
	if err := repos.Category.Seed(cat2); err != nil {
		t.Fatalf("Seed duplicate category: %v", err)
	}
	if cat2.ID != cat.ID {
		t.Errorf("duplicate Seed should return same ID: got %d, want %d", cat2.ID, cat.ID)
	}
	if cat2.Name != "Test Seed" {
		t.Errorf("duplicate Seed should not overwrite Name: got %q, want %q", cat2.Name, "Test Seed")
	}
}
