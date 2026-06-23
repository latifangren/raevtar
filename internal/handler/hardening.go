package handler

import (
	"errors"
	"log/slog"
	"net/http"
	"raevtar/internal/config"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	loginBodyLimit          int64 = 16 << 10
	apiBodyLimit            int64 = 1 << 20
	adminFormBodyLimit      int64 = 1 << 20
	mediaUploadBodyLimit    int64 = 6 << 20
	loginFailureLimit             = 5
	loginIPFailureLimit           = 20
	loginFailureWindow            = 5 * time.Minute
	loginThrottleRetryAfter       = 60
)

// initHardening configures body limits and login throttling from the application config.
// Called once at startup from handler.New().
func initHardening(cfg *config.Config) {
	if cfg.MaxUploadMB > 0 {
		mediaUploadBodyLimit = int64(cfg.MaxUploadMB) << 20
	}
	if cfg.LoginFailureLimit > 0 {
		loginFailureLimit = cfg.LoginFailureLimit
	}
	if cfg.LoginIPFailureLimit > 0 {
		loginIPFailureLimit = cfg.LoginIPFailureLimit
	}
}

type loginThrottleStore struct {
	mu       sync.Mutex
	failures map[string]loginFailureEntry
}

type loginFailureEntry struct {
	count   int
	resetAt time.Time
}

var loginThrottles = &loginThrottleStore{failures: make(map[string]loginFailureEntry)}

func capRequestBody(w http.ResponseWriter, r *http.Request, limit int64) {
	r.Body = http.MaxBytesReader(w, r.Body, limit)
}

func parseCappedForm(w http.ResponseWriter, r *http.Request, limit int64) bool {
	capRequestBody(w, r, limit)
	var err error
	if strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data") {
		err = r.ParseMultipartForm(64 << 10)
	} else {
		err = r.ParseForm()
	}
	if err == nil {
		return true
	}
	var maxBytesErr *http.MaxBytesError
	if errors.As(err, &maxBytesErr) {
		http.Error(w, "request body too large", http.StatusRequestEntityTooLarge)
		return false
	}
	http.Error(w, "invalid form", http.StatusBadRequest)
	return false
}

func adminRequestBodyLimit(r *http.Request) int64 {
	if r.URL.Path == "/admin/media" {
		return mediaUploadBodyLimit
	}
	return adminFormBodyLimit
}

func isBodyTooLarge(err error) bool {
	var maxBytesErr *http.MaxBytesError
	return errors.As(err, &maxBytesErr)
}

func internalServerError(w http.ResponseWriter, r *http.Request, err error) {
	logHandlerError(r, err)
	http.Error(w, "internal server error", http.StatusInternalServerError)
}

func internalServerJSON(w http.ResponseWriter, r *http.Request, err error) {
	logHandlerError(r, err)
	writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
}

func warnAfterMutation(r *http.Request, action string, err error) {
	if err == nil {
		return
	}
	slog.Warn("post-mutation audit failed", "action", action, "method", r.Method, "path", r.URL.Path, "error", err)
}

func logHandlerError(r *http.Request, err error) {
	if err == nil {
		return
	}
	slog.Error("handler internal error", "method", r.Method, "path", r.URL.Path, "error", err)
}

func loginThrottleKey(ip, username string) string {
	return "user\x00" + ip + "\x00" + normalizedLoginUsername(username)
}

func loginIPThrottleKey(ip string) string {
	return "ip\x00" + ip
}

func normalizedLoginUsername(username string) string {
	return strings.ToLower(strings.TrimSpace(username))
}

func loginThrottleActive(ip, username string) bool {
	return loginBucketActive(loginThrottleKey(ip, username), loginFailureLimit) || loginBucketActive(loginIPThrottleKey(ip), loginIPFailureLimit)
}

func loginBucketActive(key string, limit int) bool {
	now := time.Now()
	loginThrottles.mu.Lock()
	defer loginThrottles.mu.Unlock()
	entry, ok := loginThrottles.failures[key]
	if !ok {
		return false
	}
	if now.After(entry.resetAt) {
		delete(loginThrottles.failures, key)
		return false
	}
	return entry.count >= limit
}

func recordLoginFailure(ip, username string) {
	recordLoginFailureKey(loginThrottleKey(ip, username))
	recordLoginFailureKey(loginIPThrottleKey(ip))
}

func recordLoginFailureKey(key string) {
	now := time.Now()
	loginThrottles.mu.Lock()
	defer loginThrottles.mu.Unlock()
	entry, ok := loginThrottles.failures[key]
	if !ok || now.After(entry.resetAt) {
		loginThrottles.failures[key] = loginFailureEntry{count: 1, resetAt: now.Add(loginFailureWindow)}
		return
	}
	entry.count++
	loginThrottles.failures[key] = entry
}

func clearLoginFailures(ip, username string) {
	loginThrottles.mu.Lock()
	defer loginThrottles.mu.Unlock()
	delete(loginThrottles.failures, loginThrottleKey(ip, username))
}

func writeLoginThrottle(w http.ResponseWriter) {
	w.Header().Set("Retry-After", strconv.Itoa(loginThrottleRetryAfter))
	writeJSON(w, http.StatusTooManyRequests, map[string]string{"error": "too many login attempts"})
}
