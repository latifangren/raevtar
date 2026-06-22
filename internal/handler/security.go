package handler

import (
	"net"
	"net/http"
	"raevtar/internal/config"
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
		w.Header().Set("Content-Security-Policy", "default-src 'self'; style-src 'self' 'unsafe-inline' https://fonts.googleapis.com; font-src 'self' https://fonts.gstatic.com; script-src 'self'; img-src 'self' data:; connect-src 'self'")
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
	count   int
	resetAt time.Time
}

var limiter = &rateLimiter{
	requests: make(map[string]*bucketEntry),
}

var trustedProxyNets []*net.IPNet

var maxRequests = 60                 // max requests per window — set via initRateLimiter
var windowDuration = 1 * time.Minute // window duration — set via initRateLimiter

// initRateLimiter configures the rate limiter from the application config.
// Called once at startup from handler.New().
func initRateLimiter(cfg *config.Config) {
	if cfg.RateLimitRequests > 0 {
		maxRequests = cfg.RateLimitRequests
	}
	if cfg.RateLimitWindow > 0 {
		windowDuration = cfg.RateLimitWindow
	}
}

// RateLimit is a middleware that limits requests per IP.
// Returns 429 Too Many Requests if exceeded.
func RateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip for static files
		if strings.HasPrefix(r.URL.Path, "/static/") {
			next.ServeHTTP(w, r)
			return
		}

		ip := clientIP(r)

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

func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		if headerIP := trustedHeaderIP(host, r); headerIP != "" {
			return headerIP
		}
		return host
	}
	if headerIP := trustedHeaderIP(r.RemoteAddr, r); headerIP != "" {
		return headerIP
	}
	return r.RemoteAddr
}

func configureTrustedProxies(cfg *config.Config) {
	trustedProxyNets = trustedProxyNets[:0]
	if cfg == nil {
		return
	}
	for _, value := range cfg.TrustedProxyCIDRs {
		if strings.TrimSpace(value) == "" {
			continue
		}
		if strings.Contains(value, "/") {
			if _, network, err := net.ParseCIDR(value); err == nil {
				trustedProxyNets = append(trustedProxyNets, network)
			}
			continue
		}
		ip := net.ParseIP(value)
		if ip == nil {
			continue
		}
		bits := 32
		if ip.To4() == nil {
			bits = 128
		}
		trustedProxyNets = append(trustedProxyNets, &net.IPNet{IP: ip, Mask: net.CIDRMask(bits, bits)})
	}
}

func trustedHeaderIP(remoteHost string, r *http.Request) string {
	if !isTrustedProxy(remoteHost) {
		return ""
	}
	for _, value := range []string{r.Header.Get("CF-Connecting-IP"), firstForwardedFor(r.Header.Get("X-Forwarded-For")), r.Header.Get("X-Real-IP")} {
		ip := net.ParseIP(strings.TrimSpace(value))
		if ip != nil {
			return ip.String()
		}
	}
	return ""
}

func firstForwardedFor(value string) string {
	parts := strings.Split(value, ",")
	if len(parts) == 0 {
		return ""
	}
	return strings.TrimSpace(parts[0])
}

func isTrustedProxy(remoteHost string) bool {
	ip := net.ParseIP(remoteHost)
	if ip == nil {
		return false
	}
	for _, network := range trustedProxyNets {
		if network.Contains(ip) {
			return true
		}
	}
	return false
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
