package main

import (
	"fmt"
	"runtime"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"
)

// Challenge 166: Memory Management & GC Tuning
// Escape Analysis, Stack vs Heap, GC Tuning, Memory Profiling, Finalizers

// ===== 1. Escape Analysis Demonstrator =====

type EscapeAnalyzer struct {
	results map[string]interface{}
	mu      sync.RWMutex
}

func NewEscapeAnalyzer() *EscapeAnalyzer {
	return &EscapeAnalyzer{
		results: make(map[string]interface{}),
	}
}

// Example 1: Stack allocation (does NOT escape)
func (ea *EscapeAnalyzer) NoEscape() int {
	x := 10
	y := 20
	return x + y // No allocation, values stay on stack
}

// Example 2: Escapes due to return pointer
func (ea *EscapeAnalyzer) EscapeReturn() *int {
	x := 10
	return &x // Escapes because pointer returned
}

// Example 3: Escapes due to interface{}
func (ea *EscapeAnalyzer) EscapeInterface(x int) {
	var i interface{} = x // No escape - value type
}

// Example 4: Escapes in slice
func (ea *EscapeAnalyzer) EscapeSlice() []*int {
	slice := make([]*int, 1)
	x := 10
	slice[0] = &x // Escapes
	return slice
}

// Example 5: Closure captures (escapes)
func (ea *EscapeAnalyzer) EscapeClosure() func() int {
	x := 10
	return func() int { // x escapes into closure
		return x
	}
}

// Analyze memory allocation
func (ea *EscapeAnalyzer) AnalyzeAllocations() {
	var m1, m2 runtime.MemStats
	runtime.ReadMemStats(&m1)

	// Call allocation-heavy functions
	for i := 0; i < 1000; i++ {
		_ = ea.NoEscape()
	}

	runtime.ReadMemStats(&m2)
	noEscapeAlloc := m2.Alloc - m1.Alloc

	runtime.ReadMemStats(&m1)
	for i := 0; i < 1000; i++ {
		_ = ea.EscapeReturn()
	}
	runtime.ReadMemStats(&m2)
	escapeAlloc := m2.Alloc - m1.Alloc

	ea.mu.Lock()
	ea.results["no_escape_alloc"] = noEscapeAlloc
	ea.results["escape_alloc"] = escapeAlloc
	ea.mu.Unlock()
}

// ===== 2. Memory Pool Implementation =====

type MemoryPool struct {
	itemSize    int
	pool        sync.Pool
	allocated   int64
	deallocated int64
}

func NewMemoryPool(itemSize int) *MemoryPool {
	return &MemoryPool{
		itemSize: itemSize,
		pool: sync.Pool{
			New: func() interface{} {
				atomic.AddInt64(&MemoryPool{}.allocated, 1)
				return make([]byte, itemSize)
			},
		},
	}
}

func (mp *MemoryPool) Acquire() []byte {
	return mp.pool.Get().([]byte)
}

func (mp *MemoryPool) Release(buf []byte) {
	if len(buf) == mp.itemSize {
		atomic.AddInt64(&mp.deallocated, 1)
		mp.pool.Put(buf)
	}
}

// ===== 3. GC Analyzer =====

type GCAnalyzer struct {
	samples      []GCSample
	mu           sync.RWMutex
	monitoring   bool
	stopChan     chan struct{}
	pauseTimes   []time.Duration
	maxPauseTime time.Duration
}

type GCSample struct {
	Timestamp   time.Time
	NumGC       uint32
	PauseNs     uint64
	PauseTotal  uint64
	HeapAlloc   uint64
	HeapSys     uint64
	HeapObjects uint64
	Goroutines  int
}

func NewGCAnalyzer() *GCAnalyzer {
	return &GCAnalyzer{
		samples:  make([]GCSample, 0, 1000),
		stopChan: make(chan struct{}),
		pauseTimes: make([]time.Duration, 0),
	}
}

func (ga *GCAnalyzer) Start() {
	ga.monitoring = true
	go func() {
		ticker := time.NewTicker(10 * time.Millisecond)
		defer ticker.Stop()

		prevNumGC := uint32(0)

		for {
			select {
			case <-ga.stopChan:
				return
			case <-ticker.C:
				var m runtime.MemStats
				runtime.ReadMemStats(&m)

				sample := GCSample{
					Timestamp:   time.Now(),
					NumGC:       m.NumGC,
					HeapAlloc:   m.HeapAlloc,
					HeapSys:     m.HeapSys,
					HeapObjects: m.HeapObjects,
					Goroutines:  runtime.NumGoroutine(),
				}

				// Track pause times
				if m.NumGC > prevNumGC {
					pauseNs := m.PauseNs[(m.NumGC+255)%256]
					pauseDur := time.Duration(pauseNs) * time.Nanosecond
					ga.pauseTimes = append(ga.pauseTimes, pauseDur)
					if pauseDur > ga.maxPauseTime {
						ga.maxPauseTime = pauseDur
					}
					sample.PauseNs = pauseNs
					prevNumGC = m.NumGC
				}

				ga.mu.Lock()
				ga.samples = append(ga.samples, sample)
				ga.mu.Unlock()
			}
		}
	}()
}

func (ga *GCAnalyzer) Stop() {
	close(ga.stopChan)
	time.Sleep(50 * time.Millisecond)
	ga.monitoring = false
}

func (ga *GCAnalyzer) GetReport() map[string]interface{} {
	ga.mu.RLock()
	defer ga.mu.RUnlock()

	if len(ga.samples) == 0 {
		return make(map[string]interface{})
	}

	first := ga.samples[0]
	last := ga.samples[len(ga.samples)-1]

	var totalPause time.Duration
	for _, p := range ga.pauseTimes {
		totalPause += p
	}

	var avgPause time.Duration
	if len(ga.pauseTimes) > 0 {
		avgPause = time.Duration(totalPause.Nanoseconds() / int64(len(ga.pauseTimes)))
	}

	return map[string]interface{}{
		"samples":              len(ga.samples),
		"duration_sec":         last.Timestamp.Sub(first.Timestamp).Seconds(),
		"gc_runs":              last.NumGC - first.NumGC,
		"gc_pause_count":       len(ga.pauseTimes),
		"max_pause_time_us":    ga.maxPauseTime.Microseconds(),
		"avg_pause_time_us":    avgPause.Microseconds(),
		"total_pause_time_ms":  totalPause.Milliseconds(),
		"heap_alloc_start":     first.HeapAlloc,
		"heap_alloc_end":       last.HeapAlloc,
		"heap_sys_start":       first.HeapSys,
		"heap_sys_end":         last.HeapSys,
		"avg_heap_objects":     (first.HeapObjects + last.HeapObjects) / 2,
		"avg_goroutines":       (uint64(first.Goroutines) + uint64(last.Goroutines)) / 2,
	}
}

// ===== 4. Allocation Tracker =====

type AllocationTracker struct {
	snapshots  []AllocationSnapshot
	mu         sync.RWMutex
	monitoring bool
	stopChan   chan struct{}
}

type AllocationSnapshot struct {
	Timestamp      time.Time
	Alloc          uint64
	TotalAlloc     uint64
	Sys            uint64
	NumAlloc       uint64
	NumFrees       uint64
	AllocationRate float64
}

func NewAllocationTracker() *AllocationTracker {
	return &AllocationTracker{
		snapshots: make([]AllocationSnapshot, 0, 1000),
		stopChan:  make(chan struct{}),
	}
}

func (at *AllocationTracker) Start() {
	at.monitoring = true
	go func() {
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()

		var prevAlloc uint64

		for {
			select {
			case <-at.stopChan:
				return
			case <-ticker.C:
				var m runtime.MemStats
				runtime.ReadMemStats(&m)

				allocDiff := m.Alloc - prevAlloc
				allocationRate := float64(allocDiff) / 0.1 // per second

				snapshot := AllocationSnapshot{
					Timestamp:      time.Now(),
					Alloc:          m.Alloc,
					TotalAlloc:     m.TotalAlloc,
					Sys:            m.Sys,
					NumAlloc:       m.Mallocs,
					NumFrees:       m.Frees,
					AllocationRate: allocationRate,
				}

				at.mu.Lock()
				at.snapshots = append(at.snapshots, snapshot)
				at.mu.Unlock()

				prevAlloc = m.Alloc
			}
		}
	}()
}

func (at *AllocationTracker) Stop() {
	close(at.stopChan)
	time.Sleep(150 * time.Millisecond)
	at.monitoring = false
}

func (at *AllocationTracker) GetStats() map[string]interface{} {
	at.mu.RLock()
	defer at.mu.RUnlock()

	if len(at.snapshots) == 0 {
		return make(map[string]interface{})
	}

	first := at.snapshots[0]
	last := at.snapshots[len(at.snapshots)-1]

	var maxAllocRate float64
	var avgAllocRate float64
	for _, s := range at.snapshots {
		if s.AllocationRate > maxAllocRate {
			maxAllocRate = s.AllocationRate
		}
		avgAllocRate += s.AllocationRate
	}
	avgAllocRate /= float64(len(at.snapshots))

	return map[string]interface{}{
		"samples":            len(at.snapshots),
		"duration_sec":       last.Timestamp.Sub(first.Timestamp).Seconds(),
		"alloc_bytes_start":  first.Alloc,
		"alloc_bytes_end":    last.Alloc,
		"total_alloc_bytes":  last.TotalAlloc,
		"sys_bytes":          last.Sys,
		"num_allocs":         last.NumAlloc - first.NumAlloc,
		"num_frees":          last.NumFrees - first.NumFrees,
		"max_alloc_rate_bs":  maxAllocRate,
		"avg_alloc_rate_bs":  avgAllocRate,
	}
}

// ===== 5. Safe Unsafe Operations =====

type SafeUnsafeOps struct {
	accesses int64
	errors   int64
}

func NewSafeUnsafeOps() *SafeUnsafeOps {
	return &SafeUnsafeOps{}
}

// Safe pointer arithmetic with bounds checking
func (suo *SafeUnsafeOps) SafeByteArrayAccess(arr []byte, offset int) (byte, error) {
	atomic.AddInt64(&suo.accesses, 1)

	if offset < 0 || offset >= len(arr) {
		atomic.AddInt64(&suo.errors, 1)
		return 0, fmt.Errorf("invalid offset: %d", offset)
	}

	// Use unsafe.Pointer with validation
	ptr := unsafe.Pointer(&arr[0])
	offset64 := uintptr(offset)

	if offset64 >= uintptr(len(arr)) {
		atomic.AddInt64(&suo.errors, 1)
		return 0, fmt.Errorf("bounds check failed")
	}

	return *(*byte)(unsafe.Pointer(uintptr(ptr) + offset64)), nil
}

// Type casting without unsafe pointers (preferred)
func (suo *SafeUnsafeOps) BytesToInt64(b []byte) (int64, error) {
	if len(b) < 8 {
		return 0, fmt.Errorf("slice too small")
	}

	// Safe approach - create a new array
	var buf [8]byte
	copy(buf[:], b[:8])

	// Or use unsafe with proper validation
	if len(b) >= 8 {
		ptr := unsafe.Pointer(&b[0])
		return *(*int64)(ptr), nil
	}

	return 0, fmt.Errorf("insufficient data")
}

// ===== 6. Finalizer-based Resource Cleanup =====

type ResourceWithFinalizer struct {
	id     string
	closed bool
}

var finalizerRegistry = struct {
	mu        sync.Mutex
	resources map[string]*ResourceWithFinalizer
}{
	resources: make(map[string]*ResourceWithFinalizer),
}

func NewResourceWithFinalizer(id string) *ResourceWithFinalizer {
	r := &ResourceWithFinalizer{
		id:     id,
		closed: false,
	}

	finalizerRegistry.mu.Lock()
	finalizerRegistry.resources[id] = r
	finalizerRegistry.mu.Unlock()

	// Set finalizer
	runtime.SetFinalizer(r, (*ResourceWithFinalizer).cleanup)

	return r
}

func (r *ResourceWithFinalizer) cleanup() {
	if !r.closed {
		r.closed = true
		finalizerRegistry.mu.Lock()
		delete(finalizerRegistry.resources, r.id)
		finalizerRegistry.mu.Unlock()
	}
}

func (r *ResourceWithFinalizer) Close() error {
	if r.closed {
		return fmt.Errorf("already closed")
	}
	r.cleanup()
	runtime.SetFinalizer(r, nil) // Disable finalizer
	return nil
}

// ===== 7. GC Tuning Helper =====

type GCTuner struct {
	targetHeapSize  uint64
	currentGOGC     int
	currentMEMLIMIT string
}

func NewGCTuner() *GCTuner {
	return &GCTuner{
		currentGOGC: 100, // Default GOGC value
	}
}

func (gt *GCTuner) SetGOGC(percentage int) {
	if percentage < 0 {
		percentage = -1 // Disable GC
	}
	debug.SetGCPercent(percentage)
	gt.currentGOGC = percentage
}

func (gt *GCTuner) ForceGC() time.Duration {
	start := time.Now()
	runtime.GC()
	return time.Since(start)
}

func (gt *GCTuner) EstimateGCFrequency(allocationRate float64) time.Duration {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	if allocationRate == 0 {
		return time.Hour
	}

	// Estimate time to reach next GC threshold
	heapTarget := uint64(float64(m.HeapAlloc) * 1.0 / (1.0 - float64(gt.currentGOGC)/100))
	heapSpace := heapTarget - m.HeapAlloc

	secondsToGC := float64(heapSpace) / allocationRate
	return time.Duration(secondsToGC) * time.Second
}

// ===== 8. Optimization Patterns =====

type OptimizedBuffer struct {
	buf    []byte
	offset int
}

func NewOptimizedBuffer(size int) *OptimizedBuffer {
	return &OptimizedBuffer{
		buf:    make([]byte, size),
		offset: 0,
	}
}

func (ob *OptimizedBuffer) WriteOptimized(data []byte) error {
	if ob.offset+len(data) > len(ob.buf) {
		return fmt.Errorf("buffer full")
	}
	copy(ob.buf[ob.offset:], data)
	ob.offset += len(data)
	return nil
}

func (ob *OptimizedBuffer) Reset() {
	ob.offset = 0
	// Don't zero the buffer - data is just unreachable
}

// ===== Main Demo =====

func main() {
	fmt.Println("=== Memory Management & GC Tuning ===\n")

	// 1. Escape Analysis
	fmt.Println("1. Escape Analysis Demonstrator")
	ea := NewEscapeAnalyzer()
	ea.AnalyzeAllocations()
	fmt.Printf("Escape analysis results: %+v\n\n", ea.results)

	// 2. Memory Pool
	fmt.Println("2. Memory Pool")
	pool := NewMemoryPool(1024)
	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			buf := pool.Acquire()
			copy(buf, []byte("test data"))
			pool.Release(buf)
		}()
	}
	wg.Wait()
	fmt.Printf("Memory pool: deallocated=%d\n\n", pool.deallocated)

	// 3. GC Analysis
	fmt.Println("3. GC Analyzer")
	ga := NewGCAnalyzer()
	ga.Start()

	// Create GC pressure
	for i := 0; i < 100; i++ {
		_ = make([]byte, 1024*1024)
	}
	runtime.GC()

	time.Sleep(200 * time.Millisecond)
	ga.Stop()

	report := ga.GetReport()
	fmt.Printf("GC Report: %+v\n\n", report)

	// 4. Allocation Tracking
	fmt.Println("4. Allocation Tracker")
	at := NewAllocationTracker()
	at.Start()

	// Allocate memory
	var buffers [][]byte
	for i := 0; i < 50; i++ {
		buffers = append(buffers, make([]byte, 100000))
	}

	time.Sleep(200 * time.Millisecond)
	at.Stop()

	allocStats := at.GetStats()
	fmt.Printf("Allocation Stats: %+v\n\n", allocStats)

	// 5. Safe Unsafe Operations
	fmt.Println("5. Safe Unsafe Operations")
	suo := NewSafeUnsafeOps()

	data := []byte{1, 2, 3, 4, 5}
	val, err := suo.SafeByteArrayAccess(data, 2)
	fmt.Printf("Byte at offset 2: %d (err=%v)\n", val, err)

	val, err = suo.SafeByteArrayAccess(data, 10)
	fmt.Printf("Byte at offset 10: %d (err=%v)\n", val, err)

	num, _ := suo.BytesToInt64(append([]byte{1, 2, 3, 4}, make([]byte, 4)...))
	fmt.Printf("Int64 conversion result: %d\n", num)
	fmt.Printf("Unsafe ops - accesses: %d, errors: %d\n\n", suo.accesses, suo.errors)

	// 6. Finalizers
	fmt.Println("6. Finalizer-based Resource Cleanup")
	r := NewResourceWithFinalizer("resource-1")
	r.Close()
	fmt.Println("Resource closed explicitly")

	// Create resources without explicit close
	for i := 0; i < 5; i++ {
		_ = NewResourceWithFinalizer(fmt.Sprintf("resource-%d", i+2))
	}

	runtime.GC()
	fmt.Printf("Active resources after GC: %d\n\n", len(finalizerRegistry.resources))

	// 7. GC Tuning
	fmt.Println("7. GC Tuning")
	gt := NewGCTuner()

	gcTime := gt.ForceGC()
	fmt.Printf("GC time: %v\n", gcTime)

	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	allocRate := float64(m.Alloc) / 1024 / 1024 // MB
	estimatedGCFreq := gt.EstimateGCFrequency(allocRate * 1e6)
	fmt.Printf("Estimated GC frequency: %v\n\n", estimatedGCFreq)

	// 8. Final Memory Stats
	fmt.Println("8. Final Memory Statistics")
	runtime.ReadMemStats(&m)
	fmt.Printf("Heap Alloc: %v MB\n", m.HeapAlloc/1024/1024)
	fmt.Printf("Heap Sys: %v MB\n", m.HeapSys/1024/1024)
	fmt.Printf("Heap Objects: %d\n", m.HeapObjects)
	fmt.Printf("Total Alloc: %v MB\n", m.TotalAlloc/1024/1024)
	fmt.Printf("GC Runs: %d\n", m.NumGC)

	fmt.Println("\n=== Complete ===")
}
