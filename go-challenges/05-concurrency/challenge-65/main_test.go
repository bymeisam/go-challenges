package main

import "testing"

func TestProducerConsumer(t *testing.T) {
	results := ProducerConsumer(2, 3, 10)
	
	if len(results) != 10 {
		t.Errorf("Expected 10 results, got %d", len(results))
	}
	
	t.Log("✓ Producer-consumer pattern works!")
}
