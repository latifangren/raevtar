package repo

import (
	"path/filepath"
	"testing"
)

func TestViewRepoRecordAndCountPostViews(t *testing.T) {
	db := InitSQLite(filepath.Join(t.TempDir(), "test.db"))
	t.Cleanup(func() { _ = db.Close() })
	AutoMigrate(db)
	repos := New(db)

	if err := repos.View.RecordPostView(1, "hash-abc"); err != nil {
		t.Fatalf("RecordPostView: %v", err)
	}

	count, err := repos.View.CountPostViews(1)
	if err != nil {
		t.Fatalf("CountPostViews: %v", err)
	}
	if count != 1 {
		t.Errorf("expected count 1, got %d", count)
	}
}

func TestViewRepoMultipleViews(t *testing.T) {
	db := InitSQLite(filepath.Join(t.TempDir(), "test.db"))
	t.Cleanup(func() { _ = db.Close() })
	AutoMigrate(db)
	repos := New(db)

	postID := int64(42)
	for i := 0; i < 5; i++ {
		hash := "ip-hash-" + string(rune('0'+i))
		if err := repos.View.RecordPostView(postID, hash); err != nil {
			t.Fatalf("RecordPostView %d: %v", i, err)
		}
	}

	count, err := repos.View.CountPostViews(postID)
	if err != nil {
		t.Fatalf("CountPostViews: %v", err)
	}
	if count != 5 {
		t.Errorf("expected count 5, got %d", count)
	}
}

func TestViewRepoCountAllPostViews(t *testing.T) {
	db := InitSQLite(filepath.Join(t.TempDir(), "test.db"))
	t.Cleanup(func() { _ = db.Close() })
	AutoMigrate(db)
	repos := New(db)

	// Post 1 gets 3 views
	for i := 0; i < 3; i++ {
		if err := repos.View.RecordPostView(1, "hash-a"+string(rune('0'+i))); err != nil {
			t.Fatalf("RecordPostView for post 1: %v", err)
		}
	}
	// Post 2 gets 2 views
	for i := 0; i < 2; i++ {
		if err := repos.View.RecordPostView(2, "hash-b"+string(rune('0'+i))); err != nil {
			t.Fatalf("RecordPostView for post 2: %v", err)
		}
	}

	allViews, err := repos.View.CountAllPostViews()
	if err != nil {
		t.Fatalf("CountAllPostViews: %v", err)
	}

	if len(allViews) != 2 {
		t.Fatalf("expected 2 post entries in map, got %d", len(allViews))
	}
	if allViews[1] != 3 {
		t.Errorf("post 1: expected 3 views, got %d", allViews[1])
	}
	if allViews[2] != 2 {
		t.Errorf("post 2: expected 2 views, got %d", allViews[2])
	}
}

func TestViewRepoCountPostViewsForNonExistentPost(t *testing.T) {
	db := InitSQLite(filepath.Join(t.TempDir(), "test.db"))
	t.Cleanup(func() { _ = db.Close() })
	AutoMigrate(db)
	repos := New(db)

	count, err := repos.View.CountPostViews(99999)
	if err != nil {
		t.Fatalf("CountPostViews for non-existent post: %v", err)
	}
	if count != 0 {
		t.Errorf("expected count 0 for non-existent post, got %d", count)
	}
}

func TestViewRepoCountAllPostViewsEmpty(t *testing.T) {
	db := InitSQLite(filepath.Join(t.TempDir(), "test.db"))
	t.Cleanup(func() { _ = db.Close() })
	AutoMigrate(db)
	repos := New(db)

	allViews, err := repos.View.CountAllPostViews()
	if err != nil {
		t.Fatalf("CountAllPostViews on empty table: %v", err)
	}
	if allViews == nil {
		t.Fatal("expected non-nil empty map, got nil")
	}
	if len(allViews) != 0 {
		t.Errorf("expected empty map, got %d entries", len(allViews))
	}
}
