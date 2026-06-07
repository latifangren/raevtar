package repo

import (
	"fmt"
	"log/slog"

	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

func InitSQLite(path string) *sqlx.DB {
	slog.Info("opening database", "path", path)
	db, err := sqlx.Open("sqlite", path)
	if err != nil {
		slog.Error("failed to open database", "error", err)
		panic(err)
	}
	db.SetMaxOpenConns(1) // SQLite write lock: 1 is enough
	return db
}

func AutoMigrate(db *sqlx.DB) {
	schema := `
	CREATE TABLE IF NOT EXISTS categories (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		slug TEXT UNIQUE NOT NULL,
		name TEXT NOT NULL,
		description TEXT DEFAULT '',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS posts (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		category_id INTEGER NOT NULL REFERENCES categories(id),
		title TEXT NOT NULL,
		slug TEXT UNIQUE NOT NULL,
		content_md TEXT NOT NULL DEFAULT '',
		excerpt TEXT NOT NULL DEFAULT '',
		published INTEGER DEFAULT 1,
		cover_image_url TEXT NOT NULL DEFAULT '',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS projects (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		slug TEXT UNIQUE NOT NULL,
		content_md TEXT NOT NULL DEFAULT '',
		excerpt TEXT NOT NULL DEFAULT '',
		published INTEGER DEFAULT 1,
		state TEXT NOT NULL DEFAULT 'active',
		featured INTEGER NOT NULL DEFAULT 0,
		sort_order INTEGER NOT NULL DEFAULT 0,
		cover_image_url TEXT NOT NULL DEFAULT '',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS project_updates (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		project_id INTEGER NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
		kind TEXT NOT NULL,
		title TEXT NOT NULL,
		content_md TEXT NOT NULL DEFAULT '',
		published INTEGER NOT NULL DEFAULT 1,
		pinned INTEGER NOT NULL DEFAULT 0,
		sort_order INTEGER NOT NULL DEFAULT 0,
		event_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS content_relations (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		source_type TEXT NOT NULL,
		source_id INTEGER NOT NULL,
		target_type TEXT NOT NULL,
		target_id INTEGER NOT NULL,
		relation_kind TEXT NOT NULL DEFAULT 'related',
		sort_order INTEGER NOT NULL DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(source_type, source_id, target_type, target_id, relation_kind)
	);

	CREATE TABLE IF NOT EXISTS project_showcase_items (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		project_id INTEGER NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
		kind TEXT NOT NULL,
		title TEXT NOT NULL,
		body_md TEXT NOT NULL DEFAULT '',
		asset_url TEXT NOT NULL DEFAULT '',
		external_url TEXT NOT NULL DEFAULT '',
		embed_provider TEXT NOT NULL DEFAULT '',
		embed_ref TEXT NOT NULL DEFAULT '',
		published INTEGER NOT NULL DEFAULT 1,
		sort_order INTEGER NOT NULL DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS servers (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT UNIQUE NOT NULL,
		host TEXT NOT NULL,
		port INTEGER DEFAULT 22,
		tags TEXT DEFAULT '',
		agent_token_hash TEXT NOT NULL DEFAULT '',
		last_seen DATETIME,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS server_metrics (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		server_id INTEGER NOT NULL REFERENCES servers(id),
		cpu_percent REAL DEFAULT 0,
		cpu_load_1 REAL DEFAULT 0,
		cpu_load_5 REAL DEFAULT 0,
		cpu_load_15 REAL DEFAULT 0,
		cpu_cores INTEGER DEFAULT 0,
		ram_used_mb REAL DEFAULT 0,
		ram_total_mb REAL DEFAULT 0,
		disk_used_gb REAL DEFAULT 0,
		disk_total_gb REAL DEFAULT 0,
		temperature_c REAL DEFAULT 0,
		temperature_available INTEGER DEFAULT 0,
		uptime_seconds INTEGER DEFAULT 0,
		online INTEGER DEFAULT 1,
		recorded_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_posts_category ON posts(category_id);
	CREATE INDEX IF NOT EXISTS idx_posts_slug ON posts(slug);
	CREATE INDEX IF NOT EXISTS idx_posts_published ON posts(published);
	CREATE INDEX IF NOT EXISTS idx_projects_slug ON projects(slug);
	CREATE INDEX IF NOT EXISTS idx_projects_published ON projects(published);
	CREATE INDEX IF NOT EXISTS idx_metrics_server ON server_metrics(server_id);

	CREATE TABLE IF NOT EXISTS tags (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT UNIQUE NOT NULL,
		slug TEXT UNIQUE NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS post_tags (
		post_id INTEGER NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
		tag_id INTEGER NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
		PRIMARY KEY (post_id, tag_id)
	);

	CREATE TABLE IF NOT EXISTS project_tags (
		project_id INTEGER NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
		tag_id INTEGER NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
		PRIMARY KEY (project_id, tag_id)
	);

	CREATE INDEX IF NOT EXISTS idx_post_tags_post ON post_tags(post_id);
	CREATE INDEX IF NOT EXISTS idx_post_tags_tag ON post_tags(tag_id);
	CREATE INDEX IF NOT EXISTS idx_project_tags_project ON project_tags(project_id);
	CREATE INDEX IF NOT EXISTS idx_project_tags_tag ON project_tags(tag_id);

	CREATE TABLE IF NOT EXISTS page_contents (
		key TEXT PRIMARY KEY,
		title TEXT NOT NULL DEFAULT '',
		summary TEXT NOT NULL DEFAULT '',
		content_md TEXT NOT NULL DEFAULT '',
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS media_assets (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		original_name TEXT NOT NULL,
		stored_name TEXT UNIQUE NOT NULL,
		url TEXT UNIQUE NOT NULL,
		mime_type TEXT NOT NULL,
		size_bytes INTEGER NOT NULL DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_media_assets_created ON media_assets(created_at);

	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT UNIQUE NOT NULL,
		password_hash TEXT NOT NULL,
		role TEXT NOT NULL DEFAULT 'readonly',
		display_name TEXT DEFAULT '',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS audit_logs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user TEXT NOT NULL DEFAULT 'system',
		action TEXT NOT NULL,
		details TEXT DEFAULT '',
		ip TEXT DEFAULT '',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_audit_logs_user ON audit_logs(user);
	CREATE INDEX IF NOT EXISTS idx_audit_logs_action ON audit_logs(action);
	CREATE INDEX IF NOT EXISTS idx_audit_logs_created ON audit_logs(created_at);

	CREATE TABLE IF NOT EXISTS editorial_inbox (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		source_type TEXT NOT NULL,
		source_value TEXT NOT NULL,
		category_hint TEXT NOT NULL DEFAULT '',
		priority INTEGER NOT NULL DEFAULT 50,
		not_before DATETIME NOT NULL,
		deadline DATETIME,
		note TEXT NOT NULL DEFAULT '',
		mode TEXT NOT NULL,
		status TEXT NOT NULL,
		published_post_id INTEGER,
		failure_note TEXT NOT NULL DEFAULT '',
		failure_meta TEXT NOT NULL DEFAULT '',
		claimed_by TEXT NOT NULL DEFAULT '',
		claim_token_hash TEXT NOT NULL DEFAULT '',
		claimed_at DATETIME,
		lease_expires_at DATETIME,
		attempt_count INTEGER NOT NULL DEFAULT 0,
		completed_at DATETIME,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS editorial_policy_state (
		name TEXT PRIMARY KEY,
		value INTEGER NOT NULL DEFAULT 0,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_editorial_inbox_status ON editorial_inbox(status);
	CREATE INDEX IF NOT EXISTS idx_editorial_inbox_not_before ON editorial_inbox(not_before);
	CREATE INDEX IF NOT EXISTS idx_editorial_inbox_priority ON editorial_inbox(priority);
	`
	if _, err := db.Exec(schema); err != nil {
		slog.Error("migration failed", "error", err)
		panic(err)
	}
	ensureColumn(db, "servers", "agent_token_hash", "TEXT NOT NULL DEFAULT ''")
	ensureColumn(db, "posts", "cover_image_url", "TEXT NOT NULL DEFAULT ''")
	ensureColumn(db, "projects", "state", "TEXT NOT NULL DEFAULT 'active'")
	ensureColumn(db, "projects", "featured", "INTEGER NOT NULL DEFAULT 0")
	ensureColumn(db, "projects", "sort_order", "INTEGER NOT NULL DEFAULT 0")
	ensureIndex(db, "idx_projects_state", "CREATE INDEX IF NOT EXISTS idx_projects_state ON projects(state)")
	ensureIndex(db, "idx_projects_featured", "CREATE INDEX IF NOT EXISTS idx_projects_featured ON projects(featured)")
	ensureIndex(db, "idx_projects_sort_order", "CREATE INDEX IF NOT EXISTS idx_projects_sort_order ON projects(sort_order)")
	ensureIndex(db, "idx_project_updates_project", "CREATE INDEX IF NOT EXISTS idx_project_updates_project ON project_updates(project_id)")
	ensureIndex(db, "idx_project_updates_kind", "CREATE INDEX IF NOT EXISTS idx_project_updates_kind ON project_updates(kind)")
	ensureIndex(db, "idx_project_updates_event_at", "CREATE INDEX IF NOT EXISTS idx_project_updates_event_at ON project_updates(event_at)")
	ensureIndex(db, "idx_content_relations_source", "CREATE INDEX IF NOT EXISTS idx_content_relations_source ON content_relations(source_type, source_id)")
	ensureIndex(db, "idx_content_relations_target", "CREATE INDEX IF NOT EXISTS idx_content_relations_target ON content_relations(target_type, target_id)")
	ensureIndex(db, "idx_project_showcase_items_project", "CREATE INDEX IF NOT EXISTS idx_project_showcase_items_project ON project_showcase_items(project_id)")
	ensureColumn(db, "server_metrics", "cpu_load_1", "REAL DEFAULT 0")
	ensureColumn(db, "server_metrics", "cpu_load_5", "REAL DEFAULT 0")
	ensureColumn(db, "server_metrics", "cpu_load_15", "REAL DEFAULT 0")
	ensureColumn(db, "server_metrics", "cpu_cores", "INTEGER DEFAULT 0")
	ensureColumn(db, "server_metrics", "disk_total_gb", "REAL DEFAULT 0")
	ensureColumn(db, "server_metrics", "temperature_c", "REAL DEFAULT 0")
	ensureColumn(db, "server_metrics", "temperature_available", "INTEGER DEFAULT 0")
	ensureColumn(db, "editorial_inbox", "published_post_id", "INTEGER")
	ensureColumn(db, "editorial_inbox", "failure_note", "TEXT NOT NULL DEFAULT ''")
	ensureColumn(db, "editorial_inbox", "failure_meta", "TEXT NOT NULL DEFAULT ''")
	ensureColumn(db, "editorial_inbox", "claimed_by", "TEXT NOT NULL DEFAULT ''")
	ensureColumn(db, "editorial_inbox", "claim_token_hash", "TEXT NOT NULL DEFAULT ''")
	ensureColumn(db, "editorial_inbox", "claimed_at", "DATETIME")
	ensureColumn(db, "editorial_inbox", "lease_expires_at", "DATETIME")
	ensureColumn(db, "editorial_inbox", "attempt_count", "INTEGER NOT NULL DEFAULT 0")
	ensureColumn(db, "editorial_inbox", "completed_at", "DATETIME")
	slog.Info("database migrated")
}

func ensureColumn(db *sqlx.DB, table, column, definition string) {
	rows, err := db.Queryx("PRAGMA table_info(" + table + ")")
	if err != nil {
		slog.Error("column inspection failed", "table", table, "column", column, "error", err)
		panic(err)
	}
	defer rows.Close()

	for rows.Next() {
		var cid int
		var name, colType string
		var notNull int
		var defaultValue any
		var pk int
		if err := rows.Scan(&cid, &name, &colType, &notNull, &defaultValue, &pk); err != nil {
			slog.Error("column inspection scan failed", "table", table, "column", column, "error", err)
			panic(err)
		}
		if name == column {
			return
		}
	}
	if err := rows.Err(); err != nil {
		slog.Error("column inspection iteration failed", "table", table, "column", column, "error", err)
		panic(err)
	}

	query := fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s", table, column, definition)
	if _, err := db.Exec(query); err != nil {
		slog.Error("column migration failed", "table", table, "column", column, "error", err)
		panic(err)
	}
}

func ensureIndex(db *sqlx.DB, name, statement string) {
	if _, err := db.Exec(statement); err != nil {
		slog.Error("index migration failed", "index", name, "error", err)
		panic(err)
	}
}

// Repositories groups all repos
type Repositories struct {
	Post            *PostRepo
	Project         *ProjectRepo
	ProjectUpdate   *ProjectUpdateRepo
	ContentRelation *ContentRelationRepo
	ProjectShowcase *ProjectShowcaseRepo
	PageContent     *PageContentRepo
	Category        *CategoryRepo
	EditorialInbox  *EditorialInboxRepo
	Server          *ServerRepo
	Metric          *MetricRepo
	Tag             *TagRepo
	Media           *MediaRepo
	User            *UserRepo
	Audit           *AuditRepo
}

func New(db *sqlx.DB) *Repositories {
	return &Repositories{
		Post:            &PostRepo{db: db},
		Project:         &ProjectRepo{db: db},
		ProjectUpdate:   &ProjectUpdateRepo{db: db},
		ContentRelation: &ContentRelationRepo{db: db},
		ProjectShowcase: &ProjectShowcaseRepo{db: db},
		PageContent:     &PageContentRepo{db: db},
		Category:        &CategoryRepo{db: db},
		EditorialInbox:  &EditorialInboxRepo{db: db},
		Server:          &ServerRepo{db: db},
		Metric:          &MetricRepo{db: db},
		Tag:             &TagRepo{db: db},
		Media:           &MediaRepo{db: db},
		User:            &UserRepo{db: db},
		Audit:           &AuditRepo{db: db},
	}
}
