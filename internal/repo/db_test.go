package repo

import (
	"path/filepath"
	"testing"
)

func TestAutoMigrateAddsExtendedServerMetricColumnsToExistingDatabase(t *testing.T) {
	db := InitSQLite(filepath.Join(t.TempDir(), "legacy.db"))
	t.Cleanup(func() { _ = db.Close() })

	_, err := db.Exec(`
		CREATE TABLE server_metrics (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			server_id INTEGER NOT NULL,
			cpu_percent REAL DEFAULT 0,
			ram_used_mb REAL DEFAULT 0,
			ram_total_mb REAL DEFAULT 0,
			disk_used_gb REAL DEFAULT 0,
			uptime_seconds INTEGER DEFAULT 0,
			online INTEGER DEFAULT 1,
			recorded_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`)
	if err != nil {
		t.Fatalf("create legacy server_metrics: %v", err)
	}

	AutoMigrate(db)

	rows, err := db.Queryx("PRAGMA table_info(server_metrics)")
	if err != nil {
		t.Fatalf("inspect columns: %v", err)
	}
	defer rows.Close()

	columns := make(map[string]bool)
	for rows.Next() {
		var cid int
		var name, columnType string
		var notNull int
		var defaultValue any
		var primaryKey int
		if err := rows.Scan(&cid, &name, &columnType, &notNull, &defaultValue, &primaryKey); err != nil {
			t.Fatalf("scan column: %v", err)
		}
		columns[name] = true
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("iterate columns: %v", err)
	}

	for _, name := range []string{"cpu_load_1", "cpu_load_5", "cpu_load_15", "cpu_cores", "disk_total_gb", "temperature_c", "temperature_available"} {
		if !columns[name] {
			t.Errorf("server_metrics missing additive migration column %q", name)
		}
	}
}

func TestAutoMigrateCreatesEditorialInboxTable(t *testing.T) {
	db := InitSQLite(filepath.Join(t.TempDir(), "editorial.db"))
	t.Cleanup(func() { _ = db.Close() })

	AutoMigrate(db)

	rows, err := db.Queryx("PRAGMA table_info(editorial_inbox)")
	if err != nil {
		t.Fatalf("inspect editorial_inbox columns: %v", err)
	}
	defer rows.Close()
	columns := map[string]bool{}
	for rows.Next() {
		var cid int
		var name, columnType string
		var notNull int
		var defaultValue any
		var primaryKey int
		if err := rows.Scan(&cid, &name, &columnType, &notNull, &defaultValue, &primaryKey); err != nil {
			t.Fatalf("scan editorial column: %v", err)
		}
		columns[name] = true
	}
	for _, name := range []string{"source_type", "source_value", "category_hint", "priority", "not_before", "deadline", "note", "mode", "status", "published_post_id", "failure_note", "failure_meta", "claimed_by", "claim_token_hash", "claimed_at", "lease_expires_at", "attempt_count", "created_at", "updated_at"} {
		if !columns[name] {
			t.Fatalf("editorial_inbox missing column %q", name)
		}
	}
	if !columns["completed_at"] {
		t.Fatalf("editorial_inbox missing column %q", "completed_at")
	}

	policyRows, err := db.Queryx("PRAGMA table_info(editorial_policy_state)")
	if err != nil {
		t.Fatalf("inspect editorial_policy_state columns: %v", err)
	}
	defer policyRows.Close()
	policyColumns := map[string]bool{}
	for policyRows.Next() {
		var cid int
		var name, columnType string
		var notNull int
		var defaultValue any
		var primaryKey int
		if err := policyRows.Scan(&cid, &name, &columnType, &notNull, &defaultValue, &primaryKey); err != nil {
			t.Fatalf("scan policy column: %v", err)
		}
		policyColumns[name] = true
	}
	for _, name := range []string{"name", "value", "updated_at"} {
		if !policyColumns[name] {
			t.Fatalf("editorial_policy_state missing column %q", name)
		}
	}
}
