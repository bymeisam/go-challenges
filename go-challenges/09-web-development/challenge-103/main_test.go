package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHTTPServer(t *testing.T) {
	server := NewServer()

	// Test home route
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	server.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	if w.Body.String() != "Welcome to the home page" {
		t.Errorf("Unexpected response: %s", w.Body.String())
	}

	// Test health route
	req = httptest.NewRequest("GET", "/health", nil)
	w = httptest.NewRecorder()
	server.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var healthResp map[string]string
	json.NewDecoder(w.Body).Decode(&healthResp)
	if healthResp["status"] != "healthy" {
		t.Error("Health check failed")
	}

	// Test hello route
	req = httptest.NewRequest("GET", "/api/hello?name=Alice", nil)
	w = httptest.NewRecorder()
	server.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var helloResp map[string]string
	json.NewDecoder(w.Body).Decode(&helloResp)
	if helloResp["message"] != "Hello, Alice!" {
		t.Errorf("Unexpected message: %s", helloResp["message"])
	}

	// Test hello without name parameter
	req = httptest.NewRequest("GET", "/api/hello", nil)
	w = httptest.NewRecorder()
	server.ServeHTTP(w, req)

	json.NewDecoder(w.Body).Decode(&helloResp)
	if helloResp["message"] != "Hello, World!" {
		t.Errorf("Unexpected default message: %s", helloResp["message"])
	}

	t.Log("✓ Basic HTTP server works!")
}
