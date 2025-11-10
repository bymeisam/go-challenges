package main

import (
	"testing"
	"time"
)

func TestObjectPool(t *testing.T) {
	pool := NewObjectPool(func() interface{} {
		return make([]byte, 1024)
	})

	obj1 := pool.Get()
	if obj1 == nil {
		t.Fatal("Expected object from pool")
	}

	pool.Put(obj1)

	obj2 := pool.Get()
	if obj2 == nil {
		t.Fatal("Expected reused object from pool")
	}
}

func TestBufferPool(t *testing.T) {
	bp := NewBufferPool(1024)

	buf1 := bp.Get()
	buf1.WriteString("test")

	if buf1.String() != "test" {
		t.Fatal("Expected buffer content")
	}

	bp.Put(buf1)

	buf2 := bp.Get()
	if buf2.String() != "" {
		t.Fatal("Expected reset buffer")
	}
}

func TestLockFreeQueue(t *testing.T) {
	queue := NewLockFreeQueue()

	queue.Enqueue(1)
	queue.Enqueue(2)
	queue.Enqueue(3)

	if queue.Size() != 3 {
		t.Fatalf("Expected size 3, got %d", queue.Size())
	}

	val1, ok := queue.Dequeue()
	if !ok || val1 != 1 {
		t.Fatal("Expected first value")
	}

	if queue.Size() != 2 {
		t.Fatalf("Expected size 2 after dequeue, got %d", queue.Size())
	}
}

func TestLockFreeQueueEmpty(t *testing.T) {
	queue := NewLockFreeQueue()

	_, ok := queue.Dequeue()
	if ok {
		t.Fatal("Expected false for empty queue")
	}
}

func TestZeroAllocBuffer(t *testing.T) {
	zab := &ZeroAllocBuffer{}

	data := []byte("test")
	err := zab.Write(data)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if string(zab.Bytes()) != "test" {
		t.Fatal("Expected buffer content")
	}

	zab.Reset()
	if len(zab.Bytes()) != 0 {
		t.Fatal("Expected empty buffer after reset")
	}
}

func TestLRUCache(t *testing.T) {
	cache := NewLRUCache(2)

	cache.Set("key1", "value1", 1*time.Hour)
	cache.Set("key2", "value2", 1*time.Hour)

	val, ok := cache.Get("key1")
	if !ok || val != "value1" {
		t.Fatal("Expected cached value")
	}

	// Add third item, should evict LRU (key2)
	cache.Set("key3", "value3", 1*time.Hour)

	_, ok = cache.Get("key2")
	if ok {
		t.Fatal("Expected key2 to be evicted")
	}
}

func TestLRUCacheExpiration(t *testing.T) {
	cache := NewLRUCache(10)

	cache.Set("key1", "value1", 100*time.Millisecond)

	val, ok := cache.Get("key1")
	if !ok || val != "value1" {
		t.Fatal("Expected cached value")
	}

	time.Sleep(150 * time.Millisecond)

	_, ok = cache.Get("key1")
	if ok {
		t.Fatal("Expected expired value to be removed")
	}
}

func TestLRUCacheStats(t *testing.T) {
	cache := NewLRUCache(10)

	cache.Set("key1", "value1", 1*time.Hour)

	cache.Get("key1")
	cache.Get("key1")
	cache.Get("missing")

	stats := cache.GetStats()

	if stats["hits"] != int64(2) {
		t.Fatalf("Expected 2 hits, got %v", stats["hits"])
	}

	if stats["misses"] != int64(1) {
		t.Fatalf("Expected 1 miss, got %v", stats["misses"])
	}
}

func TestGoroutinePool(t *testing.T) {
	pool := NewGoroutinePool(4)

	counter := &AtomicCounter{}

	for i := 0; i < 10; i++ {
		pool.Submit(func() {
			counter.Increment()
		})
	}

	time.Sleep(100 * time.Millisecond)
	pool.Shutdown()

	if counter.Get() != 10 {
		t.Fatalf("Expected 10 completed tasks, got %d", counter.Get())
	}

	if pool.CompletedTasks() != 10 {
		t.Fatalf("Expected 10 completed tasks, got %d", pool.CompletedTasks())
	}
}

func TestGoroutinePoolActive(t *testing.T) {
	pool := NewGoroutinePool(2)

	for i := 0; i < 4; i++ {
		pool.Submit(func() {
			time.Sleep(100 * time.Millisecond)
		})
	}

	time.Sleep(50 * time.Millisecond)

	active := pool.ActiveWorkers()
	if active == 0 {
		t.Fatal("Expected some active workers")
	}

	time.Sleep(200 * time.Millisecond)
	pool.Shutdown()
}

func TestMetricsCollector(t *testing.T) {
	mc := NewMetricsCollector()

	mc.RecordAllocation()
	mc.RecordAllocation()
	mc.RecordDeallocation()

	metrics := mc.GetMetrics()

	if metrics.Allocations != 2 {
		t.Fatalf("Expected 2 allocations, got %d", metrics.Allocations)
	}

	if metrics.Deallocations != 1 {
		t.Fatalf("Expected 1 deallocation, got %d", metrics.Deallocations)
	}
}

func TestMetricsGCPause(t *testing.T) {
	mc := NewMetricsCollector()

	mc.RecordGCPause(10 * time.Millisecond)
	mc.RecordGCPause(20 * time.Millisecond)

	metrics := mc.GetMetrics()

	if metrics.GCPauses != 2 {
		t.Fatalf("Expected 2 GC pauses, got %d", metrics.GCPauses)
	}

	if metrics.AvgPauseTime == 0 {
		t.Fatal("Expected non-zero average pause time")
	}
}

func TestAtomicCounter(t *testing.T) {
	counter := &AtomicCounter{}

	counter.Increment()
	counter.Increment()

	if counter.Get() != 2 {
		t.Fatalf("Expected 2, got %d", counter.Get())
	}

	counter.Set(10)
	if counter.Get() != 10 {
		t.Fatalf("Expected 10, got %d", counter.Get())
	}
}

func TestOptimizedStringBuilder(t *testing.T) {
	builder := NewOptimizedStringBuilder()

	buf := builder.Build()
	buf.WriteString("hello ")
	buf.WriteString("world")

	if buf.String() != "hello world" {
		t.Fatal("Expected concatenated string")
	}

	builder.Release(buf)

	buf2 := builder.Build()
	if buf2.String() != "" {
		t.Fatal("Expected empty buffer after release")
	}
}

func TestConcurrentLRUCache(t *testing.T) {
	cache := NewLRUCache(100)

	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(id int) {
			for j := 0; j < 10; j++ {
				cache.Set("key-"+string(rune(id*10+j)), "value", 1*time.Hour)
			}
			for j := 0; j < 10; j++ {
				cache.Get("key-" + string(rune(id*10+j)))
			}
			done <- true
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	stats := cache.GetStats()
	if stats["size"] == 0 {
		t.Fatal("Expected cached values")
	}
}

func TestConcurrentGoroutinePool(t *testing.T) {
	pool := NewGoroutinePool(10)

	counter := &AtomicCounter{}

	for i := 0; i < 100; i++ {
		pool.Submit(func() {
			counter.Increment()
		})
	}

	time.Sleep(500 * time.Millisecond)
	pool.Shutdown()

	if counter.Get() != 100 {
		t.Fatalf("Expected 100 completed, got %d", counter.Get())
	}
}

// ========== Benchmarks ==========

func BenchmarkObjectPool(b *testing.B) {
	pool := NewObjectPool(func() interface{} {
		return make([]byte, 1024)
	})

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		obj := pool.Get()
		pool.Put(obj)
	}
}

func BenchmarkBufferPool(b *testing.B) {
	bp := NewBufferPool(1024)

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		buf := bp.Get()
		buf.WriteString("test")
		bp.Put(buf)
	}
}

func BenchmarkLockFreeQueue(b *testing.B) {
	queue := NewLockFreeQueue()

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		queue.Enqueue(i)
		queue.Dequeue()
	}
}

func BenchmarkLRUCache(b *testing.B) {
	cache := NewLRUCache(100)

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		cache.Set("key-"+string(rune(i%100)), "value", 1*time.Hour)
		cache.Get("key-" + string(rune(i%100)))
	}
}

func BenchmarkAtomicCounter(b *testing.B) {
	counter := &AtomicCounter{}

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		counter.Increment()
	}
}

func BenchmarkZeroAllocBuffer(b *testing.B) {
	zab := &ZeroAllocBuffer{}

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		zab.Write([]byte("test"))
		zab.Reset()
	}
}
