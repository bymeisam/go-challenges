package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestChiRouter(t *testing.T) {
	router := NewRouter()

	// Test home route
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var resp Response
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.Message != "Welcome to chi router" {
		t.Errorf("Unexpected response: %s", resp.Message)
	}

	// Test user route with parameter
	req = httptest.NewRequest("GET", "/users/123", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	json.NewDecoder(w.Body).Decode(&resp)
	if resp.UserID != "123" {
		t.Errorf("Expected user ID '123', got '%s'", resp.UserID)
	}

	// Test create user
	req = httptest.NewRequest("POST", "/users", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", w.Code)
	}

	// Test health endpoint
	req = httptest.NewRequest("GET", "/api/health", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	t.Log("✓ go-chi router works!")
}
