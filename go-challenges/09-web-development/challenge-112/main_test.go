package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestSessionManagement(t *testing.T) {
	handler := NewSessionHandler()

	t.Run("Login", func(t *testing.T) {
		loginData := map[string]string{
			"username": "alice",
		}
		jsonData, _ := json.Marshal(loginData)

		req := httptest.NewRequest("POST", "/login", bytes.NewReader(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var resp map[string]interface{}
		json.NewDecoder(w.Body).Decode(&resp)

		if resp["message"] != "Login successful" {
			t.Errorf("Unexpected message: %v", resp["message"])
		}

		// Check if session cookie is set
		cookies := w.Result().Cookies()
		var sessionCookie *http.Cookie
		for _, cookie := range cookies {
			if cookie.Name == SessionCookieName {
				sessionCookie = cookie
				break
			}
		}

		if sessionCookie == nil {
			t.Fatal("Session cookie not set")
		}

		if sessionCookie.Value == "" {
			t.Error("Session cookie value is empty")
		}

		if !sessionCookie.HttpOnly {
			t.Error("Session cookie should be HttpOnly")
		}
	})

	t.Run("ProfileWithSession", func(t *testing.T) {
		// First login to get session
		loginData := map[string]string{
			"username": "bob",
		}
		jsonData, _ := json.Marshal(loginData)

		loginReq := httptest.NewRequest("POST", "/login", bytes.NewReader(jsonData))
		loginReq.Header.Set("Content-Type", "application/json")
		loginW := httptest.NewRecorder()
		handler.ServeHTTP(loginW, loginReq)

		// Get session cookie
		var sessionCookie *http.Cookie
		for _, cookie := range loginW.Result().Cookies() {
			if cookie.Name == SessionCookieName {
				sessionCookie = cookie
				break
			}
		}

		// Access profile with session
		profileReq := httptest.NewRequest("GET", "/profile", nil)
		profileReq.AddCookie(sessionCookie)
		profileW := httptest.NewRecorder()
		handler.ServeHTTP(profileW, profileReq)

		if profileW.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", profileW.Code)
		}

		var profile map[string]interface{}
		json.NewDecoder(profileW.Body).Decode(&profile)

		if profile["username"] != "bob" {
			t.Errorf("Expected username 'bob', got %v", profile["username"])
		}

		if profile["login_time"] == nil {
			t.Error("Expected login_time to be set")
		}
	})

	t.Run("ProfileWithoutSession", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/profile", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected status 401, got %d", w.Code)
		}
	})

	t.Run("Logout", func(t *testing.T) {
		// First login
		loginData := map[string]string{
			"username": "charlie",
		}
		jsonData, _ := json.Marshal(loginData)

		loginReq := httptest.NewRequest("POST", "/login", bytes.NewReader(jsonData))
		loginReq.Header.Set("Content-Type", "application/json")
		loginW := httptest.NewRecorder()
		handler.ServeHTTP(loginW, loginReq)

		var sessionCookie *http.Cookie
		for _, cookie := range loginW.Result().Cookies() {
			if cookie.Name == SessionCookieName {
				sessionCookie = cookie
				break
			}
		}

		// Logout
		logoutReq := httptest.NewRequest("GET", "/logout", nil)
		logoutReq.AddCookie(sessionCookie)
		logoutW := httptest.NewRecorder()
		handler.ServeHTTP(logoutW, logoutReq)

		if logoutW.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", logoutW.Code)
		}

		// Try to access profile after logout
		profileReq := httptest.NewRequest("GET", "/profile", nil)
		profileReq.AddCookie(sessionCookie)
		profileW := httptest.NewRecorder()
		handler.ServeHTTP(profileW, profileReq)

		if profileW.Code != http.StatusUnauthorized {
			t.Errorf("Expected status 401 after logout, got %d", profileW.Code)
		}
	})

	t.Run("SetAndGetCookie", func(t *testing.T) {
		// Set cookie
		setReq := httptest.NewRequest("GET", "/set-cookie", nil)
		setW := httptest.NewRecorder()
		handler.ServeHTTP(setW, setReq)

		if setW.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", setW.Code)
		}

		// Get the cookie
		var prefCookie *http.Cookie
		for _, cookie := range setW.Result().Cookies() {
			if cookie.Name == "preferences" {
				prefCookie = cookie
				break
			}
		}

		if prefCookie == nil {
			t.Fatal("Preferences cookie not set")
		}

		// Get cookie value
		getReq := httptest.NewRequest("GET", "/get-cookie", nil)
		getReq.AddCookie(prefCookie)
		getW := httptest.NewRecorder()
		handler.ServeHTTP(getW, getReq)

		if getW.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", getW.Code)
		}

		var resp map[string]string
		json.NewDecoder(getW.Body).Decode(&resp)

		if resp["cookie_value"] != "dark_mode=true" {
			t.Errorf("Expected cookie_value 'dark_mode=true', got '%s'", resp["cookie_value"])
		}
	})

	t.Run("GetCookieNotFound", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/get-cookie", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status 404, got %d", w.Code)
		}
	})

	t.Log("✓ Session and cookie management works!")
}

func TestSessionExpiration(t *testing.T) {
	sm := NewSessionManager()

	// Create a session with short duration
	session, _ := sm.CreateSession()
	session.ExpiresAt = time.Now().Add(100 * time.Millisecond)

	// Session should exist initially
	if _, exists := sm.GetSession(session.ID); !exists {
		t.Error("Session should exist initially")
	}

	// Wait for expiration
	time.Sleep(200 * time.Millisecond)

	// Session should be expired
	if _, exists := sm.GetSession(session.ID); exists {
		t.Error("Session should be expired")
	}

	t.Log("✓ Session expiration works!")
}
