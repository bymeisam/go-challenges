package main

import (
	"context"
	"testing"
	"time"
)

func TestTimeout(t *testing.T) {
	_, err := OperationWithTimeout(10 * time.Millisecond)
	if err == nil {
		t.Error("Should timeout")
	}
	
	result, err := OperationWithTimeout(200 * time.Millisecond)
	if err != nil || result != "completed" {
		t.Error("Should complete")
	}
	
	t.Log("✓ Timeout pattern works!")
}

func TestContextTimeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	
	err := ContextTimeout(ctx)
	if err == nil {
		t.Error("Should timeout")
	}
	
	t.Log("✓ Context timeout works!")
}
