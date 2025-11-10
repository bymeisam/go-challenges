package main

import (
	"context"
	"testing"
	"time"
)

func TestContextTimeout(t *testing.T) {
	err := RunWithTimeout(50 * time.Millisecond)
	if err != context.DeadlineExceeded {
		t.Error("Should timeout")
	}
	
	err = RunWithTimeout(200 * time.Millisecond)
	if err != nil {
		t.Error("Should complete successfully")
	}
	
	t.Log("✓ Context timeout works!")
}
