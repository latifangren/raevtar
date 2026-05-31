package model

import "time"

type Server struct {
	ID        int64      `db:"id" json:"id"`
	Name      string     `db:"name" json:"name"`
	Host      string     `db:"host" json:"host"`
	Port      int        `db:"port" json:"port"`
	Tags      string     `db:"tags" json:"tags"`
	LastSeen  *time.Time `db:"last_seen" json:"last_seen"`
	CreatedAt time.Time  `db:"created_at" json:"created_at"`
}
