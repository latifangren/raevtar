package service

import (
	"fmt"
	"strings"

	"raevtar/internal/model"
	"raevtar/internal/repo"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
)

type PageContentService struct {
	repos    *repo.Repositories
	markdown goldmark.Markdown
}

func NewPageContentService(repos *repo.Repositories) *PageContentService {
	md := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			extension.Typographer,
		),
	)
	return &PageContentService{repos: repos, markdown: md}
}

func (s *PageContentService) ListPages() ([]model.PageContent, error) {
	pages, err := s.repos.PageContent.List()
	if err != nil {
		return nil, fmt.Errorf("list pages: %w", err)
	}
	for i := range pages {
		pages[i].ContentHTML, err = s.RenderMarkdown(pages[i].ContentMD)
		if err != nil {
			return nil, err
		}
	}
	return pages, nil
}

func (s *PageContentService) GetPage(key string) (*model.PageContent, error) {
	page, err := s.repos.PageContent.GetByKey(strings.TrimSpace(key))
	if err != nil {
		return nil, fmt.Errorf("get page %s: %w", key, err)
	}
	page.ContentHTML, err = s.RenderMarkdown(page.ContentMD)
	if err != nil {
		return nil, err
	}
	return page, nil
}

func (s *PageContentService) UpdatePage(page model.PageContent) (*model.PageContent, error) {
	page.Key = strings.TrimSpace(page.Key)
	page.Title = strings.TrimSpace(page.Title)
	page.Summary = strings.TrimSpace(page.Summary)
	page.ContentMD = strings.TrimSpace(page.ContentMD)
	if page.Key == "" || page.Title == "" || page.ContentMD == "" {
		return nil, fmt.Errorf("page key, title, and content required")
	}
	if err := s.repos.PageContent.Upsert(&page); err != nil {
		return nil, fmt.Errorf("upsert page: %w", err)
	}
	return s.GetPage(page.Key)
}

func (s *PageContentService) RenderMarkdown(content string) (string, error) {
	var buf strings.Builder
	if err := s.markdown.Convert([]byte(content), &buf); err != nil {
		return "", fmt.Errorf("render markdown: %w", err)
	}
	return buf.String(), nil
}
