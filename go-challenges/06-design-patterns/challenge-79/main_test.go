package main

import (
	"errors"
	"testing"
)

func TestCircuitBreaker(t *testing.T) {
	cb := NewCircuitBreaker(3)
	
	failingFn := func() error {
		return errors.New("failed")
	}
	
	// First 3 failures
	for i := 0; i < 3; i++ {
		cb.Call(failingFn)
	}
	
	// Circuit should be open now
	err := cb.Call(func() error { return nil })
	if err == nil || err.Error() != "circuit breaker is open" {
		t.Error("Circuit breaker should be open")
	}
	
	cb.Reset()
	err = cb.Call(func() error { return nil })
	if err != nil {
		t.Error("Should work after reset")
	}
	
	t.Log("✓ Circuit breaker works!")
}
