package repo

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestAuditRepoInsertAndList(t *testing.T) {
	db := InitSQLite(filepath.Join(t.TempDir(), "test.db"))
	t.Cleanup(func() { _ = db.Close() })
	AutoMigrate(db)
	repos := New(db)

	if err := repos.Audit.Insert("alice", "login", "Logged in from 192.168.1.1", "192.168.1.1"); err != nil {
		t.Fatalf("Insert audit log 1: %v", err)
	}
	if err := repos.Audit.Insert("alice", "create_post", "Created post 'Hello'", "192.168.1.1"); err != nil {
		t.Fatalf("Insert audit log 2: %v", err)
	}
	if err := repos.Audit.Insert("bob", "logout", "Logged out", "10.0.0.1"); err != nil {
		t.Fatalf("Insert audit log 3: %v", err)
	}

	logs, err := repos.Audit.List(10, 0)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(logs) != 3 {
		t.Fatalf("expected 3 logs, got %d", len(logs))
	}

	// Ordered by created_at DESC: logout (bob), create_post (alice), login (alice)
	if logs[0].Action != "logout" {
		t.Errorf("first log action: got %q, want %q", logs[0].Action, "logout")
	}
	if logs[1].Action != "create_post" {
		t.Errorf("second log action: got %q, want %q", logs[1].Action, "create_post")
	}
	if logs[2].Action != "login" {
		t.Errorf("third log action: got %q, want %q", logs[2].Action, "login")
	}

	// Verify fields
	if logs[2].User != "alice" {
		t.Errorf("User: got %q, want %q", logs[2].User, "alice")
	}
	if logs[2].IP != "192.168.1.1" {
		t.Errorf("IP: got %q, want %q", logs[2].IP, "192.168.1.1")
	}
	if logs[2].ID == 0 {
		t.Error("expected ID to be set")
	}
	if logs[2].CreatedAt.IsZero() {
		t.Error("CreatedAt should not be zero")
	}
}

func TestAuditRepoListWithLimitOffset(t *testing.T) {
	db := InitSQLite(filepath.Join(t.TempDir(), "test.db"))
	t.Cleanup(func() { _ = db.Close() })
	AutoMigrate(db)
	repos := New(db)

	for i := 0; i < 5; i++ {
		action := string(rune('A' + i)) + "-action"
		if err := repos.Audit.Insert("user", action, "details", ""); err != nil {
			t.Fatalf("Insert log %d: %v", i, err)
		}
	}

	logs, err := repos.Audit.List(2, 0)
	if err != nil {
		t.Fatalf("List with limit=2: %v", err)
	}
	if len(logs) != 2 {
		t.Fatalf("expected 2 logs with limit=2, got %d", len(logs))
	}

	logs2, err := repos.Audit.List(2, 2)
	if err != nil {
		t.Fatalf("List with limit=2 offset=2: %v", err)
	}
	if len(logs2) != 2 {
		t.Fatalf("expected 2 logs with limit=2 offset=2, got %d", len(logs2))
	}

	// Pages should not overlap
	if logs[0].ID == logs2[0].ID {
		t.Error("limit/offset pages overlapped")
	}
}

func TestAuditRepoListWithDefaultLimit(t *testing.T) {
	db := InitSQLite(filepath.Join(t.TempDir(), "test.db"))
	t.Cleanup(func() { _ = db.Close() })
	AutoMigrate(db)
	repos := New(db)

	for i := 0; i < 60; i++ {
		action := "action-" + string(rune('0'+(i%10)))
		if err := repos.Audit.Insert("user", action, "details", ""); err != nil {
			t.Fatalf("Insert log %d: %v", i, err)
		}
	}

	logs, err := repos.Audit.List(0, 0)
	if err != nil {
		t.Fatalf("List with zero limit: %v", err)
	}
	if len(logs) != 50 {
		t.Errorf("expected 50 logs (default limit), got %d", len(logs))
	}
}

func TestAuditRepoCount(t *testing.T) {
	db := InitSQLite(filepath.Join(t.TempDir(), "test.db"))
	t.Cleanup(func() { _ = db.Close() })
	AutoMigrate(db)
	repos := New(db)

	for i := 0; i < 3; i++ {
		if err := repos.Audit.Insert("user", "action", "details", ""); err != nil {
			t.Fatalf("Insert log %d: %v", i, err)
		}
	}

	count, err := repos.Audit.Count()
	if err != nil {
		t.Fatalf("Count: %v", err)
	}
	if count != 3 {
		t.Errorf("expected count 3, got %d", count)
	}
}

func TestAuditRepoListServerLogs(t *testing.T) {
	db := InitSQLite(filepath.Join(t.TempDir(), "test.db"))
	t.Cleanup(func() { _ = db.Close() })
	AutoMigrate(db)
	repos := New(db)

	serverID := "srv-001"
	otherServerID := "srv-002"

	entries := []struct {
		user, action, details string
	}{
		{"system", "server_checkin", "server id: srv-001"},
		{"system", "server_checkin", "server id: srv-001 (cpu: 45%, mem: 60%)"},
		{"system", "server_checkin", "server id: srv-002"},
		{"system", "server_checkin", "server id: srv-001 (cpu: 50%, mem: 70%)"},
		{"admin", "reboot", "server id: srv-001 was rebooted by admin"},
	}
	for _, e := range entries {
		if err := repos.Audit.Insert(e.user, e.action, e.details, ""); err != nil {
			t.Fatalf("Insert log %q: %v", e.details, err)
		}
	}

	logs, err := repos.Audit.ListServerLogs(serverID, 10)
	if err != nil {
		t.Fatalf("ListServerLogs: %v", err)
	}
	if len(logs) != 3 {
		t.Fatalf("expected 4 logs for server %q, got %d", serverID, len(logs))
	}
	for _, l := range logs {
		if !strings.Contains(l.Details, serverID) {
			t.Errorf("log details %q does not contain server ID %q", l.Details, serverID)
		}
	}

	otherLogs, err := repos.Audit.ListServerLogs(otherServerID, 10)
	if err != nil {
		t.Fatalf("ListServerLogs for other server: %v", err)
	}
	if len(otherLogs) != 1 {
		t.Fatalf("expected 1 log for server %q, got %d", otherServerID, len(otherLogs))
	}
}

func TestAuditRepoEmptyList(t *testing.T) {
	db := InitSQLite(filepath.Join(t.TempDir(), "test.db"))
	t.Cleanup(func() { _ = db.Close() })
	AutoMigrate(db)
	repos := New(db)

	logs, err := repos.Audit.List(10, 0)
	if err != nil {
		t.Fatalf("List on empty table: %v", err)
	}
	if len(logs) != 0 {
		t.Errorf("expected 0 logs, got %d", len(logs))
	}
}

