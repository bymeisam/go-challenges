package main

import (
	"testing"
)

func TestRunConcurrently(t *testing.T) {
	RunConcurrently(5)
	t.Log("✓ Goroutines launched!")
}

func TestConcurrentSum(t *testing.T) {
	sum := ConcurrentSum([]int{1, 2, 3, 4, 5})
	if sum != 15 {
		t.Errorf("ConcurrentSum = %d; want 15", sum)
	}
	t.Log("✓ Concurrent sum works!")
}

func TestRace(t *testing.T) {
	counter = 0
	Race()
	// Counter will likely be less than 1000 due to race
	t.Logf("Counter value: %d (race condition!)", counter)
	t.Log("✓ Race condition demonstrated!")
}
