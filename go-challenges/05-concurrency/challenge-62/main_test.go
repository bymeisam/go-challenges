package main

import (
	"context"
	"testing"
)

func TestContextValues(t *testing.T) {
	ctx := context.Background()
	ctx = WithUserID(ctx, "user123")
	ctx = WithRequestID(ctx, "req456")
	
	userID, ok := GetUserID(ctx)
	if !ok || userID != "user123" {
		t.Error("Failed to get user ID")
	}
	
	requestID, ok := GetRequestID(ctx)
	if !ok || requestID != "req456" {
		t.Error("Failed to get request ID")
	}
	
	t.Log("✓ Context values work!")
}
