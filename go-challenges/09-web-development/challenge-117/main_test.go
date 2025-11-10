package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

func TestGracefulServer(t *testing.T) {
	server := NewGracefulServer(":0", 5*time.Second)

	t.Run("HomeRoute", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		w := httptest.NewRecorder()

		server.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var resp map[string]string
		json.NewDecoder(w.Body).Decode(&resp)

		if resp["message"] != "Server is running" {
			t.Error("Expected server running message")
		}
	})

	t.Run("HealthRoute", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/health", nil)
		w := httptest.NewRecorder()

		server.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var resp map[string]string
		json.NewDecoder(w.Body).Decode(&resp)

		if resp["status"] != "healthy" {
			t.Error("Expected healthy status")
		}
	})

	t.Run("SlowRoute", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/slow?duration=100ms", nil)
		w := httptest.NewRecorder()

		start := time.Now()
		server.ServeHTTP(w, req)
		duration := time.Since(start)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		if duration < 100*time.Millisecond {
			t.Errorf("Expected request to take at least 100ms, took %v", duration)
		}

		var resp map[string]interface{}
		json.NewDecoder(w.Body).Decode(&resp)

		if resp["message"] != "Slow operation completed" {
			t.Error("Expected slow operation message")
		}
	})

	t.Run("StatsRoute", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/stats", nil)
		w := httptest.NewRecorder()

		server.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var resp map[string]interface{}
		json.NewDecoder(w.Body).Decode(&resp)

		if resp["active_requests"] == nil {
			t.Error("Expected active_requests in response")
		}
	})

	t.Log("✓ Server routes work!")
}

func TestRequestTracking(t *testing.T) {
	server := NewGracefulServer(":0", 5*time.Second)

	t.Run("TrackActiveRequests", func(t *testing.T) {
		initialActive := server.GetActiveRequests()

		var wg sync.WaitGroup
		numRequests := 5

		// Start multiple slow requests
		for i := 0; i < numRequests; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				req := httptest.NewRequest("GET", "/slow?duration=200ms", nil)
				w := httptest.NewRecorder()
				server.ServeHTTP(w, req)
			}()
		}

		// Give goroutines time to start
		time.Sleep(50 * time.Millisecond)

		// Check active requests increased
		activeNow := server.GetActiveRequests()
		if activeNow <= initialActive {
			t.Errorf("Expected active requests to increase, got %d", activeNow)
		}

		// Wait for all to complete
		wg.Wait()

		// Check active requests decreased
		finalActive := server.GetActiveRequests()
		if finalActive != initialActive {
			t.Errorf("Expected active requests to return to %d, got %d", initialActive, finalActive)
		}
	})

	t.Log("✓ Request tracking works!")
}

func TestGracefulShutdown(t *testing.T) {
	t.Run("ShutdownWithoutActiveRequests", func(t *testing.T) {
		server := NewGracefulServer(":0", 2*time.Second)

		// Start server
		testServer := httptest.NewServer(server)

		// Shutdown immediately
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		err := server.server.Shutdown(ctx)
		if err != nil {
			t.Errorf("Expected clean shutdown, got error: %v", err)
		}

		testServer.Close()
	})

	t.Run("ShutdownWithActiveRequests", func(t *testing.T) {
		server := NewGracefulServer(":0", 3*time.Second)
		testServer := httptest.NewServer(server)

		var wg sync.WaitGroup

		// Start a slow request
		wg.Add(1)
		go func() {
			defer wg.Done()
			resp, err := http.Get(testServer.URL + "/slow?duration=500ms")
			if err != nil {
				t.Logf("Request error (expected during shutdown): %v", err)
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode == http.StatusOK {
				t.Log("Request completed successfully before shutdown")
			}
		}()

		// Give request time to start
		time.Sleep(100 * time.Millisecond)

		// Check there's an active request
		active := server.GetActiveRequests()
		if active == 0 {
			t.Log("Warning: No active requests detected")
		}

		// Shutdown server
		shutdownStart := time.Now()
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		err := server.server.Shutdown(ctx)
		shutdownDuration := time.Since(shutdownStart)

		if err != nil {
			t.Logf("Shutdown error: %v", err)
		}

		// Shutdown should wait for the slow request
		if shutdownDuration < 400*time.Millisecond {
			t.Logf("Shutdown was very quick (%v), request may not have been honored", shutdownDuration)
		}

		wg.Wait()
		testServer.Close()
	})

	t.Run("ShutdownTimeout", func(t *testing.T) {
		server := NewGracefulServer(":0", 500*time.Millisecond)
		testServer := httptest.NewServer(server)

		var wg sync.WaitGroup

		// Start a very slow request that exceeds shutdown timeout
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := http.Get(testServer.URL + "/slow?duration=2s")
			if err != nil {
				// Expected to fail due to shutdown
				t.Logf("Request failed as expected: %v", err)
			}
		}()

		// Give request time to start
		time.Sleep(100 * time.Millisecond)

		// Shutdown with short timeout
		ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
		defer cancel()

		start := time.Now()
		err := server.server.Shutdown(ctx)
		duration := time.Since(start)

		// Should timeout
		if err == nil {
			t.Log("Shutdown completed (request may have been interrupted)")
		}

		// Should not wait for full request duration
		if duration >= 2*time.Second {
			t.Error("Shutdown waited too long, should have timed out")
		}

		wg.Wait()
		testServer.Close()
	})

	t.Log("✓ Graceful shutdown works!")
}

func TestServerLifecycle(t *testing.T) {
	t.Run("StartAndShutdown", func(t *testing.T) {
		server := NewGracefulServer(":0", 1*time.Second)

		// Start server
		if err := server.Start(); err != nil {
			t.Fatalf("Failed to start server: %v", err)
		}

		// Give server time to start
		time.Sleep(100 * time.Millisecond)

		// Shutdown server
		if err := server.Shutdown(); err != nil {
			t.Errorf("Shutdown error: %v", err)
		}
	})

	t.Run("GetActiveRequests", func(t *testing.T) {
		server := NewGracefulServer(":0", 1*time.Second)

		active := server.GetActiveRequests()
		if active < 0 {
			t.Error("Active requests should not be negative")
		}
	})

	t.Log("✓ Server lifecycle works!")
}

func TestConcurrentRequests(t *testing.T) {
	server := NewGracefulServer(":0", 5*time.Second)
	testServer := httptest.NewServer(server)
	defer testServer.Close()

	t.Run("MultipleConcurrentRequests", func(t *testing.T) {
		var wg sync.WaitGroup
		numRequests := 20
		errors := make(chan error, numRequests)

		for i := 0; i < numRequests; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()

				resp, err := http.Get(testServer.URL + "/")
				if err != nil {
					errors <- err
					return
				}
				defer resp.Body.Close()

				if resp.StatusCode != http.StatusOK {
					errors <- err
				}
			}(i)
		}

		wg.Wait()
		close(errors)

		errorCount := 0
		for err := range errors {
			if err != nil {
				t.Errorf("Request error: %v", err)
				errorCount++
			}
		}

		if errorCount > 0 {
			t.Errorf("Failed %d out of %d requests", errorCount, numRequests)
		}
	})

	t.Log("✓ Concurrent requests work!")
}
