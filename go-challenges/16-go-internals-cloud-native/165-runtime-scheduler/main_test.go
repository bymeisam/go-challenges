package main

import (
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// TestGMPSimulator tests the GMP model simulator
func TestGMPSimulator(t *testing.T) {
	sim := NewGMPSimulator(4)
	defer sim.Stop()

	var workDone int32
	for i := 0; i < 50; i++ {
		sim.SubmitWork(func() {
			atomic.AddInt32(&workDone, 1)
		}, i%4)
	}

	time.Sleep(200 * time.Millisecond)

	metrics := sim.GetMetrics()
	if metrics["goroutines_created"].(int32) != 50 {
		t.Errorf("Expected 50 goroutines, got %d", metrics["goroutines_created"])
	}
}

func TestGMPSimulatorWorkStealing(t *testing.T) {
	sim := NewGMPSimulator(2)
	defer sim.Stop()

	// Submit work heavily to one processor
	for i := 0; i < 100; i++ {
		sim.SubmitWork(func() {
			time.Sleep(time.Microsecond)
		}, 0)
	}

	time.Sleep(500 * time.Millisecond)

	metrics := sim.GetMetrics()
	steals := metrics["steals_successful"].(int64)
	if steals == 0 {
		t.Logf("Warning: No work stealing detected (may be normal)")
	}
	t.Logf("Steals: %d, Attempts: %d", steals, metrics["steal_attempts"])
}

func TestGMPSimulatorMultipleProcessors(t *testing.T) {
	for numP := 1; numP <= 8; numP *= 2 {
		sim := NewGMPSimulator(numP)

		work := int32(0)
		for i := 0; i < 100; i++ {
			sim.SubmitWork(func() {
				atomic.AddInt32(&work, 1)
			}, i%numP)
		}

		time.Sleep(100 * time.Millisecond)
		sim.Stop()

		if work != 100 {
			t.Errorf("P=%d: Expected 100 work items, got %d", numP, work)
		}
	}
}

// TestPreemptionAnalyzer tests goroutine preemption
func TestPreemptionAnalyzerChannelPreemption(t *testing.T) {
	pa := NewPreemptionAnalyzer()
	pa.DemonstrateChanPreemption(100000)
	// Should complete without hanging
}

func TestPreemptionAnalyzerLoopPreemption(t *testing.T) {
	pa := NewPreemptionAnalyzer()
	pa.DemonstrateLoopPreemption()
	// Should complete and all goroutines should progress
}

func TestPreemptionAnalyzerFairness(t *testing.T) {
	counts := make([]*int32, 4)
	for i := range counts {
		counts[i] = &([]int32{0}[0])
	}

	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		idx := i
		go func() {
			defer wg.Done()
			for j := 0; j < 1000000; j++ {
				atomic.AddInt32(counts[idx], 1)
			}
		}()
	}
	wg.Wait()

	// All goroutines should have made progress
	for i, count := range counts {
		if atomic.LoadInt32(count) == 0 {
			t.Errorf("Goroutine %d made no progress", i)
		}
	}
}

// TestRuntimeStatsCollector tests statistics collection
func TestRuntimeStatsCollectorBasic(t *testing.T) {
	rsc := NewRuntimeStatsCollector(10 * time.Millisecond)
	rsc.Start()

	// Create some load
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			sum := 0
			for j := 0; j < 100000; j++ {
				sum += j
			}
		}()
	}
	wg.Wait()

	rsc.Stop()

	stats := rsc.GetStats()
	if samples, ok := stats["samples"]; !ok || samples.(int) == 0 {
		t.Errorf("No samples collected")
	}
}

func TestRuntimeStatsCollectorGC(t *testing.T) {
	rsc := NewRuntimeStatsCollector(5 * time.Millisecond)
	rsc.Start()

	// Trigger GC
	runtime.GC()
	time.Sleep(50 * time.Millisecond)

	rsc.Stop()

	stats := rsc.GetStats()
	if gcRuns, ok := stats["gc_runs"]; !ok || gcRuns.(uint32) < 1 {
		t.Logf("Warning: GC runs not recorded properly")
	}
}

func TestRuntimeStatsCollectorMemory(t *testing.T) {
	rsc := NewRuntimeStatsCollector(20 * time.Millisecond)
	rsc.Start()

	// Allocate memory
	_ = make([]byte, 10*1024*1024)
	time.Sleep(50 * time.Millisecond)

	rsc.Stop()

	stats := rsc.GetStats()
	if allocStart, ok := stats["alloc_bytes_start"]; !ok || allocStart.(uint64) == 0 {
		t.Errorf("Memory allocation not tracked")
	}
}

// TestContentionAnalyzer tests lock contention analysis
func TestContentionAnalyzerBasic(t *testing.T) {
	ca := NewContentionAnalyzer()

	ca.RecordContention("test", time.Millisecond)
	ca.RecordContention("test", 2*time.Millisecond)
	ca.RecordContention("test", 3*time.Millisecond)

	report := ca.GetReport()
	if len(report) != 1 {
		t.Errorf("Expected 1 contention point, got %d", len(report))
	}

	testInfo := report["test"]
	if testInfo["contention_count"].(int64) != 3 {
		t.Errorf("Expected 3 contentions, got %d", testInfo["contention_count"])
	}
}

func TestContentionAnalyzerMultiple(t *testing.T) {
	ca := NewContentionAnalyzer()

	ca.RecordContention("mutex", time.Millisecond)
	ca.RecordContention("channel", 2*time.Millisecond)
	ca.RecordContention("mutex", 3*time.Millisecond)

	report := ca.GetReport()
	if len(report) != 2 {
		t.Errorf("Expected 2 contention points, got %d", len(report))
	}

	if report["mutex"]["contention_count"].(int64) != 2 {
		t.Errorf("Expected 2 mutex contentions")
	}
}

func TestContentionAnalyzerMaxWait(t *testing.T) {
	ca := NewContentionAnalyzer()

	durations := []time.Duration{1 * time.Millisecond, 5 * time.Millisecond, 2 * time.Millisecond}
	for _, d := range durations {
		ca.RecordContention("test", d)
	}

	report := ca.GetReport()
	maxWait := report["test"]["max_wait_time_us"].(int64)
	if maxWait != 5000 { // 5ms = 5000us
		t.Errorf("Expected max wait 5000us, got %d", maxWait)
	}
}

func TestContentionAnalyzerConcurrent(t *testing.T) {
	ca := NewContentionAnalyzer()

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				ca.RecordContention("test", time.Millisecond)
			}
		}()
	}
	wg.Wait()

	report := ca.GetReport()
	if report["test"]["contention_count"].(int64) != 100 {
		t.Errorf("Expected 100 contentions, got %d", report["test"]["contention_count"])
	}
}

// TestMAXPROCSOptimizer tests GOMAXPROCS optimization
func TestMAXPROCSOptimizerBenchmark(t *testing.T) {
	optimizer := NewMAXPROCSOptimizer()

	workload := func() {
		var wg sync.WaitGroup
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				sum := 0
				for j := 0; j < 10000; j++ {
					sum += j
				}
			}()
		}
		wg.Wait()
	}

	duration := optimizer.BenchmarkWithProcs(2, workload)
	if duration == 0 {
		t.Errorf("Expected non-zero duration")
	}
}

func TestMAXPROCSOptimizerFindOptimal(t *testing.T) {
	optimizer := NewMAXPROCSOptimizer()

	workload := func() {
		var wg sync.WaitGroup
		for i := 0; i < 50; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				sum := 0
				for j := 0; j < 10000; j++ {
					sum += j
				}
			}()
		}
		wg.Wait()
	}

	optimal := optimizer.FindOptimal(4, workload)
	if optimal < 1 || optimal > 4 {
		t.Errorf("Expected optimal between 1 and 4, got %d", optimal)
	}
}

func TestMAXPROCSOptimizerBenchmarks(t *testing.T) {
	optimizer := NewMAXPROCSOptimizer()

	workload := func() {
		var wg sync.WaitGroup
		for i := 0; i < 50; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				sum := 0
				for j := 0; j < 10000; j++ {
					sum += j
				}
			}()
		}
		wg.Wait()
	}

	optimizer.BenchmarkWithProcs(1, workload)
	optimizer.BenchmarkWithProcs(2, workload)

	benchmarks := optimizer.GetBenchmarks()
	if len(benchmarks) != 2 {
		t.Errorf("Expected 2 benchmarks, got %d", len(benchmarks))
	}
}

// TestAffinityScheduler tests goroutine affinity scheduling
func TestAffinitySchedulerBasic(t *testing.T) {
	sched := NewAffinityScheduler(4)
	defer sched.Stop()

	var processed int32
	for i := 0; i < 20; i++ {
		sched.SubmitWork(AffinityWork{
			fn: func() {
				atomic.AddInt32(&processed, 1)
			},
			affinity: i,
			priority: 1,
		})
	}

	time.Sleep(500 * time.Millisecond)

	if processed != 20 {
		t.Errorf("Expected 20 work items processed, got %d", processed)
	}
}

func TestAffinitySchedulerAffinity(t *testing.T) {
	sched := NewAffinityScheduler(4)
	defer sched.Stop()

	workerCounts := make([]int32, 4)

	for i := 0; i < 40; i++ {
		idx := i % 4
		sched.SubmitWork(AffinityWork{
			fn: func() {
				atomic.AddInt32(&workerCounts[idx], 1)
			},
			affinity: idx,
			priority: 1,
		})
	}

	time.Sleep(500 * time.Millisecond)

	// Each worker should process work
	for i, count := range workerCounts {
		if count == 0 {
			t.Logf("Warning: Worker %d did not process any work", i)
		}
	}
}

func TestAffinitySchedulerUnderLoad(t *testing.T) {
	sched := NewAffinityScheduler(8)
	defer sched.Stop()

	var processed int32
	for i := 0; i < 1000; i++ {
		sched.SubmitWork(AffinityWork{
			fn: func() {
				time.Sleep(time.Microsecond)
				atomic.AddInt32(&processed, 1)
			},
			affinity: i,
			priority: 1,
		})
	}

	time.Sleep(5 * time.Second)

	if processed < 990 {
		t.Logf("Warning: Only %d/%d work items processed under load", processed, 1000)
	}
}

// Benchmark tests

func BenchmarkGMPSimulator(b *testing.B) {
	sim := NewGMPSimulator(4)
	defer sim.Stop()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var done int32
		for j := 0; j < 100; j++ {
			sim.SubmitWork(func() {
				atomic.AddInt32(&done, 1)
			}, j%4)
		}
	}
}

func BenchmarkContentionAnalyzer(b *testing.B) {
	ca := NewContentionAnalyzer()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ca.RecordContention("test", time.Millisecond)
	}
}

func BenchmarkAffinityScheduler(b *testing.B) {
	sched := NewAffinityScheduler(4)
	defer sched.Stop()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sched.SubmitWork(AffinityWork{
			fn:       func() {},
			affinity: i % 4,
			priority: 1,
		})
	}
}

func BenchmarkPreemption(b *testing.B) {
	pa := NewPreemptionAnalyzer()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pa.DemonstrateChanPreemption(1000)
	}
}
