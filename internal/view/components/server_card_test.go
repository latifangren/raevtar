package components

import (
	"os"
	"strings"
	"testing"
)

func TestServerCardStatusUsesDurationSemantics(t *testing.T) {
	source, err := os.ReadFile("server_card.templ")
	if err != nil {
		t.Fatalf("read server card template: %v", err)
	}
	content := string(source)

	if count := strings.Count(content, "< 3*time.Minute"); count != 2 {
		t.Fatalf("online threshold count = %d, want 2", count)
	}
	if count := strings.Count(content, "< 15*time.Minute"); count != 2 {
		t.Fatalf("stale threshold count = %d, want 2", count)
	}
	for _, oldThreshold := range []string{"< 2*time.Minute", "< 10*time.Minute"} {
		if strings.Contains(content, oldThreshold) {
			t.Fatalf("server card template still contains old threshold %q", oldThreshold)
		}
	}
}
