package model

import "time"

type Project struct {
	ID            int64     `db:"id" json:"id"`
	Title         string    `db:"title" json:"title"`
	Slug          string    `db:"slug" json:"slug"`
	ContentMD     string    `db:"content_md" json:"content_md"`
	ContentHTML   string    `db:"-" json:"content_html"`
	Excerpt       string    `db:"excerpt" json:"excerpt"`
	CoverImageURL string    `db:"cover_image_url" json:"cover_image_url,omitempty"`
	Published     bool      `db:"published" json:"published"`
	State         string    `db:"state" json:"state"`
	Featured      bool      `db:"featured" json:"featured"`
	SortOrder     int       `db:"sort_order" json:"sort_order"`
	CreatedAt     time.Time `db:"created_at" json:"created_at"`
	UpdatedAt     time.Time `db:"updated_at" json:"updated_at"`

	Tags          []Tag                 `db:"-" json:"tags,omitempty"`
	Updates       []ProjectUpdateEntry  `db:"-" json:"updates,omitempty"`
	RelatedItems  []ContentRelationView `db:"-" json:"related_items,omitempty"`
	ShowcaseItems []ProjectShowcaseItem `db:"-" json:"showcase_items,omitempty"`
}

type ProjectCreate struct {
	Title         string   `json:"title"`
	ContentMD     string   `json:"content_md"`
	Excerpt       string   `json:"excerpt"`
	CoverImageURL string   `json:"cover_image_url"`
	Published     bool     `json:"published"`
	State         string   `json:"state"`
	Featured      bool     `json:"featured"`
	SortOrder     int      `json:"sort_order"`
	Tags          []string `json:"tags"`
}

type ProjectUpdate struct {
	Title         string   `json:"title"`
	ContentMD     string   `json:"content_md"`
	Excerpt       string   `json:"excerpt"`
	CoverImageURL string   `json:"cover_image_url"`
	Published     bool     `json:"published"`
	State         string   `json:"state"`
	Featured      bool     `json:"featured"`
	SortOrder     int      `json:"sort_order"`
	Tags          []string `json:"tags"`
}

const (
	ProjectStatePlanning = "planning"
	ProjectStateActive   = "active"
	ProjectStatePaused   = "paused"
	ProjectStateShipped  = "shipped"
	ProjectStateArchived = "archived"
)

func ValidProjectStates() []string {
	return []string{
		ProjectStatePlanning,
		ProjectStateActive,
		ProjectStatePaused,
		ProjectStateShipped,
		ProjectStateArchived,
	}
}

const (
	ProjectUpdateKindTimeline  = "timeline"
	ProjectUpdateKindBuildLog  = "build_log"
	ProjectUpdateKindChangelog = "changelog"
)

func ValidProjectUpdateKinds() []string {
	return []string{
		ProjectUpdateKindTimeline,
		ProjectUpdateKindBuildLog,
		ProjectUpdateKindChangelog,
	}
}

const (
	ContentRelationTypePost    = "post"
	ContentRelationTypeProject = "project"
)

func ValidContentRelationTypes() []string {
	return []string{ContentRelationTypePost, ContentRelationTypeProject}
}

const (
	ContentRelationKindRelated    = "related"
	ContentRelationKindInspiredBy = "inspired_by"
	ContentRelationKindBuildsOn   = "builds_on"
)

func ValidContentRelationKinds() []string {
	return []string{
		ContentRelationKindRelated,
		ContentRelationKindInspiredBy,
		ContentRelationKindBuildsOn,
	}
}

const (
	ProjectShowcaseKindImage  = "image"
	ProjectShowcaseKindLink   = "link"
	ProjectShowcaseKindRepo   = "repo"
	ProjectShowcaseKindMetric = "metric"
	ProjectShowcaseKindVideo  = "video"
)

func ValidProjectShowcaseKinds() []string {
	return []string{
		ProjectShowcaseKindImage,
		ProjectShowcaseKindLink,
		ProjectShowcaseKindRepo,
		ProjectShowcaseKindMetric,
		ProjectShowcaseKindVideo,
	}
}
