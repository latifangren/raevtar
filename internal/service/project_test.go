package service

import (
	"errors"
	"testing"
	"time"

	"raevtar/internal/model"
)

func createTestProject(t *testing.T, state *testServices) *model.Project {
	t.Helper()
	proj, err := state.svc.Projects.CreateProject(model.ProjectCreate{
		Title:     "Test Project",
		ContentMD: "# Test\nContent",
		Excerpt:   "Test excerpt",
		Published: true,
		Tags:      []string{"test"},
	})
	if err != nil {
		t.Fatalf("create project: %v", err)
	}
	return proj
}

func TestProjectServiceCreateUpdate(t *testing.T) {
	state := newTestServices(t)
	proj := createTestProject(t, state)

	update, err := state.svc.Projects.CreateProjectUpdate(proj.ID, model.ProjectUpdateEntryCreate{
		Kind:      model.ProjectUpdateKindChangelog,
		Title:     "v1.0 Release",
		ContentMD: "## Changes\n\n- Added feature X",
		Published: true,
	})
	if err != nil {
		t.Fatalf("create project update: %v", err)
	}
	if update.ID == 0 {
		t.Fatalf("update id = 0, want persisted id")
	}
	if update.Kind != model.ProjectUpdateKindChangelog {
		t.Fatalf("update kind = %q, want %q", update.Kind, model.ProjectUpdateKindChangelog)
	}
	if update.Title != "v1.0 Release" {
		t.Fatalf("update title = %q, want %q", update.Title, "v1.0 Release")
	}
	if !update.Published {
		t.Fatalf("update published = false, want true")
	}
	if update.ProjectID != proj.ID {
		t.Fatalf("update project_id = %d, want %d", update.ProjectID, proj.ID)
	}

	updates, err := state.svc.Projects.ListProjectUpdates(proj.ID, false, nil, 10)
	if err != nil {
		t.Fatalf("list project updates: %v", err)
	}
	if len(updates) != 1 {
		t.Fatalf("updates len = %d, want 1", len(updates))
	}
	if updates[0].ID != update.ID {
		t.Fatalf("listed update id = %d, want %d", updates[0].ID, update.ID)
	}
	if updates[0].Title != "v1.0 Release" {
		t.Fatalf("listed update title = %q, want %q", updates[0].Title, "v1.0 Release")
	}
	if updates[0].ContentHTML == "" {
		t.Fatalf("expected rendered content_html, got empty")
	}
}

func TestProjectServiceCreateUpdateWithInvalidType(t *testing.T) {
	state := newTestServices(t)
	proj := createTestProject(t, state)

	_, err := state.svc.Projects.CreateProjectUpdate(proj.ID, model.ProjectUpdateEntryCreate{
		Kind:      "roadmap",
		Title:     "Invalid Update",
		ContentMD: "This kind does not exist",
	})
	if err == nil {
		t.Fatalf("expected error for invalid update kind")
	}
	if !errors.Is(err, ErrInvalidProjectUpdateKind) {
		t.Fatalf("err = %v, want ErrInvalidProjectUpdateKind", err)
	}
}

func TestProjectServiceListProjectUpdatesFiltersByKinds(t *testing.T) {
	state := newTestServices(t)
	proj := createTestProject(t, state)

	kinds := []string{model.ProjectUpdateKindChangelog, model.ProjectUpdateKindTimeline, model.ProjectUpdateKindBuildLog}
	created := make(map[string]int64)
	for _, kind := range kinds {
		u, err := state.svc.Projects.CreateProjectUpdate(proj.ID, model.ProjectUpdateEntryCreate{
			Kind:      kind,
			Title:     "Update " + kind,
			ContentMD: "# " + kind,
			Published: true,
		})
		if err != nil {
			t.Fatalf("create %s update: %v", kind, err)
		}
		created[kind] = u.ID
	}

	// Filter by changelog only
	changelogs, err := state.svc.Projects.ListProjectUpdates(proj.ID, false, []string{model.ProjectUpdateKindChangelog}, 10)
	if err != nil {
		t.Fatalf("list updates filtered by changelog: %v", err)
	}
	if len(changelogs) != 1 {
		t.Fatalf("changelogs len = %d, want 1", len(changelogs))
	}
	if changelogs[0].ID != created[model.ProjectUpdateKindChangelog] {
		t.Fatalf("changelog id = %d, want %d", changelogs[0].ID, created[model.ProjectUpdateKindChangelog])
	}

	// Filter by timeline only
	timelines, err := state.svc.Projects.ListProjectUpdates(proj.ID, false, []string{model.ProjectUpdateKindTimeline}, 10)
	if err != nil {
		t.Fatalf("list updates filtered by timeline: %v", err)
	}
	if len(timelines) != 1 {
		t.Fatalf("timelines len = %d, want 1", len(timelines))
	}
	if timelines[0].ID != created[model.ProjectUpdateKindTimeline] {
		t.Fatalf("timeline id = %d, want %d", timelines[0].ID, created[model.ProjectUpdateKindTimeline])
	}

	// Filter by multiple kinds (changelog + build_log)
	multi, err := state.svc.Projects.ListProjectUpdates(proj.ID, false, []string{model.ProjectUpdateKindChangelog, model.ProjectUpdateKindBuildLog}, 10)
	if err != nil {
		t.Fatalf("list updates filtered by multiple kinds: %v", err)
	}
	if len(multi) != 2 {
		t.Fatalf("multi-filter len = %d, want 2", len(multi))
	}
}

func TestProjectServiceListProjectTimeline(t *testing.T) {
	state := newTestServices(t)
	proj := createTestProject(t, state)

	now := time.Now()
	_, err := state.svc.Projects.CreateProjectUpdate(proj.ID, model.ProjectUpdateEntryCreate{
		Kind:      model.ProjectUpdateKindTimeline,
		Title:     "Milestone Alpha",
		ContentMD: "Alpha completed",
		Published: true,
		EventAt:   now.Add(-48 * time.Hour),
	})
	if err != nil {
		t.Fatalf("create timeline update: %v", err)
	}

	_, err = state.svc.Projects.CreateProjectUpdate(proj.ID, model.ProjectUpdateEntryCreate{
		Kind:      model.ProjectUpdateKindBuildLog,
		Title:     "Build #42",
		ContentMD: "Build passed",
		Published: true,
		EventAt:   now.Add(-24 * time.Hour),
	})
	if err != nil {
		t.Fatalf("create build_log update: %v", err)
	}

	_, err = state.svc.Projects.CreateProjectUpdate(proj.ID, model.ProjectUpdateEntryCreate{
		Kind:      model.ProjectUpdateKindChangelog,
		Title:     "v1.1 Changelog",
		ContentMD: "Bug fixes",
		Published: true,
		EventAt:   now,
	})
	if err != nil {
		t.Fatalf("create changelog update: %v", err)
	}

	items, err := state.svc.Projects.ListProjectTimeline(proj.ID, false, 10)
	if err != nil {
		t.Fatalf("list project timeline: %v", err)
	}
	if len(items) != 3 {
		t.Fatalf("timeline items len = %d, want 3", len(items))
	}
}

func TestProjectServiceListProjectChangelog(t *testing.T) {
	state := newTestServices(t)
	proj := createTestProject(t, state)

	_, err := state.svc.Projects.CreateProjectUpdate(proj.ID, model.ProjectUpdateEntryCreate{
		Kind:      model.ProjectUpdateKindChangelog,
		Title:     "v1.0 Changelog",
		ContentMD: "Initial release",
		Published: true,
	})
	if err != nil {
		t.Fatalf("create changelog: %v", err)
	}

	_, err = state.svc.Projects.CreateProjectUpdate(proj.ID, model.ProjectUpdateEntryCreate{
		Kind:      model.ProjectUpdateKindChangelog,
		Title:     "v1.1 Changelog",
		ContentMD: "Bug fixes",
		Published: true,
	})
	if err != nil {
		t.Fatalf("create second changelog: %v", err)
	}

	_, err = state.svc.Projects.CreateProjectUpdate(proj.ID, model.ProjectUpdateEntryCreate{
		Kind:      model.ProjectUpdateKindTimeline,
		Title:     "Status Update",
		ContentMD: "Progress report",
		Published: true,
	})
	if err != nil {
		t.Fatalf("create timeline: %v", err)
	}

	changelogs, err := state.svc.Projects.ListProjectChangelog(proj.ID, false, 10)
	if err != nil {
		t.Fatalf("list project changelog: %v", err)
	}
	if len(changelogs) != 2 {
		t.Fatalf("changelog items len = %d, want 2", len(changelogs))
	}
	for _, item := range changelogs {
		if item.Kind != model.ProjectUpdateKindChangelog {
			t.Fatalf("item kind = %q, want %q", item.Kind, model.ProjectUpdateKindChangelog)
		}
	}
}

func TestProjectServiceCreateShowcase(t *testing.T) {
	state := newTestServices(t)
	proj := createTestProject(t, state)

	showcase, err := state.svc.Projects.CreateProjectShowcase(proj.ID, model.ProjectShowcaseItemCreate{
		Kind:        model.ProjectShowcaseKindLink,
		Title:       "Awesome Demo",
		BodyMD:      "A live demo of the project",
		ExternalURL: "https://demo.example.com",
		Published:   true,
	})
	if err != nil {
		t.Fatalf("create project showcase: %v", err)
	}
	if showcase.ID == 0 {
		t.Fatalf("showcase id = 0, want persisted id")
	}
	if showcase.Title != "Awesome Demo" {
		t.Fatalf("showcase title = %q, want %q", showcase.Title, "Awesome Demo")
	}
	if showcase.Kind != model.ProjectShowcaseKindLink {
		t.Fatalf("showcase kind = %q, want %q", showcase.Kind, model.ProjectShowcaseKindLink)
	}
	if showcase.ExternalURL != "https://demo.example.com" {
		t.Fatalf("showcase external_url = %q, want %q", showcase.ExternalURL, "https://demo.example.com")
	}
	if !showcase.Published {
		t.Fatalf("showcase published = false, want true")
	}

	items, err := state.svc.Projects.ListProjectShowcase(proj.ID, false)
	if err != nil {
		t.Fatalf("list project showcase: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("showcase items len = %d, want 1", len(items))
	}
	if items[0].ID != showcase.ID {
		t.Fatalf("listed showcase id = %d, want %d", items[0].ID, showcase.ID)
	}
	if items[0].ExternalURL != "https://demo.example.com" {
		t.Fatalf("listed showcase url = %q, want %q", items[0].ExternalURL, "https://demo.example.com")
	}
}

func TestProjectServiceCreateShowcaseWithImage(t *testing.T) {
	state := newTestServices(t)
	proj := createTestProject(t, state)

	showcase, err := state.svc.Projects.CreateProjectShowcase(proj.ID, model.ProjectShowcaseItemCreate{
		Kind:     model.ProjectShowcaseKindImage,
		Title:    "Screenshot",
		BodyMD:   "Application screenshot",
		AssetURL: "/static/uploads/screenshot.png",
	})
	if err != nil {
		t.Fatalf("create showcase with image: %v", err)
	}
	if showcase.AssetURL != "/static/uploads/screenshot.png" {
		t.Fatalf("showcase asset_url = %q, want %q", showcase.AssetURL, "/static/uploads/screenshot.png")
	}
	if showcase.Kind != model.ProjectShowcaseKindImage {
		t.Fatalf("showcase kind = %q, want %q", showcase.Kind, model.ProjectShowcaseKindImage)
	}

	items, err := state.svc.Projects.ListProjectShowcase(proj.ID, false)
	if err != nil {
		t.Fatalf("list project showcase: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("showcase items len = %d, want 1", len(items))
	}
	if items[0].AssetURL != "/static/uploads/screenshot.png" {
		t.Fatalf("listed asset_url = %q, want %q", items[0].AssetURL, "/static/uploads/screenshot.png")
	}
}

func TestProjectServiceCreateRelation(t *testing.T) {
	state := newTestServices(t)

	projA, err := state.svc.Projects.CreateProject(model.ProjectCreate{
		Title:     "Project Alpha",
		ContentMD: "# Alpha",
		Excerpt:   "First project",
		Published: true,
	})
	if err != nil {
		t.Fatalf("create project alpha: %v", err)
	}

	projB, err := state.svc.Projects.CreateProject(model.ProjectCreate{
		Title:     "Project Beta",
		ContentMD: "# Beta",
		Excerpt:   "Second project",
		Published: true,
	})
	if err != nil {
		t.Fatalf("create project beta: %v", err)
	}

	rel, err := state.svc.Projects.CreateProjectRelation(projA.ID, model.ContentRelationCreate{
		TargetType:   model.ContentRelationTypeProject,
		TargetID:     projB.ID,
		RelationKind: model.ContentRelationKindRelated,
	})
	if err != nil {
		t.Fatalf("create project relation: %v", err)
	}
	if rel.ID == 0 {
		t.Fatalf("relation id = 0, want persisted id")
	}
	if rel.SourceID != projA.ID {
		t.Fatalf("relation source_id = %d, want %d", rel.SourceID, projA.ID)
	}
	if rel.TargetID != projB.ID {
		t.Fatalf("relation target_id = %d, want %d", rel.TargetID, projB.ID)
	}
	if rel.TargetType != model.ContentRelationTypeProject {
		t.Fatalf("relation target_type = %q, want %q", rel.TargetType, model.ContentRelationTypeProject)
	}
	if rel.RelationKind != model.ContentRelationKindRelated {
		t.Fatalf("relation kind = %q, want %q", rel.RelationKind, model.ContentRelationKindRelated)
	}

	relations, err := state.svc.Projects.ListProjectRelations(projA.ID)
	if err != nil {
		t.Fatalf("list project relations: %v", err)
	}
	if len(relations) != 1 {
		t.Fatalf("relations len = %d, want 1", len(relations))
	}
	if relations[0].TargetID != projB.ID {
		t.Fatalf("listed relation target_id = %d, want %d", relations[0].TargetID, projB.ID)
	}
}

func TestProjectServiceGetResolvedRelations(t *testing.T) {
	state := newTestServices(t)

	projA, err := state.svc.Projects.CreateProject(model.ProjectCreate{
		Title:     "Primary Project",
		ContentMD: "# Primary",
		Excerpt:   "The main project",
		Published: true,
	})
	if err != nil {
		t.Fatalf("create primary project: %v", err)
	}

	projB, err := state.svc.Projects.CreateProject(model.ProjectCreate{
		Title:     "Related Project",
		ContentMD: "# Related",
		Excerpt:   "A related project",
		Published: true,
	})
	if err != nil {
		t.Fatalf("create related project: %v", err)
	}

	_, err = state.svc.Projects.CreateProjectRelation(projA.ID, model.ContentRelationCreate{
		TargetType:   model.ContentRelationTypeProject,
		TargetID:     projB.ID,
		RelationKind: model.ContentRelationKindRelated,
	})
	if err != nil {
		t.Fatalf("create project relation: %v", err)
	}

	views, err := state.svc.Projects.GetResolvedProjectRelations(projA.ID, false)
	if err != nil {
		t.Fatalf("get resolved relations: %v", err)
	}
	if len(views) != 1 {
		t.Fatalf("resolved views len = %d, want 1", len(views))
	}
	if views[0].TargetID != projB.ID {
		t.Fatalf("view target_id = %d, want %d", views[0].TargetID, projB.ID)
	}
	if views[0].Title != "Related Project" {
		t.Fatalf("view title = %q, want %q", views[0].Title, "Related Project")
	}
	if views[0].Slug != projB.Slug {
		t.Fatalf("view slug = %q, want %q", views[0].Slug, projB.Slug)
	}
	if views[0].Excerpt != "A related project" {
		t.Fatalf("view excerpt = %q, want %q", views[0].Excerpt, "A related project")
	}
	if views[0].URL != "/projects/"+projB.Slug {
		t.Fatalf("view url = %q, want %q", views[0].URL, "/projects/"+projB.Slug)
	}
	if !views[0].Published {
		t.Fatalf("view published = false, want true")
	}

	// Verify publishedOnly filtering hides unpublished relations
	projC, err := state.svc.Projects.CreateProject(model.ProjectCreate{
		Title:     "Unpublished Project",
		ContentMD: "# Unpublished",
		Excerpt:   "Not yet published",
		Published: false,
	})
	if err != nil {
		t.Fatalf("create unpublished project: %v", err)
	}

	_, err = state.svc.Projects.CreateProjectRelation(projA.ID, model.ContentRelationCreate{
		TargetType:   model.ContentRelationTypeProject,
		TargetID:     projC.ID,
		RelationKind: model.ContentRelationKindRelated,
	})
	if err != nil {
		t.Fatalf("create relation to unpublished: %v", err)
	}

	// publishedOnly=true should exclude the unpublished relation
	publishedViews, err := state.svc.Projects.GetResolvedProjectRelations(projA.ID, true)
	if err != nil {
		t.Fatalf("get resolved relations publishedOnly: %v", err)
	}
	if len(publishedViews) != 1 {
		t.Fatalf("published-only views len = %d, want 1 (excluding unpublished)", len(publishedViews))
	}
	if publishedViews[0].TargetID != projB.ID {
		t.Fatalf("published-only view target_id = %d, want %d", publishedViews[0].TargetID, projB.ID)
	}

	// publishedOnly=false should include both
	allViews, err := state.svc.Projects.GetResolvedProjectRelations(projA.ID, false)
	if err != nil {
		t.Fatalf("get resolved relations all: %v", err)
	}
	if len(allViews) != 2 {
		t.Fatalf("all views len = %d, want 2", len(allViews))
	}
}

// -- UpdateProjectUpdate -----------------------------------------------------

func TestProjectServiceUpdateProjectUpdate(t *testing.T) {
	state := newTestServices(t)
	proj := createTestProject(t, state)

	update, err := state.svc.Projects.CreateProjectUpdate(proj.ID, model.ProjectUpdateEntryCreate{
		Kind:      model.ProjectUpdateKindChangelog,
		Title:     "Original Title",
		ContentMD: "## Original\n\nContent",
		Published: true,
	})
	if err != nil {
		t.Fatalf("create project update: %v", err)
	}

	updated, err := state.svc.Projects.UpdateProjectUpdate(update.ID, model.ProjectUpdateEntryUpdate{
		Kind:      model.ProjectUpdateKindChangelog,
		Title:     "Updated Title",
		ContentMD: "## Updated\n\nNew content",
		Published: false,
	})
	if err != nil {
		t.Fatalf("UpdateProjectUpdate: %v", err)
	}
	if updated.ID != update.ID {
		t.Fatalf("updated id = %d, want %d", updated.ID, update.ID)
	}
	if updated.Title != "Updated Title" {
		t.Fatalf("updated title = %q, want Updated Title", updated.Title)
	}
	if updated.ContentMD != "## Updated\n\nNew content" {
		t.Fatalf("updated content_md = %q, want updated content", updated.ContentMD)
	}
	if updated.Published {
		t.Fatalf("updated published = true, want false")
	}
	if updated.ContentHTML == "" {
		t.Fatalf("updated content_html is empty (should be rendered)")
	}

	// Verify via list
	updates, err := state.svc.Projects.ListProjectUpdates(proj.ID, false, nil, 10)
	if err != nil {
		t.Fatalf("list project updates after update: %v", err)
	}
	if len(updates) != 1 {
		t.Fatalf("updates len = %d, want 1", len(updates))
	}
	if updates[0].Title != "Updated Title" {
		t.Fatalf("listed update title = %q, want Updated Title", updates[0].Title)
	}
}

func TestProjectServiceUpdateProjectUpdateNotFound(t *testing.T) {
	state := newTestServices(t)

	_, err := state.svc.Projects.UpdateProjectUpdate(9999, model.ProjectUpdateEntryUpdate{
		Kind:      model.ProjectUpdateKindChangelog,
		Title:     "Ghost Update",
		ContentMD: "## Missing",
	})
	if err == nil {
		t.Fatalf("expected error for updating nonexistent update")
	}
	if !errors.Is(err, ErrProjectUpdateNotFound) {
		t.Fatalf("err = %v, want ErrProjectUpdateNotFound", err)
	}
}

// -- DeleteProjectUpdate -----------------------------------------------------

func TestProjectServiceDeleteProjectUpdate(t *testing.T) {
	state := newTestServices(t)
	proj := createTestProject(t, state)

	update, err := state.svc.Projects.CreateProjectUpdate(proj.ID, model.ProjectUpdateEntryCreate{
		Kind:      model.ProjectUpdateKindTimeline,
		Title:     "Delete Me Update",
		ContentMD: "## Delete\n\nMe",
		Published: true,
	})
	if err != nil {
		t.Fatalf("create project update: %v", err)
	}

	// Verify it exists
	updates, err := state.svc.Projects.ListProjectUpdates(proj.ID, false, nil, 10)
	if err != nil {
		t.Fatalf("list project updates before delete: %v", err)
	}
	if len(updates) != 1 {
		t.Fatalf("updates before delete = %d, want 1", len(updates))
	}

	if err := state.svc.Projects.DeleteProjectUpdate(update.ID); err != nil {
		t.Fatalf("DeleteProjectUpdate: %v", err)
	}

	updates, err = state.svc.Projects.ListProjectUpdates(proj.ID, false, nil, 10)
	if err != nil {
		t.Fatalf("list project updates after delete: %v", err)
	}
	if len(updates) != 0 {
		t.Fatalf("updates after delete = %d, want 0", len(updates))
	}
}

func TestProjectServiceDeleteProjectUpdateNotFound(t *testing.T) {
	state := newTestServices(t)

	err := state.svc.Projects.DeleteProjectUpdate(9999)
	if err == nil {
		t.Fatalf("expected error for deleting nonexistent update")
	}
	if !errors.Is(err, ErrProjectUpdateNotFound) {
		t.Fatalf("err = %v, want ErrProjectUpdateNotFound", err)
	}
}

// -- UpdateProjectShowcase ---------------------------------------------------

func TestProjectServiceUpdateProjectShowcase(t *testing.T) {
	state := newTestServices(t)
	proj := createTestProject(t, state)

	showcase, err := state.svc.Projects.CreateProjectShowcase(proj.ID, model.ProjectShowcaseItemCreate{
		Kind:        model.ProjectShowcaseKindLink,
		Title:       "Original Showcase",
		BodyMD:      "Original description",
		ExternalURL: "https://original.example.com",
		Published:   true,
	})
	if err != nil {
		t.Fatalf("create project showcase: %v", err)
	}

	updated, err := state.svc.Projects.UpdateProjectShowcase(showcase.ID, model.ProjectShowcaseItemUpdate{
		Kind:        model.ProjectShowcaseKindLink,
		Title:       "Updated Showcase",
		BodyMD:      "Updated description",
		ExternalURL: "https://updated.example.com",
		Published:   false,
	})
	if err != nil {
		t.Fatalf("UpdateProjectShowcase: %v", err)
	}
	if updated.ID != showcase.ID {
		t.Fatalf("updated id = %d, want %d", updated.ID, showcase.ID)
	}
	if updated.Title != "Updated Showcase" {
		t.Fatalf("updated title = %q, want Updated Showcase", updated.Title)
	}
	if updated.BodyMD != "Updated description" {
		t.Fatalf("updated body_md = %q, want Updated description", updated.BodyMD)
	}
	if updated.ExternalURL != "https://updated.example.com" {
		t.Fatalf("updated external_url = %q, want https://updated.example.com", updated.ExternalURL)
	}
	if updated.Published {
		t.Fatalf("updated published = true, want false")
	}
	if updated.BodyHTML == "" {
		t.Fatalf("updated body_html is empty (should be rendered)")
	}

	// Verify via list
	items, err := state.svc.Projects.ListProjectShowcase(proj.ID, false)
	if err != nil {
		t.Fatalf("list project showcase after update: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("showcase items len = %d, want 1", len(items))
	}
	if items[0].Title != "Updated Showcase" {
		t.Fatalf("listed showcase title = %q, want Updated Showcase", items[0].Title)
	}
}

func TestProjectServiceUpdateProjectShowcaseNotFound(t *testing.T) {
	state := newTestServices(t)

	_, err := state.svc.Projects.UpdateProjectShowcase(9999, model.ProjectShowcaseItemUpdate{
		Kind:      model.ProjectShowcaseKindLink,
		Title:     "Ghost Showcase",
		BodyMD:    "Missing",
	})
	if err == nil {
		t.Fatalf("expected error for updating nonexistent showcase")
	}
	if !errors.Is(err, ErrProjectShowcaseNotFound) {
		t.Fatalf("err = %v, want ErrProjectShowcaseNotFound", err)
	}
}

// -- DeleteProjectShowcase ---------------------------------------------------

func TestProjectServiceDeleteProjectShowcase(t *testing.T) {
	state := newTestServices(t)
	proj := createTestProject(t, state)

	showcase, err := state.svc.Projects.CreateProjectShowcase(proj.ID, model.ProjectShowcaseItemCreate{
		Kind:     model.ProjectShowcaseKindImage,
		Title:    "Delete Me Showcase",
		BodyMD:   "To be deleted",
		AssetURL: "/static/uploads/delete.png",
	})
	if err != nil {
		t.Fatalf("create project showcase: %v", err)
	}

	// Verify it exists
	items, err := state.svc.Projects.ListProjectShowcase(proj.ID, false)
	if err != nil {
		t.Fatalf("list project showcase before delete: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("showcase items before delete = %d, want 1", len(items))
	}

	if err := state.svc.Projects.DeleteProjectShowcase(showcase.ID); err != nil {
		t.Fatalf("DeleteProjectShowcase: %v", err)
	}

	items, err = state.svc.Projects.ListProjectShowcase(proj.ID, false)
	if err != nil {
		t.Fatalf("list project showcase after delete: %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("showcase items after delete = %d, want 0", len(items))
	}
}

func TestProjectServiceDeleteProjectShowcaseNotFound(t *testing.T) {
	state := newTestServices(t)

	err := state.svc.Projects.DeleteProjectShowcase(9999)
	if err == nil {
		t.Fatalf("expected error for deleting nonexistent showcase")
	}
	if !errors.Is(err, ErrProjectShowcaseNotFound) {
		t.Fatalf("err = %v, want ErrProjectShowcaseNotFound", err)
	}
}

// -- DeleteProjectRelation ---------------------------------------------------

func TestProjectServiceDeleteProjectRelation(t *testing.T) {
	state := newTestServices(t)

	projA, err := state.svc.Projects.CreateProject(model.ProjectCreate{
		Title:     "Source Project",
		ContentMD: "# Source",
		Excerpt:   "Source project",
		Published: true,
	})
	if err != nil {
		t.Fatalf("create source project: %v", err)
	}

	projB, err := state.svc.Projects.CreateProject(model.ProjectCreate{
		Title:     "Target Project",
		ContentMD: "# Target",
		Excerpt:   "Target project",
		Published: true,
	})
	if err != nil {
		t.Fatalf("create target project: %v", err)
	}

	rel, err := state.svc.Projects.CreateProjectRelation(projA.ID, model.ContentRelationCreate{
		TargetType:   model.ContentRelationTypeProject,
		TargetID:     projB.ID,
		RelationKind: model.ContentRelationKindRelated,
	})
	if err != nil {
		t.Fatalf("create project relation: %v", err)
	}

	// Verify it exists
	relations, err := state.svc.Projects.ListProjectRelations(projA.ID)
	if err != nil {
		t.Fatalf("list project relations before delete: %v", err)
	}
	if len(relations) != 1 {
		t.Fatalf("relations before delete = %d, want 1", len(relations))
	}

	if err := state.svc.Projects.DeleteProjectRelation(rel.ID); err != nil {
		t.Fatalf("DeleteProjectRelation: %v", err)
	}

	relations, err = state.svc.Projects.ListProjectRelations(projA.ID)
	if err != nil {
		t.Fatalf("list project relations after delete: %v", err)
	}
	if len(relations) != 0 {
		t.Fatalf("relations after delete = %d, want 0", len(relations))
	}
}
