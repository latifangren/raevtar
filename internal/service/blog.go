package service

import (
	"database/sql"
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
var ErrPostNotFound = errors.New("post not found")
var ErrInvalidCategoryInput = errors.New("invalid category input")
var ErrCategoryNotFound = errors.New("category not found")
var ErrCategoryInUse = errors.New("category in use")

type BlogService struct {
	repos    *repo.Repositories
	markdown goldmark.Markdown
}

type BlogListOptions struct {
	CategorySlug string
	Query        string
	Page         int
	PageSize     int
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
	return s.ListPostsWithOptions(BlogListOptions{CategorySlug: categorySlug, Page: page, PageSize: pageSize})
}

func (s *BlogService) ListPostsWithOptions(opts BlogListOptions) ([]model.Post, int, error) {
	if opts.Page < 1 {
		opts.Page = 1
	}
	if opts.PageSize <= 0 {
		opts.PageSize = 10
	}
	opts.CategorySlug = strings.TrimSpace(opts.CategorySlug)
	opts.Query = strings.TrimSpace(opts.Query)
	offset := (opts.Page - 1) * opts.PageSize

	total, err := s.repos.Post.CountWithOptions(repo.PostListOptions{
		CategorySlug:  opts.CategorySlug,
		PublishedOnly: true,
		Query:         opts.Query,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("count posts: %w", err)
	}

	posts, err := s.repos.Post.ListWithOptions(repo.PostListOptions{
		CategorySlug:  opts.CategorySlug,
		PublishedOnly: true,
		Query:         opts.Query,
		Limit:         opts.PageSize,
		Offset:        offset,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("list posts: %w", err)
	}

	return posts, total, nil
}

func (s *BlogService) ListAllPosts(page, pageSize int) ([]model.Post, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize
	total, err := s.repos.Post.CountWithOptions(repo.PostListOptions{PublishedOnly: false})
	if err != nil {
		return nil, 0, fmt.Errorf("count posts: %w", err)
	}
	posts, err := s.repos.Post.ListWithOptions(repo.PostListOptions{PublishedOnly: false, Limit: pageSize, Offset: offset})
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

func (s *BlogService) GetCategoryByID(id int64) (*model.Category, int, error) {
	category, err := s.repos.Category.GetByID(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, 0, fmt.Errorf("%w: %w", ErrCategoryNotFound, err)
		}
		return nil, 0, fmt.Errorf("get category: %w", err)
	}
	count, err := s.repos.Post.CountByCategoryID(id)
	if err != nil {
		return nil, 0, fmt.Errorf("count category posts: %w", err)
	}
	return category, count, nil
}

func (s *BlogService) PostCountForCategory(categoryID int64) (int, error) {
	count, err := s.repos.Post.CountByCategoryID(categoryID)
	if err != nil {
		return 0, fmt.Errorf("count category posts: %w", err)
	}
	return count, nil
}

func (s *BlogService) CreateCategory(input model.Category) (*model.Category, error) {
	input, err := cleanCategoryInput(input)
	if err != nil {
		return nil, err
	}
	exists, err := s.repos.Category.SlugExists(input.Slug)
	if err != nil {
		return nil, fmt.Errorf("check category slug: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("%w: slug already exists", ErrInvalidCategoryInput)
	}
	now := time.Now()
	category := &model.Category{
		Slug:        input.Slug,
		Name:        input.Name,
		Description: input.Description,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := s.repos.Category.Create(category); err != nil {
		return nil, fmt.Errorf("create category: %w", err)
	}
	return category, nil
}

func (s *BlogService) UpdateCategory(id int64, input model.Category) (*model.Category, error) {
	input, err := cleanCategoryInput(input)
	if err != nil {
		return nil, err
	}
	current, err := s.repos.Category.GetByID(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w: %w", ErrCategoryNotFound, err)
		}
		return nil, fmt.Errorf("get category: %w", err)
	}
	exists, err := s.repos.Category.SlugExistsExcludingID(input.Slug, id)
	if err != nil {
		return nil, fmt.Errorf("check category slug: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("%w: slug already exists", ErrInvalidCategoryInput)
	}
	postCount, err := s.repos.Post.CountByCategoryID(id)
	if err != nil {
		return nil, fmt.Errorf("count category posts: %w", err)
	}
	if current.Slug != input.Slug && postCount > 0 {
		return nil, fmt.Errorf("%w: slug cannot change while posts exist", ErrCategoryInUse)
	}
	current.Slug = input.Slug
	current.Name = input.Name
	current.Description = input.Description
	current.UpdatedAt = time.Now()
	if err := s.repos.Category.Update(current); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w: %w", ErrCategoryNotFound, err)
		}
		return nil, fmt.Errorf("update category: %w", err)
	}
	return current, nil
}

func (s *BlogService) DeleteCategory(id int64) error {
	if _, err := s.repos.Category.GetByID(id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("%w: %w", ErrCategoryNotFound, err)
		}
		return fmt.Errorf("get category: %w", err)
	}
	postCount, err := s.repos.Post.CountByCategoryID(id)
	if err != nil {
		return fmt.Errorf("count category posts: %w", err)
	}
	if postCount > 0 {
		return fmt.Errorf("%w: category still has posts", ErrCategoryInUse)
	}
	if err := s.repos.Category.Delete(id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("%w: %w", ErrCategoryNotFound, err)
		}
		return fmt.Errorf("delete category: %w", err)
	}
	return nil
}

func (s *BlogService) GetPost(slug string) (*model.Post, error) {
	post, err := s.repos.Post.GetBySlug(slug)
	if err != nil {
		return nil, err
	}
	post.ContentHTML, err = s.RenderMarkdown(post.ContentMD)
	if err != nil {
		return nil, err
	}
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
		CategoryID:    cat.ID,
		Title:         input.Title,
		Slug:          "",
		ContentMD:     input.ContentMD,
		Excerpt:       input.Excerpt,
		CoverImageURL: input.CoverImageURL,
		Published:     input.Published,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	post.Slug, err = s.uniqueSlug(input.Title)
	if err != nil {
		return nil, fmt.Errorf("generate slug: %w", err)
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

func (s *BlogService) GetPostByID(id int64) (*model.Post, error) {
	post, err := s.repos.Post.GetByID(id)
	if err != nil {
		return nil, err
	}
	post.ContentHTML, err = s.RenderMarkdown(post.ContentMD)
	if err != nil {
		return nil, err
	}
	return post, nil
}

func (s *BlogService) RenderMarkdown(content string) (string, error) {
	// Pre-process shortcodes
	content = s.processShortcodes(content)

	var buf strings.Builder
	if err := s.markdown.Convert([]byte(content), &buf); err != nil {
		return "", fmt.Errorf("render markdown: %w", err)
	}
	return buf.String(), nil
}

func (s *BlogService) processShortcodes(content string) string {
	// Simple shortcode: [[server-status:node-name]]
	// This will be rendered as a div that HTMX can pick up
	return serverStatusRe.ReplaceAllStringFunc(content, func(match string) string {
		parts := serverStatusRe.FindStringSubmatch(match)
		if len(parts) < 2 {
			return match
		}
		nodeName := parts[1]
		return fmt.Sprintf(`<div class="nb-card bg-retro-paper p-4 my-6" hx-get="/lab/node-status/%s" hx-trigger="load" hx-swap="outerHTML">
			<p class="text-xs font-black uppercase text-retro-muted animate-pulse">Loading node status: %s...</p>
		</div>`, nodeName, nodeName)
	})
}

func (s *BlogService) UpdatePost(id int64, input model.PostUpdate) (*model.Post, error) {
	input.CategorySlug = strings.TrimSpace(input.CategorySlug)
	input.Title = strings.TrimSpace(input.Title)
	input.ContentMD = strings.TrimSpace(input.ContentMD)
	input.Excerpt = strings.TrimSpace(input.Excerpt)
	input.CoverImageURL = strings.TrimSpace(input.CoverImageURL)
	cleanTags := make([]string, 0, len(input.Tags))
	for _, tag := range input.Tags {
		tag = strings.TrimSpace(tag)
		if tag != "" {
			cleanTags = append(cleanTags, tag)
		}
	}
	input.Tags = cleanTags
	if input.Title == "" || input.CategorySlug == "" || input.ContentMD == "" {
		return nil, fmt.Errorf("%w: title, category_slug, and content_md required", ErrInvalidPostInput)
	}
	post, err := s.repos.Post.GetByID(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w: %w", ErrPostNotFound, err)
		}
		return nil, fmt.Errorf("get post: %w", err)
	}
	cat, err := s.repos.Category.GetBySlug(input.CategorySlug)
	if err != nil {
		return nil, fmt.Errorf("category not found: %s", input.CategorySlug)
	}
	post.CategoryID = cat.ID
	post.Title = input.Title
	post.ContentMD = input.ContentMD
	post.Excerpt = input.Excerpt
	post.CoverImageURL = input.CoverImageURL
	post.Published = input.Published
	post.UpdatedAt = time.Now()
	if err := s.repos.Post.Update(post); err != nil {
		return nil, fmt.Errorf("update post: %w", err)
	}
	if err := s.repos.Tag.SetTags(post.ID, input.Tags); err != nil {
		return nil, fmt.Errorf("set tags: %w", err)
	}
	return s.GetPostByID(post.ID)
}

func cleanPostCreate(input model.PostCreate) model.PostCreate {
	input.CategorySlug = strings.TrimSpace(input.CategorySlug)
	input.Title = strings.TrimSpace(input.Title)
	input.ContentMD = strings.TrimSpace(input.ContentMD)
	input.Excerpt = strings.TrimSpace(input.Excerpt)
	input.CoverImageURL = strings.TrimSpace(input.CoverImageURL)
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

func cleanCategoryInput(input model.Category) (model.Category, error) {
	input.Slug = strings.TrimSpace(strings.ToLower(input.Slug))
	input.Name = strings.TrimSpace(input.Name)
	input.Description = strings.TrimSpace(input.Description)
	if input.Slug == "" || input.Name == "" {
		return model.Category{}, fmt.Errorf("%w: slug and name required", ErrInvalidCategoryInput)
	}
	if generateSlug(input.Slug) != input.Slug {
		return model.Category{}, fmt.Errorf("%w: slug must be lowercase kebab-case", ErrInvalidCategoryInput)
	}
	return input, nil
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

func (s *BlogService) uniqueSlug(title string) (string, error) {
	base := generateSlug(title)
	for i := 1; ; i++ {
		slug := base
		if i > 1 {
			slug = fmt.Sprintf("%s-%d", base, i)
		}
		exists, err := s.repos.Post.SlugExists(slug)
		if err != nil {
			return "", err
		}
		if !exists {
			return slug, nil
		}
	}
}
