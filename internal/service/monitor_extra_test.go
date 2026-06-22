package service

import (
	"errors"
	"strings"
	"testing"
	"time"

	"raevtar/internal/model"
)

// -- CreateServer ------------------------------------------------------------

func TestMonitorServiceCreateServer(t *testing.T) {
	state := newTestServices(t)

	server, err := state.svc.Monitor.CreateServer("test-node", "10.0.0.1", 9100, "test,local")
	if err != nil {
		t.Fatalf("CreateServer: %v", err)
	}
	if server == nil {
		t.Fatalf("CreateServer returned nil")
	}
	if server.ID == 0 {
		t.Fatalf("server ID = 0, want non-zero")
	}
	if server.Name != "test-node" {
		t.Fatalf("server name = %q, want test-node", server.Name)
	}
	if server.Host != "10.0.0.1" {
		t.Fatalf("server host = %q, want 10.0.0.1", server.Host)
	}
	if server.Port != 9100 {
		t.Fatalf("server port = %d, want 9100", server.Port)
	}
	if server.Tags != "test,local" {
		t.Fatalf("server tags = %q, want test,local", server.Tags)
	}
	if server.LastSeen != nil {
		t.Fatalf("new server last_seen should be nil, got %v", server.LastSeen)
	}
}

func TestMonitorServiceDuplicateName(t *testing.T) {
	state := newTestServices(t)

	_, err := state.svc.Monitor.CreateServer("dup-name", "10.0.0.1", 9100, "first")
	if err != nil {
		t.Fatalf("create first server: %v", err)
	}

	// Duplicate name is rejected due to servers.name UNIQUE constraint
	_, err = state.svc.Monitor.CreateServer("dup-name", "10.0.0.2", 9200, "second")
	if err == nil {
		t.Fatalf("expected duplicate server name to fail due to UNIQUE constraint")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "unique") {
		t.Fatalf("err = %v, want UNIQUE constraint error", err)
	}
}

// -- RecordMetrics -----------------------------------------------------------

func TestMonitorServiceRecordMetrics(t *testing.T) {
	state := newTestServices(t)

	server, err := state.svc.Monitor.CreateServer("metrics-node", "10.0.0.1", 9100, "test")
	if err != nil {
		t.Fatalf("create server: %v", err)
	}

	err = state.svc.Monitor.RecordMetrics(server.ID, model.ServerMetric{
		CPUPercent:    45.2,
		RAMUsedMB:     1024,
		RAMTotalMB:    4096,
		DiskUsedGB:    50.0,
		UptimeSeconds: 7200,
		Online:        true,
	})
	if err != nil {
		t.Fatalf("RecordMetrics: %v", err)
	}

	metrics, err := state.svc.Monitor.GetRecentMetrics(server.ID, 10)
	if err != nil {
		t.Fatalf("GetRecentMetrics: %v", err)
	}
	if len(metrics) != 1 {
		t.Fatalf("metrics len = %d, want 1", len(metrics))
	}
	if metrics[0].ServerID != server.ID {
		t.Fatalf("metric server_id = %d, want %d", metrics[0].ServerID, server.ID)
	}
	if metrics[0].CPUPercent != 45.2 {
		t.Fatalf("cpu_percent = %f, want 45.2", metrics[0].CPUPercent)
	}
	if !metrics[0].Online {
		t.Fatalf("metric online = false, want true")
	}
	if metrics[0].RecordedAt.Location() != time.UTC {
		t.Fatalf("recorded_at location = %s, want UTC", metrics[0].RecordedAt.Location())
	}

	// Verify last_seen updated on server
	updated, err := state.svc.Monitor.GetServer(server.ID)
	if err != nil {
		t.Fatalf("GetServer: %v", err)
	}
	if updated.LastSeen == nil {
		t.Fatalf("last_seen should be set after metrics")
	}
	if !updated.LastSeen.Truncate(time.Second).Equal(metrics[0].RecordedAt.Truncate(time.Second)) {
		t.Fatalf("last_seen = %s, want equal to recorded_at %s",
			updated.LastSeen.Format(time.RFC3339), metrics[0].RecordedAt.Format(time.RFC3339))
	}
}

func TestMonitorServiceListMetrics(t *testing.T) {
	state := newTestServices(t)

	server, err := state.svc.Monitor.CreateServer("multi-metrics", "10.0.0.1", 9100, "test")
	if err != nil {
		t.Fatalf("create server: %v", err)
	}

	// Record 3 metrics with ascending CPU values
	for i := 0; i < 3; i++ {
		err = state.svc.Monitor.RecordMetrics(server.ID, model.ServerMetric{
			CPUPercent:    float64(10 * (i + 1)),
			RAMUsedMB:     512,
			RAMTotalMB:    2048,
			DiskUsedGB:    25.0,
			UptimeSeconds: 3600,
			Online:        true,
		})
		if err != nil {
			t.Fatalf("RecordMetrics %d: %v", i, err)
		}
	}

	// Limit to 2 — should return newest 2 (DESC recorded_at)
	limited, err := state.svc.Monitor.GetRecentMetrics(server.ID, 2)
	if err != nil {
		t.Fatalf("GetRecentMetrics limit 2: %v", err)
	}
	if len(limited) != 2 {
		t.Fatalf("limited metrics len = %d, want 2", len(limited))
	}

	// Limit 10 returns all 3, newest first
	all, err := state.svc.Monitor.GetRecentMetrics(server.ID, 10)
	if err != nil {
		t.Fatalf("GetRecentMetrics limit 10: %v", err)
	}
	if len(all) != 3 {
		t.Fatalf("all metrics len = %d, want 3", len(all))
	}

	// Verify DESC ordering: first recorded should have highest CPU (last inserted)
	if all[0].CPUPercent < all[1].CPUPercent {
		t.Fatalf("metrics not in DESC order: %f < %f", all[0].CPUPercent, all[1].CPUPercent)
	}
}

// -- DeleteServer ------------------------------------------------------------

func TestMonitorServiceDeleteServer(t *testing.T) {
	state := newTestServices(t)

	server, err := state.svc.Monitor.CreateServer("delete-me", "10.0.0.1", 9100, "test")
	if err != nil {
		t.Fatalf("create server: %v", err)
	}

	// Verify it exists
	servers, err := state.svc.Monitor.ListServers()
	if err != nil {
		t.Fatalf("ListServers before delete: %v", err)
	}
	if len(servers) != 1 {
		t.Fatalf("servers before delete = %d, want 1", len(servers))
	}

	// Delete via repo (MonitorService does not expose DeleteServer)
	if err := state.repos.Server.Delete(server.ID); err != nil {
		t.Fatalf("delete server via repo: %v", err)
	}

	servers, err = state.svc.Monitor.ListServers()
	if err != nil {
		t.Fatalf("ListServers after delete: %v", err)
	}
	if len(servers) != 0 {
		t.Fatalf("servers after delete = %d, want 0", len(servers))
	}

	// GetServer should fail
	_, err = state.svc.Monitor.GetServer(server.ID)
	if err == nil {
		t.Fatalf("expected GetServer for deleted server to fail")
	}
}

// -- CreateServerWithAgentToken ----------------------------------------------

func TestMonitorServiceCreateServerWithAgentToken(t *testing.T) {
	state := newTestServices(t)

	server, token, err := state.svc.Monitor.CreateServerWithAgentToken("agent-node", "10.0.0.1", 9100, "agent")
	if err != nil {
		t.Fatalf("CreateServerWithAgentToken: %v", err)
	}
	if server == nil {
		t.Fatalf("CreateServerWithAgentToken returned nil server")
	}
	if server.ID == 0 {
		t.Fatalf("server ID = 0, want non-zero")
	}
	if server.Name != "agent-node" {
		t.Fatalf("server name = %q, want agent-node", server.Name)
	}
	if server.AgentTokenHash == "" {
		t.Fatalf("server agent_token_hash is empty")
	}
	if token == "" {
		t.Fatalf("expected non-empty agent token")
	}
	if !state.svc.Monitor.VerifyAgentToken(server.ID, token) {
		t.Fatalf("generated token should verify against server hash")
	}

	servers, err := state.svc.Monitor.ListServers()
	if err != nil {
		t.Fatalf("ListServers: %v", err)
	}
	if len(servers) != 1 {
		t.Fatalf("servers len = %d, want 1", len(servers))
	}
	if servers[0].Name != "agent-node" {
		t.Fatalf("listed server name = %q, want agent-node", servers[0].Name)
	}
}

// -- GetServerByName --------------------------------------------------------

func TestMonitorServiceGetServerByName(t *testing.T) {
	state := newTestServices(t)

	server, err := state.svc.Monitor.CreateServer("by-name", "10.0.0.1", 9100, "test")
	if err != nil {
		t.Fatalf("CreateServer: %v", err)
	}

	fetched, err := state.svc.Monitor.GetServerByName("by-name")
	if err != nil {
		t.Fatalf("GetServerByName: %v", err)
	}
	if fetched.ID != server.ID {
		t.Fatalf("fetched server id = %d, want %d", fetched.ID, server.ID)
	}
	if fetched.Host != "10.0.0.1" {
		t.Fatalf("fetched host = %q, want 10.0.0.1", fetched.Host)
	}
	if fetched.Port != 9100 {
		t.Fatalf("fetched port = %d, want 9100", fetched.Port)
	}
}

func TestMonitorServiceGetServerByNameNotFound(t *testing.T) {
	state := newTestServices(t)

	_, err := state.svc.Monitor.GetServerByName("nonexistent-server")
	if err == nil {
		t.Fatalf("GetServerByName for nonexistent server should fail")
	}
}

// -- UpdateServer ----------------------------------------------------------

func TestMonitorServiceUpdateServer(t *testing.T) {
	state := newTestServices(t)

	server, err := state.svc.Monitor.CreateServer("original-name", "10.0.0.1", 9100, "original")
	if err != nil {
		t.Fatalf("CreateServer: %v", err)
	}

	updated, err := state.svc.Monitor.UpdateServer(server.ID, "new-name", "10.0.0.2", 9200, "updated,tags")
	if err != nil {
		t.Fatalf("UpdateServer: %v", err)
	}
	if updated.Name != "new-name" {
		t.Fatalf("updated name = %q, want new-name", updated.Name)
	}
	if updated.Host != "10.0.0.2" {
		t.Fatalf("updated host = %q, want 10.0.0.2", updated.Host)
	}
	if updated.Port != 9200 {
		t.Fatalf("updated port = %d, want 9200", updated.Port)
	}
	if updated.Tags != "updated,tags" {
		t.Fatalf("updated tags = %q, want updated,tags", updated.Tags)
	}

	// Verify via GetServer
	fetched, err := state.svc.Monitor.GetServer(server.ID)
	if err != nil {
		t.Fatalf("GetServer after update: %v", err)
	}
	if fetched.Name != "new-name" {
		t.Fatalf("fetched name = %q, want new-name", fetched.Name)
	}
	if fetched.Host != "10.0.0.2" {
		t.Fatalf("fetched host = %q, want 10.0.0.2", fetched.Host)
	}
	if fetched.Port != 9200 {
		t.Fatalf("fetched port = %d, want 9200", fetched.Port)
	}
}

func TestMonitorServiceUpdateServerRejectsEmptyInput(t *testing.T) {
	state := newTestServices(t)

	server, err := state.svc.Monitor.CreateServer("validate", "10.0.0.1", 9100, "test")
	if err != nil {
		t.Fatalf("CreateServer: %v", err)
	}

	_, err = state.svc.Monitor.UpdateServer(server.ID, "", "10.0.0.2", 9200, "tags")
	if err == nil {
		t.Fatalf("expected error for empty name")
	}

	_, err = state.svc.Monitor.UpdateServer(server.ID, "valid-name", "", 9200, "tags")
	if err == nil {
		t.Fatalf("expected error for empty host")
	}

	_, err = state.svc.Monitor.UpdateServer(server.ID, "valid-name", "10.0.0.2", 0, "tags")
	if err == nil {
		t.Fatalf("expected error for zero port")
	}
}

func TestMonitorServiceUpdateServerNotFound(t *testing.T) {
	state := newTestServices(t)

	_, err := state.svc.Monitor.UpdateServer(9999, "ghost", "10.0.0.1", 9100, "test")
	if err == nil {
		t.Fatalf("expected error for updating nonexistent server")
	}
	if !errors.Is(err, ErrServerNotFound) {
		t.Fatalf("err = %v, want ErrServerNotFound", err)
	}
}

// -- PingServer ------------------------------------------------------------

func TestMonitorServicePingServer(t *testing.T) {
	state := newTestServices(t)

	metric, err := state.svc.Monitor.PingServer("10.0.0.1", 9100)
	if err != nil {
		t.Fatalf("PingServer: %v", err)
	}
	if metric == nil {
		t.Fatalf("PingServer returned nil metric")
	}
	if !metric.Online {
		t.Fatalf("ping metric online = false, want true")
	}
}

// -- GetRecentMetricsWithLimit -----------------------------------------------

func TestMonitorServiceGetRecentMetricsWithLimit(t *testing.T) {
	state := newTestServices(t)

	server, err := state.svc.Monitor.CreateServer("limit-test", "10.0.0.1", 9100, "test")
	if err != nil {
		t.Fatalf("create server: %v", err)
	}

	for i := 0; i < 3; i++ {
		err = state.svc.Monitor.RecordMetrics(server.ID, model.ServerMetric{
			CPUPercent:    float64(10 * (i + 1)),
			RAMUsedMB:     512,
			RAMTotalMB:    2048,
			DiskUsedGB:    25.0,
			UptimeSeconds: 3600,
			Online:        true,
		})
		if err != nil {
			t.Fatalf("RecordMetrics %d: %v", i, err)
		}
	}

	// With limit=1, should return only 1 (the most recent)
	limited, err := state.svc.Monitor.GetRecentMetrics(server.ID, 1)
	if err != nil {
		t.Fatalf("GetRecentMetrics limit 1: %v", err)
	}
	if len(limited) != 1 {
		t.Fatalf("limited metrics len = %d, want 1", len(limited))
	}

	// With limit <= 0 (default), should return up to 20
	defaulted, err := state.svc.Monitor.GetRecentMetrics(server.ID, 0)
	if err != nil {
		t.Fatalf("GetRecentMetrics default limit: %v", err)
	}
	if len(defaulted) != 3 {
		t.Fatalf("default limited metrics len = %d, want 3", len(defaulted))
	}
}

// -- RotateAgentToken --------------------------------------------------------

func TestMonitorServiceRotateAgentToken(t *testing.T) {
	state := newTestServices(t)

	server, err := state.svc.Monitor.CreateServer("rotate-srv", "10.0.0.1", 9100, "test")
	if err != nil {
		t.Fatalf("create server: %v", err)
	}

	// Rotate and verify new token
	token, err := state.svc.Monitor.RotateAgentToken(server.ID)
	if err != nil {
		t.Fatalf("RotateAgentToken: %v", err)
	}
	if token == "" {
		t.Fatalf("expected non-empty token")
	}
	if !state.svc.Monitor.VerifyAgentToken(server.ID, token) {
		t.Fatalf("newly rotated token should verify")
	}
}

// -- VerifyAgentTokenNotFound ------------------------------------------------

func TestMonitorServiceVerifyAgentTokenNotFound(t *testing.T) {
	state := newTestServices(t)

	if state.svc.Monitor.VerifyAgentToken(99999, "some-random-token") {
		t.Fatalf("expected false for non-existent server")
	}

	// Empty token should also return false
	if state.svc.Monitor.VerifyAgentToken(99999, "") {
		t.Fatalf("expected false for empty token on non-existent server")
	}
}
