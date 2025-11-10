package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"
)

type Middleware func(http.Handler) http.Handler

// Logging middleware
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		log.Printf("Started %s %s", r.Method, r.URL.Path)
		
		next.ServeHTTP(w, r)
		
		log.Printf("Completed in %v", time.Since(start))
	})
}

// Authentication middleware
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		
		if token == "" {
			http.Error(w, "Missing authorization token", http.StatusUnauthorized)
			return
		}
		
		// Simple token validation (in real app, validate against DB/JWT)
		if !strings.HasPrefix(token, "Bearer ") {
			http.Error(w, "Invalid token format", http.StatusUnauthorized)
			return
		}
		
		// Add user info to context
		ctx := context.WithValue(r.Context(), "user_id", "123")
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// CORS middleware
func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		
		next.ServeHTTP(w, r)
	})
}

// Chain multiple middleware
func Chain(handler http.Handler, middleware ...Middleware) http.Handler {
	for i := len(middleware) - 1; i >= 0; i-- {
		handler = middleware[i](handler)
	}
	return handler
}

func PublicHandler(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Public endpoint",
	})
}

func ProtectedHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Protected endpoint",
		"user_id": userID,
	})
}

func NewMux() *http.ServeMux {
	mux := http.NewServeMux()
	
	// Public route with logging
	mux.Handle("/public", Chain(
		http.HandlerFunc(PublicHandler),
		LoggingMiddleware,
	))
	
	// Protected route with auth and logging
	mux.Handle("/protected", Chain(
		http.HandlerFunc(ProtectedHandler),
		AuthMiddleware,
		LoggingMiddleware,
	))
	
	// CORS-enabled route
	mux.Handle("/cors", Chain(
		http.HandlerFunc(PublicHandler),
		CORSMiddleware,
		LoggingMiddleware,
	))
	
	return mux
}

func main() {}
