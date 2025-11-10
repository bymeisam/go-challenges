package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func setupTestServer(t *testing.T) *APIServer {
	tmpFile := filepath.Join(t.TempDir(), "test.db")
	server, err := NewAPIServer(tmpFile, "test-secret")
	if err != nil {
		t.Fatalf("Failed to create test server: %v", err)
	}
	return server
}

func TestHealthEndpoint(t *testing.T) {
	server := setupTestServer(t)
	defer server.Close()

	req := httptest.NewRequest("GET", "/api/health", nil)
	w := httptest.NewRecorder()

	server.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]string
	json.NewDecoder(w.Body).Decode(&response)

	if response["status"] != "healthy" {
		t.Error("Expected healthy status")
	}
}

func TestRegister(t *testing.T) {
	server := setupTestServer(t)
	defer server.Close()

	tests := []struct {
		name       string
		payload    RegisterRequest
		wantStatus int
	}{
		{
			name:       "Valid registration",
			payload:    RegisterRequest{Username: "testuser", Password: "password123"},
			wantStatus: http.StatusCreated,
		},
		{
			name:       "Missing username",
			payload:    RegisterRequest{Username: "", Password: "password123"},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "Short password",
			payload:    RegisterRequest{Username: "testuser2", Password: "short"},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "Duplicate username",
			payload:    RegisterRequest{Username: "testuser", Password: "password123"},
			wantStatus: http.StatusConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.payload)
			req := httptest.NewRequest("POST", "/api/auth/register", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			server.router.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("Expected status %d, got %d", tt.wantStatus, w.Code)
			}
		})
	}
}

func TestLogin(t *testing.T) {
	server := setupTestServer(t)
	defer server.Close()

	// Register a user first
	registerReq := RegisterRequest{Username: "logintest", Password: "password123"}
	body, _ := json.Marshal(registerReq)
	req := httptest.NewRequest("POST", "/api/auth/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	tests := []struct {
		name       string
		payload    LoginRequest
		wantStatus int
	}{
		{
			name:       "Valid login",
			payload:    LoginRequest{Username: "logintest", Password: "password123"},
			wantStatus: http.StatusOK,
		},
		{
			name:       "Wrong password",
			payload:    LoginRequest{Username: "logintest", Password: "wrongpassword"},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "Non-existent user",
			payload:    LoginRequest{Username: "nonexistent", Password: "password123"},
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.payload)
			req := httptest.NewRequest("POST", "/api/auth/login", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			server.router.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("Expected status %d, got %d", tt.wantStatus, w.Code)
			}

			if tt.wantStatus == http.StatusOK {
				var response LoginResponse
				json.NewDecoder(w.Body).Decode(&response)

				if response.Token == "" {
					t.Error("Expected token in response")
				}

				if response.User.Username != tt.payload.Username {
					t.Errorf("Expected username %s, got %s", tt.payload.Username, response.User.Username)
				}
			}
		})
	}
}

func TestCreateArticle(t *testing.T) {
	server := setupTestServer(t)
	defer server.Close()

	// Get auth token
	token := registerAndLogin(t, server, "author", "password123")

	tests := []struct {
		name       string
		payload    CreateArticleRequest
		useAuth    bool
		wantStatus int
	}{
		{
			name: "Valid article",
			payload: CreateArticleRequest{
				Title:   "Test Article",
				Content: "This is test content",
				Author:  "author",
			},
			useAuth:    true,
			wantStatus: http.StatusCreated,
		},
		{
			name: "No authentication",
			payload: CreateArticleRequest{
				Title:   "Test Article",
				Content: "This is test content",
				Author:  "author",
			},
			useAuth:    false,
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "Missing title",
			payload: CreateArticleRequest{
				Title:   "",
				Content: "This is test content",
				Author:  "author",
			},
			useAuth:    true,
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.payload)
			req := httptest.NewRequest("POST", "/api/articles", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			if tt.useAuth {
				req.Header.Set("Authorization", "Bearer "+token)
			}
			w := httptest.NewRecorder()

			server.router.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("Expected status %d, got %d", tt.wantStatus, w.Code)
			}

			if tt.wantStatus == http.StatusCreated {
				var article Article
				json.NewDecoder(w.Body).Decode(&article)

				if article.Title != tt.payload.Title {
					t.Errorf("Expected title %s, got %s", tt.payload.Title, article.Title)
				}

				if article.ID == 0 {
					t.Error("Expected article to have ID")
				}
			}
		})
	}
}

func TestListArticles(t *testing.T) {
	server := setupTestServer(t)
	defer server.Close()

	token := registerAndLogin(t, server, "author", "password123")

	// Create some articles
	for i := 1; i <= 5; i++ {
		payload := CreateArticleRequest{
			Title:   "Article " + string(rune(i)),
			Content: "Content",
			Author:  "author",
		}
		body, _ := json.Marshal(payload)
		req := httptest.NewRequest("POST", "/api/articles", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		server.router.ServeHTTP(w, req)
	}

	tests := []struct {
		name      string
		url       string
		wantCount int
	}{
		{
			name:      "Default pagination",
			url:       "/api/articles",
			wantCount: 5,
		},
		{
			name:      "With limit",
			url:       "/api/articles?limit=3",
			wantCount: 3,
		},
		{
			name:      "Second page",
			url:       "/api/articles?page=2&limit=3",
			wantCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.url, nil)
			w := httptest.NewRecorder()

			server.router.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("Expected status 200, got %d", w.Code)
			}

			var response PaginatedResponse
			json.NewDecoder(w.Body).Decode(&response)

			if len(response.Data) != tt.wantCount {
				t.Errorf("Expected %d articles, got %d", tt.wantCount, len(response.Data))
			}

			if response.TotalCount != 5 {
				t.Errorf("Expected total count 5, got %d", response.TotalCount)
			}
		})
	}
}

func TestGetArticle(t *testing.T) {
	server := setupTestServer(t)
	defer server.Close()

	token := registerAndLogin(t, server, "author", "password123")

	// Create an article
	payload := CreateArticleRequest{
		Title:   "Test Article",
		Content: "Test Content",
		Author:  "author",
	}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/api/articles", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	var created Article
	json.NewDecoder(w.Body).Decode(&created)

	// Get the article
	req = httptest.NewRequest("GET", "/api/articles/"+string(rune(created.ID)), nil)
	w = httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var article Article
	json.NewDecoder(w.Body).Decode(&article)

	if article.Title != payload.Title {
		t.Errorf("Expected title %s, got %s", payload.Title, article.Title)
	}

	// Test non-existent article
	req = httptest.NewRequest("GET", "/api/articles/9999", nil)
	w = httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}
}

func TestUpdateArticle(t *testing.T) {
	server := setupTestServer(t)
	defer server.Close()

	token := registerAndLogin(t, server, "author", "password123")

	// Create an article
	createPayload := CreateArticleRequest{
		Title:   "Original Title",
		Content: "Original Content",
		Author:  "author",
	}
	body, _ := json.Marshal(createPayload)
	req := httptest.NewRequest("POST", "/api/articles", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	var created Article
	json.NewDecoder(w.Body).Decode(&created)

	// Update the article
	updatePayload := UpdateArticleRequest{
		Title:   "Updated Title",
		Content: "Updated Content",
	}
	body, _ = json.Marshal(updatePayload)
	req = httptest.NewRequest("PUT", "/api/articles/"+string(rune(created.ID)), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var updated Article
	json.NewDecoder(w.Body).Decode(&updated)

	if updated.Title != updatePayload.Title {
		t.Errorf("Expected title %s, got %s", updatePayload.Title, updated.Title)
	}
}

func TestDeleteArticle(t *testing.T) {
	server := setupTestServer(t)
	defer server.Close()

	token := registerAndLogin(t, server, "author", "password123")

	// Create an article
	createPayload := CreateArticleRequest{
		Title:   "To Delete",
		Content: "Content",
		Author:  "author",
	}
	body, _ := json.Marshal(createPayload)
	req := httptest.NewRequest("POST", "/api/articles", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	var created Article
	json.NewDecoder(w.Body).Decode(&created)

	// Delete the article
	req = httptest.NewRequest("DELETE", "/api/articles/"+string(rune(created.ID)), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Expected status 204, got %d", w.Code)
	}

	// Verify it's deleted
	req = httptest.NewRequest("GET", "/api/articles/"+string(rune(created.ID)), nil)
	w = httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}
}

func TestAuthMiddleware(t *testing.T) {
	server := setupTestServer(t)
	defer server.Close()

	payload := CreateArticleRequest{
		Title:   "Test",
		Content: "Content",
		Author:  "author",
	}
	body, _ := json.Marshal(payload)

	tests := []struct {
		name       string
		authHeader string
		wantStatus int
	}{
		{
			name:       "No auth header",
			authHeader: "",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "Invalid format",
			authHeader: "InvalidToken",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "Invalid token",
			authHeader: "Bearer invalid.token.here",
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/api/articles", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			w := httptest.NewRecorder()

			server.router.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("Expected status %d, got %d", tt.wantStatus, w.Code)
			}
		})
	}
}

// Helper function
func registerAndLogin(t *testing.T, server *APIServer, username, password string) string {
	// Register
	registerReq := RegisterRequest{Username: username, Password: password}
	body, _ := json.Marshal(registerReq)
	req := httptest.NewRequest("POST", "/api/auth/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	// Login
	loginReq := LoginRequest{Username: username, Password: password}
	body, _ = json.Marshal(loginReq)
	req = httptest.NewRequest("POST", "/api/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	var response LoginResponse
	json.NewDecoder(w.Body).Decode(&response)

	return response.Token
}

func TestDatabasePersistence(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "test.db")

	// Create server and add data
	server1, _ := NewAPIServer(tmpFile, "test-secret")
	token := registerAndLogin(t, server1, "testuser", "password123")

	payload := CreateArticleRequest{
		Title:   "Persistent Article",
		Content: "Content",
		Author:  "testuser",
	}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/api/articles", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	server1.router.ServeHTTP(w, req)
	server1.Close()

	// Create new server instance with same database
	server2, err := NewAPIServer(tmpFile, "test-secret")
	if err != nil {
		t.Fatalf("Failed to reopen database: %v", err)
	}
	defer server2.Close()

	// Verify data persisted
	req = httptest.NewRequest("GET", "/api/articles", nil)
	w = httptest.NewRecorder()
	server2.router.ServeHTTP(w, req)

	var response PaginatedResponse
	json.NewDecoder(w.Body).Decode(&response)

	if len(response.Data) != 1 {
		t.Errorf("Expected 1 article, got %d", len(response.Data))
	}

	if response.Data[0].Title != payload.Title {
		t.Error("Article data not persisted correctly")
	}
}

func TestServerCreation(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "test.db")

	server, err := NewAPIServer(tmpFile, "test-secret")
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}
	defer server.Close()

	if server.db == nil {
		t.Error("Database should be initialized")
	}

	if server.router == nil {
		t.Error("Router should be initialized")
	}

	// Verify database file was created
	if _, err := os.Stat(tmpFile); os.IsNotExist(err) {
		t.Error("Database file was not created")
	}
}
