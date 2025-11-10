package main

import (
	"sync"
	"testing"
)

func TestSafeCounter(t *testing.T) {
	counter := &SafeCounter{}
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
	
	t.Log("✓ Mutex works!")
}

func TestCache(t *testing.T) {
	cache := NewCache()
	cache.Set("key", "value")
	
	val, ok := cache.Get("key")
	if !ok || val != "value" {
		t.Error("Cache failed")
	}
	
	t.Log("✓ RWMutex works!")
}
