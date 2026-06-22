package service

import (
	"errors"
	"strconv"
	"strings"
	"testing"
	"time"

	"raevtar/internal/model"
)

// ── LogPostCreated / DeletePost ─────────────────────────────────────────────

func TestAdminServiceLogPostCreated(t *testing.T) {
	state := newTestServices(t)

	if err := state.svc.Admin.LogPostCreated("admin", "Test Article", "10.0.0.1"); err != nil {
		t.Fatalf("LogPostCreated: %v", err)
	}

	logs, err := state.svc.Admin.ListAuditLogs(10, 0)
	if err != nil {
		t.Fatalf("ListAuditLogs: %v", err)
	}

	var found bool
	for _, log := range logs {
		if log.Action == "CREATE_POST" && log.User == "admin" {
			found = true
			if !strings.Contains(log.Details, "Test Article") {
				t.Fatalf("CREATE_POST details = %q, want containing title", log.Details)
			}
			if log.IP != "10.0.0.1" {
				t.Fatalf("CREATE_POST IP = %q, want 10.0.0.1", log.IP)
			}
			break
		}
	}
	if !found {
		t.Fatalf("CREATE_POST audit entry not found")
	}
}

func TestAdminServiceDeletePost(t *testing.T) {
	state := newTestServices(t)

	post, err := state.svc.Blog.CreatePost(model.PostCreate{
		CategorySlug: "devops",
		Title:        "Post To Delete",
		ContentMD:    "# Delete Me",
		Published:    true,
	})
	if err != nil {
		t.Fatalf("create post: %v", err)
	}

	if err := state.svc.Admin.DeletePost("admin", post.ID, "10.0.0.1"); err != nil {
		t.Fatalf("DeletePost: %v", err)
	}

	// Post should be deleted
	_, err = state.svc.Blog.GetPostByID(post.ID)
	if err == nil {
		t.Fatalf("expected deleted post lookup to fail")
	}

	// Audit log should have DELETE_POST entry
	logs, err := state.svc.Admin.ListAuditLogs(10, 0)
	if err != nil {
		t.Fatalf("ListAuditLogs: %v", err)
	}

	var found bool
	for _, log := range logs {
		if log.Action == "DELETE_POST" && log.User == "admin" {
			found = true
			if !strings.Contains(log.Details, "Post To Delete") {
				t.Fatalf("DELETE_POST details = %q, want containing title", log.Details)
			}
			if log.IP != "10.0.0.1" {
				t.Fatalf("DELETE_POST IP = %q, want 10.0.0.1", log.IP)
			}
			break
		}
	}
	if !found {
		t.Fatalf("DELETE_POST audit entry not found")
	}
}

func TestAdminServiceDeletePostNotFound(t *testing.T) {
	state := newTestServices(t)

	err := state.svc.Admin.DeletePost("admin", 9999, "10.0.0.1")
	if err == nil {
		t.Fatalf("expected error for non-existent post")
	}
	if !errors.Is(err, ErrPostNotFound) {
		t.Fatalf("err = %v, want ErrPostNotFound", err)
	}
}

// ── LogServerCreated / LogServerUpdated ─────────────────────────────────────

func TestAdminServiceLogServerCreated(t *testing.T) {
	state := newTestServices(t)

	if err := state.svc.Admin.LogServerCreated("admin", "web-01", "10.0.0.1", "9100", "10.0.0.1"); err != nil {
		t.Fatalf("LogServerCreated: %v", err)
	}

	logs, err := state.svc.Admin.ListAuditLogs(10, 0)
	if err != nil {
		t.Fatalf("ListAuditLogs: %v", err)
	}

	var found bool
	for _, log := range logs {
		if log.Action == "CREATE_SERVER" && log.User == "admin" {
			found = true
			if !strings.Contains(log.Details, "web-01") {
				t.Fatalf("CREATE_SERVER details = %q, want containing server name", log.Details)
			}
			break
		}
	}
	if !found {
		t.Fatalf("CREATE_SERVER audit entry not found")
	}
}

func TestAdminServiceLogServerUpdated(t *testing.T) {
	state := newTestServices(t)

	if err := state.svc.Admin.LogServerUpdated("admin", 1, "web-01", "10.0.0.1", "9100", "10.0.0.1"); err != nil {
		t.Fatalf("LogServerUpdated: %v", err)
	}

	logs, err := state.svc.Admin.ListAuditLogs(10, 0)
	if err != nil {
		t.Fatalf("ListAuditLogs: %v", err)
	}

	var found bool
	for _, log := range logs {
		if log.Action == "UPDATE_SERVER" && log.User == "admin" {
			found = true
			if !strings.Contains(log.Details, "web-01") {
				t.Fatalf("UPDATE_SERVER details = %q, want containing server name", log.Details)
			}
			break
		}
	}
	if !found {
		t.Fatalf("UPDATE_SERVER audit entry not found")
	}
}

// ── LogCategoryCreated ──────────────────────────────────────────────────────

func TestAdminServiceLogCategoryCreated(t *testing.T) {
	state := newTestServices(t)

	if err := state.svc.Admin.LogCategoryCreated("admin", "test-category", "10.0.0.1"); err != nil {
		t.Fatalf("LogCategoryCreated: %v", err)
	}

	logs, err := state.svc.Admin.ListAuditLogs(10, 0)
	if err != nil {
		t.Fatalf("ListAuditLogs: %v", err)
	}

	var found bool
	for _, log := range logs {
		if log.Action == "CREATE_CATEGORY" && log.User == "admin" {
			found = true
			if !strings.Contains(log.Details, "test-category") {
				t.Fatalf("CREATE_CATEGORY details = %q, want containing category name", log.Details)
			}
			break
		}
	}
	if !found {
		t.Fatalf("CREATE_CATEGORY audit entry not found")
	}
}

// ── LogPageUpdated ──────────────────────────────────────────────────────────

func TestAdminServiceLogPageUpdated(t *testing.T) {
	state := newTestServices(t)

	page, err := state.svc.Pages.GetPage(model.PageKeyAbout)
	if err != nil {
		t.Fatalf("GetPage: %v", err)
	}

	if err := state.svc.Admin.LogPageUpdated("admin", page, "10.0.0.1"); err != nil {
		t.Fatalf("LogPageUpdated: %v", err)
	}

	logs, err := state.svc.Admin.ListAuditLogs(10, 0)
	if err != nil {
		t.Fatalf("ListAuditLogs: %v", err)
	}

	var found bool
	for _, log := range logs {
		if log.Action == "UPDATE_PAGE_CONTENT" && log.User == "admin" {
			found = true
			if !strings.Contains(log.Details, model.PageKeyAbout) {
				t.Fatalf("UPDATE_PAGE_CONTENT details = %q, want containing page key", log.Details)
			}
			break
		}
	}
	if !found {
		t.Fatalf("UPDATE_PAGE_CONTENT audit entry not found")
	}
}

// ── LogAudit ────────────────────────────────────────────────────────────────

func TestAdminServiceLogAudit(t *testing.T) {
	state := newTestServices(t)

	if err := state.svc.Admin.LogAudit("operator", "CUSTOM_ACTION", "custom details here", "10.0.0.1"); err != nil {
		t.Fatalf("LogAudit: %v", err)
	}

	logs, err := state.svc.Admin.ListAuditLogs(10, 0)
	if err != nil {
		t.Fatalf("ListAuditLogs: %v", err)
	}

	var found bool
	for _, log := range logs {
		if log.Action == "CUSTOM_ACTION" && log.User == "operator" {
			found = true
			if log.Details != "custom details here" {
				t.Fatalf("LogAudit details = %q, want custom details here", log.Details)
			}
			if log.IP != "10.0.0.1" {
				t.Fatalf("LogAudit IP = %q, want 10.0.0.1", log.IP)
			}
			break
		}
	}
	if !found {
		t.Fatalf("CUSTOM_ACTION audit entry not found")
	}
}

// ── DeleteProject ───────────────────────────────────────────────────────────

func TestAdminServiceDeleteProject(t *testing.T) {
	state := newTestServices(t)

	project, err := state.svc.Projects.CreateProject(model.ProjectCreate{
		Title:     "Admin Delete Project",
		ContentMD: "# Admin Delete",
		Excerpt:   "To be deleted via admin",
		Published: true,
	})
	if err != nil {
		t.Fatalf("create project: %v", err)
	}

	if err := state.svc.Admin.DeleteProject("admin", project.ID, "10.0.0.1"); err != nil {
		t.Fatalf("DeleteProject: %v", err)
	}

	// Verify deleted
	_, err = state.svc.Projects.GetProjectByID(project.ID)
	if err == nil {
		t.Fatalf("expected deleted project lookup to fail")
	}

	// Verify audit log
	logs, err := state.svc.Admin.ListAuditLogs(10, 0)
	if err != nil {
		t.Fatalf("ListAuditLogs: %v", err)
	}

	var found bool
	for _, log := range logs {
		if log.Action == "DELETE_PROJECT" && log.User == "admin" {
			found = true
			if !strings.Contains(log.Details, "Admin Delete Project") {
				t.Fatalf("DELETE_PROJECT details = %q, want containing project title", log.Details)
			}
			break
		}
	}
	if !found {
		t.Fatalf("DELETE_PROJECT audit entry not found")
	}
}

func TestAdminServiceDeleteProjectNotFound(t *testing.T) {
	state := newTestServices(t)

	err := state.svc.Admin.DeleteProject("admin", 9999, "10.0.0.1")
	if err == nil {
		t.Fatalf("expected error for non-existent project")
	}
	if !errors.Is(err, ErrProjectNotFound) {
		t.Fatalf("err = %v, want ErrProjectNotFound", err)
	}
}

// ── DeleteServer ────────────────────────────────────────────────────────────

func TestAdminServiceDeleteServer(t *testing.T) {
	state := newTestServices(t)

	server, err := state.svc.Monitor.CreateServer("admin-del-srv", "10.0.0.1", 9100, "test")
	if err != nil {
		t.Fatalf("create server: %v", err)
	}

	if err := state.svc.Admin.DeleteServer("admin", server.ID, "1", "10.0.0.1"); err != nil {
		t.Fatalf("Admin DeleteServer: %v", err)
	}

	// Verify deleted
	_, err = state.svc.Monitor.GetServer(server.ID)
	if err == nil {
		t.Fatalf("expected deleted server lookup to fail")
	}

	// Verify audit log
	logs, err := state.svc.Admin.ListAuditLogs(10, 0)
	if err != nil {
		t.Fatalf("ListAuditLogs: %v", err)
	}

	var found bool
	for _, log := range logs {
		if log.Action == "DELETE_SERVER" && log.User == "admin" {
			found = true
			if !strings.Contains(log.Details, "server id: 1") {
				t.Fatalf("DELETE_SERVER details = %q, want containing server id", log.Details)
			}
			break
		}
	}
	if !found {
		t.Fatalf("DELETE_SERVER audit entry not found")
	}
}

func TestAdminServiceDeleteServerNotFound(t *testing.T) {
	state := newTestServices(t)

	err := state.svc.Admin.DeleteServer("admin", 9999, "9999", "10.0.0.1")
	if err == nil {
		t.Fatalf("expected error for non-existent server")
	}
	if !errors.Is(err, ErrServerNotFound) {
		t.Fatalf("err = %v, want ErrServerNotFound", err)
	}
}


// ── LogPostUpdated ──────────────────────────────────────────────────────────

func TestAdminServiceLogPostUpdated(t *testing.T) {
	state := newTestServices(t)

	post, err := state.svc.Blog.CreatePost(model.PostCreate{
		CategorySlug: "devops",
		Title:        "Post To Update",
		ContentMD:    "# Update Me",
		Published:    true,
	})
	if err != nil {
		t.Fatalf("create post: %v", err)
	}

	if err := state.svc.Admin.LogPostUpdated("admin", post.Title, "10.0.0.1"); err != nil {
		t.Fatalf("LogPostUpdated: %v", err)
	}

	logs, err := state.svc.Admin.ListAuditLogs(10, 0)
	if err != nil {
		t.Fatalf("ListAuditLogs: %v", err)
	}

	var found bool
	for _, log := range logs {
		if log.Action == "UPDATE_POST" && log.User == "admin" {
			found = true
			if !strings.Contains(log.Details, "Post To Update") {
				t.Fatalf("UPDATE_POST details = %q, want containing title", log.Details)
			}
			if log.IP != "10.0.0.1" {
				t.Fatalf("UPDATE_POST IP = %q, want 10.0.0.1", log.IP)
			}
			break
		}
	}
	if !found {
		t.Fatalf("UPDATE_POST audit entry not found")
	}
}

// ── LogProjectCreated ──────────────────────────────────────────────────────

func TestAdminServiceLogProjectCreated(t *testing.T) {
	state := newTestServices(t)

	if err := state.svc.Admin.LogProjectCreated("admin", "Test Project", "10.0.0.1"); err != nil {
		t.Fatalf("LogProjectCreated: %v", err)
	}

	logs, err := state.svc.Admin.ListAuditLogs(10, 0)
	if err != nil {
		t.Fatalf("ListAuditLogs: %v", err)
	}

	var found bool
	for _, log := range logs {
		if log.Action == "CREATE_PROJECT" && log.User == "admin" {
			found = true
			if !strings.Contains(log.Details, "Test Project") {
				t.Fatalf("CREATE_PROJECT details = %q, want containing title", log.Details)
			}
			break
		}
	}
	if !found {
		t.Fatalf("CREATE_PROJECT audit entry not found")
	}
}

// ── LogProjectUpdated ──────────────────────────────────────────────────────

func TestAdminServiceLogProjectUpdated(t *testing.T) {
	state := newTestServices(t)

	if err := state.svc.Admin.LogProjectUpdated("admin", "Updated Project", "10.0.0.1"); err != nil {
		t.Fatalf("LogProjectUpdated: %v", err)
	}

	logs, err := state.svc.Admin.ListAuditLogs(10, 0)
	if err != nil {
		t.Fatalf("ListAuditLogs: %v", err)
	}

	var found bool
	for _, log := range logs {
		if log.Action == "UPDATE_PROJECT" && log.User == "admin" {
			found = true
			if !strings.Contains(log.Details, "Updated Project") {
				t.Fatalf("UPDATE_PROJECT details = %q, want containing title", log.Details)
			}
			break
		}
	}
	if !found {
		t.Fatalf("UPDATE_PROJECT audit entry not found")
	}
}

// ── LogCategoryUpdated ──────────────────────────────────────────────────────

func TestAdminServiceLogCategoryUpdated(t *testing.T) {
	state := newTestServices(t)

	if err := state.svc.Admin.LogCategoryUpdated("admin", "Updated Category", "10.0.0.1"); err != nil {
		t.Fatalf("LogCategoryUpdated: %v", err)
	}

	logs, err := state.svc.Admin.ListAuditLogs(10, 0)
	if err != nil {
		t.Fatalf("ListAuditLogs: %v", err)
	}

	var found bool
	for _, log := range logs {
		if log.Action == "UPDATE_CATEGORY" && log.User == "admin" {
			found = true
			if !strings.Contains(log.Details, "Updated Category") {
				t.Fatalf("UPDATE_CATEGORY details = %q, want containing name", log.Details)
			}
			break
		}
	}
	if !found {
		t.Fatalf("UPDATE_CATEGORY audit entry not found")
	}
}

// ── Admin DeleteCategory ────────────────────────────────────────────────────

func TestAdminServiceDeleteCategory(t *testing.T) {
	state := newTestServices(t)

	cat, err := callCreateCategory(t, state, model.Category{
		Slug:        "delete-by-admin",
		Name:        "Delete By Admin",
		Description: "Will be deleted via admin",
	})
	if err != nil {
		t.Fatalf("create category: %v", err)
	}

	if err := state.svc.Admin.DeleteCategory("admin", cat, "10.0.0.1"); err != nil {
		t.Fatalf("DeleteCategory: %v", err)
	}

	// Note: AdminService.DeleteCategory only logs, doesn't delete.
	// Category should still exist.
	_, _, err = state.svc.Blog.GetCategoryByID(cat.ID)
	if err != nil {
		t.Fatalf("category should still exist after admin log: %v", err)
	}

	logs, err := state.svc.Admin.ListAuditLogs(10, 0)
	if err != nil {
		t.Fatalf("ListAuditLogs: %v", err)
	}

	var found bool
	for _, log := range logs {
		if log.Action == "DELETE_CATEGORY" && log.User == "admin" {
			found = true
			if !strings.Contains(log.Details, "Delete By Admin") {
				t.Fatalf("DELETE_CATEGORY details = %q, want containing name", log.Details)
			}
			break
		}
	}
	if !found {
		t.Fatalf("DELETE_CATEGORY audit entry not found")
	}
}

// ── LogAgentTokenRotated ───────────────────────────────────────────────────

func TestAdminServiceLogAgentTokenRotated(t *testing.T) {
	state := newTestServices(t)

	server, err := state.svc.Monitor.CreateServer("token-srv", "10.0.0.1", 9100, "test")
	if err != nil {
		t.Fatalf("create server: %v", err)
	}

	if err := state.svc.Admin.LogAgentTokenRotated("admin", strconv.FormatInt(server.ID, 10), "10.0.0.1"); err != nil {
		t.Fatalf("LogAgentTokenRotated: %v", err)
	}

	logs, err := state.svc.Admin.ListAuditLogs(10, 0)
	if err != nil {
		t.Fatalf("ListAuditLogs: %v", err)
	}

	var found bool
	for _, log := range logs {
		if log.Action == "ROTATE_AGENT_TOKEN" && log.User == "admin" {
			found = true
			if !strings.Contains(log.Details, "server id: ") {
				t.Fatalf("ROTATE_AGENT_TOKEN details = %q, want containing server id", log.Details)
			}
			break
		}
	}
	if !found {
		t.Fatalf("ROTATE_AGENT_TOKEN audit entry not found")
	}
}

// ── LogEditorialInboxCreated ───────────────────────────────────────────────

func TestAdminServiceLogEditorialInboxCreated(t *testing.T) {
	state := newTestServices(t)
	now := time.Now().UTC().Truncate(time.Minute)

	item, err := state.svc.Editorial.CreateInboxItem(model.EditorialInboxCreate{
		SourceType:  "repo",
		SourceValue: "https://github.com/example/inbox-log",
		Priority:    50,
		NotBefore:   now,
		Mode:        model.EditorialModeScheduled,
		Status:      model.EditorialStatusApproved,
	})
	if err != nil {
		t.Fatalf("create inbox item: %v", err)
	}

	if err := state.svc.Admin.LogEditorialInboxCreated("admin", item, "10.0.0.1"); err != nil {
		t.Fatalf("LogEditorialInboxCreated: %v", err)
	}

	logs, err := state.svc.Admin.ListAuditLogs(10, 0)
	if err != nil {
		t.Fatalf("ListAuditLogs: %v", err)
	}

	var found bool
	for _, log := range logs {
		if log.Action == "CREATE_EDITORIAL_ITEM" && log.User == "admin" {
			found = true
			if !strings.Contains(log.Details, item.SourceValue) {
				t.Fatalf("CREATE_EDITORIAL_ITEM details = %q, want containing source value", log.Details)
			}
			break
		}
	}
	if !found {
		t.Fatalf("CREATE_EDITORIAL_ITEM audit entry not found")
	}
}

// ── LogEditorialInboxUpdated ───────────────────────────────────────────────

func TestAdminServiceLogEditorialInboxUpdated(t *testing.T) {
	state := newTestServices(t)
	now := time.Now().UTC().Truncate(time.Minute)

	item, err := state.svc.Editorial.CreateInboxItem(model.EditorialInboxCreate{
		SourceType:  "repo",
		SourceValue: "https://github.com/example/inbox-update",
		Priority:    50,
		NotBefore:   now,
		Mode:        model.EditorialModeScheduled,
		Status:      model.EditorialStatusApproved,
	})
	if err != nil {
		t.Fatalf("create inbox item: %v", err)
	}

	if err := state.svc.Admin.LogEditorialInboxUpdated("admin", item, "10.0.0.1"); err != nil {
		t.Fatalf("LogEditorialInboxUpdated: %v", err)
	}

	logs, err := state.svc.Admin.ListAuditLogs(10, 0)
	if err != nil {
		t.Fatalf("ListAuditLogs: %v", err)
	}

	var found bool
	for _, log := range logs {
		if log.Action == "UPDATE_EDITORIAL_ITEM" && log.User == "admin" {
			found = true
			if !strings.Contains(log.Details, "status="+item.Status) {
				t.Fatalf("UPDATE_EDITORIAL_ITEM details = %q, want containing status", log.Details)
			}
			if !strings.Contains(log.Details, item.Status) {
				t.Fatalf("UPDATE_EDITORIAL_ITEM details = %q, want containing status", log.Details)
			}
			break
		}
	}
	if !found {
		t.Fatalf("UPDATE_EDITORIAL_ITEM audit entry not found")
	}
}

// ── LogEditorialInboxDeleted ───────────────────────────────────────────────

func TestAdminServiceLogEditorialInboxDeleted(t *testing.T) {
	state := newTestServices(t)
	now := time.Now().UTC().Truncate(time.Minute)

	item, err := state.svc.Editorial.CreateInboxItem(model.EditorialInboxCreate{
		SourceType:  "repo",
		SourceValue: "https://github.com/example/inbox-delete",
		Priority:    50,
		NotBefore:   now,
		Mode:        model.EditorialModeScheduled,
		Status:      model.EditorialStatusApproved,
	})
	if err != nil {
		t.Fatalf("create inbox item: %v", err)
	}

	if err := state.svc.Admin.LogEditorialInboxDeleted("admin", item, "10.0.0.1"); err != nil {
		t.Fatalf("LogEditorialInboxDeleted: %v", err)
	}

	logs, err := state.svc.Admin.ListAuditLogs(10, 0)
	if err != nil {
		t.Fatalf("ListAuditLogs: %v", err)
	}

	var found bool
	for _, log := range logs {
		if log.Action == "DELETE_EDITORIAL_ITEM" && log.User == "admin" {
			found = true
			if !strings.Contains(log.Details, item.SourceValue) {
				t.Fatalf("DELETE_EDITORIAL_ITEM details = %q, want containing source value", log.Details)
			}
			break
		}
	}
	if !found {
		t.Fatalf("DELETE_EDITORIAL_ITEM audit entry not found")
	}
}






