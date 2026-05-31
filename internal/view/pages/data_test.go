package pages

import (
	"testing"
	"time"
)

func TestServerStatusTextUsesDurationSemantics(t *testing.T) {
	now := time.Now()
	online := now.Add(-time.Minute)
	stale := now.Add(-5 * time.Minute)
	offline := now.Add(-11 * time.Minute)

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
