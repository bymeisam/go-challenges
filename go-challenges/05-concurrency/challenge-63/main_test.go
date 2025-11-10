package main

import (
	"sync"
	"testing"
	"time"
)

func TestSemaphore(t *testing.T) {
	sem := NewSemaphore(2)
	var wg sync.WaitGroup
	active := 0
	maxActive := 0
	var mu sync.Mutex
	
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			
			sem.Acquire()
			defer sem.Release()
			
			mu.Lock()
			active++
			if active > maxActive {
				maxActive = active
			}
			mu.Unlock()
			
			time.Sleep(10 * time.Millisecond)
			
			mu.Lock()
			active--
			mu.Unlock()
		}()
	}
	
	wg.Wait()
	
	if maxActive > 2 {
		t.Errorf("Max concurrent should be 2, got %d", maxActive)
	}
	
	t.Log("✓ Semaphore works!")
}
