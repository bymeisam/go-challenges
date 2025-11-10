package main

import "testing"

func TestMiddleware(t *testing.T) {
	handler := func(input string) string {
		return "Handled: " + input
	}
	
	chained := Chain(handler, AuthMiddleware, LoggingMiddleware)
	
	result := chained("test")
	if result != "Logged: Handled: test" {
		t.Errorf("Middleware chain failed: %s", result)
	}
	
	result = chained("unauthenticated")
	if result != "Logged: Auth failed" {
		t.Error("Auth middleware failed")
	}
	
	t.Log("✓ Middleware pattern works!")
}
