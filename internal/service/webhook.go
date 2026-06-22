package service

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"raevtar/internal/model"
	"raevtar/internal/repo"
)

type WebhookService struct {
	repos *repo.Repositories
	http  *http.Client
}

func NewWebhookService(repos *repo.Repositories) *WebhookService {
	return &WebhookService{
		repos: repos,
		http: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (s *WebhookService) ListConfigs() ([]model.WebhookConfig, error) {
	return s.repos.Webhook.ListConfigs()
}

func (s *WebhookService) GetConfig(id int64) (*model.WebhookConfig, error) {
	return s.repos.Webhook.GetConfig(id)
}

func (s *WebhookService) CreateConfig(name, url, secret string, enabled bool) (*model.WebhookConfig, error) {
	cfg := &model.WebhookConfig{
		Name:    name,
		URL:     url,
		Secret:  secret,
		Enabled: enabled,
	}
	if err := s.repos.Webhook.InsertConfig(cfg); err != nil {
		return nil, fmt.Errorf("create webhook config: %w", err)
	}
	return cfg, nil
}

func (s *WebhookService) UpdateConfig(cfg *model.WebhookConfig) error {
	return s.repos.Webhook.UpdateConfig(cfg)
}

func (s *WebhookService) DeleteConfig(id int64) error {
	return s.repos.Webhook.DeleteConfig(id)
}

func (s *WebhookService) ListEvents(limit int) ([]model.WebhookEvent, error) {
	return s.repos.Webhook.ListEvents(limit)
}

// Thresholds for alert evaluation.
type AlertThresholds struct {
	CPUPercent  float64
	RAMPercent  float64
	DiskPercent float64
}

var DefaultThresholds = AlertThresholds{
	CPUPercent:  90,
	RAMPercent:  90,
	DiskPercent: 90,
}

// EvaluateAndFire checks metric thresholds and fires webhooks asynchronously.
func (s *WebhookService) EvaluateAndFire(serverID int64, metric model.ServerMetric) {
	cfg, err := s.repos.Webhook.ListEnabledConfigs()
	if err != nil || len(cfg) == 0 {
		return
	}

	alerts := s.evaluateThresholds(metric)
	if len(alerts) == 0 {
		return
	}

	for _, alertType := range alerts {
		for _, c := range cfg {
			go s.fireWebhook(c, alertType, serverID, metric)
		}
	}
}

func (s *WebhookService) evaluateThresholds(metric model.ServerMetric) []string {
	var alerts []string

	if metric.CPUPercent >= DefaultThresholds.CPUPercent {
		alerts = append(alerts, "cpu_high")
	}
	if metric.RAMTotalMB > 0 {
		ramPct := (metric.RAMUsedMB / metric.RAMTotalMB) * 100
		if ramPct >= DefaultThresholds.RAMPercent {
			alerts = append(alerts, "ram_high")
		}
	}
	if metric.DiskTotalGB > 0 {
		diskPct := (metric.DiskUsedGB / metric.DiskTotalGB) * 100
		if diskPct >= DefaultThresholds.DiskPercent {
			alerts = append(alerts, "disk_high")
		}
	}
	return alerts
}

type webhookPayload struct {
	Event     string             `json:"event"`
	ServerID  int64              `json:"server_id"`
	Timestamp string             `json:"timestamp"`
	Metric    model.ServerMetric `json:"metric"`
}

func (s *WebhookService) fireWebhook(cfg model.WebhookConfig, eventType string, serverID int64, metric model.ServerMetric) {
	payload := webhookPayload{
		Event:     eventType,
		ServerID:  serverID,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Metric:    metric,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		slog.Error("webhook marshal", "error", err)
		return
	}

	req, err := http.NewRequest(http.MethodPost, cfg.URL, bytes.NewReader(body))
	if err != nil {
		slog.Error("webhook request", "error", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	// HMAC-SHA256 signature
	if cfg.Secret != "" {
		mac := hmac.New(sha256.New, []byte(cfg.Secret))
		mac.Write(body)
		sig := hex.EncodeToString(mac.Sum(nil))
		req.Header.Set("X-Webhook-Signature-256", sig)
	}

	resp, err := s.http.Do(req)
	if err != nil {
		slog.Error("webhook fire", "event", eventType, "url", cfg.URL, "error", err)
		_ = s.recordEvent(cfg.ID, eventType, serverID, string(body), 0, err.Error())
		return
	}
	defer resp.Body.Close()

	buf := new(bytes.Buffer)
	_, _ = buf.ReadFrom(resp.Body)
	_ = s.recordEvent(cfg.ID, eventType, serverID, string(body), resp.StatusCode, buf.String())
}

// FireAlert fires a generic event (no metric payload) to all enabled webhooks.
// Used for stale-server alerts and similar non-metric notifications.
func (s *WebhookService) FireAlert(serverID int64, eventType string, message string) {
	cfgs, err := s.repos.Webhook.ListEnabledConfigs()
	if err != nil || len(cfgs) == 0 {
		return
	}
	for _, c := range cfgs {
		go func(cfg model.WebhookConfig) {
			payload := map[string]string{
				"event":     eventType,
				"message":   message,
				"timestamp": time.Now().UTC().Format(time.RFC3339),
			}
			body, err := json.Marshal(payload)
			if err != nil {
				slog.Error("fire alert marshal", "error", err)
				return
			}
			req, err := http.NewRequest(http.MethodPost, cfg.URL, bytes.NewReader(body))
			if err != nil {
				slog.Error("fire alert request", "error", err)
				return
			}
			req.Header.Set("Content-Type", "application/json")
			if cfg.Secret != "" {
				mac := hmac.New(sha256.New, []byte(cfg.Secret))
				mac.Write(body)
				sig := hex.EncodeToString(mac.Sum(nil))
				req.Header.Set("X-Webhook-Signature-256", sig)
			}
			resp, err := s.http.Do(req)
			if err != nil {
				slog.Error("fire alert", "event", eventType, "url", cfg.URL, "error", err)
				_ = s.recordEvent(cfg.ID, eventType, serverID, string(body), 0, err.Error())
				return
			}
			defer resp.Body.Close()
			buf := new(bytes.Buffer)
			_, _ = buf.ReadFrom(resp.Body)
			_ = s.recordEvent(cfg.ID, eventType, serverID, string(body), resp.StatusCode, buf.String())
		}(c)
	}
}

func (s *WebhookService) recordEvent(webhookID int64, eventType string, serverID int64, payload string, code int, body string) error {
	evt := &model.WebhookEvent{
		WebhookID:    webhookID,
		EventType:    eventType,
		ServerID:     serverID,
		Payload:      truncateString(payload, 1000),
		ResponseCode: code,
		ResponseBody: truncateString(body, 1000),
	}
	return s.repos.Webhook.InsertEvent(evt)
}

func truncateString(s string, maxLen int) string {
	r := []rune(s)
	if len(r) > maxLen {
		return string(r[:maxLen])
	}
	return s
}

// formatAlertValue returns a human-readable alert value string.
func formatAlertValue(metric model.ServerMetric, alertType string) string {
	switch alertType {
	case "cpu_high":
		return fmt.Sprintf("%.1f%%", metric.CPUPercent)
	case "ram_high":
		if metric.RAMTotalMB > 0 {
			pct := (metric.RAMUsedMB / metric.RAMTotalMB) * 100
			return fmt.Sprintf("%.1f%% (%.0f/%.0f MB)", pct, metric.RAMUsedMB, metric.RAMTotalMB)
		}
		return fmt.Sprintf("%.0f MB used", metric.RAMUsedMB)
	case "disk_high":
		if metric.DiskTotalGB > 0 {
			pct := (metric.DiskUsedGB / metric.DiskTotalGB) * 100
			return fmt.Sprintf("%.1f%% (%.1f/%.1f GB)", pct, metric.DiskUsedGB, metric.DiskTotalGB)
		}
		return fmt.Sprintf("%.1f GB used", metric.DiskUsedGB)
	default:
		return ""
	}
}

// IDText converts int64 to string for templates.
func IDText(id int64) string {
	return strconv.FormatInt(id, 10)
}

// CountText returns "N" as string for template display.
func CountText(n int) string {
	return strconv.Itoa(n)
}
