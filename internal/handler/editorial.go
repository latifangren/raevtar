package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"raevtar/internal/model"
	"raevtar/internal/service"
	adminview "raevtar/internal/view/admin"
)

func (h *Handler) adminEditorialInbox(w http.ResponseWriter, r *http.Request) {
	items, err := h.svc.Editorial.ListInboxItems(service.EditorialInboxListFilter{})
	if err != nil {
		internalServerError(w, r, err)
		return
	}
	counts, err := h.svc.Editorial.CountInboxStatuses()
	if err != nil {
		internalServerError(w, r, err)
		return
	}
	categories, err := h.svc.Blog.ListCategories()
	if err != nil {
		internalServerError(w, r, err)
		return
	}
	renderHTML(w, r, adminview.EditorialInbox(adminview.EditorialInboxData{
		CurrentPath: r.URL.Path,
		CSRFToken:   csrfTokenForRequest(r),
		Items:       items,
		Counts:      counts,
		Categories:  categories,
		Modes:       model.ValidEditorialModes(),
		Statuses:    model.ValidEditorialStatuses(),
	}))
}

func (h *Handler) adminCreateEditorialInbox(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST required", http.StatusMethodNotAllowed)
		return
	}
	entry, _ := getSessionEntry(r)
	input, err := editorialInboxCreateFromForm(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	item, err := h.svc.Editorial.CreateInboxItem(input)
	if err != nil {
		if errors.Is(err, service.ErrInvalidEditorialInboxInput) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		internalServerError(w, r, err)
		return
	}
	warnAfterMutation(r, "create_editorial_audit", h.svc.Admin.LogEditorialInboxCreated(entry.username, item, clientIP(r)))
	http.Redirect(w, r, "/admin/editorial-inbox", http.StatusSeeOther)
}

func (h *Handler) adminUpdateEditorialInbox(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST required", http.StatusMethodNotAllowed)
		return
	}
	entry, _ := getSessionEntry(r)
	id, err := strconv.ParseInt(r.PathValue("itemID"), 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	input, err := editorialInboxUpdateFromForm(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	item, err := h.svc.Editorial.UpdateInboxItem(id, input)
	if err != nil {
		if errors.Is(err, service.ErrInvalidEditorialInboxInput) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if errors.Is(err, service.ErrEditorialInboxNotFound) {
			http.Error(w, "item not found", http.StatusNotFound)
			return
		}
		internalServerError(w, r, err)
		return
	}
	warnAfterMutation(r, "update_editorial_audit", h.svc.Admin.LogEditorialInboxUpdated(entry.username, item, clientIP(r)))
	http.Redirect(w, r, "/admin/editorial-inbox", http.StatusSeeOther)
}

func (h *Handler) apiEditorialInboxContract(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"source_of_truth": "raevtar",
		"resource":        "editorial_inbox",
		"modes":           model.ValidEditorialModes(),
		"statuses":        model.ValidEditorialStatuses(),
		"ready_selection": map[string]any{
			"status":         model.EditorialStatusApproved,
			"not_before_lte": "now",
			"order":          []string{"priority desc", "deadline asc nulls last", "created_at asc"},
		},
		"fields": []string{"source_type", "source_value", "category_hint", "priority", "not_before", "deadline", "note", "mode", "status", "published_post_id", "failure_note", "failure_meta"},
	})
}

func (h *Handler) apiListEditorialInbox(w http.ResponseWriter, r *http.Request) {
	filter := service.EditorialInboxListFilter{
		Status: strings.TrimSpace(r.URL.Query().Get("status")),
		Mode:   strings.TrimSpace(r.URL.Query().Get("mode")),
		Ready:  r.URL.Query().Get("ready") == "true",
		Limit:  100,
	}
	items, err := h.svc.Editorial.ListInboxItems(filter)
	if err != nil {
		internalServerJSON(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func (h *Handler) apiGetEditorialInbox(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("itemID"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}
	item, err := h.svc.Editorial.GetInboxItem(id)
	if err != nil {
		if errors.Is(err, service.ErrEditorialInboxNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
			return
		}
		internalServerJSON(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *Handler) apiCreateEditorialInbox(w http.ResponseWriter, r *http.Request) {
	capRequestBody(w, r, apiBodyLimit)
	var input model.EditorialInboxCreate
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		if isBodyTooLarge(err) {
			writeJSON(w, http.StatusRequestEntityTooLarge, map[string]string{"error": "request body too large"})
			return
		}
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
		return
	}
	item, err := h.svc.Editorial.CreateInboxItem(input)
	if err != nil {
		if errors.Is(err, service.ErrInvalidEditorialInboxInput) {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		internalServerJSON(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, item)
}

func (h *Handler) apiUpdateEditorialInbox(w http.ResponseWriter, r *http.Request) {
	capRequestBody(w, r, apiBodyLimit)
	id, err := strconv.ParseInt(r.PathValue("itemID"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}
	var input model.EditorialInboxUpdate
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		if isBodyTooLarge(err) {
			writeJSON(w, http.StatusRequestEntityTooLarge, map[string]string{"error": "request body too large"})
			return
		}
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
		return
	}
	item, err := h.svc.Editorial.UpdateInboxItem(id, input)
	if err != nil {
		if errors.Is(err, service.ErrInvalidEditorialInboxInput) {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		if errors.Is(err, service.ErrEditorialInboxNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
			return
		}
		internalServerJSON(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func editorialInboxCreateFromForm(r *http.Request) (model.EditorialInboxCreate, error) {
	notBefore, deadline, err := editorialTimesFromForm(r)
	if err != nil {
		return model.EditorialInboxCreate{}, err
	}
	priority, err := strconv.Atoi(strings.TrimSpace(r.FormValue("priority")))
	if err != nil {
		return model.EditorialInboxCreate{}, errors.New("priority must be a number")
	}
	return model.EditorialInboxCreate{
		SourceType:      r.FormValue("source_type"),
		SourceValue:     r.FormValue("source_value"),
		CategoryHint:    r.FormValue("category_hint"),
		Priority:        priority,
		NotBefore:       notBefore,
		Deadline:        deadline,
		Note:            r.FormValue("note"),
		Mode:            r.FormValue("mode"),
		Status:          r.FormValue("status"),
		PublishedPostID: editorialPostIDFromForm(r),
		FailureNote:     r.FormValue("failure_note"),
		FailureMeta:     r.FormValue("failure_meta"),
	}, nil
}

func editorialInboxUpdateFromForm(r *http.Request) (model.EditorialInboxUpdate, error) {
	notBefore, deadline, err := editorialTimesFromForm(r)
	if err != nil {
		return model.EditorialInboxUpdate{}, err
	}
	priority, err := strconv.Atoi(strings.TrimSpace(r.FormValue("priority")))
	if err != nil {
		return model.EditorialInboxUpdate{}, errors.New("priority must be a number")
	}
	return model.EditorialInboxUpdate{
		SourceType:      r.FormValue("source_type"),
		SourceValue:     r.FormValue("source_value"),
		CategoryHint:    r.FormValue("category_hint"),
		Priority:        priority,
		NotBefore:       notBefore,
		Deadline:        deadline,
		Note:            r.FormValue("note"),
		Mode:            r.FormValue("mode"),
		Status:          r.FormValue("status"),
		PublishedPostID: editorialPostIDFromForm(r),
		FailureNote:     r.FormValue("failure_note"),
		FailureMeta:     r.FormValue("failure_meta"),
	}, nil
}

func editorialPostIDFromForm(r *http.Request) *int64 {
	value := strings.TrimSpace(r.FormValue("published_post_id"))
	if value == "" {
		return nil
	}
	id, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return nil
	}
	return &id
}

func editorialTimesFromForm(r *http.Request) (time.Time, *time.Time, error) {
	notBeforeText := strings.TrimSpace(r.FormValue("not_before"))
	if notBeforeText == "" {
		return time.Time{}, nil, errors.New("not_before required")
	}
	notBefore, err := time.Parse("2006-01-02T15:04", notBeforeText)
	if err != nil {
		return time.Time{}, nil, errors.New("invalid not_before")
	}
	deadlineText := strings.TrimSpace(r.FormValue("deadline"))
	if deadlineText == "" {
		return notBefore, nil, nil
	}
	deadline, err := time.Parse("2006-01-02T15:04", deadlineText)
	if err != nil {
		return time.Time{}, nil, errors.New("invalid deadline")
	}
	return notBefore, &deadline, nil
}
