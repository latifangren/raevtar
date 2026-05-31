package components

import (
	"testing"
	"time"
)

func TestServerCardStatusUsesDurationSemantics(t *testing.T) {
	now := time.Now()
	online := now.Add(-time.Minute)
	stale := now.Add(-5 * time.Minute)
	offline := now.Add(-11 * time.Minute)

	if got := serverStatusText(&online); got != "Online" {
		t.Fatalf("online status = %q", got)
	}
	if got := serverStatusText(&stale); got != "Stale" {
		t.Fatalf("stale status = %q", got)
	}
	if got := serverStatusText(&offline); got != "Offline" {
		t.Fatalf("offline status = %q", got)
	}
	if got := serverStatusText(nil); got != "Offline" {
		t.Fatalf("nil status = %q", got)
	}
}
