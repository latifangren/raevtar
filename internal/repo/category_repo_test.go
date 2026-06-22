package repo

import (
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	"raevtar/internal/model"
)

func TestCategoryRepoCreateAndGetByID(t *testing.T) {
	db := InitSQLite(filepath.Join(t.TempDir(), "test.db"))
	t.Cleanup(func() { _ = db.Close() })
	AutoMigrate(db)
	repos := New(db)

	now := time.Now().Truncate(time.Millisecond)
	cat := &model.Category{
		Slug:        "test-category",
		Name:        "Test Category",
		Description: "A category for testing",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := repos.Category.Create(cat); err != nil {
		t.Fatalf("create category: %v", err)
	}
	if cat.ID == 0 {
		t.Fatal("expected category.ID to be set after Create")
	}

	loaded, err := repos.Category.GetByID(cat.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if loaded == nil {
		t.Fatal("GetByID returned nil")
	}
	if loaded.ID != cat.ID {
		t.Errorf("ID: got %d, want %d", loaded.ID, cat.ID)
	}
	if loaded.Slug != cat.Slug {
		t.Errorf("Slug: got %q, want %q", loaded.Slug, cat.Slug)
	}
	if loaded.Name != cat.Name {
		t.Errorf("Name: got %q, want %q", loaded.Name, cat.Name)
	}
	if loaded.Description != cat.Description {
		t.Errorf("Description: got %q, want %q", loaded.Description, cat.Description)
	}
	if !loaded.CreatedAt.Equal(now) {
		t.Errorf("CreatedAt: got %v, want %v", loaded.CreatedAt, now)
	}
	if !loaded.UpdatedAt.Equal(now) {
		t.Errorf("UpdatedAt: got %v, want %v", loaded.UpdatedAt, now)
	}
}

func TestCategoryRepoGetBySlug(t *testing.T) {
	db := InitSQLite(filepath.Join(t.TempDir(), "test.db"))
	t.Cleanup(func() { _ = db.Close() })
	AutoMigrate(db)
	repos := New(db)

	now := time.Now().Truncate(time.Millisecond)
	cat := &model.Category{
		Slug:        "find-by-slug",
		Name:        "Find By Slug",
		Description: "Lookup via slug",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := repos.Category.Create(cat); err != nil {
		t.Fatalf("create category: %v", err)
	}

	loaded, err := repos.Category.GetBySlug("find-by-slug")
	if err != nil {
		t.Fatalf("GetBySlug: %v", err)
	}
	if loaded == nil {
		t.Fatal("GetBySlug returned nil")
	}
	if loaded.ID != cat.ID {
		t.Errorf("ID: got %d, want %d", loaded.ID, cat.ID)
	}
	if loaded.Slug != "find-by-slug" {
		t.Errorf("Slug: got %q, want %q", loaded.Slug, "find-by-slug")
	}
	if loaded.Name != "Find By Slug" {
		t.Errorf("Name: got %q, want %q", loaded.Name, "Find By Slug")
	}
}

func TestCategoryRepoGetBySlugNotFound(t *testing.T) {
	db := InitSQLite(filepath.Join(t.TempDir(), "test.db"))
	t.Cleanup(func() { _ = db.Close() })
	AutoMigrate(db)
	repos := New(db)

	loaded, err := repos.Category.GetBySlug("non-existent-slug")
	if err == nil {
		t.Fatal("expected error for non-existent slug")
	}
	if err != sql.ErrNoRows {
		t.Fatalf("expected sql.ErrNoRows, got %v", err)
	}
	if loaded != nil {
		t.Fatal("expected nil category for non-existent slug")
	}
}

func TestCategoryRepoList(t *testing.T) {
	db := InitSQLite(filepath.Join(t.TempDir(), "test.db"))
	t.Cleanup(func() { _ = db.Close() })
	AutoMigrate(db)
	repos := New(db)

	now := time.Now().Truncate(time.Millisecond)
	cats := []*model.Category{
		{Slug: "cat-c", Name: "Alpha", Description: "First alphabetically", CreatedAt: now, UpdatedAt: now},
		{Slug: "cat-a", Name: "Beta", Description: "Second alphabetically", CreatedAt: now, UpdatedAt: now},
		{Slug: "cat-b", Name: "Gamma", Description: "Third alphabetically", CreatedAt: now, UpdatedAt: now},
	}
	for _, c := range cats {
		if err := repos.Category.Create(c); err != nil {
			t.Fatalf("create category %q: %v", c.Slug, err)
		}
	}

	result, err := repos.Category.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(result) != 3 {
		t.Fatalf("expected 3 categories, got %d", len(result))
	}

	// Ordered by name ascending: Alpha, Beta, Gamma
	expectedNames := []string{"Alpha", "Beta", "Gamma"}
	for i, name := range expectedNames {
		if result[i].Name != name {
			t.Errorf("position %d: expected name %q, got %q", i, name, result[i].Name)
		}
	}
}

func TestCategoryRepoListEmpty(t *testing.T) {
	db := InitSQLite(filepath.Join(t.TempDir(), "test.db"))
	t.Cleanup(func() { _ = db.Close() })
	AutoMigrate(db)
	repos := New(db)

	result, err := repos.Category.List()
	if err != nil {
		t.Fatalf("List on empty table: %v", err)
	}
	if len(result) != 0 {
		t.Fatalf("expected 0 categories, got %d", len(result))
	}
}

func TestCategoryRepoUpdate(t *testing.T) {
	db := InitSQLite(filepath.Join(t.TempDir(), "test.db"))
	t.Cleanup(func() { _ = db.Close() })
	AutoMigrate(db)
	repos := New(db)

	now := time.Now().Truncate(time.Millisecond)
	cat := &model.Category{
		Slug:        "original",
		Name:        "Original Name",
		Description: "Original description",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := repos.Category.Create(cat); err != nil {
		t.Fatalf("create category: %v", err)
	}

	updatedAt := now.Add(1 * time.Hour)
	cat.Name = "Updated Name"
	cat.Description = "Updated description"
	cat.Slug = "updated"
	cat.UpdatedAt = updatedAt

	if err := repos.Category.Update(cat); err != nil {
		t.Fatalf("Update: %v", err)
	}

	loaded, err := repos.Category.GetByID(cat.ID)
	if err != nil {
		t.Fatalf("GetByID after update: %v", err)
	}

	if loaded.Name != "Updated Name" {
		t.Errorf("Name: got %q, want %q", loaded.Name, "Updated Name")
	}
	if loaded.Description != "Updated description" {
		t.Errorf("Description: got %q, want %q", loaded.Description, "Updated description")
	}
	if loaded.Slug != "updated" {
		t.Errorf("Slug: got %q, want %q", loaded.Slug, "updated")
	}
	if !loaded.UpdatedAt.Equal(updatedAt) {
		t.Errorf("UpdatedAt: got %v, want %v", loaded.UpdatedAt, updatedAt)
	}
	if !loaded.CreatedAt.Equal(now) {
		t.Errorf("CreatedAt changed after update: got %v, want %v", loaded.CreatedAt, now)
	}
}

func TestCategoryRepoDelete(t *testing.T) {
	db := InitSQLite(filepath.Join(t.TempDir(), "test.db"))
	t.Cleanup(func() { _ = db.Close() })
	AutoMigrate(db)
	repos := New(db)

	now := time.Now().Truncate(time.Millisecond)
	cat := &model.Category{
		Slug:      "delete-me",
		Name:      "Delete Me",
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := repos.Category.Create(cat); err != nil {
		t.Fatalf("create category: %v", err)
	}

	if err := repos.Category.Delete(cat.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	loaded, err := repos.Category.GetByID(cat.ID)
	if err == nil {
		t.Fatal("expected error after deleting category")
	}
	if err != sql.ErrNoRows {
		t.Fatalf("expected sql.ErrNoRows, got %v", err)
	}
	if loaded != nil {
		t.Fatal("expected nil category after deletion")
	}
}

func TestCategoryRepoSlugExists(t *testing.T) {
	db := InitSQLite(filepath.Join(t.TempDir(), "test.db"))
	t.Cleanup(func() { _ = db.Close() })
	AutoMigrate(db)
	repos := New(db)

	now := time.Now().Truncate(time.Millisecond)
	cat := &model.Category{
		Slug:      "existing-slug",
		Name:      "Existing",
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := repos.Category.Create(cat); err != nil {
		t.Fatalf("create category: %v", err)
	}

	exists, err := repos.Category.SlugExists("existing-slug")
	if err != nil {
		t.Fatalf("SlugExists for existing slug: %v", err)
	}
	if !exists {
		t.Error("SlugExists returned false for existing slug 'existing-slug'")
	}

	notExists, err := repos.Category.SlugExists("non-existent-slug")
	if err != nil {
		t.Fatalf("SlugExists for non-existent slug: %v", err)
	}
	if notExists {
		t.Error("SlugExists returned true for non-existent slug")
	}
}

func TestCategoryRepoSlugExistsExcludingID(t *testing.T) {
	db := InitSQLite(filepath.Join(t.TempDir(), "test.db"))
	t.Cleanup(func() { _ = db.Close() })
	AutoMigrate(db)
	repos := New(db)

	now := time.Now().Truncate(time.Millisecond)

	// Create two categories with different slugs
	catA := &model.Category{
		Slug:      "shared-slug",
		Name:      "Category A",
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := repos.Category.Create(catA); err != nil {
		t.Fatalf("create category A: %v", err)
	}

	catB := &model.Category{
		Slug:      "other-slug",
		Name:      "Category B",
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := repos.Category.Create(catB); err != nil {
		t.Fatalf("create category B: %v", err)
	}

	// Slug exists for a DIFFERENT ID (catB), exclude catB's ID
	existsOther, err := repos.Category.SlugExistsExcludingID("shared-slug", catB.ID)
	if err != nil {
		t.Fatalf("SlugExistsExcludingID for existing slug with exclude: %v", err)
	}
	if !existsOther {
		t.Error("SlugExistsExcludingID returned false for 'shared-slug' excluding catB's ID")
	}

	// Slug does NOT exist (different ID excluded), exclude catA's ID
	existsSelf, err := repos.Category.SlugExistsExcludingID("shared-slug", catA.ID)
	if err != nil {
		t.Fatalf("SlugExistsExcludingID for existing slug excluding own ID: %v", err)
	}
	if existsSelf {
		t.Error("SlugExistsExcludingID returned true for 'shared-slug' excluding its own ID")
	}

	// Non-existent slug
	notExists, err := repos.Category.SlugExistsExcludingID("no-such-slug", 0)
	if err != nil {
		t.Fatalf("SlugExistsExcludingID for non-existent slug: %v", err)
	}
	if notExists {
		t.Error("SlugExistsExcludingID returned true for non-existent slug")
	}
}
