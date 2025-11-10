package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type GracefulServer struct {
	server          *http.Server
	shutdownTimeout time.Duration
	activeRequests  int
	mu              sync.Mutex
	shutdownChan    chan struct{}
	logger          *log.Logger
}

func NewGracefulServer(addr string, shutdownTimeout time.Duration) *GracefulServer {
	mux := http.NewServeMux()

	gs := &GracefulServer{
		server: &http.Server{
			Addr:    addr,
			Handler: mux,
		},
		shutdownTimeout: shutdownTimeout,
		shutdownChan:    make(chan struct{}),
		logger:          log.New(os.Stdout, "[GracefulServer] ", log.LstdFlags),
	}

	// Set up routes
	mux.HandleFunc("/", gs.trackRequest(gs.handleHome()))
	mux.HandleFunc("/slow", gs.trackRequest(gs.handleSlow()))
	mux.HandleFunc("/health", gs.trackRequest(gs.handleHealth()))
	mux.HandleFunc("/stats", gs.trackRequest(gs.handleStats()))

	return gs
}

// Middleware to track active requests
func (gs *GracefulServer) trackRequest(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gs.mu.Lock()
		gs.activeRequests++
		gs.mu.Unlock()

		defer func() {
			gs.mu.Lock()
			gs.activeRequests--
			gs.mu.Unlock()
		}()

		next(w, r)
	}
}

func (gs *GracefulServer) handleHome() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Server is running",
		})
	}
}

func (gs *GracefulServer) handleSlow() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Simulate slow operation
		duration := 2 * time.Second

		// Check for custom duration
		if d := r.URL.Query().Get("duration"); d != "" {
			if parsed, err := time.ParseDuration(d); err == nil {
				duration = parsed
			}
		}

		time.Sleep(duration)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message":  "Slow operation completed",
			"duration": duration.String(),
		})
	}
}

func (gs *GracefulServer) handleHealth() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status": "healthy",
		})
	}
}

func (gs *GracefulServer) handleStats() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gs.mu.Lock()
		active := gs.activeRequests
		gs.mu.Unlock()

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"active_requests": active,
		})
	}
}

func (gs *GracefulServer) Start() error {
	gs.logger.Printf("Starting server on %s", gs.server.Addr)

	// Start server in goroutine
	go func() {
		if err := gs.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			gs.logger.Printf("Server error: %v", err)
		}
	}()

	return nil
}

func (gs *GracefulServer) Shutdown() error {
	gs.logger.Println("Shutdown signal received")

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), gs.shutdownTimeout)
	defer cancel()

	// Log active requests
	gs.mu.Lock()
	active := gs.activeRequests
	gs.mu.Unlock()
	gs.logger.Printf("Active requests: %d", active)

	// Attempt graceful shutdown
	gs.logger.Printf("Waiting for active requests to complete (timeout: %s)", gs.shutdownTimeout)

	if err := gs.server.Shutdown(ctx); err != nil {
		gs.logger.Printf("Shutdown error: %v", err)
		return err
	}

	gs.logger.Println("Server stopped gracefully")
	close(gs.shutdownChan)
	return nil
}

func (gs *GracefulServer) Wait() {
	<-gs.shutdownChan
}

func (gs *GracefulServer) GetActiveRequests() int {
	gs.mu.Lock()
	defer gs.mu.Unlock()
	return gs.activeRequests
}

// Run starts the server and handles shutdown signals
func (gs *GracefulServer) Run() error {
	// Start server
	if err := gs.Start(); err != nil {
		return err
	}

	// Set up signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	// Wait for shutdown signal
	sig := <-sigChan
	gs.logger.Printf("Received signal: %v", sig)

	// Perform graceful shutdown
	return gs.Shutdown()
}

// For testing purposes
func (gs *GracefulServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	gs.server.Handler.ServeHTTP(w, r)
}

func main() {
	server := NewGracefulServer(":8080", 30*time.Second)

	fmt.Println("Server starting on :8080")
	fmt.Println("Press Ctrl+C to stop")

	if err := server.Run(); err != nil {
		log.Fatal(err)
	}
}
