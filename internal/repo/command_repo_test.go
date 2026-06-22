package repo

import (
	"path/filepath"
	"testing"
	"time"

	"raevtar/internal/model"
)

func TestCommandRepoInsertAndGetByID(t *testing.T) {
	db := InitSQLite(filepath.Join(t.TempDir(), "test.db"))
	t.Cleanup(func() { _ = db.Close() })
	AutoMigrate(db)
	repos := New(db)

	srv := &model.Server{Name: "cmd-server", Host: "10.0.0.1", Port: 22}
	if err := repos.Server.Create(srv); err != nil {
		t.Fatalf("create server: %v", err)
	}

	cmd := &model.ServerCommand{
		ServerID: srv.ID,
		Command:  "uptime",
		Payload:  "",
	}
	if err := repos.Command.Insert(cmd); err != nil {
		t.Fatalf("Insert: %v", err)
	}
	if cmd.ID == 0 {
		t.Fatal("expected cmd.ID to be set after Insert")
	}
	if cmd.Status != model.CommandPending {
		t.Errorf("Status: got %q, want %q", cmd.Status, model.CommandPending)
	}
	if cmd.QueuedAt.IsZero() {
		t.Error("QueuedAt should not be zero")
	}

	loaded, err := repos.Command.GetByID(cmd.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if loaded == nil {
		t.Fatal("GetByID returned nil")
	}
	if loaded.ID != cmd.ID {
		t.Errorf("ID: got %d, want %d", loaded.ID, cmd.ID)
	}
	if loaded.ServerID != srv.ID {
		t.Errorf("ServerID: got %d, want %d", loaded.ServerID, srv.ID)
	}
	if loaded.Command != "uptime" {
		t.Errorf("Command: got %q, want %q", loaded.Command, "uptime")
	}
	if loaded.Status != model.CommandPending {
		t.Errorf("Status: got %q, want %q", loaded.Status, model.CommandPending)
	}
}

func TestCommandRepoGetByIDNotFound(t *testing.T) {
	db := InitSQLite(filepath.Join(t.TempDir(), "test.db"))
	t.Cleanup(func() { _ = db.Close() })
	AutoMigrate(db)
	repos := New(db)

	loaded, err := repos.Command.GetByID(99999)
	if err != nil {
		t.Fatalf("GetByID for non-existent ID: %v", err)
	}
	if loaded != nil {
		t.Fatal("expected nil for non-existent command")
	}
}

func TestCommandRepoPendingByServerID(t *testing.T) {
	db := InitSQLite(filepath.Join(t.TempDir(), "test.db"))
	t.Cleanup(func() { _ = db.Close() })
	AutoMigrate(db)
	repos := New(db)

	srv := &model.Server{Name: "pending-srv", Host: "10.0.0.2", Port: 22}
	if err := repos.Server.Create(srv); err != nil {
		t.Fatalf("create server: %v", err)
	}

	cmds := []*model.ServerCommand{
		{ServerID: srv.ID, Command: "first"},
		{ServerID: srv.ID, Command: "second"},
	}
	for _, c := range cmds {
		if err := repos.Command.Insert(c); err != nil {
			t.Fatalf("Insert %q: %v", c.Command, err)
		}
	}

	// Mark one as running
	if err := repos.Command.MarkRunning(cmds[0].ID); err != nil {
		t.Fatalf("MarkRunning: %v", err)
	}

	pending, err := repos.Command.PendingByServerID(srv.ID)
	if err != nil {
		t.Fatalf("PendingByServerID: %v", err)
	}
	if len(pending) != 1 {
		t.Fatalf("expected 1 pending command, got %d", len(pending))
	}
	if pending[0].Command != "second" {
		t.Errorf("expected pending command 'second', got %q", pending[0].Command)
	}
}

func TestCommandRepoPendingByServerIDEmpty(t *testing.T) {
	db := InitSQLite(filepath.Join(t.TempDir(), "test.db"))
	t.Cleanup(func() { _ = db.Close() })
	AutoMigrate(db)
	repos := New(db)

	srv := &model.Server{Name: "empty-srv", Host: "10.0.0.3", Port: 22}
	if err := repos.Server.Create(srv); err != nil {
		t.Fatalf("create server: %v", err)
	}

	pending, err := repos.Command.PendingByServerID(srv.ID)
	if err != nil {
		t.Fatalf("PendingByServerID: %v", err)
	}
	if len(pending) != 0 {
		t.Errorf("expected empty pending list, got %d", len(pending))
	}
}

func TestCommandRepoMarkRunning(t *testing.T) {
	db := InitSQLite(filepath.Join(t.TempDir(), "test.db"))
	t.Cleanup(func() { _ = db.Close() })
	AutoMigrate(db)
	repos := New(db)

	srv := &model.Server{Name: "run-srv", Host: "10.0.0.4", Port: 22}
	if err := repos.Server.Create(srv); err != nil {
		t.Fatalf("create server: %v", err)
	}

	cmd := &model.ServerCommand{ServerID: srv.ID, Command: "df -h"}
	if err := repos.Command.Insert(cmd); err != nil {
		t.Fatalf("Insert: %v", err)
	}

	if err := repos.Command.MarkRunning(cmd.ID); err != nil {
		t.Fatalf("MarkRunning: %v", err)
	}

	loaded, err := repos.Command.GetByID(cmd.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if loaded.Status != model.CommandRunning {
		t.Errorf("Status: got %q, want %q", loaded.Status, model.CommandRunning)
	}
	if loaded.StartedAt == nil {
		t.Fatal("StartedAt should be set after MarkRunning")
	}
	if loaded.StartedAt.IsZero() {
		t.Error("StartedAt should not be zero")
	}
}

func TestCommandRepoMarkCompleted(t *testing.T) {
	db := InitSQLite(filepath.Join(t.TempDir(), "test.db"))
	t.Cleanup(func() { _ = db.Close() })
	AutoMigrate(db)
	repos := New(db)

	srv := &model.Server{Name: "comp-srv", Host: "10.0.0.5", Port: 22}
	if err := repos.Server.Create(srv); err != nil {
		t.Fatalf("create server: %v", err)
	}

	cmd := &model.ServerCommand{ServerID: srv.ID, Command: "ls -la"}
	if err := repos.Command.Insert(cmd); err != nil {
		t.Fatalf("Insert: %v", err)
	}

	if err := repos.Command.MarkCompleted(cmd.ID, "all files listed"); err != nil {
		t.Fatalf("MarkCompleted: %v", err)
	}

	loaded, err := repos.Command.GetByID(cmd.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if loaded.Status != model.CommandCompleted {
		t.Errorf("Status: got %q, want %q", loaded.Status, model.CommandCompleted)
	}
	if loaded.Result != "all files listed" {
		t.Errorf("Result: got %q, want %q", loaded.Result, "all files listed")
	}
	if loaded.CompletedAt == nil {
		t.Fatal("CompletedAt should be set after MarkCompleted")
	}
	if loaded.CompletedAt.IsZero() {
		t.Error("CompletedAt should not be zero")
	}
}

func TestCommandRepoMarkFailed(t *testing.T) {
	db := InitSQLite(filepath.Join(t.TempDir(), "test.db"))
	t.Cleanup(func() { _ = db.Close() })
	AutoMigrate(db)
	repos := New(db)

	srv := &model.Server{Name: "fail-srv", Host: "10.0.0.6", Port: 22}
	if err := repos.Server.Create(srv); err != nil {
		t.Fatalf("create server: %v", err)
	}

	cmd := &model.ServerCommand{ServerID: srv.ID, Command: "dangerous-command"}
	if err := repos.Command.Insert(cmd); err != nil {
		t.Fatalf("Insert: %v", err)
	}

	if err := repos.Command.MarkFailed(cmd.ID, "exit code 1"); err != nil {
		t.Fatalf("MarkFailed: %v", err)
	}

	loaded, err := repos.Command.GetByID(cmd.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if loaded.Status != model.CommandFailed {
		t.Errorf("Status: got %q, want %q", loaded.Status, model.CommandFailed)
	}
	if loaded.Result != "exit code 1" {
		t.Errorf("Result: got %q, want %q", loaded.Result, "exit code 1")
	}
	if loaded.CompletedAt == nil {
		t.Fatal("CompletedAt should be set after MarkFailed")
	}
}

func TestCommandRepoListByServerID(t *testing.T) {
	db := InitSQLite(filepath.Join(t.TempDir(), "test.db"))
	t.Cleanup(func() { _ = db.Close() })
	AutoMigrate(db)
	repos := New(db)

	srv := &model.Server{Name: "list-srv", Host: "10.0.0.7", Port: 22}
	if err := repos.Server.Create(srv); err != nil {
		t.Fatalf("create server: %v", err)
	}

	// Insert 3 commands, each with a slight delay to ensure different queued_at
	for i := 0; i < 3; i++ {
		cmd := &model.ServerCommand{
			ServerID: srv.ID,
			Command:  "cmd-" + string(rune('0'+i)),
		}
		if err := repos.Command.Insert(cmd); err != nil {
			t.Fatalf("Insert cmd-%d: %v", i, err)
		}
		time.Sleep(2 * time.Millisecond)
	}

	// List with limit 2
	cmds, err := repos.Command.ListByServerID(srv.ID, 2)
	if err != nil {
		t.Fatalf("ListByServerID: %v", err)
	}
	if len(cmds) != 2 {
		t.Fatalf("expected 2 commands, got %d", len(cmds))
	}

	// Ordered by queued_at DESC: newest first
	if cmds[0].Command != "cmd-2" {
		t.Errorf("expected newest command 'cmd-2', got %q", cmds[0].Command)
	}
	if cmds[1].Command != "cmd-1" {
		t.Errorf("expected second newest 'cmd-1', got %q", cmds[1].Command)
	}
}

func TestCommandRepoListByServerIDNoCommands(t *testing.T) {
	db := InitSQLite(filepath.Join(t.TempDir(), "test.db"))
	t.Cleanup(func() { _ = db.Close() })
	AutoMigrate(db)
	repos := New(db)

	srv := &model.Server{Name: "no-cmd-srv", Host: "10.0.0.8", Port: 22}
	if err := repos.Server.Create(srv); err != nil {
		t.Fatalf("create server: %v", err)
	}

	cmds, err := repos.Command.ListByServerID(srv.ID, 10)
	if err != nil {
		t.Fatalf("ListByServerID: %v", err)
	}
	if len(cmds) != 0 {
		t.Errorf("expected empty list, got %d", len(cmds))
	}
}
