package model

import (
	"testing"
)

// ---------------------------------------------------------------------------
// Editorial validators
// ---------------------------------------------------------------------------

func TestValidEditorialModes(t *testing.T) {
	got := ValidEditorialModes()
	want := []string{"scheduled_assignment", "opportunistic_assignment", "campaign_theme", "autonomous_seed"}
	if len(got) != len(want) {
		t.Fatalf("ValidEditorialModes() returned %d items, want %d: got %v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("ValidEditorialModes()[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestValidEditorialStatuses(t *testing.T) {
	got := ValidEditorialStatuses()
	want := []string{"queued", "approved", "paused", "running", "failed", "done", "cancelled"}
	if len(got) != len(want) {
		t.Fatalf("ValidEditorialStatuses() returned %d items, want %d: got %v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("ValidEditorialStatuses()[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestIsValidEditorialMode(t *testing.T) {
	validModes := []string{
		EditorialModeScheduled,
		EditorialModeOpportunistic,
		EditorialModeCampaign,
		EditorialModeSeed,
	}
	for _, m := range validModes {
		if !IsValidEditorialMode(m) {
			t.Fatalf("IsValidEditorialMode(%q) = false, want true", m)
		}
	}
	invalidModes := []string{"", "unknown", "scheduled", "campaign"}
	for _, m := range invalidModes {
		if IsValidEditorialMode(m) {
			t.Fatalf("IsValidEditorialMode(%q) = true, want false", m)
		}
	}
}

func TestIsValidEditorialStatus(t *testing.T) {
	validStatuses := []string{
		EditorialStatusQueued,
		EditorialStatusApproved,
		EditorialStatusPaused,
		EditorialStatusRunning,
		EditorialStatusFailed,
		EditorialStatusDone,
		EditorialStatusCancelled,
	}
	for _, s := range validStatuses {
		if !IsValidEditorialStatus(s) {
			t.Fatalf("IsValidEditorialStatus(%q) = false, want true", s)
		}
	}
	invalidStatuses := []string{"", "unknown", "pending", "complete"}
	for _, s := range invalidStatuses {
		if IsValidEditorialStatus(s) {
			t.Fatalf("IsValidEditorialStatus(%q) = true, want false", s)
		}
	}
}

// ---------------------------------------------------------------------------
// User validators
// ---------------------------------------------------------------------------

func TestValidRoles(t *testing.T) {
	got := ValidRoles()
	want := []string{RoleOwner, RoleAdmin, RoleOperator, RoleReadonly}
	if len(got) != len(want) {
		t.Fatalf("ValidRoles() returned %d items, want %d: got %v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("ValidRoles()[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestRoleLevel(t *testing.T) {
	cases := []struct {
		role string
		want int
	}{
		{RoleOwner, 0},
		{RoleAdmin, 1},
		{RoleOperator, 2},
		{RoleReadonly, 3},
		{"unknown", 999},
		{"", 999},
	}
	for _, c := range cases {
		got := RoleLevel(c.role)
		if got != c.want {
			t.Fatalf("RoleLevel(%q) = %d, want %d", c.role, got, c.want)
		}
	}
}

func TestIsValidRole(t *testing.T) {
	validRoles := []string{RoleOwner, RoleAdmin, RoleOperator, RoleReadonly}
	for _, r := range validRoles {
		if !IsValidRole(r) {
			t.Fatalf("IsValidRole(%q) = false, want true", r)
		}
	}
	invalidRoles := []string{"", "superadmin", "editor", "viewer"}
	for _, r := range invalidRoles {
		if IsValidRole(r) {
			t.Fatalf("IsValidRole(%q) = true, want false", r)
		}
	}
}

func TestCanManage(t *testing.T) {
	// owner (level 0) can manage everyone
	if !CanManage(RoleOwner, RoleAdmin) {
		t.Fatalf("CanManage(owner, admin) = false, want true")
	}
	if !CanManage(RoleOwner, RoleOperator) {
		t.Fatalf("CanManage(owner, operator) = false, want true")
	}
	if !CanManage(RoleOwner, RoleReadonly) {
		t.Fatalf("CanManage(owner, readonly) = false, want true")
	}

	// admin (level 1) can manage lower-privilege roles
	if !CanManage(RoleAdmin, RoleOperator) {
		t.Fatalf("CanManage(admin, operator) = false, want true")
	}
	if !CanManage(RoleAdmin, RoleReadonly) {
		t.Fatalf("CanManage(admin, readonly) = false, want true")
	}

	// operator (level 2) can manage readonly (level 3)
	if !CanManage(RoleOperator, RoleReadonly) {
		t.Fatalf("CanManage(operator, readonly) = false, want true")
	}

	// readonly (level 3) cannot manage anyone
	if CanManage(RoleReadonly, RoleOwner) {
		t.Fatalf("CanManage(readonly, owner) = true, want false")
	}
	if CanManage(RoleReadonly, RoleAdmin) {
		t.Fatalf("CanManage(readonly, admin) = true, want false")
	}
	if CanManage(RoleReadonly, RoleOperator) {
		t.Fatalf("CanManage(readonly, operator) = true, want false")
	}

	// cannot manage same level
	if CanManage(RoleOwner, RoleOwner) {
		t.Fatalf("CanManage(owner, owner) = true, want false")
	}
	if CanManage(RoleAdmin, RoleAdmin) {
		t.Fatalf("CanManage(admin, admin) = true, want false")
	}
	if CanManage(RoleOperator, RoleOperator) {
		t.Fatalf("CanManage(operator, operator) = true, want false")
	}
	if CanManage(RoleReadonly, RoleReadonly) {
		t.Fatalf("CanManage(readonly, readonly) = true, want false")
	}

	// higher-privilege cannot be managed by lower-privilege
	if CanManage(RoleAdmin, RoleOwner) {
		t.Fatalf("CanManage(admin, owner) = true, want false")
	}
	if CanManage(RoleOperator, RoleOwner) {
		t.Fatalf("CanManage(operator, owner) = true, want false")
	}
	if CanManage(RoleOperator, RoleAdmin) {
		t.Fatalf("CanManage(operator, admin) = true, want false")
	}
}

// ---------------------------------------------------------------------------
// JSON-LD
// ---------------------------------------------------------------------------

func TestMustJSONLD(t *testing.T) {
	// valid input returns JSON string
	input := map[string]string{"@context": "https://schema.org"}
	got := MustJSONLD(input)
	want := `{"@context":"https://schema.org"}`
	if got != want {
		t.Fatalf("MustJSONLD(%v) = %q, want %q", input, got, want)
	}

	// struct with json tags
	type person struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}
	got = MustJSONLD(person{Name: "Alice", Age: 30})
	want = `{"name":"Alice","age":30}`
	if got != want {
		t.Fatalf("MustJSONLD(person{Alice, 30}) = %q, want %q", got, want)
	}

	// nil input marshals to "null"
	got = MustJSONLD(nil)
	if got != "null" {
		t.Fatalf("MustJSONLD(nil) = %q, want %q", got, "null")
	}

	// error-producing input panics
	didPanic := false
	func() {
		defer func() {
			if r := recover(); r != nil {
				didPanic = true
			}
		}()
		MustJSONLD(make(chan int))
	}()
	if !didPanic {
		t.Fatalf("MustJSONLD(make(chan int)): want panic, got none")
	}
}

// ---------------------------------------------------------------------------
// Project validators
// ---------------------------------------------------------------------------

func TestValidProjectStates(t *testing.T) {
	got := ValidProjectStates()
	want := []string{ProjectStatePlanning, ProjectStateActive, ProjectStatePaused, ProjectStateShipped, ProjectStateArchived}
	if len(got) != len(want) {
		t.Fatalf("ValidProjectStates() returned %d items, want %d: got %v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("ValidProjectStates()[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestValidProjectUpdateKinds(t *testing.T) {
	got := ValidProjectUpdateKinds()
	want := []string{ProjectUpdateKindTimeline, ProjectUpdateKindBuildLog, ProjectUpdateKindChangelog}
	if len(got) != len(want) {
		t.Fatalf("ValidProjectUpdateKinds() returned %d items, want %d: got %v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("ValidProjectUpdateKinds()[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestValidContentRelationTypes(t *testing.T) {
	got := ValidContentRelationTypes()
	want := []string{ContentRelationTypePost, ContentRelationTypeProject}
	if len(got) != len(want) {
		t.Fatalf("ValidContentRelationTypes() returned %d items, want %d: got %v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("ValidContentRelationTypes()[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestValidContentRelationKinds(t *testing.T) {
	got := ValidContentRelationKinds()
	want := []string{ContentRelationKindRelated, ContentRelationKindInspiredBy, ContentRelationKindBuildsOn}
	if len(got) != len(want) {
		t.Fatalf("ValidContentRelationKinds() returned %d items, want %d: got %v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("ValidContentRelationKinds()[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestValidProjectShowcaseKinds(t *testing.T) {
	got := ValidProjectShowcaseKinds()
	want := []string{ProjectShowcaseKindImage, ProjectShowcaseKindLink, ProjectShowcaseKindRepo, ProjectShowcaseKindMetric, ProjectShowcaseKindVideo}
	if len(got) != len(want) {
		t.Fatalf("ValidProjectShowcaseKinds() returned %d items, want %d: got %v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("ValidProjectShowcaseKinds()[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}
