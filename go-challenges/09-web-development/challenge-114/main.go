package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

type TokenBucket struct {
	tokens       int
	capacity     int
	refillRate   int           // tokens per second
	lastRefill   time.Time
	mu           sync.Mutex
}

func NewTokenBucket(capacity, refillRate int) *TokenBucket {
	return &TokenBucket{
		tokens:     capacity,
		capacity:   capacity,
		refillRate: refillRate,
		lastRefill: time.Now(),
	}
}

func (tb *TokenBucket) Allow() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	// Refill tokens based on time passed
	now := time.Now()
	elapsed := now.Sub(tb.lastRefill).Seconds()
	tokensToAdd := int(elapsed * float64(tb.refillRate))

	if tokensToAdd > 0 {
		tb.tokens += tokensToAdd
		if tb.tokens > tb.capacity {
			tb.tokens = tb.capacity
		}
		tb.lastRefill = now
	}

	// Check if we have tokens available
	if tb.tokens > 0 {
		tb.tokens--
		return true
	}

	return false
}

func (tb *TokenBucket) Remaining() int {
	tb.mu.Lock()
	defer tb.mu.Unlock()
	return tb.tokens
}

type RateLimiter struct {
	buckets  map[string]*TokenBucket
	mu       sync.RWMutex
	capacity int
	refillRate int
}

func NewRateLimiter(capacity, refillRate int) *RateLimiter {
	rl := &RateLimiter{
		buckets:    make(map[string]*TokenBucket),
		capacity:   capacity,
		refillRate: refillRate,
	}
	// Cleanup old buckets periodically
	go rl.cleanup()
	return rl
}

func (rl *RateLimiter) getBucket(key string) *TokenBucket {
	rl.mu.RLock()
	bucket, exists := rl.buckets[key]
	rl.mu.RUnlock()

	if !exists {
		rl.mu.Lock()
		// Double-check after acquiring write lock
		bucket, exists = rl.buckets[key]
		if !exists {
			bucket = NewTokenBucket(rl.capacity, rl.refillRate)
			rl.buckets[key] = bucket
		}
		rl.mu.Unlock()
	}

	return bucket
}

func (rl *RateLimiter) Allow(key string) bool {
	bucket := rl.getBucket(key)
	return bucket.Allow()
}

func (rl *RateLimiter) Remaining(key string) int {
	bucket := rl.getBucket(key)
	return bucket.Remaining()
}

func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for key, bucket := range rl.buckets {
			// Remove buckets that haven't been used in 10 minutes
			if now.Sub(bucket.lastRefill) > 10*time.Minute {
				delete(rl.buckets, key)
			}
		}
		rl.mu.Unlock()
	}
}

type RateLimitedServer struct {
	mux     *http.ServeMux
	limiter *RateLimiter
}

func NewRateLimitedServer(requestsPerSecond int) *RateLimitedServer {
	server := &RateLimitedServer{
		mux:     http.NewServeMux(),
		limiter: NewRateLimiter(requestsPerSecond*10, requestsPerSecond), // 10 second capacity
	}
	server.routes()
	return server
}

func (s *RateLimitedServer) routes() {
	s.mux.HandleFunc("/api/data", s.rateLimitMiddleware(s.handleData()))
	s.mux.HandleFunc("/api/unlimited", s.handleUnlimited())
}

func (s *RateLimitedServer) rateLimitMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Use IP address as the key for rate limiting
		ip := getIP(r)

		// Check rate limit
		if !s.limiter.Allow(ip) {
			remaining := s.limiter.Remaining(ip)
			w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", s.limiter.capacity))
			w.Header().Set("X-RateLimit-Remaining", "0")
			w.Header().Set("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(time.Second).Unix()))

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error":   "Rate limit exceeded",
				"message": "Too many requests. Please try again later.",
			})
			return
		}

		// Add rate limit headers
		remaining := s.limiter.Remaining(ip)
		w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", s.limiter.capacity))
		w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))

		// Call next handler
		next(w, r)
	}
}

func (s *RateLimitedServer) handleData() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"message": "This endpoint is rate limited",
			"data":    "Some important data",
		})
	}
}

func (s *RateLimitedServer) handleUnlimited() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"message": "This endpoint has no rate limit",
		})
	}
}

func getIP(r *http.Request) string {
	// Check X-Forwarded-For header first
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		return forwarded
	}

	// Fall back to RemoteAddr
	return r.RemoteAddr
}

func (s *RateLimitedServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func main() {}
