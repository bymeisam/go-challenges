package main

import (
	"encoding/json"
	"encoding/xml"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestResponseWriter(t *testing.T) {
	rw := NewResponseWriter()

	t.Run("JSONResponse", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/json", nil)
		w := httptest.NewRecorder()

		rw.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		contentType := w.Header().Get("Content-Type")
		if contentType != "application/json" {
			t.Errorf("Expected Content-Type 'application/json', got '%s'", contentType)
		}

		var user User
		if err := json.NewDecoder(w.Body).Decode(&user); err != nil {
			t.Fatalf("Failed to decode JSON: %v", err)
		}

		if user.ID != 1 {
			t.Errorf("Expected ID 1, got %d", user.ID)
		}
		if user.Name != "John Doe" {
			t.Errorf("Expected name 'John Doe', got '%s'", user.Name)
		}
		if user.Email != "john@example.com" {
			t.Errorf("Expected email 'john@example.com', got '%s'", user.Email)
		}
	})

	t.Run("XMLResponse", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/xml", nil)
		w := httptest.NewRecorder()

		rw.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		contentType := w.Header().Get("Content-Type")
		if contentType != "application/xml" {
			t.Errorf("Expected Content-Type 'application/xml', got '%s'", contentType)
		}

		var user User
		if err := xml.NewDecoder(w.Body).Decode(&user); err != nil {
			t.Fatalf("Failed to decode XML: %v", err)
		}

		if user.ID != 2 {
			t.Errorf("Expected ID 2, got %d", user.ID)
		}
		if user.Name != "Jane Smith" {
			t.Errorf("Expected name 'Jane Smith', got '%s'", user.Name)
		}
		if user.Email != "jane@example.com" {
			t.Errorf("Expected email 'jane@example.com', got '%s'", user.Email)
		}
	})

	t.Run("HTMLResponse", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/html", nil)
		w := httptest.NewRecorder()

		rw.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		contentType := w.Header().Get("Content-Type")
		if !strings.Contains(contentType, "text/html") {
			t.Errorf("Expected Content-Type to contain 'text/html', got '%s'", contentType)
		}

		body := w.Body.String()

		// Check for key HTML elements
		if !strings.Contains(body, "<!DOCTYPE html>") {
			t.Error("Expected HTML doctype")
		}
		if !strings.Contains(body, "<h1>User Profile</h1>") {
			t.Error("Expected heading 'User Profile'")
		}
		if !strings.Contains(body, "Bob Johnson") {
			t.Error("Expected name 'Bob Johnson' in HTML")
		}
		if !strings.Contains(body, "bob@example.com") {
			t.Error("Expected email 'bob@example.com' in HTML")
		}
		if !strings.Contains(body, "bobjohnson") {
			t.Error("Expected username 'bobjohnson' in HTML")
		}
	})

	t.Log("✓ Response writing works!")
}
