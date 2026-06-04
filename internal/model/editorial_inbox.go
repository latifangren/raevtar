package model

import "time"

const (
	EditorialModeScheduled     = "scheduled_assignment"
	EditorialModeOpportunistic = "opportunistic_assignment"
	EditorialModeCampaign      = "campaign_theme"
	EditorialModeSeed          = "autonomous_seed"
)

const (
	EditorialStatusQueued    = "queued"
	EditorialStatusApproved  = "approved"
	EditorialStatusPaused    = "paused"
	EditorialStatusRunning   = "running"
	EditorialStatusFailed    = "failed"
	EditorialStatusDone      = "done"
	EditorialStatusCancelled = "cancelled"
)

type EditorialInboxItem struct {
	ID              int64      `db:"id" json:"id"`
	SourceType      string     `db:"source_type" json:"source_type"`
	SourceValue     string     `db:"source_value" json:"source_value"`
	CategoryHint    string     `db:"category_hint" json:"category_hint"`
	Priority        int        `db:"priority" json:"priority"`
	NotBefore       time.Time  `db:"not_before" json:"not_before"`
	Deadline        *time.Time `db:"deadline" json:"deadline,omitempty"`
	Note            string     `db:"note" json:"note"`
	Mode            string     `db:"mode" json:"mode"`
	Status          string     `db:"status" json:"status"`
	PublishedPostID *int64     `db:"published_post_id" json:"published_post_id,omitempty"`
	FailureNote     string     `db:"failure_note" json:"failure_note,omitempty"`
	FailureMeta     string     `db:"failure_meta" json:"failure_meta,omitempty"`
	CreatedAt       time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time  `db:"updated_at" json:"updated_at"`
}

type EditorialInboxCreate struct {
	SourceType      string     `json:"source_type"`
	SourceValue     string     `json:"source_value"`
	CategoryHint    string     `json:"category_hint"`
	Priority        int        `json:"priority"`
	NotBefore       time.Time  `json:"not_before"`
	Deadline        *time.Time `json:"deadline"`
	Note            string     `json:"note"`
	Mode            string     `json:"mode"`
	Status          string     `json:"status"`
	PublishedPostID *int64     `json:"published_post_id"`
	FailureNote     string     `json:"failure_note"`
	FailureMeta     string     `json:"failure_meta"`
}

type EditorialInboxUpdate struct {
	SourceType      string     `json:"source_type"`
	SourceValue     string     `json:"source_value"`
	CategoryHint    string     `json:"category_hint"`
	Priority        int        `json:"priority"`
	NotBefore       time.Time  `json:"not_before"`
	Deadline        *time.Time `json:"deadline"`
	Note            string     `json:"note"`
	Mode            string     `json:"mode"`
	Status          string     `json:"status"`
	PublishedPostID *int64     `json:"published_post_id"`
	FailureNote     string     `json:"failure_note"`
	FailureMeta     string     `json:"failure_meta"`
}

func ValidEditorialModes() []string {
	return []string{
		EditorialModeScheduled,
		EditorialModeOpportunistic,
		EditorialModeCampaign,
		EditorialModeSeed,
	}
}

func ValidEditorialStatuses() []string {
	return []string{
		EditorialStatusQueued,
		EditorialStatusApproved,
		EditorialStatusPaused,
		EditorialStatusRunning,
		EditorialStatusFailed,
		EditorialStatusDone,
		EditorialStatusCancelled,
	}
}

func IsValidEditorialMode(mode string) bool {
	for _, valid := range ValidEditorialModes() {
		if mode == valid {
			return true
		}
	}
	return false
}

func IsValidEditorialStatus(status string) bool {
	for _, valid := range ValidEditorialStatuses() {
		if status == valid {
			return true
		}
	}
	return false
}
