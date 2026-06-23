package repo

import (
	"time"

	"raevtar/internal/model"

	"github.com/jmoiron/sqlx"
)

type ServerRepo struct{ db *sqlx.DB }

func (r *ServerRepo) List() ([]model.Server, error) {
	var servers []model.Server
	return servers, r.db.Select(&servers, "SELECT * FROM servers ORDER BY name")
}

func (r *ServerRepo) GetByID(id int64) (*model.Server, error) {
	var s model.Server
	err := r.db.Get(&s, "SELECT * FROM servers WHERE id = ?", id)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *ServerRepo) GetByName(name string) (*model.Server, error) {
	var s model.Server
	err := r.db.Get(&s, "SELECT * FROM servers WHERE name = ?", name)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *ServerRepo) Create(s *model.Server) error {
	now := time.Now()
	s.CreatedAt = now
	result, err := r.db.Exec(
		"INSERT INTO servers (name, host, port, tags, agent_token_hash, created_at) VALUES (?, ?, ?, ?, ?, ?)",
		s.Name, s.Host, s.Port, s.Tags, s.AgentTokenHash, now,
	)
	if err != nil {
		return err
	}
	id, _ := result.LastInsertId()
	s.ID = id
	return nil
}

func (r *ServerRepo) Delete(id int64) error {
	_, err := r.db.Exec("DELETE FROM servers WHERE id = ?", id)
	return err
}

func (r *ServerRepo) Update(s *model.Server) error {
	_, err := r.db.Exec(
		"UPDATE servers SET name = ?, host = ?, port = ?, tags = ? WHERE id = ?",
		s.Name, s.Host, s.Port, s.Tags, s.ID,
	)
	return err
}

func (r *ServerRepo) UpdateLastSeen(id int64, lastSeen time.Time) error {
	_, err := r.db.Exec("UPDATE servers SET last_seen = ? WHERE id = ?", lastSeen, id)
	return err
}

func (r *ServerRepo) UpdateAgentTokenHash(id int64, hash string) error {
	_, err := r.db.Exec("UPDATE servers SET agent_token_hash = ? WHERE id = ?", hash, id)
	return err
}

func (r *ServerRepo) GetStaleServers(cutoff time.Time) ([]model.Server, error) {
	var servers []model.Server
	err := r.db.Select(&servers,
		"SELECT * FROM servers WHERE last_seen IS NULL OR last_seen < ? ORDER BY name",
		cutoff)
	if err != nil {
		return nil, err
	}
	return servers, nil
}
