package pages

import (
	"testing"
	"time"

	"raevtar/internal/model"
)

func TestServerStatusTextUsesDurationSemantics(t *testing.T) {
	now := time.Now()
	online := now.Add(-2 * time.Minute)
	stale := now.Add(-14 * time.Minute)
	offline := now.Add(-16 * time.Minute)

	if got := ServerStatusText(&online); got != "Online" {
		t.Fatalf("online status = %q", got)
	}
	if got := ServerStatusText(&stale); got != "Stale" {
		t.Fatalf("stale status = %q", got)
	}
	if got := ServerStatusText(&offline); got != "Offline" {
		t.Fatalf("offline status = %q", got)
	}
	if got := ServerStatusText(nil); got != "Offline" {
		t.Fatalf("nil status = %q", got)
	}
}

func TestServerFreshnessHelpersAreDeterministic(t *testing.T) {
	now := time.Date(2026, 5, 31, 12, 0, 0, 0, time.UTC)
	lastSeen := now.Add(-4*time.Minute - 30*time.Second)
	metrics := []model.ServerMetric{{RecordedAt: now.Add(-5 * time.Minute)}}

	if got := AgeText(now.Sub(lastSeen)); got != "4m" {
		t.Fatalf("age text = %q", got)
	}
	if got := LastSignalAgeText(&lastSeen, now); got != "4m ago" {
		t.Fatalf("last signal age = %q", got)
	}
	if got := LatestMetricTimestampText(metrics); got != "May 31 11:55:00 UTC" {
		t.Fatalf("latest metric text = %q", got)
	}
	if got := RefreshTimeText(now); got != "May 31 12:00:00 UTC" {
		t.Fatalf("refresh time = %q", got)
	}
	if got := FreshnessCauseHint(&lastSeen, metrics, now); got != "Telemetry is delayed. Agent may be between scheduled reports." {
		t.Fatalf("freshness hint = %q", got)
	}
}

func TestFreshnessCauseHintStates(t *testing.T) {
	now := time.Date(2026, 5, 31, 12, 0, 0, 0, time.UTC)
	metrics := []model.ServerMetric{{RecordedAt: now.Add(-time.Minute)}}
	fresh := now.Add(-2 * time.Minute)
	offline := now.Add(-16 * time.Minute)
	missingMetrics := now.Add(-time.Minute)

	tests := []struct {
		name     string
		lastSeen *time.Time
		metrics  []model.ServerMetric
		want     string
	}{
		{name: "never seen", want: "No telemetry received yet. Waiting for first agent signal."},
		{name: "fresh", lastSeen: &fresh, metrics: metrics, want: "Telemetry is fresh. Latest agent signal arrived recently."},
		{name: "offline", lastSeen: &offline, metrics: metrics, want: "Telemetry is offline. No recent agent signal has reached Raevtar."},
		{name: "missing metrics", lastSeen: &missingMetrics, want: "Signal timestamp exists, but no metrics sample is stored yet."},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FreshnessCauseHint(tt.lastSeen, tt.metrics, now); got != tt.want {
				t.Fatalf("hint = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDashboardFreshnessReasonAtExplainsStatusWindows(t *testing.T) {
	now := time.Date(2026, 5, 31, 12, 0, 0, 0, time.UTC)
	fresh := now.Add(-2 * time.Minute)
	stale := now.Add(-8 * time.Minute)
	offline := now.Add(-16 * time.Minute)

	tests := []struct {
		name     string
		lastSeen *time.Time
		want     string
	}{
		{name: "never seen", want: "Why: no last agent signal has reached Raevtar yet."},
		{name: "fresh", lastSeen: &fresh, want: "Why: last agent signal is inside the <3m online window."},
		{name: "stale", lastSeen: &stale, want: "Why: last agent signal is older than 3m but still inside the 15m stale window."},
		{name: "offline", lastSeen: &offline, want: "Why: last agent signal is older than 15m, so public status is offline."},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := DashboardFreshnessReasonAt(tt.lastSeen, now); got != tt.want {
				t.Fatalf("reason = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestMetricHistoryHelpers(t *testing.T) {
	now := time.Date(2026, 5, 31, 12, 0, 0, 0, time.UTC)
	metrics := []model.ServerMetric{
		{RecordedAt: now.Add(-1 * time.Minute)},
		{RecordedAt: now.Add(-6 * time.Minute)},
	}

	if got := MetricSampleCountText(metrics); got != "2 samples" {
		t.Fatalf("sample count = %q", got)
	}
	if got := MetricWindowText(metrics); got != "5m captured" {
		t.Fatalf("metric window = %q", got)
	}
	if got := MetricRecordedAgeText(metrics[0].RecordedAt, now); got != "1m ago" {
		t.Fatalf("metric age = %q", got)
	}
	if got := MetricWindowText(nil); got != "No history yet" {
		t.Fatalf("empty metric window = %q", got)
	}
}

func TestNodeSystemHealthHelpersPreserveOldPayloadValues(t *testing.T) {
	metric := model.ServerMetric{
		CPUPercent:    12.5,
		RAMUsedMB:     256,
		RAMTotalMB:    1024,
		DiskUsedGB:    8.5,
		UptimeSeconds: 3600,
	}

	for name, tt := range map[string]struct {
		got  string
		want string
	}{
		"cpu load":    {got: CPULoadText(metric), want: "N/A"},
		"cpu cores":   {got: CPUCoresText(metric), want: "N/A"},
		"ram":         {got: RAMHealthText(metric), want: "256.0 / 1024.0 MB · 25.0%"},
		"disk":        {got: DiskHealthText(metric), want: "8.5 GB / N/A"},
		"temperature": {got: TemperatureText(metric), want: "N/A"},
	} {
		t.Run(name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Fatalf("text = %q, want %q", tt.got, tt.want)
			}
		})
	}
}

func TestNodeSystemHealthHelpersFormatExtendedPayload(t *testing.T) {
	metric := model.ServerMetric{
		CPULoad1:             0,
		CPULoad5:             0.3,
		CPULoad15:            0.2,
		CPUCores:             4,
		DiskUsedGB:           8,
		DiskTotalGB:          32,
		TemperatureC:         48.5,
		TemperatureAvailable: true,
	}

	for name, tt := range map[string]struct {
		got  string
		want string
	}{
		"cpu load":    {got: CPULoadText(metric), want: "0.0 / 0.3 / 0.2"},
		"cpu cores":   {got: CPUCoresText(metric), want: "4"},
		"disk":        {got: DiskHealthText(metric), want: "8.0 / 32.0 GB · 25.0%"},
		"temperature": {got: TemperatureText(metric), want: "48.5°C"},
	} {
		t.Run(name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Fatalf("text = %q, want %q", tt.got, tt.want)
			}
		})
	}
}
