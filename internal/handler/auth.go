package handler

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"raevtar/internal/model"
	"raevtar/internal/repo"
)

// --- In-memory session store (with role) ---

type sessionEntry struct {
	userID    int64
	username  string
	role      string
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
	rand.Read(b)
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

// --- Login/Logout handlers ---

func (h *Handler) adminLoginPage(w http.ResponseWriter, r *http.Request) {
	// Already logged in? Redirect to admin panel
	if _, ok := getSessionEntry(r); ok {
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
		return
	}

	html := `<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<title>Login — Raevtar Admin</title>
	<link rel="preconnect" href="https://fonts.googleapis.com">
	<link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
	<link href="https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700;900&family=JetBrains+Mono:wght@400;500;700&display=swap" rel="stylesheet">
	<link rel="stylesheet" href="/static/css/style.css">
	<link rel="icon" type="image/svg+xml" href="/static/favicon.svg">
</head>
<body class="bg-neutral-100 text-black font-sans min-h-screen flex items-center justify-center">
	<div class="w-full max-w-sm mx-4">
		<div class="bg-white border-2 border-black p-8 shadow-[8px_8px_0px_0px_#000]">
			<div class="text-center mb-8">
				<div class="w-12 h-12 bg-emerald-400 border-2 border-black flex items-center justify-center mx-auto mb-3">
					<span class="text-xl font-black text-black">R</span>
				</div>
				<h1 class="text-2xl font-black text-black uppercase">Raevtar</h1>
				<p class="text-sm font-bold text-neutral-600 mt-1">admin panel</p>
			</div>

			<form method="POST" action="/admin/login" class="space-y-4">
				<div>
					<label class="block text-sm font-bold text-neutral-700 mb-1">Username</label>
					<input type="text" name="username" required autocomplete="username"
						class="w-full px-3 py-2 border-2 border-black bg-white text-sm font-bold text-black placeholder:text-neutral-400 focus:outline-none focus:ring-2 focus:ring-black"
						placeholder="admin">
				</div>
				<div>
					<label class="block text-sm font-bold text-neutral-700 mb-1">Password</label>
					<input type="password" name="password" required autocomplete="current-password"
						class="w-full px-3 py-2 border-2 border-black bg-white text-sm font-bold text-black placeholder:text-neutral-400 focus:outline-none focus:ring-2 focus:ring-black"
						placeholder="••••••••">
				</div>
				<button type="submit"
					class="w-full px-4 py-3 border-2 border-black bg-black text-white text-sm font-bold shadow-[4px_4px_0px_0px_#000] hover:shadow-[2px_2px_0px_0px_#000] hover:translate-x-[2px] hover:translate-y-[2px] transition-all cursor-pointer">
					Sign in
				</button>
			</form>

			<p class="text-xs font-bold text-neutral-500 text-center mt-6">
				Unauthorized access is prohibited.
			</p>
		</div>
	</div>
</body>
</html>`
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(html))
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

	// Look up user from DB
	user, err := h.svc.Repos.User.GetByUsername(username)
	if err != nil {
		sessions.cleanup()
		// Log failed attempt
		h.svc.Repos.Audit.Insert(username, "LOGIN_FAILED", "user not found", r.RemoteAddr)
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid credentials"})
		return
	}

	if !repo.CheckPassword(password, user.PasswordHash) {
		sessions.cleanup()
		h.svc.Repos.Audit.Insert(username, "LOGIN_FAILED", "wrong password", r.RemoteAddr)
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid credentials"})
		return
	}

	token := sessions.create(user.ID, user.Username, user.Role)

	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(sessionMaxAge.Seconds()),
	})

	h.svc.Repos.Audit.Insert(user.Username, "LOGIN", "login via admin panel", r.RemoteAddr)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok", "redirect": "/admin"})
}

func (h *Handler) adminLogout(w http.ResponseWriter, r *http.Request) {
	entry, _ := getSessionEntry(r)
	cookie, err := r.Cookie(sessionCookieName)
	if err == nil {
		sessions.delete(cookie.Value)
	}

	if entry.username != "" {
		h.svc.Repos.Audit.Insert(entry.username, "LOGOUT", "manual logout", r.RemoteAddr)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})

	http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
}
