package repo

import (
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"
	"raevtar/internal/model"
)

type CommandRepo struct{ db *sqlx.DB }

func (r *CommandRepo) Insert(cmd *model.ServerCommand) error {
	now := time.Now().UTC()
	cmd.QueuedAt = now
	cmd.Status = model.CommandPending
	res, err := r.db.Exec(`INSERT INTO server_commands (server_id, command, status, payload, result, queued_at) VALUES (?, ?, ?, ?, ?, ?)`,
		cmd.ServerID, cmd.Command, cmd.Status, cmd.Payload, cmd.Result, cmd.QueuedAt)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	cmd.ID = id
	return nil
}

func (r *CommandRepo) PendingByServerID(serverID int64) ([]model.ServerCommand, error) {
	var cmds []model.ServerCommand
	err := r.db.Select(&cmds, `SELECT * FROM server_commands WHERE server_id = ? AND status = 'pending' ORDER BY queued_at ASC`, serverID)
	if err != nil {
		return nil, err
	}
	return cmds, nil
}

func (r *CommandRepo) GetByID(id int64) (*model.ServerCommand, error) {
	var cmd model.ServerCommand
	err := r.db.Get(&cmd, `SELECT * FROM server_commands WHERE id = ?`, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &cmd, nil
}

func (r *CommandRepo) MarkRunning(id int64) error {
	now := time.Now().UTC()
	_, err := r.db.Exec(`UPDATE server_commands SET status = ?, started_at = ? WHERE id = ?`, model.CommandRunning, now, id)
	return err
}

func (r *CommandRepo) MarkCompleted(id int64, result string) error {
	now := time.Now().UTC()
	_, err := r.db.Exec(`UPDATE server_commands SET status = ?, result = ?, completed_at = ? WHERE id = ?`, model.CommandCompleted, result, now, id)
	return err
}

func (r *CommandRepo) MarkFailed(id int64, result string) error {
	now := time.Now().UTC()
	_, err := r.db.Exec(`UPDATE server_commands SET status = ?, result = ?, completed_at = ? WHERE id = ?`, model.CommandFailed, result, now, id)
	return err
}

func (r *CommandRepo) ListByServerID(serverID int64, limit int) ([]model.ServerCommand, error) {
	var cmds []model.ServerCommand
	err := r.db.Select(&cmds, `SELECT * FROM server_commands WHERE server_id = ? ORDER BY queued_at DESC LIMIT ?`, serverID, limit)
	if err != nil {
		return nil, err
	}
	return cmds, nil
}
