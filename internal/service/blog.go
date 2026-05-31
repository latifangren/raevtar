package service

import (
	"fmt"
	"regexp"
	"strings"
	"time"
	"unicode"

	"raevtar/internal/model"
	"raevtar/internal/repo"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
)

var slugRe = regexp.MustCompile(`[^a-z0-9-]`)

type BlogService struct {
	repos    *repo.Repositories
	markdown goldmark.Markdown
}

func NewBlogService(repos *repo.Repositories) *BlogService {
	md := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			extension.Typographer,
		),
	)
	return &BlogService{repos: repos, markdown: md}
}

func (s *BlogService) ListPosts(categorySlug string, page, pageSize int) ([]model.Post, int, error) {
	if pageSize <= 0 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize

	total, err := s.repos.Post.Count(categorySlug)
	if err != nil {
		return nil, 0, fmt.Errorf("count posts: %w", err)
	}

	posts, err := s.repos.Post.List(categorySlug, true, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list posts: %w", err)
	}

	return posts, total, nil
}

func (s *BlogService) GetPost(slug string) (*model.Post, error) {
	post, err := s.repos.Post.GetBySlug(slug)
	if err != nil {
		return nil, err
	}
	// render markdown to HTML
	var buf strings.Builder
	if err := s.markdown.Convert([]byte(post.ContentMD), &buf); err != nil {
		return nil, fmt.Errorf("render markdown: %w", err)
	}
	post.ContentHTML = buf.String()
	return post, nil
}

func (s *BlogService) CreatePost(input model.PostCreate) (*model.Post, error) {
	cat, err := s.repos.Category.GetBySlug(input.CategorySlug)
	if err != nil {
		return nil, fmt.Errorf("category not found: %s", input.CategorySlug)
	}

	post := &model.Post{
		CategoryID: cat.ID,
		Title:      input.Title,
		Slug:       generateSlug(input.Title),
		ContentMD:  input.ContentMD,
		Excerpt:    input.Excerpt,
		Published:  input.Published,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	if err := s.repos.Post.Create(post); err != nil {
		return nil, fmt.Errorf("create post: %w", err)
	}

	// handle tags
	if len(input.Tags) > 0 {
		if err := s.repos.Tag.SetTags(post.ID, input.Tags); err != nil {
			return nil, fmt.Errorf("set tags: %w", err)
		}
	}

	return s.GetPost(post.Slug) // reload with tags + rendered markdown
}

func generateSlug(title string) string {
	slug := strings.ToLower(title)
	slug = strings.ReplaceAll(slug, " ", "-")
	slug = slugRe.ReplaceAllString(slug, "")
	slug = strings.TrimFunc(slug, func(r rune) bool {
		return r == '-' || unicode.IsSpace(r)
	})
	if slug == "" {
		slug = fmt.Sprintf("post-%d", time.Now().Unix())
	}
	return slug
}
