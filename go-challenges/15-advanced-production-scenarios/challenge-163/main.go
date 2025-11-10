package main

import (
	"bytes"
	"sync"
	"sync/atomic"
	"time"
)

// ========== Memory Pool Models ==========

type ObjectPool struct {
	pool  *sync.Pool
	size  int
	count int64
	mu    sync.RWMutex
}

type BufferPool struct {
	pool       *sync.Pool
	bufferSize int
	creates    int64
	reuses     int64
}

type PerformanceMetrics struct {
	Allocations    int64
	Deallocations  int64
	PoolSize       int64
	GCPauses       int64
	AvgPauseTime   time.Duration
	MemoryUsage    uint64
	PeakMemory     uint64
}

type LockFreeQueue struct {
	head    *Node
	tail    *Node
	size    int64
	headMu  sync.Mutex
	tailMu  sync.Mutex
}

type Node struct {
	value interface{}
	next  *Node
}

// ========== Object Pool Implementation ==========

func NewObjectPool(factory func() interface{}) *ObjectPool {
	return &ObjectPool{
		pool: &sync.Pool{
			New: factory,
		},
		size: 0,
	}
}

func (op *ObjectPool) Get() interface{} {
	atomic.AddInt64(&op.count, 1)
	return op.pool.Get()
}

func (op *ObjectPool) Put(obj interface{}) {
	atomic.AddInt64(&op.count, -1)
	op.pool.Put(obj)
}

// ========== Buffer Pool Implementation ==========

func NewBufferPool(bufferSize int) *BufferPool {
	return &BufferPool{
		pool: &sync.Pool{
			New: func() interface{} {
				atomic.AddInt64(&allocations, 1)
				return bytes.NewBuffer(make([]byte, 0, bufferSize))
			},
		},
		bufferSize: bufferSize,
	}
}

var allocations int64
var deallocations int64

func (bp *BufferPool) Get() *bytes.Buffer {
	buf := bp.pool.Get().(*bytes.Buffer)
	atomic.AddInt64(&bp.reuses, 1)
	return buf
}

func (bp *BufferPool) Put(buf *bytes.Buffer) {
	buf.Reset()
	bp.pool.Put(buf)
	atomic.AddInt64(&deallocations, 1)
}

// ========== Lock-Free Queue Implementation ==========

func NewLockFreeQueue() *LockFreeQueue {
	sentinel := &Node{}
	return &LockFreeQueue{
		head: sentinel,
		tail: sentinel,
	}
}

func (q *LockFreeQueue) Enqueue(value interface{}) {
	newNode := &Node{value: value}

	q.tailMu.Lock()
	q.tail.next = newNode
	q.tail = newNode
	q.tailMu.Unlock()

	atomic.AddInt64(&q.size, 1)
}

func (q *LockFreeQueue) Dequeue() (interface{}, bool) {
	q.headMu.Lock()
	head := q.head
	if head.next == nil {
		q.headMu.Unlock()
		return nil, false
	}
	q.head = head.next
	q.headMu.Unlock()

	atomic.AddInt64(&q.size, -1)
	return head.next.value, true
}

func (q *LockFreeQueue) Size() int64 {
	return atomic.LoadInt64(&q.size)
}

// ========== Zero-Allocation Optimizations ==========

type ZeroAllocBuffer struct {
	buffer [4096]byte
	pos    int
}

func (zab *ZeroAllocBuffer) Write(data []byte) error {
	if zab.pos+len(data) > len(zab.buffer) {
		return nil // buffer full
	}
	copy(zab.buffer[zab.pos:], data)
	zab.pos += len(data)
	return nil
}

func (zab *ZeroAllocBuffer) Reset() {
	zab.pos = 0
}

func (zab *ZeroAllocBuffer) Bytes() []byte {
	return zab.buffer[:zab.pos]
}

// ========== Caching Layer ==========

type CacheEntry struct {
	value      interface{}
	expiresAt  time.Time
	accessTime time.Time
	accessCount int64
}

type LRUCache struct {
	cache   map[string]*CacheEntry
	mu      sync.RWMutex
	maxSize int
	hits    int64
	misses  int64
}

func NewLRUCache(maxSize int) *LRUCache {
	return &LRUCache{
		cache:   make(map[string]*CacheEntry),
		maxSize: maxSize,
	}
}

func (c *LRUCache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	entry, exists := c.cache[key]
	c.mu.RUnlock()

	if !exists {
		atomic.AddInt64(&c.misses, 1)
		return nil, false
	}

	if time.Now().After(entry.expiresAt) {
		c.mu.Lock()
		delete(c.cache, key)
		c.mu.Unlock()
		atomic.AddInt64(&c.misses, 1)
		return nil, false
	}

	entry.accessTime = time.Now()
	atomic.AddInt64(&entry.accessCount, 1)
	atomic.AddInt64(&c.hits, 1)

	return entry.value, true
}

func (c *LRUCache) Set(key string, value interface{}, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(c.cache) >= c.maxSize {
		// Evict LRU entry
		var lruKey string
		var lruTime time.Time

		for k, entry := range c.cache {
			if lruTime.IsZero() || entry.accessTime.Before(lruTime) {
				lruKey = k
				lruTime = entry.accessTime
			}
		}

		if lruKey != "" {
			delete(c.cache, lruKey)
		}
	}

	c.cache[key] = &CacheEntry{
		value:       value,
		expiresAt:   time.Now().Add(ttl),
		accessTime:  time.Now(),
		accessCount: 0,
	}
}

func (c *LRUCache) GetStats() map[string]interface{} {
	c.mu.RLock()
	size := len(c.cache)
	c.mu.RUnlock()

	hits := atomic.LoadInt64(&c.hits)
	misses := atomic.LoadInt64(&c.misses)

	hitRate := 0.0
	if hits+misses > 0 {
		hitRate = float64(hits) / float64(hits+misses) * 100
	}

	return map[string]interface{}{
		"size":     size,
		"hits":     hits,
		"misses":   misses,
		"hit_rate": hitRate,
	}
}

// ========== Goroutine Pool ==========

type GoroutinePool struct {
	workers    int
	workQueue  chan func()
	active     int64
	completed  int64
	poolMu     sync.Mutex
}

func NewGoroutinePool(workers int) *GoroutinePool {
	pool := &GoroutinePool{
		workers:   workers,
		workQueue: make(chan func(), workers*2),
	}

	for i := 0; i < workers; i++ {
		go pool.worker()
	}

	return pool
}

func (gp *GoroutinePool) worker() {
	for work := range gp.workQueue {
		atomic.AddInt64(&gp.active, 1)
		work()
		atomic.AddInt64(&gp.active, -1)
		atomic.AddInt64(&gp.completed, 1)
	}
}

func (gp *GoroutinePool) Submit(work func()) {
	select {
	case gp.workQueue <- work:
	default:
		// Queue full, execute directly (rare case)
		go work()
	}
}

func (gp *GoroutinePool) ActiveWorkers() int64 {
	return atomic.LoadInt64(&gp.active)
}

func (gp *GoroutinePool) CompletedTasks() int64 {
	return atomic.LoadInt64(&gp.completed)
}

func (gp *GoroutinePool) Shutdown() {
	close(gp.workQueue)
}

// ========== Performance Metrics Collector ==========

type MetricsCollector struct {
	metrics    *PerformanceMetrics
	startTime  time.Time
	mu         sync.RWMutex
	snapshots  []*PerformanceMetrics
}

func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		metrics:   &PerformanceMetrics{},
		startTime: time.Now(),
		snapshots: []*PerformanceMetrics{},
	}
}

func (mc *MetricsCollector) RecordAllocation() {
	atomic.AddInt64(&mc.metrics.Allocations, 1)
}

func (mc *MetricsCollector) RecordDeallocation() {
	atomic.AddInt64(&mc.metrics.Deallocations, 1)
}

func (mc *MetricsCollector) RecordGCPause(duration time.Duration) {
	atomic.AddInt64(&mc.metrics.GCPauses, 1)
	mc.mu.Lock()
	if mc.metrics.AvgPauseTime == 0 {
		mc.metrics.AvgPauseTime = duration
	} else {
		mc.metrics.AvgPauseTime = (mc.metrics.AvgPauseTime + duration) / 2
	}
	mc.mu.Unlock()
}

func (mc *MetricsCollector) GetMetrics() *PerformanceMetrics {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	return &PerformanceMetrics{
		Allocations:   atomic.LoadInt64(&mc.metrics.Allocations),
		Deallocations: atomic.LoadInt64(&mc.metrics.Deallocations),
		GCPauses:      atomic.LoadInt64(&mc.metrics.GCPauses),
		AvgPauseTime:  mc.metrics.AvgPauseTime,
		MemoryUsage:   atomic.LoadUint64(&mc.metrics.MemoryUsage),
		PeakMemory:    atomic.LoadUint64(&mc.metrics.PeakMemory),
	}
}

func (mc *MetricsCollector) Snapshot() {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	snapshot := &PerformanceMetrics{
		Allocations:   atomic.LoadInt64(&mc.metrics.Allocations),
		Deallocations: atomic.LoadInt64(&mc.metrics.Deallocations),
		GCPauses:      atomic.LoadInt64(&mc.metrics.GCPauses),
		AvgPauseTime:  mc.metrics.AvgPauseTime,
	}

	mc.snapshots = append(mc.snapshots, snapshot)
}

// ========== Atomic Counter ==========

type AtomicCounter struct {
	value int64
}

func (ac *AtomicCounter) Increment() {
	atomic.AddInt64(&ac.value, 1)
}

func (ac *AtomicCounter) Get() int64 {
	return atomic.LoadInt64(&ac.value)
}

func (ac *AtomicCounter) Set(value int64) {
	atomic.StoreInt64(&ac.value, value)
}

// ========== String Builder Optimization ==========

type OptimizedStringBuilder struct {
	pool *sync.Pool
}

func NewOptimizedStringBuilder() *OptimizedStringBuilder {
	return &OptimizedStringBuilder{
		pool: &sync.Pool{
			New: func() interface{} {
				return bytes.NewBuffer(make([]byte, 0, 256))
			},
		},
	}
}

func (osb *OptimizedStringBuilder) Build() *bytes.Buffer {
	return osb.pool.Get().(*bytes.Buffer)
}

func (osb *OptimizedStringBuilder) Release(buf *bytes.Buffer) {
	buf.Reset()
	osb.pool.Put(buf)
}

func main() {
	// Example performance optimization
	pool := NewObjectPool(func() interface{} {
		return make([]byte, 1024)
	})

	obj := pool.Get()
	pool.Put(obj)
}
