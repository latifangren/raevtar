package model

import "time"

type ContentRelation struct {
	ID           int64     `db:"id" json:"id"`
	SourceType   string    `db:"source_type" json:"source_type"`
	SourceID     int64     `db:"source_id" json:"source_id"`
	TargetType   string    `db:"target_type" json:"target_type"`
	TargetID     int64     `db:"target_id" json:"target_id"`
	RelationKind string    `db:"relation_kind" json:"relation_kind"`
	SortOrder    int       `db:"sort_order" json:"sort_order"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
}

type ContentRelationCreate struct {
	SourceType   string `json:"source_type"`
	SourceID     int64  `json:"source_id"`
	TargetType   string `json:"target_type"`
	TargetID     int64  `json:"target_id"`
	RelationKind string `json:"relation_kind"`
	SortOrder    int    `json:"sort_order"`
}

type ContentRelationView struct {
	ID           int64  `json:"id"`
	RelationKind string `json:"relation_kind"`
	TargetType   string `json:"target_type"`
	TargetID     int64  `json:"target_id"`
	Title        string `json:"title"`
	Slug         string `json:"slug"`
	Excerpt      string `json:"excerpt"`
	URL          string `json:"url"`
	Published    bool   `json:"published"`
}
