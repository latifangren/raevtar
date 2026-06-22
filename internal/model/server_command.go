package model

import "time"

type CommandStatus string

const (
	CommandPending   CommandStatus = "pending"
	CommandRunning   CommandStatus = "running"
	CommandCompleted CommandStatus = "completed"
	CommandFailed    CommandStatus = "failed"
)

type ServerCommand struct {
	ID          int64         `db:"id" json:"id"`
	ServerID    int64         `db:"server_id" json:"server_id"`
	Command     string        `db:"command" json:"command"`
	Status      CommandStatus `db:"status" json:"status"`
	Payload     string        `db:"payload" json:"payload"`
	Result      string        `db:"result" json:"result"`
	QueuedAt    time.Time     `db:"queued_at" json:"queued_at"`
	StartedAt   *time.Time    `db:"started_at" json:"started_at"`
	CompletedAt *time.Time    `db:"completed_at" json:"completed_at"`
}
