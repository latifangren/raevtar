package handler

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"raevtar/internal/model"
	"raevtar/internal/service"
	"strconv"
	"strings"
	"testing"
)

// =============================================================================
// Admin page rendering (GET)
// =============================================================================

// TestAdminDashboard verifies the admin dashboard page renders correctly.
func TestAdminDashboard(t *testing.T) {
	app := newPublicTestApp(t)
	sessionCookie, _ := loginAsAdmin(t, app)

	status, body := getBody(t, app, "/admin", sessionCookie)
	if status != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", status, http.StatusOK, body)
	}
	assertContains(t, body, "Dashboard")
	assertContains(t, body, "Online Servers")
	assertContains(t, body, "Quick Actions")
	assertContains(t, body, "Server Status")
	assertContains(t, body, "System Health")
}

// TestAdminUsersPage verifies the manage users list page renders.
func TestAdminUsersPage(t *testing.T) {
	app := newPublicTestApp(t)
	sessionCookie, _ := loginAsAdmin(t, app)

	status, body := getBody(t, app, "/admin/manage-users", sessionCookie)
	if status != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", status, http.StatusOK, body)
	}
	assertContains(t, body, "Manage Users")
	assertContains(t, body, "Username")
	assertContains(t, body, "Role")
	assertContains(t, body, "admin")
}

// TestAdminUserCreatePage verifies the user creation form is present on the
// manage users page.
func TestAdminUserCreatePage(t *testing.T) {
	app := newPublicTestApp(t)
	sessionCookie, _ := loginAsAdmin(t, app)

	status, body := getBody(t, app, "/admin/manage-users", sessionCookie)
	if status != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", status, http.StatusOK, body)
	}
	assertContains(t, body, "Add New User")
	assertContains(t, body, `name="username"`)
	assertContains(t, body, `name="password"`)
	assertContains(t, body, `name="role"`)
	assertContains(t, body, `action="/admin/manage-users"`)
}

// TestAdminAuditLogPage verifies the audit log page renders with the expected
// table structure.
func TestAdminAuditLogPage(t *testing.T) {
	app := newPublicTestApp(t)
	sessionCookie, _ := loginAsAdmin(t, app)

	status, body := getBody(t, app, "/admin/audit-log", sessionCookie)
	if status != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", status, http.StatusOK, body)
	}
	assertContains(t, body, "Audit Log")
	assertContains(t, body, "Action")
	assertContains(t, body, "User")
	assertContains(t, body, "IP")
}

// TestAdminCategoriesPage verifies the topics (categories) page renders with
// the seed category.
func TestAdminCategoriesPage(t *testing.T) {
	app := newPublicTestApp(t)
	sessionCookie, _ := loginAsAdmin(t, app)

	status, body := getBody(t, app, "/admin/topics", sessionCookie)
	if status != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", status, http.StatusOK, body)
	}
	assertContains(t, body, "Manage Topics")
	assertContains(t, body, "DevOps")
	assertContains(t, body, "Name")
	assertContains(t, body, "Slug")
	assertContains(t, body, "Create New Topic")
}

// TestAdminMediaPage verifies the media library page renders with the upload
// form.
func TestAdminMediaPage(t *testing.T) {
	app := newPublicTestApp(t)
	sessionCookie, _ := loginAsAdmin(t, app)

	status, body := getBody(t, app, "/admin/media", sessionCookie)
	if status != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", status, http.StatusOK, body)
	}
	assertContains(t, body, "Media Library")
	assertContains(t, body, "Upload image")
	assertContains(t, body, `name="file"`)
	assertContains(t, body, `action="/admin/media"`)
}

// TestAdminPostsPage verifies the manage posts page renders with the seed post
// and the create form.
func TestAdminPostsPage(t *testing.T) {
	app := newPublicTestApp(t)
	sessionCookie, _ := loginAsAdmin(t, app)

	status, body := getBody(t, app, "/admin/posts", sessionCookie)
	if status != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", status, http.StatusOK, body)
	}
	assertContains(t, body, "Manage Posts")
	assertContains(t, body, "Hello Raevtar")
	assertContains(t, body, "devops")
	assertContains(t, body, "Views")
	assertContains(t, body, "published")
}

// TestAdminPostCreatePage verifies the create post form is present on the posts
// page with the required fields and the seed category in the dropdown.
func TestAdminPostCreatePage(t *testing.T) {
	app := newPublicTestApp(t)
	sessionCookie, _ := loginAsAdmin(t, app)

	status, body := getBody(t, app, "/admin/posts", sessionCookie)
	if status != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", status, http.StatusOK, body)
	}
	assertContains(t, body, "Create New Post")
	assertContains(t, body, `name="title"`)
	assertContains(t, body, `name="category_slug"`)
	assertContains(t, body, `name="content"`)
	assertContains(t, body, `name="intent" value="draft"`)
	assertContains(t, body, `name="intent" value="publish"`)
	assertContains(t, body, "DevOps")
}

// TestAdminProjectsPage verifies the manage projects page renders with the seed
// project.
func TestAdminProjectsPage(t *testing.T) {
	app := newPublicTestApp(t)
	sessionCookie, _ := loginAsAdmin(t, app)

	status, body := getBody(t, app, "/admin/projects", sessionCookie)
	if status != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", status, http.StatusOK, body)
	}
	assertContains(t, body, "Manage Projects")
	assertContains(t, body, "Whyred Watchtower")
	assertContains(t, body, "Status")
	assertContains(t, body, "Order")
}

// TestAdminProjectCreatePage verifies the create project form is present on the
// projects page.
func TestAdminProjectCreatePage(t *testing.T) {
	app := newPublicTestApp(t)
	sessionCookie, _ := loginAsAdmin(t, app)

	status, body := getBody(t, app, "/admin/projects", sessionCookie)
	if status != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", status, http.StatusOK, body)
	}
	assertContains(t, body, "Create New Project")
	assertContains(t, body, `name="title"`)
	assertContains(t, body, `name="content"`)
	assertContains(t, body, `name="excerpt"`)
	assertContains(t, body, `name="tags"`)
	assertContains(t, body, `name="featured"`)
}

// TestAdminPagesPage verifies the pages list page renders with the seed pages.
func TestAdminPagesPage(t *testing.T) {
	app := newPublicTestApp(t)
	sessionCookie, _ := loginAsAdmin(t, app)

	status, body := getBody(t, app, "/admin/pages", sessionCookie)
	if status != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", status, http.StatusOK, body)
	}
	assertContains(t, body, "Manage Pages")
	assertContains(t, body, "about")
	assertContains(t, body, "contact")
	assertContains(t, body, "Hero support copy")
}

// TestAdminWebhooksPage verifies the webhooks page renders with the create form
// and threshold info.
func TestAdminWebhooksPage(t *testing.T) {
	app := newPublicTestApp(t)
	sessionCookie, _ := loginAsAdmin(t, app)

	status, body := getBody(t, app, "/admin/webhooks", sessionCookie)
	if status != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", status, http.StatusOK, body)
	}
	assertContains(t, body, "Webhook Alerts")
	assertContains(t, body, "Create Webhook")
	assertContains(t, body, `name="name"`)
	assertContains(t, body, `name="url"`)
	assertContains(t, body, "CPU Usage")
	assertContains(t, body, "RAM Usage")
	assertContains(t, body, "Disk Usage")
}

// TestAdminServerCommandsPage verifies the server detail page shows the remote
// command center with available commands.
func TestAdminServerCommandsPage(t *testing.T) {
	app := newPublicTestApp(t)
	sessionCookie, _ := loginAsAdmin(t, app)

	path := "/admin/servers/" + strconv.FormatInt(app.serverID, 10)
	status, body := getBody(t, app, path, sessionCookie)
	if status != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", status, http.StatusOK, body)
	}
	assertContains(t, body, "Remote Command Center")
	assertContains(t, body, "Restart Agent")
	assertContains(t, body, "Clear Cache")
	assertContains(t, body, "Reboot Node")
	assertContains(t, body, "Update Agent")
	assertContains(t, body, "whyred")
	assertContains(t, body, "Node Metadata")
}

// =============================================================================
// Admin form submission (POST)
// =============================================================================

// TestAdminCreateUser verifies that POSTing valid user creation form data
// returns a 303 redirect to the manage users page.
func TestAdminCreateUser(t *testing.T) {
	app := newPublicTestApp(t)
	sessionCookie, csrfToken := loginAsAdmin(t, app)

	form := url.Values{
		"username": {"createduser"},
		"password": {"strong-password"},
		"role":     {"operator"},
		"_csrf":    {csrfToken},
	}
	req := httptest.NewRequest(http.MethodPost, "/admin/manage-users", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(sessionCookie)
	testRequestCounter++
	req.RemoteAddr = fmt.Sprintf("203.0.113.%d:1234", testRequestCounter%250+1)
	rr := httptest.NewRecorder()
	app.handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusSeeOther, rr.Body.String())
	}
	if got := rr.Header().Get("Location"); got != "/admin/manage-users" {
		t.Fatalf("Location = %q, want /admin/manage-users", got)
	}

	// Verify the user was actually created
	users, err := app.svc.Admin.ListUsers()
	if err != nil {
		t.Fatalf("list users: %v", err)
	}
	var found bool
	for _, u := range users {
		if u.Username == "createduser" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("created user not found in listing")
	}
}

// TestAdminDeleteUser verifies that POSTing to delete an existing user returns
// a 303 redirect.
func TestAdminDeleteUser(t *testing.T) {
	app := newPublicTestApp(t)
	sessionCookie, csrfToken := loginAsAdmin(t, app)

	// Create a user to delete
	user, err := app.svc.Admin.CreateUser(model.RoleOwner, "admin", "delete-target", "target-password123", model.RoleOperator, "127.0.0.1")
	if err != nil {
		t.Fatalf("create target user: %v", err)
	}

	form := url.Values{"_csrf": {csrfToken}}
	req := httptest.NewRequest(http.MethodPost, "/admin/manage-users/delete/"+strconv.FormatInt(user.ID, 10), strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(sessionCookie)
	testRequestCounter++
	req.RemoteAddr = fmt.Sprintf("203.0.113.%d:1234", testRequestCounter%250+1)
	rr := httptest.NewRecorder()
	app.handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusSeeOther, rr.Body.String())
	}
	if got := rr.Header().Get("Location"); got != "/admin/manage-users" {
		t.Fatalf("Location = %q, want /admin/manage-users", got)
	}
}

// TestAdminCreateCategory verifies that POSTing valid topic creation data
// returns a 303 redirect to the topics page.
func TestAdminCreateCategory(t *testing.T) {
	app := newPublicTestApp(t)
	sessionCookie, csrfToken := loginAsAdmin(t, app)

	form := url.Values{
		"slug":        {"test-category"},
		"name":        {"Test Category"},
		"description": {"A test category created in admin."},
		"_csrf":       {csrfToken},
	}
	req := httptest.NewRequest(http.MethodPost, "/admin/topics", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(sessionCookie)
	testRequestCounter++
	req.RemoteAddr = fmt.Sprintf("203.0.113.%d:1234", testRequestCounter%250+1)
	rr := httptest.NewRecorder()
	app.handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusSeeOther, rr.Body.String())
	}
	if got := rr.Header().Get("Location"); got != "/admin/topics" {
		t.Fatalf("Location = %q, want /admin/topics", got)
	}

	// Verify the topic was actually created
	cats, err := app.svc.Blog.ListCategories()
	if err != nil {
		t.Fatalf("list categories: %v", err)
	}
	var foundCat *model.Category
	for i := range cats {
		if cats[i].Slug == "test-category" {
			foundCat = &cats[i]
			break
		}
	}
	if foundCat == nil {
		t.Fatalf("created category not found by slug")
	}
	if foundCat.Name != "Test Category" {
		t.Fatalf("category name = %q, want %q", foundCat.Name, "Test Category")
	}
}

// TestAdminCreateMediaUpload verifies that POSTing a valid image file to the
// media endpoint returns a 303 redirect.
func TestAdminCreateMediaUpload(t *testing.T) {
	app := newPublicTestApp(t)
	sessionCookie, csrfToken := loginAsAdmin(t, app)

	req := mediaUploadRequest(t, "/admin/media", sessionCookie.Value, csrfToken, "test.png", testPNG(t))
	testRequestCounter++
	req.RemoteAddr = fmt.Sprintf("203.0.113.%d:1234", testRequestCounter%250+1)
	rr := httptest.NewRecorder()
	app.handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusSeeOther, rr.Body.String())
	}
	if got := rr.Header().Get("Location"); got != "/admin/media" {
		t.Fatalf("Location = %q, want /admin/media", got)
	}

	// Verify the asset was created
	assets, err := app.svc.Media.ListAssets()
	if err != nil {
		t.Fatalf("list assets: %v", err)
	}
	if len(assets) == 0 {
		t.Fatalf("expected at least one media asset after upload")
	}
	found := false
	for _, a := range assets {
		if a.OriginalName == "test.png" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("uploaded test.png not found in assets: %+v", assets)
	}
}

// TestAdminCreatePost verifies that POSTing valid post creation data returns a
// 303 redirect to the posts page.
func TestAdminCreatePost(t *testing.T) {
	app := newPublicTestApp(t)
	sessionCookie, csrfToken := loginAsAdmin(t, app)

	form := url.Values{
		"title":         {"Admin Created Post"},
		"category_slug": {"devops"},
		"content":       {"# Admin Created Post\n\nThis was created via admin POST."},
		"excerpt":       {"Created via admin POST test."},
		"intent":        {"draft"},
		"_csrf":         {csrfToken},
	}
	req := httptest.NewRequest(http.MethodPost, "/admin/posts", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(sessionCookie)
	testRequestCounter++
	req.RemoteAddr = fmt.Sprintf("203.0.113.%d:1234", testRequestCounter%250+1)
	rr := httptest.NewRecorder()
	app.handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusSeeOther, rr.Body.String())
	}
	if got := rr.Header().Get("Location"); got != "/admin/posts" {
		t.Fatalf("Location = %q, want /admin/posts", got)
	}

	// Verify the post was created
	posts, _, err := app.svc.Blog.ListAllPosts(1, 9999)
	if err != nil {
		t.Fatalf("list posts: %v", err)
	}
	var found bool
	for _, p := range posts {
		if p.Title == "Admin Created Post" {
			found = true
			if p.Published {
				t.Fatalf("post marked as published, want draft")
			}
			break
		}
	}
	if !found {
		t.Fatalf("created post not found in listing")
	}
}

// TestAdminUpdatePost verifies that POSTing valid update data to an existing
// post returns a 303 redirect.
func TestAdminUpdatePost(t *testing.T) {
	app := newPublicTestApp(t)
	sessionCookie, csrfToken := loginAsAdmin(t, app)

	// Find the seed post by title
	posts, _, err := app.svc.Blog.ListAllPosts(1, 9999)
	if err != nil {
		t.Fatalf("list posts: %v", err)
	}
	var postID int64
	for _, p := range posts {
		if p.Title == "Hello Raevtar" {
			postID = p.ID
			break
		}
	}
	if postID == 0 {
		t.Fatalf("seed post not found")
	}

	form := url.Values{
		"title":         {"Hello Raevtar (Updated)"},
		"category_slug": {"devops"},
		"content":       {"# Hello Raevtar\n\nUpdated via admin POST."},
		"excerpt":       {"Updated excerpt."},
		"intent":        {"update"},
		"_csrf":         {csrfToken},
	}
	req := httptest.NewRequest(http.MethodPost, "/admin/posts/update/"+strconv.FormatInt(postID, 10), strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(sessionCookie)
	testRequestCounter++
	req.RemoteAddr = fmt.Sprintf("203.0.113.%d:1234", testRequestCounter%250+1)
	rr := httptest.NewRecorder()
	app.handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusSeeOther, rr.Body.String())
	}
	if got := rr.Header().Get("Location"); got != "/admin/posts" {
		t.Fatalf("Location = %q, want /admin/posts", got)
	}

	// Verify the post was updated
	updated, err := app.svc.Blog.GetPostByID(postID)
	if err != nil {
		t.Fatalf("get post: %v", err)
	}
	if updated.Title != "Hello Raevtar (Updated)" {
		t.Fatalf("post title = %q, want %q", updated.Title, "Hello Raevtar (Updated)")
	}
}

// TestAdminCreateProject verifies that POSTing valid project creation data
// returns a 303 redirect.
func TestAdminCreateProject(t *testing.T) {
	app := newPublicTestApp(t)
	sessionCookie, csrfToken := loginAsAdmin(t, app)

	form := url.Values{
		"title":      {"Admin Created Project"},
		"content":    {"# Admin Created Project\n\nThis project was created via admin POST."},
		"excerpt":    {"Created via admin POST test."},
		"intent":     {"draft"},
		"state":      {"active"},
		"sort_order": {"5"},
		"_csrf":      {csrfToken},
	}
	req := httptest.NewRequest(http.MethodPost, "/admin/projects", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(sessionCookie)
	testRequestCounter++
	req.RemoteAddr = fmt.Sprintf("203.0.113.%d:1234", testRequestCounter%250+1)
	rr := httptest.NewRecorder()
	app.handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusSeeOther, rr.Body.String())
	}
	if got := rr.Header().Get("Location"); got != "/admin/projects" {
		t.Fatalf("Location = %q, want /admin/projects", got)
	}

	// Verify the project was created
	projects, _, err := app.svc.Projects.ListAllProjects(1, 9999, service.ProjectListOptions{})
	if err != nil {
		t.Fatalf("list projects: %v", err)
	}
	var found bool
	for _, p := range projects {
		if p.Title == "Admin Created Project" {
			found = true
			if p.Published {
				t.Fatalf("project marked as published, want draft")
			}
			if p.State != "active" {
				t.Fatalf("project state = %q, want %q", p.State, "active")
			}
			break
		}
	}
	if !found {
		t.Fatalf("created project not found in listing")
	}
}

// TestAdminUpdateProject verifies that POSTing valid update data to an existing
// project returns a 303 redirect.
func TestAdminUpdateProject(t *testing.T) {
	app := newPublicTestApp(t)
	sessionCookie, csrfToken := loginAsAdmin(t, app)

	// Find the seed project
	projects, _, err := app.svc.Projects.ListAllProjects(1, 9999, service.ProjectListOptions{})
	if err != nil {
		t.Fatalf("list projects: %v", err)
	}
	var projectID int64
	for _, p := range projects {
		if p.Title == "Whyred Watchtower" {
			projectID = p.ID
			break
		}
	}
	if projectID == 0 {
		t.Fatalf("seed project not found")
	}

	form := url.Values{
		"title":      {"Whyred Watchtower (Updated)"},
		"content":    {"# Whyred Watchtower\n\nUpdated via admin POST."},
		"excerpt":    {"Updated project excerpt."},
		"intent":     {"publish"},
		"state":      {"active"},
		"featured":   {"on"},
		"sort_order": {"2"},
		"_csrf":      {csrfToken},
	}
	req := httptest.NewRequest(http.MethodPost, "/admin/projects/update/"+strconv.FormatInt(projectID, 10), strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(sessionCookie)
	testRequestCounter++
	req.RemoteAddr = fmt.Sprintf("203.0.113.%d:1234", testRequestCounter%250+1)
	rr := httptest.NewRecorder()
	app.handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusSeeOther, rr.Body.String())
	}
	if got := rr.Header().Get("Location"); got != "/admin/projects" {
		t.Fatalf("Location = %q, want /admin/projects", got)
	}

	// Verify the project was updated
	updated, err := app.svc.Projects.GetProjectByID(projectID)
	if err != nil {
		t.Fatalf("get project: %v", err)
	}
	if updated.Title != "Whyred Watchtower (Updated)" {
		t.Fatalf("project title = %q, want %q", updated.Title, "Whyred Watchtower (Updated)")
	}
	if !updated.Published {
		t.Fatalf("project should be published after update")
	}
}

// TestAdminDeleteProject verifies that POSTing to delete an existing project
// returns a 303 redirect.
func TestAdminDeleteProject(t *testing.T) {
	app := newPublicTestApp(t)
	sessionCookie, csrfToken := loginAsAdmin(t, app)

	// Create a project to delete
	project, err := app.svc.Projects.CreateProject(model.ProjectCreate{
		Title:     "Delete Target Project",
		ContentMD: "# To Be Deleted",
		Excerpt:   "This project will be deleted.",
		Published: false,
	})
	if err != nil {
		t.Fatalf("create target project: %v", err)
	}

	form := url.Values{"_csrf": {csrfToken}}
	req := httptest.NewRequest(http.MethodPost, "/admin/projects/delete/"+strconv.FormatInt(project.ID, 10), strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(sessionCookie)
	testRequestCounter++
	req.RemoteAddr = fmt.Sprintf("203.0.113.%d:1234", testRequestCounter%250+1)
	rr := httptest.NewRecorder()
	app.handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusSeeOther, rr.Body.String())
	}
	if got := rr.Header().Get("Location"); got != "/admin/projects" {
		t.Fatalf("Location = %q, want /admin/projects", got)
	}

	// Verify the project was deleted
	_, err = app.svc.Projects.GetProjectByID(project.ID)
	if err == nil {
		t.Fatalf("project should have been deleted")
	}
}

// TestAdminCreateWebhook verifies that POSTing valid webhook creation data
// returns a 303 redirect.
func TestAdminCreateWebhook(t *testing.T) {
	app := newPublicTestApp(t)
	sessionCookie, csrfToken := loginAsAdmin(t, app)

	form := url.Values{
		"name":    {"Test Webhook"},
		"url":     {"https://hooks.example.com/alerts"},
		"secret":  {"shared-secret"},
		"enabled": {"1"},
		"_csrf":   {csrfToken},
	}
	req := httptest.NewRequest(http.MethodPost, "/admin/webhooks", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(sessionCookie)
	testRequestCounter++
	req.RemoteAddr = fmt.Sprintf("203.0.113.%d:1234", testRequestCounter%250+1)
	rr := httptest.NewRecorder()
	app.handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusSeeOther, rr.Body.String())
	}
	if got := rr.Header().Get("Location"); got != "/admin/webhooks" {
		t.Fatalf("Location = %q, want /admin/webhooks", got)
	}

	// Verify the webhook was created
	cfgs, err := app.svc.Webhook.ListConfigs()
	if err != nil {
		t.Fatalf("list webhooks: %v", err)
	}
	var found bool
	for _, wh := range cfgs {
		if wh.Name == "Test Webhook" && wh.URL == "https://hooks.example.com/alerts" {
			found = true
			if !wh.Enabled {
				t.Fatalf("webhook should be enabled")
			}
			break
		}
	}
	if !found {
		t.Fatalf("created webhook not found in listing")
	}
}

// TestAdminUpdateWebhook verifies that POSTing valid update data to an existing
// webhook returns a 303 redirect.
func TestAdminUpdateWebhook(t *testing.T) {
	app := newPublicTestApp(t)
	sessionCookie, csrfToken := loginAsAdmin(t, app)

	// Create a webhook to update
	wh, err := app.svc.Webhook.CreateConfig("Original Webhook", "https://original.example.com/hook", "orig-secret", true)
	if err != nil {
		t.Fatalf("create webhook: %v", err)
	}

	form := url.Values{
		"name":    {"Updated Webhook"},
		"url":     {"https://updated.example.com/hook"},
		"secret":  {"updated-secret"},
		"enabled": {"1"},
		"_csrf":   {csrfToken},
	}
	req := httptest.NewRequest(http.MethodPost, "/admin/webhooks/update/"+strconv.FormatInt(wh.ID, 10), strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(sessionCookie)
	testRequestCounter++
	req.RemoteAddr = fmt.Sprintf("203.0.113.%d:1234", testRequestCounter%250+1)
	rr := httptest.NewRecorder()
	app.handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusSeeOther, rr.Body.String())
	}
	if got := rr.Header().Get("Location"); got != "/admin/webhooks" {
		t.Fatalf("Location = %q, want /admin/webhooks", got)
	}

	// Verify the webhook was updated
	updated, err := app.svc.Webhook.GetConfig(wh.ID)
	if err != nil {
		t.Fatalf("get webhook: %v", err)
	}
	if updated.Name != "Updated Webhook" {
		t.Fatalf("webhook name = %q, want %q", updated.Name, "Updated Webhook")
	}
	if updated.URL != "https://updated.example.com/hook" {
		t.Fatalf("webhook url = %q, want %q", updated.URL, "https://updated.example.com/hook")
	}
	if !updated.Enabled {
		t.Fatalf("webhook should be enabled")
	}
}

// TestAdminDeleteWebhook verifies that POSTing to delete an existing webhook
// returns a 303 redirect.
func TestAdminDeleteWebhook(t *testing.T) {
	app := newPublicTestApp(t)
	sessionCookie, csrfToken := loginAsAdmin(t, app)

	// Create a webhook to delete
	wh, err := app.svc.Webhook.CreateConfig("Delete Target", "https://delete.example.com/hook", "del-secret", false)
	if err != nil {
		t.Fatalf("create webhook: %v", err)
	}

	form := url.Values{"_csrf": {csrfToken}}
	req := httptest.NewRequest(http.MethodPost, "/admin/webhooks/delete/"+strconv.FormatInt(wh.ID, 10), strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(sessionCookie)
	testRequestCounter++
	req.RemoteAddr = fmt.Sprintf("203.0.113.%d:1234", testRequestCounter%250+1)
	rr := httptest.NewRecorder()
	app.handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusSeeOther, rr.Body.String())
	}
	if got := rr.Header().Get("Location"); got != "/admin/webhooks" {
		t.Fatalf("Location = %q, want /admin/webhooks", got)
	}

	// Verify the webhook was deleted
	_, err = app.svc.Webhook.GetConfig(wh.ID)
	if err == nil {
		t.Fatalf("webhook should have been deleted")
	}
}

// TestAdminUpdatePage verifies that POSTing valid page update data returns a
// 303 redirect to the pages list.
func TestAdminUpdatePage(t *testing.T) {
	app := newPublicTestApp(t)
	sessionCookie, csrfToken := loginAsAdmin(t, app)

	form := url.Values{
		"title":   {"Updated About Page"},
		"summary": {"Updated summary for the about page."},
		"content": {"# About\n\nThis about page was updated via the admin POST handler."},
		"_csrf":   {csrfToken},
	}
	req := httptest.NewRequest(http.MethodPost, "/admin/pages/update/about", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(sessionCookie)
	testRequestCounter++
	req.RemoteAddr = fmt.Sprintf("203.0.113.%d:1234", testRequestCounter%250+1)
	rr := httptest.NewRecorder()
	app.handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusSeeOther, rr.Body.String())
	}
	if got := rr.Header().Get("Location"); got != "/admin/pages" {
		t.Fatalf("Location = %q, want /admin/pages", got)
	}

	// Verify the page was updated
	page, err := app.svc.Pages.GetPage("about")
	if err != nil {
		t.Fatalf("get page: %v", err)
	}
	if page.Title != "Updated About Page" {
		t.Fatalf("page title = %q, want %q", page.Title, "Updated About Page")
	}
	if page.Summary != "Updated summary for the about page." {
		t.Fatalf("page summary = %q, want %q", page.Summary, "Updated summary for the about page.")
	}
}

// =============================================================================
// Admin error handling
// =============================================================================

// TestAdminCreatePostNoCategory verifies that POSTing a post without a category
// slug returns a 400 error.
func TestAdminCreatePostNoCategory(t *testing.T) {
	app := newPublicTestApp(t)
	sessionCookie, csrfToken := loginAsAdmin(t, app)

	form := url.Values{
		"title":   {"Post Without Category"},
		"content": {"# No category provided"},
		"_csrf":   {csrfToken},
	}
	req := httptest.NewRequest(http.MethodPost, "/admin/posts", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(sessionCookie)
	testRequestCounter++
	req.RemoteAddr = fmt.Sprintf("203.0.113.%d:1234", testRequestCounter%250+1)
	rr := httptest.NewRecorder()
	app.handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusBadRequest, rr.Body.String())
	}
	assertContains(t, rr.Body.String(), "title, category_slug, and content required")
}

// TestAdminUpdateMissingPost verifies that updating a non-existent post returns
// a 404 error.
func TestAdminUpdateMissingPost(t *testing.T) {
	app := newPublicTestApp(t)
	sessionCookie, csrfToken := loginAsAdmin(t, app)

	form := url.Values{
		"title":         {"Ghost Post"},
		"category_slug": {"devops"},
		"content":       {"# Ghost"},
		"_csrf":         {csrfToken},
	}
	req := httptest.NewRequest(http.MethodPost, "/admin/posts/update/99999", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(sessionCookie)
	testRequestCounter++
	req.RemoteAddr = fmt.Sprintf("203.0.113.%d:1234", testRequestCounter%250+1)
	rr := httptest.NewRecorder()
	app.handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusNotFound, rr.Body.String())
	}
	assertContains(t, rr.Body.String(), "post not found")
}

// TestAdminDeleteMissingProject verifies that deleting a non-existent project
// returns a 404 error.
func TestAdminDeleteMissingProject(t *testing.T) {
	app := newPublicTestApp(t)
	sessionCookie, csrfToken := loginAsAdmin(t, app)

	form := url.Values{"_csrf": {csrfToken}}
	req := httptest.NewRequest(http.MethodPost, "/admin/projects/delete/99999", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(sessionCookie)
	testRequestCounter++
	req.RemoteAddr = fmt.Sprintf("203.0.113.%d:1234", testRequestCounter%250+1)
	rr := httptest.NewRecorder()
	app.handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusNotFound, rr.Body.String())
	}
	assertContains(t, rr.Body.String(), "project not found")
}
// =============================================================================
// Admin topic edit (GET)
// =============================================================================

// TestAdminEditTopicGet verifies the topic edit form renders for an existing
// category with its name, slug, and CSRF field.
func TestAdminEditTopicGet(t *testing.T) {
	app := newPublicTestApp(t)
	sessionCookie, _ := loginAsAdmin(t, app)

	cats, err := app.svc.Blog.ListCategories()
	if err != nil {
		t.Fatalf("list categories: %v", err)
	}
	var catID int64
	for _, c := range cats {
		if c.Slug == "devops" {
			catID = c.ID
			break
		}
	}
	if catID == 0 {
		t.Fatalf("seed category not found")
	}

	status, body := getBody(t, app, "/admin/topics/edit/"+strconv.FormatInt(catID, 10), sessionCookie)
	if status != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", status, http.StatusOK, body)
	}
	assertContains(t, body, "Edit Topic")
	assertContains(t, body, "DevOps")
	assertContains(t, body, `name="_csrf"`)
}

// TestAdminEditTopicNotFound verifies that GET /admin/topics/edit/{id} with a
// non-existent ID returns 404.
func TestAdminEditTopicNotFound(t *testing.T) {
	app := newPublicTestApp(t)
	sessionCookie, _ := loginAsAdmin(t, app)

	status, body := getBody(t, app, "/admin/topics/edit/99999", sessionCookie)
	if status != http.StatusNotFound {
		t.Fatalf("status = %d, want %d; body: %s", status, http.StatusNotFound, body)
	}
	assertContains(t, body, "topic not found")
}

// =============================================================================
// Admin post edit (GET)
// =============================================================================

// TestAdminEditPostGet verifies the post edit form renders for an existing post
// with its title, intent buttons, and CSRF field.
func TestAdminEditPostGet(t *testing.T) {
	app := newPublicTestApp(t)
	sessionCookie, _ := loginAsAdmin(t, app)

	post, err := app.svc.Blog.GetPost("hello-raevtar")
	if err != nil {
		t.Fatalf("get seed post: %v", err)
	}

	status, body := getBody(t, app, "/admin/posts/edit/"+strconv.FormatInt(post.ID, 10), sessionCookie)
	if status != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", status, http.StatusOK, body)
	}
	assertContains(t, body, "Edit post")
	assertContains(t, body, "Hello Raevtar")
	assertContains(t, body, `name="_csrf"`)
	assertContains(t, body, `name="intent" value="draft"`)
	assertContains(t, body, `name="intent" value="update"`)
	assertContains(t, body, `name="intent" value="publish"`)
}

// TestAdminEditPostNotFound verifies that GET /admin/posts/edit/{id} with a
// non-existent ID returns 404.
func TestAdminEditPostNotFound(t *testing.T) {
	app := newPublicTestApp(t)
	sessionCookie, _ := loginAsAdmin(t, app)

	status, body := getBody(t, app, "/admin/posts/edit/99999", sessionCookie)
	if status != http.StatusNotFound {
		t.Fatalf("status = %d, want %d; body: %s", status, http.StatusNotFound, body)
	}
	assertContains(t, body, "post not found")
}

// =============================================================================
// Admin server detail (GET)
// =============================================================================

// TestAdminServerDetail verifies the admin server detail page renders with
// server metadata, command center, audit logs, and CSRF field.
func TestAdminServerDetail(t *testing.T) {
	app := newPublicTestApp(t)
	sessionCookie, _ := loginAsAdmin(t, app)

	path := "/admin/servers/" + strconv.FormatInt(app.serverID, 10)
	status, body := getBody(t, app, path, sessionCookie)
	if status != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", status, http.StatusOK, body)
	}
	assertContains(t, body, "Server Detail")
	assertContains(t, body, "whyred")
	assertContains(t, body, "Node Metadata")
	assertContains(t, body, "Remote Command Center")
	assertContains(t, body, "Audit Logs")
	assertContains(t, body, `name="_csrf"`)
}

// =============================================================================
// Admin server mutations (POST)
// =============================================================================

// TestAdminUpdateServer verifies that POSTing valid update data to an existing
// server returns a 303 redirect and updates the server record.
func TestAdminUpdateServer(t *testing.T) {
	app := newPublicTestApp(t)
	sessionCookie, csrfToken := loginAsAdmin(t, app)

	form := url.Values{
		"name":  {"whyred-updated"},
		"host":  {"127.0.0.1"},
		"port":  {"9100"},
		"tags":  {"local,updated"},
		"_csrf": {csrfToken},
	}
	req := httptest.NewRequest(http.MethodPost, "/admin/servers/update/"+strconv.FormatInt(app.serverID, 10), strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(sessionCookie)
	testRequestCounter++
	req.RemoteAddr = fmt.Sprintf("203.0.113.%d:1234", testRequestCounter%250+1)
	rr := httptest.NewRecorder()
	app.handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusSeeOther, rr.Body.String())
	}
	wantLoc := "/admin/servers/" + strconv.FormatInt(app.serverID, 10)
	if got := rr.Header().Get("Location"); got != wantLoc {
		t.Fatalf("Location = %q, want %q", got, wantLoc)
	}

	// Verify the server was updated
	server, err := app.svc.Monitor.GetServer(app.serverID)
	if err != nil {
		t.Fatalf("get server: %v", err)
	}
	if server.Name != "whyred-updated" {
		t.Fatalf("server name = %q, want %q", server.Name, "whyred-updated")
	}
	if server.Host != "127.0.0.1" {
		t.Fatalf("server host = %q, want %q", server.Host, "127.0.0.1")
	}
	if server.Port != 9100 {
		t.Fatalf("server port = %d, want %d", server.Port, 9100)
	}
	if server.Tags != "local,updated" {
		t.Fatalf("server tags = %q, want %q", server.Tags, "local,updated")
	}
}

// TestAdminRotateServerToken verifies that POSTing to rotate the agent token
// returns a 303 redirect and updates the server's token hash.
func TestAdminRotateServerToken(t *testing.T) {
	app := newPublicTestApp(t)
	sessionCookie, csrfToken := loginAsAdmin(t, app)

	form := url.Values{"_csrf": {csrfToken}}
	req := httptest.NewRequest(http.MethodPost, "/admin/servers/rotate-token/"+strconv.FormatInt(app.serverID, 10), strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(sessionCookie)
	testRequestCounter++
	req.RemoteAddr = fmt.Sprintf("203.0.113.%d:1234", testRequestCounter%250+1)
	rr := httptest.NewRecorder()
	app.handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusSeeOther, rr.Body.String())
	}
	if got := rr.Header().Get("Location"); got != "/admin/servers" {
		t.Fatalf("Location = %q, want /admin/servers", got)
	}

	// Verify the server now has a non-empty token hash
	server, err := app.svc.Monitor.GetServer(app.serverID)
	if err != nil {
		t.Fatalf("get server: %v", err)
	}
	if server.AgentTokenHash == "" {
		t.Fatalf("expected non-empty agent token hash after rotation")
	}
}
