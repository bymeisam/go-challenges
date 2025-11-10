package main

import (
	"sync"
	"testing"
)

func TestSingleton(t *testing.T) {
	var wg sync.WaitGroup
	instances := make([]*Database, 100)
	
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			instances[index] = GetInstance()
		}(i)
	}
	
	wg.Wait()
	
	// All instances should be the same
	first := instances[0]
	for i, inst := range instances {
		if inst != first {
			t.Errorf("Instance %d is different", i)
		}
	}
	
	t.Log("✓ Singleton pattern works!")
}
