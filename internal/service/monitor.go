package service

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"raevtar/internal/model"
	"raevtar/internal/repo"
)

type MonitorService struct {
	repos *repo.Repositories
}

var ErrServerNotFound = errors.New("server not found")

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

func (s *MonitorService) CreateServerWithAgentToken(name, host string, port int, tags string) (*model.Server, string, error) {
	token, hash, err := newAgentToken()
	if err != nil {
		return nil, "", fmt.Errorf("create agent token: %w", err)
	}
	server := &model.Server{
		Name:           name,
		Host:           host,
		Port:           port,
		Tags:           tags,
		AgentTokenHash: hash,
	}
	if err := s.repos.Server.Create(server); err != nil {
		return nil, "", fmt.Errorf("create server: %w", err)
	}
	return server, token, nil
}

func (s *MonitorService) ListServers() ([]model.Server, error) {
	return s.repos.Server.List()
}

func (s *MonitorService) GetServer(id int64) (*model.Server, error) {
	return s.repos.Server.GetByID(id)
}

func (s *MonitorService) RecordMetrics(serverID int64, m model.ServerMetric) error {
	if _, err := s.repos.Server.GetByID(serverID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("%w: %w", ErrServerNotFound, err)
		}
		return fmt.Errorf("get server: %w", err)
	}

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

func (s *MonitorService) RotateAgentToken(serverID int64) (string, error) {
	if _, err := s.repos.Server.GetByID(serverID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", fmt.Errorf("%w: %w", ErrServerNotFound, err)
		}
		return "", fmt.Errorf("get server: %w", err)
	}

	token, hash, err := newAgentToken()
	if err != nil {
		return "", fmt.Errorf("create agent token: %w", err)
	}
	if err := s.repos.Server.UpdateAgentTokenHash(serverID, hash); err != nil {
		return "", fmt.Errorf("update agent token: %w", err)
	}
	return token, nil
}

func (s *MonitorService) VerifyAgentToken(serverID int64, token string) bool {
	if token == "" {
		return false
	}
	server, err := s.repos.Server.GetByID(serverID)
	if err != nil || server.AgentTokenHash == "" {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(agentTokenHash(token)), []byte(server.AgentTokenHash)) == 1
}

func newAgentToken() (string, string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", "", err
	}
	token := hex.EncodeToString(b)
	return token, agentTokenHash(token), nil
}

func agentTokenHash(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

// PingServer is called by agents (including Hermes cron) to report status
func (s *MonitorService) PingServer(host string, port int) (*model.ServerMetric, error) {
	// Stub: in production, do actual HTTP/ICMP check
	slog.Debug("pinging server", "host", host, "port", port)
	return &model.ServerMetric{
		Online: true,
	}, nil
}
