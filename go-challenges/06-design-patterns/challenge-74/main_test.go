package main

import "testing"

func TestChainOfResponsibility(t *testing.T) {
	auth := &AuthHandler{}
	validation := &ValidationHandler{}
	
	auth.SetNext(validation)
	
	result := auth.Handle("valid")
	if result != "Validation passed" {
		t.Errorf("Chain failed: %s", result)
	}
	
	result = auth.Handle("unauthenticated")
	if result != "Auth failed" {
		t.Error("Auth should fail")
	}
	
	t.Log("✓ Chain of Responsibility works!")
}
