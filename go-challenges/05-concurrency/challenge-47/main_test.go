package main

import "testing"

func TestWaitGroup(t *testing.T) {
	items := []string{"a", "b", "c"}
	results := ProcessWithWaitGroup(items)
	
	if len(results) != 3 {
		t.Errorf("Expected 3 results, got %d", len(results))
	}
	
	for _, result := range results {
		if result == "" {
			t.Error("Some goroutines didn't complete")
		}
	}
	
	t.Log("✓ WaitGroup works!")
}
