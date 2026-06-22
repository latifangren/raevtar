package repo

import (
	"database/sql"
	"path/filepath"
	"testing"

	"raevtar/internal/model"
)

func TestMediaRepoCreateAndGetByID(t *testing.T) {
	db := InitSQLite(filepath.Join(t.TempDir(), "test.db"))
	t.Cleanup(func() { _ = db.Close() })
	AutoMigrate(db)
	repos := New(db)

	asset := &model.MediaAsset{
		OriginalName: "photo.jpg",
		StoredName:   "abc123.jpg",
		URL:          "/uploads/abc123.jpg",
		MimeType:     "image/jpeg",
		SizeBytes:    102400,
	}

	if err := repos.Media.Create(asset); err != nil {
		t.Fatalf("Create media asset: %v", err)
	}
	if asset.ID == 0 {
		t.Fatal("expected asset.ID to be set after Create")
	}

	loaded, err := repos.Media.GetByID(asset.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if loaded == nil {
		t.Fatal("GetByID returned nil asset")
	}

	if loaded.ID != asset.ID {
		t.Errorf("ID: got %d, want %d", loaded.ID, asset.ID)
	}
	if loaded.OriginalName != "photo.jpg" {
		t.Errorf("OriginalName: got %q, want %q", loaded.OriginalName, "photo.jpg")
	}
	if loaded.StoredName != "abc123.jpg" {
		t.Errorf("StoredName: got %q, want %q", loaded.StoredName, "abc123.jpg")
	}
	if loaded.URL != "/uploads/abc123.jpg" {
		t.Errorf("URL: got %q, want %q", loaded.URL, "/uploads/abc123.jpg")
	}
	if loaded.MimeType != "image/jpeg" {
		t.Errorf("MimeType: got %q, want %q", loaded.MimeType, "image/jpeg")
	}
	if loaded.SizeBytes != 102400 {
		t.Errorf("SizeBytes: got %d, want %d", loaded.SizeBytes, 102400)
	}
	if loaded.CreatedAt.IsZero() {
		t.Error("CreatedAt should not be zero")
	}
}

func TestMediaRepoGetByIDNotFound(t *testing.T) {
	db := InitSQLite(filepath.Join(t.TempDir(), "test.db"))
	t.Cleanup(func() { _ = db.Close() })
	AutoMigrate(db)
	repos := New(db)

	loaded, err := repos.Media.GetByID(99999)
	if err == nil {
		t.Fatal("expected error for non-existent ID")
	}
	if err != sql.ErrNoRows {
		t.Fatalf("expected sql.ErrNoRows, got %v", err)
	}
	if loaded != nil {
		t.Fatal("expected nil asset for non-existent ID")
	}
}

func TestMediaRepoList(t *testing.T) {
	db := InitSQLite(filepath.Join(t.TempDir(), "test.db"))
	t.Cleanup(func() { _ = db.Close() })
	AutoMigrate(db)
	repos := New(db)

	assets := []*model.MediaAsset{
		{OriginalName: "first.png", StoredName: "f1.png", URL: "/uploads/f1.png", MimeType: "image/png", SizeBytes: 100},
		{OriginalName: "second.png", StoredName: "f2.png", URL: "/uploads/f2.png", MimeType: "image/png", SizeBytes: 200},
		{OriginalName: "third.png", StoredName: "f3.png", URL: "/uploads/f3.png", MimeType: "image/png", SizeBytes: 300},
	}
	for _, a := range assets {
		if err := repos.Media.Create(a); err != nil {
			t.Fatalf("Create asset %q: %v", a.OriginalName, err)
		}
	}

	result, err := repos.Media.List(10)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(result) != 3 {
		t.Fatalf("expected 3 assets, got %d", len(result))
	}

	// Ordered by created_at DESC, id DESC: third, second, first
	if result[0].OriginalName != "third.png" {
		t.Errorf("first item: got %q, want %q", result[0].OriginalName, "third.png")
	}
	if result[2].OriginalName != "first.png" {
		t.Errorf("last item: got %q, want %q", result[2].OriginalName, "first.png")
	}

	// Verify all fields on one entry
	if result[2].StoredName != "f1.png" {
		t.Errorf("StoredName: got %q, want %q", result[2].StoredName, "f1.png")
	}
	if result[2].MimeType != "image/png" {
		t.Errorf("MimeType: got %q, want %q", result[2].MimeType, "image/png")
	}
	if result[2].SizeBytes != 100 {
		t.Errorf("SizeBytes: got %d, want %d", result[2].SizeBytes, 100)
	}
}

func TestMediaRepoListWithZeroLimit(t *testing.T) {
	db := InitSQLite(filepath.Join(t.TempDir(), "test.db"))
	t.Cleanup(func() { _ = db.Close() })
	AutoMigrate(db)
	repos := New(db)

	for i := 0; i < 5; i++ {
		a := &model.MediaAsset{
			OriginalName: "file" + string(rune('0'+i)) + ".txt",
			StoredName:   "s" + string(rune('0'+i)) + ".txt",
			URL:          "/uploads/s" + string(rune('0'+i)) + ".txt",
			MimeType:     "text/plain",
			SizeBytes:    int64(i * 10),
		}
		if err := repos.Media.Create(a); err != nil {
			t.Fatalf("Create asset %d: %v", i, err)
		}
	}

	result, err := repos.Media.List(0)
	if err != nil {
		t.Fatalf("List with zero limit: %v", err)
	}
	if len(result) != 5 {
		t.Errorf("expected 5 assets (default limit 100), got %d", len(result))
	}
}
