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
