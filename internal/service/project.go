package service

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"raevtar/internal/model"
	"raevtar/internal/repo"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
)

var ErrInvalidProjectInput = errors.New("invalid project input")
var ErrProjectNotFound = errors.New("project not found")

type ProjectService struct {
	repos    *repo.Repositories
	markdown goldmark.Markdown
}

func NewProjectService(repos *repo.Repositories) *ProjectService {
	md := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			extension.Typographer,
		),
	)
	return &ProjectService{repos: repos, markdown: md}
}

func (s *ProjectService) ListProjects(page, pageSize int) ([]model.Project, int, error) {
	if pageSize <= 0 {
		pageSize = 10
	}
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * pageSize
	total, err := s.repos.Project.Count(true)
	if err != nil {
		return nil, 0, fmt.Errorf("count projects: %w", err)
	}
	projects, err := s.repos.Project.List(true, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list projects: %w", err)
	}
	return projects, total, nil
}

func (s *ProjectService) ListAllProjects(page, pageSize int) ([]model.Project, int, error) {
	if pageSize <= 0 {
		pageSize = 10
	}
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * pageSize
	total, err := s.repos.Project.Count(false)
	if err != nil {
		return nil, 0, fmt.Errorf("count projects: %w", err)
	}
	projects, err := s.repos.Project.List(false, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list projects: %w", err)
	}
	return projects, total, nil
}

func (s *ProjectService) GetProject(slug string) (*model.Project, error) {
	project, err := s.repos.Project.GetBySlug(slug)
	if err != nil {
		return nil, err
	}
	project.ContentHTML, err = s.RenderMarkdown(project.ContentMD)
	if err != nil {
		return nil, err
	}
	return project, nil
}

func (s *ProjectService) GetPublishedProject(slug string) (*model.Project, error) {
	project, err := s.GetProject(slug)
	if err != nil {
		return nil, err
	}
	if !project.Published {
		return nil, fmt.Errorf("project not found: %s", slug)
	}
	return project, nil
}

func (s *ProjectService) GetProjectByID(id int64) (*model.Project, error) {
	project, err := s.repos.Project.GetByID(id)
	if err != nil {
		return nil, err
	}
	project.ContentHTML, err = s.RenderMarkdown(project.ContentMD)
	if err != nil {
		return nil, err
	}
	return project, nil
}

func (s *ProjectService) CreateProject(input model.ProjectCreate) (*model.Project, error) {
	input = cleanProjectCreate(input)
	if input.Title == "" || input.ContentMD == "" {
		return nil, fmt.Errorf("%w: title and content_md required", ErrInvalidProjectInput)
	}
	project := &model.Project{
		Title:         input.Title,
		ContentMD:     input.ContentMD,
		Excerpt:       input.Excerpt,
		CoverImageURL: input.CoverImageURL,
		Published:     input.Published,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	slug, err := s.uniqueSlug(input.Title)
	if err != nil {
		return nil, fmt.Errorf("generate slug: %w", err)
	}
	project.Slug = slug
	if err := s.repos.Project.Create(project); err != nil {
		return nil, fmt.Errorf("create project: %w", err)
	}
	if err := s.repos.Tag.SetProjectTags(project.ID, input.Tags); err != nil {
		return nil, fmt.Errorf("set project tags: %w", err)
	}
	return s.GetProject(project.Slug)
}

func (s *ProjectService) UpdateProject(id int64, input model.ProjectUpdate) (*model.Project, error) {
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
	if input.Title == "" || input.ContentMD == "" {
		return nil, fmt.Errorf("%w: title and content_md required", ErrInvalidProjectInput)
	}
	project, err := s.repos.Project.GetByID(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w: %w", ErrProjectNotFound, err)
		}
		return nil, fmt.Errorf("get project: %w", err)
	}
	project.Title = input.Title
	project.ContentMD = input.ContentMD
	project.Excerpt = input.Excerpt
	project.CoverImageURL = input.CoverImageURL
	project.Published = input.Published
	project.UpdatedAt = time.Now()
	if err := s.repos.Project.Update(project); err != nil {
		return nil, fmt.Errorf("update project: %w", err)
	}
	if err := s.repos.Tag.SetProjectTags(project.ID, input.Tags); err != nil {
		return nil, fmt.Errorf("set project tags: %w", err)
	}
	return s.GetProjectByID(project.ID)
}

func (s *ProjectService) RenderMarkdown(content string) (string, error) {
	var buf strings.Builder
	if err := s.markdown.Convert([]byte(content), &buf); err != nil {
		return "", fmt.Errorf("render markdown: %w", err)
	}
	return buf.String(), nil
}

func cleanProjectCreate(input model.ProjectCreate) model.ProjectCreate {
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

func (s *ProjectService) uniqueSlug(title string) (string, error) {
	base := generateSlug(title)
	for i := 1; ; i++ {
		slug := base
		if i > 1 {
			slug = fmt.Sprintf("%s-%d", base, i)
		}
		exists, err := s.repos.Project.SlugExists(slug)
		if err != nil {
			return "", err
		}
		if !exists {
			return slug, nil
		}
	}
}
