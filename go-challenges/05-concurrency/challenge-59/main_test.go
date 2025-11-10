package main

import (
	"sync"
	"testing"
)

func TestAtomic(t *testing.T) {
	counter := &AtomicCounter{}
	var wg sync.WaitGroup
	
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			counter.Inc()
		}()
	}
	
	wg.Wait()
	
	if counter.Value() != 1000 {
		t.Errorf("Counter = %d; want 1000", counter.Value())
	}
	
	t.Log("✓ Atomic operations work!")
}

func TestCompareAndSwap(t *testing.T) {
	counter := &AtomicCounter{}
	
	if !counter.CompareAndSwap(0, 10) {
		t.Error("CAS should succeed")
	}
	
	if counter.Value() != 10 {
		t.Error("Value should be 10")
	}
	
	t.Log("✓ Compare-and-swap works!")
}
