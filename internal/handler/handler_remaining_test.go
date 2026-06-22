package handler

import (
	"net/http/httptest"
	"strings"
	"testing"

	"raevtar/internal/model"
)

// ---------- nodeStatusShortcode ----------

func TestNodeStatusShortcodeExisting(t *testing.T) {
	app := newPublicTestApp(t)
	status, body := getBody(t, app, "/lab/node-status/whyred", nil)
	if status != 200 {
		t.Fatalf("status = %d, want 200; body: %s", status, body)
	}
	assertContains(t, body, "whyred")
	assertContains(t, body, "Live Node Status")
	assertContains(t, body, "Offline")
}

func TestNodeStatusShortcodeNotFound(t *testing.T) {
	app := newPublicTestApp(t)
	status, body := getBody(t, app, "/lab/node-status/nonexistent", nil)
	if status != 200 {
		t.Fatalf("status = %d, want 200; body: %s", status, body)
	}
	assertContains(t, body, "Node \"nonexistent\" not found")
	assertContains(t, body, "bg-retro-blush")
}

func TestNodeStatusShortcodeWithMetrics(t *testing.T) {
	app := newPublicTestApp(t)
	if err := app.svc.Monitor.RecordMetrics(app.serverID, model.ServerMetric{
		CPUPercent:    42.0,
		RAMUsedMB:     512,
		RAMTotalMB:    2048,
		DiskUsedGB:    100,
		UptimeSeconds: 7200,
		Online:        true,
	}); err != nil {
		t.Fatalf("record metrics: %v", err)
	}
	status, body := getBody(t, app, "/lab/node-status/whyred", nil)
	if status != 200 {
		t.Fatalf("status = %d, want 200; body: %s", status, body)
	}
	assertContains(t, body, "whyred")
	assertContains(t, body, "Online")
	assertContains(t, body, "bg-retro-sage")
}

// ---------- projectChangelogPage ----------

func TestProjectChangelogPageExisting(t *testing.T) {
	app := newPublicTestApp(t)
	status, body := getBody(t, app, "/projects/whyred-watchtower/changelog", nil)
	if status != 200 {
		t.Fatalf("status = %d, want 200; body: %s", status, body)
	}
	assertContains(t, body, "Changelog")
	assertContains(t, body, "Whyred Watchtower")
}

func TestProjectChangelogPageNonexistentSlug(t *testing.T) {
	app := newPublicTestApp(t)
	status, body := getBody(t, app, "/projects/nonexistent/changelog", nil)
	if status != 404 {
		t.Fatalf("status = %d, want 404; body: %s", status, body)
	}
	assertContains(t, body, "Project not found")
}

// ---------- serveBlogOGImage + serveProjectOGImage ----------

func TestServeBlogOGImageExisting(t *testing.T) {
	app := newPublicTestApp(t)
	req := httptest.NewRequest("GET", "/og-image/blog/hello-raevtar", nil)
	rr := httptest.NewRecorder()
	app.handler.ServeHTTP(rr, req)
	if rr.Code != 200 {
		t.Fatalf("status = %d, want 200; body: %s", rr.Code, rr.Body.String())
	}
	ct := rr.Header().Get("Content-Type")
	if !strings.HasPrefix(ct, "image/svg+xml") {
		t.Fatalf("Content-Type = %q, want image/svg+xml", ct)
	}
	body := rr.Body.String()
	assertContains(t, body, "Hello Raevtar")
	assertContains(t, body, "<svg")
	assertContains(t, body, "Blog dispatch from Raevtar")
}

func TestServeBlogOGImageNotFound(t *testing.T) {
	app := newPublicTestApp(t)
	req := httptest.NewRequest("GET", "/og-image/blog/nonexistent-post", nil)
	rr := httptest.NewRecorder()
	app.handler.ServeHTTP(rr, req)
	if rr.Code != 404 {
		t.Fatalf("status = %d, want 404; body: %s", rr.Code, rr.Body.String())
	}
}

func TestServeProjectOGImageExisting(t *testing.T) {
	app := newPublicTestApp(t)
	req := httptest.NewRequest("GET", "/og-image/project/whyred-watchtower", nil)
	rr := httptest.NewRecorder()
	app.handler.ServeHTTP(rr, req)
	if rr.Code != 200 {
		t.Fatalf("status = %d, want 200; body: %s", rr.Code, rr.Body.String())
	}
	ct := rr.Header().Get("Content-Type")
	if !strings.HasPrefix(ct, "image/svg+xml") {
		t.Fatalf("Content-Type = %q, want image/svg+xml", ct)
	}
	body := rr.Body.String()
	assertContains(t, body, "Whyred Watchtower")
	assertContains(t, body, "Project from small-machine lab")
}

// ---------- formatBytes ----------

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		kb   uint64
		want string
	}{
		{kb: 0, want: "0 KB"},
		{kb: 1, want: "1 KB"},
		{kb: 1023, want: "1023 KB"},
		{kb: 1024, want: "1.0 MB"},
		{kb: 1536, want: "1.5 MB"},
		{kb: 2048, want: "2.0 MB"},
		{kb: 1048576, want: "1.0 GB"},
		{kb: 1572864, want: "1.5 GB"},
		{kb: 1073741824, want: "1.0 TB"},
	}
	for _, tt := range tests {
		got := formatBytes(tt.kb)
		if got != tt.want {
			t.Errorf("formatBytes(%d) = %q, want %q", tt.kb, got, tt.want)
		}
	}
}

func TestFormatBytesRounding(t *testing.T) {
	tests := []struct {
		kb   uint64
		want string
	}{
		{kb: 1025, want: "1.0 MB"},
		{kb: 1049, want: "1.0 MB"},
		{kb: 1050, want: "1.0 MB"},
		{kb: 1075, want: "1.0 MB"},
		{kb: 1126, want: "1.1 MB"},
		{kb: 2047, want: "2.0 MB"},
		{kb: 2049, want: "2.0 MB"},
		{kb: 2560, want: "2.5 MB"},
		{kb: 1048576 + 524288, want: "1.5 GB"},
		{kb: 1048576 + 104858, want: "1.1 GB"},
	}
	for _, tt := range tests {
		got := formatBytes(tt.kb)
		if got != tt.want {
			t.Errorf("formatBytes(%d) = %q, want %q", tt.kb, got, tt.want)
		}
	}
}

// ---------- splitTags ----------

func TestSplitTags(t *testing.T) {
	tests := []struct {
		input string
		want  []string
	}{
		{input: "a,b,c", want: []string{"a", "b", "c"}},
		{input: "a, b, c", want: []string{"a", "b", "c"}},
		{input: "", want: nil},
		{input: "single", want: []string{"single"}},
		{input: "a,,b", want: []string{"a", "b"}},
		{input: ",,", want: nil},
	}
	for _, tt := range tests {
		got := splitTags(tt.input)
		if len(got) != len(tt.want) {
			t.Errorf("splitTags(%q) = %v (len=%d), want %v", tt.input, got, len(got), tt.want)
			continue
		}
		for i := range got {
			if got[i] != tt.want[i] {
				t.Errorf("splitTags(%q)[%d] = %q, want %q", tt.input, i, got[i], tt.want[i])
			}
		}
	}
}

func TestSplitTagsTrimsWhitespace(t *testing.T) {
	tests := []struct {
		input string
		want  []string
	}{
		{input: "  a  ,  b  ", want: []string{"a", "b"}},
		{input: "a , b , c ", want: []string{"a", "b", "c"}},
		{input: "\talpha\n,\tbeta ", want: []string{"alpha", "beta"}},
		{input: "  spaced out  ,  tag  ", want: []string{"spaced out", "tag"}},
	}
	for _, tt := range tests {
		got := splitTags(tt.input)
		if len(got) != len(tt.want) {
			t.Errorf("splitTags(%q) = %v (len=%d), want %v", tt.input, got, len(got), tt.want)
			continue
		}
		for i := range got {
			if got[i] != tt.want[i] {
				t.Errorf("splitTags(%q)[%d] = %q, want %q", tt.input, i, got[i], tt.want[i])
			}
		}
	}
}

// ---------- xmlEscape ----------

func TestXMLEscape(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{input: "&", want: "&amp;"},
		{input: "<", want: "&lt;"},
		{input: ">", want: "&gt;"},
		{input: "\u0022", want: "&quot;"},
		{input: "'", want: "&apos;"},
		{input: "Tom & Jerry <3", want: "Tom &amp; Jerry &lt;3"},
		{input: "<a href=\"x\">", want: "&lt;a href=&quot;x&quot;&gt;"},
		{input: "it's", want: "it&apos;s"},
		{input: "a & b < c > d \"e\" 'f'", want: "a &amp; b &lt; c &gt; d &quot;e&quot; &apos;f&apos;"},
	}
	for _, tt := range tests {
		got := xmlEscape(tt.input)
		if got != tt.want {
			t.Errorf("xmlEscape(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestXMLEscapeNormalText(t *testing.T) {
	tests := []string{
		"hello world",
		"abc123",
		"Hello, World!",
		"Normal text with spaces and punctuation.",
		"1234567890",
	}
	for _, input := range tests {
		got := xmlEscape(input)
		if got != input {
			t.Errorf("xmlEscape(%q) = %q, want unmodified input", input, got)
		}
	}
}

func TestXMLEscapeEmptyString(t *testing.T) {
	got := xmlEscape("")
	if got != "" {
		t.Errorf("xmlEscape('') = %q, want empty string", got)
	}
}
