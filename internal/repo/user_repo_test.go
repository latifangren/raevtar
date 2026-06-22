package repo

import (
	"database/sql"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"raevtar/internal/model"
)

func TestUserRepoCreateAndGetByID(t *testing.T) {
	db := InitSQLite(filepath.Join(t.TempDir(), "test.db"))
	t.Cleanup(func() { _ = db.Close() })
	AutoMigrate(db)
	repos := New(db)

	hash, err := HashPassword("secure-password-123")
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}

	u, err := repos.User.Create("testuser", hash, model.RoleAdmin, "Test User")
	if err != nil {
		t.Fatalf("Create user: %v", err)
	}
	if u.ID == 0 {
		t.Fatal("expected user.ID to be set after Create")
	}
	if u.Username != "testuser" {
		t.Errorf("Username: got %q, want %q", u.Username, "testuser")
	}
	if u.Role != model.RoleAdmin {
		t.Errorf("Role: got %q, want %q", u.Role, model.RoleAdmin)
	}
	if u.DisplayName != "Test User" {
		t.Errorf("DisplayName: got %q, want %q", u.DisplayName, "Test User")
	}
	if u.PasswordHash == "" {
		t.Error("PasswordHash should not be empty")
	}
	if u.CreatedAt.IsZero() {
		t.Error("CreatedAt should not be zero")
	}
	if u.UpdatedAt.IsZero() {
		t.Error("UpdatedAt should not be zero")
	}

	loaded, err := repos.User.GetByID(u.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if loaded == nil {
		t.Fatal("GetByID returned nil user")
	}
	if loaded.Username != "testuser" {
		t.Errorf("Username from DB: got %q, want %q", loaded.Username, "testuser")
	}
	if !CheckPassword("secure-password-123", loaded.PasswordHash) {
		t.Error("stored password hash should match original password")
	}
}

func TestUserRepoGetByUsername(t *testing.T) {
	db := InitSQLite(filepath.Join(t.TempDir(), "test.db"))
	t.Cleanup(func() { _ = db.Close() })
	AutoMigrate(db)
	repos := New(db)

	hash, err := HashPassword("pass")
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}

	if _, err := repos.User.Create("alice", hash, model.RoleOperator, "Alice"); err != nil {
		t.Fatalf("Create user: %v", err)
	}

	loaded, err := repos.User.GetByUsername("alice")
	if err != nil {
		t.Fatalf("GetByUsername: %v", err)
	}
	if loaded == nil {
		t.Fatal("GetByUsername returned nil")
	}
	if loaded.Username != "alice" {
		t.Errorf("Username: got %q, want %q", loaded.Username, "alice")
	}
	if loaded.DisplayName != "Alice" {
		t.Errorf("DisplayName: got %q, want %q", loaded.DisplayName, "Alice")
	}
}

func TestUserRepoGetByUsernameNotFound(t *testing.T) {
	db := InitSQLite(filepath.Join(t.TempDir(), "test.db"))
	t.Cleanup(func() { _ = db.Close() })
	AutoMigrate(db)
	repos := New(db)

	loaded, err := repos.User.GetByUsername("nonexistent")
	if err == nil {
		t.Fatal("expected error for non-existent username")
	}
	if err != sql.ErrNoRows {
		t.Fatalf("expected sql.ErrNoRows, got %v", err)
	}
	if loaded != nil {
		t.Fatal("expected nil user for non-existent username")
	}
}

func TestUserRepoGetByIDNotFound(t *testing.T) {
	db := InitSQLite(filepath.Join(t.TempDir(), "test.db"))
	t.Cleanup(func() { _ = db.Close() })
	AutoMigrate(db)
	repos := New(db)

	loaded, err := repos.User.GetByID(99999)
	if err == nil {
		t.Fatal("expected error for non-existent ID")
	}
	if err != sql.ErrNoRows {
		t.Fatalf("expected sql.ErrNoRows, got %v", err)
	}
	if loaded != nil {
		t.Fatal("expected nil user for non-existent ID")
	}
}

func TestUserRepoUpdatePassword(t *testing.T) {
	db := InitSQLite(filepath.Join(t.TempDir(), "test.db"))
	t.Cleanup(func() { _ = db.Close() })
	AutoMigrate(db)
	repos := New(db)

	hash, err := HashPassword("old-password")
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}

	u, err := repos.User.Create("updatepass", hash, model.RoleReadonly, "Update Pass")
	if err != nil {
		t.Fatalf("Create user: %v", err)
	}

	newHash, err := HashPassword("new-password")
	if err != nil {
		t.Fatalf("HashPassword (new): %v", err)
	}

	if err := repos.User.UpdatePassword(u.ID, newHash); err != nil {
		t.Fatalf("UpdatePassword: %v", err)
	}

	loaded, err := repos.User.GetByID(u.ID)
	if err != nil {
		t.Fatalf("GetByID after update: %v", err)
	}
	if CheckPassword("old-password", loaded.PasswordHash) {
		t.Error("old password should no longer match after update")
	}
	if !CheckPassword("new-password", loaded.PasswordHash) {
		t.Error("new password should match after update")
	}
}

func TestUserRepoUpdateRole(t *testing.T) {
	db := InitSQLite(filepath.Join(t.TempDir(), "test.db"))
	t.Cleanup(func() { _ = db.Close() })
	AutoMigrate(db)
	repos := New(db)

	hash, err := HashPassword("pass")
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}

	u, err := repos.User.Create("changerole", hash, model.RoleReadonly, "Change Role")
	if err != nil {
		t.Fatalf("Create user: %v", err)
	}

	if err := repos.User.UpdateRole(u.ID, model.RoleAdmin); err != nil {
		t.Fatalf("UpdateRole: %v", err)
	}

	loaded, err := repos.User.GetByID(u.ID)
	if err != nil {
		t.Fatalf("GetByID after role update: %v", err)
	}
	if loaded.Role != model.RoleAdmin {
		t.Errorf("Role: got %q, want %q", loaded.Role, model.RoleAdmin)
	}
}

func TestUserRepoList(t *testing.T) {
	db := InitSQLite(filepath.Join(t.TempDir(), "test.db"))
	t.Cleanup(func() { _ = db.Close() })
	AutoMigrate(db)
	repos := New(db)

	hash, err := HashPassword("pass")
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}

	users := []struct {
		username, role, displayName string
	}{
		{"zara", model.RoleReadonly, "Zara"},
		{"bob", model.RoleOperator, "Bob"},
		{"alice", model.RoleAdmin, "Alice"},
	}
	for _, u := range users {
		if _, err := repos.User.Create(u.username, hash, u.role, u.displayName); err != nil {
			t.Fatalf("Create user %q: %v", u.username, err)
		}
		time.Sleep(2 * time.Millisecond)
	}

	result, err := repos.User.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(result) != 3 {
		t.Fatalf("expected 3 users, got %d", len(result))
	}

	expectedOrder := []string{"zara", "bob", "alice"}
	for i, username := range expectedOrder {
		if result[i].Username != username {
			t.Errorf("position %d: expected username %q, got %q", i, username, result[i].Username)
		}
	}
}

func TestUserRepoDelete(t *testing.T) {
	db := InitSQLite(filepath.Join(t.TempDir(), "test.db"))
	t.Cleanup(func() { _ = db.Close() })
	AutoMigrate(db)
	repos := New(db)

	hash, err := HashPassword("pass")
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}

	u, err := repos.User.Create("deleteme", hash, model.RoleReadonly, "Delete Me")
	if err != nil {
		t.Fatalf("Create user: %v", err)
	}

	if err := repos.User.Delete(u.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	loaded, err := repos.User.GetByID(u.ID)
	if err == nil {
		t.Fatal("expected error after deleting user")
	}
	if err != sql.ErrNoRows {
		t.Fatalf("expected sql.ErrNoRows, got %v", err)
	}
	if loaded != nil {
		t.Fatal("expected nil user after deletion")
	}
}

func TestUserRepoCount(t *testing.T) {
	db := InitSQLite(filepath.Join(t.TempDir(), "test.db"))
	t.Cleanup(func() { _ = db.Close() })
	AutoMigrate(db)
	repos := New(db)

	hash, err := HashPassword("pass")
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}

	for i := 0; i < 4; i++ {
		username := "countuser" + string(rune('0'+i))
		if _, err := repos.User.Create(username, hash, model.RoleReadonly, "Count User"); err != nil {
			t.Fatalf("Create user %q: %v", username, err)
		}
	}

	count, err := repos.User.Count()
	if err != nil {
		t.Fatalf("Count: %v", err)
	}
	if count != 4 {
		t.Errorf("expected count 4, got %d", count)
	}
}

func TestHashPasswordAndCheckPassword(t *testing.T) {
	hash, err := HashPassword("my-secret")
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}
	if hash == "" {
		t.Fatal("HashPassword returned empty string")
	}
	if !strings.HasPrefix(hash, "$2a$") && !strings.HasPrefix(hash, "$2b$") {
		t.Errorf("expected bcrypt hash prefix, got %q", hash[:4])
	}

	if !CheckPassword("my-secret", hash) {
		t.Error("CheckPassword should return true for correct password")
	}
	if CheckPassword("wrong-password", hash) {
		t.Error("CheckPassword should return false for wrong password")
	}
	if CheckPassword("", hash) {
		t.Error("CheckPassword should return false for empty password")
	}
}
