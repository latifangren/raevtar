package service

import (
	"path/filepath"
	"strings"
	"testing"

	"raevtar/internal/config"
	"raevtar/internal/model"
	"raevtar/internal/repo"
)

// adminTestServices creates a test Service with an admin user seeded
// via AdminPass config and SeedData.
func adminTestServices(t *testing.T) *testServices {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "raevtar_admin_test.db")
	cfg := &config.Config{
		DatabasePath: dbPath,
		Domain:       "raevtar.test",
		AdminUser:    "admin",
		AdminPass:    "test-pass",
		MediaDir:     filepath.Join(t.TempDir(), "uploads"),
	}

	db := repo.InitSQLite(cfg.DatabasePath)
	t.Cleanup(func() {
		_ = db.Close()
	})
	repo.AutoMigrate(db)

	repos := repo.New(db)
	svc := New(repos, cfg)
	if err := svc.SeedData(); err != nil {
		t.Fatalf("seed data: %v", err)
	}

	return &testServices{svc: svc, repos: repos}
}

func TestAdminServiceListUsersEmpty(t *testing.T) {
	state := newTestServices(t)

	users, err := state.svc.Admin.ListUsers()
	if err != nil {
		t.Fatalf("ListUsers: %v", err)
	}
	if len(users) != 0 {
		t.Fatalf("users len = %d, want 0", len(users))
	}
}

func TestAdminServiceSeedDataCreatesAdminUser(t *testing.T) {
	state := adminTestServices(t)

	users, err := state.svc.Admin.ListUsers()
	if err != nil {
		t.Fatalf("ListUsers: %v", err)
	}
	if len(users) != 1 {
		t.Fatalf("users len = %d, want 1", len(users))
	}
	if users[0].Username != "admin" {
		t.Fatalf("user username = %q, want admin", users[0].Username)
	}
	if users[0].Role != model.RoleOwner {
		t.Fatalf("user role = %q, want owner", users[0].Role)
	}
}

func TestAdminServiceAuthenticateValidCredentials(t *testing.T) {
	state := adminTestServices(t)

	user, err := state.svc.Admin.Authenticate("admin", "test-pass", "127.0.0.1")
	if err != nil {
		t.Fatalf("Authenticate: %v", err)
	}
	if user == nil {
		t.Fatalf("Authenticate returned nil user")
	}
	if user.Username != "admin" {
		t.Fatalf("user username = %q, want admin", user.Username)
	}
	if user.Role != model.RoleOwner {
		t.Fatalf("user role = %q, want owner", user.Role)
	}
}

func TestAdminServiceAuthenticateInvalidPassword(t *testing.T) {
	state := adminTestServices(t)

	_, err := state.svc.Admin.Authenticate("admin", "wrong-password", "127.0.0.1")
	if err == nil {
		t.Fatalf("Authenticate with wrong password should fail")
	}
}

func TestAdminServiceAuthenticateNonexistentUser(t *testing.T) {
	state := adminTestServices(t)

	_, err := state.svc.Admin.Authenticate("nonexistent", "test-pass", "127.0.0.1")
	if err == nil {
		t.Fatalf("Authenticate with nonexistent user should fail")
	}
}

func TestAdminServiceAuthenticateLogsLoginAudit(t *testing.T) {
	state := adminTestServices(t)

	_, err := state.svc.Admin.Authenticate("admin", "test-pass", "10.0.0.1")
	if err != nil {
		t.Fatalf("Authenticate: %v", err)
	}

	logs, err := state.svc.Admin.ListAuditLogs(10, 0)
	if err != nil {
		t.Fatalf("ListAuditLogs: %v", err)
	}

	var found bool
	for _, log := range logs {
		if log.Action == "LOGIN" && log.User == "admin" {
			found = true
			if log.IP != "10.0.0.1" {
				t.Fatalf("LOGIN audit IP = %q, want 10.0.0.1", log.IP)
			}
			break
		}
	}
	if !found {
		t.Fatalf("LOGIN audit entry not found")
	}
}

func TestAdminServiceCreateUser(t *testing.T) {
	state := adminTestServices(t)

	created, err := state.svc.Admin.CreateUser("owner", "admin", "newadmin", "newpass", "admin", "127.0.0.1")
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	if created == nil {
		t.Fatalf("CreateUser returned nil user")
	}
	if created.Username != "newadmin" {
		t.Fatalf("created username = %q, want newadmin", created.Username)
	}
	if created.Role != model.RoleAdmin {
		t.Fatalf("created role = %q, want admin", created.Role)
	}
	if created.DisplayName != "newadmin" {
		t.Fatalf("created display_name = %q, want newadmin", created.DisplayName)
	}

	users, err := state.svc.Admin.ListUsers()
	if err != nil {
		t.Fatalf("ListUsers: %v", err)
	}
	if len(users) != 2 {
		t.Fatalf("len(users) = %d, want 2", len(users))
	}
}

func TestAdminServiceCreateUserForbiddenRole(t *testing.T) {
	state := adminTestServices(t)

	// Create an admin user as owner
	_, err := state.svc.Admin.CreateUser("owner", "admin", "admin2", "admin2-pass", "admin", "127.0.0.1")
	if err != nil {
		t.Fatalf("create admin user: %v", err)
	}

	// Try to create an owner as admin — forbidden because admin cannot manage owner
	_, err = state.svc.Admin.CreateUser("admin", "admin2", "bad-actor", "bad-pass", "owner", "127.0.0.1")
	if err == nil {
		t.Fatalf("CreateUser with forbidden role should fail")
	}
	if !strings.Contains(err.Error(), "forbidden role") {
		t.Fatalf("err = %v, want containing 'forbidden role'", err)
	}
}

func TestAdminServiceDeleteUser(t *testing.T) {
	state := adminTestServices(t)

	created, err := state.svc.Admin.CreateUser("owner", "admin", "to-delete", "delete-pass", "operator", "127.0.0.1")
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}

	users, err := state.svc.Admin.ListUsers()
	if err != nil {
		t.Fatalf("ListUsers before delete: %v", err)
	}
	if len(users) != 2 {
		t.Fatalf("users before delete = %d, want 2", len(users))
	}

	err = state.svc.Admin.DeleteUser("owner", "admin", created.ID, "127.0.0.1")
	if err != nil {
		t.Fatalf("DeleteUser: %v", err)
	}

	users, err = state.svc.Admin.ListUsers()
	if err != nil {
		t.Fatalf("ListUsers after delete: %v", err)
	}
	if len(users) != 1 {
		t.Fatalf("users after delete = %d, want 1", len(users))
	}
	if users[0].ID == created.ID {
		t.Fatalf("deleted user still present in list")
	}
}

func TestAdminServiceDeleteUserSelf(t *testing.T) {
	state := adminTestServices(t)

	users, err := state.svc.Admin.ListUsers()
	if err != nil {
		t.Fatalf("ListUsers: %v", err)
	}
	if len(users) != 1 {
		t.Fatalf("users len = %d, want 1", len(users))
	}

	// Owner trying to delete the sole owner user — CanManage("owner","owner") is false
	err = state.svc.Admin.DeleteUser("owner", "admin", users[0].ID, "127.0.0.1")
	if err == nil {
		t.Fatalf("DeleteUser self should fail")
	}
}

func TestAdminServiceLogLogout(t *testing.T) {
	state := adminTestServices(t)

	err := state.svc.Admin.LogLogout("admin", "10.0.0.1")
	if err != nil {
		t.Fatalf("LogLogout: %v", err)
	}

	logs, err := state.svc.Admin.ListAuditLogs(10, 0)
	if err != nil {
		t.Fatalf("ListAuditLogs: %v", err)
	}

	var found bool
	for _, log := range logs {
		if log.Action == "LOGOUT" && log.User == "admin" {
			found = true
			if log.IP != "10.0.0.1" {
				t.Fatalf("LOGOUT audit IP = %q, want 10.0.0.1", log.IP)
			}
			break
		}
	}
	if !found {
		t.Fatalf("LOGOUT audit entry not found")
	}
}

func TestAdminServiceLogLogoutEmptyUsername(t *testing.T) {
	state := adminTestServices(t)

	err := state.svc.Admin.LogLogout("", "10.0.0.1")
	if err != nil {
		t.Fatalf("LogLogout with empty username: %v", err)
	}

	logs, err := state.svc.Admin.ListAuditLogs(10, 0)
	if err != nil {
		t.Fatalf("ListAuditLogs: %v", err)
	}
	if len(logs) != 0 {
		t.Fatalf("audit logs len = %d, want 0 (empty username should not log)", len(logs))
	}
}

func TestAdminServiceListServerAuditLogs(t *testing.T) {
	state := adminTestServices(t)

	// Insert server-related audit entries directly via repo.
	// Only entries with details matching "server id: <N>" patterns are returned.
	if err := state.repos.Audit.Insert("admin", "CREATE_SERVER", "created server: web01 (10.0.0.1:9100)", "127.0.0.1"); err != nil {
		t.Fatalf("insert create server audit: %v", err)
	}
	if err := state.repos.Audit.Insert("admin", "UPDATE_SERVER", "updated server id: 1 (web01 10.0.0.1:9100)", "127.0.0.1"); err != nil {
		t.Fatalf("insert update server audit: %v", err)
	}
	if err := state.repos.Audit.Insert("admin", "DELETE_SERVER", "deleted server id: 1", "127.0.0.1"); err != nil {
		t.Fatalf("insert delete server audit: %v", err)
	}
	// Unrelated log — should not match server ID filtering
	if err := state.repos.Audit.Insert("admin", "LOGIN", "login via admin panel", "10.0.0.1"); err != nil {
		t.Fatalf("insert unrelated audit: %v", err)
	}

	logs, err := state.svc.Admin.ListServerAuditLogs(1, 10)
	if err != nil {
		t.Fatalf("ListServerAuditLogs: %v", err)
	}
	if len(logs) != 2 {
		t.Fatalf("server audit logs len = %d, want 2 (UPDATE_SERVER + DELETE_SERVER)", len(logs))
	}
}
