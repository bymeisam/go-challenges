package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCORS(t *testing.T) {
	config := &CORSConfig{
		AllowedOrigins:   []string{"http://example.com", "http://localhost:3000"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		ExposedHeaders:   []string{"X-Custom-Header"},
		AllowCredentials: true,
		MaxAge:           3600,
	}

	server := NewCORSServer(config)

	t.Run("PreflightRequest", func(t *testing.T) {
		req := httptest.NewRequest("OPTIONS", "/api/data", nil)
		req.Header.Set("Origin", "http://example.com")
		req.Header.Set("Access-Control-Request-Method", "POST")
		req.Header.Set("Access-Control-Request-Headers", "Content-Type")
		w := httptest.NewRecorder()

		server.ServeHTTP(w, req)

		if w.Code != http.StatusNoContent {
			t.Errorf("Expected status 204, got %d", w.Code)
		}

		// Check CORS headers
		allowOrigin := w.Header().Get("Access-Control-Allow-Origin")
		if allowOrigin != "http://example.com" {
			t.Errorf("Expected Allow-Origin 'http://example.com', got '%s'", allowOrigin)
		}

		allowMethods := w.Header().Get("Access-Control-Allow-Methods")
		if !strings.Contains(allowMethods, "POST") {
			t.Errorf("Expected Allow-Methods to contain 'POST', got '%s'", allowMethods)
		}

		allowHeaders := w.Header().Get("Access-Control-Allow-Headers")
		if !strings.Contains(allowHeaders, "Content-Type") {
			t.Errorf("Expected Allow-Headers to contain 'Content-Type', got '%s'", allowHeaders)
		}

		allowCredentials := w.Header().Get("Access-Control-Allow-Credentials")
		if allowCredentials != "true" {
			t.Errorf("Expected Allow-Credentials 'true', got '%s'", allowCredentials)
		}
	})

	t.Run("PreflightRequestInvalidMethod", func(t *testing.T) {
		req := httptest.NewRequest("OPTIONS", "/api/data", nil)
		req.Header.Set("Origin", "http://example.com")
		req.Header.Set("Access-Control-Request-Method", "PATCH")
		w := httptest.NewRecorder()

		server.ServeHTTP(w, req)

		if w.Code != http.StatusForbidden {
			t.Errorf("Expected status 403 for invalid method, got %d", w.Code)
		}
	})

	t.Run("PreflightRequestInvalidHeader", func(t *testing.T) {
		req := httptest.NewRequest("OPTIONS", "/api/data", nil)
		req.Header.Set("Origin", "http://example.com")
		req.Header.Set("Access-Control-Request-Method", "POST")
		req.Header.Set("Access-Control-Request-Headers", "X-Invalid-Header")
		w := httptest.NewRecorder()

		server.ServeHTTP(w, req)

		if w.Code != http.StatusForbidden {
			t.Errorf("Expected status 403 for invalid header, got %d", w.Code)
		}
	})

	t.Run("ActualRequestWithAllowedOrigin", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/data", nil)
		req.Header.Set("Origin", "http://localhost:3000")
		w := httptest.NewRecorder()

		server.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		allowOrigin := w.Header().Get("Access-Control-Allow-Origin")
		if allowOrigin != "http://localhost:3000" {
			t.Errorf("Expected Allow-Origin 'http://localhost:3000', got '%s'", allowOrigin)
		}

		exposeHeaders := w.Header().Get("Access-Control-Expose-Headers")
		if !strings.Contains(exposeHeaders, "X-Custom-Header") {
			t.Errorf("Expected Expose-Headers to contain 'X-Custom-Header', got '%s'", exposeHeaders)
		}
	})

	t.Run("ActualRequestWithDisallowedOrigin", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/data", nil)
		req.Header.Set("Origin", "http://evil.com")
		w := httptest.NewRecorder()

		server.ServeHTTP(w, req)

		// Request should still process but without CORS headers
		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		allowOrigin := w.Header().Get("Access-Control-Allow-Origin")
		if allowOrigin == "http://evil.com" {
			t.Error("Should not set Allow-Origin for disallowed origin")
		}
	})

	t.Run("WildcardOrigin", func(t *testing.T) {
		wildcardConfig := &CORSConfig{
			AllowedOrigins: []string{"*"},
			AllowedMethods: []string{"GET", "POST"},
			AllowedHeaders: []string{"Content-Type"},
		}

		wildcardServer := NewCORSServer(wildcardConfig)

		req := httptest.NewRequest("GET", "/api/data", nil)
		req.Header.Set("Origin", "http://any-domain.com")
		w := httptest.NewRecorder()

		wildcardServer.ServeHTTP(w, req)

		allowOrigin := w.Header().Get("Access-Control-Allow-Origin")
		if allowOrigin != "*" {
			t.Errorf("Expected Allow-Origin '*', got '%s'", allowOrigin)
		}
	})

	t.Run("DifferentHTTPMethods", func(t *testing.T) {
		methods := []string{"GET", "POST", "PUT", "DELETE"}

		for _, method := range methods {
			req := httptest.NewRequest(method, "/api/data", nil)
			req.Header.Set("Origin", "http://example.com")
			w := httptest.NewRecorder()

			server.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("Expected status 200 for %s, got %d", method, w.Code)
			}

			allowOrigin := w.Header().Get("Access-Control-Allow-Origin")
			if allowOrigin != "http://example.com" {
				t.Errorf("Expected Allow-Origin for %s method", method)
			}
		}
	})

	t.Run("DefaultConfig", func(t *testing.T) {
		defaultServer := NewCORSServer(nil)

		req := httptest.NewRequest("GET", "/api/data", nil)
		req.Header.Set("Origin", "http://any-domain.com")
		w := httptest.NewRecorder()

		defaultServer.ServeHTTP(w, req)

		allowOrigin := w.Header().Get("Access-Control-Allow-Origin")
		if allowOrigin != "*" {
			t.Errorf("Default config should allow all origins, got '%s'", allowOrigin)
		}
	})

	t.Run("CredentialsWithWildcard", func(t *testing.T) {
		// Note: In real CORS, you can't use credentials with wildcard origin
		// But our implementation allows it for testing purposes
		credConfig := &CORSConfig{
			AllowedOrigins:   []string{"*"},
			AllowedMethods:   []string{"GET"},
			AllowedHeaders:   []string{"Content-Type"},
			AllowCredentials: true,
		}

		credServer := NewCORSServer(credConfig)

		req := httptest.NewRequest("GET", "/api/data", nil)
		req.Header.Set("Origin", "http://example.com")
		w := httptest.NewRecorder()

		credServer.ServeHTTP(w, req)

		allowCredentials := w.Header().Get("Access-Control-Allow-Credentials")
		if allowCredentials != "true" {
			t.Errorf("Expected Allow-Credentials 'true', got '%s'", allowCredentials)
		}
	})

	t.Run("MultipleRequestHeaders", func(t *testing.T) {
		req := httptest.NewRequest("OPTIONS", "/api/data", nil)
		req.Header.Set("Origin", "http://example.com")
		req.Header.Set("Access-Control-Request-Method", "POST")
		req.Header.Set("Access-Control-Request-Headers", "Content-Type, Authorization")
		w := httptest.NewRecorder()

		server.ServeHTTP(w, req)

		if w.Code != http.StatusNoContent {
			t.Errorf("Expected status 204, got %d", w.Code)
		}
	})

	t.Log("✓ CORS handling works!")
}
