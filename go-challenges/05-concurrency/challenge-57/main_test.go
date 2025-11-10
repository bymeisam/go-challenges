package main

import (
	"testing"
	"time"
)

func TestRateLimit(t *testing.T) {
	start := time.Now()
	items := []string{"a", "b", "c"}
	
	ProcessWithRateLimit(items, 10) // 10 requests per second
	
	elapsed := time.Since(start)
	if elapsed < 200*time.Millisecond {
		t.Error("Should rate limit")
	}
	
	t.Log("✓ Rate limiting works!")
}
