package main

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"sync"
	"time"
)

const (
	SessionCookieName = "session_id"
	SessionDuration   = 30 * time.Minute
)

type Session struct {
	ID        string
	Data      map[string]interface{}
	ExpiresAt time.Time
}

type SessionManager struct {
	sessions map[string]*Session
	mu       sync.RWMutex
}

func NewSessionManager() *SessionManager {
	sm := &SessionManager{
		sessions: make(map[string]*Session),
	}
	// Start cleanup goroutine
	go sm.cleanupExpiredSessions()
	return sm
}

func (sm *SessionManager) generateSessionID() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func (sm *SessionManager) CreateSession() (*Session, error) {
	id, err := sm.generateSessionID()
	if err != nil {
		return nil, err
	}

	session := &Session{
		ID:        id,
		Data:      make(map[string]interface{}),
		ExpiresAt: time.Now().Add(SessionDuration),
	}

	sm.mu.Lock()
	sm.sessions[id] = session
	sm.mu.Unlock()

	return session, nil
}

func (sm *SessionManager) GetSession(id string) (*Session, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	session, exists := sm.sessions[id]
	if !exists {
		return nil, false
	}

	// Check if session is expired
	if time.Now().After(session.ExpiresAt) {
		return nil, false
	}

	return session, true
}

func (sm *SessionManager) DestroySession(id string) {
	sm.mu.Lock()
	delete(sm.sessions, id)
	sm.mu.Unlock()
}

func (sm *SessionManager) cleanupExpiredSessions() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		sm.mu.Lock()
		now := time.Now()
		for id, session := range sm.sessions {
			if now.After(session.ExpiresAt) {
				delete(sm.sessions, id)
			}
		}
		sm.mu.Unlock()
	}
}

type SessionHandler struct {
	mux     *http.ServeMux
	sm      *SessionManager
}

func NewSessionHandler() *SessionHandler {
	h := &SessionHandler{
		mux: http.NewServeMux(),
		sm:  NewSessionManager(),
	}
	h.routes()
	return h
}

func (h *SessionHandler) routes() {
	h.mux.HandleFunc("/login", h.handleLogin())
	h.mux.HandleFunc("/profile", h.handleProfile())
	h.mux.HandleFunc("/logout", h.handleLogout())
	h.mux.HandleFunc("/set-cookie", h.handleSetCookie())
	h.mux.HandleFunc("/get-cookie", h.handleGetCookie())
}

func (h *SessionHandler) handleLogin() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var creds struct {
			Username string `json:"username"`
		}

		if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		// Create session
		session, err := h.sm.CreateSession()
		if err != nil {
			http.Error(w, "Failed to create session", http.StatusInternalServerError)
			return
		}

		// Store user data in session
		session.Data["username"] = creds.Username
		session.Data["login_time"] = time.Now()

		// Set session cookie
		http.SetCookie(w, &http.Cookie{
			Name:     SessionCookieName,
			Value:    session.ID,
			Path:     "/",
			HttpOnly: true,
			MaxAge:   int(SessionDuration.Seconds()),
		})

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message":    "Login successful",
			"session_id": session.ID,
		})
	}
}

func (h *SessionHandler) handleProfile() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get session cookie
		cookie, err := r.Cookie(SessionCookieName)
		if err != nil {
			http.Error(w, "Not authenticated", http.StatusUnauthorized)
			return
		}

		// Get session
		session, exists := h.sm.GetSession(cookie.Value)
		if !exists {
			http.Error(w, "Session expired or invalid", http.StatusUnauthorized)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"username":   session.Data["username"],
			"login_time": session.Data["login_time"],
			"expires_at": session.ExpiresAt,
		})
	}
}

func (h *SessionHandler) handleLogout() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get session cookie
		cookie, err := r.Cookie(SessionCookieName)
		if err == nil {
			h.sm.DestroySession(cookie.Value)
		}

		// Delete cookie
		http.SetCookie(w, &http.Cookie{
			Name:     SessionCookieName,
			Value:    "",
			Path:     "/",
			MaxAge:   -1,
			HttpOnly: true,
		})

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Logout successful",
		})
	}
}

func (h *SessionHandler) handleSetCookie() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Set a custom cookie
		http.SetCookie(w, &http.Cookie{
			Name:  "preferences",
			Value: "dark_mode=true",
			Path:  "/",
			MaxAge: 3600,
		})

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Cookie set successfully",
		})
	}
}

func (h *SessionHandler) handleGetCookie() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("preferences")
		if err != nil {
			http.Error(w, "Cookie not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"cookie_value": cookie.Value,
		})
	}
}

func (h *SessionHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mux.ServeHTTP(w, r)
}

func main() {}
