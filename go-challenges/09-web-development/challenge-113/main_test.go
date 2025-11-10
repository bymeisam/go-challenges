package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestJWTAuth(t *testing.T) {
	auth := NewJWTAuth()

	t.Run("Login", func(t *testing.T) {
		creds := map[string]string{
			"username": "alice",
			"password": "password123",
		}
		jsonData, _ := json.Marshal(creds)

		req := httptest.NewRequest("POST", "/login", bytes.NewReader(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		auth.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var resp map[string]string
		json.NewDecoder(w.Body).Decode(&resp)

		if resp["token"] == "" {
			t.Error("Expected token to be returned")
		}

		// Verify token format (should have 3 parts)
		parts := strings.Split(resp["token"], ".")
		if len(parts) != 3 {
			t.Errorf("Expected token to have 3 parts, got %d", len(parts))
		}
	})

	t.Run("LoginInvalidCredentials", func(t *testing.T) {
		creds := map[string]string{
			"username": "",
			"password": "",
		}
		jsonData, _ := json.Marshal(creds)

		req := httptest.NewRequest("POST", "/login", bytes.NewReader(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		auth.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected status 401, got %d", w.Code)
		}
	})

	t.Run("AccessProtectedWithValidToken", func(t *testing.T) {
		// First login to get token
		creds := map[string]string{
			"username": "bob",
			"password": "password123",
		}
		jsonData, _ := json.Marshal(creds)

		loginReq := httptest.NewRequest("POST", "/login", bytes.NewReader(jsonData))
		loginReq.Header.Set("Content-Type", "application/json")
		loginW := httptest.NewRecorder()
		auth.ServeHTTP(loginW, loginReq)

		var loginResp map[string]string
		json.NewDecoder(loginW.Body).Decode(&loginResp)
		token := loginResp["token"]

		// Access protected route
		protectedReq := httptest.NewRequest("GET", "/protected", nil)
		protectedReq.Header.Set("Authorization", "Bearer "+token)
		protectedW := httptest.NewRecorder()
		auth.ServeHTTP(protectedW, protectedReq)

		if protectedW.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", protectedW.Code)
		}

		var protectedResp map[string]interface{}
		json.NewDecoder(protectedW.Body).Decode(&protectedResp)

		if protectedResp["username"] != "bob" {
			t.Errorf("Expected username 'bob', got %v", protectedResp["username"])
		}
		if protectedResp["message"] != "This is protected data" {
			t.Error("Expected protected data message")
		}
	})

	t.Run("AccessProtectedWithoutToken", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/protected", nil)
		w := httptest.NewRecorder()

		auth.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected status 401, got %d", w.Code)
		}
	})

	t.Run("AccessProtectedWithInvalidToken", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "Bearer invalid.token.here")
		w := httptest.NewRecorder()

		auth.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected status 401, got %d", w.Code)
		}
	})

	t.Run("AccessProtectedWithMalformedHeader", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "InvalidFormat")
		w := httptest.NewRecorder()

		auth.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected status 401, got %d", w.Code)
		}
	})

	t.Run("AccessPublicRoute", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/public", nil)
		w := httptest.NewRecorder()

		auth.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var resp map[string]string
		json.NewDecoder(w.Body).Decode(&resp)

		if resp["message"] != "This is public data" {
			t.Error("Expected public data message")
		}
	})

	t.Log("✓ JWT authentication works!")
}

func TestTokenValidation(t *testing.T) {
	t.Run("ValidToken", func(t *testing.T) {
		token, err := createToken(1, "testuser")
		if err != nil {
			t.Fatalf("Failed to create token: %v", err)
		}

		claims, err := validateToken(token)
		if err != nil {
			t.Errorf("Expected valid token, got error: %v", err)
		}

		if claims.Username != "testuser" {
			t.Errorf("Expected username 'testuser', got '%s'", claims.Username)
		}
		if claims.UserID != 1 {
			t.Errorf("Expected user_id 1, got %d", claims.UserID)
		}
	})

	t.Run("ExpiredToken", func(t *testing.T) {
		// Create token that's already expired
		claims := Claims{
			UserID:   1,
			Username: "testuser",
			Exp:      time.Now().Add(-1 * time.Hour).Unix(), // Expired 1 hour ago
		}

		// Manually create expired token
		header := map[string]string{"alg": "HS256", "typ": "JWT"}
		headerJSON, _ := json.Marshal(header)
		headerEncoded := base64Encode(headerJSON)

		payloadJSON, _ := json.Marshal(claims)
		payloadEncoded := base64Encode(payloadJSON)

		message := headerEncoded + "." + payloadEncoded
		signature := createSignature(message)
		token := message + "." + signature

		_, err := validateToken(token)
		if err == nil {
			t.Error("Expected error for expired token")
		}
		if !strings.Contains(err.Error(), "expired") {
			t.Errorf("Expected 'expired' error, got: %v", err)
		}
	})

	t.Run("InvalidSignature", func(t *testing.T) {
		token, _ := createToken(1, "testuser")
		// Tamper with token
		parts := strings.Split(token, ".")
		parts[2] = "invalid-signature"
		tamperedToken := strings.Join(parts, ".")

		_, err := validateToken(tamperedToken)
		if err == nil {
			t.Error("Expected error for invalid signature")
		}
	})

	t.Log("✓ Token validation works!")
}

func base64Encode(data []byte) string {
	return strings.TrimRight(
		strings.Replace(
			strings.Replace(
				string(data),
				"+", "-", -1),
			"/", "_", -1),
		"=")
}
