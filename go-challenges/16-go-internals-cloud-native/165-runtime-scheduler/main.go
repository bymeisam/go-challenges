package main

import (
	"fmt"
	"math/rand"
	"runtime"
	"runtime/debug"
	"runtime/metrics"
	"sync"
	"sync/atomic"
	"time"
)

// Challenge 165: Go Runtime & Scheduler Internals
// GMP Model, Scheduler Behavior, Runtime Stats, Work Stealing

// ===== 1. GMP Model Simulator =====

type GMPSimulator struct {
	numP          int
	numG          int32
	numM          int32
	globalQueue   chan func()
	processors    []*Processor
	machines      []*Machine
	metrics       *SchedulerMetrics
	mu            sync.RWMutex
	goroutineMap  map[int64]*GoroutineInfo
	stop          chan struct{}
	stealAttempts int64
	steals        int64
}

type GoroutineInfo struct {
	id        int64
	processor int
	startTime time.Time
	duration  time.Duration
	state     string // "running", "waiting", "done"
}

type Processor struct {
	id         int
	localQueue chan func()
	machine    *Machine
	mu         sync.Mutex
}

type Machine struct {
	id        int
	processor *Processor
	running   bool
	workDone  int64
}

type SchedulerMetrics struct {
	goroutinesCreated int64
	contextSwitches   int64
	stealsSuccessful  int64
	stealsFailed      int64
	avgScheduleTime   time.Duration
	totalScheduleTime time.Duration
	mu                sync.RWMutex
}

func NewGMPSimulator(numP int) *GMPSimulator {
	sim := &GMPSimulator{
		numP:         numP,
		globalQueue:  make(chan func(), 1000),
		processors:   make([]*Processor, numP),
		machines:     make([]*Machine, numP),
		metrics:      &SchedulerMetrics{},
		goroutineMap: make(map[int64]*GoroutineInfo),
		stop:         make(chan struct{}),
	}

	for i := 0; i < numP; i++ {
		p := &Processor{
			id:         i,
			localQueue: make(chan func(), 100),
		}
		m := &Machine{
			id:        i,
			processor: p,
			running:   true,
		}
		p.machine = m
		sim.processors[i] = p
		sim.machines[i] = m

		// Start machine worker
		go sim.machineWorker(m)
	}

	return sim
}

func (sim *GMPSimulator) machineWorker(m *Machine) {
	for {
		select {
		case <-sim.stop:
			m.running = false
			return
		case work := <-m.processor.localQueue:
			start := time.Now()
			work()
			atomic.AddInt64(&m.workDone, 1)
			atomic.AddInt64(&sim.metrics.contextSwitches, 1)
			sim.metrics.mu.Lock()
			sim.metrics.totalScheduleTime += time.Since(start)
			sim.metrics.mu.Unlock()
		case work := <-sim.globalQueue:
			start := time.Now()
			work()
			atomic.AddInt64(&m.workDone, 1)
			sim.metrics.mu.Lock()
			sim.metrics.totalScheduleTime += time.Since(start)
			sim.metrics.mu.Unlock()
		default:
			// Try work stealing
			if sim.tryWorkSteal(m.processor) {
				atomic.AddInt64(&sim.steals, 1)
			} else {
				atomic.AddInt64(&sim.stealAttempts, 1)
			}
			time.Sleep(time.Microsecond)
		}
	}
}

func (sim *GMPSimulator) tryWorkSteal(p *Processor) bool {
	// Try stealing from other processors
	for i := 0; i < len(sim.processors); i++ {
		if i == p.id {
			continue
		}
		other := sim.processors[i]
		select {
		case work := <-other.localQueue:
			// Successfully stole work
			select {
			case p.localQueue <- work:
				return true
			case sim.globalQueue <- work:
				return true
			}
		default:
			// No work available
		}
	}
	return false
}

func (sim *GMPSimulator) SubmitWork(work func(), p int) {
	if p >= 0 && p < sim.numP {
		select {
		case sim.processors[p].localQueue <- work:
		case sim.globalQueue <- work:
		}
	} else {
		sim.globalQueue <- work
	}
	atomic.AddInt32(&sim.numG, 1)
	atomic.AddInt64(&sim.metrics.goroutinesCreated, 1)
}

func (sim *GMPSimulator) GetMetrics() map[string]interface{} {
	sim.metrics.mu.RLock()
	defer sim.metrics.mu.RUnlock()

	avgSchedule := time.Duration(0)
	if sim.metrics.goroutinesCreated > 0 {
		avgSchedule = time.Duration(sim.metrics.totalScheduleTime.Nanoseconds() / sim.metrics.goroutinesCreated)
	}

	machineMetrics := make(map[string]int64)
	for _, m := range sim.machines {
		machineMetrics[fmt.Sprintf("M%d_work", m.id)] = atomic.LoadInt64(&m.workDone)
	}

	return map[string]interface{}{
		"processors_count":      sim.numP,
		"goroutines_created":    atomic.LoadInt32(&sim.numG),
		"context_switches":      atomic.LoadInt64(&sim.metrics.contextSwitches),
		"steals_successful":     atomic.LoadInt64(&sim.steals),
		"steal_attempts":        atomic.LoadInt64(&sim.stealAttempts),
		"avg_schedule_time_us":  avgSchedule.Microseconds(),
		"machine_work":          machineMetrics,
	}
}

func (sim *GMPSimulator) Stop() {
	close(sim.stop)
	time.Sleep(100 * time.Millisecond)
}

// ===== 2. Goroutine Preemption Analyzer =====

type PreemptionAnalyzer struct {
	preemptionPoints []PreemptionPoint
	mu               sync.RWMutex
}

type PreemptionPoint struct {
	Location string
	Count    int64
	LastSeen time.Time
}

func NewPreemptionAnalyzer() *PreemptionAnalyzer {
	return &PreemptionAnalyzer{
		preemptionPoints: make([]PreemptionPoint, 0),
	}
}

// Cooperative preemption through channel operations
func (pa *PreemptionAnalyzer) DemonstrateChanPreemption(iterations int) {
	done := make(chan bool)
	count := 0

	for i := 0; i < 10; i++ {
		go func(id int) {
			for j := 0; j < iterations; j++ {
				// Channel send/receive is a preemption point
				select {
				case <-done:
					return
				default:
					count++
				}
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
}

// Demonstrate cooperative preemption
func (pa *PreemptionAnalyzer) DemonstrateLoopPreemption() {
	// Modern Go: loops with backward jumps can be preempted
	var wg sync.WaitGroup
	results := make([]int, 4)

	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 1000000; j++ {
				results[id]++
				// Implicit preemption point in modern Go (1.16+)
			}
		}(i)
	}
	wg.Wait()
}

// ===== 3. Runtime Statistics Collector =====

type RuntimeStatsCollector struct {
	samples  []RuntimeSnapshot
	interval time.Duration
	mu       sync.RWMutex
	stop     chan struct{}
}

type RuntimeSnapshot struct {
	Timestamp      time.Time
	NumGoroutine   int
	Alloc          uint64
	TotalAlloc     uint64
	Sys            uint64
	NumGC          uint32
	PauseNs        []uint64
	PauseEnd       []uint64
	CPUFraction    float64
	NumCgoCall     int64
	MemStats       runtime.MemStats
}

func NewRuntimeStatsCollector(interval time.Duration) *RuntimeStatsCollector {
	return &RuntimeStatsCollector{
		samples:  make([]RuntimeSnapshot, 0, 1000),
		interval: interval,
		stop:     make(chan struct{}),
	}
}

func (rsc *RuntimeStatsCollector) Start() {
	go func() {
		ticker := time.NewTicker(rsc.interval)
		defer ticker.Stop()

		for {
			select {
			case <-rsc.stop:
				return
			case <-ticker.C:
				var m runtime.MemStats
				runtime.ReadMemStats(&m)

				snapshot := RuntimeSnapshot{
					Timestamp:    time.Now(),
					NumGoroutine: runtime.NumGoroutine(),
					Alloc:        m.Alloc,
					TotalAlloc:   m.TotalAlloc,
					Sys:          m.Sys,
					NumGC:        m.NumGC,
					NumCgoCall:   m.NumCgoCall,
					MemStats:     m,
				}

				rsc.mu.Lock()
				rsc.samples = append(rsc.samples, snapshot)
				rsc.mu.Unlock()
			}
		}
	}()
}

func (rsc *RuntimeStatsCollector) Stop() {
	close(rsc.stop)
	time.Sleep(100 * time.Millisecond)
}

func (rsc *RuntimeStatsCollector) GetStats() map[string]interface{} {
	rsc.mu.RLock()
	defer rsc.mu.RUnlock()

	if len(rsc.samples) == 0 {
		return map[string]interface{}{}
	}

	first := rsc.samples[0]
	last := rsc.samples[len(rsc.samples)-1]

	return map[string]interface{}{
		"samples":              len(rsc.samples),
		"duration_sec":         last.Timestamp.Sub(first.Timestamp).Seconds(),
		"avg_goroutines":       calculateAvg(rsc.samples, func(s RuntimeSnapshot) uint64 { return uint64(s.NumGoroutine) }),
		"max_goroutines":       calculateMax(rsc.samples, func(s RuntimeSnapshot) uint64 { return uint64(s.NumGoroutine) }),
		"alloc_bytes_start":    first.Alloc,
		"alloc_bytes_end":      last.Alloc,
		"total_alloc_bytes":    last.TotalAlloc,
		"sys_bytes":            last.Sys,
		"gc_runs":              last.NumGC - first.NumGC,
		"heap_alloc":           last.MemStats.HeapAlloc,
		"heap_sys":             last.MemStats.HeapSys,
		"heap_idle":            last.MemStats.HeapIdle,
		"stack_inuse":          last.MemStats.StackInuse,
		"mspan_inuse":          last.MemStats.MSpanInuse,
		"mcache_inuse":         last.MemStats.MCacheInuse,
	}
}

func calculateAvg(samples []RuntimeSnapshot, fn func(RuntimeSnapshot) uint64) uint64 {
	var total uint64
	for _, s := range samples {
		total += fn(s)
	}
	return total / uint64(len(samples))
}

func calculateMax(samples []RuntimeSnapshot, fn func(RuntimeSnapshot) uint64) uint64 {
	var max uint64
	for _, s := range samples {
		if v := fn(s); v > max {
			max = v
		}
	}
	return max
}

// ===== 4. Scheduler Contention Analyzer =====

type ContentionAnalyzer struct {
	contentionPoints map[string]*ContentionInfo
	mu               sync.RWMutex
}

type ContentionInfo struct {
	Name             string
	WaitTime         time.Duration
	ContentionCount  int64
	MaxWaitTime      time.Duration
	TotalWaitTime    time.Duration
	LastContentionAt time.Time
}

func NewContentionAnalyzer() *ContentionAnalyzer {
	return &ContentionAnalyzer{
		contentionPoints: make(map[string]*ContentionInfo),
	}
}

func (ca *ContentionAnalyzer) RecordContention(name string, duration time.Duration) {
	ca.mu.Lock()
	defer ca.mu.Unlock()

	info, exists := ca.contentionPoints[name]
	if !exists {
		info = &ContentionInfo{Name: name}
		ca.contentionPoints[name] = info
	}

	info.ContentionCount++
	info.TotalWaitTime += duration
	info.LastContentionAt = time.Now()
	if duration > info.MaxWaitTime {
		info.MaxWaitTime = duration
	}
}

func (ca *ContentionAnalyzer) GetReport() map[string]map[string]interface{} {
	ca.mu.RLock()
	defer ca.mu.RUnlock()

	report := make(map[string]map[string]interface{})
	for name, info := range ca.contentionPoints {
		avgWait := time.Duration(0)
		if info.ContentionCount > 0 {
			avgWait = time.Duration(info.TotalWaitTime.Nanoseconds() / info.ContentionCount)
		}

		report[name] = map[string]interface{}{
			"contention_count": info.ContentionCount,
			"avg_wait_time_us": avgWait.Microseconds(),
			"max_wait_time_us": info.MaxWaitTime.Microseconds(),
			"total_wait_time_ms": info.TotalWaitTime.Milliseconds(),
		}
	}
	return report
}

// ===== 5. GOMAXPROCS Optimizer =====

type MAXPROCSOptimizer struct {
	benchmarks map[int]time.Duration
	mu         sync.RWMutex
}

func NewMAXPROCSOptimizer() *MAXPROCSOptimizer {
	return &MAXPROCSOptimizer{
		benchmarks: make(map[int]time.Duration),
	}
}

func (mo *MAXPROCSOptimizer) BenchmarkWithProcs(maxProcs int, workload func()) time.Duration {
	oldProcs := runtime.GOMAXPROCS(maxProcs)
	defer runtime.GOMAXPROCS(oldProcs)

	start := time.Now()
	workload()
	duration := time.Since(start)

	mo.mu.Lock()
	mo.benchmarks[maxProcs] = duration
	mo.mu.Unlock()

	return duration
}

func (mo *MAXPROCSOptimizer) FindOptimal(maxProcs int, workload func()) int {
	optimalProcs := 1
	var minDuration time.Duration = time.Hour

	for p := 1; p <= maxProcs; p++ {
		duration := mo.BenchmarkWithProcs(p, workload)
		if duration < minDuration {
			minDuration = duration
			optimalProcs = p
		}
	}

	return optimalProcs
}

func (mo *MAXPROCSOptimizer) GetBenchmarks() map[int]time.Duration {
	mo.mu.RLock()
	defer mo.mu.RUnlock()

	result := make(map[int]time.Duration)
	for k, v := range mo.benchmarks {
		result[k] = v
	}
	return result
}

// ===== 6. Goroutine Affinity Pattern =====

type AffinityScheduler struct {
	workers   []*AffinityWorker
	workQueue chan AffinityWork
	numProcs  int
}

type AffinityWorker struct {
	id      int
	work    chan AffinityWork
	affinity int // CPU affinity (simulated)
}

type AffinityWork struct {
	fn       func()
	affinity int // Preferred processor
	priority int
}

func NewAffinityScheduler(numWorkers int) *AffinityScheduler {
	as := &AffinityScheduler{
		workers:   make([]*AffinityWorker, numWorkers),
		workQueue: make(chan AffinityWork, 1000),
		numProcs:  numWorkers,
	}

	for i := 0; i < numWorkers; i++ {
		w := &AffinityWorker{
			id:       i,
			work:     make(chan AffinityWork, 100),
			affinity: i,
		}
		as.workers[i] = w

		go func(worker *AffinityWorker) {
			for work := range worker.work {
				work.fn()
			}
		}(w)
	}

	return as
}

func (as *AffinityScheduler) SubmitWork(work AffinityWork) {
	// Schedule on preferred processor with affinity
	preferredWorker := as.workers[work.affinity%len(as.workers)]
	select {
	case preferredWorker.work <- work:
	default:
		// Fallback to round-robin if preferred is full
		for i := 0; i < len(as.workers); i++ {
			worker := as.workers[(work.affinity+i)%len(as.workers)]
			select {
			case worker.work <- work:
				return
			default:
			}
		}
		// Last resort: global queue
		as.workQueue <- work
	}
}

func (as *AffinityScheduler) Stop() {
	for _, w := range as.workers {
		close(w.work)
	}
}

// ===== Main Demo =====

func main() {
	fmt.Println("=== Go Runtime & Scheduler Internals ===\n")

	// 1. GMP Simulator
	fmt.Println("1. GMP Model Simulator")
	sim := NewGMPSimulator(4)
	workDone := int32(0)

	// Generate work
	for i := 0; i < 100; i++ {
		id := i
		sim.SubmitWork(func() {
			// Simulate work
			sum := 0
			for j := 0; j < 10000; j++ {
				sum += rand.Intn(100)
			}
			_ = sum
			atomic.AddInt32(&workDone, 1)
		}, id%4)
	}

	time.Sleep(500 * time.Millisecond)
	sim.Stop()

	metrics := sim.GetMetrics()
	fmt.Printf("Scheduler Metrics: %v\n", metrics)
	fmt.Printf("Work Done: %d\n\n", workDone)

	// 2. Runtime Statistics
	fmt.Println("2. Runtime Statistics Collector")
	rsc := NewRuntimeStatsCollector(50 * time.Millisecond)
	rsc.Start()

	// Create goroutine load
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 1000000; j++ {
				_ = j % 2
			}
		}()
	}
	wg.Wait()

	rsc.Stop()
	stats := rsc.GetStats()
	fmt.Printf("Runtime Stats: %+v\n\n", stats)

	// 3. Preemption Analysis
	fmt.Println("3. Preemption Analyzer")
	pa := NewPreemptionAnalyzer()
	pa.DemonstrateChanPreemption(100000)
	pa.DemonstrateLoopPreemption()
	fmt.Println("Preemption demonstrations completed\n")

	// 4. Contention Analysis
	fmt.Println("4. Contention Analyzer")
	ca := NewContentionAnalyzer()
	mu := sync.Mutex{}

	for i := 0; i < 100; i++ {
		go func() {
			for j := 0; j < 10; j++ {
				start := time.Now()
				mu.Lock()
				time.Sleep(1 * time.Millisecond)
				mu.Unlock()
				ca.RecordContention("mutex", time.Since(start))
			}
		}()
	}

	time.Sleep(2 * time.Second)
	report := ca.GetReport()
	for name, info := range report {
		fmt.Printf("Contention (%s): %+v\n", name, info)
	}
	fmt.Println()

	// 5. GOMAXPROCS Optimizer
	fmt.Println("5. GOMAXPROCS Optimizer")
	optimizer := NewMAXPROCSOptimizer()
	workload := func() {
		var wg sync.WaitGroup
		for i := 0; i < 1000; i++ {
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
	}

	optimal := optimizer.FindOptimal(runtime.NumCPU(), workload)
	fmt.Printf("Optimal GOMAXPROCS: %d\n", optimal)
	benchmarks := optimizer.GetBenchmarks()
	for procs, duration := range benchmarks {
		fmt.Printf("  Procs=%d: %v\n", procs, duration)
	}
	fmt.Println()

	// 6. Affinity Scheduler
	fmt.Println("6. Goroutine Affinity Pattern")
	affSched := NewAffinityScheduler(4)
	processed := int32(0)

	for i := 0; i < 40; i++ {
		affSched.SubmitWork(AffinityWork{
			fn: func() {
				time.Sleep(time.Millisecond)
				atomic.AddInt32(&processed, 1)
			},
			affinity: i,
			priority: 1,
		})
	}

	time.Sleep(1 * time.Second)
	affSched.Stop()
	fmt.Printf("Affinity-scheduled work processed: %d\n", processed)

	// 7. Runtime Metrics
	fmt.Println("\n7. Live Runtime Metrics")
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	fmt.Printf("Goroutines: %d\n", runtime.NumGoroutine())
	fmt.Printf("Heap Alloc: %v MB\n", m.HeapAlloc/1024/1024)
	fmt.Printf("NumGC: %d\n", m.NumGC)
	fmt.Printf("Pause (last): %v us\n", m.PauseNs[(m.NumGC+255)%256]/1000)

	fmt.Println("\n=== Complete ===")
}
