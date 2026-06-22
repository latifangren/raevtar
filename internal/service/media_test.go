package service

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMediaServiceUploadAndGetAsset(t *testing.T) {
	state := newTestServices(t)

	pngData := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0x00, 0x00, 0x00, 0x0D, 0x49, 0x48, 0x44, 0x52}
	asset, err := state.svc.Media.Upload(MediaUpload{
		OriginalName: "test.png",
		Reader:       bytes.NewReader(pngData),
	})
	if err != nil {
		t.Fatalf("upload png: %v", err)
	}
	if asset.ID == 0 {
		t.Fatalf("asset id = 0, want persisted id")
	}
	if asset.OriginalName != "test.png" {
		t.Fatalf("original name = %q, want test.png", asset.OriginalName)
	}
	if asset.MimeType != "image/png" {
		t.Fatalf("mime type = %q, want image/png", asset.MimeType)
	}
	if asset.SizeBytes <= 0 {
		t.Fatalf("size bytes = %d, want > 0", asset.SizeBytes)
	}
	if asset.URL == "" {
		t.Fatalf("url is empty, want non-empty")
	}
	if !strings.HasPrefix(asset.URL, "/uploads/") {
		t.Fatalf("url = %q, want /uploads/ prefix", asset.URL)
	}

	fetched, err := state.svc.Media.GetAsset(asset.ID)
	if err != nil {
		t.Fatalf("get asset: %v", err)
	}
	if fetched.OriginalName != asset.OriginalName {
		t.Fatalf("fetched original name = %q, want %q", fetched.OriginalName, asset.OriginalName)
	}
	if fetched.MimeType != asset.MimeType {
		t.Fatalf("fetched mime = %q, want %q", fetched.MimeType, asset.MimeType)
	}
	if fetched.SizeBytes != asset.SizeBytes {
		t.Fatalf("fetched size = %d, want %d", fetched.SizeBytes, asset.SizeBytes)
	}
	if fetched.URL != asset.URL {
		t.Fatalf("fetched url = %q, want %q", fetched.URL, asset.URL)
	}
}

func TestMediaServiceListAssets(t *testing.T) {
	state := newTestServices(t)

	names := []string{"first.png", "second.png", "third.png"}
	for _, name := range names {
		_, err := state.svc.Media.Upload(MediaUpload{
			OriginalName: name,
			Reader:       bytes.NewReader([]byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}),
		})
		if err != nil {
			t.Fatalf("upload %s: %v", name, err)
		}
	}

	assets, err := state.svc.Media.ListAssets()
	if err != nil {
		t.Fatalf("list assets: %v", err)
	}
	if len(assets) != 3 {
		t.Fatalf("assets len = %d, want 3", len(assets))
	}
	if assets[0].OriginalName != "third.png" {
		t.Fatalf("first asset = %q, want third.png (DESC order)", assets[0].OriginalName)
	}
	if assets[1].OriginalName != "second.png" {
		t.Fatalf("second asset = %q, want second.png", assets[1].OriginalName)
	}
	if assets[2].OriginalName != "first.png" {
		t.Fatalf("third asset = %q, want first.png", assets[2].OriginalName)
	}
	if assets[0].CreatedAt.IsZero() {
		t.Fatalf("first asset created_at is zero")
	}
}

func TestMediaServiceListAssetsEmpty(t *testing.T) {
	state := newTestServices(t)

	assets, err := state.svc.Media.ListAssets()
	if err != nil {
		t.Fatalf("list assets: %v", err)
	}
	if len(assets) != 0 {
		t.Fatalf("assets len = %d, want 0", len(assets))
	}
}

func TestMediaServiceGetAssetNotFound(t *testing.T) {
	state := newTestServices(t)

	_, err := state.svc.Media.GetAsset(99999)
	if err == nil {
		t.Fatalf("get non-existent asset should fail")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Fatalf("err = %v, want 'not found' message", err)
	}
}

func TestMediaServiceUploadVariousMimeTypes(t *testing.T) {
	state := newTestServices(t)

	pngAsset, err := state.svc.Media.Upload(MediaUpload{
		OriginalName: "photo.png",
		Reader:       bytes.NewReader([]byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0x00, 0x00, 0x00, 0x0D, 0x49, 0x48, 0x44, 0x52}),
	})
	if err != nil {
		t.Fatalf("upload png: %v", err)
	}
	if pngAsset.MimeType != "image/png" {
		t.Fatalf("png mime = %q, want image/png", pngAsset.MimeType)
	}

	jpgAsset, err := state.svc.Media.Upload(MediaUpload{
		OriginalName: "photo.jpg",
		Reader:       bytes.NewReader([]byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46, 0x49, 0x46, 0x00, 0x01}),
	})
	if err != nil {
		t.Fatalf("upload jpg: %v", err)
	}
	if jpgAsset.MimeType != "image/jpeg" {
		t.Fatalf("jpg mime = %q, want image/jpeg", jpgAsset.MimeType)
	}

	gifAsset, err := state.svc.Media.Upload(MediaUpload{
		OriginalName: "animation.gif",
		Reader:       bytes.NewReader([]byte("GIF89a")),
	})
	if err != nil {
		t.Fatalf("upload gif: %v", err)
	}
	if gifAsset.MimeType != "image/gif" {
		t.Fatalf("gif mime = %q, want image/gif", gifAsset.MimeType)
	}
}

func TestMediaServiceFilePath(t *testing.T) {
	state := newTestServices(t)

	data := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
	asset, err := state.svc.Media.Upload(MediaUpload{
		OriginalName: "onsite.png",
		Reader:       bytes.NewReader(data),
	})
	if err != nil {
		t.Fatalf("upload: %v", err)
	}

	path, err := state.svc.Media.FilePath(asset.StoredName)
	if err != nil {
		t.Fatalf("file path: %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat stored file: %v", err)
	}
	if info.Size() != int64(len(data)) {
		t.Fatalf("stored file size = %d, want %d", info.Size(), len(data))
	}
	if !filepath.IsAbs(path) {
		t.Fatalf("path is not absolute: %q", path)
	}
}

func TestMediaServiceUploadEmptyInput(t *testing.T) {
	state := newTestServices(t)

	_, err := state.svc.Media.Upload(MediaUpload{
		OriginalName: "",
		Reader:       bytes.NewReader([]byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}),
	})
	if err == nil {
		t.Fatalf("upload with empty name should fail")
	}
	if !strings.Contains(err.Error(), "file required") {
		t.Fatalf("err = %v, want 'file required'", err)
	}

	_, err = state.svc.Media.Upload(MediaUpload{
		OriginalName: "test.png",
		Reader:       nil,
	})
	if err == nil {
		t.Fatalf("upload with nil reader should fail")
	}
	if !strings.Contains(err.Error(), "file required") {
		t.Fatalf("err = %v, want 'file required'", err)
	}
}

