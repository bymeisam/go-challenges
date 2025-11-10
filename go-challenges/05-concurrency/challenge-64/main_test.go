package main

import (
	"sync"
	"testing"
)

func TestOnce(t *testing.T) {
	config := &Config{}
	var wg sync.WaitGroup
	
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			config.Load()
		}()
	}
	
	wg.Wait()
	
	if config.Get() != "loaded" {
		t.Error("Config should be loaded")
	}
	
	t.Log("✓ sync.Once works!")
}

func TestSingleton(t *testing.T) {
	c1 := GetConfig()
	c2 := GetConfig()
	
	if c1 != c2 {
		t.Error("Should return same instance")
	}
	
	t.Log("✓ Singleton pattern works!")
}
