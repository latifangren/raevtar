package repo

import (
	"path/filepath"
	"testing"
	"time"

	"raevtar/internal/model"
)

func TestWebhookRepoInsertAndGetConfig(t *testing.T) {
	db := InitSQLite(filepath.Join(t.TempDir(), "test.db"))
	t.Cleanup(func() { _ = db.Close() })
	AutoMigrate(db)
	repos := New(db)

	cfg := &model.WebhookConfig{
		Name:    "test-webhook",
		URL:     "https://hooks.example.com/test",
		Secret:  "supersecret",
		Enabled: true,
	}
	if err := repos.Webhook.InsertConfig(cfg); err != nil {
		t.Fatalf("InsertConfig: %v", err)
	}
	if cfg.ID == 0 {
		t.Fatal("expected cfg.ID to be set after InsertConfig")
	}

	loaded, err := repos.Webhook.GetConfig(cfg.ID)
	if err != nil {
		t.Fatalf("GetConfig: %v", err)
	}
	if loaded == nil {
		t.Fatal("GetConfig returned nil")
	}
	if loaded.ID != cfg.ID {
		t.Errorf("ID: got %d, want %d", loaded.ID, cfg.ID)
	}
	if loaded.Name != "test-webhook" {
		t.Errorf("Name: got %q, want %q", loaded.Name, "test-webhook")
	}
	if loaded.URL != "https://hooks.example.com/test" {
		t.Errorf("URL: got %q, want %q", loaded.URL, "https://hooks.example.com/test")
	}
	if loaded.Secret != "supersecret" {
		t.Errorf("Secret: got %q, want %q", loaded.Secret, "supersecret")
	}
	if !loaded.Enabled {
		t.Error("Enabled should be true")
	}
	if loaded.CreatedAt.IsZero() {
		t.Error("CreatedAt should not be zero")
	}
}

func TestWebhookRepoListConfigs(t *testing.T) {
	db := InitSQLite(filepath.Join(t.TempDir(), "test.db"))
	t.Cleanup(func() { _ = db.Close() })
	AutoMigrate(db)
	repos := New(db)

	names := []string{"z-webhook", "a-webhook", "m-webhook"}
	for _, name := range names {
		cfg := &model.WebhookConfig{
			Name: name, URL: "https://hooks.example.com/" + name,
			Secret: "sec-" + name, Enabled: true,
		}
		if err := repos.Webhook.InsertConfig(cfg); err != nil {
			t.Fatalf("InsertConfig %q: %v", name, err)
		}
	}

	cfgs, err := repos.Webhook.ListConfigs()
	if err != nil {
		t.Fatalf("ListConfigs: %v", err)
	}
	if len(cfgs) != 3 {
		t.Fatalf("expected 3 configs, got %d", len(cfgs))
	}
	if cfgs[0].Name != "a-webhook" {
		t.Errorf("cfgs[0].Name: got %q, want %q", cfgs[0].Name, "a-webhook")
	}
	if cfgs[1].Name != "m-webhook" {
		t.Errorf("cfgs[1].Name: got %q, want %q", cfgs[1].Name, "m-webhook")
	}
	if cfgs[2].Name != "z-webhook" {
		t.Errorf("cfgs[2].Name: got %q, want %q", cfgs[2].Name, "z-webhook")
	}
}

func TestWebhookRepoListEnabledConfigs(t *testing.T) {
	db := InitSQLite(filepath.Join(t.TempDir(), "test.db"))
	t.Cleanup(func() { _ = db.Close() })
	AutoMigrate(db)
	repos := New(db)

	enabled := &model.WebhookConfig{
		Name: "enabled-hook", URL: "https://hooks.example.com/enabled",
		Secret: "s1", Enabled: true,
	}
	disabled := &model.WebhookConfig{
		Name: "disabled-hook", URL: "https://hooks.example.com/disabled",
		Secret: "s2", Enabled: false,
	}
	if err := repos.Webhook.InsertConfig(enabled); err != nil {
		t.Fatalf("InsertConfig enabled: %v", err)
	}
	if err := repos.Webhook.InsertConfig(disabled); err != nil {
		t.Fatalf("InsertConfig disabled: %v", err)
	}

	cfgs, err := repos.Webhook.ListEnabledConfigs()
	if err != nil {
		t.Fatalf("ListEnabledConfigs: %v", err)
	}
	if len(cfgs) != 1 {
		t.Fatalf("expected 1 enabled config, got %d", len(cfgs))
	}
	if cfgs[0].Name != "enabled-hook" {
		t.Errorf("Name: got %q, want %q", cfgs[0].Name, "enabled-hook")
	}
	if !cfgs[0].Enabled {
		t.Error("Enabled should be true for enabled config")
	}
}

func TestWebhookRepoUpdateConfig(t *testing.T) {
	db := InitSQLite(filepath.Join(t.TempDir(), "test.db"))
	t.Cleanup(func() { _ = db.Close() })
	AutoMigrate(db)
	repos := New(db)

	cfg := &model.WebhookConfig{
		Name: "original", URL: "https://hooks.example.com/original",
		Secret: "original-secret", Enabled: true,
	}
	if err := repos.Webhook.InsertConfig(cfg); err != nil {
		t.Fatalf("InsertConfig: %v", err)
	}

	cfg.Name = "updated"
	cfg.URL = "https://hooks.example.com/updated"
	cfg.Secret = "updated-secret"
	cfg.Enabled = false

	if err := repos.Webhook.UpdateConfig(cfg); err != nil {
		t.Fatalf("UpdateConfig: %v", err)
	}

	loaded, err := repos.Webhook.GetConfig(cfg.ID)
	if err != nil {
		t.Fatalf("GetConfig after update: %v", err)
	}
	if loaded.Name != "updated" {
		t.Errorf("Name: got %q, want %q", loaded.Name, "updated")
	}
	if loaded.URL != "https://hooks.example.com/updated" {
		t.Errorf("URL: got %q, want %q", loaded.URL, "https://hooks.example.com/updated")
	}
	if loaded.Secret != "updated-secret" {
		t.Errorf("Secret: got %q, want %q", loaded.Secret, "updated-secret")
	}
	if loaded.Enabled {
		t.Error("Enabled should be false after update")
	}
}

func TestWebhookRepoDeleteConfig(t *testing.T) {
	db := InitSQLite(filepath.Join(t.TempDir(), "test.db"))
	t.Cleanup(func() { _ = db.Close() })
	AutoMigrate(db)
	repos := New(db)

	cfg := &model.WebhookConfig{
		Name: "delete-me", URL: "https://hooks.example.com/delete",
		Secret: "del", Enabled: true,
	}
	if err := repos.Webhook.InsertConfig(cfg); err != nil {
		t.Fatalf("InsertConfig: %v", err)
	}

	if err := repos.Webhook.DeleteConfig(cfg.ID); err != nil {
		t.Fatalf("DeleteConfig: %v", err)
	}

	loaded, err := repos.Webhook.GetConfig(cfg.ID)
	if err == nil {
		t.Fatal("expected error after deletion")
	}
	if loaded != nil {
		t.Fatal("expected nil config after deletion")
	}
}

func TestWebhookRepoInsertEventAndListEvents(t *testing.T) {
	db := InitSQLite(filepath.Join(t.TempDir(), "test.db"))
	t.Cleanup(func() { _ = db.Close() })
	AutoMigrate(db)
	repos := New(db)

	cfg := &model.WebhookConfig{
		Name: "event-test", URL: "https://hooks.example.com/event",
		Secret: "evt", Enabled: true,
	}
	if err := repos.Webhook.InsertConfig(cfg); err != nil {
		t.Fatalf("InsertConfig: %v", err)
	}

	payload := `{"cpu": 95}`
	evt := &model.WebhookEvent{
		WebhookID:    cfg.ID,
		EventType:    "server.alert",
		ServerID:     42,
		Payload:      payload,
		ResponseCode: 200,
		ResponseBody: "OK",
	}
	if err := repos.Webhook.InsertEvent(evt); err != nil {
		t.Fatalf("InsertEvent: %v", err)
	}
	if evt.ID == 0 {
		t.Fatal("expected event.ID to be set after InsertEvent")
	}

	events, err := repos.Webhook.ListEvents(10)
	if err != nil {
		t.Fatalf("ListEvents: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].ID != evt.ID {
		t.Errorf("ID: got %d, want %d", events[0].ID, evt.ID)
	}
	if events[0].WebhookID != cfg.ID {
		t.Errorf("WebhookID: got %d, want %d", events[0].WebhookID, cfg.ID)
	}
	if events[0].EventType != "server.alert" {
		t.Errorf("EventType: got %q, want %q", events[0].EventType, "server.alert")
	}
	if events[0].Payload != payload {
		t.Errorf("Payload: got %q, want %q", events[0].Payload, payload)
	}
	if events[0].ResponseCode != 200 {
		t.Errorf("ResponseCode: got %d, want %d", events[0].ResponseCode, 200)
	}
	if events[0].FiredAt.IsZero() {
		t.Error("FiredAt should not be zero")
	}
}

func TestWebhookRepoListEventsRespectsLimit(t *testing.T) {
	db := InitSQLite(filepath.Join(t.TempDir(), "test.db"))
	t.Cleanup(func() { _ = db.Close() })
	AutoMigrate(db)
	repos := New(db)

	cfg := &model.WebhookConfig{
		Name: "limit-test", URL: "https://hooks.example.com/limit",
		Secret: "lim", Enabled: true,
	}
	if err := repos.Webhook.InsertConfig(cfg); err != nil {
		t.Fatalf("InsertConfig: %v", err)
	}

	for i := 0; i < 5; i++ {
		evt := &model.WebhookEvent{
			WebhookID: cfg.ID,
			EventType: "test.event",
			Payload:   "event-" + string(rune('0'+i)),
		}
		if err := repos.Webhook.InsertEvent(evt); err != nil {
			t.Fatalf("InsertEvent %d: %v", i, err)
		}
		time.Sleep(2 * time.Millisecond)
	}

	events, err := repos.Webhook.ListEvents(3)
	if err != nil {
		t.Fatalf("ListEvents with limit=3: %v", err)
	}
	if len(events) != 3 {
		t.Fatalf("expected 3 events, got %d", len(events))
	}

	if events[0].Payload != "event-4" {
		t.Errorf("expected newest event 'event-4', got %q", events[0].Payload)
	}
	if events[2].Payload != "event-2" {
		t.Errorf("expected third event 'event-2', got %q", events[2].Payload)
	}
}

func TestWebhookRepoListEventsEmpty(t *testing.T) {
	db := InitSQLite(filepath.Join(t.TempDir(), "test.db"))
	t.Cleanup(func() { _ = db.Close() })
	AutoMigrate(db)
	repos := New(db)

	events, err := repos.Webhook.ListEvents(10)
	if err != nil {
		t.Fatalf("ListEvents: %v", err)
	}
	if len(events) != 0 {
		t.Errorf("expected empty events, got %d", len(events))
	}
}

