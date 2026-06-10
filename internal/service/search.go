package service

import (
	"fmt"
	"strings"

	"raevtar/internal/model"
	"raevtar/internal/repo"
)

const (
	SearchScopeAll      = "all"
	SearchScopePosts    = "posts"
	SearchScopeProjects = "projects"
	SearchScopePages    = "pages"
)

type SearchService struct {
	blog     *BlogService
	projects *ProjectService
	pages    *PageContentService
	repos    *repo.Repositories
}

type SearchOptions struct {
	Query    string
	Scope    string
	Page     int
	PageSize int
}

type SearchResults struct {
	Query        string
	Scope        string
	Page         int
	PageSize     int
	Total        int
	TotalPages   int
	Paginated    bool
	Posts        []model.Post
	Projects     []model.Project
	Pages        []model.PageContent
	PostCount    int
	ProjectCount int
	PageCount    int
}

func NewSearchService(repos *repo.Repositories, blog *BlogService, projects *ProjectService, pages *PageContentService) *SearchService {
	return &SearchService{repos: repos, blog: blog, projects: projects, pages: pages}
}

func (s *SearchService) SearchPublic(opts SearchOptions) (SearchResults, error) {
	query := strings.TrimSpace(opts.Query)
	if opts.Page < 1 {
		opts.Page = 1
	}
	if opts.PageSize <= 0 {
		opts.PageSize = 10
	}
	scope, err := normalizeSearchScope(opts.Scope)
	if err != nil {
		return SearchResults{}, err
	}
	results := SearchResults{Query: query, Scope: scope, Page: opts.Page, PageSize: opts.PageSize}
	if query == "" {
		return results, nil
	}
	page := opts.Page
	if scope == SearchScopeAll {
		page = 1
	}

	if scope == SearchScopeAll || scope == SearchScopePosts {
		posts, total, err := s.blog.ListPostsWithOptions(BlogListOptions{Query: query, Page: page, PageSize: opts.PageSize})
		if err != nil {
			return SearchResults{}, fmt.Errorf("search posts: %w", err)
		}
		results.Posts = posts
		results.PostCount = total
	}
	if scope == SearchScopeAll || scope == SearchScopeProjects {
		projects, total, err := s.projects.ListProjects(page, opts.PageSize, ProjectListOptions{Query: query})
		if err != nil {
			return SearchResults{}, fmt.Errorf("search projects: %w", err)
		}
		results.Projects = projects
		results.ProjectCount = total
	}
	if scope == SearchScopeAll || scope == SearchScopePages {
		pageCount, err := s.repos.PageContent.CountSearch(publicSearchPageKeys(), query)
		if err != nil {
			return SearchResults{}, fmt.Errorf("count pages: %w", err)
		}
		pages, err := s.repos.PageContent.Search(repo.PageContentSearchOptions{Keys: publicSearchPageKeys(), Query: query, Limit: opts.PageSize, Offset: (page - 1) * opts.PageSize})
		if err != nil {
			return SearchResults{}, fmt.Errorf("search pages: %w", err)
		}
		for i := range pages {
			pages[i].ContentHTML, err = s.pages.RenderMarkdown(pages[i].ContentMD)
			if err != nil {
				return SearchResults{}, err
			}
		}
		results.Pages = pages
		results.PageCount = pageCount
	}

	results.Total = results.PostCount + results.ProjectCount + results.PageCount
	results.Paginated = scope != SearchScopeAll
	if results.Paginated {
		results.TotalPages = pageCountFor(results.Total, opts.PageSize)
	}
	return results, nil
}

func normalizeSearchScope(scope string) (string, error) {
	scope = strings.TrimSpace(strings.ToLower(scope))
	switch scope {
	case "", SearchScopeAll:
		return SearchScopeAll, nil
	case SearchScopePosts, SearchScopeProjects, SearchScopePages:
		return scope, nil
	default:
		return "", fmt.Errorf("invalid search scope: %s", scope)
	}
}

func publicSearchPageKeys() []string {
	return []string{model.PageKeyAbout, model.PageKeyContact}
}

func pageCountFor(total, pageSize int) int {
	if total <= 0 || pageSize <= 0 {
		return 0
	}
	pages := total / pageSize
	if total%pageSize != 0 {
		pages++
	}
	return pages
}
