package repo

import (
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	"raevtar/internal/model"
)

func TestEditorialInboxRepoCreateAndGetByID(t *testing.T) {
	db := InitSQLite(filepath.Join(t.TempDir(), "test.db"))
	t.Cleanup(func() { _ = db.Close() })
	AutoMigrate(db)
	repos := New(db)

	now := time.Now().Truncate(time.Second)
	deadline := now.Add(24 * time.Hour)
	item := &model.EditorialInboxItem{
		SourceType:   "github",
		SourceValue:  "raevtar/raevtar#42",
		CategoryHint: "tech",
		Priority:     75,
		NotBefore:    now.Add(-1 * time.Hour),
		Deadline:     &deadline,
		Note:         "Write a blog post about Go testing",
		Mode:         model.EditorialModeScheduled,
		Status:       model.EditorialStatusQueued,
		AttemptCount: 0,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := repos.EditorialInbox.Create(item); err != nil {
		t.Fatalf("create item: %v", err)
	}
	if item.ID == 0 {
		t.Fatal("expected item.ID to be set after Create")
	}

	loaded, err := repos.EditorialInbox.GetByID(item.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if loaded == nil {
		t.Fatal("GetByID returned nil")
	}

	if loaded.ID != item.ID {
		t.Errorf("ID: got %d, want %d", loaded.ID, item.ID)
	}
	if loaded.SourceType != item.SourceType {
		t.Errorf("SourceType: got %q, want %q", loaded.SourceType, item.SourceType)
	}
	if loaded.SourceValue != item.SourceValue {
		t.Errorf("SourceValue: got %q, want %q", loaded.SourceValue, item.SourceValue)
	}
	if loaded.CategoryHint != item.CategoryHint {
		t.Errorf("CategoryHint: got %q, want %q", loaded.CategoryHint, item.CategoryHint)
	}
	if loaded.Priority != item.Priority {
		t.Errorf("Priority: got %d, want %d", loaded.Priority, item.Priority)
	}
	if !loaded.NotBefore.Truncate(time.Second).Equal(now.Add(-1 * time.Hour).Truncate(time.Second)) {
		t.Errorf("NotBefore: got %v, want %v", loaded.NotBefore, now.Add(-1*time.Hour))
	}
	if loaded.Deadline == nil {
		t.Fatal("expected Deadline to be set")
	}
	if !loaded.Deadline.Truncate(time.Second).Equal(deadline.Truncate(time.Second)) {
		t.Errorf("Deadline: got %v, want %v", loaded.Deadline, deadline)
	}
	if loaded.Note != item.Note {
		t.Errorf("Note: got %q, want %q", loaded.Note, item.Note)
	}
	if loaded.Mode != item.Mode {
		t.Errorf("Mode: got %q, want %q", loaded.Mode, item.Mode)
	}
	if loaded.Status != item.Status {
		t.Errorf("Status: got %q, want %q", loaded.Status, item.Status)
	}
	if loaded.AttemptCount != 0 {
		t.Errorf("AttemptCount: got %d, want 0", loaded.AttemptCount)
	}
}

func TestEditorialInboxRepoUpdate(t *testing.T) {
	db := InitSQLite(filepath.Join(t.TempDir(), "test.db"))
	t.Cleanup(func() { _ = db.Close() })
	AutoMigrate(db)
	repos := New(db)

	now := time.Now().Truncate(time.Second)
	item := &model.EditorialInboxItem{
		SourceType:  "manual",
		SourceValue: "brainstorm",
		Priority:    50,
		NotBefore:   now,
		Mode:        model.EditorialModeSeed,
		Status:      model.EditorialStatusQueued,
		Note:        "original note",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := repos.EditorialInbox.Create(item); err != nil {
		t.Fatalf("create item: %v", err)
	}

	updatedAt := now.Add(1 * time.Hour)
	item.Status = model.EditorialStatusApproved
	item.Priority = 90
	item.Note = "updated note"
	item.UpdatedAt = updatedAt

	if err := repos.EditorialInbox.Update(item); err != nil {
		t.Fatalf("Update: %v", err)
	}

	loaded, err := repos.EditorialInbox.GetByID(item.ID)
	if err != nil {
		t.Fatalf("GetByID after update: %v", err)
	}

	if loaded.Status != model.EditorialStatusApproved {
		t.Errorf("Status: got %q, want %q", loaded.Status, model.EditorialStatusApproved)
	}
	if loaded.Priority != 90 {
		t.Errorf("Priority: got %d, want 90", loaded.Priority)
	}
	if loaded.Note != "updated note" {
		t.Errorf("Note: got %q, want %q", loaded.Note, "updated note")
	}
	if loaded.AttemptCount != 0 {
		t.Errorf("AttemptCount should remain 0, got %d", loaded.AttemptCount)
	}
	if !loaded.UpdatedAt.Equal(updatedAt) {
		t.Errorf("UpdatedAt: got %v, want %v", loaded.UpdatedAt, updatedAt)
	}
	if loaded.SourceValue != "brainstorm" {
		t.Errorf("SourceValue changed: got %q, want %q", loaded.SourceValue, "brainstorm")
	}
}

func TestEditorialInboxRepoDelete(t *testing.T) {
	db := InitSQLite(filepath.Join(t.TempDir(), "test.db"))
	t.Cleanup(func() { _ = db.Close() })
	AutoMigrate(db)
	repos := New(db)

	now := time.Now().Truncate(time.Second)
	item := &model.EditorialInboxItem{
		SourceType:  "test",
		SourceValue: "delete-test",
		Priority:    50,
		NotBefore:   now,
		Mode:        model.EditorialModeScheduled,
		Status:      model.EditorialStatusQueued,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := repos.EditorialInbox.Create(item); err != nil {
		t.Fatalf("create item: %v", err)
	}

	if err := repos.EditorialInbox.Delete(item.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	loaded, err := repos.EditorialInbox.GetByID(item.ID)
	if err == nil {
		t.Fatal("expected error after deleting item")
	}
	if err != sql.ErrNoRows {
		t.Fatalf("expected sql.ErrNoRows, got %v", err)
	}
	if loaded != nil {
		t.Fatal("expected nil item after deletion")
	}
}

func TestEditorialInboxRepoList(t *testing.T) {
	db := InitSQLite(filepath.Join(t.TempDir(), "test.db"))
	t.Cleanup(func() { _ = db.Close() })
	AutoMigrate(db)
	repos := New(db)

	now := time.Now().Truncate(time.Second)
	items := []*model.EditorialInboxItem{
		{
			SourceType: "github", SourceValue: "item-1", Priority: 50,
			NotBefore: now.Add(-2 * time.Hour), Mode: model.EditorialModeScheduled,
			Status: model.EditorialStatusApproved, Note: "item 1",
			CreatedAt: now, UpdatedAt: now,
		},
		{
			SourceType: "github", SourceValue: "item-2", Priority: 50,
			NotBefore: now.Add(-1 * time.Hour), Mode: model.EditorialModeSeed,
			Status: model.EditorialStatusPaused, Note: "item 2",
			CreatedAt: now, UpdatedAt: now,
		},
		{
			SourceType: "manual", SourceValue: "item-3", Priority: 50,
			NotBefore: now.Add(1 * time.Hour), Mode: model.EditorialModeScheduled,
			Status: model.EditorialStatusApproved, Note: "item 3",
			CreatedAt: now, UpdatedAt: now,
		},
	}
	for _, it := range items {
		if err := repos.EditorialInbox.Create(it); err != nil {
			t.Fatalf("create item %q: %v", it.SourceValue, err)
		}
	}

	t.Run("filter by status", func(t *testing.T) {
		result, err := repos.EditorialInbox.List(EditorialInboxFilter{Status: model.EditorialStatusApproved})
		if err != nil {
			t.Fatalf("List by status: %v", err)
		}
		if len(result) != 2 {
			t.Fatalf("expected 2 approved items, got %d", len(result))
		}
		for _, r := range result {
			if r.Status != model.EditorialStatusApproved {
				t.Errorf("item %d has status %q, want %q", r.ID, r.Status, model.EditorialStatusApproved)
			}
		}
	})

	t.Run("filter by mode", func(t *testing.T) {
		result, err := repos.EditorialInbox.List(EditorialInboxFilter{Mode: model.EditorialModeSeed})
		if err != nil {
			t.Fatalf("List by mode: %v", err)
		}
		if len(result) != 1 {
			t.Fatalf("expected 1 seed item, got %d", len(result))
		}
		if result[0].Mode != model.EditorialModeSeed {
			t.Errorf("item mode: got %q, want %q", result[0].Mode, model.EditorialModeSeed)
		}
	})

	t.Run("filter by ready", func(t *testing.T) {
		result, err := repos.EditorialInbox.List(EditorialInboxFilter{Ready: true, Now: now})
		if err != nil {
			t.Fatalf("List by ready: %v", err)
		}
		if len(result) != 1 {
			t.Fatalf("expected 1 ready item (approved+past), got %d", len(result))
		}
		if result[0].Status != model.EditorialStatusApproved {
			t.Errorf("ready item status: got %q, want %q", result[0].Status, model.EditorialStatusApproved)
		}
		if result[0].SourceValue != "item-1" {
			t.Errorf("expected item-1 as ready, got %q", result[0].SourceValue)
		}
	})

	t.Run("no filter returns all", func(t *testing.T) {
		result, err := repos.EditorialInbox.List(EditorialInboxFilter{})
		if err != nil {
			t.Fatalf("List with no filter: %v", err)
		}
		if len(result) != 3 {
			t.Fatalf("expected 3 items, got %d", len(result))
		}
	})
}

func TestEditorialInboxRepoListDefaults(t *testing.T) {
	db := InitSQLite(filepath.Join(t.TempDir(), "test.db"))
	t.Cleanup(func() { _ = db.Close() })
	AutoMigrate(db)
	repos := New(db)

	now := time.Now().Truncate(time.Second)
	for i := 0; i < 5; i++ {
		item := &model.EditorialInboxItem{
			SourceType: "test", SourceValue: "default-item",
			Priority: 50, NotBefore: now, Mode: model.EditorialModeScheduled,
			Status:    model.EditorialStatusQueued,
			CreatedAt: now, UpdatedAt: now,
		}
		if err := repos.EditorialInbox.Create(item); err != nil {
			t.Fatalf("create item %d: %v", i, err)
		}
	}

	// Limit=0, Offset=0 should return all items (no pagination applied)
	result, err := repos.EditorialInbox.List(EditorialInboxFilter{Limit: 0, Offset: 0})
	if err != nil {
		t.Fatalf("List with zero defaults: %v", err)
	}
	if len(result) != 5 {
		t.Fatalf("expected 5 items with limit=0, got %d", len(result))
	}
}

func TestEditorialInboxRepoClaimNextReady(t *testing.T) {
	db := InitSQLite(filepath.Join(t.TempDir(), "test.db"))
	t.Cleanup(func() { _ = db.Close() })
	AutoMigrate(db)
	repos := New(db)

	now := time.Now().Truncate(time.Second)
	item := &model.EditorialInboxItem{
		SourceType:  "github",
		SourceValue: "raevtar/raevtar#100",
		Priority:    80,
		NotBefore:   now.Add(-30 * time.Minute),
		Mode:        model.EditorialModeScheduled,
		Status:      model.EditorialStatusApproved,
		Note:        "Claim me",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := repos.EditorialInbox.Create(item); err != nil {
		t.Fatalf("create item: %v", err)
	}

	claimed, err := repos.EditorialInbox.ClaimNextReady(EditorialInboxClaimParams{
		Worker:         "test-worker",
		ClaimTokenHash: "claim-hash-abc",
		Now:            now,
		LeaseExpiresAt: now.Add(5 * time.Minute),
	})
	if err != nil {
		t.Fatalf("ClaimNextReady: %v", err)
	}
	if claimed == nil {
		t.Fatal("expected a claimed item, got nil")
	}
	if claimed.ID != item.ID {
		t.Errorf("claimed ID: got %d, want %d", claimed.ID, item.ID)
	}
	if claimed.Status != model.EditorialStatusRunning {
		t.Errorf("Status after claim: got %q, want %q", claimed.Status, model.EditorialStatusRunning)
	}
	if claimed.ClaimedBy != "test-worker" {
		t.Errorf("ClaimedBy: got %q, want %q", claimed.ClaimedBy, "test-worker")
	}
	if claimed.ClaimTokenHash != "claim-hash-abc" {
		t.Errorf("ClaimTokenHash: got %q, want %q", claimed.ClaimTokenHash, "claim-hash-abc")
	}
	if claimed.AttemptCount != 1 {
		t.Errorf("AttemptCount: got %d, want 1", claimed.AttemptCount)
	}
	if claimed.ClaimedAt == nil {
		t.Fatal("ClaimedAt should be set")
	}
	if claimed.LeaseExpiresAt == nil {
		t.Fatal("LeaseExpiresAt should be set")
	}
}

func TestEditorialInboxRepoClaimNextReadySkipsFuture(t *testing.T) {
	db := InitSQLite(filepath.Join(t.TempDir(), "test.db"))
	t.Cleanup(func() { _ = db.Close() })
	AutoMigrate(db)
	repos := New(db)

	now := time.Now().Truncate(time.Second)
	item := &model.EditorialInboxItem{
		SourceType:  "github",
		SourceValue: "future-item",
		Priority:    80,
		NotBefore:   now.Add(1 * time.Hour),
		Mode:        model.EditorialModeScheduled,
		Status:      model.EditorialStatusApproved,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := repos.EditorialInbox.Create(item); err != nil {
		t.Fatalf("create item: %v", err)
	}

	claimed, err := repos.EditorialInbox.ClaimNextReady(EditorialInboxClaimParams{
		Worker:         "worker",
		ClaimTokenHash: "hash",
		Now:            now,
		LeaseExpiresAt: now.Add(5 * time.Minute),
	})
	if err != nil {
		t.Fatalf("ClaimNextReady: %v", err)
	}
	if claimed != nil {
		t.Fatal("expected nil claim for future item, got a claimed item")
	}

	// Item should still be approved
	loaded, err := repos.EditorialInbox.GetByID(item.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if loaded.Status != model.EditorialStatusApproved {
		t.Errorf("Status should still be approved, got %q", loaded.Status)
	}
}

func TestEditorialInboxRepoClaimNextReadySkipsNonApproved(t *testing.T) {
	db := InitSQLite(filepath.Join(t.TempDir(), "test.db"))
	t.Cleanup(func() { _ = db.Close() })
	AutoMigrate(db)
	repos := New(db)

	now := time.Now().Truncate(time.Second)
	statuses := []string{
		model.EditorialStatusQueued,
		model.EditorialStatusPaused,
		model.EditorialStatusCancelled,
	}
	for _, s := range statuses {
		item := &model.EditorialInboxItem{
			SourceType: "test", SourceValue: "non-approved-" + s,
			Priority: 50, NotBefore: now.Add(-1 * time.Hour), Mode: model.EditorialModeScheduled,
			Status: s, CreatedAt: now, UpdatedAt: now,
		}
		if err := repos.EditorialInbox.Create(item); err != nil {
			t.Fatalf("create %s item: %v", s, err)
		}
	}

	claimed, err := repos.EditorialInbox.ClaimNextReady(EditorialInboxClaimParams{
		Worker:         "worker",
		ClaimTokenHash: "hash",
		Now:            now,
		LeaseExpiresAt: now.Add(5 * time.Minute),
	})
	if err != nil {
		t.Fatalf("ClaimNextReady: %v", err)
	}
	if claimed != nil {
		t.Fatal("expected nil claim when only non-approved items exist")
	}
}

func TestEditorialInboxRepoClaimNextReadyOrdersByPriority(t *testing.T) {
	db := InitSQLite(filepath.Join(t.TempDir(), "test.db"))
	t.Cleanup(func() { _ = db.Close() })
	AutoMigrate(db)
	repos := New(db)

	now := time.Now().Truncate(time.Second)

	lowPriority := &model.EditorialInboxItem{
		SourceType: "test", SourceValue: "low-priority",
		Priority: 10, NotBefore: now.Add(-1 * time.Hour), Mode: model.EditorialModeScheduled,
		Status: model.EditorialStatusApproved, CreatedAt: now, UpdatedAt: now,
	}
	if err := repos.EditorialInbox.Create(lowPriority); err != nil {
		t.Fatalf("create low priority item: %v", err)
	}

	highPriority := &model.EditorialInboxItem{
		SourceType: "test", SourceValue: "high-priority",
		Priority: 90, NotBefore: now.Add(-1 * time.Hour), Mode: model.EditorialModeScheduled,
		Status: model.EditorialStatusApproved, CreatedAt: now, UpdatedAt: now,
	}
	if err := repos.EditorialInbox.Create(highPriority); err != nil {
		t.Fatalf("create high priority item: %v", err)
	}

	claimed, err := repos.EditorialInbox.ClaimNextReady(EditorialInboxClaimParams{
		Worker:         "worker",
		ClaimTokenHash: "hash",
		Now:            now,
		LeaseExpiresAt: now.Add(5 * time.Minute),
	})
	if err != nil {
		t.Fatalf("ClaimNextReady: %v", err)
	}
	if claimed == nil {
		t.Fatal("expected a claimed item")
	}
	if claimed.ID != highPriority.ID {
		t.Errorf("expected high priority item (ID=%d, priority=90) to be claimed first, got ID=%d priority=%d",
			highPriority.ID, claimed.ID, claimed.Priority)
	}
}

func TestEditorialInboxRepoCompleteClaim(t *testing.T) {
	db := InitSQLite(filepath.Join(t.TempDir(), "test.db"))
	t.Cleanup(func() { _ = db.Close() })
	AutoMigrate(db)
	repos := New(db)

	now := time.Now().Truncate(time.Second)
	item := &model.EditorialInboxItem{
		SourceType: "test", SourceValue: "complete-me",
		Priority: 50, NotBefore: now.Add(-1 * time.Hour), Mode: model.EditorialModeScheduled,
		Status: model.EditorialStatusApproved, CreatedAt: now, UpdatedAt: now,
	}
	if err := repos.EditorialInbox.Create(item); err != nil {
		t.Fatalf("create item: %v", err)
	}

	claimed, err := repos.EditorialInbox.ClaimNextReady(EditorialInboxClaimParams{
		Worker:         "worker",
		ClaimTokenHash: "claim-hash-complete",
		Now:            now,
		LeaseExpiresAt: now.Add(5 * time.Minute),
	})
	if err != nil {
		t.Fatalf("ClaimNextReady: %v", err)
	}
	if claimed == nil {
		t.Fatal("expected a claimed item")
	}

	ok, err := repos.EditorialInbox.CompleteClaim(EditorialInboxCompletionParams{
		ID:              claimed.ID,
		ClaimTokenHash:  "claim-hash-complete",
		PublishedPostID: 42,
		Now:             now.Add(10 * time.Minute),
	})
	if err != nil {
		t.Fatalf("CompleteClaim: %v", err)
	}
	if !ok {
		t.Fatal("CompleteClaim returned false, expected true")
	}

	loaded, err := repos.EditorialInbox.GetByID(item.ID)
	if err != nil {
		t.Fatalf("GetByID after complete: %v", err)
	}
	if loaded.Status != model.EditorialStatusDone {
		t.Errorf("Status: got %q, want %q", loaded.Status, model.EditorialStatusDone)
	}
	if loaded.PublishedPostID == nil {
		t.Fatal("PublishedPostID should be set")
	}
	if *loaded.PublishedPostID != 42 {
		t.Errorf("PublishedPostID: got %d, want 42", *loaded.PublishedPostID)
	}
	if loaded.CompletedAt == nil {
		t.Fatal("CompletedAt should be set")
	}
	if loaded.ClaimedBy != "" {
		t.Errorf("ClaimedBy should be cleared, got %q", loaded.ClaimedBy)
	}
	if loaded.ClaimTokenHash != "" {
		t.Errorf("ClaimTokenHash should be cleared, got %q", loaded.ClaimTokenHash)
	}
}

func TestEditorialInboxRepoFailClaimRetryable(t *testing.T) {
	db := InitSQLite(filepath.Join(t.TempDir(), "test.db"))
	t.Cleanup(func() { _ = db.Close() })
	AutoMigrate(db)
	repos := New(db)

	now := time.Now().Truncate(time.Second)
	item := &model.EditorialInboxItem{
		SourceType: "test", SourceValue: "fail-retry",
		Priority: 50, NotBefore: now.Add(-1 * time.Hour), Mode: model.EditorialModeScheduled,
		Status: model.EditorialStatusApproved, CreatedAt: now, UpdatedAt: now,
	}
	if err := repos.EditorialInbox.Create(item); err != nil {
		t.Fatalf("create item: %v", err)
	}

	claimed, err := repos.EditorialInbox.ClaimNextReady(EditorialInboxClaimParams{
		Worker:         "worker",
		ClaimTokenHash: "claim-hash-fail",
		Now:            now,
		LeaseExpiresAt: now.Add(5 * time.Minute),
	})
	if err != nil {
		t.Fatalf("ClaimNextReady: %v", err)
	}
	if claimed == nil {
		t.Fatal("expected a claimed item")
	}

	retryNotBefore := now.Add(1 * time.Hour)
	ok, err := repos.EditorialInbox.FailClaim(EditorialInboxFailureParams{
		ID:             claimed.ID,
		ClaimTokenHash: "claim-hash-fail",
		Status:         model.EditorialStatusApproved,
		NotBefore:      retryNotBefore,
		FailureNote:    "temporary network error",
		FailureMeta:    `{"error":"timeout","code":503}`,
		Now:            now,
	})
	if err != nil {
		t.Fatalf("FailClaim: %v", err)
	}
	if !ok {
		t.Fatal("FailClaim returned false, expected true")
	}

	loaded, err := repos.EditorialInbox.GetByID(item.ID)
	if err != nil {
		t.Fatalf("GetByID after fail: %v", err)
	}
	if loaded.Status != model.EditorialStatusApproved {
		t.Errorf("Status: got %q, want %q (reverted after retryable fail)", loaded.Status, model.EditorialStatusApproved)
	}
	if !loaded.NotBefore.Truncate(time.Second).Equal(retryNotBefore.Truncate(time.Second)) {
		t.Errorf("NotBefore: got %v, want %v", loaded.NotBefore, retryNotBefore)
	}
	if loaded.FailureNote != "temporary network error" {
		t.Errorf("FailureNote: got %q, want %q", loaded.FailureNote, "temporary network error")
	}
	if loaded.FailureMeta != `{"error":"timeout","code":503}` {
		t.Errorf("FailureMeta: got %q, want %q", loaded.FailureMeta, `{"error":"timeout","code":503}`)
	}
	if loaded.ClaimedBy != "" {
		t.Errorf("ClaimedBy should be cleared, got %q", loaded.ClaimedBy)
	}
	if loaded.ClaimTokenHash != "" {
		t.Errorf("ClaimTokenHash should be cleared, got %q", loaded.ClaimTokenHash)
	}
	if loaded.AttemptCount != 1 {
		t.Errorf("AttemptCount should remain 1, got %d", loaded.AttemptCount)
	}

	// Item should be claimable again after the new not_before passes
	item2 := &model.EditorialInboxItem{
		SourceType: "test", SourceValue: "another-item",
		Priority: 50, NotBefore: now.Add(-1 * time.Hour), Mode: model.EditorialModeScheduled,
		Status: model.EditorialStatusApproved, CreatedAt: now, UpdatedAt: now,
	}
	if err := repos.EditorialInbox.Create(item2); err != nil {
		t.Fatalf("create another item: %v", err)
	}

	claimed2, err := repos.EditorialInbox.ClaimNextReady(EditorialInboxClaimParams{
		Worker:         "worker",
		ClaimTokenHash: "hash2",
		Now:            now,
		LeaseExpiresAt: now.Add(5 * time.Minute),
	})
	if err != nil {
		t.Fatalf("ClaimNextReady after fail: %v", err)
	}
	if claimed2 == nil {
		t.Fatal("expected a different item to be claimable after fail")
	}
	if claimed2.ID == item.ID {
		t.Errorf("failed item (ID=%d) should not be claimable yet since NotBefore is in the future", item.ID)
	}
}

func TestEditorialInboxRepoCountByStatus(t *testing.T) {
	db := InitSQLite(filepath.Join(t.TempDir(), "test.db"))
	t.Cleanup(func() { _ = db.Close() })
	AutoMigrate(db)
	repos := New(db)

	now := time.Now().Truncate(time.Second)
	type statusCount struct {
		status string
		count  int
	}
	scenarios := []statusCount{
		{model.EditorialStatusQueued, 3},
		{model.EditorialStatusApproved, 2},
		{model.EditorialStatusRunning, 1},
		{model.EditorialStatusDone, 4},
		{model.EditorialStatusFailed, 1},
		{model.EditorialStatusPaused, 2},
	}

	for _, sc := range scenarios {
		for i := 0; i < sc.count; i++ {
			item := &model.EditorialInboxItem{
				SourceType: "test", SourceValue: sc.status + "-" + string(rune('0'+i)),
				Priority: 50, NotBefore: now, Mode: model.EditorialModeScheduled,
				Status: sc.status, CreatedAt: now, UpdatedAt: now,
			}
			if err := repos.EditorialInbox.Create(item); err != nil {
				t.Fatalf("create %s item: %v", sc.status, err)
			}
		}
	}

	counts, err := repos.EditorialInbox.CountByStatus()
	if err != nil {
		t.Fatalf("CountByStatus: %v", err)
	}

	for _, sc := range scenarios {
		got := counts[sc.status]
		if got != sc.count {
			t.Errorf("CountByStatus[%q]: got %d, want %d", sc.status, got, sc.count)
		}
	}
}

func TestEditorialInboxRepoPolicyState(t *testing.T) {
	db := InitSQLite(filepath.Join(t.TempDir(), "test.db"))
	t.Cleanup(func() { _ = db.Close() })
	AutoMigrate(db)
	repos := New(db)

	now := time.Now().Truncate(time.Second)

	// Get a non-existent policy key
	val, err := repos.EditorialInbox.GetPolicyState("non_existent_key")
	if err != nil {
		t.Fatalf("GetPolicyState for non-existent key: %v", err)
	}
	if val != 0 {
		t.Errorf("expected 0 for non-existent key, got %d", val)
	}

	// Set and get
	if err := repos.EditorialInbox.SetPolicyState("max_claims_per_day", 5, now); err != nil {
		t.Fatalf("SetPolicyState: %v", err)
	}

	val, err = repos.EditorialInbox.GetPolicyState("max_claims_per_day")
	if err != nil {
		t.Fatalf("GetPolicyState: %v", err)
	}
	if val != 5 {
		t.Errorf("expected 5, got %d", val)
	}

	// Update existing
	if err := repos.EditorialInbox.SetPolicyState("max_claims_per_day", 10, now.Add(1*time.Hour)); err != nil {
		t.Fatalf("SetPolicyState update: %v", err)
	}

	val, err = repos.EditorialInbox.GetPolicyState("max_claims_per_day")
	if err != nil {
		t.Fatalf("GetPolicyState after update: %v", err)
	}
	if val != 10 {
		t.Errorf("expected 10, got %d", val)
	}

	// Separate key doesn't interfere
	val2, err := repos.EditorialInbox.GetPolicyState("another_key")
	if err != nil {
		t.Fatalf("GetPolicyState for another key: %v", err)
	}
	if val2 != 0 {
		t.Errorf("expected 0 for non-existent key, got %d", val2)
	}
}

func TestEditorialInboxRepoCountOverdue(t *testing.T) {
	db := InitSQLite(filepath.Join(t.TempDir(), "test.db"))
	t.Cleanup(func() { _ = db.Close() })
	AutoMigrate(db)
	repos := New(db)

	now := time.Now().Truncate(time.Second)

	// Create items without deadline — we set deadline via SQL datetime()
	// to ensure SQLite stores it in a format compatible with date comparison.
	items := []*model.EditorialInboxItem{
		{SourceType: "test", SourceValue: "overdue-approved", Priority: 50, NotBefore: now.Add(-3 * time.Hour), Mode: model.EditorialModeScheduled, Status: model.EditorialStatusApproved, CreatedAt: now, UpdatedAt: now},
		{SourceType: "test", SourceValue: "future-approved", Priority: 50, NotBefore: now.Add(-3 * time.Hour), Mode: model.EditorialModeScheduled, Status: model.EditorialStatusApproved, CreatedAt: now, UpdatedAt: now},
		{SourceType: "test", SourceValue: "overdue-running", Priority: 50, NotBefore: now.Add(-3 * time.Hour), Mode: model.EditorialModeScheduled, Status: model.EditorialStatusRunning, CreatedAt: now, UpdatedAt: now},
		{SourceType: "test", SourceValue: "no-deadline-running", Priority: 50, NotBefore: now.Add(-3 * time.Hour), Mode: model.EditorialModeScheduled, Status: model.EditorialStatusRunning, CreatedAt: now, UpdatedAt: now},
		{SourceType: "test", SourceValue: "overdue-done", Priority: 50, NotBefore: now.Add(-3 * time.Hour), Mode: model.EditorialModeScheduled, Status: model.EditorialStatusDone, CreatedAt: now, UpdatedAt: now},
		{SourceType: "test", SourceValue: "on-time-done", Priority: 50, NotBefore: now.Add(-3 * time.Hour), Mode: model.EditorialModeScheduled, Status: model.EditorialStatusDone, CreatedAt: now, UpdatedAt: now},
	}
	for _, it := range items {
		if err := repos.EditorialInbox.Create(it); err != nil {
			t.Fatalf("create %q: %v", it.SourceValue, err)
		}
	}

	// Use SQL datetime() for deadline (2 hours ago = overdue, 2 hours ahead = not overdue)
	if _, err := db.Exec("UPDATE editorial_inbox SET deadline = datetime('now', '-2 hours') WHERE source_value = 'overdue-approved'"); err != nil {
		t.Fatalf("set overdue-approved deadline: %v", err)
	}
	if _, err := db.Exec("UPDATE editorial_inbox SET deadline = datetime('now', '2 hours') WHERE source_value = 'future-approved'"); err != nil {
		t.Fatalf("set future-approved deadline: %v", err)
	}
	if _, err := db.Exec("UPDATE editorial_inbox SET deadline = datetime('now', '-2 hours') WHERE source_value = 'overdue-running'"); err != nil {
		t.Fatalf("set overdue-running deadline: %v", err)
	}
	if _, err := db.Exec("UPDATE editorial_inbox SET deadline = datetime('now', '-2 hours'), completed_at = datetime('now') WHERE source_value = 'overdue-done'"); err != nil {
		t.Fatalf("set overdue-done deadline and completed_at: %v", err)
	}
	if _, err := db.Exec("UPDATE editorial_inbox SET deadline = datetime('now', '2 hours'), completed_at = datetime('now', '-1 hours') WHERE source_value = 'on-time-done'"); err != nil {
		t.Fatalf("set on-time-done deadline and completed_at: %v", err)
	}

	summary, err := repos.EditorialInbox.CountOverdue(now)
	if err != nil {
		t.Fatalf("CountOverdue: %v", err)
	}

	if summary.ApprovedCount != 1 {
		t.Errorf("ApprovedCount: got %d, want 1", summary.ApprovedCount)
	}
	if summary.RunningCount != 1 {
		t.Errorf("RunningCount: got %d, want 1", summary.RunningCount)
	}
	if summary.CompletedCount != 1 {
		t.Errorf("CompletedCount: got %d, want 1", summary.CompletedCount)
	}
}

func TestEditorialInboxRepoGetAnalytics(t *testing.T) {
	db := InitSQLite(filepath.Join(t.TempDir(), "test.db"))
	t.Cleanup(func() { _ = db.Close() })
	AutoMigrate(db)
	repos := New(db)

	now := time.Now().Truncate(time.Second)
	postID := int64(100)

	// Create items with basic fields, then set timestamps via SQL datetime()
	// to ensure julianday() can parse them.
	doneItems := []*model.EditorialInboxItem{
		{SourceType: "test", SourceValue: "done-with-post-1", Priority: 50, NotBefore: now.Add(-2 * time.Hour), Mode: model.EditorialModeScheduled, Status: model.EditorialStatusDone, PublishedPostID: &postID, CreatedAt: now, UpdatedAt: now},
		{SourceType: "test", SourceValue: "done-with-post-2", Priority: 50, NotBefore: now.Add(-1 * time.Hour), Mode: model.EditorialModeOpportunistic, Status: model.EditorialStatusDone, PublishedPostID: &postID, CreatedAt: now, UpdatedAt: now},
	}
	for _, it := range doneItems {
		if err := repos.EditorialInbox.Create(it); err != nil {
			t.Fatalf("create done item %q: %v", it.SourceValue, err)
		}
	}

	failedItem := &model.EditorialInboxItem{
		SourceType: "test", SourceValue: "failed-item", Priority: 50, NotBefore: now.Add(-1 * time.Hour), Mode: model.EditorialModeScheduled, Status: model.EditorialStatusFailed, CreatedAt: now, UpdatedAt: now,
	}
	if err := repos.EditorialInbox.Create(failedItem); err != nil {
		t.Fatalf("create failed item: %v", err)
	}

	// Set precise timestamps via SQL datetime() so julianday() can compute deltas.
	// Item 1: created 5h ago, ready 2h ago, completed now → queue=18000s, ready=7200s
	// Item 2: created 3h ago, ready 1h ago, completed now → queue=10800s, ready=3600s
	if _, err := db.Exec(`UPDATE editorial_inbox SET created_at = datetime('now', '-5 hours'), not_before = datetime('now', '-2 hours'), completed_at = datetime('now') WHERE source_value = 'done-with-post-1'`); err != nil {
		t.Fatalf("set timestamps for done-with-post-1: %v", err)
	}
	if _, err := db.Exec(`UPDATE editorial_inbox SET created_at = datetime('now', '-3 hours'), not_before = datetime('now', '-1 hours'), completed_at = datetime('now') WHERE source_value = 'done-with-post-2'`); err != nil {
		t.Fatalf("set timestamps for done-with-post-2: %v", err)
	}

	analytics, err := repos.EditorialInbox.GetAnalytics()
	if err != nil {
		t.Fatalf("GetAnalytics: %v", err)
	}

	if analytics.DoneCount != 2 {
		t.Errorf("DoneCount: got %d, want 2", analytics.DoneCount)
	}
	if analytics.FailedCount != 1 {
		t.Errorf("FailedCount: got %d, want 1", analytics.FailedCount)
	}
	if analytics.CompletedWithPostCount != 2 {
		t.Errorf("CompletedWithPostCount: got %d, want 2", analytics.CompletedWithPostCount)
	}
	// AverageQueueWaitSeconds: (18000 + 10800) / 2 = 14400
	// SQLite stores DATETIME with second precision, so allow ±1s tolerance
	if analytics.AverageQueueWaitSeconds < 14399 || analytics.AverageQueueWaitSeconds > 14401 {
		t.Errorf("AverageQueueWaitSeconds: got %d, want ~14400", analytics.AverageQueueWaitSeconds)
	}
	// AverageReadyWaitSeconds: (7200 + 3600) / 2 = 5400
	if analytics.AverageReadyWaitSeconds < 5399 || analytics.AverageReadyWaitSeconds > 5401 {
		t.Errorf("AverageReadyWaitSeconds: got %d, want ~5400", analytics.AverageReadyWaitSeconds)
	}
	if analytics.OverdueCompletedCount != 0 {
		t.Errorf("OverdueCompletedCount: got %d, want 0", analytics.OverdueCompletedCount)
	}
}

func TestEditorialInboxRepoCountDoneByMode(t *testing.T) {
	db := InitSQLite(filepath.Join(t.TempDir(), "test.db"))
	t.Cleanup(func() { _ = db.Close() })
	AutoMigrate(db)
	repos := New(db)

	now := time.Now().Truncate(time.Second)
	items := []*model.EditorialInboxItem{
		{SourceType: "test", SourceValue: "scheduled-1", Priority: 50, NotBefore: now, Mode: model.EditorialModeScheduled, Status: model.EditorialStatusDone, CreatedAt: now, UpdatedAt: now},
		{SourceType: "test", SourceValue: "scheduled-2", Priority: 50, NotBefore: now, Mode: model.EditorialModeScheduled, Status: model.EditorialStatusDone, CreatedAt: now, UpdatedAt: now},
		{SourceType: "test", SourceValue: "opportunistic-1", Priority: 50, NotBefore: now, Mode: model.EditorialModeOpportunistic, Status: model.EditorialStatusDone, CreatedAt: now, UpdatedAt: now},
		{SourceType: "test", SourceValue: "seed-1", Priority: 50, NotBefore: now, Mode: model.EditorialModeSeed, Status: model.EditorialStatusFailed, CreatedAt: now, UpdatedAt: now},
	}
	for _, it := range items {
		if err := repos.EditorialInbox.Create(it); err != nil {
			t.Fatalf("create %q: %v", it.SourceValue, err)
		}
	}

	rows, err := repos.EditorialInbox.CountDoneByMode()
	if err != nil {
		t.Fatalf("CountDoneByMode: %v", err)
	}

	if len(rows) != 2 {
		t.Fatalf("expected 2 modes with done items, got %d", len(rows))
	}

	for _, row := range rows {
		switch row.Mode {
		case model.EditorialModeScheduled:
			if row.Count != 2 {
				t.Errorf("Count for scheduled: got %d, want 2", row.Count)
			}
		case model.EditorialModeOpportunistic:
			if row.Count != 1 {
				t.Errorf("Count for opportunistic: got %d, want 1", row.Count)
			}
		default:
			t.Errorf("unexpected mode in results: %q", row.Mode)
		}
	}
}
