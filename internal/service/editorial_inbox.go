package service

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"raevtar/internal/model"
	"raevtar/internal/repo"
)

var ErrInvalidEditorialInboxInput = errors.New("invalid editorial inbox input")
var ErrEditorialInboxNotFound = errors.New("editorial inbox item not found")
var ErrEditorialInboxNoClaimableItem = errors.New("no claimable editorial inbox item")
var ErrEditorialInboxInvalidClaim = errors.New("invalid editorial inbox claim")

const editorialInboxLeaseTTL = 30 * time.Minute
const editorialFairnessGapThreshold = 3
const editorialPolicyStateNonUrgentClaimStreak = "non_urgent_claim_streak"
const editorialPolicyStateAutonomousGapOpened = "autonomous_gap_opened"

type EditorialInboxListFilter struct {
	Status string
	Mode   string
	Ready  bool
	Limit  int
	Offset int
}

type EditorialInboxService struct {
	repos *repo.Repositories
}

func NewEditorialInboxService(repos *repo.Repositories) *EditorialInboxService {
	return &EditorialInboxService{repos: repos}
}

func (s *EditorialInboxService) CreateInboxItem(input model.EditorialInboxCreate) (*model.EditorialInboxItem, error) {
	item, err := s.buildInboxItem(0, input.SourceType, input.SourceValue, input.CategoryHint, input.Priority, input.NotBefore, input.Deadline, input.Note, input.Mode, input.Status, input.PublishedPostID, input.FailureNote, input.FailureMeta)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	item.CreatedAt = now
	item.UpdatedAt = now
	if err := s.repos.EditorialInbox.Create(item); err != nil {
		return nil, fmt.Errorf("create editorial inbox item: %w", err)
	}
	return item, nil
}

func (s *EditorialInboxService) ListInboxItems(filter EditorialInboxListFilter) ([]model.EditorialInboxItem, error) {
	items, err := s.repos.EditorialInbox.List(repo.EditorialInboxFilter{
		Status: strings.TrimSpace(filter.Status),
		Mode:   strings.TrimSpace(filter.Mode),
		Ready:  filter.Ready,
		Now:    time.Now().UTC(),
		Limit:  filter.Limit,
		Offset: filter.Offset,
	})
	if err != nil {
		return nil, fmt.Errorf("list editorial inbox items: %w", err)
	}
	return items, nil
}

func (s *EditorialInboxService) GetInboxItem(id int64) (*model.EditorialInboxItem, error) {
	item, err := s.repos.EditorialInbox.GetByID(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w: %w", ErrEditorialInboxNotFound, err)
		}
		return nil, fmt.Errorf("get editorial inbox item: %w", err)
	}
	return item, nil
}

func (s *EditorialInboxService) UpdateInboxItem(id int64, input model.EditorialInboxUpdate) (*model.EditorialInboxItem, error) {
	existing, err := s.repos.EditorialInbox.GetByID(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w: %w", ErrEditorialInboxNotFound, err)
		}
		return nil, fmt.Errorf("get editorial inbox item: %w", err)
	}
	item, err := s.buildInboxItem(id, input.SourceType, input.SourceValue, input.CategoryHint, input.Priority, input.NotBefore, input.Deadline, input.Note, input.Mode, input.Status, input.PublishedPostID, input.FailureNote, input.FailureMeta)
	if err != nil {
		return nil, err
	}
	item.AttemptCount = existing.AttemptCount
	item.CompletedAt = existing.CompletedAt
	item.UpdatedAt = time.Now().UTC()
	if err := s.repos.EditorialInbox.Update(item); err != nil {
		return nil, fmt.Errorf("update editorial inbox item: %w", err)
	}
	return s.GetInboxItem(id)
}

func (s *EditorialInboxService) ClaimNextInboxItem(worker string, now time.Time) (*model.EditorialInboxClaimResult, error) {
	worker = strings.TrimSpace(worker)
	if worker == "" {
		return nil, fmt.Errorf("%w: worker required", ErrInvalidEditorialInboxInput)
	}
	policyNow := now.UTC()
	ready, err := s.listClaimCandidates(policyNow, 100)
	if err != nil {
		return nil, err
	}
	if len(ready) == 0 {
		return nil, ErrEditorialInboxNoClaimableItem
	}
	urgentExists := false
	for _, candidate := range ready {
		if isEditorialOverdue(candidate, policyNow) || isStaleRunningClaim(candidate, policyNow) {
			urgentExists = true
			break
		}
	}
	if !urgentExists {
		streak, err := s.repos.EditorialInbox.GetPolicyState(editorialPolicyStateNonUrgentClaimStreak)
		if err != nil {
			return nil, fmt.Errorf("get fairness streak: %w", err)
		}
		if streak >= editorialFairnessGapThreshold {
			if err := s.repos.EditorialInbox.SetPolicyState(editorialPolicyStateNonUrgentClaimStreak, 0, policyNow); err != nil {
				return nil, fmt.Errorf("reset fairness streak: %w", err)
			}
			if err := s.repos.EditorialInbox.SetPolicyState(editorialPolicyStateAutonomousGapOpened, 1, policyNow); err != nil {
				return nil, fmt.Errorf("set autonomous gap flag: %w", err)
			}
			return nil, ErrEditorialInboxNoClaimableItem
		}
	}
	if err := s.repos.EditorialInbox.SetPolicyState(editorialPolicyStateAutonomousGapOpened, 0, policyNow); err != nil {
		return nil, fmt.Errorf("clear autonomous gap flag: %w", err)
	}
	claimToken, claimTokenHash, err := newEditorialClaimToken()
	if err != nil {
		return nil, fmt.Errorf("create editorial claim token: %w", err)
	}
	item, err := s.repos.EditorialInbox.ClaimNextReady(repo.EditorialInboxClaimParams{
		ItemID:         ready[0].ID,
		Worker:         worker,
		ClaimTokenHash: claimTokenHash,
		Now:            policyNow,
		LeaseExpiresAt: policyNow.Add(editorialInboxLeaseTTL),
	})
	if err != nil {
		return nil, fmt.Errorf("claim editorial inbox item: %w", err)
	}
	if item == nil {
		return nil, ErrEditorialInboxNoClaimableItem
	}
	if isEditorialOverdue(*item, policyNow) || isStaleRunningClaim(*item, policyNow) {
		if err := s.repos.EditorialInbox.SetPolicyState(editorialPolicyStateNonUrgentClaimStreak, 0, policyNow); err != nil {
			return nil, fmt.Errorf("reset fairness streak after urgent claim: %w", err)
		}
	} else {
		streak, err := s.repos.EditorialInbox.GetPolicyState(editorialPolicyStateNonUrgentClaimStreak)
		if err != nil {
			return nil, fmt.Errorf("get fairness streak after claim: %w", err)
		}
		if err := s.repos.EditorialInbox.SetPolicyState(editorialPolicyStateNonUrgentClaimStreak, streak+1, policyNow); err != nil {
			return nil, fmt.Errorf("increment fairness streak: %w", err)
		}
	}
	return &model.EditorialInboxClaimResult{Item: item, ClaimToken: claimToken}, nil
}

func (s *EditorialInboxService) CompleteInboxItemClaim(id int64, claimToken string, publishedPostID int64) (*model.EditorialInboxItem, error) {
	claimToken = strings.TrimSpace(claimToken)
	if claimToken == "" {
		return nil, fmt.Errorf("%w: claim_token required", ErrInvalidEditorialInboxInput)
	}
	if publishedPostID <= 0 {
		return nil, fmt.Errorf("%w: published_post_id must be positive", ErrInvalidEditorialInboxInput)
	}
	ok, err := s.repos.EditorialInbox.CompleteClaim(repo.EditorialInboxCompletionParams{
		ID:              id,
		ClaimTokenHash:  editorialClaimTokenHash(claimToken),
		PublishedPostID: publishedPostID,
		Now:             time.Now().UTC(),
	})
	if err != nil {
		return nil, fmt.Errorf("complete editorial inbox claim: %w", err)
	}
	if !ok {
		return nil, ErrEditorialInboxInvalidClaim
	}
	return s.GetInboxItem(id)
}

func (s *EditorialInboxService) FailInboxItemClaim(id int64, claimToken, failureNote, failureMeta string, retryable bool, now time.Time) (*model.EditorialInboxItem, error) {
	claimToken = strings.TrimSpace(claimToken)
	failureNote = strings.TrimSpace(failureNote)
	failureMeta = strings.TrimSpace(failureMeta)
	if claimToken == "" {
		return nil, fmt.Errorf("%w: claim_token required", ErrInvalidEditorialInboxInput)
	}
	if failureNote == "" {
		return nil, fmt.Errorf("%w: failure_note required", ErrInvalidEditorialInboxInput)
	}
	status := model.EditorialStatusFailed
	notBefore := now.UTC()
	if retryable {
		item, err := s.GetInboxItem(id)
		if err != nil {
			return nil, err
		}
		status = model.EditorialStatusApproved
		notBefore = now.UTC().Add(editorialRetryBackoff(item.AttemptCount))
	}
	ok, err := s.repos.EditorialInbox.FailClaim(repo.EditorialInboxFailureParams{
		ID:             id,
		ClaimTokenHash: editorialClaimTokenHash(claimToken),
		Status:         status,
		NotBefore:      notBefore,
		FailureNote:    failureNote,
		FailureMeta:    failureMeta,
		Now:            now.UTC(),
	})
	if err != nil {
		return nil, fmt.Errorf("fail editorial inbox claim: %w", err)
	}
	if !ok {
		return nil, ErrEditorialInboxInvalidClaim
	}
	return s.GetInboxItem(id)
}

func (s *EditorialInboxService) ListReadyInboxItems(now time.Time, limit int) ([]model.EditorialInboxItem, error) {
	items, err := s.repos.EditorialInbox.List(repo.EditorialInboxFilter{
		Ready: true,
		Now:   now.UTC(),
		Limit: limit,
	})
	if err != nil {
		return nil, fmt.Errorf("list ready editorial inbox items: %w", err)
	}
	sortReadyItems(items, now.UTC())
	return items, nil
}

func (s *EditorialInboxService) listClaimCandidates(now time.Time, limit int) ([]model.EditorialInboxItem, error) {
	items, err := s.repos.EditorialInbox.List(repo.EditorialInboxFilter{Now: now.UTC(), Limit: 0})
	if err != nil {
		return nil, fmt.Errorf("list editorial claim candidates: %w", err)
	}
	out := make([]model.EditorialInboxItem, 0, len(items))
	for _, item := range items {
		if (item.Status == model.EditorialStatusApproved && !item.NotBefore.After(now)) || (item.Status == model.EditorialStatusRunning && item.LeaseExpiresAt != nil && !item.LeaseExpiresAt.After(now)) {
			out = append(out, item)
		}
	}
	sortReadyItems(out, now)
	if limit > 0 && len(out) > limit {
		out = out[:limit]
	}
	return out, nil
}

func (s *EditorialInboxService) CountInboxStatuses() (map[string]int, error) {
	counts, err := s.repos.EditorialInbox.CountByStatus()
	if err != nil {
		return nil, fmt.Errorf("count editorial inbox statuses: %w", err)
	}
	return counts, nil
}

func (s *EditorialInboxService) GetInboxSummary(now time.Time) (*model.EditorialInboxSummary, error) {
	fairnessStreak, err := s.repos.EditorialInbox.GetPolicyState(editorialPolicyStateNonUrgentClaimStreak)
	if err != nil {
		return nil, fmt.Errorf("get fairness streak: %w", err)
	}
	autonomousGapOpened, err := s.repos.EditorialInbox.GetPolicyState(editorialPolicyStateAutonomousGapOpened)
	if err != nil {
		return nil, fmt.Errorf("get autonomous gap flag: %w", err)
	}
	overdue, err := s.repos.EditorialInbox.CountOverdue(now.UTC())
	if err != nil {
		return nil, fmt.Errorf("count overdue editorial inbox items: %w", err)
	}
	analyticsRow, err := s.repos.EditorialInbox.GetAnalytics()
	if err != nil {
		return nil, fmt.Errorf("get editorial analytics: %w", err)
	}
	byModeRows, err := s.repos.EditorialInbox.CountDoneByMode()
	if err != nil {
		return nil, fmt.Errorf("get editorial analytics by mode: %w", err)
	}
	byMode := make([]model.EditorialModeAnalytics, 0, len(byModeRows))
	for _, row := range byModeRows {
		byMode = append(byMode, model.EditorialModeAnalytics{Mode: row.Mode, Count: row.Count})
	}
	return &model.EditorialInboxSummary{
		Fairness: model.EditorialFairnessSummary{
			NonUrgentClaimStreak:   fairnessStreak,
			AutonomousGapThreshold: editorialFairnessGapThreshold,
			AutonomousGapOpened:    autonomousGapOpened > 0,
		},
		Overdue: overdue,
		Analytics: model.EditorialPublishAnalytics{
			DoneCount:               analyticsRow.DoneCount,
			FailedCount:             analyticsRow.FailedCount,
			CompletedWithPostCount:  analyticsRow.CompletedWithPostCount,
			AverageQueueWaitSeconds: analyticsRow.AverageQueueWaitSeconds,
			AverageReadyWaitSeconds: analyticsRow.AverageReadyWaitSeconds,
			ByMode:                  byMode,
		},
	}, nil
}

func (s *EditorialInboxService) buildInboxItem(id int64, sourceType, sourceValue, categoryHint string, priority int, notBefore time.Time, deadline *time.Time, note, mode, status string, publishedPostID *int64, failureNote, failureMeta string) (*model.EditorialInboxItem, error) {
	sourceType = strings.TrimSpace(sourceType)
	sourceValue = strings.TrimSpace(sourceValue)
	categoryHint = strings.TrimSpace(categoryHint)
	note = strings.TrimSpace(note)
	mode = strings.TrimSpace(mode)
	status = strings.TrimSpace(status)
	failureNote = strings.TrimSpace(failureNote)
	failureMeta = strings.TrimSpace(failureMeta)
	if sourceType == "" || sourceValue == "" {
		return nil, fmt.Errorf("%w: source_type and source_value required", ErrInvalidEditorialInboxInput)
	}
	if notBefore.IsZero() {
		return nil, fmt.Errorf("%w: not_before required", ErrInvalidEditorialInboxInput)
	}
	if deadline != nil && deadline.Before(notBefore) {
		return nil, fmt.Errorf("%w: deadline must be after not_before", ErrInvalidEditorialInboxInput)
	}
	if priority < 0 || priority > 100 {
		return nil, fmt.Errorf("%w: priority must be between 0 and 100", ErrInvalidEditorialInboxInput)
	}
	if !model.IsValidEditorialMode(mode) {
		return nil, fmt.Errorf("%w: invalid mode", ErrInvalidEditorialInboxInput)
	}
	if !model.IsValidEditorialStatus(status) {
		return nil, fmt.Errorf("%w: invalid status", ErrInvalidEditorialInboxInput)
	}
	if publishedPostID != nil && *publishedPostID <= 0 {
		return nil, fmt.Errorf("%w: published_post_id must be positive", ErrInvalidEditorialInboxInput)
	}
	if status == model.EditorialStatusDone && publishedPostID == nil {
		return nil, fmt.Errorf("%w: published_post_id required when status is done", ErrInvalidEditorialInboxInput)
	}
	if status != model.EditorialStatusDone {
		publishedPostID = nil
	}
	claimedBy := ""
	var claimedAt *time.Time
	var leaseExpiresAt *time.Time
	var completedAt *time.Time
	claimTokenHash := ""
	attemptCount := 0
	if status != model.EditorialStatusRunning {
		claimedBy = ""
		claimedAt = nil
		leaseExpiresAt = nil
		claimTokenHash = ""
	}
	if status != model.EditorialStatusFailed {
		failureNote = ""
		failureMeta = ""
	}
	if status != model.EditorialStatusDone {
		completedAt = nil
	}
	if categoryHint != "" {
		if _, err := s.repos.Category.GetBySlug(categoryHint); err != nil {
			return nil, fmt.Errorf("%w: invalid category_hint", ErrInvalidEditorialInboxInput)
		}
	}
	return &model.EditorialInboxItem{
		ID:              id,
		SourceType:      sourceType,
		SourceValue:     sourceValue,
		CategoryHint:    categoryHint,
		Priority:        priority,
		NotBefore:       notBefore.UTC(),
		Deadline:        deadline,
		Note:            note,
		Mode:            mode,
		Status:          status,
		PublishedPostID: publishedPostID,
		FailureNote:     failureNote,
		FailureMeta:     failureMeta,
		ClaimedBy:       claimedBy,
		ClaimTokenHash:  claimTokenHash,
		ClaimedAt:       claimedAt,
		LeaseExpiresAt:  leaseExpiresAt,
		AttemptCount:    attemptCount,
		CompletedAt:     completedAt,
	}, nil
}

func isEditorialOverdue(item model.EditorialInboxItem, now time.Time) bool {
	return item.Deadline != nil && item.Deadline.Before(now) && (item.Status == model.EditorialStatusApproved || item.Status == model.EditorialStatusRunning || item.Status == model.EditorialStatusDone)
}

func isStaleRunningClaim(item model.EditorialInboxItem, now time.Time) bool {
	return item.Status == model.EditorialStatusRunning && item.LeaseExpiresAt != nil && !item.LeaseExpiresAt.After(now)
}

func sortReadyItems(items []model.EditorialInboxItem, now time.Time) {
	slices.SortFunc(items, func(a, b model.EditorialInboxItem) int {
		aOverdue := isEditorialOverdue(a, now)
		bOverdue := isEditorialOverdue(b, now)
		if aOverdue != bOverdue {
			if aOverdue {
				return -1
			}
			return 1
		}
		if a.Priority != b.Priority {
			return b.Priority - a.Priority
		}
		if a.Deadline == nil && b.Deadline != nil {
			return 1
		}
		if a.Deadline != nil && b.Deadline == nil {
			return -1
		}
		if a.Deadline != nil && b.Deadline != nil {
			if a.Deadline.Before(*b.Deadline) {
				return -1
			}
			if b.Deadline.Before(*a.Deadline) {
				return 1
			}
		}
		if a.CreatedAt.Before(b.CreatedAt) {
			return -1
		}
		if b.CreatedAt.Before(a.CreatedAt) {
			return 1
		}
		return 0
	})
}

func editorialRetryBackoff(attemptCount int) time.Duration {
	switch {
	case attemptCount <= 1:
		return 15 * time.Minute
	case attemptCount == 2:
		return time.Hour
	default:
		return 6 * time.Hour
	}
}

func newEditorialClaimToken() (string, string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", "", err
	}
	token := hex.EncodeToString(b)
	return token, editorialClaimTokenHash(token), nil
}

func editorialClaimTokenHash(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
