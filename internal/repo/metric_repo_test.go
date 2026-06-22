package repo

import (
	"path/filepath"
	"testing"
	"time"

	"raevtar/internal/model"
)

func testMetricRepoDB(t *testing.T) (*Repositories, *model.Server) {
	t.Helper()
	db := InitSQLite(filepath.Join(t.TempDir(), "test.db"))
	t.Cleanup(func() { _ = db.Close() })
	AutoMigrate(db)
	repos := New(db)

	srv := &model.Server{Name: "metric-host", Host: "10.0.0.1", Port: 22}
	if err := repos.Server.Create(srv); err != nil {
		t.Fatalf("create server: %v", err)
	}
	return repos, srv
}

func baseMetric(serverID int64, cpuPercent float64, recordedAt time.Time) model.ServerMetric {
	return model.ServerMetric{
		ServerID:             serverID,
		CPUPercent:           cpuPercent,
		CPULoad1:             0.5,
		CPULoad5:             0.4,
		CPULoad15:            0.3,
		CPUCores:             4,
		RAMUsedMB:            1024,
		RAMTotalMB:           8192,
		DiskUsedGB:           50,
		DiskTotalGB:          100,
		TemperatureC:         0,
		TemperatureAvailable: false,
		UptimeSeconds:        3600,
		Online:               true,
		TopProcesses:         "[]",
		Logs:                 "",
		RecordedAt:           recordedAt,
	}
}

func TestMetricRepoInsertAndGetByServerID(t *testing.T) {
	repos, srv := testMetricRepoDB(t)

	now := time.Now().Truncate(time.Millisecond)
	metrics := []model.ServerMetric{
		baseMetric(srv.ID, 10.5, now.Add(-2*time.Hour)),
		baseMetric(srv.ID, 20.5, now.Add(-1*time.Hour)),
		baseMetric(srv.ID, 30.5, now),
	}

	for i := range metrics {
		if err := repos.Metric.Insert(&metrics[i]); err != nil {
			t.Fatalf("Insert metric %d: %v", i, err)
		}
	}

	results, err := repos.Metric.GetByServerID(srv.ID, 10)
	if err != nil {
		t.Fatalf("GetByServerID: %v", err)
	}
	if len(results) != 3 {
		t.Fatalf("got %d metrics, want 3", len(results))
	}

	// Verify DESC ordering by recorded_at: newest first
	if results[0].CPUPercent != 30.5 {
		t.Errorf("results[0].CPUPercent = %f, want 30.5", results[0].CPUPercent)
	}
	if results[1].CPUPercent != 20.5 {
		t.Errorf("results[1].CPUPercent = %f, want 20.5", results[1].CPUPercent)
	}
	if results[2].CPUPercent != 10.5 {
		t.Errorf("results[2].CPUPercent = %f, want 10.5", results[2].CPUPercent)
	}

	// Verify fields round-trip correctly
	if results[0].ServerID != srv.ID {
		t.Errorf("results[0].ServerID = %d, want %d", results[0].ServerID, srv.ID)
	}
	if results[0].CPUCores != 4 {
		t.Errorf("results[0].CPUCores = %d, want 4", results[0].CPUCores)
	}
}

func TestMetricRepoInsertSetsID(t *testing.T) {
	repos, srv := testMetricRepoDB(t)

	now := time.Now().Truncate(time.Millisecond)
	m := baseMetric(srv.ID, 42.0, now)
	if err := repos.Metric.Insert(&m); err != nil {
		t.Fatalf("Insert: %v", err)
	}
	if m.ID == 0 {
		t.Error("metric.ID should be set after Insert")
	}
}

func TestMetricRepoGetByServerIDWithLimit(t *testing.T) {
	repos, srv := testMetricRepoDB(t)

	now := time.Now().Truncate(time.Millisecond)
	for i := 0; i < 5; i++ {
		m := baseMetric(srv.ID, float64(i+1)*10.0, now.Add(-time.Duration(i)*time.Hour))
		if err := repos.Metric.Insert(&m); err != nil {
			t.Fatalf("Insert metric %d: %v", i, err)
		}
	}

	results, err := repos.Metric.GetByServerID(srv.ID, 2)
	if err != nil {
		t.Fatalf("GetByServerID: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("got %d metrics, want 2", len(results))
	}
}

func TestMetricRepoGetByServerIDNoMetrics(t *testing.T) {
	repos, srv := testMetricRepoDB(t)

	results, err := repos.Metric.GetByServerID(srv.ID, 10)
	if err != nil {
		t.Fatalf("GetByServerID: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("got %d metrics, want 0", len(results))
	}
}

func TestMetricRepoGetByServerIDWrongServer(t *testing.T) {
	repos, srv := testMetricRepoDB(t)

	srv2 := &model.Server{Name: "other-host", Host: "10.0.0.2", Port: 22}
	if err := repos.Server.Create(srv2); err != nil {
		t.Fatalf("create second server: %v", err)
	}

	// Insert metric for srv, not srv2
	now := time.Now().Truncate(time.Millisecond)
	m := baseMetric(srv.ID, 50.0, now)
	if err := repos.Metric.Insert(&m); err != nil {
		t.Fatalf("Insert: %v", err)
	}

	results, err := repos.Metric.GetByServerID(srv2.ID, 10)
	if err != nil {
		t.Fatalf("GetByServerID: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("got %d metrics, want 0", len(results))
	}
}
