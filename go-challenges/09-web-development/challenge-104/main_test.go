package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHTTPHandlers(t *testing.T) {
	mux := NewMux()

	// Test custom handler
	req := httptest.NewRequest("GET", "/greet?name=Alice", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	if w.Body.String() != "Welcome, Alice!" {
		t.Errorf("Unexpected response: %s", w.Body.String())
	}

	// Test time handler
	req = httptest.NewRequest("GET", "/time", nil)
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	var timeResp map[string]string
	json.NewDecoder(w.Body).Decode(&timeResp)
	if timeResp["time"] == "" {
		t.Error("Time handler didn't return time")
	}

	// Test custom message handler
	req = httptest.NewRequest("GET", "/custom", nil)
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	var msgResp map[string]string
	json.NewDecoder(w.Body).Decode(&msgResp)
	if msgResp["message"] != "This is a custom message" {
		t.Error("Custom message handler failed")
	}

	// Test method handler - GET
	req = httptest.NewRequest("GET", "/method", nil)
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	var methodResp map[string]string
	json.NewDecoder(w.Body).Decode(&methodResp)
	if methodResp["method"] != "GET" {
		t.Error("Method handler GET failed")
	}

	// Test method handler - POST
	req = httptest.NewRequest("POST", "/method", nil)
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	json.NewDecoder(w.Body).Decode(&methodResp)
	if methodResp["method"] != "POST" {
		t.Error("Method handler POST failed")
	}

	t.Log("✓ HTTP handlers work!")
}
