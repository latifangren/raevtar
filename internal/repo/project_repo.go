package repo

import (
	"raevtar/internal/model"

	"github.com/jmoiron/sqlx"
)

type ProjectRepo struct{ db *sqlx.DB }

type ProjectListOptions struct {
	PublishedOnly bool
	FeaturedOnly  bool
	Sort          string
	Limit         int
	Offset        int
}

func (r *ProjectRepo) List(opts ProjectListOptions) ([]model.Project, error) {
	query := `SELECT * FROM projects WHERE 1=1`
	args := make([]interface{}, 0, 4)
	if opts.PublishedOnly {
		query += " AND published = 1"
	}
	if opts.FeaturedOnly {
		query += " AND featured = 1"
	}
	query += " ORDER BY " + projectOrderClause(opts.Sort) + " LIMIT ? OFFSET ?"
	args = append(args, opts.Limit, opts.Offset)

	var projects []model.Project
	if err := r.db.Select(&projects, query, args...); err != nil {
		return nil, err
	}

	tagRepo := &TagRepo{db: r.db}
	ids := make([]int64, len(projects))
	for i, project := range projects {
		ids[i] = project.ID
	}
	tagMap, err := tagRepo.GetByProjectIDs(ids)
	if err != nil {
		return nil, err
	}
	for i, project := range projects {
		projects[i].Tags = tagMap[project.ID]
	}

	return projects, nil
}

func projectOrderClause(sort string) string {
	switch sort {
	case "oldest":
		return "featured DESC, sort_order ASC, created_at ASC, id ASC"
	case "newest", "":
		fallthrough
	default:
		return "featured DESC, sort_order ASC, created_at DESC, id DESC"
	}
}

func (r *ProjectRepo) GetBySlug(slug string) (*model.Project, error) {
	var project model.Project
	if err := r.db.Get(&project, `SELECT * FROM projects WHERE slug = ?`, slug); err != nil {
		return nil, err
	}
	tags, err := (&TagRepo{db: r.db}).GetByProjectID(project.ID)
	if err != nil {
		return nil, err
	}
	project.Tags = tags
	return &project, nil
}

func (r *ProjectRepo) GetByID(id int64) (*model.Project, error) {
	var project model.Project
	if err := r.db.Get(&project, `SELECT * FROM projects WHERE id = ?`, id); err != nil {
		return nil, err
	}
	tags, err := (&TagRepo{db: r.db}).GetByProjectID(project.ID)
	if err != nil {
		return nil, err
	}
	project.Tags = tags
	return &project, nil
}

func (r *ProjectRepo) Create(project *model.Project) error {
	result, err := r.db.Exec(`
		INSERT INTO projects (title, slug, content_md, excerpt, cover_image_url, published, featured, sort_order)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		project.Title, project.Slug, project.ContentMD, project.Excerpt, project.CoverImageURL, project.Published, project.Featured, project.SortOrder,
	)
	if err != nil {
		return err
	}
	id, _ := result.LastInsertId()
	project.ID = id
	return nil
}

func (r *ProjectRepo) Update(project *model.Project) error {
	_, err := r.db.Exec(`
		UPDATE projects
		SET title = ?, content_md = ?, excerpt = ?, cover_image_url = ?, published = ?, featured = ?, sort_order = ?, updated_at = ?
		WHERE id = ?`,
		project.Title, project.ContentMD, project.Excerpt, project.CoverImageURL, project.Published, project.Featured, project.SortOrder, project.UpdatedAt, project.ID,
	)
	return err
}

func (r *ProjectRepo) SlugExists(slug string) (bool, error) {
	var count int
	if err := r.db.Get(&count, `SELECT COUNT(*) FROM projects WHERE slug = ?`, slug); err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *ProjectRepo) Delete(id int64) error {
	_, err := r.db.Exec(`DELETE FROM projects WHERE id = ?`, id)
	return err
}

func (r *ProjectRepo) Count(publishedOnly bool, featuredOnly bool) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM projects WHERE 1=1`
	if publishedOnly {
		query += ` AND published = 1`
	}
	if featuredOnly {
		query += ` AND featured = 1`
	}
	return count, r.db.Get(&count, query)
}
