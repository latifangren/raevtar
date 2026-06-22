package repo

import (
	"raevtar/internal/model"

	"github.com/jmoiron/sqlx"
)

type MediaRepo struct{ db *sqlx.DB }

func (r *MediaRepo) Create(asset *model.MediaAsset) error {
	result, err := r.db.Exec(`
		INSERT INTO media_assets (original_name, stored_name, url, mime_type, size_bytes, alt_text)
		VALUES (?, ?, ?, ?, ?, ?)`,
		asset.OriginalName, asset.StoredName, asset.URL, asset.MimeType, asset.SizeBytes, asset.AltText,
	)
	if err != nil {
		return err
	}
	asset.ID, _ = result.LastInsertId()
	return nil
}

func (r *MediaRepo) List(limit int) ([]model.MediaAsset, error) {
	var assets []model.MediaAsset
	if limit <= 0 {
		limit = 100
	}
	return assets, r.db.Select(&assets, `
		SELECT * FROM media_assets
		ORDER BY created_at DESC, id DESC
		LIMIT ?`, limit)
}

func (r *MediaRepo) GetByID(id int64) (*model.MediaAsset, error) {
	var asset model.MediaAsset
	if err := r.db.Get(&asset, "SELECT * FROM media_assets WHERE id = ?", id); err != nil {
		return nil, err
	}
	return &asset, nil
}
