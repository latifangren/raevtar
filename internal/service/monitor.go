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
	"math"
	"strings"
	"time"

	"raevtar/internal/model"
	"raevtar/internal/repo"
)

type MonitorService struct {
	repos *repo.Repositories
}

var ErrServerNotFound = errors.New("server not found")
var ErrInvalidMetricPayload = errors.New("invalid metric payload")

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

func (s *MonitorService) GetServerByName(name string) (*model.Server, error) {
	return s.repos.Server.GetByName(name)
}

func (s *MonitorService) UpdateServer(id int64, name, host string, port int, tags string) (*model.Server, error) {
	server, err := s.repos.Server.GetByID(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w: %w", ErrServerNotFound, err)
		}
		return nil, fmt.Errorf("get server: %w", err)
	}

	name = strings.TrimSpace(name)
	host = strings.TrimSpace(host)
	tags = strings.TrimSpace(tags)
	if name == "" || host == "" {
		return nil, fmt.Errorf("update server: name and host required")
	}
	if port <= 0 {
		return nil, fmt.Errorf("update server: port required")
	}

	server.Name = name
	server.Host = host
	server.Port = port
	server.Tags = tags
	if err := s.repos.Server.Update(server); err != nil {
		return nil, fmt.Errorf("update server: %w", err)
	}
	return server, nil
}

func (s *MonitorService) RecordMetrics(serverID int64, m model.ServerMetric) error {
	if _, err := s.repos.Server.GetByID(serverID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("%w: %w", ErrServerNotFound, err)
		}
		return fmt.Errorf("get server: %w", err)
	}
	if err := validateServerMetric(m); err != nil {
		return err
	}

	recordedAt := time.Now().UTC()
	m.ServerID = serverID
	m.RecordedAt = recordedAt
	if err := s.repos.Metric.Insert(&m); err != nil {
		return fmt.Errorf("record metrics: %w", err)
	}
	if err := s.repos.Server.UpdateLastSeen(serverID, recordedAt); err != nil {
		return fmt.Errorf("update last seen: %w", err)
	}
	return nil
}

func validateServerMetric(m model.ServerMetric) error {
	checks := []struct {
		name  string
		value float64
	}{
		{name: "cpu percent", value: m.CPUPercent},
		{name: "cpu load1", value: m.CPULoad1},
		{name: "cpu load5", value: m.CPULoad5},
		{name: "cpu load15", value: m.CPULoad15},
		{name: "ram used", value: m.RAMUsedMB},
		{name: "ram total", value: m.RAMTotalMB},
		{name: "disk used", value: m.DiskUsedGB},
		{name: "disk total", value: m.DiskTotalGB},
		{name: "temperature", value: m.TemperatureC},
	}
	for _, check := range checks {
		if math.IsNaN(check.value) || math.IsInf(check.value, 0) {
			return fmt.Errorf("%w: %s must be finite", ErrInvalidMetricPayload, check.name)
		}
	}

	if m.CPUPercent < 0 || m.CPUPercent > 100 {
		return fmt.Errorf("%w: cpu percent out of range", ErrInvalidMetricPayload)
	}
	if m.CPULoad1 < 0 || m.CPULoad5 < 0 || m.CPULoad15 < 0 {
		return fmt.Errorf("%w: cpu load cannot be negative", ErrInvalidMetricPayload)
	}
	if m.CPUCores < 0 {
		return fmt.Errorf("%w: cpu cores cannot be negative", ErrInvalidMetricPayload)
	}
	if m.RAMUsedMB < 0 || m.RAMTotalMB < 0 {
		return fmt.Errorf("%w: ram cannot be negative", ErrInvalidMetricPayload)
	}
	if m.RAMTotalMB > 0 && m.RAMUsedMB > m.RAMTotalMB {
		return fmt.Errorf("%w: ram used cannot exceed total", ErrInvalidMetricPayload)
	}
	if m.DiskUsedGB < 0 || m.DiskTotalGB < 0 {
		return fmt.Errorf("%w: disk cannot be negative", ErrInvalidMetricPayload)
	}
	if m.DiskTotalGB > 0 && m.DiskUsedGB > m.DiskTotalGB {
		return fmt.Errorf("%w: disk used cannot exceed total", ErrInvalidMetricPayload)
	}
	if m.TemperatureAvailable && (m.TemperatureC < -50 || m.TemperatureC > 150) {
		return fmt.Errorf("%w: temperature out of range", ErrInvalidMetricPayload)
	}
	if m.UptimeSeconds < 0 {
		return fmt.Errorf("%w: uptime cannot be negative", ErrInvalidMetricPayload)
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
