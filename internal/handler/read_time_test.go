package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"raevtar/internal/repo"
	adminview "raevtar/internal/view/admin"
)

func TestAPIIncrementPostReadTime(t *testing.T) {
	app := newPublicTestApp(t)

	// Fetch posts to find a valid post ID
	posts, _, err := app.svc.Blog.ListPosts("", 1, 10)
	if err != nil {
		t.Fatalf("list posts: %v", err)
	}
	if len(posts) == 0 {
		t.Fatalf("no posts seeded")
	}
	postID := posts[0].ID
	postIDStr := strconv.FormatInt(postID, 10)

	// Table test cases
	tests := []struct {
		name       string
		postIDText string
		payload    any
		wantStatus int
	}{
		{
			name:       "valid increment",
			postIDText: postIDStr,
			payload:    map[string]any{"seconds": 15},
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "invalid post ID",
			postIDText: "invalid",
			payload:    map[string]any{"seconds": 15},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid seconds negative",
			postIDText: postIDStr,
			payload:    map[string]any{"seconds": -5},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid seconds zero",
			postIDText: postIDStr,
			payload:    map[string]any{"seconds": 0},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid seconds over cap",
			postIDText: postIDStr,
			payload:    map[string]any{"seconds": 65},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid JSON",
			postIDText: postIDStr,
			payload:    "invalid-json-body",
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var bodyReader *strings.Reader
			if s, ok := tt.payload.(string); ok {
				bodyReader = strings.NewReader(s)
			} else {
				data, _ := json.Marshal(tt.payload)
				bodyReader = strings.NewReader(string(data))
			}

			req := httptest.NewRequest(http.MethodPost, "/api/v1/posts/"+tt.postIDText+"/read-time", bodyReader)
			req.Header.Set("Content-Type", "application/json")
			req.RemoteAddr = "192.168.1.1:1234"

			rr := httptest.NewRecorder()
			app.handler.ServeHTTP(rr, req)

			if rr.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d (body: %s)", rr.Code, tt.wantStatus, rr.Body.String())
			}
		})
	}

	// Verify increment accumulation
	readTimes, err := app.svc.Blog.AllPostAverageReadTimes()
	if err != nil {
		t.Fatalf("get average read times: %v", err)
	}
	if readTimes[postID] != 15 {
		t.Fatalf("expected post read time of 15, got %d", readTimes[postID])
	}

	// Send another from different IP hash to test average calculation
	payload := mustMarshal(t, map[string]any{"seconds": 45})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/posts/"+postIDStr+"/read-time", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	req.RemoteAddr = "192.168.1.5:1234" // different host

	rr := httptest.NewRecorder()
	app.handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want 204", rr.Code)
	}

	// Average of (15 + 45) / 2 = 30 seconds
	readTimes, err = app.svc.Blog.AllPostAverageReadTimes()
	if err != nil {
		t.Fatalf("get average read times: %v", err)
	}
	if readTimes[postID] != 30 {
		t.Fatalf("expected average post read time of 30, got %d", readTimes[postID])
	}
}

func TestFormatReadTime(t *testing.T) {
	tests := []struct {
		sec  int
		want string
	}{
		{0, "-"},
		{-1, "-"},
		{45, "45s"},
		{60, "1m 0s"},
		{75, "1m 15s"},
		{125, "2m 5s"},
	}

	for _, tt := range tests {
		got := adminview.FormatReadTime(tt.sec)
		if got != tt.want {
			t.Errorf("FormatReadTime(%d) = %q, want %q", tt.sec, got, tt.want)
		}
	}
}

func TestViewRepoReadTimes(t *testing.T) {
	app := newPublicTestApp(t)

	// Test repo directly with seed post
	posts, _, _ := app.svc.Blog.ListPosts("", 1, 1)
	postID := posts[0].ID

	viewRepo := repo.New(app.db).View

	err := viewRepo.IncrementPostReadTime(postID, "hash1", 10)
	if err != nil {
		t.Fatalf("increment read time: %v", err)
	}
	err = viewRepo.IncrementPostReadTime(postID, "hash1", 10) // conflict update
	if err != nil {
		t.Fatalf("increment read time: %v", err)
	}
	err = viewRepo.IncrementPostReadTime(postID, "hash2", 40)
	if err != nil {
		t.Fatalf("increment read time: %v", err)
	}

	res, err := viewRepo.AveragePostReadTimes()
	if err != nil {
		t.Fatalf("average read times: %v", err)
	}
	// Average: (20 + 40) / 2 = 30
	if res[postID] != 30 {
		t.Fatalf("expected avg read time to be 30, got %d", res[postID])
	}
}
