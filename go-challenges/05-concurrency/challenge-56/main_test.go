package main

import "testing"

func TestPipeline(t *testing.T) {
	results := Pipeline([]int{1, 2, 3, 4, 5})
	
	// Only squares > 10: 16, 25
	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}
	
	t.Log("✓ Pipeline works!")
}
