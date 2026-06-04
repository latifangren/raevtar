package admin

import (
	"testing"
	"time"

	"raevtar/internal/model"
)

func TestStatusTextUsesDurationSemantics(t *testing.T) {
	now := time.Now()
	online := now.Add(-2 * time.Minute)
	stale := now.Add(-14 * time.Minute)
	offline := now.Add(-16 * time.Minute)

	if got := StatusText(&online); got != "Online" {
		t.Fatalf("online status = %q", got)
	}
	if got := StatusText(&stale); got != "Stale" {
		t.Fatalf("stale status = %q", got)
	}
	if got := StatusText(&offline); got != "Offline" {
		t.Fatalf("offline status = %q", got)
	}
	if got := StatusText(nil); got != "Offline" {
		t.Fatalf("nil status = %q", got)
	}
}

func TestLastPayloadSummaryText(t *testing.T) {
	now := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	metrics := []model.ServerMetric{{
		CPUPercent:    12.5,
		RAMUsedMB:     256,
		RAMTotalMB:    1024,
		DiskUsedGB:    8.5,
		UptimeSeconds: 3600,
		Online:        true,
		RecordedAt:    now.Add(-2 * time.Minute),
	}}

	got := LastPayloadSummaryText(metrics, now)
	want := "CPU 12.5% · RAM 256.0 / 1024.0 MB · Disk 8.5 GB · Uptime 1h 0m · 2m ago"
	if got != want {
		t.Fatalf("summary = %q, want %q", got, want)
	}
}

func TestMetricTimelineTextDetectsTransitions(t *testing.T) {
	metrics := []model.ServerMetric{{Online: true}, {Online: false}}
	if got := MetricTimelineText(metrics, 0); got != "Offline to Online" {
		t.Fatalf("transition = %q", got)
	}
	if got := MetricTimelineText(metrics, 1); got != "Offline sample recorded" {
		t.Fatalf("oldest = %q", got)
	}
}

func TestEditorialBadgeHelpers(t *testing.T) {
	if got := EditorialStatusBadgeClass(model.EditorialStatusApproved); got == "" {
		t.Fatalf("expected approved badge class")
	}
	if got := EditorialModeBadgeClass(model.EditorialModeScheduled); got == "" {
		t.Fatalf("expected scheduled badge class")
	}
}

func TestEditorialPhase4Helpers(t *testing.T) {
	now := time.Date(2026, 6, 5, 12, 0, 0, 0, time.UTC)
	overdue := now.Add(-45 * time.Minute)
	if got := EditorialEscalationText(&overdue, model.EditorialStatusApproved, now); got != "Overdue" {
		t.Fatalf("escalation text = %q, want Overdue", got)
	}
	if got := EditorialEscalationBadgeClass(&overdue, model.EditorialStatusApproved, now); got == "" {
		t.Fatalf("expected overdue escalation badge class")
	}
	if got := EditorialFairnessStateText(3, true); got != "Autonomous slot open next claim" {
		t.Fatalf("fairness state text = %q", got)
	}
	if got := EditorialDurationText(95 * time.Minute); got != "1h 35m" {
		t.Fatalf("duration text = %q, want 1h 35m", got)
	}
}

func TestEditorialItemMutabilityHelpers(t *testing.T) {
	mutable := model.EditorialInboxItem{Status: model.EditorialStatusApproved, AttemptCount: 0}
	if !EditorialItemMutable(mutable) {
		t.Fatalf("expected approved item with zero attempts to be mutable")
	}
	locked := model.EditorialInboxItem{Status: model.EditorialStatusApproved, AttemptCount: 1}
	if EditorialItemMutable(locked) {
		t.Fatalf("expected attempted item to be locked")
	}
	if got := EditorialItemLockedReason(locked); got != "Locked after first execution attempt" {
		t.Fatalf("locked reason = %q", got)
	}
}
