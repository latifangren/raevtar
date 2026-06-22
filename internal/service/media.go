package service

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"raevtar/internal/model"
	"raevtar/internal/repo"
)

var ErrInvalidMediaInput = errors.New("invalid media input")

const maxMediaUploadBytes int64 = 5 << 20

type MediaService struct {
	repos *repo.Repositories
	dir   string
}

type MediaUpload struct {
	OriginalName string
	Reader       io.Reader
	AltText      string
}

func NewMediaService(repos *repo.Repositories, dir string) *MediaService {
	if strings.TrimSpace(dir) == "" {
		dir = filepath.Join(os.TempDir(), "raevtar-uploads")
	}
	return &MediaService{repos: repos, dir: dir}
}

func (s *MediaService) ListAssets() ([]model.MediaAsset, error) {
	assets, err := s.repos.Media.List(200)
	if err != nil {
		return nil, fmt.Errorf("list media: %w", err)
	}
	return assets, nil
}

func (s *MediaService) GetAsset(id int64) (*model.MediaAsset, error) {
	asset, err := s.repos.Media.GetByID(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("media not found: %w", err)
		}
		return nil, fmt.Errorf("get media: %w", err)
	}
	return asset, nil
}

func (s *MediaService) Upload(input MediaUpload) (*model.MediaAsset, error) {
	name := filepath.Base(strings.TrimSpace(input.OriginalName))
	if name == "." || name == "" || input.Reader == nil {
		return nil, fmt.Errorf("%w: file required", ErrInvalidMediaInput)
	}

	ext := strings.ToLower(filepath.Ext(name))
	if !allowedMediaExt(ext) {
		return nil, fmt.Errorf("%w: unsupported image type", ErrInvalidMediaInput)
	}

	data, err := io.ReadAll(io.LimitReader(input.Reader, maxMediaUploadBytes+1))
	if err != nil {
		return nil, fmt.Errorf("read media: %w", err)
	}
	if int64(len(data)) > maxMediaUploadBytes {
		return nil, fmt.Errorf("%w: file too large", ErrInvalidMediaInput)
	}
	if len(data) == 0 {
		return nil, fmt.Errorf("%w: empty file", ErrInvalidMediaInput)
	}

	mimeType := http.DetectContentType(data)
	if !allowedMediaMIME(mimeType) {
		return nil, fmt.Errorf("%w: unsupported image content", ErrInvalidMediaInput)
	}

	storedName, err := randomStoredName(ext)
	if err != nil {
		return nil, fmt.Errorf("name media: %w", err)
	}
	if err := os.MkdirAll(s.dir, 0o755); err != nil {
		return nil, fmt.Errorf("create media dir: %w", err)
	}
	path := filepath.Join(s.dir, storedName)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return nil, fmt.Errorf("write media: %w", err)
	}

	asset := &model.MediaAsset{
		OriginalName: name,
		StoredName:   storedName,
		URL:          "/uploads/" + storedName,
		MimeType:     mimeType,
		SizeBytes:    int64(len(data)),
	}
	altText := strings.TrimSpace(input.AltText)
	if altText == "" {
		altText = strings.TrimSuffix(name, ext)
		altText = strings.ReplaceAll(altText, "-", " ")
		altText = strings.ReplaceAll(altText, "_", " ")
	}
	asset.AltText = altText
	if err := s.repos.Media.Create(asset); err != nil {
		_ = os.Remove(path)
		return nil, fmt.Errorf("create media record: %w", err)
	}
	return s.repos.Media.GetByID(asset.ID)
}

func (s *MediaService) FilePath(storedName string) (string, error) {
	name := filepath.Base(strings.TrimSpace(storedName))
	if name == "." || name == "" || name != storedName {
		return "", fmt.Errorf("%w: invalid media path", ErrInvalidMediaInput)
	}
	return filepath.Join(s.dir, name), nil
}

func allowedMediaExt(ext string) bool {
	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif", ".webp":
		return true
	default:
		return false
	}
}

func allowedMediaMIME(mimeType string) bool {
	if strings.HasPrefix(mimeType, "image/") {
		return mimeType == "image/jpeg" || mimeType == "image/png" || mimeType == "image/gif" || mimeType == "image/webp"
	}
	exts, _ := mime.ExtensionsByType(mimeType)
	return len(exts) > 0 && allowedMediaExt(exts[0])
}

func randomStoredName(ext string) (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf) + ext, nil
}
