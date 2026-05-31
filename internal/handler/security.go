package handler

import (
	"net/http"
	"strings"
	"sync"
	"time"
)

// WithSecurityHeaders wraps an http.Handler with security headers.
// This is applied at the server level (to all routes).
func WithSecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Prevent MIME sniffing
		w.Header().Set("X-Content-Type-Options", "nosniff")
		// Prevent clickjacking
		w.Header().Set("X-Frame-Options", "DENY")
		// Enable XSS filter in older browsers
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		// Referrer policy
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		// Content Security Policy (allow inline styles for Tailwind)
		w.Header().Set("Content-Security-Policy", "default-src 'self'; style-src 'self' 'unsafe-inline' https://fonts.googleapis.com; font-src 'self' https://fonts.gstatic.com; script-src 'self' 'unsafe-inline' https://unpkg.com; img-src 'self' data:; connect-src 'self'")
		// HSTS (only in production — enable when HTTPS is confirmed working)
		// w.Header().Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains; preload")

		next.ServeHTTP(w, r)
	})
}

// --- Rate limiter (simple token bucket per IP) ---

type rateLimiter struct {
	mu       sync.Mutex
	requests map[string]*bucketEntry
}

type bucketEntry struct {
	count    int
	resetAt  time.Time
}

var limiter = &rateLimiter{
	requests: make(map[string]*bucketEntry),
}

const maxRequests = 60          // max requests per window
const windowDuration = 1 * time.Minute

// RateLimit is a middleware that limits requests per IP.
// Returns 429 Too Many Requests if exceeded.
func RateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip for static files
		if strings.HasPrefix(r.URL.Path, "/static/") {
			next.ServeHTTP(w, r)
			return
		}

		ip := r.RemoteAddr

		limiter.mu.Lock()
		entry, exists := limiter.requests[ip]

		now := time.Now()
		if !exists || now.After(entry.resetAt) {
			limiter.requests[ip] = &bucketEntry{
				count:   1,
				resetAt: now.Add(windowDuration),
			}
			limiter.mu.Unlock()
			next.ServeHTTP(w, r)
			return
		}

		entry.count++
		if entry.count > maxRequests {
			limiter.mu.Unlock()
			w.Header().Set("Retry-After", "60")
			http.Error(w, "429 — too many requests", http.StatusTooManyRequests)
			return
		}

		limiter.mu.Unlock()
		next.ServeHTTP(w, r)
	})
}

// Cleanup rate limiter entries periodically (call from background goroutine)
func init() {
	go func() {
		for {
			time.Sleep(5 * time.Minute)
			limiter.mu.Lock()
			for ip, entry := range limiter.requests {
				if time.Since(entry.resetAt) > windowDuration {
					delete(limiter.requests, ip)
				}
			}
			limiter.mu.Unlock()
		}
	}()
}
