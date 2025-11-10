package main

import "testing"

func TestFanOutFanIn(t *testing.T) {
	input := []int{1, 2, 3, 4, 5, 6}
	
	channels := FanOut(input, 2)
	results := FanIn(channels)
	
	if len(results) != len(input) {
		t.Errorf("Expected %d results, got %d", len(input), len(results))
	}
	
	t.Log("✓ Fan-out fan-in works!")
}
