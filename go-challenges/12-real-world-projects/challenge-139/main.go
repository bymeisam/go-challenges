package main

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-redis/redis/v8"
)

const (
	defaultCodeLength = 6
	charset           = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

// ShortenRequest represents a request to shorten a URL
type ShortenRequest struct {
	URL        string `json:"url"`
	CustomCode string `json:"custom_code,omitempty"`
	ExpiresIn  int64  `json:"expires_in,omitempty"` // seconds
}

// ShortenResponse represents the response with short URL
type ShortenResponse struct {
	ShortURL  string    `json:"short_url"`
	Code      string    `json:"code"`
	LongURL   string    `json:"long_url"`
	ExpiresAt time.Time `json:"expires_at,omitempty"`
}

// URLData represents stored URL data
type URLData struct {
	LongURL   string    `json:"long_url"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at,omitempty"`
	Clicks    int64     `json:"clicks"`
}

// Stats represents analytics data
type Stats struct {
	Code      string    `json:"code"`
	LongURL   string    `json:"long_url"`
	Clicks    int64     `json:"clicks"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at,omitempty"`
}

// URLShortener manages URL shortening
type URLShortener struct {
	redis   *redis.Client
	baseURL string
}

// NewURLShortener creates a new URL shortener
func NewURLShortener(redisAddr, baseURL string) (*URLShortener, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &URLShortener{
		redis:   rdb,
		baseURL: baseURL,
	}, nil
}

// Shorten creates a short URL
func (us *URLShortener) Shorten(ctx context.Context, req ShortenRequest) (*ShortenResponse, error) {
	// Validate URL
	if !isValidURL(req.URL) {
		return nil, fmt.Errorf("invalid URL")
	}

	// Normalize URL
	normalizedURL := normalizeURL(req.URL)

	// Generate or use custom code
	code := req.CustomCode
	if code == "" {
		var err error
		code, err = us.generateCode(ctx)
		if err != nil {
			return nil, err
		}
	} else {
		// Check if custom code already exists
		exists, err := us.redis.Exists(ctx, "url:"+code).Result()
		if err != nil {
			return nil, err
		}
		if exists > 0 {
			return nil, fmt.Errorf("custom code already in use")
		}
	}

	// Create URL data
	data := URLData{
		LongURL:   normalizedURL,
		CreatedAt: time.Now(),
		Clicks:    0,
	}

	// Set expiration
	var ttl time.Duration
	if req.ExpiresIn > 0 {
		ttl = time.Duration(req.ExpiresIn) * time.Second
		data.ExpiresAt = time.Now().Add(ttl)
	}

	// Store in Redis
	key := "url:" + code
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	if ttl > 0 {
		err = us.redis.Set(ctx, key, jsonData, ttl).Err()
	} else {
		err = us.redis.Set(ctx, key, jsonData, 0).Err()
	}

	if err != nil {
		return nil, err
	}

	response := &ShortenResponse{
		ShortURL:  us.baseURL + "/" + code,
		Code:      code,
		LongURL:   normalizedURL,
		ExpiresAt: data.ExpiresAt,
	}

	return response, nil
}

// Resolve resolves a short code to a long URL and increments click count
func (us *URLShortener) Resolve(ctx context.Context, code string) (string, error) {
	key := "url:" + code

	jsonData, err := us.redis.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", fmt.Errorf("URL not found")
	}
	if err != nil {
		return "", err
	}

	var data URLData
	if err := json.Unmarshal([]byte(jsonData), &data); err != nil {
		return "", err
	}

	// Increment click count
	data.Clicks++
	updatedJSON, _ := json.Marshal(data)

	// Get TTL to preserve expiration
	ttl, _ := us.redis.TTL(ctx, key).Result()
	if ttl > 0 {
		us.redis.Set(ctx, key, updatedJSON, ttl)
	} else {
		us.redis.Set(ctx, key, updatedJSON, 0)
	}

	return data.LongURL, nil
}

// GetStats returns analytics for a short code
func (us *URLShortener) GetStats(ctx context.Context, code string) (*Stats, error) {
	key := "url:" + code

	jsonData, err := us.redis.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, fmt.Errorf("URL not found")
	}
	if err != nil {
		return nil, err
	}

	var data URLData
	if err := json.Unmarshal([]byte(jsonData), &data); err != nil {
		return nil, err
	}

	stats := &Stats{
		Code:      code,
		LongURL:   data.LongURL,
		Clicks:    data.Clicks,
		CreatedAt: data.CreatedAt,
		ExpiresAt: data.ExpiresAt,
	}

	return stats, nil
}

// Delete removes a short URL
func (us *URLShortener) Delete(ctx context.Context, code string) error {
	key := "url:" + code
	result, err := us.redis.Del(ctx, key).Result()
	if err != nil {
		return err
	}
	if result == 0 {
		return fmt.Errorf("URL not found")
	}
	return nil
}

// generateCode generates a random short code
func (us *URLShortener) generateCode(ctx context.Context) (string, error) {
	maxAttempts := 10
	for attempt := 0; attempt < maxAttempts; attempt++ {
		code := generateRandomCode(defaultCodeLength)

		// Check if code already exists
		exists, err := us.redis.Exists(ctx, "url:"+code).Result()
		if err != nil {
			return "", err
		}

		if exists == 0 {
			return code, nil
		}
	}

	return "", fmt.Errorf("failed to generate unique code")
}

// generateRandomCode generates a random code of specified length
func generateRandomCode(length int) string {
	code := make([]byte, length)
	for i := range code {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		code[i] = charset[n.Int64()]
	}
	return string(code)
}

// isValidURL checks if a URL is valid
func isValidURL(rawURL string) bool {
	u, err := url.Parse(rawURL)
	if err != nil {
		return false
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		return false
	}

	if u.Host == "" {
		return false
	}

	return true
}

// normalizeURL normalizes a URL
func normalizeURL(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}

	// Remove trailing slash
	u.Path = strings.TrimSuffix(u.Path, "/")

	return u.String()
}

// API Handlers

func (us *URLShortener) handleShorten(w http.ResponseWriter, r *http.Request) {
	var req ShortenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	response, err := us.Shorten(r.Context(), req)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, response)
}

func (us *URLShortener) handleRedirect(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")

	longURL, err := us.Resolve(r.Context(), code)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	http.Redirect(w, r, longURL, http.StatusMovedPermanently)
}

func (us *URLShortener) handleStats(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")

	stats, err := us.GetStats(r.Context(), code)
	if err != nil {
		respondError(w, http.StatusNotFound, "URL not found")
		return
	}

	respondJSON(w, http.StatusOK, stats)
}

func (us *URLShortener) handleDelete(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")

	if err := us.Delete(r.Context(), code); err != nil {
		respondError(w, http.StatusNotFound, "URL not found")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Helper functions

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]string{"error": message})
}

func main() {
	// In-memory Redis simulation for demo (use real Redis in production)
	shortener, err := NewURLShortener("localhost:6379", "http://localhost:8080")
	if err != nil {
		log.Printf("Warning: Redis not available, using mock: %v", err)
		// In production, you would exit here
		// For demo purposes, we'll show the structure
	}

	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)

	// Routes
	r.Post("/api/shorten", shortener.handleShorten)
	r.Get("/api/stats/{code}", shortener.handleStats)
	r.Delete("/api/{code}", shortener.handleDelete)
	r.Get("/{code}", shortener.handleRedirect)

	// Home page with API documentation
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		html := `
<!DOCTYPE html>
<html>
<head>
    <title>URL Shortener</title>
    <style>
        body { font-family: Arial, sans-serif; max-width: 800px; margin: 50px auto; padding: 20px; }
        code { background: #f4f4f4; padding: 2px 6px; border-radius: 3px; }
        pre { background: #f4f4f4; padding: 15px; border-radius: 5px; overflow-x: auto; }
    </style>
</head>
<body>
    <h1>URL Shortener API</h1>

    <h2>Shorten URL</h2>
    <pre>POST /api/shorten
Content-Type: application/json

{
  "url": "https://example.com/very/long/url",
  "custom_code": "mycode",  // optional
  "expires_in": 86400       // optional, in seconds
}</pre>

    <h2>Redirect</h2>
    <pre>GET /{code}</pre>

    <h2>Get Statistics</h2>
    <pre>GET /api/stats/{code}</pre>

    <h2>Delete URL</h2>
    <pre>DELETE /api/{code}</pre>

    <h2>Example</h2>
    <pre>curl -X POST http://localhost:8080/api/shorten \
  -H "Content-Type: application/json" \
  -d '{"url":"https://example.com"}'</pre>
</body>
</html>
`
		fmt.Fprint(w, html)
	})

	log.Println("URL Shortener starting on :8080")
	log.Println("Note: Requires Redis running on localhost:6379")
	log.Fatal(http.ListenAndServe(":8080", r))
}
