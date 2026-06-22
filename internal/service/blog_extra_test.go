package service

import (
	"errors"
	"strings"
	"strconv"
	"testing"

	"raevtar/internal/model"
)

// -- ListCategories ----------------------------------------------------------

func TestBlogServiceListCategories(t *testing.T) {
	state := newTestServices(t)

	categories, err := state.svc.Blog.ListCategories()
	if err != nil {
		t.Fatalf("ListCategories: %v", err)
	}
	if len(categories) < 5 {
		t.Fatalf("categories len = %d, want >= 5 (seeded)", len(categories))
	}

	var found bool
	for _, cat := range categories {
		if cat.Slug == "devops" {
			found = true
			if cat.Name != "DevOps" {
				t.Fatalf("devops name = %q, want DevOps", cat.Name)
			}
			break
		}
	}
	if !found {
		t.Fatalf("seeded category 'devops' not found in list")
	}
}

// -- GetCategoryByID ---------------------------------------------------------

func TestBlogServiceGetCategoryByID(t *testing.T) {
	state := newTestServices(t)

	devops, err := state.repos.Category.GetBySlug("devops")
	if err != nil {
		t.Fatalf("get devops category: %v", err)
	}

	cat, count, err := state.svc.Blog.GetCategoryByID(devops.ID)
	if err != nil {
		t.Fatalf("GetCategoryByID(%d): %v", devops.ID, err)
	}
	if cat == nil {
		t.Fatalf("GetCategoryByID returned nil category")
	}
	if cat.Slug != "devops" {
		t.Fatalf("category slug = %q, want devops", cat.Slug)
	}
	if cat.Name != "DevOps" {
		t.Fatalf("category name = %q, want DevOps", cat.Name)
	}
	if count != 0 {
		t.Fatalf("post count = %d, want 0 (no posts yet)", count)
	}
}

func TestBlogServiceGetCategoryByIDNotFound(t *testing.T) {
	state := newTestServices(t)

	cat, count, err := state.svc.Blog.GetCategoryByID(9999)
	if err == nil {
		t.Fatalf("expected error for non-existent category, got %+v count=%d", cat, count)
	}
	if !errors.Is(err, ErrCategoryNotFound) {
		t.Fatalf("err = %v, want ErrCategoryNotFound", err)
	}
	if cat != nil {
		t.Fatalf("expected nil category, got %+v", cat)
	}
	if count != 0 {
		t.Fatalf("count = %d, want 0", count)
	}
}

// -- PostCountForCategory ----------------------------------------------------

func TestBlogServicePostCountForCategory(t *testing.T) {
	state := newTestServices(t)

	devops, err := state.repos.Category.GetBySlug("devops")
	if err != nil {
		t.Fatalf("get devops category: %v", err)
	}

	// No posts yet — count should be 0
	count, err := state.svc.Blog.PostCountForCategory(devops.ID)
	if err != nil {
		t.Fatalf("PostCountForCategory: %v", err)
	}
	if count != 0 {
		t.Fatalf("count = %d, want 0", count)
	}

	// Create a post in this category
	_, err = state.svc.Blog.CreatePost(model.PostCreate{
		CategorySlug: "devops",
		Title:        "Count Test Post",
		ContentMD:    "# Count Test",
		Published:    true,
	})
	if err != nil {
		t.Fatalf("create post: %v", err)
	}

	count, err = state.svc.Blog.PostCountForCategory(devops.ID)
	if err != nil {
		t.Fatalf("PostCountForCategory after create: %v", err)
	}
	if count != 1 {
		t.Fatalf("count = %d, want 1", count)
	}
}

// -- DeleteCategory ----------------------------------------------------------

func TestBlogServiceDeleteCategory(t *testing.T) {
	state := newTestServices(t)

	cat, err := callCreateCategory(t, state, model.Category{
		Slug:        "temp-cat",
		Name:        "Temporary Category",
		Description: "Will be deleted",
	})
	if err != nil {
		t.Fatalf("create category: %v", err)
	}

	if err := state.svc.Blog.DeleteCategory(cat.ID); err != nil {
		t.Fatalf("DeleteCategory: %v", err)
	}

	// Verify deleted
	_, _, err = state.svc.Blog.GetCategoryByID(cat.ID)
	if err == nil {
		t.Fatalf("expected deleted category lookup to fail")
	}
	if !errors.Is(err, ErrCategoryNotFound) {
		t.Fatalf("err after delete = %v, want ErrCategoryNotFound", err)
	}
}

func TestBlogServiceDeleteCategoryInUse(t *testing.T) {
	state := newTestServices(t)

	cat, err := callCreateCategory(t, state, model.Category{
		Slug: "in-use-cat",
		Name: "Category In Use",
	})
	if err != nil {
		t.Fatalf("create category: %v", err)
	}

	_, err = state.svc.Blog.CreatePost(model.PostCreate{
		CategorySlug: cat.Slug,
		Title:        "Occupied Post",
		ContentMD:    "# Occupied",
		Published:    true,
	})
	if err != nil {
		t.Fatalf("create post in category: %v", err)
	}

	err = state.svc.Blog.DeleteCategory(cat.ID)
	if err == nil {
		t.Fatalf("expected DeleteCategory to fail for category with posts")
	}
	if !errors.Is(err, ErrCategoryInUse) {
		t.Fatalf("err = %v, want ErrCategoryInUse", err)
	}

	// Category should still exist
	_, _, err = state.svc.Blog.GetCategoryByID(cat.ID)
	if err != nil {
		t.Fatalf("category should still exist after failed delete: %v", err)
	}
}

func TestBlogServiceDeleteCategoryNotFound(t *testing.T) {
	state := newTestServices(t)

	err := state.svc.Blog.DeleteCategory(9999)
	if err == nil {
		t.Fatalf("expected DeleteCategory(9999) to fail")
	}
	if !errors.Is(err, ErrCategoryNotFound) {
		t.Fatalf("err = %v, want ErrCategoryNotFound", err)
	}
}

// -- uniqueSlug edge cases --------------------------------------------------

func TestUniqueSlugEdgeCases(t *testing.T) {
	state := newTestServices(t)

	// Fresh title — base slug
	slug, err := state.svc.Blog.uniqueSlug("Edge Case Title")
	if err != nil {
		t.Fatalf("uniqueSlug fresh: %v", err)
	}
	if slug != "edge-case-title" {
		t.Fatalf("slug = %q, want edge-case-title", slug)
	}

	// Create a post to occupy the base slug
	_, err = state.svc.Blog.CreatePost(model.PostCreate{
		CategorySlug: "devops",
		Title:        "Edge Case Title",
		ContentMD:    "# First",
		Published:    true,
	})
	if err != nil {
		t.Fatalf("create first collision post: %v", err)
	}

	// First collision ? -2 suffix
	slug2, err := state.svc.Blog.uniqueSlug("Edge Case Title")
	if err != nil {
		t.Fatalf("uniqueSlug first collision: %v", err)
	}
	if slug2 != "edge-case-title-2" {
		t.Fatalf("slug = %q, want edge-case-title-2", slug2)
	}

	// Occupy -2 by creating a second post
	_, err = state.svc.Blog.CreatePost(model.PostCreate{
		CategorySlug: "tools",
		Title:        "Edge Case Title",
		ContentMD:    "# Second",
		Published:    true,
	})
	if err != nil {
		t.Fatalf("create second collision post: %v", err)
	}

	// Second collision ? -3 suffix
	slug3, err := state.svc.Blog.uniqueSlug("Edge Case Title")
	if err != nil {
		t.Fatalf("uniqueSlug second collision: %v", err)
	}
	if slug3 != "edge-case-title-3" {
		t.Fatalf("slug = %q, want edge-case-title-3", slug3)
	}
}

// -- RenderMarkdown edge cases -----------------------------------------------

func TestBlogServiceRenderMarkdownEdgeCases(t *testing.T) {
	state := newTestServices(t)

	// Basic markdown produces valid HTML
	html, err := state.svc.Blog.RenderMarkdown("# Hello World")
	if err != nil {
		t.Fatalf("RenderMarkdown basic: %v", err)
	}
	if !strings.Contains(html, "<h1") || !strings.Contains(html, "Hello World") {
		t.Fatalf("basic markdown html = %q, want h1 with Hello World", html)
	}

	// Bold and italic
	html, err = state.svc.Blog.RenderMarkdown("**bold** and *italic*")
	if err != nil {
		t.Fatalf("RenderMarkdown emphasis: %v", err)
	}
	if !strings.Contains(html, "<strong>bold</strong>") || !strings.Contains(html, "<em>italic</em>") {
		t.Fatalf("emphasis html = %q, want strong/em tags", html)
	}

	// Link
	html, err = state.svc.Blog.RenderMarkdown("[link](https://example.com)")
	if err != nil {
		t.Fatalf("RenderMarkdown link: %v", err)
	}
	if !strings.Contains(html, `href="https://example.com"`) {
		t.Fatalf("link html = %q, want anchor with href", html)
	}

	// Shortcode processing: goldmark escapes raw HTML without html.WithUnsafe(),
	// but the shortcode text [[server-status:...]] should be removed from output.
	html, err = state.svc.Blog.RenderMarkdown("before [[server-status:node-1]] after")
	if err != nil {
		t.Fatalf("RenderMarkdown shortcode: %v", err)
	}
	if strings.Contains(html, "[[server-status:node-1]]") {
		t.Fatalf("shortcode syntax should have been replaced in output: %q", html)
	}
	if !strings.Contains(html, "node-1") {
		t.Fatalf("shortcode html = %q, want node-1 in output (replacement occurred)", html)
	}
}


// -- RecordPostView ----------------------------------------------------------

func TestBlogServiceRecordPostView(t *testing.T) {
	state := newTestServices(t)

	post, err := state.svc.Blog.CreatePost(model.PostCreate{
		CategorySlug: "devops",
		Title:        "View Test Post",
		ContentMD:    "# View Test",
		Excerpt:      "View test",
		Published:    true,
	})
	if err != nil {
		t.Fatalf("create post: %v", err)
	}

	if err := state.svc.Blog.RecordPostView(post.ID, "abc123hash"); err != nil {
		t.Fatalf("RecordPostView: %v", err)
	}
}

// -- PostViewCount -----------------------------------------------------------

func TestBlogServicePostViewCount(t *testing.T) {
	state := newTestServices(t)

	post, err := state.svc.Blog.CreatePost(model.PostCreate{
		CategorySlug: "devops",
		Title:        "View Count Post",
		ContentMD:    "# View Count",
		Excerpt:      "View count test",
		Published:    true,
	})
	if err != nil {
		t.Fatalf("create post: %v", err)
	}

	count, err := state.svc.Blog.PostViewCount(post.ID)
	if err != nil {
		t.Fatalf("PostViewCount before recording: %v", err)
	}
	if count != 0 {
		t.Fatalf("count before recording = %d, want 0", count)
	}

	if err := state.svc.Blog.RecordPostView(post.ID, "hash-1"); err != nil {
		t.Fatalf("RecordPostView 1: %v", err)
	}

	count, err = state.svc.Blog.PostViewCount(post.ID)
	if err != nil {
		t.Fatalf("PostViewCount after recording: %v", err)
	}
	if count != 1 {
		t.Fatalf("count after recording = %d, want 1", count)
	}
}

// -- AllPostViewCounts -------------------------------------------------------

func TestBlogServiceAllPostViewCounts(t *testing.T) {
	state := newTestServices(t)

	postA, err := state.svc.Blog.CreatePost(model.PostCreate{
		CategorySlug: "devops",
		Title:        "Post A Views",
		ContentMD:    "# Post A",
		Excerpt:      "Post A excerpt",
		Published:    true,
	})
	if err != nil {
		t.Fatalf("create post A: %v", err)
	}

	postB, err := state.svc.Blog.CreatePost(model.PostCreate{
		CategorySlug: "tools",
		Title:        "Post B Views",
		ContentMD:    "# Post B",
		Excerpt:      "Post B excerpt",
		Published:    true,
	})
	if err != nil {
		t.Fatalf("create post B: %v", err)
	}

	for i := 0; i < 3; i++ {
		if err := state.svc.Blog.RecordPostView(postA.ID, "hash-a-"+strconv.Itoa(i)); err != nil {
			t.Fatalf("RecordPostView post A %d: %v", i, err)
		}
	}
	if err := state.svc.Blog.RecordPostView(postB.ID, "hash-b-1"); err != nil {
		t.Fatalf("RecordPostView post B: %v", err)
	}

	views, err := state.svc.Blog.AllPostViewCounts()
	if err != nil {
		t.Fatalf("AllPostViewCounts: %v", err)
	}

	if views[postA.ID] != 3 {
		t.Fatalf("post A views = %d, want 3", views[postA.ID])
	}
	if views[postB.ID] != 1 {
		t.Fatalf("post B views = %d, want 1", views[postB.ID])
	}
}

func TestBlogServiceAllPostViewCountsEmpty(t *testing.T) {
	state := newTestServices(t)

	views, err := state.svc.Blog.AllPostViewCounts()
	if err != nil {
		t.Fatalf("AllPostViewCounts empty: %v", err)
	}
	if len(views) != 0 {
		t.Fatalf("views len = %d, want 0", len(views))
	}
}

