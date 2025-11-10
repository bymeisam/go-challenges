package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// Unit Tests - Repository Layer

func TestInMemoryUserRepository_Create(t *testing.T) {
	repo := NewInMemoryUserRepository()

	user := User{
		Name:     "Alice",
		Email:    "alice@example.com",
		Password: "password123",
	}

	err := repo.Create(user)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Verify user was created with ID
	created, err := repo.GetByEmail("alice@example.com")
	if err != nil {
		t.Fatalf("GetByEmail failed: %v", err)
	}

	if created.ID == 0 {
		t.Error("user ID should be assigned")
	}

	if created.Name != "Alice" {
		t.Errorf("name = %q; want %q", created.Name, "Alice")
	}

	t.Log("✓ Repository Create test passed!")
}

func TestInMemoryUserRepository_DuplicateEmail(t *testing.T) {
	repo := NewInMemoryUserRepository()

	user := User{Name: "Bob", Email: "bob@example.com", Password: "pass"}

	err := repo.Create(user)
	if err != nil {
		t.Fatalf("first Create failed: %v", err)
	}

	// Try to create another user with same email
	err = repo.Create(user)
	if err == nil {
		t.Fatal("expected error for duplicate email")
	}

	t.Log("✓ Repository duplicate email test passed!")
}

// Unit Tests - Service Layer

func TestUserService_RegisterUser(t *testing.T) {
	repo := NewInMemoryUserRepository()
	service := NewUserService(repo)

	tests := []struct {
		name        string
		userName    string
		email       string
		password    string
		expectError bool
	}{
		{"valid user", "Charlie", "charlie@example.com", "password123", false},
		{"empty name", "", "test@example.com", "password123", true},
		{"invalid email", "Dave", "notanemail", "password123", true},
		{"short password", "Eve", "eve@example.com", "123", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := service.RegisterUser(tt.userName, tt.email, tt.password)

			if tt.expectError {
				if err == nil {
					t.Error("expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if user.ID == 0 {
					t.Error("user ID should be assigned")
				}
			}
		})
	}

	t.Log("✓ Service RegisterUser tests passed!")
}

func TestUserService_Authenticate(t *testing.T) {
	repo := NewInMemoryUserRepository()
	service := NewUserService(repo)

	// Register a user
	_, err := service.RegisterUser("Frank", "frank@example.com", "correctpassword")
	if err != nil {
		t.Fatalf("RegisterUser failed: %v", err)
	}

	t.Run("valid credentials", func(t *testing.T) {
		user, err := service.Authenticate("frank@example.com", "correctpassword")
		if err != nil {
			t.Errorf("Authenticate failed: %v", err)
		}
		if user.Email != "frank@example.com" {
			t.Errorf("user email = %q; want %q", user.Email, "frank@example.com")
		}
	})

	t.Run("wrong password", func(t *testing.T) {
		_, err := service.Authenticate("frank@example.com", "wrongpassword")
		if err == nil {
			t.Error("expected error for wrong password")
		}
	})

	t.Run("non-existent user", func(t *testing.T) {
		_, err := service.Authenticate("nobody@example.com", "password")
		if err == nil {
			t.Error("expected error for non-existent user")
		}
	})

	t.Log("✓ Service Authenticate tests passed!")
}

// Integration Tests - HTTP Layer

func TestHTTPServer_Register(t *testing.T) {
	repo := NewInMemoryUserRepository()
	service := NewUserService(repo)
	server := NewHTTPServer(service)

	t.Run("successful registration", func(t *testing.T) {
		body := bytes.NewBufferString(`{"name":"Grace","email":"grace@example.com","password":"password123"}`)
		req := httptest.NewRequest(http.MethodPost, "/register", body)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		server.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("status = %d; want %d", w.Code, http.StatusCreated)
		}

		var user User
		if err := json.NewDecoder(w.Body).Decode(&user); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if user.Name != "Grace" {
			t.Errorf("name = %q; want %q", user.Name, "Grace")
		}

		if user.Password != "" {
			t.Error("password should not be returned")
		}
	})

	t.Run("invalid request", func(t *testing.T) {
		body := bytes.NewBufferString(`{"name":"","email":"invalid","password":""}`)
		req := httptest.NewRequest(http.MethodPost, "/register", body)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		server.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("status = %d; want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Log("✓ HTTP Register tests passed!")
}

func TestHTTPServer_Login(t *testing.T) {
	repo := NewInMemoryUserRepository()
	service := NewUserService(repo)
	server := NewHTTPServer(service)

	// Register a user first
	service.RegisterUser("Henry", "henry@example.com", "password123")

	t.Run("successful login", func(t *testing.T) {
		body := bytes.NewBufferString(`{"email":"henry@example.com","password":"password123"}`)
		req := httptest.NewRequest(http.MethodPost, "/login", body)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		server.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("status = %d; want %d", w.Code, http.StatusOK)
		}

		var user User
		json.NewDecoder(w.Body).Decode(&user)

		if user.Email != "henry@example.com" {
			t.Errorf("email = %q; want %q", user.Email, "henry@example.com")
		}
	})

	t.Run("failed login", func(t *testing.T) {
		body := bytes.NewBufferString(`{"email":"henry@example.com","password":"wrongpassword"}`)
		req := httptest.NewRequest(http.MethodPost, "/login", body)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		server.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("status = %d; want %d", w.Code, http.StatusUnauthorized)
		}
	})

	t.Log("✓ HTTP Login tests passed!")
}

func TestHTTPServer_Profile(t *testing.T) {
	repo := NewInMemoryUserRepository()
	service := NewUserService(repo)
	server := NewHTTPServer(service)

	// Register a user
	user, _ := service.RegisterUser("Iris", "iris@example.com", "password123")

	t.Run("get profile", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/profile/%d", user.ID), nil)
		w := httptest.NewRecorder()

		server.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("status = %d; want %d", w.Code, http.StatusOK)
		}

		var profile User
		json.NewDecoder(w.Body).Decode(&profile)

		if profile.Name != "Iris" {
			t.Errorf("name = %q; want %q", profile.Name, "Iris")
		}

		if profile.Password != "" {
			t.Error("password should not be in profile")
		}
	})

	t.Log("✓ HTTP Profile tests passed!")
}

// Integration Test - External API Mock

func TestExternalAPIClient(t *testing.T) {
	// Create mock external API server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/users/123" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"id":123,"name":"Mock User","status":"active"}`))
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer mockServer.Close()

	client := NewExternalAPIClient(mockServer.URL)

	t.Run("fetch existing user", func(t *testing.T) {
		data, err := client.FetchUserData(123)
		if err != nil {
			t.Fatalf("FetchUserData failed: %v", err)
		}

		if data["name"] != "Mock User" {
			t.Errorf("name = %v; want %q", data["name"], "Mock User")
		}
	})

	t.Run("fetch non-existent user", func(t *testing.T) {
		_, err := client.FetchUserData(999)
		if err == nil {
			t.Error("expected error for non-existent user")
		}
	})

	t.Log("✓ External API Client tests passed!")
}

// End-to-End Integration Test

func TestEndToEnd_UserRegistrationAndLogin(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping end-to-end test in short mode")
	}

	// Setup complete system
	repo := NewInMemoryUserRepository()
	service := NewUserService(repo)
	server := NewHTTPServer(service)

	// Step 1: Register user
	registerBody := bytes.NewBufferString(`{"name":"Jack","email":"jack@example.com","password":"password123"}`)
	registerReq := httptest.NewRequest(http.MethodPost, "/register", registerBody)
	registerReq.Header.Set("Content-Type", "application/json")
	registerW := httptest.NewRecorder()

	server.ServeHTTP(registerW, registerReq)

	if registerW.Code != http.StatusCreated {
		t.Fatalf("registration failed with status %d", registerW.Code)
	}

	var registeredUser User
	json.NewDecoder(registerW.Body).Decode(&registeredUser)

	// Step 2: Login
	loginBody := bytes.NewBufferString(`{"email":"jack@example.com","password":"password123"}`)
	loginReq := httptest.NewRequest(http.MethodPost, "/login", loginBody)
	loginReq.Header.Set("Content-Type", "application/json")
	loginW := httptest.NewRecorder()

	server.ServeHTTP(loginW, loginReq)

	if loginW.Code != http.StatusOK {
		t.Fatalf("login failed with status %d", loginW.Code)
	}

	var loggedInUser User
	json.NewDecoder(loginW.Body).Decode(&loggedInUser)

	// Step 3: Get profile
	profileReq := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/profile/%d", registeredUser.ID), nil)
	profileW := httptest.NewRecorder()

	server.ServeHTTP(profileW, profileReq)

	if profileW.Code != http.StatusOK {
		t.Fatalf("get profile failed with status %d", profileW.Code)
	}

	var profile User
	json.NewDecoder(profileW.Body).Decode(&profile)

	// Verify complete flow
	if profile.Name != "Jack" {
		t.Errorf("profile name = %q; want %q", profile.Name, "Jack")
	}

	t.Log("✓ End-to-end test passed!")
}

// Performance/Integration Test

func TestConcurrentRegistrations(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping concurrent test in short mode")
	}

	repo := NewInMemoryUserRepository()
	service := NewUserService(repo)

	// Register 100 users concurrently
	done := make(chan bool)
	errors := make(chan error, 100)

	for i := 0; i < 100; i++ {
		go func(id int) {
			name := "User" + string(rune(id))
			email := "user" + string(rune(id)) + "@example.com"
			_, err := service.RegisterUser(name, email, "password123")
			if err != nil {
				errors <- err
			}
			done <- true
		}(i)
	}

	// Wait for all to complete
	timeout := time.After(5 * time.Second)
	for i := 0; i < 100; i++ {
		select {
		case <-done:
			// Success
		case <-timeout:
			t.Fatal("test timed out")
		}
	}

	close(errors)

	// Check for errors
	errorCount := 0
	for err := range errors {
		t.Logf("error: %v", err)
		errorCount++
	}

	if errorCount > 0 {
		t.Errorf("got %d errors during concurrent registrations", errorCount)
	}

	t.Log("✓ Concurrent registrations test passed!")
}
