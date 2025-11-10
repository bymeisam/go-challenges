package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

// User represents a user in the system
type User struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password,omitempty"` // In production, store hashed!
}

// UserRepository interface for data access
type UserRepository interface {
	Create(user User) error
	GetByID(id int) (User, error)
	GetByEmail(email string) (User, error)
	Delete(id int) error
}

// InMemoryUserRepository is an in-memory implementation for testing
type InMemoryUserRepository struct {
	mu      sync.RWMutex
	users   map[int]User
	nextID  int
	byEmail map[string]int
}

func NewInMemoryUserRepository() *InMemoryUserRepository {
	return &InMemoryUserRepository{
		users:   make(map[int]User),
		nextID:  1,
		byEmail: make(map[string]int),
	}
}

func (r *InMemoryUserRepository) Create(user User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if email already exists
	if _, exists := r.byEmail[user.Email]; exists {
		return errors.New("email already exists")
	}

	user.ID = r.nextID
	r.users[user.ID] = user
	r.byEmail[user.Email] = user.ID
	r.nextID++

	return nil
}

func (r *InMemoryUserRepository) GetByID(id int) (User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, exists := r.users[id]
	if !exists {
		return User{}, errors.New("user not found")
	}

	return user, nil
}

func (r *InMemoryUserRepository) GetByEmail(email string) (User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	id, exists := r.byEmail[email]
	if !exists {
		return User{}, errors.New("user not found")
	}

	return r.users[id], nil
}

func (r *InMemoryUserRepository) Delete(id int) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	user, exists := r.users[id]
	if !exists {
		return errors.New("user not found")
	}

	delete(r.users, id)
	delete(r.byEmail, user.Email)

	return nil
}

// UserService contains business logic
type UserService struct {
	repo UserRepository
}

func NewUserService(repo UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) RegisterUser(name, email, password string) (User, error) {
	// Validate input
	if name == "" || email == "" || password == "" {
		return User{}, errors.New("name, email, and password are required")
	}

	if !strings.Contains(email, "@") {
		return User{}, errors.New("invalid email format")
	}

	if len(password) < 6 {
		return User{}, errors.New("password must be at least 6 characters")
	}

	// Create user
	user := User{
		Name:     name,
		Email:    email,
		Password: password, // In production: hash the password!
	}

	if err := s.repo.Create(user); err != nil {
		return User{}, err
	}

	// Get created user with ID
	return s.repo.GetByEmail(email)
}

func (s *UserService) Authenticate(email, password string) (User, error) {
	user, err := s.repo.GetByEmail(email)
	if err != nil {
		return User{}, errors.New("invalid credentials")
	}

	// In production: compare hashed passwords
	if user.Password != password {
		return User{}, errors.New("invalid credentials")
	}

	return user, nil
}

func (s *UserService) GetUserProfile(id int) (User, error) {
	user, err := s.repo.GetByID(id)
	if err != nil {
		return User{}, err
	}

	// Don't return password in profile
	user.Password = ""
	return user, nil
}

// HTTPServer handles HTTP requests
type HTTPServer struct {
	service *UserService
	mux     *http.ServeMux
}

func NewHTTPServer(service *UserService) *HTTPServer {
	server := &HTTPServer{
		service: service,
		mux:     http.NewServeMux(),
	}

	server.mux.HandleFunc("/register", server.HandleRegister)
	server.mux.HandleFunc("/login", server.HandleLogin)
	server.mux.HandleFunc("/profile/", server.HandleProfile)

	return server
}

func (s *HTTPServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func (s *HTTPServer) HandleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	user, err := s.service.RegisterUser(req.Name, req.Email, req.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Don't return password
	user.Password = ""

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

func (s *HTTPServer) HandleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	user, err := s.service.Authenticate(req.Email, req.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	// Don't return password
	user.Password = ""

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func (s *HTTPServer) HandleProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract ID from path: /profile/123
	path := strings.TrimPrefix(r.URL.Path, "/profile/")
	id, err := strconv.Atoi(path)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	user, err := s.service.GetUserProfile(id)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// ExternalAPIClient simulates an external API client
type ExternalAPIClient struct {
	BaseURL string
}

func NewExternalAPIClient(baseURL string) *ExternalAPIClient {
	return &ExternalAPIClient{BaseURL: baseURL}
}

func (c *ExternalAPIClient) FetchUserData(userID int) (map[string]interface{}, error) {
	url := fmt.Sprintf("%s/users/%d", c.BaseURL, userID)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	var data map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	return data, nil
}

func main() {
	// Example usage
	repo := NewInMemoryUserRepository()
	service := NewUserService(repo)

	// Register a user
	user, err := service.RegisterUser("John Doe", "john@example.com", "password123")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Registered user: %s (ID: %d)\n", user.Name, user.ID)

	// Start HTTP server
	server := NewHTTPServer(service)
	fmt.Println("Server would start on :8080")
	_ = server // In production: http.ListenAndServe(":8080", server)
}
