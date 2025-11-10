package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRateLimiting(t *testing.T) {
	server := NewRateLimitedServer(2) // 2 requests per second

	t.Run("AllowedRequests", func(t *testing.T) {
		// First request should succeed
		req := httptest.NewRequest("GET", "/api/data", nil)
		req.RemoteAddr = "192.168.1.1:12345"
		w := httptest.NewRecorder()

		server.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		// Check rate limit headers
		limit := w.Header().Get("X-RateLimit-Limit")
		if limit == "" {
			t.Error("Expected X-RateLimit-Limit header")
		}

		remaining := w.Header().Get("X-RateLimit-Remaining")
		if remaining == "" {
			t.Error("Expected X-RateLimit-Remaining header")
		}
	})

	t.Run("RateLimitExceeded", func(t *testing.T) {
		// Create a new server with very low limit for testing
		testServer := NewRateLimitedServer(1) // 1 request per second

		// Make requests until rate limited
		ip := "192.168.1.2:12345"
		var lastStatus int

		// Try to make 15 requests rapidly
		for i := 0; i < 15; i++ {
			req := httptest.NewRequest("GET", "/api/data", nil)
			req.RemoteAddr = ip
			w := httptest.NewRecorder()

			testServer.ServeHTTP(w, req)
			lastStatus = w.Code

			if w.Code == http.StatusTooManyRequests {
				// Verify rate limit response
				var resp map[string]interface{}
				json.NewDecoder(w.Body).Decode(&resp)

				if resp["error"] != "Rate limit exceeded" {
					t.Errorf("Expected rate limit error message")
				}

				remaining := w.Header().Get("X-RateLimit-Remaining")
				if remaining != "0" {
					t.Errorf("Expected remaining to be 0, got %s", remaining)
				}

				break
			}

			// Small delay between requests
			time.Sleep(10 * time.Millisecond)
		}

		if lastStatus != http.StatusTooManyRequests {
			t.Error("Expected to be rate limited after multiple requests")
		}
	})

	t.Run("DifferentIPsIndependentLimits", func(t *testing.T) {
		testServer := NewRateLimitedServer(1)

		// First IP
		req1 := httptest.NewRequest("GET", "/api/data", nil)
		req1.RemoteAddr = "192.168.1.3:12345"
		w1 := httptest.NewRecorder()
		testServer.ServeHTTP(w1, req1)

		// Second IP (different)
		req2 := httptest.NewRequest("GET", "/api/data", nil)
		req2.RemoteAddr = "192.168.1.4:12345"
		w2 := httptest.NewRecorder()
		testServer.ServeHTTP(w2, req2)

		// Both should succeed as they're from different IPs
		if w1.Code != http.StatusOK {
			t.Errorf("Expected first IP request to succeed, got %d", w1.Code)
		}
		if w2.Code != http.StatusOK {
			t.Errorf("Expected second IP request to succeed, got %d", w2.Code)
		}
	})

	t.Run("UnlimitedEndpoint", func(t *testing.T) {
		// Unlimited endpoint should not be rate limited
		for i := 0; i < 20; i++ {
			req := httptest.NewRequest("GET", "/api/unlimited", nil)
			req.RemoteAddr = "192.168.1.5:12345"
			w := httptest.NewRecorder()

			server.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("Unlimited endpoint should not be rate limited, got status %d", w.Code)
			}
		}
	})

	t.Run("TokenRefill", func(t *testing.T) {
		testServer := NewRateLimitedServer(2) // 2 requests per second

		ip := "192.168.1.6:12345"

		// Exhaust the bucket
		for i := 0; i < 25; i++ {
			req := httptest.NewRequest("GET", "/api/data", nil)
			req.RemoteAddr = ip
			w := httptest.NewRecorder()
			testServer.ServeHTTP(w, req)
		}

		// Should be rate limited now
		req := httptest.NewRequest("GET", "/api/data", nil)
		req.RemoteAddr = ip
		w := httptest.NewRecorder()
		testServer.ServeHTTP(w, req)

		if w.Code != http.StatusTooManyRequests {
			t.Errorf("Expected to be rate limited, got status %d", w.Code)
		}

		// Wait for token refill (1 second should give us 2 tokens)
		time.Sleep(1100 * time.Millisecond)

		// Should succeed now after refill
		req = httptest.NewRequest("GET", "/api/data", nil)
		req.RemoteAddr = ip
		w = httptest.NewRecorder()
		testServer.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected request to succeed after token refill, got status %d", w.Code)
		}
	})

	t.Run("XForwardedForHeader", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/data", nil)
		req.Header.Set("X-Forwarded-For", "10.0.0.1")
		req.RemoteAddr = "192.168.1.7:12345"
		w := httptest.NewRecorder()

		server.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}
	})

	t.Log("✓ Rate limiting works!")
}

func TestTokenBucket(t *testing.T) {
	t.Run("InitialCapacity", func(t *testing.T) {
		tb := NewTokenBucket(10, 1)

		if tb.Remaining() != 10 {
			t.Errorf("Expected 10 tokens initially, got %d", tb.Remaining())
		}
	})

	t.Run("ConsumeTokens", func(t *testing.T) {
		tb := NewTokenBucket(5, 1)

		// Consume 3 tokens
		for i := 0; i < 3; i++ {
			if !tb.Allow() {
				t.Error("Expected to allow request")
			}
		}

		remaining := tb.Remaining()
		if remaining != 2 {
			t.Errorf("Expected 2 tokens remaining, got %d", remaining)
		}
	})

	t.Run("ExhaustBucket", func(t *testing.T) {
		tb := NewTokenBucket(3, 1)

		// Consume all tokens
		for i := 0; i < 3; i++ {
			tb.Allow()
		}

		// Should not allow more requests
		if tb.Allow() {
			t.Error("Expected to deny request when bucket is empty")
		}
	})

	t.Run("TokenRefill", func(t *testing.T) {
		tb := NewTokenBucket(5, 2) // 2 tokens per second

		// Consume all tokens
		for i := 0; i < 5; i++ {
			tb.Allow()
		}

		// Wait for refill
		time.Sleep(1100 * time.Millisecond)

		// Should have refilled 2 tokens
		if !tb.Allow() {
			t.Error("Expected to allow request after refill")
		}
		if !tb.Allow() {
			t.Error("Expected to allow second request after refill")
		}
	})

	t.Log("✓ Token bucket works!")
}
