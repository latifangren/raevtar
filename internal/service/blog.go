package service

import (
	"errors"
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

var ErrInvalidPostInput = errors.New("invalid post input")

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

	total, err := s.repos.Post.Count(categorySlug, true)
	if err != nil {
		return nil, 0, fmt.Errorf("count posts: %w", err)
	}

	posts, err := s.repos.Post.List(categorySlug, true, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list posts: %w", err)
	}

	return posts, total, nil
}

func (s *BlogService) ListCategories() ([]model.Category, error) {
	categories, err := s.repos.Category.List()
	if err != nil {
		return nil, fmt.Errorf("list categories: %w", err)
	}
	return categories, nil
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

func (s *BlogService) GetPublishedPost(slug string) (*model.Post, error) {
	post, err := s.GetPost(slug)
	if err != nil {
		return nil, err
	}
	if !post.Published {
		return nil, fmt.Errorf("post not found: %s", slug)
	}
	return post, nil
}

func (s *BlogService) CreatePost(input model.PostCreate) (*model.Post, error) {
	input = cleanPostCreate(input)
	if input.Title == "" || input.CategorySlug == "" || input.ContentMD == "" {
		return nil, fmt.Errorf("%w: title, category_slug, and content_md required", ErrInvalidPostInput)
	}

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

func cleanPostCreate(input model.PostCreate) model.PostCreate {
	input.CategorySlug = strings.TrimSpace(input.CategorySlug)
	input.Title = strings.TrimSpace(input.Title)
	input.ContentMD = strings.TrimSpace(input.ContentMD)
	input.Excerpt = strings.TrimSpace(input.Excerpt)
	cleanTags := make([]string, 0, len(input.Tags))
	for _, tag := range input.Tags {
		tag = strings.TrimSpace(tag)
		if tag != "" {
			cleanTags = append(cleanTags, tag)
		}
	}
	input.Tags = cleanTags
	return input
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
