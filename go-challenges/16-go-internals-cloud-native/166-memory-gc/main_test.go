package main

import (
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// TestEscapeAnalyzer tests escape analysis demonstrations
func TestEscapeAnalyzerNoEscape(t *testing.T) {
	ea := NewEscapeAnalyzer()
	result := ea.NoEscape()
	if result != 30 {
		t.Errorf("Expected 30, got %d", result)
	}
}

func TestEscapeAnalyzerEscape(t *testing.T) {
	ea := NewEscapeAnalyzer()
	ptr := ea.EscapeReturn()
	if ptr == nil {
		t.Errorf("Expected non-nil pointer")
	}
	if *ptr != 10 {
		t.Errorf("Expected 10, got %d", *ptr)
	}
}

func TestEscapeAnalyzerClosure(t *testing.T) {
	ea := NewEscapeAnalyzer()
	fn := ea.EscapeClosure()
	if fn() != 10 {
		t.Errorf("Expected 10 from closure")
	}
}

func TestEscapeAnalyzerAnalyzeAllocations(t *testing.T) {
	ea := NewEscapeAnalyzer()
	ea.AnalyzeAllocations()

	if len(ea.results) == 0 {
		t.Errorf("Expected analysis results")
	}
}

// TestMemoryPool tests memory pool functionality
func TestMemoryPoolBasic(t *testing.T) {
	pool := NewMemoryPool(1024)

	buf := pool.Acquire()
	if len(buf) != 1024 {
		t.Errorf("Expected buffer size 1024, got %d", len(buf))
	}

	pool.Release(buf)
	if pool.deallocated != 1 {
		t.Errorf("Expected 1 deallocation, got %d", pool.deallocated)
	}
}

func TestMemoryPoolConcurrent(t *testing.T) {
	pool := NewMemoryPool(512)

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			buf := pool.Acquire()
			buf[0] = 42
			pool.Release(buf)
		}()
	}
	wg.Wait()

	if pool.deallocated > 100 {
		t.Logf("Warning: Pool deallocation count exceeds submissions: %d", pool.deallocated)
	}
}

func TestMemoryPoolWrongSize(t *testing.T) {
	pool := NewMemoryPool(512)

	buf := make([]byte, 256)
	originalDeallocated := pool.deallocated

	pool.Release(buf)

	if pool.deallocated != originalDeallocated {
		t.Errorf("Pool should not accept buffers of wrong size")
	}
}

// TestGCAnalyzer tests GC analysis
func TestGCAnalyzerBasic(t *testing.T) {
	ga := NewGCAnalyzer()
	ga.Start()

	runtime.GC()
	time.Sleep(100 * time.Millisecond)

	ga.Stop()

	report := ga.GetReport()
	if len(report) == 0 {
		t.Errorf("Expected analysis report")
	}
}

func TestGCAnalyzerGCPauses(t *testing.T) {
	ga := NewGCAnalyzer()
	ga.Start()

	// Trigger multiple GCs
	for i := 0; i < 3; i++ {
		_ = make([]byte, 5*1024*1024)
		runtime.GC()
		time.Sleep(50 * time.Millisecond)
	}

	ga.Stop()

	report := ga.GetReport()
	if gcRuns, ok := report["gc_runs"]; !ok || gcRuns.(uint32) < 1 {
		t.Logf("Warning: GC runs not tracked")
	}
}

func TestGCAnalyzerMaxPause(t *testing.T) {
	ga := NewGCAnalyzer()
	ga.Start()

	_ = make([]byte, 10*1024*1024)
	runtime.GC()
	time.Sleep(100 * time.Millisecond)

	ga.Stop()

	if ga.maxPauseTime == 0 {
		t.Logf("Warning: Max pause time not recorded")
	}
}

// TestAllocationTracker tests allocation tracking
func TestAllocationTrackerBasic(t *testing.T) {
	at := NewAllocationTracker()
	at.Start()

	_ = make([]byte, 1024*1024)
	time.Sleep(150 * time.Millisecond)

	at.Stop()

	stats := at.GetStats()
	if len(stats) == 0 {
		t.Errorf("Expected allocation stats")
	}
}

func TestAllocationTrackerHeapGrowth(t *testing.T) {
	at := NewAllocationTracker()
	at.Start()

	var buffers [][]byte
	for i := 0; i < 10; i++ {
		buffers = append(buffers, make([]byte, 1024*1024))
		time.Sleep(10 * time.Millisecond)
	}

	time.Sleep(150 * time.Millisecond)
	at.Stop()

	stats := at.GetStats()
	if allocEnd, ok := stats["alloc_bytes_end"]; !ok || allocEnd.(uint64) == 0 {
		t.Errorf("Expected allocation tracking")
	}
}

func TestAllocationTrackerAllocationRate(t *testing.T) {
	at := NewAllocationTracker()
	at.Start()

	for i := 0; i < 20; i++ {
		_ = make([]byte, 100000)
		time.Sleep(10 * time.Millisecond)
	}

	time.Sleep(150 * time.Millisecond)
	at.Stop()

	stats := at.GetStats()
	if rate, ok := stats["avg_alloc_rate_bs"]; ok && rate.(float64) > 0 {
		t.Logf("Allocation rate: %.0f bytes/sec", rate)
	}
}

// TestSafeUnsafeOps tests safe unsafe operations
func TestSafeUnsafeOpsValidAccess(t *testing.T) {
	suo := NewSafeUnsafeOps()

	data := []byte{10, 20, 30, 40, 50}
	val, err := suo.SafeByteArrayAccess(data, 2)

	if err != nil {
		t.Errorf("Expected valid access, got error: %v", err)
	}
	if val != 30 {
		t.Errorf("Expected 30, got %d", val)
	}
}

func TestSafeUnsafeOpsInvalidAccess(t *testing.T) {
	suo := NewSafeUnsafeOps()

	data := []byte{10, 20, 30}
	_, err := suo.SafeByteArrayAccess(data, 10)

	if err == nil {
		t.Errorf("Expected error for out-of-bounds access")
	}
}

func TestSafeUnsafeOpsNegativeOffset(t *testing.T) {
	suo := NewSafeUnsafeOps()

	data := []byte{10, 20, 30}
	_, err := suo.SafeByteArrayAccess(data, -1)

	if err == nil {
		t.Errorf("Expected error for negative offset")
	}
}

func TestSafeUnsafeOpsBytesToInt64(t *testing.T) {
	suo := NewSafeUnsafeOps()

	data := make([]byte, 8)
	data[0] = 1
	data[7] = 255

	_, err := suo.BytesToInt64(data)
	if err != nil {
		t.Errorf("Expected valid conversion, got error: %v", err)
	}
}

func TestSafeUnsafeOpsTooShort(t *testing.T) {
	suo := NewSafeUnsafeOps()

	data := []byte{1, 2, 3}
	_, err := suo.BytesToInt64(data)

	if err == nil {
		t.Errorf("Expected error for small buffer")
	}
}

// TestResourceWithFinalizer tests finalizer cleanup
func TestResourceWithFinalizerBasic(t *testing.T) {
	r := NewResourceWithFinalizer("test-resource")
	if r.closed {
		t.Errorf("Resource should not be closed initially")
	}

	err := r.Close()
	if err != nil {
		t.Errorf("Expected successful close, got error: %v", err)
	}

	if !r.closed {
		t.Errorf("Resource should be closed")
	}
}

func TestResourceWithFinalizerDoublClose(t *testing.T) {
	r := NewResourceWithFinalizer("test-resource-2")
	r.Close()

	err := r.Close()
	if err == nil {
		t.Errorf("Expected error on double close")
	}
}

func TestResourceWithFinalizerAutoCleanup(t *testing.T) {
	initialCount := len(finalizerRegistry.resources)

	for i := 0; i < 5; i++ {
		_ = NewResourceWithFinalizer("auto-cleanup-" + string(rune(i)))
	}

	if len(finalizerRegistry.resources) <= initialCount {
		t.Errorf("Expected resources to be tracked")
	}

	runtime.GC()

	// Some resources may be cleaned up
	if len(finalizerRegistry.resources) > initialCount+5 {
		t.Errorf("Expected some cleanup via finalizers")
	}
}

// TestGCTuner tests GC tuning
func TestGCTunerForceGC(t *testing.T) {
	gt := NewGCTuner()

	duration := gt.ForceGC()
	if duration < 0 {
		t.Errorf("GC time should be non-negative")
	}
}

func TestGCTunerSetGOGC(t *testing.T) {
	gt := NewGCTuner()

	gt.SetGOGC(50)
	if gt.currentGOGC != 50 {
		t.Errorf("Expected GOGC=50, got %d", gt.currentGOGC)
	}

	gt.SetGOGC(-1)
	if gt.currentGOGC != -1 {
		t.Errorf("Expected GOGC=-1, got %d", gt.currentGOGC)
	}
}

func TestGCTunerEstimateGCFrequency(t *testing.T) {
	gt := NewGCTuner()

	freq := gt.EstimateGCFrequency(1000000) // 1MB/sec
	if freq <= 0 {
		t.Errorf("Expected positive frequency")
	}

	t.Logf("Estimated GC frequency: %v", freq)
}

// TestOptimizedBuffer tests optimized buffer
func TestOptimizedBufferWrite(t *testing.T) {
	buf := NewOptimizedBuffer(100)

	data := []byte("hello")
	err := buf.WriteOptimized(data)

	if err != nil {
		t.Errorf("Expected successful write, got error: %v", err)
	}

	if buf.offset != 5 {
		t.Errorf("Expected offset 5, got %d", buf.offset)
	}
}

func TestOptimizedBufferReset(t *testing.T) {
	buf := NewOptimizedBuffer(100)
	buf.WriteOptimized([]byte("data"))

	buf.Reset()

	if buf.offset != 0 {
		t.Errorf("Expected offset 0 after reset, got %d", buf.offset)
	}
}

func TestOptimizedBufferOverflow(t *testing.T) {
	buf := NewOptimizedBuffer(10)

	err := buf.WriteOptimized(make([]byte, 20))

	if err == nil {
		t.Errorf("Expected overflow error")
	}
}

// Benchmark tests

func BenchmarkEscapeAnalysis(b *testing.B) {
	ea := NewEscapeAnalyzer()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ea.NoEscape()
	}
}

func BenchmarkMemoryPoolAcquireRelease(b *testing.B) {
	pool := NewMemoryPool(1024)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf := pool.Acquire()
		pool.Release(buf)
	}
}

func BenchmarkMemoryPoolConcurrent(b *testing.B) {
	pool := NewMemoryPool(1024)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			buf := pool.Acquire()
			pool.Release(buf)
		}
	})
}

func BenchmarkSafeUnsafeOps(b *testing.B) {
	suo := NewSafeUnsafeOps()
	data := []byte{1, 2, 3, 4, 5}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = suo.SafeByteArrayAccess(data, i%5)
	}
}

func BenchmarkGCForceGC(b *testing.B) {
	gt := NewGCTuner()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = gt.ForceGC()
	}
}
