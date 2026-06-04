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
	ClaimedBy       string     `db:"claimed_by" json:"claimed_by,omitempty"`
	ClaimTokenHash  string     `db:"claim_token_hash" json:"-"`
	ClaimedAt       *time.Time `db:"claimed_at" json:"claimed_at,omitempty"`
	LeaseExpiresAt  *time.Time `db:"lease_expires_at" json:"lease_expires_at,omitempty"`
	AttemptCount    int        `db:"attempt_count" json:"attempt_count"`
	CompletedAt     *time.Time `db:"completed_at" json:"completed_at,omitempty"`
	CreatedAt       time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time  `db:"updated_at" json:"updated_at"`
}

type EditorialFairnessSummary struct {
	NonUrgentClaimStreak   int  `json:"non_urgent_claim_streak"`
	AutonomousGapThreshold int  `json:"autonomous_gap_threshold"`
	AutonomousGapOpened    bool `json:"autonomous_gap_opened"`
}

type EditorialOverdueSummary struct {
	ApprovedCount  int `db:"approved_count" json:"approved_count"`
	RunningCount   int `db:"running_count" json:"running_count"`
	CompletedCount int `db:"completed_count" json:"completed_count"`
}

type EditorialModeAnalytics struct {
	Mode  string `json:"mode"`
	Count int    `json:"count"`
}

type EditorialPublishAnalytics struct {
	DoneCount               int                      `json:"done_count"`
	FailedCount             int                      `json:"failed_count"`
	CompletedWithPostCount  int                      `json:"completed_with_post_count"`
	AverageQueueWaitSeconds int64                    `json:"average_queue_wait_seconds"`
	AverageReadyWaitSeconds int64                    `json:"average_ready_wait_seconds"`
	ByMode                  []EditorialModeAnalytics `json:"by_mode"`
}

type EditorialInboxSummary struct {
	Fairness  EditorialFairnessSummary  `json:"fairness"`
	Overdue   EditorialOverdueSummary   `json:"overdue"`
	Analytics EditorialPublishAnalytics `json:"analytics"`
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

type EditorialInboxClaimRequest struct {
	Worker string `json:"worker"`
}

type EditorialInboxClaimResult struct {
	Item       *EditorialInboxItem `json:"item"`
	ClaimToken string              `json:"claim_token"`
}

type EditorialInboxCompleteRequest struct {
	ClaimToken      string `json:"claim_token"`
	PublishedPostID int64  `json:"published_post_id"`
}

type EditorialInboxFailRequest struct {
	ClaimToken  string `json:"claim_token"`
	FailureNote string `json:"failure_note"`
	FailureMeta string `json:"failure_meta"`
	Retryable   bool   `json:"retryable"`
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
