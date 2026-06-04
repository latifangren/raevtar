package service

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"raevtar/internal/model"
	"raevtar/internal/repo"
)

var ErrInvalidEditorialInboxInput = errors.New("invalid editorial inbox input")
var ErrEditorialInboxNotFound = errors.New("editorial inbox item not found")

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
	item, err := s.buildInboxItem(0, input.SourceType, input.SourceValue, input.CategoryHint, input.Priority, input.NotBefore, input.Deadline, input.Note, input.Mode, input.Status)
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
	_, err := s.repos.EditorialInbox.GetByID(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w: %w", ErrEditorialInboxNotFound, err)
		}
		return nil, fmt.Errorf("get editorial inbox item: %w", err)
	}
	item, err := s.buildInboxItem(id, input.SourceType, input.SourceValue, input.CategoryHint, input.Priority, input.NotBefore, input.Deadline, input.Note, input.Mode, input.Status)
	if err != nil {
		return nil, err
	}
	item.UpdatedAt = time.Now().UTC()
	if err := s.repos.EditorialInbox.Update(item); err != nil {
		return nil, fmt.Errorf("update editorial inbox item: %w", err)
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
	return items, nil
}

func (s *EditorialInboxService) CountInboxStatuses() (map[string]int, error) {
	counts, err := s.repos.EditorialInbox.CountByStatus()
	if err != nil {
		return nil, fmt.Errorf("count editorial inbox statuses: %w", err)
	}
	return counts, nil
}

func (s *EditorialInboxService) buildInboxItem(id int64, sourceType, sourceValue, categoryHint string, priority int, notBefore time.Time, deadline *time.Time, note, mode, status string) (*model.EditorialInboxItem, error) {
	sourceType = strings.TrimSpace(sourceType)
	sourceValue = strings.TrimSpace(sourceValue)
	categoryHint = strings.TrimSpace(categoryHint)
	note = strings.TrimSpace(note)
	mode = strings.TrimSpace(mode)
	status = strings.TrimSpace(status)
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
	if categoryHint != "" {
		if _, err := s.repos.Category.GetBySlug(categoryHint); err != nil {
			return nil, fmt.Errorf("%w: invalid category_hint", ErrInvalidEditorialInboxInput)
		}
	}
	return &model.EditorialInboxItem{
		ID:           id,
		SourceType:   sourceType,
		SourceValue:  sourceValue,
		CategoryHint: categoryHint,
		Priority:     priority,
		NotBefore:    notBefore.UTC(),
		Deadline:     deadline,
		Note:         note,
		Mode:         mode,
		Status:       status,
	}, nil
}
