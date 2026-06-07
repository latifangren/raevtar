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
var ErrInvalidProjectSort = errors.New("invalid project sort")
var ErrInvalidProjectState = errors.New("invalid project state")
var ErrInvalidProjectUpdateKind = errors.New("invalid project update kind")
var ErrProjectUpdateNotFound = errors.New("project update not found")
var ErrProjectShowcaseNotFound = errors.New("project showcase item not found")
var ErrInvalidContentRelation = errors.New("invalid content relation")
var ErrInvalidProjectShowcase = errors.New("invalid project showcase")

type ProjectListOptions struct {
	FeaturedOnly bool
	State        string
	Sort         string
}

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

func (s *ProjectService) ListProjects(page, pageSize int, opts ProjectListOptions) ([]model.Project, int, error) {
	if pageSize <= 0 {
		pageSize = 10
	}
	if page < 1 {
		page = 1
	}
	opts, err := normalizeProjectListOptions(opts)
	if err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * pageSize
	total, err := s.repos.Project.Count(true, opts.FeaturedOnly, opts.State)
	if err != nil {
		return nil, 0, fmt.Errorf("count projects: %w", err)
	}
	projects, err := s.repos.Project.List(repo.ProjectListOptions{
		PublishedOnly: true,
		FeaturedOnly:  opts.FeaturedOnly,
		State:         opts.State,
		Sort:          opts.Sort,
		Limit:         pageSize,
		Offset:        offset,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("list projects: %w", err)
	}
	return projects, total, nil
}

func (s *ProjectService) ListAllProjects(page, pageSize int, opts ProjectListOptions) ([]model.Project, int, error) {
	if pageSize <= 0 {
		pageSize = 10
	}
	if page < 1 {
		page = 1
	}
	opts, err := normalizeProjectListOptions(opts)
	if err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * pageSize
	total, err := s.repos.Project.Count(false, opts.FeaturedOnly, opts.State)
	if err != nil {
		return nil, 0, fmt.Errorf("count projects: %w", err)
	}
	projects, err := s.repos.Project.List(repo.ProjectListOptions{
		PublishedOnly: false,
		FeaturedOnly:  opts.FeaturedOnly,
		State:         opts.State,
		Sort:          opts.Sort,
		Limit:         pageSize,
		Offset:        offset,
	})
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
		State:         normalizeProjectInputState(input.State),
		Featured:      input.Featured,
		SortOrder:     normalizeProjectSortOrder(input.SortOrder),
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
	project.State = normalizeProjectInputState(input.State)
	project.Featured = input.Featured
	project.SortOrder = normalizeProjectSortOrder(input.SortOrder)
	project.UpdatedAt = time.Now()
	if err := s.repos.Project.Update(project); err != nil {
		return nil, fmt.Errorf("update project: %w", err)
	}
	if err := s.repos.Tag.SetProjectTags(project.ID, input.Tags); err != nil {
		return nil, fmt.Errorf("set project tags: %w", err)
	}
	return s.GetProjectByID(project.ID)
}

func (s *ProjectService) DeleteProject(id int64) error {
	if _, err := s.repos.Project.GetByID(id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("%w: %w", ErrProjectNotFound, err)
		}
		return fmt.Errorf("get project: %w", err)
	}
	if err := s.repos.Project.Delete(id); err != nil {
		return fmt.Errorf("delete project: %w", err)
	}
	return nil
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
	input.State = normalizeProjectInputState(input.State)
	cleanTags := make([]string, 0, len(input.Tags))
	for _, tag := range input.Tags {
		tag = strings.TrimSpace(tag)
		if tag != "" {
			cleanTags = append(cleanTags, tag)
		}
	}
	input.Tags = cleanTags
	input.SortOrder = normalizeProjectSortOrder(input.SortOrder)
	input.State = normalizeProjectInputState(input.State)
	return input
}

func normalizeProjectSortOrder(sortOrder int) int {
	if sortOrder < 0 {
		return 0
	}
	return sortOrder
}

func normalizeProjectListOptions(opts ProjectListOptions) (ProjectListOptions, error) {
	sort := strings.TrimSpace(strings.ToLower(opts.Sort))
	state := normalizeProjectFilterState(opts.State)
	switch sort {
	case "", "newest", "oldest":
		opts.Sort = sort
		opts.State = state
		if state != "" && !isValidProjectState(state) {
			return ProjectListOptions{}, fmt.Errorf("%w: %s", ErrInvalidProjectState, state)
		}
		return opts, nil
	default:
		return ProjectListOptions{}, fmt.Errorf("%w: %s", ErrInvalidProjectSort, sort)
	}
}

func normalizeProjectInputState(state string) string {
	state = strings.TrimSpace(strings.ToLower(state))
	if state == "" {
		return model.ProjectStateActive
	}
	return state
}

func isValidProjectState(state string) bool {
	for _, valid := range model.ValidProjectStates() {
		if state == valid {
			return true
		}
	}
	return false
}

func normalizeProjectFilterState(state string) string {
	return strings.TrimSpace(strings.ToLower(state))
}

func isValidProjectUpdateKind(kind string) bool {
	for _, valid := range model.ValidProjectUpdateKinds() {
		if kind == valid {
			return true
		}
	}
	return false
}

func (s *ProjectService) ListProjectUpdates(projectID int64, publishedOnly bool, kinds []string, limit int) ([]model.ProjectUpdateEntry, error) {
	items, err := s.repos.ProjectUpdate.List(repo.ProjectUpdateListOptions{ProjectID: projectID, PublishedOnly: publishedOnly, Kinds: kinds, Limit: limit})
	if err != nil {
		return nil, fmt.Errorf("list project updates: %w", err)
	}
	for i := range items {
		items[i].ContentHTML, err = s.RenderMarkdown(items[i].ContentMD)
		if err != nil {
			return nil, err
		}
	}
	return items, nil
}

func (s *ProjectService) ListProjectTimeline(projectID int64, publishedOnly bool, limit int) ([]model.ProjectUpdateEntry, error) {
	return s.ListProjectUpdates(projectID, publishedOnly, []string{model.ProjectUpdateKindTimeline, model.ProjectUpdateKindBuildLog, model.ProjectUpdateKindChangelog}, limit)
}

func (s *ProjectService) ListProjectChangelog(projectID int64, publishedOnly bool, limit int) ([]model.ProjectUpdateEntry, error) {
	return s.ListProjectUpdates(projectID, publishedOnly, []string{model.ProjectUpdateKindChangelog}, limit)
}

func (s *ProjectService) ListProjectShowcase(projectID int64, publishedOnly bool) ([]model.ProjectShowcaseItem, error) {
	items, err := s.repos.ProjectShowcase.ListByProjectID(projectID, publishedOnly)
	if err != nil {
		return nil, fmt.Errorf("list project showcase: %w", err)
	}
	for i := range items {
		items[i].BodyHTML, err = s.RenderMarkdown(items[i].BodyMD)
		if err != nil {
			return nil, err
		}
	}
	return items, nil
}

func (s *ProjectService) ListProjectRelations(projectID int64) ([]model.ContentRelation, error) {
	items, err := s.repos.ContentRelation.ListBySource(model.ContentRelationTypeProject, projectID)
	if err != nil {
		return nil, fmt.Errorf("list project relations: %w", err)
	}
	return items, nil
}

func (s *ProjectService) GetResolvedProjectRelations(projectID int64, publishedOnly bool) ([]model.ContentRelationView, error) {
	relations, err := s.ListProjectRelations(projectID)
	if err != nil {
		return nil, err
	}
	views := make([]model.ContentRelationView, 0, len(relations))
	for _, rel := range relations {
		switch rel.TargetType {
		case model.ContentRelationTypeProject:
			project, err := s.repos.Project.GetByID(rel.TargetID)
			if err != nil {
				continue
			}
			if publishedOnly && !project.Published {
				continue
			}
			views = append(views, model.ContentRelationView{ID: rel.ID, RelationKind: rel.RelationKind, TargetType: rel.TargetType, TargetID: rel.TargetID, Title: project.Title, Slug: project.Slug, Excerpt: project.Excerpt, URL: "/projects/" + project.Slug, Published: project.Published})
		case model.ContentRelationTypePost:
			post, err := s.repos.Post.GetByID(rel.TargetID)
			if err != nil {
				continue
			}
			if publishedOnly && !post.Published {
				continue
			}
			views = append(views, model.ContentRelationView{ID: rel.ID, RelationKind: rel.RelationKind, TargetType: rel.TargetType, TargetID: rel.TargetID, Title: post.Title, Slug: post.Slug, Excerpt: post.Excerpt, URL: "/blog/" + post.Slug, Published: post.Published})
		}
	}
	return views, nil
}

func (s *ProjectService) CreateProjectUpdate(projectID int64, input model.ProjectUpdateEntryCreate) (*model.ProjectUpdateEntry, error) {
	if _, err := s.repos.Project.GetByID(projectID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w: %w", ErrProjectNotFound, err)
		}
		return nil, fmt.Errorf("get project: %w", err)
	}
	input.Kind = strings.TrimSpace(strings.ToLower(input.Kind))
	input.Title = strings.TrimSpace(input.Title)
	input.ContentMD = strings.TrimSpace(input.ContentMD)
	if input.Title == "" || input.ContentMD == "" {
		return nil, fmt.Errorf("%w: title and content_md required", ErrInvalidProjectInput)
	}
	if !isValidProjectUpdateKind(input.Kind) {
		return nil, fmt.Errorf("%w: %s", ErrInvalidProjectUpdateKind, input.Kind)
	}
	if input.EventAt.IsZero() {
		input.EventAt = time.Now()
	}
	item := &model.ProjectUpdateEntry{ProjectID: projectID, Kind: input.Kind, Title: input.Title, ContentMD: input.ContentMD, Published: input.Published, Pinned: input.Pinned, SortOrder: normalizeProjectSortOrder(input.SortOrder), EventAt: input.EventAt, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	if err := s.repos.ProjectUpdate.Create(item); err != nil {
		return nil, fmt.Errorf("create project update: %w", err)
	}
	item.ContentHTML, _ = s.RenderMarkdown(item.ContentMD)
	return item, nil
}

func (s *ProjectService) UpdateProjectUpdate(id int64, input model.ProjectUpdateEntryUpdate) (*model.ProjectUpdateEntry, error) {
	item, err := s.repos.ProjectUpdate.GetByID(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w: %w", ErrProjectUpdateNotFound, err)
		}
		return nil, fmt.Errorf("get project update: %w", err)
	}
	input.Kind = strings.TrimSpace(strings.ToLower(input.Kind))
	input.Title = strings.TrimSpace(input.Title)
	input.ContentMD = strings.TrimSpace(input.ContentMD)
	if input.Title == "" || input.ContentMD == "" {
		return nil, fmt.Errorf("%w: title and content_md required", ErrInvalidProjectInput)
	}
	if !isValidProjectUpdateKind(input.Kind) {
		return nil, fmt.Errorf("%w: %s", ErrInvalidProjectUpdateKind, input.Kind)
	}
	if input.EventAt.IsZero() {
		input.EventAt = item.EventAt
	}
	item.Kind = input.Kind
	item.Title = input.Title
	item.ContentMD = input.ContentMD
	item.Published = input.Published
	item.Pinned = input.Pinned
	item.SortOrder = normalizeProjectSortOrder(input.SortOrder)
	item.EventAt = input.EventAt
	item.UpdatedAt = time.Now()
	if err := s.repos.ProjectUpdate.Update(item); err != nil {
		return nil, fmt.Errorf("update project update: %w", err)
	}
	item.ContentHTML, _ = s.RenderMarkdown(item.ContentMD)
	return item, nil
}

func (s *ProjectService) DeleteProjectUpdate(id int64) error {
	if _, err := s.repos.ProjectUpdate.GetByID(id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("%w: %w", ErrProjectUpdateNotFound, err)
		}
		return fmt.Errorf("get project update: %w", err)
	}
	if err := s.repos.ProjectUpdate.Delete(id); err != nil {
		return fmt.Errorf("delete project update: %w", err)
	}
	return nil
}

func (s *ProjectService) CreateProjectShowcase(projectID int64, input model.ProjectShowcaseItemCreate) (*model.ProjectShowcaseItem, error) {
	if _, err := s.repos.Project.GetByID(projectID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w: %w", ErrProjectNotFound, err)
		}
		return nil, fmt.Errorf("get project: %w", err)
	}
	input.Kind = strings.TrimSpace(strings.ToLower(input.Kind))
	input.Title = strings.TrimSpace(input.Title)
	input.BodyMD = strings.TrimSpace(input.BodyMD)
	if input.Title == "" || !isValidProjectShowcaseKind(input.Kind) {
		return nil, fmt.Errorf("%w: invalid kind or title", ErrInvalidProjectShowcase)
	}
	item := &model.ProjectShowcaseItem{ProjectID: projectID, Kind: input.Kind, Title: input.Title, BodyMD: input.BodyMD, AssetURL: strings.TrimSpace(input.AssetURL), ExternalURL: strings.TrimSpace(input.ExternalURL), EmbedProvider: strings.TrimSpace(strings.ToLower(input.EmbedProvider)), EmbedRef: strings.TrimSpace(input.EmbedRef), Published: input.Published, SortOrder: normalizeProjectSortOrder(input.SortOrder), CreatedAt: time.Now(), UpdatedAt: time.Now()}
	if err := s.repos.ProjectShowcase.Create(item); err != nil {
		return nil, fmt.Errorf("create project showcase: %w", err)
	}
	item.BodyHTML, _ = s.RenderMarkdown(item.BodyMD)
	return item, nil
}

func (s *ProjectService) UpdateProjectShowcase(id int64, input model.ProjectShowcaseItemUpdate) (*model.ProjectShowcaseItem, error) {
	item, err := s.repos.ProjectShowcase.GetByID(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w: %w", ErrProjectShowcaseNotFound, err)
		}
		return nil, fmt.Errorf("get project showcase: %w", err)
	}
	input.Kind = strings.TrimSpace(strings.ToLower(input.Kind))
	input.Title = strings.TrimSpace(input.Title)
	input.BodyMD = strings.TrimSpace(input.BodyMD)
	if input.Title == "" || !isValidProjectShowcaseKind(input.Kind) {
		return nil, fmt.Errorf("%w: invalid kind or title", ErrInvalidProjectShowcase)
	}
	item.Kind = input.Kind
	item.Title = input.Title
	item.BodyMD = input.BodyMD
	item.AssetURL = strings.TrimSpace(input.AssetURL)
	item.ExternalURL = strings.TrimSpace(input.ExternalURL)
	item.EmbedProvider = strings.TrimSpace(strings.ToLower(input.EmbedProvider))
	item.EmbedRef = strings.TrimSpace(input.EmbedRef)
	item.Published = input.Published
	item.SortOrder = normalizeProjectSortOrder(input.SortOrder)
	item.UpdatedAt = time.Now()
	if err := s.repos.ProjectShowcase.Update(item); err != nil {
		return nil, fmt.Errorf("update project showcase: %w", err)
	}
	item.BodyHTML, _ = s.RenderMarkdown(item.BodyMD)
	return item, nil
}

func (s *ProjectService) DeleteProjectShowcase(id int64) error {
	if _, err := s.repos.ProjectShowcase.GetByID(id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("%w: %w", ErrProjectShowcaseNotFound, err)
		}
		return fmt.Errorf("get project showcase: %w", err)
	}
	if err := s.repos.ProjectShowcase.Delete(id); err != nil {
		return fmt.Errorf("delete project showcase: %w", err)
	}
	return nil
}

func (s *ProjectService) CreateProjectRelation(projectID int64, input model.ContentRelationCreate) (*model.ContentRelation, error) {
	if _, err := s.repos.Project.GetByID(projectID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w: %w", ErrProjectNotFound, err)
		}
		return nil, fmt.Errorf("get project: %w", err)
	}
	input.TargetType = strings.TrimSpace(strings.ToLower(input.TargetType))
	input.RelationKind = strings.TrimSpace(strings.ToLower(input.RelationKind))
	if !isValidContentRelationType(input.TargetType) || !isValidContentRelationKind(input.RelationKind) || input.TargetID <= 0 {
		return nil, fmt.Errorf("%w: invalid relation target", ErrInvalidContentRelation)
	}
	if input.TargetType == model.ContentRelationTypeProject && input.TargetID == projectID {
		return nil, fmt.Errorf("%w: self relation forbidden", ErrInvalidContentRelation)
	}
	relation := &model.ContentRelation{SourceType: model.ContentRelationTypeProject, SourceID: projectID, TargetType: input.TargetType, TargetID: input.TargetID, RelationKind: input.RelationKind, SortOrder: normalizeProjectSortOrder(input.SortOrder), CreatedAt: time.Now()}
	if err := s.repos.ContentRelation.Create(relation); err != nil {
		return nil, fmt.Errorf("create content relation: %w", err)
	}
	return relation, nil
}

func (s *ProjectService) DeleteProjectRelation(id int64) error {
	if err := s.repos.ContentRelation.Delete(id); err != nil {
		return fmt.Errorf("delete content relation: %w", err)
	}
	return nil
}

func isValidProjectShowcaseKind(kind string) bool {
	for _, valid := range model.ValidProjectShowcaseKinds() {
		if kind == valid {
			return true
		}
	}
	return false
}

func isValidContentRelationType(kind string) bool {
	for _, valid := range model.ValidContentRelationTypes() {
		if kind == valid {
			return true
		}
	}
	return false
}

func isValidContentRelationKind(kind string) bool {
	for _, valid := range model.ValidContentRelationKinds() {
		if kind == valid {
			return true
		}
	}
	return false
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
