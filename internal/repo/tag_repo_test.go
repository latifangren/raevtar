package repo

import (
	"path/filepath"
	"strings"
	"testing"
	"time"

	"raevtar/internal/model"
)

func TestTagRepoEnsureCreatesTag(t *testing.T) {
	db := InitSQLite(filepath.Join(t.TempDir(), "test.db"))
	t.Cleanup(func() { _ = db.Close() })
	AutoMigrate(db)
	repos := New(db)

	tag, err := repos.Tag.Ensure("Golang")
	if err != nil {
		t.Fatalf("Ensure: %v", err)
	}
	if tag.ID == 0 {
		t.Fatal("expected tag.ID to be set after Ensure")
	}
	if tag.Name != "Golang" {
		t.Errorf("Name: got %q, want %q", tag.Name, "Golang")
	}
	if tag.Slug != "golang" {
		t.Errorf("Slug: got %q, want %q", tag.Slug, "golang")
	}
	if tag.CreatedAt.IsZero() {
		t.Error("CreatedAt should not be zero")
	}
}

func TestTagRepoEnsureReturnsExistingOnDuplicateSlug(t *testing.T) {
	db := InitSQLite(filepath.Join(t.TempDir(), "test.db"))
	t.Cleanup(func() { _ = db.Close() })
	AutoMigrate(db)
	repos := New(db)

	tag1, err := repos.Tag.Ensure("Golang")
	if err != nil {
		t.Fatalf("first Ensure: %v", err)
	}
	if tag1.ID == 0 {
		t.Fatal("expected first tag.ID to be set")
	}

	tag2, err := repos.Tag.Ensure("Golang")
	if err != nil {
		t.Fatalf("second Ensure: %v", err)
	}
	if tag1.ID != tag2.ID {
		t.Errorf("expected same ID for duplicate slug: got %d and %d", tag1.ID, tag2.ID)
	}
	if tag2.Name != "Golang" {
		t.Errorf("Name: got %q, want %q", tag2.Name, "Golang")
	}
}

func TestTagRepoEnsureWithInvalidNameReturnsError(t *testing.T) {
	db := InitSQLite(filepath.Join(t.TempDir(), "test.db"))
	t.Cleanup(func() { _ = db.Close() })
	AutoMigrate(db)
	repos := New(db)

	_, err := repos.Tag.Ensure("")
	if err == nil {
		t.Fatal("expected error for empty tag name")
	}
	if !strings.Contains(err.Error(), "invalid tag name") {
		t.Errorf("expected error to contain 'invalid tag name', got %v", err)
	}
}

func TestTagRepoSetTagsAndGetByPostID(t *testing.T) {
	db := InitSQLite(filepath.Join(t.TempDir(), "test.db"))
	t.Cleanup(func() { _ = db.Close() })
	AutoMigrate(db)
	repos := New(db)

	cat := &model.Category{
		Slug:      "tag-post-cat",
		Name:      "Tag Post Category",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := repos.Category.Create(cat); err != nil {
		t.Fatalf("create category: %v", err)
	}

	now := time.Now().Truncate(time.Millisecond)
	post := &model.Post{
		CategoryID: cat.ID,
		Title:      "Tagged Post",
		Slug:       "tagged-post",
		ContentMD:  "content with tags",
		Excerpt:    "tagged excerpt",
		Published:  true,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	if err := repos.Post.Create(post); err != nil {
		t.Fatalf("create post: %v", err)
	}

	if err := repos.Tag.SetTags(post.ID, []string{"Go", "Testing", "Database"}); err != nil {
		t.Fatalf("SetTags: %v", err)
	}

	tags, err := repos.Tag.GetByPostID(post.ID)
	if err != nil {
		t.Fatalf("GetByPostID: %v", err)
	}
	if len(tags) != 3 {
		t.Fatalf("expected 3 tags, got %d", len(tags))
	}

	expected := []string{"Database", "Go", "Testing"}
	for i, tag := range tags {
		if tag.Name != expected[i] {
			t.Errorf("tag[%d].Name: got %q, want %q", i, tag.Name, expected[i])
		}
		if tag.ID == 0 {
			t.Errorf("tag[%d].ID is zero after SetTags", i)
		}
	}
}

func TestTagRepoSetTagsReplacesOldTags(t *testing.T) {
	db := InitSQLite(filepath.Join(t.TempDir(), "test.db"))
	t.Cleanup(func() { _ = db.Close() })
	AutoMigrate(db)
	repos := New(db)

	cat := &model.Category{
		Slug:      "replace-cat",
		Name:      "Replace Category",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := repos.Category.Create(cat); err != nil {
		t.Fatalf("create category: %v", err)
	}

	now := time.Now().Truncate(time.Millisecond)
	post := &model.Post{
		CategoryID: cat.ID,
		Title:      "Replace Tags Post",
		Slug:       "replace-tags-post",
		ContentMD:  "content",
		Excerpt:    "excerpt",
		Published:  true,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	if err := repos.Post.Create(post); err != nil {
		t.Fatalf("create post: %v", err)
	}

	if err := repos.Tag.SetTags(post.ID, []string{"Old", "Tags"}); err != nil {
		t.Fatalf("first SetTags: %v", err)
	}

	if err := repos.Tag.SetTags(post.ID, []string{"New", "Replacement"}); err != nil {
		t.Fatalf("second SetTags: %v", err)
	}

	tags, err := repos.Tag.GetByPostID(post.ID)
	if err != nil {
		t.Fatalf("GetByPostID: %v", err)
	}
	if len(tags) != 2 {
		t.Fatalf("expected 2 tags after replacement, got %d", len(tags))
	}
	if tags[0].Name != "New" {
		t.Errorf("expected first tag 'New', got %q", tags[0].Name)
	}
	if tags[1].Name != "Replacement" {
		t.Errorf("expected second tag 'Replacement', got %q", tags[1].Name)
	}
}

func TestTagRepoGetByPostIDs(t *testing.T) {
	db := InitSQLite(filepath.Join(t.TempDir(), "test.db"))
	t.Cleanup(func() { _ = db.Close() })
	AutoMigrate(db)
	repos := New(db)

	cat := &model.Category{
		Slug:      "multi-post-cat",
		Name:      "Multi Post Category",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := repos.Category.Create(cat); err != nil {
		t.Fatalf("create category: %v", err)
	}

	now := time.Now().Truncate(time.Millisecond)
	postA := &model.Post{
		CategoryID: cat.ID, Title: "Post A", Slug: "post-a",
		ContentMD: "a", Excerpt: "a", Published: true,
		CreatedAt: now, UpdatedAt: now,
	}
	postB := &model.Post{
		CategoryID: cat.ID, Title: "Post B", Slug: "post-b",
		ContentMD: "b", Excerpt: "b", Published: true,
		CreatedAt: now, UpdatedAt: now,
	}
	if err := repos.Post.Create(postA); err != nil {
		t.Fatalf("create post A: %v", err)
	}
	if err := repos.Post.Create(postB); err != nil {
		t.Fatalf("create post B: %v", err)
	}

	if err := repos.Tag.SetTags(postA.ID, []string{"Alpha", "Beta"}); err != nil {
		t.Fatalf("SetTags for post A: %v", err)
	}
	if err := repos.Tag.SetTags(postB.ID, []string{"Beta", "Gamma"}); err != nil {
		t.Fatalf("SetTags for post B: %v", err)
	}

	result, err := repos.Tag.GetByPostIDs([]int64{postA.ID, postB.ID})
	if err != nil {
		t.Fatalf("GetByPostIDs: %v", err)
	}

	if len(result) != 2 {
		t.Fatalf("expected 2 entries in result map, got %d", len(result))
	}

	tagsA := result[postA.ID]
	if len(tagsA) != 2 {
		t.Fatalf("expected 2 tags for post A, got %d", len(tagsA))
	}
	if tagsA[0].Name != "Alpha" {
		t.Errorf("postA first tag: got %q, want %q", tagsA[0].Name, "Alpha")
	}
	if tagsA[1].Name != "Beta" {
		t.Errorf("postA second tag: got %q, want %q", tagsA[1].Name, "Beta")
	}

	tagsB := result[postB.ID]
	if len(tagsB) != 2 {
		t.Fatalf("expected 2 tags for post B, got %d", len(tagsB))
	}
}

func TestTagRepoGetByPostIDsEmptyInput(t *testing.T) {
	db := InitSQLite(filepath.Join(t.TempDir(), "test.db"))
	t.Cleanup(func() { _ = db.Close() })
	AutoMigrate(db)
	repos := New(db)

	result, err := repos.Tag.GetByPostIDs([]int64{})
	if err != nil {
		t.Fatalf("GetByPostIDs with empty slice: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected empty map, got %d entries", len(result))
	}
}

func TestTagRepoSetProjectTagsAndGetByProjectID(t *testing.T) {
	db := InitSQLite(filepath.Join(t.TempDir(), "test.db"))
	t.Cleanup(func() { _ = db.Close() })
	AutoMigrate(db)
	repos := New(db)

	proj := &model.Project{
		Title:     "Tagged Project",
		Slug:      "tagged-project",
		ContentMD: "project with tags",
		Excerpt:   "tagged project excerpt",
		Published: true,
		State:     model.ProjectStateActive,
	}
	if err := repos.Project.Create(proj); err != nil {
		t.Fatalf("create project: %v", err)
	}

	if err := repos.Tag.SetProjectTags(proj.ID, []string{"Frontend", "React", "TypeScript"}); err != nil {
		t.Fatalf("SetProjectTags: %v", err)
	}

	tags, err := repos.Tag.GetByProjectID(proj.ID)
	if err != nil {
		t.Fatalf("GetByProjectID: %v", err)
	}
	if len(tags) != 3 {
		t.Fatalf("expected 3 tags, got %d", len(tags))
	}

	expected := []string{"Frontend", "React", "TypeScript"}
	for i, tag := range tags {
		if tag.Name != expected[i] {
			t.Errorf("tag[%d].Name: got %q, want %q", i, tag.Name, expected[i])
		}
	}
}

func TestTagRepoGetByProjectIDs(t *testing.T) {
	db := InitSQLite(filepath.Join(t.TempDir(), "test.db"))
	t.Cleanup(func() { _ = db.Close() })
	AutoMigrate(db)
	repos := New(db)

	projA := &model.Project{
		Title: "Project A", Slug: "proj-a",
		ContentMD: "a", Excerpt: "a",
		Published: true, State: model.ProjectStateActive,
	}
	projB := &model.Project{
		Title: "Project B", Slug: "proj-b",
		ContentMD: "b", Excerpt: "b",
		Published: true, State: model.ProjectStateActive,
	}
	if err := repos.Project.Create(projA); err != nil {
		t.Fatalf("create project A: %v", err)
	}
	if err := repos.Project.Create(projB); err != nil {
		t.Fatalf("create project B: %v", err)
	}

	if err := repos.Tag.SetProjectTags(projA.ID, []string{"API", "Backend"}); err != nil {
		t.Fatalf("SetProjectTags for A: %v", err)
	}
	if err := repos.Tag.SetProjectTags(projB.ID, []string{"Frontend", "UI"}); err != nil {
		t.Fatalf("SetProjectTags for B: %v", err)
	}

	result, err := repos.Tag.GetByProjectIDs([]int64{projA.ID, projB.ID})
	if err != nil {
		t.Fatalf("GetByProjectIDs: %v", err)
	}

	if len(result) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(result))
	}
	if len(result[projA.ID]) != 2 {
		t.Errorf("expected 2 tags for project A, got %d", len(result[projA.ID]))
	}
	if len(result[projB.ID]) != 2 {
		t.Errorf("expected 2 tags for project B, got %d", len(result[projB.ID]))
	}
}

func TestTagRepoGetByProjectIDsEmptyInput(t *testing.T) {
	db := InitSQLite(filepath.Join(t.TempDir(), "test.db"))
	t.Cleanup(func() { _ = db.Close() })
	AutoMigrate(db)
	repos := New(db)

	result, err := repos.Tag.GetByProjectIDs([]int64{})
	if err != nil {
		t.Fatalf("GetByProjectIDs with empty slice: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected empty map, got %d entries", len(result))
	}
}
