package service

import (
	"strings"
	"testing"

	"raevtar/internal/model"
)

// ── GetPage ─────────────────────────────────────────────────────────────────

func TestPageContentServiceGetPage(t *testing.T) {
	state := newTestServices(t)

	page, err := state.svc.Pages.GetPage(model.PageKeyAbout)
	if err != nil {
		t.Fatalf("GetPage(about): %v", err)
	}
	if page == nil {
		t.Fatalf("GetPage returned nil page")
	}
	if page.Key != model.PageKeyAbout {
		t.Fatalf("page key = %q, want about", page.Key)
	}
	if page.Title == "" {
		t.Fatalf("page title is empty")
	}
	if page.ContentMD == "" {
		t.Fatalf("page content_md is empty")
	}
	if page.ContentHTML == "" {
		t.Fatalf("page content_html is empty (should be rendered)")
	}
	if !strings.Contains(page.ContentHTML, "<p>") {
		t.Fatalf("content_html = %q, expected paragraph HTML", page.ContentHTML)
	}
}

func TestPageContentServiceGetPageNotFound(t *testing.T) {
	state := newTestServices(t)

	_, err := state.svc.Pages.GetPage("nonexistent-key")
	if err == nil {
		t.Fatalf("expected error for non-existent page key")
	}
	if !strings.Contains(err.Error(), "nonexistent-key") {
		t.Fatalf("err = %v, want containing key name", err)
	}
}

// ── UpdatePage ──────────────────────────────────────────────────────────────

func TestPageContentServiceUpdatePage(t *testing.T) {
	state := newTestServices(t)

	updated, err := state.svc.Pages.UpdatePage(model.PageContent{
		Key:       model.PageKeyAbout,
		Title:     "Updated About Title",
		Summary:   "Updated summary text",
		ContentMD: "# Updated Content\n\nNew body here.",
	})
	if err != nil {
		t.Fatalf("UpdatePage: %v", err)
	}
	if updated.Key != model.PageKeyAbout {
		t.Fatalf("updated key = %q, want about", updated.Key)
	}
	if updated.Title != "Updated About Title" {
		t.Fatalf("updated title = %q, want Updated About Title", updated.Title)
	}
	if updated.Summary != "Updated summary text" {
		t.Fatalf("updated summary = %q, want Updated summary text", updated.Summary)
	}
	if updated.ContentHTML == "" || !strings.Contains(updated.ContentHTML, "Updated Content") {
		t.Fatalf("updated content_html = %q, want rendered HTML with updated content", updated.ContentHTML)
	}
}

func TestPageContentServiceUpdatePageCreateNew(t *testing.T) {
	state := newTestServices(t)

	// Update a non-existent key — Upsert creates it
	updated, err := state.svc.Pages.UpdatePage(model.PageContent{
		Key:       "new-page",
		Title:     "Brand New Page",
		Summary:   "Created via UpdatePage",
		ContentMD: "# New Page\n\nFresh content.",
	})
	if err != nil {
		t.Fatalf("UpdatePage new key: %v", err)
	}
	if updated.Key != "new-page" {
		t.Fatalf("key = %q, want new-page", updated.Key)
	}

	// Verify it persisted
	fetched, err := state.svc.Pages.GetPage("new-page")
	if err != nil {
		t.Fatalf("GetPage after create: %v", err)
	}
	if fetched.Title != "Brand New Page" {
		t.Fatalf("fetched title = %q, want Brand New Page", fetched.Title)
	}
	if fetched.Summary != "Created via UpdatePage" {
		t.Fatalf("fetched summary = %q, want Created via UpdatePage", fetched.Summary)
	}
}

func TestPageContentServiceUpdatePageEmptyContent(t *testing.T) {
	state := newTestServices(t)

	_, err := state.svc.Pages.UpdatePage(model.PageContent{
		Key:       model.PageKeyAbout,
		Title:     "Title Only",
		Summary:   "Summary",
		ContentMD: "",
	})
	if err == nil {
		t.Fatalf("expected error for empty content_md")
	}
	if !strings.Contains(err.Error(), "content") {
		t.Fatalf("err = %v, want error containing 'content'", err)
	}
}

func TestPageContentServiceUpdatePageEmptyTitle(t *testing.T) {
	state := newTestServices(t)

	_, err := state.svc.Pages.UpdatePage(model.PageContent{
		Key:       model.PageKeyAbout,
		Title:     "",
		Summary:   "Summary",
		ContentMD: "# Some content",
	})
	if err == nil {
		t.Fatalf("expected error for empty title")
	}
	if !strings.Contains(err.Error(), "title") {
		t.Fatalf("err = %v, want error containing 'title'", err)
	}
}

func TestPageContentServiceUpdatePageEmptyKey(t *testing.T) {
	state := newTestServices(t)

	_, err := state.svc.Pages.UpdatePage(model.PageContent{
		Key:       "",
		Title:     "No Key",
		Summary:   "Summary",
		ContentMD: "# Content",
	})
	if err == nil {
		t.Fatalf("expected error for empty key")
	}
	if !strings.Contains(err.Error(), "key") {
		t.Fatalf("err = %v, want error containing 'key'", err)
	}
}

// ── ListPages ──────────────────────────────────────────────────────────────

func TestPageContentServiceListPages(t *testing.T) {
	state := newTestServices(t)

	pages, err := state.svc.Pages.ListPages()
	if err != nil {
		t.Fatalf("ListPages: %v", err)
	}
	if len(pages) < 2 {
		t.Fatalf("pages len = %d, want at least 2 (seeded about + contact)", len(pages))
	}

	// Verify seeded pages are present
	found := map[string]bool{}
	for _, p := range pages {
		found[p.Key] = true
	}
	if !found[model.PageKeyAbout] {
		t.Fatalf("ListPages missing seeded about page")
	}
	if !found[model.PageKeyContact] {
		t.Fatalf("ListPages missing seeded contact page")
	}
	// Verify about page has rendered HTML
	for _, p := range pages {
		if p.Key == model.PageKeyAbout {
			if p.ContentHTML == "" {
				t.Fatalf("about page content_html is empty (should be rendered)")
			}
		}
	}
}

func TestPageContentServiceListPagesIncludesNewPages(t *testing.T) {
	state := newTestServices(t)

	// Create a new page via UpdatePage (Upsert)
	_, err := state.svc.Pages.UpdatePage(model.PageContent{
		Key:       "custom-page",
		Title:     "Custom Page",
		Summary:   "A custom page created after seed",
		ContentMD: "# Custom\n\nContent.",
	})
	if err != nil {
		t.Fatalf("UpdatePage custom page: %v", err)
	}

	pages, err := state.svc.Pages.ListPages()
	if err != nil {
		t.Fatalf("ListPages after add: %v", err)
	}
	if len(pages) < 3 {
		t.Fatalf("pages len = %d, want at least 3 (seeded 2 + custom)", len(pages))
	}

	found := false
	for _, p := range pages {
		if p.Key == "custom-page" {
			found = true
			if p.Title != "Custom Page" {
				t.Fatalf("custom page title = %q, want Custom Page", p.Title)
			}
			if p.ContentHTML == "" {
				t.Fatalf("custom page content_html is empty (should be rendered)")
			}
			if p.Summary != "A custom page created after seed" {
				t.Fatalf("custom page summary = %q, want summary text", p.Summary)
			}
		}
	}
	if !found {
		t.Fatalf("ListPages missing custom-page")
	}
}
