package service

import (
	"testing"
	"time"

	"raevtar/internal/model"
)

func TestCommandQueueServiceQueueCommand(t *testing.T) {
	state := newTestServices(t)

	srv, err := state.svc.Monitor.CreateServer("test-node", "10.0.0.1", 22, "test")
	if err != nil {
		t.Fatalf("create server: %v", err)
	}

	cmd, err := state.svc.CommandQ.QueueCommand(srv.ID, "uptime", "")
	if err != nil {
		t.Fatalf("queue command: %v", err)
	}
	if cmd.ID == 0 {
		t.Fatalf("command id = 0, want positive")
	}
	if cmd.ServerID != srv.ID {
		t.Fatalf("server_id = %d, want %d", cmd.ServerID, srv.ID)
	}
	if cmd.Command != "uptime" {
		t.Fatalf("command = %q, want uptime", cmd.Command)
	}
	if cmd.Status != model.CommandPending {
		t.Fatalf("status = %q, want pending", cmd.Status)
	}
	if cmd.QueuedAt.IsZero() {
		t.Fatalf("queued_at should be set")
	}
}

func TestCommandQueueServiceQueueCommandInvalidServer(t *testing.T) {
	state := newTestServices(t)

	cmd, err := state.svc.CommandQ.QueueCommand(99999, "invalid-server-test", "")
	if err != nil {
		t.Fatalf("queue command with non-existent server: %v", err)
	}
	if cmd.ID == 0 {
		t.Fatalf("command id = 0, want positive")
	}
	if cmd.ServerID != 99999 {
		t.Fatalf("server_id = %d, want 99999", cmd.ServerID)
	}
}

func TestCommandQueueServicePendingCommands(t *testing.T) {
	state := newTestServices(t)

	srv, err := state.svc.Monitor.CreateServer("pending-node", "10.0.0.2", 22, "test")
	if err != nil {
		t.Fatalf("create server: %v", err)
	}

	cmd1, err := state.svc.CommandQ.QueueCommand(srv.ID, "df -h", "")
	if err != nil {
		t.Fatalf("queue first cmd: %v", err)
	}
	cmd2, err := state.svc.CommandQ.QueueCommand(srv.ID, "free -m", "")
	if err != nil {
		t.Fatalf("queue second cmd: %v", err)
	}

	pending, err := state.svc.CommandQ.PendingCommands(srv.ID)
	if err != nil {
		t.Fatalf("pending commands: %v", err)
	}
	if len(pending) != 2 {
		t.Fatalf("pending len = %d, want 2", len(pending))
	}

	// Mark one as running — it should no longer appear in pending.
	if err := state.svc.CommandQ.TakeAndRun(cmd1.ID); err != nil {
		t.Fatalf("take and run first cmd: %v", err)
	}

	pendingAfterRun, err := state.svc.CommandQ.PendingCommands(srv.ID)
	if err != nil {
		t.Fatalf("pending after run: %v", err)
	}
	if len(pendingAfterRun) != 1 {
		t.Fatalf("pending after run len = %d, want 1", len(pendingAfterRun))
	}
	if pendingAfterRun[0].ID != cmd2.ID {
		t.Fatalf("remaining pending id = %d, want %d", pendingAfterRun[0].ID, cmd2.ID)
	}
}

func TestCommandQueueServicePendingCommandsEmpty(t *testing.T) {
	state := newTestServices(t)

	srv, err := state.svc.Monitor.CreateServer("empty-node", "10.0.0.3", 22, "test")
	if err != nil {
		t.Fatalf("create server: %v", err)
	}

	pending, err := state.svc.CommandQ.PendingCommands(srv.ID)
	if err != nil {
		t.Fatalf("pending commands: %v", err)
	}
	if len(pending) != 0 {
		t.Fatalf("pending len = %d, want 0", len(pending))
	}
}

func TestCommandQueueServiceMarkRunning(t *testing.T) {
	state := newTestServices(t)

	srv, err := state.svc.Monitor.CreateServer("run-node", "10.0.0.4", 22, "test")
	if err != nil {
		t.Fatalf("create server: %v", err)
	}
	cmd, err := state.svc.CommandQ.QueueCommand(srv.ID, "systemctl status", "")
	if err != nil {
		t.Fatalf("queue cmd: %v", err)
	}

	before := time.Now().UTC()
	if err := state.svc.CommandQ.TakeAndRun(cmd.ID); err != nil {
		t.Fatalf("take and run: %v", err)
	}
	after := time.Now().UTC()

	// Verify it is no longer pending.
	pending, err := state.svc.CommandQ.PendingCommands(srv.ID)
	if err != nil {
		t.Fatalf("pending commands: %v", err)
	}
	if len(pending) != 0 {
		t.Fatalf("pending after run len = %d, want 0", len(pending))
	}

	// Verify via repo that status and started_at are set.
	reloaded, err := state.repos.Command.GetByID(cmd.ID)
	if err != nil {
		t.Fatalf("get command by id: %v", err)
	}
	if reloaded.Status != model.CommandRunning {
		t.Fatalf("status = %q, want running", reloaded.Status)
	}
	if reloaded.StartedAt == nil {
		t.Fatalf("started_at should be set")
	}
	if reloaded.StartedAt.Before(before) || reloaded.StartedAt.After(after) {
		t.Fatalf("started_at = %s, want between %s and %s", reloaded.StartedAt.Format(time.RFC3339), before.Format(time.RFC3339), after.Format(time.RFC3339))
	}
}

func TestCommandQueueServiceMarkCompleted(t *testing.T) {
	state := newTestServices(t)

	srv, err := state.svc.Monitor.CreateServer("complete-node", "10.0.0.5", 22, "test")
	if err != nil {
		t.Fatalf("create server: %v", err)
	}
	cmd, err := state.svc.CommandQ.QueueCommand(srv.ID, "uptime", "")
	if err != nil {
		t.Fatalf("queue cmd: %v", err)
	}
	if err := state.svc.CommandQ.TakeAndRun(cmd.ID); err != nil {
		t.Fatalf("take and run: %v", err)
	}

	before := time.Now().UTC()
	if err := state.svc.CommandQ.CompleteCommand(cmd.ID, "OK: uptime 30 days"); err != nil {
		t.Fatalf("complete command: %v", err)
	}
	after := time.Now().UTC()

	reloaded, err := state.repos.Command.GetByID(cmd.ID)
	if err != nil {
		t.Fatalf("get command: %v", err)
	}
	if reloaded.Status != model.CommandCompleted {
		t.Fatalf("status = %q, want completed", reloaded.Status)
	}
	if reloaded.Result != "OK: uptime 30 days" {
		t.Fatalf("result = %q, want OK: uptime 30 days", reloaded.Result)
	}
	if reloaded.CompletedAt == nil {
		t.Fatalf("completed_at should be set")
	}
	if reloaded.CompletedAt.Before(before) || reloaded.CompletedAt.After(after) {
		t.Fatalf("completed_at = %s, want between %s and %s", reloaded.CompletedAt.Format(time.RFC3339), before.Format(time.RFC3339), after.Format(time.RFC3339))
	}
}

func TestCommandQueueServiceMarkFailed(t *testing.T) {
	state := newTestServices(t)

	srv, err := state.svc.Monitor.CreateServer("fail-node", "10.0.0.6", 22, "test")
	if err != nil {
		t.Fatalf("create server: %v", err)
	}
	cmd, err := state.svc.CommandQ.QueueCommand(srv.ID, "unreliable-command", "")
	if err != nil {
		t.Fatalf("queue cmd: %v", err)
	}
	if err := state.svc.CommandQ.TakeAndRun(cmd.ID); err != nil {
		t.Fatalf("take and run: %v", err)
	}

	before := time.Now().UTC()
	if err := state.svc.CommandQ.FailCommand(cmd.ID, "exit code 1"); err != nil {
		t.Fatalf("fail command: %v", err)
	}
	after := time.Now().UTC()

	reloaded, err := state.repos.Command.GetByID(cmd.ID)
	if err != nil {
		t.Fatalf("get command: %v", err)
	}
	if reloaded.Status != model.CommandFailed {
		t.Fatalf("status = %q, want failed", reloaded.Status)
	}
	if reloaded.Result != "exit code 1" {
		t.Fatalf("result = %q, want exit code 1", reloaded.Result)
	}
	if reloaded.CompletedAt == nil {
		t.Fatalf("completed_at should be set")
	}
	if reloaded.CompletedAt.Before(before) || reloaded.CompletedAt.After(after) {
		t.Fatalf("completed_at = %s, want between %s and %s", reloaded.CompletedAt.Format(time.RFC3339), before.Format(time.RFC3339), after.Format(time.RFC3339))
	}
}

func TestCommandQueueServiceListByServerID(t *testing.T) {
	state := newTestServices(t)

	srv1, err := state.svc.Monitor.CreateServer("list-node-1", "10.0.0.7", 22, "test")
	if err != nil {
		t.Fatalf("create server1: %v", err)
	}
	srv2, err := state.svc.Monitor.CreateServer("list-node-2", "10.0.0.8", 22, "test")
	if err != nil {
		t.Fatalf("create server2: %v", err)
	}

	for i := 0; i < 3; i++ {
		_, err := state.svc.CommandQ.QueueCommand(srv1.ID, "cmd-srv1-"+string(rune('a'+i)), "")
		if err != nil {
			t.Fatalf("queue srv1 cmd %d: %v", i, err)
		}
	}

	for i := 0; i < 2; i++ {
		_, err := state.svc.CommandQ.QueueCommand(srv2.ID, "cmd-srv2-"+string(rune('a'+i)), "")
		if err != nil {
			t.Fatalf("queue srv2 cmd %d: %v", i, err)
		}
	}

	// ListByServerID uses CommandHistory to get per-server commands.
	history1, err := state.svc.CommandQ.CommandHistory(srv1.ID, 10)
	if err != nil {
		t.Fatalf("command history srv1: %v", err)
	}
	if len(history1) != 3 {
		t.Fatalf("srv1 history len = %d, want 3", len(history1))
	}

	// Respect limit.
	limited, err := state.svc.CommandQ.CommandHistory(srv1.ID, 2)
	if err != nil {
		t.Fatalf("limited history srv1: %v", err)
	}
	if len(limited) != 2 {
		t.Fatalf("limited len = %d, want 2", len(limited))
	}

	history2, err := state.svc.CommandQ.CommandHistory(srv2.ID, 10)
	if err != nil {
		t.Fatalf("command history srv2: %v", err)
	}
	if len(history2) != 2 {
		t.Fatalf("srv2 history len = %d, want 2", len(history2))
	}
}
