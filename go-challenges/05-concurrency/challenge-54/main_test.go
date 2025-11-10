package main

import "testing"

func TestWorkerPool(t *testing.T) {
	jobs := []Job{
		{ID: 1, Value: 10},
		{ID: 2, Value: 20},
		{ID: 3, Value: 30},
	}
	
	results := WorkerPool(2, jobs)
	
	if len(results) != 3 {
		t.Errorf("Expected 3 results, got %d", len(results))
	}
	
	t.Log("✓ Worker pool works!")
}
