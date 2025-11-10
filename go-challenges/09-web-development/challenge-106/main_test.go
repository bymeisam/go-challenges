package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRouter(t *testing.T) {
	router := NewRouter()

	// Define routes
	router.Get("/users/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"user_id": id,
		})
	})

	router.Get("/posts/{postId}/comments/{commentId}", func(w http.ResponseWriter, r *http.Request) {
		postId := r.URL.Query().Get("postId")
		commentId := r.URL.Query().Get("commentId")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"post_id":    postId,
			"comment_id": commentId,
		})
	})

	router.Post("/users", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "User created",
		})
	})

	router.Delete("/users/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "User deleted",
			"id":      id,
		})
	})

	// Test GET with single parameter
	req := httptest.NewRequest("GET", "/users/123", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var resp map[string]string
	json.NewDecoder(w.Body).Decode(&resp)
	if resp["user_id"] != "123" {
		t.Errorf("Expected user_id 123, got %s", resp["user_id"])
	}

	// Test GET with multiple parameters
	req = httptest.NewRequest("GET", "/posts/456/comments/789", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	json.NewDecoder(w.Body).Decode(&resp)
	if resp["post_id"] != "456" {
		t.Errorf("Expected post_id 456, got %s", resp["post_id"])
	}
	if resp["comment_id"] != "789" {
		t.Errorf("Expected comment_id 789, got %s", resp["comment_id"])
	}

	// Test POST
	req = httptest.NewRequest("POST", "/users", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", w.Code)
	}

	// Test DELETE
	req = httptest.NewRequest("DELETE", "/users/999", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	json.NewDecoder(w.Body).Decode(&resp)
	if resp["id"] != "999" {
		t.Errorf("Expected id 999, got %s", resp["id"])
	}

	// Test 404
	req = httptest.NewRequest("GET", "/notfound", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}

	// Test wrong method
	req = httptest.NewRequest("POST", "/users/123", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404 for wrong method, got %d", w.Code)
	}

	t.Log("✓ Routing with pattern matching and URL parameters works!")
}
