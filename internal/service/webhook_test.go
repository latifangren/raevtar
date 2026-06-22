package service

import (
	"testing"
	"time"

	"raevtar/internal/model"
)

func TestWebhookServiceCreateAndListConfigs(t *testing.T) {
	state := newTestServices(t)

	cfg, err := state.svc.Webhook.CreateConfig("test-webhook", "https://hooks.example.com/alert", "secret123", true)
	if err != nil {
		t.Fatalf("create webhook config: %v", err)
	}
	if cfg.ID == 0 {
		t.Fatalf("config id = 0, want persisted id")
	}
	if cfg.Name != "test-webhook" {
		t.Fatalf("config name = %q, want %q", cfg.Name, "test-webhook")
	}
	if cfg.URL != "https://hooks.example.com/alert" {
		t.Fatalf("config url = %q, want %q", cfg.URL, "https://hooks.example.com/alert")
	}
	if !cfg.Enabled {
		t.Fatalf("config enabled = false, want true")
	}

	cfgs, err := state.svc.Webhook.ListConfigs()
	if err != nil {
		t.Fatalf("list webhook configs: %v", err)
	}
	if len(cfgs) != 1 {
		t.Fatalf("configs len = %d, want 1", len(cfgs))
	}
	if cfgs[0].ID != cfg.ID {
		t.Fatalf("listed config id = %d, want %d", cfgs[0].ID, cfg.ID)
	}
	if cfgs[0].Name != "test-webhook" {
		t.Fatalf("listed config name = %q, want %q", cfgs[0].Name, "test-webhook")
	}
	if cfgs[0].URL != "https://hooks.example.com/alert" {
		t.Fatalf("listed config url = %q, want %q", cfgs[0].URL, "https://hooks.example.com/alert")
	}
}

func TestWebhookServiceGetConfig(t *testing.T) {
	state := newTestServices(t)

	created, err := state.svc.Webhook.CreateConfig("get-test", "https://example.com/webhook", "s3cr3t", true)
	if err != nil {
		t.Fatalf("create webhook config: %v", err)
	}

	cfg, err := state.svc.Webhook.GetConfig(created.ID)
	if err != nil {
		t.Fatalf("get webhook config: %v", err)
	}
	if cfg.ID != created.ID {
		t.Fatalf("config id = %d, want %d", cfg.ID, created.ID)
	}
	if cfg.Name != "get-test" {
		t.Fatalf("config name = %q, want %q", cfg.Name, "get-test")
	}
	if cfg.URL != "https://example.com/webhook" {
		t.Fatalf("config url = %q, want %q", cfg.URL, "https://example.com/webhook")
	}
	if cfg.Secret != "s3cr3t" {
		t.Fatalf("config secret = %q, want %q", cfg.Secret, "s3cr3t")
	}
	if !cfg.Enabled {
		t.Fatalf("config enabled = false, want true")
	}
	if cfg.CreatedAt.IsZero() {
		t.Fatalf("config created_at is zero")
	}
}

func TestWebhookServiceUpdateConfig(t *testing.T) {
	state := newTestServices(t)

	cfg, err := state.svc.Webhook.CreateConfig("original", "https://original.example.com/hook", "orig-secret", false)
	if err != nil {
		t.Fatalf("create webhook config: %v", err)
	}

	cfg.Name = "updated-name"
	cfg.URL = "https://updated.example.com/hook"
	cfg.Secret = "new-secret"
	cfg.Enabled = true

	if err := state.svc.Webhook.UpdateConfig(cfg); err != nil {
		t.Fatalf("update webhook config: %v", err)
	}

	updated, err := state.svc.Webhook.GetConfig(cfg.ID)
	if err != nil {
		t.Fatalf("get updated config: %v", err)
	}
	if updated.Name != "updated-name" {
		t.Fatalf("updated name = %q, want %q", updated.Name, "updated-name")
	}
	if updated.URL != "https://updated.example.com/hook" {
		t.Fatalf("updated url = %q, want %q", updated.URL, "https://updated.example.com/hook")
	}
	if updated.Secret != "new-secret" {
		t.Fatalf("updated secret = %q, want %q", updated.Secret, "new-secret")
	}
	if !updated.Enabled {
		t.Fatalf("updated enabled = false, want true")
	}
}

func TestWebhookServiceDeleteConfig(t *testing.T) {
	state := newTestServices(t)

	cfg, err := state.svc.Webhook.CreateConfig("delete-me", "https://example.com/delete", "del-secret", true)
	if err != nil {
		t.Fatalf("create webhook config: %v", err)
	}

	if err := state.svc.Webhook.DeleteConfig(cfg.ID); err != nil {
		t.Fatalf("delete webhook config: %v", err)
	}

	cfgs, err := state.svc.Webhook.ListConfigs()
	if err != nil {
		t.Fatalf("list configs after delete: %v", err)
	}
	if len(cfgs) != 0 {
		t.Fatalf("configs len after delete = %d, want 0", len(cfgs))
	}

	// GetConfig should also fail
	if _, err := state.svc.Webhook.GetConfig(cfg.ID); err == nil {
		t.Fatalf("expected error getting deleted config")
	}
}

func TestWebhookServiceListEventsEmpty(t *testing.T) {
	state := newTestServices(t)

	events, err := state.svc.Webhook.ListEvents(100)
	if err != nil {
		t.Fatalf("list events: %v", err)
	}
	if len(events) != 0 {
		t.Fatalf("events len = %d, want 0", len(events))
	}
}

func TestWebhookServiceEvaluateAndFireBelowThreshold(t *testing.T) {
	state := newTestServices(t)

	_, err := state.svc.Webhook.CreateConfig("below-test", "http://127.0.0.1:1/webhook", "test", true)
	if err != nil {
		t.Fatalf("create webhook config: %v", err)
	}

	metric := model.ServerMetric{
		CPUPercent:  50.0,
		RAMUsedMB:   512,
		RAMTotalMB:  2048,
		DiskUsedGB:  30,
		DiskTotalGB: 500,
		RecordedAt:  time.Now(),
	}
	state.svc.Webhook.EvaluateAndFire(1, metric)

	// Allow async goroutines to complete
	time.Sleep(200 * time.Millisecond)

	events, err := state.svc.Webhook.ListEvents(100)
	if err != nil {
		t.Fatalf("list events: %v", err)
	}
	if len(events) != 0 {
		t.Fatalf("events len = %d, want 0 (below threshold)", len(events))
	}
}

func TestWebhookServiceEvaluateAndFireCPUThreshold(t *testing.T) {
	state := newTestServices(t)

	_, err := state.svc.Webhook.CreateConfig("cpu-test", "http://127.0.0.1:1/webhook", "test", true)
	if err != nil {
		t.Fatalf("create webhook config: %v", err)
	}

	metric := model.ServerMetric{
		CPUPercent: 95.0,
		// RAMTotalMB=0 and DiskTotalGB=0 skip those checks
		RecordedAt: time.Now(),
	}
	state.svc.Webhook.EvaluateAndFire(1, metric)

	// Allow async goroutines to complete
	time.Sleep(200 * time.Millisecond)

	events, err := state.svc.Webhook.ListEvents(100)
	if err != nil {
		t.Fatalf("list events: %v", err)
	}
	if len(events) == 0 {
		t.Fatalf("expected at least one event for cpu threshold exceeded")
	}

	// The first event (most recent, ordered by fired_at DESC) should be cpu_high
	found := false
	for _, evt := range events {
		if evt.EventType == "cpu_high" {
			found = true
			if evt.WebhookID == 0 {
				t.Fatalf("cpu_high event webhook_id = 0, want valid id")
			}
			if evt.ServerID != 1 {
				t.Fatalf("cpu_high event server_id = %d, want 1", evt.ServerID)
			}
			break
		}
	}
	if !found {
		t.Fatalf("no cpu_high event recorded; events found: %d", len(events))
	}
}

func TestWebhookServiceEvaluateAndFireDisabledConfig(t *testing.T) {
	state := newTestServices(t)

	_, err := state.svc.Webhook.CreateConfig("disabled-test", "http://127.0.0.1:1/webhook", "test", false)
	if err != nil {
		t.Fatalf("create disabled webhook config: %v", err)
	}

	metric := model.ServerMetric{
		CPUPercent:  95.0,
		RAMUsedMB:   16000,
		RAMTotalMB:  16000,
		DiskUsedGB:  500,
		DiskTotalGB: 500,
		RecordedAt:  time.Now(),
	}
	state.svc.Webhook.EvaluateAndFire(1, metric)

	// Allow async goroutines to complete
	time.Sleep(200 * time.Millisecond)

	events, err := state.svc.Webhook.ListEvents(100)
	if err != nil {
		t.Fatalf("list events: %v", err)
	}
	if len(events) != 0 {
		t.Fatalf("events len = %d, want 0 (disabled config should not fire)", len(events))
	}
}

func TestWebhookServiceEvaluateAndFireNoConfigs(t *testing.T) {
	state := newTestServices(t)

	metric := model.ServerMetric{
		CPUPercent:  99.0,
		RAMUsedMB:   16000,
		RAMTotalMB:  16000,
		DiskUsedGB:  500,
		DiskTotalGB: 500,
		RecordedAt:  time.Now(),
	}

	// Should not panic, should not error
	state.svc.Webhook.EvaluateAndFire(1, metric)

	// Allow async goroutines to complete
	time.Sleep(100 * time.Millisecond)

	events, err := state.svc.Webhook.ListEvents(100)
	if err != nil {
		t.Fatalf("list events: %v", err)
	}
	if len(events) != 0 {
		t.Fatalf("events len = %d, want 0 (no configs defined)", len(events))
	}
}



// ── formatAlertValue ────────────────────────────────────────────────────────

func TestFormatAlertValueCPUHigh(t *testing.T) {
	metric := model.ServerMetric{CPUPercent: 95.3}
	got := formatAlertValue(metric, "cpu_high")
	want := "95.3%"
	if got != want {
		t.Fatalf("formatAlertValue(cpu_high) = %q, want %q", got, want)
	}
}

func TestFormatAlertValueRAMHighWithTotal(t *testing.T) {
	metric := model.ServerMetric{RAMUsedMB: 512, RAMTotalMB: 2048}
	got := formatAlertValue(metric, "ram_high")
	want := "25.0% (512/2048 MB)"
	if got != want {
		t.Fatalf("formatAlertValue(ram_high with total) = %q, want %q", got, want)
	}
}

func TestFormatAlertValueRAMHighWithoutTotal(t *testing.T) {
	metric := model.ServerMetric{RAMUsedMB: 1024, RAMTotalMB: 0}
	got := formatAlertValue(metric, "ram_high")
	want := "1024 MB used"
	if got != want {
		t.Fatalf("formatAlertValue(ram_high no total) = %q, want %q", got, want)
	}
}

func TestFormatAlertValueDiskHighWithTotal(t *testing.T) {
	metric := model.ServerMetric{DiskUsedGB: 25.5, DiskTotalGB: 100}
	got := formatAlertValue(metric, "disk_high")
	want := "25.5% (25.5/100.0 GB)"
	if got != want {
		t.Fatalf("formatAlertValue(disk_high with total) = %q, want %q", got, want)
	}
}

func TestFormatAlertValueDiskHighWithoutTotal(t *testing.T) {
	metric := model.ServerMetric{DiskUsedGB: 50.5, DiskTotalGB: 0}
	got := formatAlertValue(metric, "disk_high")
	want := "50.5 GB used"
	if got != want {
		t.Fatalf("formatAlertValue(disk_high no total) = %q, want %q", got, want)
	}
}

func TestFormatAlertValueDefault(t *testing.T) {
	metric := model.ServerMetric{}
	got := formatAlertValue(metric, "unknown_alert")
	if got != "" {
		t.Fatalf("formatAlertValue(default) = %q, want empty string", got)
	}
}

// ── IDText ───────────────────────────────────────────────────────────────────

func TestIDText(t *testing.T) {
	tests := []struct {
		input int64
		want  string
	}{
		{input: 0, want: "0"},
		{input: 1, want: "1"},
		{input: 42, want: "42"},
		{input: 999, want: "999"},
		{input: 1234567890, want: "1234567890"},
	}
	for _, tt := range tests {
		got := IDText(tt.input)
		if got != tt.want {
			t.Fatalf("IDText(%d) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

// ── CountText ────────────────────────────────────────────────────────────────

func TestCountText(t *testing.T) {
	tests := []struct {
		input int
		want  string
	}{
		{input: 0, want: "0"},
		{input: 1, want: "1"},
		{input: 42, want: "42"},
		{input: 999, want: "999"},
	}
	for _, tt := range tests {
		got := CountText(tt.input)
		if got != tt.want {
			t.Fatalf("CountText(%d) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

