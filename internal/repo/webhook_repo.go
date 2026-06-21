package repo

import (
	"time"

	"github.com/jmoiron/sqlx"
	"raevtar/internal/model"
)

type WebhookRepo struct{ db *sqlx.DB }

func (r *WebhookRepo) InsertConfig(cfg *model.WebhookConfig) error {
	now := time.Now().UTC()
	cfg.CreatedAt = now
	res, err := r.db.Exec(`INSERT INTO webhook_configs (name, url, secret, enabled, created_at) VALUES (?, ?, ?, ?, ?)`,
		cfg.Name, cfg.URL, cfg.Secret, boolInt(cfg.Enabled), now)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	cfg.ID = id
	return nil
}

func (r *WebhookRepo) ListConfigs() ([]model.WebhookConfig, error) {
	var cfgs []model.WebhookConfig
	err := r.db.Select(&cfgs, `SELECT id, name, url, enabled, created_at FROM webhook_configs ORDER BY name`)
	if err != nil {
		return nil, err
	}
	return cfgs, nil
}

func (r *WebhookRepo) ListEnabledConfigs() ([]model.WebhookConfig, error) {
	var cfgs []model.WebhookConfig
	err := r.db.Select(&cfgs, `SELECT * FROM webhook_configs WHERE enabled = 1 ORDER BY name`)
	if err != nil {
		return nil, err
	}
	return cfgs, nil
}

func (r *WebhookRepo) GetConfig(id int64) (*model.WebhookConfig, error) {
	var cfg model.WebhookConfig
	err := r.db.Get(&cfg, `SELECT * FROM webhook_configs WHERE id = ?`, id)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (r *WebhookRepo) UpdateConfig(cfg *model.WebhookConfig) error {
	_, err := r.db.Exec(`UPDATE webhook_configs SET name = ?, url = ?, secret = ?, enabled = ? WHERE id = ?`,
		cfg.Name, cfg.URL, cfg.Secret, boolInt(cfg.Enabled), cfg.ID)
	return err
}

func (r *WebhookRepo) DeleteConfig(id int64) error {
	_, err := r.db.Exec(`DELETE FROM webhook_configs WHERE id = ?`, id)
	return err
}

func (r *WebhookRepo) InsertEvent(event *model.WebhookEvent) error {
	now := time.Now().UTC()
	event.FiredAt = now
	res, err := r.db.Exec(`INSERT INTO webhook_events (webhook_id, event_type, server_id, payload, response_code, response_body, fired_at) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		event.WebhookID, event.EventType, event.ServerID, event.Payload, event.ResponseCode, event.ResponseBody, now)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	event.ID = id
	return nil
}

func (r *WebhookRepo) ListEvents(limit int) ([]model.WebhookEvent, error) {
	var events []model.WebhookEvent
	err := r.db.Select(&events, `SELECT * FROM webhook_events ORDER BY fired_at DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	return events, nil
}

func boolInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
