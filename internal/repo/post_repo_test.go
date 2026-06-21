package repo

import (
	"database/sql"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"raevtar/internal/model"
)

func TestPostRepoCreateAndGetByID(t *testing.T) {
	db := InitSQLite(filepath.Join(t.TempDir(), "test.db"))
	t.Cleanup(func() { _ = db.Close() })
	AutoMigrate(db)
	repos := New(db)

	cat := &model.Category{
		Slug:        "test-cat",
		Name:        "Test Category",
		Description: "A category for testing",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	if err := repos.Category.Create(cat); err != nil {
		t.Fatalf("create category: %v", err)
	}

	now := time.Now().Truncate(time.Millisecond)
	post := &model.Post{
		CategoryID:    cat.ID,
		Title:         "Test Post Title",
		Slug:          "test-post-title",
		ContentMD:     "# Hello\n\nThis is markdown content.",
		Excerpt:       "A short excerpt for the test post",
		CoverImageURL: "https://example.com/cover.jpg",
		Published:     true,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if err := repos.Post.Create(post); err != nil {
		t.Fatalf("create post: %v", err)
	}
	if post.ID == 0 {
		t.Fatal("expected post.ID to be set after Create")
	}

	loaded, err := repos.Post.GetByID(post.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if loaded == nil {
		t.Fatal("GetByID returned nil post")
	}

	if loaded.ID != post.ID {
		t.Errorf("ID: got %d, want %d", loaded.ID, post.ID)
	}
	if loaded.CategoryID != post.CategoryID {
		t.Errorf("CategoryID: got %d, want %d", loaded.CategoryID, post.CategoryID)
	}
	if loaded.Title != post.Title {
		t.Errorf("Title: got %q, want %q", loaded.Title, post.Title)
	}
	if loaded.Slug != post.Slug {
		t.Errorf("Slug: got %q, want %q", loaded.Slug, post.Slug)
	}
	if loaded.ContentMD != post.ContentMD {
		t.Errorf("ContentMD: got %q, want %q", loaded.ContentMD, post.ContentMD)
	}
	if loaded.Excerpt != post.Excerpt {
		t.Errorf("Excerpt: got %q, want %q", loaded.Excerpt, post.Excerpt)
	}
	if loaded.CoverImageURL != post.CoverImageURL {
		t.Errorf("CoverImageURL: got %q, want %q", loaded.CoverImageURL, post.CoverImageURL)
	}
	if loaded.Published != post.Published {
		t.Errorf("Published: got %v, want %v", loaded.Published, post.Published)
	}
	if loaded.CreatedAt.IsZero() {
		t.Error("CreatedAt should not be zero (set by SQLite DEFAULT)")
	}
	if loaded.UpdatedAt.IsZero() {
		t.Error("UpdatedAt should not be zero (set by SQLite DEFAULT)")
	}
	if loaded.CategoryName != cat.Name {
		t.Errorf("CategoryName: got %q, want %q", loaded.CategoryName, cat.Name)
	}
	if loaded.CategorySlug != cat.Slug {
		t.Errorf("CategorySlug: got %q, want %q", loaded.CategorySlug, cat.Slug)
	}
}

func TestPostRepoGetBySlug(t *testing.T) {
	db := InitSQLite(filepath.Join(t.TempDir(), "test.db"))
	t.Cleanup(func() { _ = db.Close() })
	AutoMigrate(db)
	repos := New(db)

	cat := &model.Category{
		Slug:      "slug-cat",
		Name:      "Slug Category",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := repos.Category.Create(cat); err != nil {
		t.Fatalf("create category: %v", err)
	}

	now := time.Now().Truncate(time.Millisecond)
	post := &model.Post{
		CategoryID: cat.ID,
		Title:      "Find Me By Slug",
		Slug:       "find-me-by-slug",
		ContentMD:  "slug-based content",
		Excerpt:    "slug excerpt",
		Published:  true,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	if err := repos.Post.Create(post); err != nil {
		t.Fatalf("create post: %v", err)
	}

	loaded, err := repos.Post.GetBySlug("find-me-by-slug")
	if err != nil {
		t.Fatalf("GetBySlug: %v", err)
	}
	if loaded == nil {
		t.Fatal("GetBySlug returned nil")
	}
	if loaded.ID != post.ID {
		t.Errorf("ID: got %d, want %d", loaded.ID, post.ID)
	}
	if loaded.Slug != "find-me-by-slug" {
		t.Errorf("Slug: got %q, want %q", loaded.Slug, "find-me-by-slug")
	}
	if loaded.CategorySlug != "slug-cat" {
		t.Errorf("CategorySlug: got %q, want %q", loaded.CategorySlug, "slug-cat")
	}
}

func TestPostRepoGetBySlugNotFound(t *testing.T) {
	db := InitSQLite(filepath.Join(t.TempDir(), "test.db"))
	t.Cleanup(func() { _ = db.Close() })
	AutoMigrate(db)
	repos := New(db)

	loaded, err := repos.Post.GetBySlug("non-existent-slug")
	if err == nil {
		t.Fatal("expected error for non-existent slug")
	}
	if err != sql.ErrNoRows {
		t.Fatalf("expected sql.ErrNoRows, got %v", err)
	}
	if loaded != nil {
		t.Fatal("expected nil post for non-existent slug")
	}
}

func TestPostRepoGetByIDNotFound(t *testing.T) {
	db := InitSQLite(filepath.Join(t.TempDir(), "test.db"))
	t.Cleanup(func() { _ = db.Close() })
	AutoMigrate(db)
	repos := New(db)

	loaded, err := repos.Post.GetByID(99999)
	if err == nil {
		t.Fatal("expected error for non-existent ID")
	}
	if err != sql.ErrNoRows {
		t.Fatalf("expected sql.ErrNoRows, got %v", err)
	}
	if loaded != nil {
		t.Fatal("expected nil post for non-existent ID")
	}
}

func TestPostRepoList(t *testing.T) {
	db := InitSQLite(filepath.Join(t.TempDir(), "test.db"))
	t.Cleanup(func() { _ = db.Close() })
	AutoMigrate(db)
	repos := New(db)

	cat := &model.Category{
		Slug:      "list-cat",
		Name:      "List Category",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := repos.Category.Create(cat); err != nil {
		t.Fatalf("create category: %v", err)
	}

	now := time.Now().Truncate(time.Second)
	for _, p := range []*model.Post{
		{CategoryID: cat.ID, Title: "Third Post", Slug: "third-post", ContentMD: "content three", Excerpt: "excerpt three", Published: true},
		{CategoryID: cat.ID, Title: "Second Post", Slug: "second-post", ContentMD: "content two", Excerpt: "excerpt two", Published: true},
		{CategoryID: cat.ID, Title: "First Post", Slug: "first-post", ContentMD: "content one", Excerpt: "excerpt one", Published: true},
	} {
		if err := repos.Post.Create(p); err != nil {
			t.Fatalf("create post %q: %v", p.Slug, err)
		}
	}

	// Set specific created_at values (Create doesn't store them — uses SQLite DEFAULT)
	if _, err := db.Exec("UPDATE posts SET created_at = ? WHERE slug = ?", now.Add(-2*time.Hour), "third-post"); err != nil {
		t.Fatalf("set created_at for third-post: %v", err)
	}
	if _, err := db.Exec("UPDATE posts SET created_at = ? WHERE slug = ?", now.Add(-1*time.Hour), "second-post"); err != nil {
		t.Fatalf("set created_at for second-post: %v", err)
	}
	if _, err := db.Exec("UPDATE posts SET created_at = ? WHERE slug = ?", now, "first-post"); err != nil {
		t.Fatalf("set created_at for first-post: %v", err)
	}

	result, err := repos.Post.List("", false, 10, 0)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(result) != 3 {
		t.Fatalf("expected 3 posts, got %d", len(result))
	}

	// Ordered by created_at DESC: First (now), Second (now-1h), Third (now-2h)
	expectedOrder := []string{"First Post", "Second Post", "Third Post"}
	for i, title := range expectedOrder {
		if result[i].Title != title {
			t.Errorf("position %d: expected title %q, got %q", i, title, result[i].Title)
		}
	}
}

func TestPostRepoListWithLimitOffset(t *testing.T) {
	db := InitSQLite(filepath.Join(t.TempDir(), "test.db"))
	t.Cleanup(func() { _ = db.Close() })
	AutoMigrate(db)
	repos := New(db)

	cat := &model.Category{
		Slug:      "lo-cat",
		Name:      "Limit Offset Category",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := repos.Category.Create(cat); err != nil {
		t.Fatalf("create category: %v", err)
	}

	slugs := []string{"post-a", "post-b", "post-c", "post-d", "post-e"}
	for i, slug := range slugs {
		p := &model.Post{
			CategoryID: cat.ID,
			Title:      "Post " + string(rune('A'+i)),
			Slug:       slug,
			ContentMD:  "content",
			Excerpt:    "excerpt",
			Published:  true,
		}
		if err := repos.Post.Create(p); err != nil {
			t.Fatalf("create post %q: %v", slug, err)
		}
	}

	// Set distinct created_at values for deterministic ordering
	now := time.Now().Truncate(time.Second)
	for i, slug := range slugs {
		if _, err := db.Exec("UPDATE posts SET created_at = ? WHERE slug = ?", now.Add(-time.Duration(i)*time.Minute), slug); err != nil {
			t.Fatalf("set created_at for %q: %v", slug, err)
		}
	}

	result, err := repos.Post.List("", false, 2, 0)
	if err != nil {
		t.Fatalf("List with limit=2: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 posts with limit=2, got %d", len(result))
	}

	result2, err := repos.Post.List("", false, 2, 2)
	if err != nil {
		t.Fatalf("List with limit=2 offset=2: %v", err)
	}
	if len(result2) != 2 {
		t.Fatalf("expected 2 posts with limit=2 offset=2, got %d", len(result2))
	}

	// First page should have newer posts than second page
	if result[0].ID == result2[0].ID {
		t.Error("limit/offset pages overlapped")
	}
	if result[0].ID == result2[1].ID || result[1].ID == result2[0].ID {
		t.Error("pages should not share posts")
	}
}

func TestPostRepoListWithCategoryFilter(t *testing.T) {
	db := InitSQLite(filepath.Join(t.TempDir(), "test.db"))
	t.Cleanup(func() { _ = db.Close() })
	AutoMigrate(db)
	repos := New(db)

	catA := &model.Category{
		Slug: "cat-a", Name: "Category A",
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	catB := &model.Category{
		Slug: "cat-b", Name: "Category B",
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	if err := repos.Category.Create(catA); err != nil {
		t.Fatalf("create category A: %v", err)
	}
	if err := repos.Category.Create(catB); err != nil {
		t.Fatalf("create category B: %v", err)
	}

	now := time.Now().Truncate(time.Millisecond)
	posts := []*model.Post{
		{CategoryID: catA.ID, Title: "A-One", Slug: "a-one", ContentMD: "c", Excerpt: "e", Published: true, CreatedAt: now.Add(-3 * time.Minute), UpdatedAt: now},
		{CategoryID: catA.ID, Title: "A-Two", Slug: "a-two", ContentMD: "c", Excerpt: "e", Published: true, CreatedAt: now.Add(-2 * time.Minute), UpdatedAt: now},
		{CategoryID: catB.ID, Title: "B-One", Slug: "b-one", ContentMD: "c", Excerpt: "e", Published: true, CreatedAt: now.Add(-1 * time.Minute), UpdatedAt: now},
	}
	for _, p := range posts {
		if err := repos.Post.Create(p); err != nil {
			t.Fatalf("create post %q: %v", p.Slug, err)
		}
	}

	result, err := repos.Post.List("cat-a", false, 10, 0)
	if err != nil {
		t.Fatalf("List with category filter: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 posts in cat-a, got %d", len(result))
	}
	for _, p := range result {
		if p.CategorySlug != "cat-a" {
			t.Errorf("post %q has CategorySlug=%q, want %q", p.Slug, p.CategorySlug, "cat-a")
		}
	}

	resultB, err := repos.Post.List("cat-b", false, 10, 0)
	if err != nil {
		t.Fatalf("List with category filter cat-b: %v", err)
	}
	if len(resultB) != 1 {
		t.Fatalf("expected 1 post in cat-b, got %d", len(resultB))
	}
	if resultB[0].Slug != "b-one" {
		t.Errorf("expected slug 'b-one', got %q", resultB[0].Slug)
	}
}

func TestPostRepoListPublishedOnly(t *testing.T) {
	db := InitSQLite(filepath.Join(t.TempDir(), "test.db"))
	t.Cleanup(func() { _ = db.Close() })
	AutoMigrate(db)
	repos := New(db)

	cat := &model.Category{
		Slug: "pub-cat", Name: "Pub Category",
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	if err := repos.Category.Create(cat); err != nil {
		t.Fatalf("create category: %v", err)
	}

	now := time.Now().Truncate(time.Millisecond)
	draft := &model.Post{
		CategoryID: cat.ID, Title: "Draft Post", Slug: "draft-post",
		ContentMD: "draft content", Excerpt: "draft excerpt",
		Published: false, CreatedAt: now.Add(-1 * time.Hour), UpdatedAt: now,
	}
	published := &model.Post{
		CategoryID: cat.ID, Title: "Published Post", Slug: "published-post",
		ContentMD: "published content", Excerpt: "published excerpt",
		Published: true, CreatedAt: now, UpdatedAt: now,
	}
	for _, p := range []*model.Post{draft, published} {
		if err := repos.Post.Create(p); err != nil {
			t.Fatalf("create post %q: %v", p.Slug, err)
		}
	}

	result, err := repos.Post.List("", true, 10, 0)
	if err != nil {
		t.Fatalf("List published only: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 published post, got %d", len(result))
	}
	if result[0].Slug != "published-post" {
		t.Errorf("expected slug 'published-post', got %q", result[0].Slug)
	}
}

func TestPostRepoListWithOptionsQuery(t *testing.T) {
	db := InitSQLite(filepath.Join(t.TempDir(), "test.db"))
	t.Cleanup(func() { _ = db.Close() })
	AutoMigrate(db)
	repos := New(db)

	cat := &model.Category{
		Slug: "query-cat", Name: "Query Category",
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	if err := repos.Category.Create(cat); err != nil {
		t.Fatalf("create category: %v", err)
	}

	now := time.Now().Truncate(time.Millisecond)
	for _, p := range []*model.Post{
		{CategoryID: cat.ID, Title: "Golang Tutorial", Slug: "golang-tutorial", ContentMD: "Learn Go programming", Excerpt: "Go intro", Published: true, CreatedAt: now.Add(-2 * time.Minute), UpdatedAt: now},
		{CategoryID: cat.ID, Title: "Rust Basics", Slug: "rust-basics", ContentMD: "Learn Rust programming", Excerpt: "Rust intro", Published: true, CreatedAt: now.Add(-1 * time.Minute), UpdatedAt: now},
		{CategoryID: cat.ID, Title: "Advanced Go Patterns", Slug: "advanced-go", ContentMD: "Deep Go concepts", Excerpt: "Go advanced", Published: true, CreatedAt: now, UpdatedAt: now},
	} {
		if err := repos.Post.Create(p); err != nil {
			t.Fatalf("create post %q: %v", p.Slug, err)
		}
	}

	result, err := repos.Post.ListWithOptions(PostListOptions{Query: "Golang", Limit: 10, Offset: 0})
	if err != nil {
		t.Fatalf("ListWithOptions with query: %v", err)
	}
	if len(result) == 0 {
		t.Fatal("expected at least one result for query 'Golang'")
	}
	for _, p := range result {
		if !strings.Contains(p.Title, "Golang") &&
			!strings.Contains(p.ContentMD, "Golang") &&
			!strings.Contains(p.Excerpt, "Golang") {
			t.Errorf("post %q matched query but no field contains 'Golang'", p.Slug)
		}
	}
}

func TestPostRepoListWithOptionsQueryCaseInsensitive(t *testing.T) {
	db := InitSQLite(filepath.Join(t.TempDir(), "test.db"))
	t.Cleanup(func() { _ = db.Close() })
	AutoMigrate(db)
	repos := New(db)

	cat := &model.Category{
		Slug: "ci-cat", Name: "CI Category",
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	if err := repos.Category.Create(cat); err != nil {
		t.Fatalf("create category: %v", err)
	}

	now := time.Now().Truncate(time.Millisecond)
	p := &model.Post{
		CategoryID: cat.ID, Title: "Golang Dependency Injection", Slug: "go-di",
		ContentMD: "Using wire for DI", Excerpt: "DI in Go",
		Published: true, CreatedAt: now, UpdatedAt: now,
	}
	if err := repos.Post.Create(p); err != nil {
		t.Fatalf("create post: %v", err)
	}

	result, err := repos.Post.ListWithOptions(PostListOptions{Query: "dependency", Limit: 10, Offset: 0})
	if err != nil {
		t.Fatalf("ListWithOptions with lowercase query: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 result for case-insensitive query, got %d", len(result))
	}
	if result[0].Slug != "go-di" {
		t.Errorf("expected slug 'go-di', got %q", result[0].Slug)
	}
}

func TestPostRepoUpdate(t *testing.T) {
	db := InitSQLite(filepath.Join(t.TempDir(), "test.db"))
	t.Cleanup(func() { _ = db.Close() })
	AutoMigrate(db)
	repos := New(db)

	cat := &model.Category{
		Slug: "upd-cat", Name: "Update Category",
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	if err := repos.Category.Create(cat); err != nil {
		t.Fatalf("create category: %v", err)
	}

	post := &model.Post{
		CategoryID: cat.ID,
		Title:      "Original Title",
		Slug:       "original-title",
		ContentMD:  "original content",
		Excerpt:    "original excerpt",
		Published:  true,
	}
	if err := repos.Post.Create(post); err != nil {
		t.Fatalf("create post: %v", err)
	}

	updatedAt := time.Now().Truncate(time.Second)
	post.Title = "Updated Title"
	post.ContentMD = "updated content"
	post.Excerpt = "updated excerpt"
	post.Published = false
	post.UpdatedAt = updatedAt

	if err := repos.Post.Update(post); err != nil {
		t.Fatalf("Update: %v", err)
	}

	loaded, err := repos.Post.GetByID(post.ID)
	if err != nil {
		t.Fatalf("GetByID after update: %v", err)
	}

	if loaded.Title != "Updated Title" {
		t.Errorf("Title: got %q, want %q", loaded.Title, "Updated Title")
	}
	if loaded.ContentMD != "updated content" {
		t.Errorf("ContentMD: got %q, want %q", loaded.ContentMD, "updated content")
	}
	if loaded.Excerpt != "updated excerpt" {
		t.Errorf("Excerpt: got %q, want %q", loaded.Excerpt, "updated excerpt")
	}
	if loaded.Published {
		t.Error("Published should be false after update")
	}
	if loaded.Slug != "original-title" {
		t.Errorf("Slug changed after update: got %q, want %q", loaded.Slug, "original-title")
	}
	if !loaded.UpdatedAt.Equal(updatedAt) {
		t.Errorf("UpdatedAt: got %v, want %v", loaded.UpdatedAt, updatedAt)
	}
	if loaded.CreatedAt.IsZero() {
		t.Error("CreatedAt should not be zero")
	}
}

func TestPostRepoDelete(t *testing.T) {
	db := InitSQLite(filepath.Join(t.TempDir(), "test.db"))
	t.Cleanup(func() { _ = db.Close() })
	AutoMigrate(db)
	repos := New(db)

	cat := &model.Category{
		Slug: "del-cat", Name: "Delete Category",
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	if err := repos.Category.Create(cat); err != nil {
		t.Fatalf("create category: %v", err)
	}

	now := time.Now().Truncate(time.Millisecond)
	p := &model.Post{
		CategoryID: cat.ID, Title: "Delete Me", Slug: "delete-me",
		ContentMD: "to be deleted", Excerpt: "bye",
		Published: true, CreatedAt: now, UpdatedAt: now,
	}
	if err := repos.Post.Create(p); err != nil {
		t.Fatalf("create post: %v", err)
	}

	if err := repos.Post.Delete(p.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	loaded, err := repos.Post.GetByID(p.ID)
	if err == nil {
		t.Fatal("expected error after deleting post")
	}
	if err != sql.ErrNoRows {
		t.Fatalf("expected sql.ErrNoRows, got %v", err)
	}
	if loaded != nil {
		t.Fatal("expected nil post after deletion")
	}
}

func TestPostRepoSlugExists(t *testing.T) {
	db := InitSQLite(filepath.Join(t.TempDir(), "test.db"))
	t.Cleanup(func() { _ = db.Close() })
	AutoMigrate(db)
	repos := New(db)

	cat := &model.Category{
		Slug: "slug-ex-cat", Name: "Slug Exists Category",
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	if err := repos.Category.Create(cat); err != nil {
		t.Fatalf("create category: %v", err)
	}

	now := time.Now().Truncate(time.Millisecond)
	p := &model.Post{
		CategoryID: cat.ID, Title: "Slug Check", Slug: "slug-check",
		ContentMD: "content", Excerpt: "excerpt",
		Published: true, CreatedAt: now, UpdatedAt: now,
	}
	if err := repos.Post.Create(p); err != nil {
		t.Fatalf("create post: %v", err)
	}

	exists, err := repos.Post.SlugExists("slug-check")
	if err != nil {
		t.Fatalf("SlugExists for existing slug: %v", err)
	}
	if !exists {
		t.Error("SlugExists returned false for existing slug 'slug-check'")
	}

	notExists, err := repos.Post.SlugExists("non-existent-slug")
	if err != nil {
		t.Fatalf("SlugExists for non-existent slug: %v", err)
	}
	if notExists {
		t.Error("SlugExists returned true for non-existent slug")
	}
}

func TestPostRepoCount(t *testing.T) {
	db := InitSQLite(filepath.Join(t.TempDir(), "test.db"))
	t.Cleanup(func() { _ = db.Close() })
	AutoMigrate(db)
	repos := New(db)

	cat := &model.Category{
		Slug: "cnt-cat", Name: "Count Category",
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	if err := repos.Category.Create(cat); err != nil {
		t.Fatalf("create category: %v", err)
	}

	now := time.Now().Truncate(time.Millisecond)
	for i := 0; i < 5; i++ {
		p := &model.Post{
			CategoryID: cat.ID,
			Title:      "Count Post",
			Slug:       "count-post-" + string(rune('0'+i)),
			ContentMD:  "content",
			Excerpt:    "excerpt",
			Published:  true,
			CreatedAt:  now,
			UpdatedAt:  now,
		}
		if err := repos.Post.Create(p); err != nil {
			t.Fatalf("create post %d: %v", i, err)
		}
	}

	count, err := repos.Post.Count("", false)
	if err != nil {
		t.Fatalf("Count: %v", err)
	}
	if count != 5 {
		t.Errorf("expected count 5, got %d", count)
	}
}

func TestPostRepoCountWithCategoryFilter(t *testing.T) {
	db := InitSQLite(filepath.Join(t.TempDir(), "test.db"))
	t.Cleanup(func() { _ = db.Close() })
	AutoMigrate(db)
	repos := New(db)

	catA := &model.Category{
		Slug: "cnt-a", Name: "Count A",
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	catB := &model.Category{
		Slug: "cnt-b", Name: "Count B",
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	if err := repos.Category.Create(catA); err != nil {
		t.Fatalf("create category A: %v", err)
	}
	if err := repos.Category.Create(catB); err != nil {
		t.Fatalf("create category B: %v", err)
	}

	now := time.Now().Truncate(time.Millisecond)
	for i := 0; i < 3; i++ {
		p := &model.Post{
			CategoryID: catA.ID, Title: "A Post", Slug: "a-post-" + string(rune('0'+i)),
			ContentMD: "c", Excerpt: "e", Published: true, CreatedAt: now, UpdatedAt: now,
		}
		if err := repos.Post.Create(p); err != nil {
			t.Fatalf("create post in A: %v", err)
		}
	}
	for i := 0; i < 2; i++ {
		p := &model.Post{
			CategoryID: catB.ID, Title: "B Post", Slug: "b-post-" + string(rune('0'+i)),
			ContentMD: "c", Excerpt: "e", Published: true, CreatedAt: now, UpdatedAt: now,
		}
		if err := repos.Post.Create(p); err != nil {
			t.Fatalf("create post in B: %v", err)
		}
	}

	countA, err := repos.Post.Count("cnt-a", false)
	if err != nil {
		t.Fatalf("Count with category filter: %v", err)
	}
	if countA != 3 {
		t.Errorf("expected 3 posts in cnt-a, got %d", countA)
	}

	countB, err := repos.Post.Count("cnt-b", false)
	if err != nil {
		t.Fatalf("Count with category filter cnt-b: %v", err)
	}
	if countB != 2 {
		t.Errorf("expected 2 posts in cnt-b, got %d", countB)
	}
}

func TestPostRepoCountWithPublishedFilter(t *testing.T) {
	db := InitSQLite(filepath.Join(t.TempDir(), "test.db"))
	t.Cleanup(func() { _ = db.Close() })
	AutoMigrate(db)
	repos := New(db)

	cat := &model.Category{
		Slug: "cnt-pub", Name: "Count Pub",
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	if err := repos.Category.Create(cat); err != nil {
		t.Fatalf("create category: %v", err)
	}

	now := time.Now().Truncate(time.Millisecond)
	posts := []*model.Post{
		{CategoryID: cat.ID, Title: "Pub 1", Slug: "pub-1", ContentMD: "c", Excerpt: "e", Published: true, CreatedAt: now.Add(-2 * time.Minute), UpdatedAt: now},
		{CategoryID: cat.ID, Title: "Pub 2", Slug: "pub-2", ContentMD: "c", Excerpt: "e", Published: true, CreatedAt: now.Add(-1 * time.Minute), UpdatedAt: now},
		{CategoryID: cat.ID, Title: "Draft", Slug: "draft", ContentMD: "c", Excerpt: "e", Published: false, CreatedAt: now, UpdatedAt: now},
	}
	for _, p := range posts {
		if err := repos.Post.Create(p); err != nil {
			t.Fatalf("create post %q: %v", p.Slug, err)
		}
	}

	total, err := repos.Post.Count("", false)
	if err != nil {
		t.Fatalf("Count all: %v", err)
	}
	if total != 3 {
		t.Errorf("expected total count 3, got %d", total)
	}

	published, err := repos.Post.Count("", true)
	if err != nil {
		t.Fatalf("Count published only: %v", err)
	}
	if published != 2 {
		t.Errorf("expected published count 2, got %d", published)
	}
}

func TestPostRepoCountByCategoryID(t *testing.T) {
	db := InitSQLite(filepath.Join(t.TempDir(), "test.db"))
	t.Cleanup(func() { _ = db.Close() })
	AutoMigrate(db)
	repos := New(db)

	cat1 := &model.Category{
		Slug: "cntby-1", Name: "Count By 1",
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	cat2 := &model.Category{
		Slug: "cntby-2", Name: "Count By 2",
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	if err := repos.Category.Create(cat1); err != nil {
		t.Fatalf("create category 1: %v", err)
	}
	if err := repos.Category.Create(cat2); err != nil {
		t.Fatalf("create category 2: %v", err)
	}

	now := time.Now().Truncate(time.Millisecond)
	for i := 0; i < 3; i++ {
		p := &model.Post{
			CategoryID: cat1.ID, Title: "Cat1 Post", Slug: "cat1-post-" + string(rune('0'+i)),
			ContentMD: "c", Excerpt: "e", Published: true, CreatedAt: now, UpdatedAt: now,
		}
		if err := repos.Post.Create(p); err != nil {
			t.Fatalf("create post in cat1: %v", err)
		}
	}

	p := &model.Post{
		CategoryID: cat2.ID, Title: "Cat2 Post", Slug: "cat2-post",
		ContentMD: "c", Excerpt: "e", Published: true, CreatedAt: now, UpdatedAt: now,
	}
	if err := repos.Post.Create(p); err != nil {
		t.Fatalf("create post in cat2: %v", err)
	}

	count1, err := repos.Post.CountByCategoryID(cat1.ID)
	if err != nil {
		t.Fatalf("CountByCategoryID for cat1: %v", err)
	}
	if count1 != 3 {
		t.Errorf("expected count 3 for cat1, got %d", count1)
	}

	count2, err := repos.Post.CountByCategoryID(cat2.ID)
	if err != nil {
		t.Fatalf("CountByCategoryID for cat2: %v", err)
	}
	if count2 != 1 {
		t.Errorf("expected count 1 for cat2, got %d", count2)
	}

	countEmpty, err := repos.Post.CountByCategoryID(99999)
	if err != nil {
		t.Fatalf("CountByCategoryID for non-existent category: %v", err)
	}
	if countEmpty != 0 {
		t.Errorf("expected count 0 for non-existent category, got %d", countEmpty)
	}
}

func TestPostRepoCreateSetsID(t *testing.T) {
	db := InitSQLite(filepath.Join(t.TempDir(), "test.db"))
	t.Cleanup(func() { _ = db.Close() })
	AutoMigrate(db)
	repos := New(db)

	cat := &model.Category{
		Slug: "id-cat", Name: "ID Cat",
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	if err := repos.Category.Create(cat); err != nil {
		t.Fatalf("create category: %v", err)
	}

	now := time.Now().Truncate(time.Millisecond)
	p1 := &model.Post{
		CategoryID: cat.ID, Title: "First", Slug: "first-id",
		ContentMD: "c", Excerpt: "e", Published: true,
		CreatedAt: now, UpdatedAt: now,
	}
	p2 := &model.Post{
		CategoryID: cat.ID, Title: "Second", Slug: "second-id",
		ContentMD: "c", Excerpt: "e", Published: true,
		CreatedAt: now, UpdatedAt: now,
	}

	if err := repos.Post.Create(p1); err != nil {
		t.Fatalf("create first post: %v", err)
	}
	if err := repos.Post.Create(p2); err != nil {
		t.Fatalf("create second post: %v", err)
	}

	if p1.ID == p2.ID {
		t.Error("expected unique IDs for separate posts")
	}
	if p2.ID <= p1.ID {
		t.Errorf("expected second post ID (%d) > first post ID (%d)", p2.ID, p1.ID)
	}
}
