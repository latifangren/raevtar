package handler

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"net/http"
	"sync"
	"time"

	"raevtar/internal/model"
	adminview "raevtar/internal/view/admin"
)

// --- In-memory session store (with role) ---

type sessionEntry struct {
	userID    int64
	username  string
	role      string
	csrfToken string
	createdAt time.Time
}

type sessionStore struct {
	mu       sync.RWMutex
	sessions map[string]sessionEntry
}

var sessions = &sessionStore{
	sessions: make(map[string]sessionEntry),
}

const sessionCookieName = "raevtar_session"
const sessionMaxAge = 24 * time.Hour

func generateSessionToken() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		panic("generate session token: " + err.Error())
	}
	return hex.EncodeToString(b)
}

func (s *sessionStore) create(userID int64, username, role string) string {
	s.mu.Lock()
	defer s.mu.Unlock()
	token := generateSessionToken()
	s.sessions[token] = sessionEntry{
		userID:    userID,
		username:  username,
		role:      role,
		csrfToken: generateSessionToken(),
		createdAt: time.Now(),
	}
	return token
}

func (s *sessionStore) get(token string) (sessionEntry, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	entry, ok := s.sessions[token]
	if !ok {
		return sessionEntry{}, false
	}
	if time.Since(entry.createdAt) > sessionMaxAge {
		return sessionEntry{}, false
	}
	return entry, true
}

func (s *sessionStore) delete(token string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.sessions, token)
}

func (s *sessionStore) cleanup() {
	s.mu.Lock()
	defer s.mu.Unlock()
	for token, entry := range s.sessions {
		if time.Since(entry.createdAt) > sessionMaxAge {
			delete(s.sessions, token)
		}
	}
}

// --- Session helper ---

func getSessionEntry(r *http.Request) (sessionEntry, bool) {
	cookie, err := r.Cookie(sessionCookieName)
	if err != nil {
		return sessionEntry{}, false
	}
	return sessions.get(cookie.Value)
}

func canManageServer(r *http.Request) bool {
	entry, ok := getSessionEntry(r)
	if !ok {
		return false
	}
	return entry.role == model.RoleOwner || entry.role == model.RoleAdmin
}

func csrfTokenForRequest(r *http.Request) string {
	entry, ok := getSessionEntry(r)
	if !ok {
		return ""
	}
	return entry.csrfToken
}

// --- RBAC Middleware ---

// adminRequired: any authenticated user (role = owner/admin/operator/readonly)
func (h *Handler) adminRequired(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := getSessionEntry(r); !ok {
			http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// ownerRequired: only owner
func (h *Handler) ownerRequired(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		entry, ok := getSessionEntry(r)
		if !ok {
			http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
			return
		}
		if entry.role != model.RoleOwner {
			http.Error(w, "403 — only owner can access this", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// ownerOrAdmin: owner or admin
func (h *Handler) ownerOrAdminRequired(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		entry, ok := getSessionEntry(r)
		if !ok {
			http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
			return
		}
		if entry.role != model.RoleOwner && entry.role != model.RoleAdmin {
			http.Error(w, "403 — admin access required", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// adminAPIAuth validates session for API calls (returns JSON 401 instead of redirect).
func (h *Handler) adminAPIAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := getSessionEntry(r); !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "not authenticated"})
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (h *Handler) adminCSRF(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			next.ServeHTTP(w, r)
			return
		}

		entry, ok := getSessionEntry(r)
		if !ok || entry.csrfToken == "" {
			http.Error(w, "403 — invalid CSRF token", http.StatusForbidden)
			return
		}

		sent := r.FormValue("_csrf")
		if subtle.ConstantTimeCompare([]byte(sent), []byte(entry.csrfToken)) != 1 {
			http.Error(w, "403 — invalid CSRF token", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// --- Login/Logout handlers ---

func (h *Handler) adminLoginPage(w http.ResponseWriter, r *http.Request) {
	// Already logged in? Redirect to admin panel
	if _, ok := getSessionEntry(r); ok {
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
		return
	}

	renderHTML(w, r, adminview.Login())
}

func (h *Handler) adminLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "POST required"})
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")

	if username == "" || password == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "username and password required"})
		return
	}

	user, err := h.svc.Admin.Authenticate(username, password, clientIP(r))
	if err != nil {
		sessions.cleanup()
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid credentials"})
		return
	}

	token := sessions.create(user.ID, user.Username, user.Role)

	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   h.cfg.IsProduction,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   int(sessionMaxAge.Seconds()),
	})

	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

func (h *Handler) adminLogout(w http.ResponseWriter, r *http.Request) {
	entry, _ := getSessionEntry(r)
	cookie, err := r.Cookie(sessionCookieName)
	if err == nil {
		sessions.delete(cookie.Value)
	}

	if entry.username != "" {
		_ = h.svc.Admin.LogLogout(entry.username, clientIP(r))
	}

	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   h.cfg.IsProduction,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   -1,
	})

	http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
}
