package handler

import (

	"context"
	"errors"
	"fmt"
	"io"

	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"raevtar/internal/config"
)

// ---------------------------------------------------------------------------
// Test helpers
// ---------------------------------------------------------------------------

// nopComponent is a templ.Component that renders successfully.
type nopComponent struct{}

func (nopComponent) Render(_ context.Context, w io.Writer) error {
	_, err := io.WriteString(w, "OK")
	return err
}

// errComponent is a templ.Component that always returns an error.
type errComponent struct{}

func (errComponent) Render(_ context.Context, _ io.Writer) error {
	return fmt.Errorf("render error")
}

func resetRateLimiter() {
	limiter.mu.Lock()
	defer limiter.mu.Unlock()
	limiter.requests = make(map[string]*bucketEntry)
}

func resetLoginThrottles() {
	loginThrottles.mu.Lock()
	defer loginThrottles.mu.Unlock()
	loginThrottles.failures = make(map[string]loginFailureEntry)
}

// ---------------------------------------------------------------------------
// 1. RateLimit middleware
// ---------------------------------------------------------------------------

func TestRateLimitAllowsNormalRequests(t *testing.T) {
	resetRateLimiter()
	handler := RateLimit(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	for i := 0; i < 10; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = "203.0.113.1:1234"
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("request %d: status = %d, want %d", i+1, rr.Code, http.StatusOK)
		}
	}
}

func TestRateLimitBlocksExcessiveRequests(t *testing.T) {
	resetRateLimiter()
	handler := RateLimit(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	ip := "198.51.100.1:9999"
	for i := 0; i < maxRequests; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = ip
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("request %d within limit: status = %d, want %d", i+1, rr.Code, http.StatusOK)
		}
	}

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = ip
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusTooManyRequests {
		t.Fatalf("status = %d, want %d (429)", rr.Code, http.StatusTooManyRequests)
	}
	if rr.Header().Get("Retry-After") != "60" {
		t.Fatalf("Retry-After = %q, want %q", rr.Header().Get("Retry-After"), "60")
	}
	if !strings.Contains(rr.Body.String(), "429") {
		t.Fatalf("body missing 429: %s", rr.Body.String())
	}
}

func TestRateLimitResetsAfterWindow(t *testing.T) {
	resetRateLimiter()
	handler := RateLimit(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	ip := "192.0.2.1:1234"
	for i := 0; i < maxRequests; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = ip
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
	}

	limiter.mu.Lock()
	limiter.requests["192.0.2.1"] = &bucketEntry{
		count:   maxRequests,
		resetAt: time.Now().Add(-time.Minute),
	}
	limiter.mu.Unlock()

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = ip
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("after window reset: status = %d, want %d", rr.Code, http.StatusOK)
	}
}

func TestRateLimitDifferentIPsIndependentBuckets(t *testing.T) {
	resetRateLimiter()
	handler := RateLimit(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	ipA := "10.0.0.1:1234"
	for i := 0; i < maxRequests; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = ipA
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
	}

	reqA := httptest.NewRequest(http.MethodGet, "/test", nil)
	reqA.RemoteAddr = ipA
	rrA := httptest.NewRecorder()
	handler.ServeHTTP(rrA, reqA)
	if rrA.Code != http.StatusTooManyRequests {
		t.Fatalf("IP A should be blocked, got %d", rrA.Code)
	}

	reqB := httptest.NewRequest(http.MethodGet, "/test", nil)
	reqB.RemoteAddr = "10.0.0.2:1234"
	rrB := httptest.NewRecorder()
	handler.ServeHTTP(rrB, reqB)
	if rrB.Code != http.StatusOK {
		t.Fatalf("IP B should not be blocked, got %d", rrB.Code)
	}
}

func TestRateLimitSkipsStaticPaths(t *testing.T) {
	resetRateLimiter()
	handler := RateLimit(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	ip := "10.0.0.1:1234"
	for i := 0; i < maxRequests+20; i++ {
		req := httptest.NewRequest(http.MethodGet, "/static/css/style.css", nil)
		req.RemoteAddr = ip
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("static request %d: status = %d, want %d", i+1, rr.Code, http.StatusOK)
		}
	}
}

// ---------------------------------------------------------------------------
// 2. WithSecurityHeaders
// ---------------------------------------------------------------------------

func TestWithSecurityHeaders(t *testing.T) {
	handler := WithSecurityHeaders(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}

	want := map[string]string{
		"X-Content-Type-Options": "nosniff",
		"X-Frame-Options":        "DENY",
		"X-XSS-Protection":       "1; mode=block",
		"Referrer-Policy":        "strict-origin-when-cross-origin",
		"Content-Security-Policy": "default-src 'self'; style-src 'self' 'unsafe-inline' https://fonts.googleapis.com; font-src 'self' https://fonts.gstatic.com; script-src 'self'; img-src 'self' data:; connect-src 'self'",
	}
	for key, wantVal := range want {
		if got := rr.Header().Get(key); got != wantVal {
			t.Fatalf("header %q = %q, want %q", key, got, wantVal)
		}
	}
}

// ---------------------------------------------------------------------------
// 3. clientIP
// ---------------------------------------------------------------------------

func TestClientIPFromRemoteAddr(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "203.0.113.50:1234"

	ip := clientIP(req)
	if ip != "203.0.113.50" {
		t.Fatalf("clientIP = %q, want %q", ip, "203.0.113.50")
	}
}

func TestClientIPFromRemoteAddrNoPort(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "203.0.113.50"

	ip := clientIP(req)
	if ip != "203.0.113.50" {
		t.Fatalf("clientIP = %q, want %q", ip, "203.0.113.50")
	}
}

func TestClientIPFromXForwardedFor(t *testing.T) {
	configureTrustedProxies(&config.Config{TrustedProxyCIDRs: []string{"192.0.2.0/24"}})
	defer configureTrustedProxies(&config.Config{})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "192.0.2.10:1234"
	req.Header.Set("X-Forwarded-For", "203.0.113.50")

	ip := clientIP(req)
	if ip != "203.0.113.50" {
		t.Fatalf("clientIP = %q, want %q", ip, "203.0.113.50")
	}
}

func TestClientIPFromCFConnectingIP(t *testing.T) {
	configureTrustedProxies(&config.Config{TrustedProxyCIDRs: []string{"192.0.2.0/24"}})
	defer configureTrustedProxies(&config.Config{})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "192.0.2.10:1234"
	req.Header.Set("CF-Connecting-IP", "203.0.113.99")

	ip := clientIP(req)
	if ip != "203.0.113.99" {
		t.Fatalf("clientIP = %q, want %q", ip, "203.0.113.99")
	}
}

func TestClientIPPrefersCFFirst(t *testing.T) {
	configureTrustedProxies(&config.Config{TrustedProxyCIDRs: []string{"192.0.2.0/24"}})
	defer configureTrustedProxies(&config.Config{})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "192.0.2.10:1234"
	req.Header.Set("CF-Connecting-IP", "10.0.0.1")
	req.Header.Set("X-Forwarded-For", "203.0.113.50")

	ip := clientIP(req)
	if ip != "10.0.0.1" {
		t.Fatalf("clientIP = %q, want CF IP %q", ip, "10.0.0.1")
	}
}

func TestClientIPIgnoresHeadersFromUntrustedProxy(t *testing.T) {
	configureTrustedProxies(&config.Config{TrustedProxyCIDRs: []string{"10.0.0.0/8"}})
	defer configureTrustedProxies(&config.Config{})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "192.0.2.10:1234"
	req.Header.Set("CF-Connecting-IP", "203.0.113.99")

	ip := clientIP(req)
	if ip != "192.0.2.10" {
		t.Fatalf("clientIP = %q, want RemoteAddr %q", ip, "192.0.2.10")
	}
}

// ---------------------------------------------------------------------------
// 4. trustedHeaderIP / firstForwardedFor / isTrustedProxy
// ---------------------------------------------------------------------------

func TestTrustedHeaderIPTrustedProxy(t *testing.T) {
	configureTrustedProxies(&config.Config{TrustedProxyCIDRs: []string{"192.0.2.0/24"}})
	defer configureTrustedProxies(&config.Config{})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Forwarded-For", "203.0.113.50")

	ip := trustedHeaderIP("192.0.2.10", req)
	if ip != "203.0.113.50" {
		t.Fatalf("trustedHeaderIP = %q, want %q", ip, "203.0.113.50")
	}
}

func TestTrustedHeaderIPUntrustedProxy(t *testing.T) {
	configureTrustedProxies(&config.Config{TrustedProxyCIDRs: []string{"192.0.2.0/24"}})
	defer configureTrustedProxies(&config.Config{})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Forwarded-For", "203.0.113.50")

	ip := trustedHeaderIP("198.51.100.10", req)
	if ip != "" {
		t.Fatalf("trustedHeaderIP = %q, want empty for untrusted proxy", ip)
	}
}

func TestTrustedHeaderIPNoHeaders(t *testing.T) {
	configureTrustedProxies(&config.Config{TrustedProxyCIDRs: []string{"192.0.2.0/24"}})
	defer configureTrustedProxies(&config.Config{})

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	ip := trustedHeaderIP("192.0.2.10", req)
	if ip != "" {
		t.Fatalf("trustedHeaderIP = %q, want empty when no headers present", ip)
	}
}

func TestFirstForwardedFor(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  string
	}{
		{name: "single IP", value: "203.0.113.50", want: "203.0.113.50"},
		{name: "multiple IPs", value: "203.0.113.50, 198.51.100.10, 192.0.2.20", want: "203.0.113.50"},
		{name: "trailing comma", value: "10.0.0.1, ", want: "10.0.0.1"},
		{name: "empty string", value: "", want: ""},
		{name: "whitespace only", value: "  ", want: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := firstForwardedFor(tt.value)
			if got != tt.want {
				t.Fatalf("firstForwardedFor(%q) = %q, want %q", tt.value, got, tt.want)
			}
		})
	}
}

func TestIsTrustedProxyMatching(t *testing.T) {
	configureTrustedProxies(&config.Config{TrustedProxyCIDRs: []string{"192.0.2.0/24"}})
	defer configureTrustedProxies(&config.Config{})

	if !isTrustedProxy("192.0.2.10") {
		t.Fatalf("isTrustedProxy should match 192.0.2.10 in 192.0.2.0/24")
	}
}

func TestIsTrustedProxyNonMatching(t *testing.T) {
	configureTrustedProxies(&config.Config{TrustedProxyCIDRs: []string{"192.0.2.0/24"}})
	defer configureTrustedProxies(&config.Config{})

	if isTrustedProxy("198.51.100.10") {
		t.Fatalf("isTrustedProxy should NOT match 198.51.100.10")
	}
}

func TestIsTrustedProxyInvalidIP(t *testing.T) {
	configureTrustedProxies(&config.Config{TrustedProxyCIDRs: []string{"192.0.2.0/24"}})
	defer configureTrustedProxies(&config.Config{})

	if isTrustedProxy("not-an-ip") {
		t.Fatalf("isTrustedProxy should return false for invalid IP")
	}
}

func TestIsTrustedProxyEmptyNetworks(t *testing.T) {
	configureTrustedProxies(&config.Config{})

	if isTrustedProxy("192.0.2.10") {
		t.Fatalf("isTrustedProxy should return false when no proxies configured")
	}
}

func TestConfigureTrustedProxiesParsesMultipleCIDRs(t *testing.T) {
	configureTrustedProxies(&config.Config{
		TrustedProxyCIDRs: []string{"10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16"},
	})
	defer configureTrustedProxies(&config.Config{})

	if !isTrustedProxy("10.1.2.3") {
		t.Fatalf("10.1.2.3 should be trusted")
	}
	if !isTrustedProxy("172.16.0.1") {
		t.Fatalf("172.16.0.1 should be trusted")
	}
	if !isTrustedProxy("192.168.1.1") {
		t.Fatalf("192.168.1.1 should be trusted")
	}
	if isTrustedProxy("8.8.8.8") {
		t.Fatalf("8.8.8.8 should NOT be trusted")
	}
}

// ---------------------------------------------------------------------------
// 5. internalServerError / internalServerJSON
// ---------------------------------------------------------------------------

func TestInternalServerError(t *testing.T) {
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)

	internalServerError(rr, req, fmt.Errorf("something broke"))

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusInternalServerError)
	}
	if !strings.Contains(rr.Body.String(), "internal server error") {
		t.Fatalf("body = %q, want generic message", rr.Body.String())
	}
	if strings.Contains(rr.Body.String(), "something broke") {
		t.Fatalf("body leaked internal error details: %s", rr.Body.String())
	}
}

func TestInternalServerJSON(t *testing.T) {
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)

	internalServerJSON(rr, req, fmt.Errorf("db failure"))

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusInternalServerError)
	}
	if ct := rr.Header().Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
		t.Fatalf("Content-Type = %q, want application/json", ct)
	}
	if !strings.Contains(rr.Body.String(), `"internal server error"`) {
		t.Fatalf("body = %q, want json error", rr.Body.String())
	}
	if strings.Contains(rr.Body.String(), "db failure") {
		t.Fatalf("body leaked error details: %s", rr.Body.String())
	}
}

// ---------------------------------------------------------------------------
// 6. capRequestBody / parseCappedForm
// ---------------------------------------------------------------------------

func TestCapRequestBodySmallBodyPasses(t *testing.T) {
	rr := httptest.NewRecorder()
	body := strings.NewReader("small body")
	req := httptest.NewRequest(http.MethodPost, "/test", body)
	capRequestBody(rr, req, 1<<20)

	data, err := io.ReadAll(req.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	if string(data) != "small body" {
		t.Fatalf("body = %q, want %q", string(data), "small body")
	}
}

func TestCapRequestBodyLargeBodyRejected(t *testing.T) {
	rr := httptest.NewRecorder()
	body := strings.NewReader(strings.Repeat("x", 1000))
	req := httptest.NewRequest(http.MethodPost, "/test", body)
	capRequestBody(rr, req, 10)

	_, err := io.ReadAll(req.Body)
	if err == nil {
		t.Fatalf("expected error for oversized body")
	}
	var maxBytesErr *http.MaxBytesError
	if !errors.As(err, &maxBytesErr) {
		t.Fatalf("error type = %T, want *http.MaxBytesError", err)
	}
}

func TestParseCappedFormValid(t *testing.T) {
	rr := httptest.NewRecorder()
	form := url.Values{"username": {"admin"}, "action": {"login"}}
	req := httptest.NewRequest(http.MethodPost, "/admin/login",
		strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	if !parseCappedForm(rr, req, 1<<20) {
		t.Fatalf("parseCappedForm returned false for valid form; body: %s", rr.Body.String())
	}
}

func TestParseCappedFormOversized(t *testing.T) {
	rr := httptest.NewRecorder()
	body := strings.NewReader("key=" + strings.Repeat("x", 500))
	req := httptest.NewRequest(http.MethodPost, "/admin/login", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	ok := parseCappedForm(rr, req, 10)
	if ok {
		t.Fatalf("parseCappedForm returned true for oversized body")
	}
	if rr.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("status = %d, want %d (413); body: %s", rr.Code, http.StatusRequestEntityTooLarge, rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), "request body too large") {
		t.Fatalf("body = %q, want 413 message", rr.Body.String())
	}
}

// ---------------------------------------------------------------------------
// 7. adminRequestBodyLimit / isBodyTooLarge
// ---------------------------------------------------------------------------

func TestAdminRequestBodyLimit(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/admin/posts", nil)
	if got := adminRequestBodyLimit(req); got != adminFormBodyLimit {
		t.Fatalf("admin posts limit = %d, want %d", got, adminFormBodyLimit)
	}

	req = httptest.NewRequest(http.MethodGet, "/admin/media", nil)
	if got := adminRequestBodyLimit(req); got != mediaUploadBodyLimit {
		t.Fatalf("admin media limit = %d, want %d", got, mediaUploadBodyLimit)
	}
}

func TestIsBodyTooLarge(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{name: "MaxBytesError returns true", err: &http.MaxBytesError{Limit: 100}, want: true},
		{name: "nil returns false", err: nil, want: false},
		{name: "generic error returns false", err: fmt.Errorf("parse error"), want: false},
		{name: "io error returns false", err: io.ErrUnexpectedEOF, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isBodyTooLarge(tt.err); got != tt.want {
				t.Fatalf("isBodyTooLarge(%v) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// 8. loginThrottle
// ---------------------------------------------------------------------------

func TestLoginThrottleInactiveOnFirstAttempt(t *testing.T) {
	resetLoginThrottles()

	if loginThrottleActive("203.0.113.1", "testuser") {
		t.Fatalf("throttle should be inactive on first attempt")
	}
}

func TestLoginThrottleActiveAfterFailures(t *testing.T) {
	resetLoginThrottles()

	for i := 0; i < loginFailureLimit; i++ {
		recordLoginFailure("203.0.113.1", "testuser")
	}

	if !loginThrottleActive("203.0.113.1", "testuser") {
		t.Fatalf("throttle should be active after %d failures", loginFailureLimit)
	}
}

func TestLoginThrottleActiveAfterIPLimit(t *testing.T) {
	resetLoginThrottles()

	for i := 0; i < loginFailureLimit; i++ {
		recordLoginFailure("203.0.113.1", fmt.Sprintf("user%d", i))
	}

	if loginThrottleActive("203.0.113.1", "user0") {
		t.Fatalf("throttle should be inactive after 1 failure for user0")
	}

	for i := loginFailureLimit; i < loginIPFailureLimit; i++ {
		recordLoginFailure("203.0.113.1", fmt.Sprintf("user%d", i))
	}

	if !loginThrottleActive("203.0.113.1", "newuser") {
		t.Fatalf("throttle should be active after IP-level limit (%d)", loginIPFailureLimit)
	}
}

func TestRecordLoginFailureIncrementsCounter(t *testing.T) {
	resetLoginThrottles()

	for i := 0; i < 3; i++ {
		recordLoginFailure("10.0.0.1", "victim")
	}

	if !loginBucketActive(loginThrottleKey("10.0.0.1", "victim"), 3) {
		t.Fatalf("bucket should be active at limit=3 after 3 failures")
	}
	if loginBucketActive(loginThrottleKey("10.0.0.1", "victim"), 4) {
		t.Fatalf("bucket should NOT be active at limit=4 after 3 failures")
	}
}

func TestClearLoginFailuresResetsCounter(t *testing.T) {
	resetLoginThrottles()

	for i := 0; i < loginFailureLimit; i++ {
		recordLoginFailure("10.0.0.1", "victim")
	}

	if !loginThrottleActive("10.0.0.1", "victim") {
		t.Fatalf("throttle should be active before clear")
	}

	clearLoginFailures("10.0.0.1", "victim")

	if loginThrottleActive("10.0.0.1", "victim") {
		t.Fatalf("throttle still active after clear")
	}
}

func TestLoginThrottleExpiresAfterWindow(t *testing.T) {
	resetLoginThrottles()

	for i := 0; i < loginFailureLimit; i++ {
		recordLoginFailure("10.0.0.1", "victim")
	}

	if !loginThrottleActive("10.0.0.1", "victim") {
		t.Fatalf("throttle should be active")
	}

	loginThrottles.mu.Lock()
	loginThrottles.failures[loginThrottleKey("10.0.0.1", "victim")] = loginFailureEntry{
		count:   loginFailureLimit,
		resetAt: time.Now().Add(-time.Minute),
	}
	loginThrottles.mu.Unlock()

	if loginThrottleActive("10.0.0.1", "victim") {
		t.Fatalf("throttle should be inactive after window expires")
	}
}

func TestWriteLoginThrottle(t *testing.T) {
	rr := httptest.NewRecorder()
	writeLoginThrottle(rr)

	if rr.Code != http.StatusTooManyRequests {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusTooManyRequests)
	}
	if rr.Header().Get("Retry-After") != "60" {
		t.Fatalf("Retry-After = %q, want %q", rr.Header().Get("Retry-After"), "60")
	}
	if ct := rr.Header().Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
		t.Fatalf("Content-Type = %q, want application/json", ct)
	}
	if !strings.Contains(rr.Body.String(), "too many login attempts") {
		t.Fatalf("body = %q, want login throttle message", rr.Body.String())
	}
}

// ---------------------------------------------------------------------------
// 9. renderHTML
// ---------------------------------------------------------------------------

func TestRenderHTML(t *testing.T) {
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	renderHTML(rr, req, nopComponent{})

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}
	if ct := rr.Header().Get("Content-Type"); ct != "text/html; charset=utf-8" {
		t.Fatalf("Content-Type = %q, want text/html; charset=utf-8", ct)
	}
	if rr.Body.String() != "OK" {
		t.Fatalf("body = %q, want %q", rr.Body.String(), "OK")
	}
}

func TestRenderHTMLError(t *testing.T) {
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	renderHTML(rr, req, errComponent{})

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusInternalServerError)
	}
	if !strings.Contains(rr.Body.String(), "internal server error") {
		t.Fatalf("body = %q, want generic error", rr.Body.String())
	}
}

// ---------------------------------------------------------------------------
// 10. Additional edge cases
// ---------------------------------------------------------------------------

func TestNormalizedLoginUsername(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{input: "Admin", want: "admin"},
		{input: "  Alice  ", want: "alice"},
		{input: "ROOT", want: "root"},
		{input: "", want: ""},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := normalizedLoginUsername(tt.input); got != tt.want {
				t.Fatalf("normalizedLoginUsername(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestLoginThrottleKeyIncludesNormalizedUsername(t *testing.T) {
	key := loginThrottleKey("10.0.0.1", "  Alice  ")
	if !strings.Contains(key, "alice") {
		t.Fatalf("key = %q, expected normalized username 'alice'", key)
	}
	if !strings.Contains(key, "10.0.0.1") {
		t.Fatalf("key = %q, expected IP", key)
	}
}

func TestLoginIPThrottleKey(t *testing.T) {
	key := loginIPThrottleKey("10.0.0.1")
	if !strings.HasPrefix(key, "ip\x00") {
		t.Fatalf("key = %q, want ip\\x00 prefix", key)
	}
	if !strings.Contains(key, "10.0.0.1") {
		t.Fatalf("key = %q, expected IP", key)
	}
}

func TestRateLimitLogHandlerErrorNil(t *testing.T) {
	// logHandlerError should not panic when called with nil error
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	logHandlerError(req, nil)
	// If we reach here without panic, the test passes
}

func TestWarnAfterMutationNoPanic(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	warnAfterMutation(req, "delete", nil)
	warnAfterMutation(req, "delete", fmt.Errorf("mutation failed"))
	// Should not panic in either case
}
