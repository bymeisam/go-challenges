package main

import (
	"fmt"
	"io"
	"net/http"
	"sync"
)

// Counter is an UNSAFE counter that has race conditions
type Counter struct {
	count int
}

func NewCounter() *Counter {
	return &Counter{}
}

func (c *Counter) Increment() {
	c.count++ // Race condition: multiple goroutines can access this
}

func (c *Counter) Value() int {
	return c.count // Race condition: reading while others might be writing
}

// SafeCounter is a thread-safe counter using mutex
type SafeCounter struct {
	mu    sync.Mutex
	count int
}

func NewSafeCounter() *SafeCounter {
	return &SafeCounter{}
}

func (sc *SafeCounter) Increment() {
	sc.mu.Lock()
	sc.count++
	sc.mu.Unlock()
}

func (sc *SafeCounter) Value() int {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	return sc.count
}

// Cache is an UNSAFE cache that has race conditions
type Cache struct {
	data map[string]interface{}
}

func NewCache() *Cache {
	return &Cache{
		data: make(map[string]interface{}),
	}
}

func (c *Cache) Set(key string, value interface{}) {
	c.data[key] = value // Race condition: map access from multiple goroutines
}

func (c *Cache) Get(key string) (interface{}, bool) {
	value, ok := c.data[key] // Race condition: reading map while others might be writing
	return value, ok
}

// SafeCache is a thread-safe cache using RWMutex
type SafeCache struct {
	mu   sync.RWMutex
	data map[string]interface{}
}

func NewSafeCache() *SafeCache {
	return &SafeCache{
		data: make(map[string]interface{}),
	}
}

func (sc *SafeCache) Set(key string, value interface{}) {
	sc.mu.Lock()
	sc.data[key] = value
	sc.mu.Unlock()
}

func (sc *SafeCache) Get(key string) (interface{}, bool) {
	sc.mu.RLock() // RLock allows multiple readers
	defer sc.mu.RUnlock()
	value, ok := sc.data[key]
	return value, ok
}

// URLFetcher fetches URLs concurrently in a safe manner
type URLFetcher struct{}

func (uf *URLFetcher) FetchURLs(urls []string) map[string]string {
	results := make(map[string]string)
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, url := range urls {
		wg.Add(1)
		go func(u string) {
			defer wg.Done()

			// Fetch URL (simplified - just return the URL for testing)
			// In real code: resp, err := http.Get(u)
			content := fetchURL(u)

			// Safely store result
			mu.Lock()
			results[u] = content
			mu.Unlock()
		}(url)
	}

	wg.Wait()
	return results
}

// fetchURL is a helper function to simulate URL fetching
func fetchURL(url string) string {
	// In production, this would be:
	// resp, err := http.Get(url)
	// if err != nil { return "" }
	// defer resp.Body.Close()
	// body, _ := io.ReadAll(resp.Body)
	// return string(body)

	// For testing, we'll just do a simple GET if it's a valid URL
	// or return a mock response
	if resp, err := http.Get(url); err == nil {
		defer resp.Body.Close()
		if body, err := io.ReadAll(resp.Body); err == nil {
			return string(body)
		}
	}
	return "mock content for " + url
}

// AtomicCounter demonstrates using sync/atomic for simple counters
type AtomicCounter struct {
	count int64
}

func NewAtomicCounter() *AtomicCounter {
	return &AtomicCounter{}
}

// Note: In real code, you would use sync/atomic package:
// import "sync/atomic"
// func (ac *AtomicCounter) Increment() {
//     atomic.AddInt64(&ac.count, 1)
// }

// SharedResource demonstrates a resource that needs protection
type SharedResource struct {
	mu       sync.Mutex
	data     []int
	accessed int
}

func NewSharedResource() *SharedResource {
	return &SharedResource{
		data: make([]int, 0),
	}
}

func (sr *SharedResource) Add(value int) {
	sr.mu.Lock()
	defer sr.mu.Unlock()
	sr.data = append(sr.data, value)
	sr.accessed++
}

func (sr *SharedResource) GetAll() []int {
	sr.mu.Lock()
	defer sr.mu.Unlock()
	// Return a copy to avoid races on the slice
	result := make([]int, len(sr.data))
	copy(result, sr.data)
	return result
}

func (sr *SharedResource) AccessCount() int {
	sr.mu.Lock()
	defer sr.mu.Unlock()
	return sr.accessed
}

func main() {
	// Example usage
	fmt.Println("Race Detection Examples")

	// Safe counter
	sc := NewSafeCounter()
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			sc.Increment()
		}()
	}
	wg.Wait()
	fmt.Printf("Safe counter: %d\n", sc.Value())
}
