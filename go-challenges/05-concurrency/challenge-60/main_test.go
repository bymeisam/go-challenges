package main

import (
	"context"
	"testing"
	"time"
)

func TestContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	
	go func() {
		time.Sleep(20 * time.Millisecond)
		cancel()
	}()
	
	err := WorkWithContext(ctx)
	if err != context.Canceled {
		t.Error("Should be cancelled")
	}
	
	t.Log("✓ Context cancellation works!")
}

func TestCancellableOperation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately
	
	_, err := CancellableOperation(ctx)
	if err == nil {
		t.Error("Should be cancelled")
	}
	
	t.Log("✓ Cancellable operation works!")
}
