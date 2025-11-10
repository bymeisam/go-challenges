package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/golang-jwt/jwt/v5"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

// Models
type User struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Password string `json:"-"`
}

type Article struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	Author    string    `json:"author"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Request/Response types
type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

type CreateArticleRequest struct {
	Title   string `json:"title"`
	Content string `json:"content"`
	Author  string `json:"author"`
}

type UpdateArticleRequest struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

type PaginatedResponse struct {
	Data       []Article `json:"data"`
	Page       int       `json:"page"`
	Limit      int       `json:"limit"`
	TotalCount int       `json:"total_count"`
	TotalPages int       `json:"total_pages"`
}

// API Server
type APIServer struct {
	db        *sql.DB
	router    chi.Router
	jwtSecret []byte
}

func NewAPIServer(dbPath string, jwtSecret string) (*APIServer, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	server := &APIServer{
		db:        db,
		router:    chi.NewRouter(),
		jwtSecret: []byte(jwtSecret),
	}

	if err := server.initDB(); err != nil {
		return nil, err
	}

	server.setupRoutes()
	return server, nil
}

func (s *APIServer) initDB() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT UNIQUE NOT NULL,
			password TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS articles (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			content TEXT NOT NULL,
			author TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
	}

	for _, query := range queries {
		if _, err := s.db.Exec(query); err != nil {
			return err
		}
	}

	return nil
}

func (s *APIServer) setupRoutes() {
	// Middleware
	s.router.Use(middleware.Logger)
	s.router.Use(middleware.Recoverer)
	s.router.Use(middleware.RequestID)
	s.router.Use(middleware.RealIP)
	s.router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Routes
	s.router.Route("/api", func(r chi.Router) {
		// Public routes
		r.Get("/health", s.handleHealth)

		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", s.handleRegister)
			r.Post("/login", s.handleLogin)
		})

		// Protected routes
		r.Route("/articles", func(r chi.Router) {
			r.Get("/", s.handleListArticles)
			r.Get("/{id}", s.handleGetArticle)

			// Authenticated endpoints
			r.Group(func(r chi.Router) {
				r.Use(s.authMiddleware)
				r.Post("/", s.handleCreateArticle)
				r.Put("/{id}", s.handleUpdateArticle)
				r.Delete("/{id}", s.handleDeleteArticle)
			})
		})
	})
}

// Middleware
func (s *APIServer) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			s.respondError(w, http.StatusUnauthorized, "missing authorization header")
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			s.respondError(w, http.StatusUnauthorized, "invalid authorization header format")
			return
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method")
			}
			return s.jwtSecret, nil
		})

		if err != nil || !token.Valid {
			s.respondError(w, http.StatusUnauthorized, "invalid token")
			return
		}

		next.ServeHTTP(w, r)
	})
}

// Handlers
func (s *APIServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	s.respondJSON(w, http.StatusOK, map[string]string{
		"status": "healthy",
		"time":   time.Now().Format(time.RFC3339),
	})
}

func (s *APIServer) handleRegister(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Username == "" || req.Password == "" {
		s.respondError(w, http.StatusBadRequest, "username and password required")
		return
	}

	if len(req.Password) < 6 {
		s.respondError(w, http.StatusBadRequest, "password must be at least 6 characters")
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		s.respondError(w, http.StatusInternalServerError, "failed to hash password")
		return
	}

	result, err := s.db.Exec("INSERT INTO users (username, password) VALUES (?, ?)",
		req.Username, string(hashedPassword))
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			s.respondError(w, http.StatusConflict, "username already exists")
			return
		}
		s.respondError(w, http.StatusInternalServerError, "failed to create user")
		return
	}

	id, _ := result.LastInsertId()
	user := User{ID: int(id), Username: req.Username}

	s.respondJSON(w, http.StatusCreated, user)
}

func (s *APIServer) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	var user User
	err := s.db.QueryRow("SELECT id, username, password FROM users WHERE username = ?",
		req.Username).Scan(&user.ID, &user.Username, &user.Password)
	if err != nil {
		s.respondError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		s.respondError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
		"exp":      time.Now().Add(24 * time.Hour).Unix(),
	})

	tokenString, err := token.SignedString(s.jwtSecret)
	if err != nil {
		s.respondError(w, http.StatusInternalServerError, "failed to generate token")
		return
	}

	s.respondJSON(w, http.StatusOK, LoginResponse{
		Token: tokenString,
		User:  User{ID: user.ID, Username: user.Username},
	})
}

func (s *APIServer) handleListArticles(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 10
	}

	offset := (page - 1) * limit

	// Get total count
	var totalCount int
	err := s.db.QueryRow("SELECT COUNT(*) FROM articles").Scan(&totalCount)
	if err != nil {
		s.respondError(w, http.StatusInternalServerError, "failed to count articles")
		return
	}

	// Get articles
	rows, err := s.db.Query(`
		SELECT id, title, content, author, created_at, updated_at
		FROM articles
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?`, limit, offset)
	if err != nil {
		s.respondError(w, http.StatusInternalServerError, "failed to fetch articles")
		return
	}
	defer rows.Close()

	articles := make([]Article, 0)
	for rows.Next() {
		var article Article
		if err := rows.Scan(&article.ID, &article.Title, &article.Content,
			&article.Author, &article.CreatedAt, &article.UpdatedAt); err != nil {
			continue
		}
		articles = append(articles, article)
	}

	totalPages := (totalCount + limit - 1) / limit

	s.respondJSON(w, http.StatusOK, PaginatedResponse{
		Data:       articles,
		Page:       page,
		Limit:      limit,
		TotalCount: totalCount,
		TotalPages: totalPages,
	})
}

func (s *APIServer) handleGetArticle(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var article Article
	err := s.db.QueryRow(`
		SELECT id, title, content, author, created_at, updated_at
		FROM articles WHERE id = ?`, id).Scan(
		&article.ID, &article.Title, &article.Content,
		&article.Author, &article.CreatedAt, &article.UpdatedAt)

	if err == sql.ErrNoRows {
		s.respondError(w, http.StatusNotFound, "article not found")
		return
	}

	if err != nil {
		s.respondError(w, http.StatusInternalServerError, "failed to fetch article")
		return
	}

	s.respondJSON(w, http.StatusOK, article)
}

func (s *APIServer) handleCreateArticle(w http.ResponseWriter, r *http.Request) {
	var req CreateArticleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Title == "" || req.Content == "" || req.Author == "" {
		s.respondError(w, http.StatusBadRequest, "title, content, and author required")
		return
	}

	result, err := s.db.Exec(`
		INSERT INTO articles (title, content, author)
		VALUES (?, ?, ?)`, req.Title, req.Content, req.Author)
	if err != nil {
		s.respondError(w, http.StatusInternalServerError, "failed to create article")
		return
	}

	id, _ := result.LastInsertId()

	var article Article
	s.db.QueryRow(`
		SELECT id, title, content, author, created_at, updated_at
		FROM articles WHERE id = ?`, id).Scan(
		&article.ID, &article.Title, &article.Content,
		&article.Author, &article.CreatedAt, &article.UpdatedAt)

	s.respondJSON(w, http.StatusCreated, article)
}

func (s *APIServer) handleUpdateArticle(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req UpdateArticleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	result, err := s.db.Exec(`
		UPDATE articles
		SET title = ?, content = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?`, req.Title, req.Content, id)
	if err != nil {
		s.respondError(w, http.StatusInternalServerError, "failed to update article")
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		s.respondError(w, http.StatusNotFound, "article not found")
		return
	}

	var article Article
	s.db.QueryRow(`
		SELECT id, title, content, author, created_at, updated_at
		FROM articles WHERE id = ?`, id).Scan(
		&article.ID, &article.Title, &article.Content,
		&article.Author, &article.CreatedAt, &article.UpdatedAt)

	s.respondJSON(w, http.StatusOK, article)
}

func (s *APIServer) handleDeleteArticle(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	result, err := s.db.Exec("DELETE FROM articles WHERE id = ?", id)
	if err != nil {
		s.respondError(w, http.StatusInternalServerError, "failed to delete article")
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		s.respondError(w, http.StatusNotFound, "article not found")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Helper methods
func (s *APIServer) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (s *APIServer) respondError(w http.ResponseWriter, status int, message string) {
	s.respondJSON(w, status, ErrorResponse{
		Error:   http.StatusText(status),
		Message: message,
	})
}

func (s *APIServer) Start(addr string) error {
	log.Printf("Starting server on %s", addr)
	return http.ListenAndServe(addr, s.router)
}

func (s *APIServer) Close() error {
	return s.db.Close()
}

func main() {
	server, err := NewAPIServer("blog.db", "your-secret-key-change-in-production")
	if err != nil {
		log.Fatal(err)
	}
	defer server.Close()

	// Graceful shutdown
	serverErrors := make(chan error, 1)
	go func() {
		serverErrors <- server.Start(":8080")
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		log.Fatalf("Server error: %v", err)
	case sig := <-shutdown:
		log.Printf("Received signal %v, shutting down gracefully...", sig)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Cleanup
		done := make(chan struct{})
		go func() {
			server.Close()
			close(done)
		}()

		select {
		case <-done:
			log.Println("Shutdown complete")
		case <-ctx.Done():
			log.Println("Shutdown timeout exceeded")
		}
	}
}
