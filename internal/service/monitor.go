package service

import (
	"fmt"
	"log/slog"
	"time"

	"raevtar/internal/model"
	"raevtar/internal/repo"
)

type MonitorService struct {
	repos *repo.Repositories
}

func (s *MonitorService) CreateServer(name, host string, port int, tags string) (*model.Server, error) {
	server := &model.Server{
		Name: name,
		Host: host,
		Port: port,
		Tags: tags,
	}
	if err := s.repos.Server.Create(server); err != nil {
		return nil, fmt.Errorf("create server: %w", err)
	}
	return server, nil
}

func (s *MonitorService) ListServers() ([]model.Server, error) {
	return s.repos.Server.List()
}

func (s *MonitorService) GetServer(id int64) (*model.Server, error) {
	return s.repos.Server.GetByID(id)
}

func (s *MonitorService) RecordMetrics(serverID int64, m model.ServerMetric) error {
	m.ServerID = serverID
	m.RecordedAt = time.Now()
	if err := s.repos.Metric.Insert(&m); err != nil {
		return fmt.Errorf("record metrics: %w", err)
	}
	if err := s.repos.Server.UpdateLastSeen(serverID); err != nil {
		return fmt.Errorf("update last seen: %w", err)
	}
	return nil
}

func (s *MonitorService) GetRecentMetrics(serverID int64, limit int) ([]model.ServerMetric, error) {
	if limit <= 0 {
		limit = 20
	}
	return s.repos.Metric.GetByServerID(serverID, limit)
}

// PingServer is called by agents (including Hermes cron) to report status
func (s *MonitorService) PingServer(host string, port int) (*model.ServerMetric, error) {
	// Stub: in production, do actual HTTP/ICMP check
	slog.Debug("pinging server", "host", host, "port", port)
	return &model.ServerMetric{
		Online: true,
	}, nil
}
