package service

import (
	"errors"
	"strings"
	"testing"
	"time"

	"raevtar/internal/model"
)

func TestEditorialInboxServiceCreate(t *testing.T) {
	state := newTestServices(t)
	now := time.Now().UTC().Truncate(time.Second)
	deadline := now.Add(24 * time.Hour)

	item, err := state.svc.Editorial.CreateInboxItem(model.EditorialInboxCreate{
		SourceType:   "repo",
		SourceValue:  "https://github.com/raevtar/test",
		CategoryHint: "devops",
		Priority:     75,
		NotBefore:    now,
		Deadline:     &deadline,
		Note:         "test item note",
		Mode:         model.EditorialModeScheduled,
		Status:       model.EditorialStatusApproved,
	})
	if err != nil {
		t.Fatalf("create inbox item: %v", err)
	}
	if item.ID == 0 {
		t.Fatalf("item id = 0, want positive")
	}
	checks := []struct{ name, got, want string }{
		{"source_type", item.SourceType, "repo"},
		{"source_value", item.SourceValue, "https://github.com/raevtar/test"},
		{"category_hint", item.CategoryHint, "devops"},
		{"note", item.Note, "test item note"},
		{"mode", item.Mode, model.EditorialModeScheduled},
		{"status", item.Status, model.EditorialStatusApproved},
	}
	for _, c := range checks {
		if c.got != c.want {
			t.Fatalf("%s = %q, want %q", c.name, c.got, c.want)
		}
	}
	if item.Priority != 75 {
		t.Fatalf("priority = %d, want 75", item.Priority)
	}
	if !item.NotBefore.Equal(now) {
		t.Fatalf("not_before = %s, want %s", item.NotBefore.Format(time.RFC3339), now.Format(time.RFC3339))
	}
	if item.Deadline == nil || !item.Deadline.Equal(deadline) {
		t.Fatalf("deadline = %v, want %v", item.Deadline, deadline.Format(time.RFC3339))
	}
	if item.AttemptCount != 0 {
		t.Fatalf("attempt_count = %d, want 0", item.AttemptCount)
	}
	if item.PublishedPostID != nil {
		t.Fatalf("published_post_id = %v, want nil for non-done item", item.PublishedPostID)
	}
	if item.CreatedAt.IsZero() {
		t.Fatalf("created_at should be set")
	}
}

func TestEditorialInboxServiceCreateMissingSource(t *testing.T) {
	state := newTestServices(t)
	now := time.Now().UTC()

	_, err := state.svc.Editorial.CreateInboxItem(model.EditorialInboxCreate{
		SourceType:  "",
		SourceValue: "",
		NotBefore:   now,
		Mode:        model.EditorialModeScheduled,
		Status:      model.EditorialStatusApproved,
	})
	if err == nil {
		t.Fatalf("expected error for missing source_type/source_value")
	}
	if !errors.Is(err, ErrInvalidEditorialInboxInput) {
		t.Fatalf("err = %v, want ErrInvalidEditorialInboxInput", err)
	}
	if !strings.Contains(err.Error(), "source_type") {
		t.Fatalf("err = %v, want source_type in error message", err)
	}
}

func TestEditorialInboxServiceGetInboxItem(t *testing.T) {
	state := newTestServices(t)
	now := time.Now().UTC().Truncate(time.Second)

	created, err := state.svc.Editorial.CreateInboxItem(model.EditorialInboxCreate{
		SourceType:  "topic",
		SourceValue: "kubernetes networking",
		NotBefore:   now,
		Mode:        model.EditorialModeScheduled,
		Status:      model.EditorialStatusApproved,
	})
	if err != nil {
		t.Fatalf("create inbox item: %v", err)
	}

	fetched, err := state.svc.Editorial.GetInboxItem(created.ID)
	if err != nil {
		t.Fatalf("get inbox item: %v", err)
	}
	if fetched.ID != created.ID {
		t.Fatalf("fetched id = %d, want %d", fetched.ID, created.ID)
	}
	if fetched.SourceValue != "kubernetes networking" {
		t.Fatalf("source_value = %q, want kubernetes networking", fetched.SourceValue)
	}
}

func TestEditorialInboxServiceGetInboxItemNotFound(t *testing.T) {
	state := newTestServices(t)

	_, err := state.svc.Editorial.GetInboxItem(99999)
	if err == nil {
		t.Fatalf("expected error for non-existent item")
	}
	if !errors.Is(err, ErrEditorialInboxNotFound) {
		t.Fatalf("err = %v, want ErrEditorialInboxNotFound", err)
	}
}

func TestEditorialInboxServiceListInboxItems(t *testing.T) {
	state := newTestServices(t)
	now := time.Now().UTC().Truncate(time.Second)

	for i := 0; i < 5; i++ {
		_, err := state.svc.Editorial.CreateInboxItem(model.EditorialInboxCreate{
			SourceType:  "repo",
			SourceValue: "https://github.com/raevtar/item-" + string(rune('a'+i)),
			NotBefore:   now,
			Mode:        model.EditorialModeScheduled,
			Status:      model.EditorialStatusApproved,
		})
		if err != nil {
			t.Fatalf("create item %d: %v", i, err)
		}
	}

	all, err := state.svc.Editorial.ListInboxItems(EditorialInboxListFilter{Limit: 10})
	if err != nil {
		t.Fatalf("list inbox items: %v", err)
	}
	if len(all) != 5 {
		t.Fatalf("list len = %d, want 5", len(all))
	}

	paginated, err := state.svc.Editorial.ListInboxItems(EditorialInboxListFilter{Limit: 2, Offset: 0})
	if err != nil {
		t.Fatalf("list paginated: %v", err)
	}
	if len(paginated) != 2 {
		t.Fatalf("paginated len = %d, want 2", len(paginated))
	}
}

func TestEditorialInboxServiceListInboxItemsByStatus(t *testing.T) {
	state := newTestServices(t)
	now := time.Now().UTC().Truncate(time.Second)

	_, err := state.svc.Editorial.CreateInboxItem(model.EditorialInboxCreate{
		SourceType:  "repo",
		SourceValue: "https://github.com/raevtar/approved",
		NotBefore:   now,
		Mode:        model.EditorialModeScheduled,
		Status:      model.EditorialStatusApproved,
	})
	if err != nil {
		t.Fatalf("create approved item: %v", err)
	}

	_, err = state.svc.Editorial.CreateInboxItem(model.EditorialInboxCreate{
		SourceType:  "repo",
		SourceValue: "https://github.com/raevtar/paused",
		NotBefore:   now,
		Mode:        model.EditorialModeOpportunistic,
		Status:      model.EditorialStatusPaused,
	})
	if err != nil {
		t.Fatalf("create paused item: %v", err)
	}

	approved, err := state.svc.Editorial.ListInboxItems(EditorialInboxListFilter{Status: model.EditorialStatusApproved, Limit: 10})
	if err != nil {
		t.Fatalf("list approved: %v", err)
	}
	if len(approved) != 1 {
		t.Fatalf("approved len = %d, want 1", len(approved))
	}
	if approved[0].SourceValue != "https://github.com/raevtar/approved" {
		t.Fatalf("approved source_value = %q, want approved value", approved[0].SourceValue)
	}

	paused, err := state.svc.Editorial.ListInboxItems(EditorialInboxListFilter{Status: model.EditorialStatusPaused, Limit: 10})
	if err != nil {
		t.Fatalf("list paused: %v", err)
	}
	if len(paused) != 1 {
		t.Fatalf("paused len = %d, want 1", len(paused))
	}
	if paused[0].SourceValue != "https://github.com/raevtar/paused" {
		t.Fatalf("paused source_value = %q, want paused value", paused[0].SourceValue)
	}
}

func TestEditorialInboxServiceClaimNextReady(t *testing.T) {
	state := newTestServices(t)
	now := time.Now().UTC().Truncate(time.Minute)

	item, err := state.svc.Editorial.CreateInboxItem(model.EditorialInboxCreate{
		SourceType:  "repo",
		SourceValue: "https://github.com/raevtar/claim-test",
		Priority:    60,
		NotBefore:   now.Add(-30 * time.Minute),
		Mode:        model.EditorialModeScheduled,
		Status:      model.EditorialStatusApproved,
	})
	if err != nil {
		t.Fatalf("create item: %v", err)
	}

	claim, err := state.svc.Editorial.ClaimNextInboxItem("test-worker", now)
	if err != nil {
		t.Fatalf("claim item: %v", err)
	}
	if claim.Item.ID != item.ID {
		t.Fatalf("claimed item id = %d, want %d", claim.Item.ID, item.ID)
	}
	if claim.Item.Status != model.EditorialStatusRunning {
		t.Fatalf("claimed status = %q, want running", claim.Item.Status)
	}
	if claim.Item.AttemptCount != 1 {
		t.Fatalf("attempt_count = %d, want 1", claim.Item.AttemptCount)
	}
	if claim.Item.ClaimedBy != "test-worker" {
		t.Fatalf("claimed_by = %q, want test-worker", claim.Item.ClaimedBy)
	}
	if claim.ClaimToken == "" {
		t.Fatalf("claim_token should not be empty")
	}
	if claim.Item.ClaimedAt == nil {
		t.Fatalf("claimed_at should be set")
	}
	if claim.Item.LeaseExpiresAt == nil {
		t.Fatalf("lease_expires_at should be set")
	}
}

func TestEditorialInboxServiceClaimNextReadyNoReady(t *testing.T) {
	state := newTestServices(t)
	now := time.Now().UTC()

	// Only create an approved item with future not_before — no ready items.
	_, err := state.svc.Editorial.CreateInboxItem(model.EditorialInboxCreate{
		SourceType:  "repo",
		SourceValue: "https://github.com/raevtar/future",
		NotBefore:   now.Add(2 * time.Hour),
		Mode:        model.EditorialModeScheduled,
		Status:      model.EditorialStatusApproved,
	})
	if err != nil {
		t.Fatalf("create future item: %v", err)
	}

	_, err = state.svc.Editorial.ClaimNextInboxItem("test-worker", now)
	if !errors.Is(err, ErrEditorialInboxNoClaimableItem) {
		t.Fatalf("err = %v, want ErrEditorialInboxNoClaimableItem", err)
	}
}

func TestEditorialInboxServiceCompleteClaim(t *testing.T) {
	state := newTestServices(t)
	now := time.Now().UTC().Truncate(time.Minute)

	item, err := state.svc.Editorial.CreateInboxItem(model.EditorialInboxCreate{
		SourceType:  "repo",
		SourceValue: "https://github.com/raevtar/complete-test",
		Priority:    50,
		NotBefore:   now.Add(-15 * time.Minute),
		Mode:        model.EditorialModeScheduled,
		Status:      model.EditorialStatusApproved,
	})
	if err != nil {
		t.Fatalf("create item: %v", err)
	}

	claim, err := state.svc.Editorial.ClaimNextInboxItem("worker-complete", now)
	if err != nil {
		t.Fatalf("claim item: %v", err)
	}
	if claim.Item.ID != item.ID {
		t.Fatalf("claimed item id = %d, want %d", claim.Item.ID, item.ID)
	}

	var publishedPostID int64 = 101
	done, err := state.svc.Editorial.CompleteInboxItemClaim(item.ID, claim.ClaimToken, publishedPostID)
	if err != nil {
		t.Fatalf("complete claim: %v", err)
	}
	if done.Status != model.EditorialStatusDone {
		t.Fatalf("status = %q, want done", done.Status)
	}
	if done.PublishedPostID == nil || *done.PublishedPostID != publishedPostID {
		t.Fatalf("published_post_id = %v, want %d", done.PublishedPostID, publishedPostID)
	}
	if done.ClaimedBy != "" {
		t.Fatalf("claimed_by = %q, want empty after completion", done.ClaimedBy)
	}
	if done.LeaseExpiresAt != nil {
		t.Fatalf("lease_expires_at should be nil after completion")
	}
	if done.CompletedAt == nil {
		t.Fatalf("completed_at should be set")
	}
}

func TestEditorialInboxServiceFailClaim(t *testing.T) {
	state := newTestServices(t)
	now := time.Now().UTC().Truncate(time.Minute)

	item, err := state.svc.Editorial.CreateInboxItem(model.EditorialInboxCreate{
		SourceType:  "repo",
		SourceValue: "https://github.com/raevtar/fail-test",
		Priority:    50,
		NotBefore:   now.Add(-15 * time.Minute),
		Mode:        model.EditorialModeScheduled,
		Status:      model.EditorialStatusApproved,
	})
	if err != nil {
		t.Fatalf("create item: %v", err)
	}

	claim, err := state.svc.Editorial.ClaimNextInboxItem("worker-fail", now)
	if err != nil {
		t.Fatalf("claim item: %v", err)
	}
	if claim.Item.ID != item.ID {
		t.Fatalf("claimed item id = %d, want %d", claim.Item.ID, item.ID)
	}
	if claim.Item.AttemptCount != 1 {
		t.Fatalf("attempt_count = %d, want 1", claim.Item.AttemptCount)
	}

	failed, err := state.svc.Editorial.FailInboxItemClaim(item.ID, claim.ClaimToken, "test failure", `{"reason":"timeout"}`, true, now)
	if err != nil {
		t.Fatalf("fail claim: %v", err)
	}
	if failed.Status != model.EditorialStatusApproved {
		t.Fatalf("retryable status = %q, want approved", failed.Status)
	}
	expectedNotBefore := now.Add(15 * time.Minute)
	if failed.NotBefore.Before(expectedNotBefore.Add(-time.Second)) {
		t.Fatalf("retryable not_before = %s, want around %s", failed.NotBefore.Format(time.RFC3339), expectedNotBefore.Format(time.RFC3339))
	}
	if failed.FailureNote != "test failure" {
		t.Fatalf("failure_note = %q, want test failure", failed.FailureNote)
	}
	if failed.FailureMeta != `{"reason":"timeout"}` {
		t.Fatalf("failure_meta = %q, want {\"reason\":\"timeout\"}", failed.FailureMeta)
	}
	if failed.AttemptCount != 1 {
		t.Fatalf("attempt_count = %d, want 1", failed.AttemptCount)
	}
}

func TestEditorialInboxServiceGetAnalytics(t *testing.T) {
	state := newTestServices(t)
	now := time.Now().UTC().Truncate(time.Minute)

	// Create a done item with published_post_id directly.
	postID1 := int64(201)
	_, err := state.svc.Editorial.CreateInboxItem(model.EditorialInboxCreate{
		SourceType:      "repo",
		SourceValue:     "https://github.com/raevtar/done-with-post",
		Priority:        50,
		NotBefore:       now.Add(-2 * time.Hour),
		Mode:            model.EditorialModeScheduled,
		Status:          model.EditorialStatusDone,
		PublishedPostID: &postID1,
	})
	if err != nil {
		t.Fatalf("create done with post: %v", err)
	}

	// Create an approved item, claim it, complete it with a post ID.
	item, err := state.svc.Editorial.CreateInboxItem(model.EditorialInboxCreate{
		SourceType:  "repo",
		SourceValue: "https://github.com/raevtar/claim-then-done",
		Priority:    40,
		NotBefore:   now.Add(-1 * time.Hour),
		Mode:        model.EditorialModeOpportunistic,
		Status:      model.EditorialStatusApproved,
	})
	if err != nil {
		t.Fatalf("create claimable item: %v", err)
	}
	claim, err := state.svc.Editorial.ClaimNextInboxItem("analytics-worker", now)
	if err != nil {
		t.Fatalf("claim item: %v", err)
	}
	postID2 := int64(202)
	_, err = state.svc.Editorial.CompleteInboxItemClaim(item.ID, claim.ClaimToken, postID2)
	if err != nil {
		t.Fatalf("complete claim: %v", err)
	}

	// Create another approved item, claim and fail it non-retryable.
	failItem, err := state.svc.Editorial.CreateInboxItem(model.EditorialInboxCreate{
		SourceType:  "topic",
		SourceValue: "failed-task",
		NotBefore:   now.Add(-3 * time.Hour),
		Mode:        model.EditorialModeScheduled,
		Status:      model.EditorialStatusApproved,
	})
	if err != nil {
		t.Fatalf("create fail item: %v", err)
	}
	claim2, err := state.svc.Editorial.ClaimNextInboxItem("analytics-worker", failItem.NotBefore)
	if err != nil {
		t.Fatalf("claim fail item: %v", err)
	}
	_, err = state.svc.Editorial.FailInboxItemClaim(failItem.ID, claim2.ClaimToken, "non retryable", `{}`, false, failItem.NotBefore)
	if err != nil {
		t.Fatalf("fail claim: %v", err)
	}

	counts, err := state.svc.Editorial.CountInboxStatuses()
	if err != nil {
		t.Fatalf("count statuses: %v", err)
	}
	if counts[model.EditorialStatusDone] != 2 {
		t.Fatalf("done count = %d, want 2", counts[model.EditorialStatusDone])
	}
	if counts[model.EditorialStatusFailed] != 1 {
		t.Fatalf("failed count = %d, want 1", counts[model.EditorialStatusFailed])
	}

	summary, err := state.svc.Editorial.GetInboxSummary(now.Add(1 * time.Hour))
	if err != nil {
		t.Fatalf("get summary: %v", err)
	}
	if summary.Analytics.DoneCount != 2 {
		t.Fatalf("analytics done count = %d, want 2", summary.Analytics.DoneCount)
	}
	if summary.Analytics.FailedCount != 1 {
		t.Fatalf("analytics failed count = %d, want 1", summary.Analytics.FailedCount)
	}
	if summary.Analytics.CompletedWithPostCount != 2 {
		t.Fatalf("completed with post count = %d, want 2", summary.Analytics.CompletedWithPostCount)
	}
	if len(summary.Analytics.ByMode) == 0 {
		t.Fatalf("expected by_mode analytics entries")
	}
}

// -- DeleteInboxItemNotFound -------------------------------------------------

func TestEditorialInboxDeleteInboxItemNotFound(t *testing.T) {
	state := newTestServices(t)

	_, err := state.svc.Editorial.DeleteInboxItem(99999)
	if err == nil {
		t.Fatalf("expected error for deleting non-existent inbox item")
	}
	if !errors.Is(err, ErrEditorialInboxNotFound) {
		t.Fatalf("err = %v, want ErrEditorialInboxNotFound", err)
	}
}

// -- sortReadyItems -----------------------------------------------------------

func TestEditorialInboxSortReadyItems(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Minute)
	past := now.Add(-1 * time.Hour)

	// High priority item
	high := model.EditorialInboxItem{
		ID:        1,
		Priority:  90,
		NotBefore: past,
		Status:    model.EditorialStatusApproved,
		CreatedAt: now.Add(-30 * time.Minute),
	}

	// Low priority item
	low := model.EditorialInboxItem{
		ID:        2,
		Priority:  10,
		NotBefore: past,
		Status:    model.EditorialStatusApproved,
		CreatedAt: now.Add(-30 * time.Minute),
	}

	items := []model.EditorialInboxItem{low, high}
	sortReadyItems(items, now)

	if len(items) != 2 {
		t.Fatalf("items len = %d, want 2", len(items))
	}
	if items[0].ID != high.ID {
		t.Fatalf("first item id = %d (priority %d), want id %d (priority %d)",
			items[0].ID, items[0].Priority, high.ID, high.Priority)
	}
	if items[1].ID != low.ID {
		t.Fatalf("second item id = %d (priority %d), want id %d (priority %d)",
			items[1].ID, items[1].Priority, low.ID, low.Priority)
	}

	// Test overdue items sort before non-overdue regardless of priority
	overdue := model.EditorialInboxItem{
		ID:        3,
		Priority:  5,
		NotBefore: past,
		Deadline:  &past,
		Status:    model.EditorialStatusApproved,
		CreatedAt: now.Add(-30 * time.Minute),
	}

	items = []model.EditorialInboxItem{high, overdue}
	sortReadyItems(items, now)

	if items[0].ID != overdue.ID {
		t.Fatalf("overdue item (id=%d) should sort before high priority (id=%d), first = %d",
			overdue.ID, high.ID, items[0].ID)
	}
}

// -- editorialRetryBackoff ---------------------------------------------------

func TestEditorialRetryBackoff(t *testing.T) {
	tests := []struct {
		attemptCount int
		want         time.Duration
	}{
		{attemptCount: 0, want: 15 * time.Minute},
		{attemptCount: 1, want: 15 * time.Minute},
		{attemptCount: 2, want: 1 * time.Hour},
		{attemptCount: 5, want: 6 * time.Hour},
	}

	for _, tt := range tests {
		got := editorialRetryBackoff(tt.attemptCount)
		if got != tt.want {
			t.Fatalf("editorialRetryBackoff(%d) = %v, want %v", tt.attemptCount, got, tt.want)
		}
	}
}
