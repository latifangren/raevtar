package admin

import (
	"testing"
	"time"
)

func TestStatusTextUsesDurationSemantics(t *testing.T) {
	now := time.Now()
	online := now.Add(-time.Minute)
	stale := now.Add(-5 * time.Minute)
	offline := now.Add(-11 * time.Minute)

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
