package repo

import (
	"path/filepath"
	"testing"
	"time"

	"raevtar/internal/model"
)

func testServerRepoDB(t *testing.T) *Repositories {
	t.Helper()
	db := InitSQLite(filepath.Join(t.TempDir(), "test.db"))
	t.Cleanup(func() { _ = db.Close() })
	AutoMigrate(db)
	return New(db)
}

func TestServerRepoCreateAndGetByID(t *testing.T) {
	repos := testServerRepoDB(t)

	s := &model.Server{
		Name: "prod-web-01",
		Host: "10.0.0.1",
		Port: 22,
		Tags: "production,web",
	}
	if err := repos.Server.Create(s); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if s.ID == 0 {
		t.Fatal("server.ID should be set after Create")
	}
	if s.CreatedAt.IsZero() {
		t.Fatal("server.CreatedAt should be set after Create")
	}

	got, err := repos.Server.GetByID(s.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}

	if got.Name != "prod-web-01" {
		t.Errorf("Name = %q, want %q", got.Name, "prod-web-01")
	}
	if got.Host != "10.0.0.1" {
		t.Errorf("Host = %q, want %q", got.Host, "10.0.0.1")
	}
	if got.Port != 22 {
		t.Errorf("Port = %d, want %d", got.Port, 22)
	}
	if got.Tags != "production,web" {
		t.Errorf("Tags = %q, want %q", got.Tags, "production,web")
	}
	if got.LastSeen != nil {
		t.Errorf("LastSeen = %v, want nil", got.LastSeen)
	}
	if got.CreatedAt.IsZero() {
		t.Error("CreatedAt should not be zero")
	}
}

func TestServerRepoGetByName(t *testing.T) {
	repos := testServerRepoDB(t)

	s := &model.Server{Name: "db-master", Host: "10.0.0.2", Port: 5432}
	if err := repos.Server.Create(s); err != nil {
		t.Fatalf("Create: %v", err)
	}

	got, err := repos.Server.GetByName("db-master")
	if err != nil {
		t.Fatalf("GetByName: %v", err)
	}
	if got.ID != s.ID {
		t.Errorf("ID = %d, want %d", got.ID, s.ID)
	}
	if got.Host != "10.0.0.2" {
		t.Errorf("Host = %q, want %q", got.Host, "10.0.0.2")
	}
}

func TestServerRepoGetByNameNotFound(t *testing.T) {
	repos := testServerRepoDB(t)

	_, err := repos.Server.GetByName("nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent server name")
	}
}

func TestServerRepoList(t *testing.T) {
	repos := testServerRepoDB(t)

	names := []string{"gamma", "alpha", "beta"}
	for _, name := range names {
		s := &model.Server{Name: name, Host: "10.0.0.1", Port: 22}
		if err := repos.Server.Create(s); err != nil {
			t.Fatalf("Create %s: %v", name, err)
		}
	}

	servers, err := repos.Server.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(servers) != 3 {
		t.Fatalf("got %d servers, want 3", len(servers))
	}
	// List returns ORDER BY name: alpha, beta, gamma
	if servers[0].Name != "alpha" {
		t.Errorf("servers[0].Name = %q, want %q", servers[0].Name, "alpha")
	}
	if servers[1].Name != "beta" {
		t.Errorf("servers[1].Name = %q, want %q", servers[1].Name, "beta")
	}
	if servers[2].Name != "gamma" {
		t.Errorf("servers[2].Name = %q, want %q", servers[2].Name, "gamma")
	}
}

func TestServerRepoListEmpty(t *testing.T) {
	repos := testServerRepoDB(t)

	servers, err := repos.Server.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(servers) != 0 {
		t.Errorf("got %d servers, want 0", len(servers))
	}
}

func TestServerRepoUpdate(t *testing.T) {
	repos := testServerRepoDB(t)

	s := &model.Server{Name: "old-name", Host: "10.0.0.1", Port: 22, Tags: "old"}
	if err := repos.Server.Create(s); err != nil {
		t.Fatalf("Create: %v", err)
	}

	s.Name = "new-name"
	s.Host = "10.0.0.2"
	s.Port = 2022
	s.Tags = "new,updated"
	if err := repos.Server.Update(s); err != nil {
		t.Fatalf("Update: %v", err)
	}

	got, err := repos.Server.GetByID(s.ID)
	if err != nil {
		t.Fatalf("GetByID after Update: %v", err)
	}
	if got.Name != "new-name" {
		t.Errorf("Name = %q, want %q", got.Name, "new-name")
	}
	if got.Host != "10.0.0.2" {
		t.Errorf("Host = %q, want %q", got.Host, "10.0.0.2")
	}
	if got.Port != 2022 {
		t.Errorf("Port = %d, want %d", got.Port, 2022)
	}
	if got.Tags != "new,updated" {
		t.Errorf("Tags = %q, want %q", got.Tags, "new,updated")
	}
	// Other fields should not be affected
	if got.CreatedAt.IsZero() {
		t.Error("CreatedAt should not be zero after Update")
	}
}

func TestServerRepoDelete(t *testing.T) {
	repos := testServerRepoDB(t)

	s := &model.Server{Name: "to-delete", Host: "10.0.0.1", Port: 22}
	if err := repos.Server.Create(s); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if err := repos.Server.Delete(s.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	got, err := repos.Server.GetByID(s.ID)
	if err == nil {
		t.Fatalf("expected error after Delete, got server: %+v", got)
	}
}

func TestServerRepoUpdateLastSeen(t *testing.T) {
	repos := testServerRepoDB(t)

	s := &model.Server{Name: "seen-server", Host: "10.0.0.1", Port: 22}
	if err := repos.Server.Create(s); err != nil {
		t.Fatalf("Create: %v", err)
	}

	now := time.Now().Truncate(time.Millisecond)
	if err := repos.Server.UpdateLastSeen(s.ID, now); err != nil {
		t.Fatalf("UpdateLastSeen: %v", err)
	}

	got, err := repos.Server.GetByID(s.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.LastSeen == nil {
		t.Fatal("LastSeen should not be nil after UpdateLastSeen")
	}
	if !got.LastSeen.Equal(now) {
		t.Errorf("LastSeen = %v, want %v", got.LastSeen, now)
	}
}

func TestServerRepoUpdateAgentTokenHash(t *testing.T) {
	repos := testServerRepoDB(t)

	s := &model.Server{Name: "hash-server", Host: "10.0.0.1", Port: 22}
	if err := repos.Server.Create(s); err != nil {
		t.Fatalf("Create: %v", err)
	}

	hash := "sha256$abc123deadbeef"
	if err := repos.Server.UpdateAgentTokenHash(s.ID, hash); err != nil {
		t.Fatalf("UpdateAgentTokenHash: %v", err)
	}

	got, err := repos.Server.GetByID(s.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.AgentTokenHash != hash {
		t.Errorf("AgentTokenHash = %q, want %q", got.AgentTokenHash, hash)
	}
}
