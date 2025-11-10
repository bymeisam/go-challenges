package main

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v8"
)

func setupTestRedis(t *testing.T) (*URLShortener, *miniredis.Miniredis) {
	// Create mini redis server
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("Failed to create miniredis: %v", err)
	}

	shortener := &URLShortener{
		redis: redis.NewClient(&redis.Options{
			Addr: mr.Addr(),
		}),
		baseURL: "http://localhost:8080",
	}

	return shortener, mr
}

func TestGenerateRandomCode(t *testing.T) {
	code1 := generateRandomCode(6)
	if len(code1) != 6 {
		t.Errorf("Expected code length 6, got %d", len(code1))
	}

	code2 := generateRandomCode(6)
	if code1 == code2 {
		t.Error("Generated codes should be different (very unlikely to be same)")
	}

	// Test different lengths
	code3 := generateRandomCode(10)
	if len(code3) != 10 {
		t.Errorf("Expected code length 10, got %d", len(code3))
	}
}

func TestIsValidURL(t *testing.T) {
	tests := []struct {
		url   string
		valid bool
	}{
		{"https://example.com", true},
		{"http://example.com", true},
		{"https://example.com/path", true},
		{"https://example.com/path?query=1", true},
		{"ftp://example.com", false},
		{"invalid-url", false},
		{"", false},
		{"http://", false},
	}

	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			result := isValidURL(tt.url)
			if result != tt.valid {
				t.Errorf("Expected %v for %s, got %v", tt.valid, tt.url, result)
			}
		})
	}
}

func TestNormalizeURL(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"https://example.com/", "https://example.com"},
		{"https://example.com/path/", "https://example.com/path"},
		{"https://example.com/path", "https://example.com/path"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := normalizeURL(tt.input)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestShortenURL(t *testing.T) {
	shortener, mr := setupTestRedis(t)
	defer mr.Close()

	ctx := context.Background()

	req := ShortenRequest{
		URL: "https://example.com/very/long/url",
	}

	response, err := shortener.Shorten(ctx, req)
	if err != nil {
		t.Fatalf("Shorten failed: %v", err)
	}

	if response.Code == "" {
		t.Error("Code should not be empty")
	}

	if response.LongURL != "https://example.com/very/long/url" {
		t.Errorf("Expected long URL to match, got %s", response.LongURL)
	}

	if response.ShortURL == "" {
		t.Error("Short URL should not be empty")
	}
}

func TestShortenWithCustomCode(t *testing.T) {
	shortener, mr := setupTestRedis(t)
	defer mr.Close()

	ctx := context.Background()

	req := ShortenRequest{
		URL:        "https://example.com",
		CustomCode: "mycode",
	}

	response, err := shortener.Shorten(ctx, req)
	if err != nil {
		t.Fatalf("Shorten failed: %v", err)
	}

	if response.Code != "mycode" {
		t.Errorf("Expected code 'mycode', got %s", response.Code)
	}
}

func TestShortenWithDuplicateCustomCode(t *testing.T) {
	shortener, mr := setupTestRedis(t)
	defer mr.Close()

	ctx := context.Background()

	req1 := ShortenRequest{
		URL:        "https://example.com",
		CustomCode: "duplicate",
	}

	_, err := shortener.Shorten(ctx, req1)
	if err != nil {
		t.Fatalf("First shorten failed: %v", err)
	}

	req2 := ShortenRequest{
		URL:        "https://other.com",
		CustomCode: "duplicate",
	}

	_, err = shortener.Shorten(ctx, req2)
	if err == nil {
		t.Error("Expected error for duplicate custom code")
	}
}

func TestShortenWithExpiration(t *testing.T) {
	shortener, mr := setupTestRedis(t)
	defer mr.Close()

	ctx := context.Background()

	req := ShortenRequest{
		URL:       "https://example.com",
		ExpiresIn: 2, // 2 seconds
	}

	response, err := shortener.Shorten(ctx, req)
	if err != nil {
		t.Fatalf("Shorten failed: %v", err)
	}

	if response.ExpiresAt.IsZero() {
		t.Error("ExpiresAt should be set")
	}

	// Fast-forward time in miniredis
	mr.FastForward(3 * time.Second)

	// Try to resolve - should fail
	_, err = shortener.Resolve(ctx, response.Code)
	if err == nil {
		t.Error("Expected error for expired URL")
	}
}

func TestShortenInvalidURL(t *testing.T) {
	shortener, mr := setupTestRedis(t)
	defer mr.Close()

	ctx := context.Background()

	req := ShortenRequest{
		URL: "not-a-valid-url",
	}

	_, err := shortener.Shorten(ctx, req)
	if err == nil {
		t.Error("Expected error for invalid URL")
	}
}

func TestResolveURL(t *testing.T) {
	shortener, mr := setupTestRedis(t)
	defer mr.Close()

	ctx := context.Background()

	// Create short URL
	req := ShortenRequest{
		URL: "https://example.com",
	}

	response, err := shortener.Shorten(ctx, req)
	if err != nil {
		t.Fatalf("Shorten failed: %v", err)
	}

	// Resolve it
	longURL, err := shortener.Resolve(ctx, response.Code)
	if err != nil {
		t.Fatalf("Resolve failed: %v", err)
	}

	if longURL != "https://example.com" {
		t.Errorf("Expected https://example.com, got %s", longURL)
	}
}

func TestResolveNonExistent(t *testing.T) {
	shortener, mr := setupTestRedis(t)
	defer mr.Close()

	ctx := context.Background()

	_, err := shortener.Resolve(ctx, "nonexistent")
	if err == nil {
		t.Error("Expected error for non-existent code")
	}
}

func TestClickCounting(t *testing.T) {
	shortener, mr := setupTestRedis(t)
	defer mr.Close()

	ctx := context.Background()

	// Create short URL
	req := ShortenRequest{
		URL: "https://example.com",
	}

	response, err := shortener.Shorten(ctx, req)
	if err != nil {
		t.Fatalf("Shorten failed: %v", err)
	}

	// Resolve multiple times
	for i := 0; i < 5; i++ {
		_, err := shortener.Resolve(ctx, response.Code)
		if err != nil {
			t.Fatalf("Resolve failed: %v", err)
		}
	}

	// Get stats
	stats, err := shortener.GetStats(ctx, response.Code)
	if err != nil {
		t.Fatalf("GetStats failed: %v", err)
	}

	if stats.Clicks != 5 {
		t.Errorf("Expected 5 clicks, got %d", stats.Clicks)
	}
}

func TestGetStats(t *testing.T) {
	shortener, mr := setupTestRedis(t)
	defer mr.Close()

	ctx := context.Background()

	// Create short URL
	req := ShortenRequest{
		URL: "https://example.com",
	}

	response, err := shortener.Shorten(ctx, req)
	if err != nil {
		t.Fatalf("Shorten failed: %v", err)
	}

	// Get stats
	stats, err := shortener.GetStats(ctx, response.Code)
	if err != nil {
		t.Fatalf("GetStats failed: %v", err)
	}

	if stats.Code != response.Code {
		t.Errorf("Expected code %s, got %s", response.Code, stats.Code)
	}

	if stats.LongURL != "https://example.com" {
		t.Errorf("Expected long URL https://example.com, got %s", stats.LongURL)
	}

	if stats.Clicks != 0 {
		t.Errorf("Expected 0 clicks, got %d", stats.Clicks)
	}

	if stats.CreatedAt.IsZero() {
		t.Error("CreatedAt should be set")
	}
}

func TestGetStatsNonExistent(t *testing.T) {
	shortener, mr := setupTestRedis(t)
	defer mr.Close()

	ctx := context.Background()

	_, err := shortener.GetStats(ctx, "nonexistent")
	if err == nil {
		t.Error("Expected error for non-existent code")
	}
}

func TestDelete(t *testing.T) {
	shortener, mr := setupTestRedis(t)
	defer mr.Close()

	ctx := context.Background()

	// Create short URL
	req := ShortenRequest{
		URL: "https://example.com",
	}

	response, err := shortener.Shorten(ctx, req)
	if err != nil {
		t.Fatalf("Shorten failed: %v", err)
	}

	// Delete it
	err = shortener.Delete(ctx, response.Code)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Try to resolve - should fail
	_, err = shortener.Resolve(ctx, response.Code)
	if err == nil {
		t.Error("Expected error after deletion")
	}
}

func TestDeleteNonExistent(t *testing.T) {
	shortener, mr := setupTestRedis(t)
	defer mr.Close()

	ctx := context.Background()

	err := shortener.Delete(ctx, "nonexistent")
	if err == nil {
		t.Error("Expected error for non-existent code")
	}
}

func TestGenerateCode(t *testing.T) {
	shortener, mr := setupTestRedis(t)
	defer mr.Close()

	ctx := context.Background()

	code, err := shortener.generateCode(ctx)
	if err != nil {
		t.Fatalf("generateCode failed: %v", err)
	}

	if len(code) != defaultCodeLength {
		t.Errorf("Expected code length %d, got %d", defaultCodeLength, len(code))
	}

	// Generate another - should be different
	code2, err := shortener.generateCode(ctx)
	if err != nil {
		t.Fatalf("generateCode failed: %v", err)
	}

	if code == code2 {
		t.Error("Generated codes should be different")
	}
}

func TestURLDataSerialization(t *testing.T) {
	data := URLData{
		LongURL:   "https://example.com",
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(24 * time.Hour),
		Clicks:    5,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var decoded URLData
	err = json.Unmarshal(jsonData, &decoded)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if decoded.LongURL != data.LongURL {
		t.Error("LongURL mismatch")
	}

	if decoded.Clicks != data.Clicks {
		t.Error("Clicks mismatch")
	}
}

func TestConcurrentShorten(t *testing.T) {
	shortener, mr := setupTestRedis(t)
	defer mr.Close()

	ctx := context.Background()

	done := make(chan bool)
	errors := make(chan error, 10)

	for i := 0; i < 10; i++ {
		go func(id int) {
			req := ShortenRequest{
				URL: "https://example.com/" + string(rune('0'+id)),
			}

			_, err := shortener.Shorten(ctx, req)
			if err != nil {
				errors <- err
			}

			done <- true
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	close(errors)
	for err := range errors {
		t.Errorf("Concurrent shorten error: %v", err)
	}
}

func TestURLNormalization(t *testing.T) {
	shortener, mr := setupTestRedis(t)
	defer mr.Close()

	ctx := context.Background()

	// URLs that should normalize to the same thing
	urls := []string{
		"https://example.com/path/",
		"https://example.com/path",
	}

	responses := make([]*ShortenResponse, len(urls))
	for i, url := range urls {
		req := ShortenRequest{URL: url}
		resp, err := shortener.Shorten(ctx, req)
		if err != nil {
			t.Fatalf("Shorten failed: %v", err)
		}
		responses[i] = resp
	}

	// Both should normalize to the same URL
	if responses[0].LongURL != responses[1].LongURL {
		t.Error("URLs should normalize to the same value")
	}
}

func TestExpirationPreservation(t *testing.T) {
	shortener, mr := setupTestRedis(t)
	defer mr.Close()

	ctx := context.Background()

	req := ShortenRequest{
		URL:       "https://example.com",
		ExpiresIn: 10,
	}

	response, err := shortener.Shorten(ctx, req)
	if err != nil {
		t.Fatalf("Shorten failed: %v", err)
	}

	// Resolve to increment counter
	_, err = shortener.Resolve(ctx, response.Code)
	if err != nil {
		t.Fatalf("Resolve failed: %v", err)
	}

	// Check that expiration is still set in Redis
	ttl, err := shortener.redis.TTL(ctx, "url:"+response.Code).Result()
	if err != nil {
		t.Fatalf("TTL check failed: %v", err)
	}

	if ttl <= 0 {
		t.Error("TTL should still be set after resolve")
	}
}

func TestStatsTimestamps(t *testing.T) {
	shortener, mr := setupTestRedis(t)
	defer mr.Close()

	ctx := context.Background()

	before := time.Now()

	req := ShortenRequest{
		URL: "https://example.com",
	}

	response, err := shortener.Shorten(ctx, req)
	if err != nil {
		t.Fatalf("Shorten failed: %v", err)
	}

	after := time.Now()

	stats, err := shortener.GetStats(ctx, response.Code)
	if err != nil {
		t.Fatalf("GetStats failed: %v", err)
	}

	if stats.CreatedAt.Before(before) || stats.CreatedAt.After(after) {
		t.Error("CreatedAt timestamp is out of expected range")
	}
}

func TestMultipleResolves(t *testing.T) {
	shortener, mr := setupTestRedis(t)
	defer mr.Close()

	ctx := context.Background()

	req := ShortenRequest{
		URL: "https://example.com",
	}

	response, err := shortener.Shorten(ctx, req)
	if err != nil {
		t.Fatalf("Shorten failed: %v", err)
	}

	// Resolve multiple times
	for i := 0; i < 3; i++ {
		longURL, err := shortener.Resolve(ctx, response.Code)
		if err != nil {
			t.Fatalf("Resolve %d failed: %v", i, err)
		}

		if longURL != "https://example.com" {
			t.Errorf("Resolve %d: expected https://example.com, got %s", i, longURL)
		}
	}

	// Check click count
	stats, _ := shortener.GetStats(ctx, response.Code)
	if stats.Clicks != 3 {
		t.Errorf("Expected 3 clicks, got %d", stats.Clicks)
	}
}
