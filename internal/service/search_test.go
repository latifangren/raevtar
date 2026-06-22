package service

import (
	"fmt"
	"testing"

	"raevtar/internal/model"
)

func TestSearchServicePublicSearchFindsPosts(t *testing.T) {
	state := newTestServices(t)

	_, err := state.svc.Blog.CreatePost(model.PostCreate{
		CategorySlug: "devops",
		Title:        "Unique Post Finding",
		ContentMD:    "# Unique Post Finding\n\nSearch test content.",
		Excerpt:      "excerpt for unique-post-finding",
		Published:    true,
	})
	if err != nil {
		t.Fatalf("create post: %v", err)
	}

	results, err := state.svc.Search.SearchPublic(SearchOptions{
		Query:    "Unique Post Finding",
		Scope:    SearchScopeAll,
		Page:     1,
		PageSize: 10,
	})
	if err != nil {
		t.Fatalf("search all: %v", err)
	}
	if results.PostCount != 1 {
		t.Fatalf("post count = %d, want 1", results.PostCount)
	}
	if len(results.Posts) != 1 {
		t.Fatalf("posts len = %d, want 1", len(results.Posts))
	}
	if results.Posts[0].Title != "Unique Post Finding" {
		t.Fatalf("post title = %q, want Unique Post Finding", results.Posts[0].Title)
	}
}

func TestSearchServicePublicSearchFindsProjects(t *testing.T) {
	state := newTestServices(t)

	_, err := state.svc.Projects.CreateProject(model.ProjectCreate{
		Title:     "Unique Project Searchable",
		ContentMD: "# Unique Project Searchable",
		Excerpt:   "excerpt for unique-project-searchable",
		Published: true,
		State:     model.ProjectStateActive,
	})
	if err != nil {
		t.Fatalf("create project: %v", err)
	}

	results, err := state.svc.Search.SearchPublic(SearchOptions{
		Query:    "Unique Project Searchable",
		Scope:    SearchScopeAll,
		Page:     1,
		PageSize: 10,
	})
	if err != nil {
		t.Fatalf("search all: %v", err)
	}
	if results.ProjectCount != 1 {
		t.Fatalf("project count = %d, want 1", results.ProjectCount)
	}
	if len(results.Projects) != 1 {
		t.Fatalf("projects len = %d, want 1", len(results.Projects))
	}
	if results.Projects[0].Title != "Unique Project Searchable" {
		t.Fatalf("project title = %q, want Unique Project Searchable", results.Projects[0].Title)
	}
}

func TestSearchServicePublicSearchFindsPages(t *testing.T) {
	state := newTestServices(t)

	_, err := state.svc.Pages.UpdatePage(model.PageContent{
		Key:       model.PageKeyAbout,
		Title:     "About Unique Page Search Test",
		Summary:   "summary for unique-page-search",
		ContentMD: "# About\n\nThis page has unique page search content.",
	})
	if err != nil {
		t.Fatalf("update about page: %v", err)
	}

	results, err := state.svc.Search.SearchPublic(SearchOptions{
		Query:    "unique page search content",
		Scope:    SearchScopeAll,
		Page:     1,
		PageSize: 10,
	})
	if err != nil {
		t.Fatalf("search all: %v", err)
	}
	if results.PageCount != 1 {
		t.Fatalf("page count = %d, want 1", results.PageCount)
	}
	if len(results.Pages) != 1 {
		t.Fatalf("pages len = %d, want 1", len(results.Pages))
	}
	if results.Pages[0].Key != model.PageKeyAbout {
		t.Fatalf("page key = %q, want %q", results.Pages[0].Key, model.PageKeyAbout)
	}
}

func TestSearchServiceEmptyQueryReturnsNoResults(t *testing.T) {
	state := newTestServices(t)

	results, err := state.svc.Search.SearchPublic(SearchOptions{
		Query:    "",
		Scope:    SearchScopeAll,
		Page:     1,
		PageSize: 10,
	})
	if err != nil {
		t.Fatalf("empty search: %v", err)
	}
	if results.Total != 0 {
		t.Fatalf("total = %d, want 0", results.Total)
	}
	if len(results.Posts) != 0 {
		t.Fatalf("posts len = %d, want 0", len(results.Posts))
	}
	if len(results.Projects) != 0 {
		t.Fatalf("projects len = %d, want 0", len(results.Projects))
	}
	if len(results.Pages) != 0 {
		t.Fatalf("pages len = %d, want 0", len(results.Pages))
	}

	results, err = state.svc.Search.SearchPublic(SearchOptions{
		Query:    "   ",
		Scope:    SearchScopeAll,
		Page:     1,
		PageSize: 10,
	})
	if err != nil {
		t.Fatalf("whitespace search: %v", err)
	}
	if results.Total != 0 {
		t.Fatalf("whitespace total = %d, want 0", results.Total)
	}
}

func TestSearchServiceScopePostsFiltersCorrectly(t *testing.T) {
	state := newTestServices(t)

	_, err := state.svc.Blog.CreatePost(model.PostCreate{
		CategorySlug: "devops",
		Title:        "Search Scope Test Post",
		ContentMD:    "# Search Scope Test Post",
		Excerpt:      "scope-filter-test",
		Published:    true,
	})
	if err != nil {
		t.Fatalf("create post: %v", err)
	}
	_, err = state.svc.Projects.CreateProject(model.ProjectCreate{
		Title:     "Search Scope Test Project",
		ContentMD: "# Search Scope Test Project",
		Excerpt:   "scope-filter-test",
		Published: true,
		State:     model.ProjectStateActive,
	})
	if err != nil {
		t.Fatalf("create project: %v", err)
	}

	postsOnly, err := state.svc.Search.SearchPublic(SearchOptions{
		Query:    "scope-filter-test",
		Scope:    SearchScopePosts,
		Page:     1,
		PageSize: 10,
	})
	if err != nil {
		t.Fatalf("search posts scope: %v", err)
	}
	if postsOnly.PostCount != 1 {
		t.Fatalf("post count = %d, want 1", postsOnly.PostCount)
	}
	if len(postsOnly.Posts) != 1 {
		t.Fatalf("posts len = %d, want 1", len(postsOnly.Posts))
	}
	if postsOnly.ProjectCount != 0 {
		t.Fatalf("project count = %d, want 0", postsOnly.ProjectCount)
	}
	if postsOnly.PageCount != 0 {
		t.Fatalf("page count = %d, want 0", postsOnly.PageCount)
	}
}

func TestSearchServiceScopeProjectsFiltersCorrectly(t *testing.T) {
	state := newTestServices(t)

	_, err := state.svc.Blog.CreatePost(model.PostCreate{
		CategorySlug: "devops",
		Title:        "Scope Projects Test Post",
		ContentMD:    "# Scope Projects Test Post",
		Excerpt:      "scope-projects-filter",
		Published:    true,
	})
	if err != nil {
		t.Fatalf("create post: %v", err)
	}
	_, err = state.svc.Projects.CreateProject(model.ProjectCreate{
		Title:     "Scope Projects Test Project",
		ContentMD: "# Scope Projects Test Project",
		Excerpt:   "scope-projects-filter",
		Published: true,
		State:     model.ProjectStateActive,
	})
	if err != nil {
		t.Fatalf("create project: %v", err)
	}

	projectsOnly, err := state.svc.Search.SearchPublic(SearchOptions{
		Query:    "scope-projects-filter",
		Scope:    SearchScopeProjects,
		Page:     1,
		PageSize: 10,
	})
	if err != nil {
		t.Fatalf("search projects scope: %v", err)
	}
	if projectsOnly.ProjectCount != 1 {
		t.Fatalf("project count = %d, want 1", projectsOnly.ProjectCount)
	}
	if len(projectsOnly.Projects) != 1 {
		t.Fatalf("projects len = %d, want 1", len(projectsOnly.Projects))
	}
	if len(projectsOnly.Posts) != 0 {
		t.Fatalf("posts len = %d, want 0", len(projectsOnly.Posts))
	}
	if projectsOnly.PageCount != 0 {
		t.Fatalf("page count = %d, want 0", projectsOnly.PageCount)
	}
}

func TestSearchServicePagination(t *testing.T) {
	state := newTestServices(t)

	for i := 0; i < 15; i++ {
		_, err := state.svc.Blog.CreatePost(model.PostCreate{
			CategorySlug: "devops",
			Title:        fmt.Sprintf("Pagination Search Article %d", i+1),
			ContentMD:    "# Article " + fmt.Sprint(i+1),
			Excerpt:      "pagination-count-seed",
			Published:    true,
		})
		if err != nil {
			t.Fatalf("create post %d: %v", i+1, err)
		}
	}

	first, err := state.svc.Search.SearchPublic(SearchOptions{
		Query:    "pagination-count-seed",
		Scope:    SearchScopePosts,
		Page:     1,
		PageSize: 10,
	})
	if err != nil {
		t.Fatalf("search page 1: %v", err)
	}
	if first.PostCount != 15 {
		t.Fatalf("post count = %d, want 15", first.PostCount)
	}
	if len(first.Posts) != 10 {
		t.Fatalf("page 1 posts len = %d, want 10", len(first.Posts))
	}

	second, err := state.svc.Search.SearchPublic(SearchOptions{
		Query:    "pagination-count-seed",
		Scope:    SearchScopePosts,
		Page:     2,
		PageSize: 10,
	})
	if err != nil {
		t.Fatalf("search page 2: %v", err)
	}
	if second.PostCount != 15 {
		t.Fatalf("page 2 post count = %d, want 15", second.PostCount)
	}
	if len(second.Posts) != 5 {
		t.Fatalf("page 2 posts len = %d, want 5", len(second.Posts))
	}
	if second.Posts[0].Title == first.Posts[0].Title {
		t.Fatalf("page 2 first title = %q, should differ from page 1 first title %q", second.Posts[0].Title, first.Posts[0].Title)
	}
}

func TestSearchServiceCaseInsensitive(t *testing.T) {
	state := newTestServices(t)

	_, err := state.svc.Blog.CreatePost(model.PostCreate{
		CategorySlug: "devops",
		Title:        "Case Insensitive Search Test",
		ContentMD:    "# Case Insensitive",
		Excerpt:      "CaseInsensitiveMatchFindMe",
		Published:    true,
	})
	if err != nil {
		t.Fatalf("create post: %v", err)
	}

	lower, err := state.svc.Search.SearchPublic(SearchOptions{
		Query:    "caseinsensitivematchfindme",
		Scope:    SearchScopePosts,
		Page:     1,
		PageSize: 10,
	})
	if err != nil {
		t.Fatalf("lowercase search: %v", err)
	}
	if lower.PostCount != 1 {
		t.Fatalf("lowercase post count = %d, want 1", lower.PostCount)
	}

	upper, err := state.svc.Search.SearchPublic(SearchOptions{
		Query:    "CASEINSENSITIVEMATCHFINDME",
		Scope:    SearchScopePosts,
		Page:     1,
		PageSize: 10,
	})
	if err != nil {
		t.Fatalf("uppercase search: %v", err)
	}
	if upper.PostCount != 1 {
		t.Fatalf("uppercase post count = %d, want 1", upper.PostCount)
	}

	mixed, err := state.svc.Search.SearchPublic(SearchOptions{
		Query:    "CaseInsensitiveMatch",
		Scope:    SearchScopePosts,
		Page:     1,
		PageSize: 10,
	})
	if err != nil {
		t.Fatalf("mixed case search: %v", err)
	}
	if mixed.PostCount != 1 {
		t.Fatalf("mixed case post count = %d, want 1", mixed.PostCount)
	}
}
