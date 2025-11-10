package main

import (
	"sync"
	"testing"
)

func TestCounter_HasRace(t *testing.T) {
	// This test WILL show race conditions when run with -race flag
	// Run: go test -race -v -run TestCounter_HasRace

	t.Log("This test demonstrates UNSAFE counter with race conditions")
	t.Log("Run with: go test -race -v -run TestCounter_HasRace")

	counter := NewCounter()
	var wg sync.WaitGroup

	// Launch 100 goroutines that increment the counter
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			counter.Increment()
		}()
	}

	wg.Wait()

	// The final value might not be 100 due to race conditions!
	t.Logf("Counter value: %d (expected 100, but might be less due to races)", counter.Value())
}

func TestSafeCounter(t *testing.T) {
	// This test should NOT show race conditions
	counter := NewSafeCounter()
	var wg sync.WaitGroup

	// Launch 1000 goroutines that increment the counter
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			counter.Increment()
		}()
	}

	wg.Wait()

	if counter.Value() != 1000 {
		t.Errorf("SafeCounter value = %d; want 1000", counter.Value())
	}

	t.Log("✓ SafeCounter is race-free!")
}

func TestCache_HasRace(t *testing.T) {
	// This test WILL show race conditions when run with -race flag
	// Run: go test -race -v -run TestCache_HasRace

	t.Log("This test demonstrates UNSAFE cache with race conditions")
	t.Log("Run with: go test -race -v -run TestCache_HasRace")

	cache := NewCache()
	var wg sync.WaitGroup

	// Writers
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			cache.Set("key", id)
		}(i)
	}

	// Readers
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			cache.Get("key")
		}()
	}

	wg.Wait()
	t.Log("Cache operations completed (but with races!)")
}

func TestSafeCache(t *testing.T) {
	// This test should NOT show race conditions
	cache := NewSafeCache()
	var wg sync.WaitGroup

	// Writers
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			cache.Set("key", id)
		}(i)
	}

	// Readers
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = cache.Get("key")
		}()
	}

	wg.Wait()

	// Verify cache works
	cache.Set("test", "value")
	value, ok := cache.Get("test")
	if !ok {
		t.Error("expected to find 'test' key")
	}
	if value != "value" {
		t.Errorf("cache value = %v; want 'value'", value)
	}

	t.Log("✓ SafeCache is race-free!")
}

func TestSafeCache_ConcurrentReadWrites(t *testing.T) {
	cache := NewSafeCache()
	var wg sync.WaitGroup

	// Multiple writers writing different keys
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			key := string(rune('a' + id%26))
			cache.Set(key, id)
		}(i)
	}

	// Multiple readers reading different keys
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			key := string(rune('a' + id%26))
			cache.Get(key)
		}(i)
	}

	wg.Wait()
	t.Log("✓ Concurrent read/write operations completed safely!")
}

func TestURLFetcher(t *testing.T) {
	// Test URL fetcher with mock URLs
	fetcher := &URLFetcher{}

	urls := []string{
		"http://example.com/1",
		"http://example.com/2",
		"http://example.com/3",
	}

	results := fetcher.FetchURLs(urls)

	if len(results) != len(urls) {
		t.Errorf("expected %d results, got %d", len(urls), len(results))
	}

	for _, url := range urls {
		if _, ok := results[url]; !ok {
			t.Errorf("missing result for URL: %s", url)
		}
	}

	t.Log("✓ URLFetcher is race-free!")
}

func TestSharedResource(t *testing.T) {
	resource := NewSharedResource()
	var wg sync.WaitGroup

	// Multiple goroutines adding values
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(val int) {
			defer wg.Done()
			resource.Add(val)
		}(i)
	}

	wg.Wait()

	data := resource.GetAll()
	if len(data) != 100 {
		t.Errorf("expected 100 items, got %d", len(data))
	}

	accessCount := resource.AccessCount()
	if accessCount != 100 {
		t.Errorf("expected 100 accesses, got %d", accessCount)
	}

	t.Log("✓ SharedResource is race-free!")
}

// Benchmark to compare safe vs unsafe (run without -race for performance comparison)
func BenchmarkCounter(b *testing.B) {
	b.Run("Unsafe", func(b *testing.B) {
		counter := NewCounter()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				counter.Increment()
			}
		})
	})

	b.Run("Safe", func(b *testing.B) {
		counter := NewSafeCounter()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				counter.Increment()
			}
		})
	})
}

func BenchmarkCache(b *testing.B) {
	b.Run("Unsafe", func(b *testing.B) {
		cache := NewCache()
		cache.Set("key", "value") // Pre-populate

		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				if i%2 == 0 {
					cache.Set("key", i)
				} else {
					cache.Get("key")
				}
				i++
			}
		})
	})

	b.Run("Safe", func(b *testing.B) {
		cache := NewSafeCache()
		cache.Set("key", "value") // Pre-populate

		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				if i%2 == 0 {
					cache.Set("key", i)
				} else {
					cache.Get("key")
				}
				i++
			}
		})
	})
}

// Test demonstrating proper goroutine synchronization patterns
func TestSynchronizationPatterns(t *testing.T) {
	t.Run("WaitGroup pattern", func(t *testing.T) {
		var wg sync.WaitGroup
		counter := NewSafeCounter()

		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				counter.Increment()
			}()
		}

		wg.Wait()

		if counter.Value() != 10 {
			t.Errorf("counter = %d; want 10", counter.Value())
		}
	})

	t.Run("Channel pattern", func(t *testing.T) {
		ch := make(chan int, 10)

		// Producer
		go func() {
			for i := 0; i < 10; i++ {
				ch <- i
			}
			close(ch)
		}()

		// Consumer
		count := 0
		for range ch {
			count++
		}

		if count != 10 {
			t.Errorf("received %d values; want 10", count)
		}
	})

	t.Log("✓ Synchronization patterns work correctly!")
}
