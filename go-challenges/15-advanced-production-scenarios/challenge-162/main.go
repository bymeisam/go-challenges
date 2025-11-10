package main

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

// ========== Load Testing Models ==========

type LoadTestConfig struct {
	Name              string
	Duration          time.Duration
	RampUpTime        time.Duration
	TargetRPS         int
	MaxConcurrency    int
	Timeout           time.Duration
	FailureThreshold  float64 // percentage
}

type LoadTestResult struct {
	StartTime          time.Time
	EndTime            time.Time
	TotalRequests      int64
	SuccessfulRequests int64
	FailedRequests     int64
	TotalErrors        int64
	TotalLatency       time.Duration
	MinLatency         time.Duration
	MaxLatency         time.Duration
	AvgLatency         time.Duration
	P50Latency         time.Duration
	P95Latency         time.Duration
	P99Latency         time.Duration
	Throughput         float64
	ErrorRate          float64
	Latencies          []time.Duration
}

type RequestResult struct {
	RequestID  string
	Latency    time.Duration
	Success    bool
	Error      error
	Timestamp  time.Time
}

// ========== Chaos Engineering Models ==========

type ChaosScenario struct {
	Name              string
	Description       string
	Duration          time.Duration
	InjectionRate     float64 // percentage of requests
	Faults            []FaultInjection
}

type FaultInjection struct {
	Type      string // "latency", "error", "timeout", "abort"
	Severity  float64
	Duration  time.Duration
	ErrorCode int
}

type ResilienceMetrics struct {
	TotalAttempts    int64
	SuccessfulAttempts int64
	FailedAttempts   int64
	CircuitBreakerTrips int64
	RetryAttempts    int64
	TimeoutOccurrences int64
	RecoveryTime     time.Duration
}

// ========== Load Test Engine ==========

type LoadTestEngine struct {
	config              *LoadTestConfig
	results             *LoadTestResult
	resultsMu           sync.RWMutex
	chaosScenarios      []*ChaosScenario
	activeChaos         *ChaosScenario
	activeChaosEnabled  int32
	metricsCollector    *MetricsCollector
	resilienceMetrics   *ResilienceMetrics
	rateLimiter         *RateLimiter
	circuitBreaker      *CircuitBreaker
}

type MetricsCollector struct {
	latencies    []time.Duration
	latenciesMu  sync.RWMutex
	errors       []error
	errorsMu     sync.RWMutex
	startTime    time.Time
}

type RateLimiter struct {
	maxRPS    int
	mu        sync.Mutex
	lastTime  time.Time
	allowance float64
}

type CircuitBreaker struct {
	state             string // "closed", "open", "half-open"
	failureCount      int
	successCount      int
	failureThreshold  int
	successThreshold  int
	timeout           time.Duration
	lastFailureTime   time.Time
	mu                sync.RWMutex
}

// ========== Load Test Engine Implementation ==========

func NewLoadTestEngine(config *LoadTestConfig) *LoadTestEngine {
	return &LoadTestEngine{
		config: config,
		results: &LoadTestResult{
			Latencies: []time.Duration{},
		},
		metricsCollector: &MetricsCollector{
			latencies: []time.Duration{},
			errors:    []error{},
			startTime: time.Now(),
		},
		resilienceMetrics: &ResilienceMetrics{},
		rateLimiter:       NewRateLimiter(config.TargetRPS),
		circuitBreaker:    NewCircuitBreaker(5, 2, 30*time.Second),
	}
}

// RunLoadTest executes the load test
func (lte *LoadTestEngine) RunLoadTest(ctx context.Context, handler func(context.Context) error) (*LoadTestResult, error) {
	lte.results.StartTime = time.Now()

	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, lte.config.Duration)
	defer cancel()

	// Start workers
	var wg sync.WaitGroup
	resultChan := make(chan *RequestResult, lte.config.MaxConcurrency)

	// Spawn workers
	for i := 0; i < lte.config.MaxConcurrency; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			lte.workerLoop(ctx, workerID, handler, resultChan)
		}(i)
	}

	// Collect results
	resultAggregator := sync.NewCond(&sync.Mutex{})
	go func() {
		for result := range resultChan {
			lte.recordResult(result)
		}
		resultAggregator.Broadcast()
	}()

	// Wait for all workers to finish
	wg.Wait()
	close(resultChan)

	lte.results.EndTime = time.Now()
	lte.finalizeResults()

	return lte.results, nil
}

func (lte *LoadTestEngine) workerLoop(ctx context.Context, workerID int, handler func(context.Context) error, results chan<- *RequestResult) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		// Rate limiting
		lte.rateLimiter.Wait()

		// Check circuit breaker
		if !lte.circuitBreaker.IsHealthy() {
			lte.resilienceMetrics.CircuitBreakerTrips++
			time.Sleep(100 * time.Millisecond)
			continue
		}

		// Execute request
		startTime := time.Now()
		err := handler(ctx)
		latency := time.Since(startTime)

		result := &RequestResult{
			RequestID: fmt.Sprintf("req_%d_%d", workerID, time.Now().UnixNano()),
			Latency:   latency,
			Success:   err == nil,
			Error:     err,
			Timestamp: time.Now(),
		}

		// Update circuit breaker
		if err != nil {
			lte.circuitBreaker.RecordFailure()
			lte.resilienceMetrics.FailedAttempts++
		} else {
			lte.circuitBreaker.RecordSuccess()
			lte.resilienceMetrics.SuccessfulAttempts++
		}

		results <- result
	}
}

func (lte *LoadTestEngine) recordResult(result *RequestResult) {
	lte.resultsMu.Lock()
	defer lte.resultsMu.Unlock()

	lte.results.TotalRequests++
	lte.results.TotalLatency += result.Latency
	lte.results.Latencies = append(lte.results.Latencies, result.Latency)

	if result.Latency < lte.results.MinLatency || lte.results.MinLatency == 0 {
		lte.results.MinLatency = result.Latency
	}
	if result.Latency > lte.results.MaxLatency {
		lte.results.MaxLatency = result.Latency
	}

	if result.Success {
		lte.results.SuccessfulRequests++
	} else {
		lte.results.FailedRequests++
		if result.Error != nil {
			lte.results.TotalErrors++
		}
	}

	lte.metricsCollector.addLatency(result.Latency)
	if result.Error != nil {
		lte.metricsCollector.addError(result.Error)
	}
}

func (lte *LoadTestEngine) finalizeResults() {
	lte.resultsMu.Lock()
	defer lte.resultsMu.Unlock()

	duration := lte.results.EndTime.Sub(lte.results.StartTime).Seconds()

	if lte.results.TotalRequests > 0 {
		lte.results.AvgLatency = time.Duration(int64(lte.results.TotalLatency) / lte.results.TotalRequests)
		lte.results.ErrorRate = float64(lte.results.FailedRequests) / float64(lte.results.TotalRequests) * 100
	}

	lte.results.Throughput = float64(lte.results.SuccessfulRequests) / duration

	// Calculate percentiles
	if len(lte.results.Latencies) > 0 {
		sort.Slice(lte.results.Latencies, func(i, j int) bool {
			return lte.results.Latencies[i] < lte.results.Latencies[j]
		})

		lte.results.P50Latency = lte.results.Latencies[len(lte.results.Latencies)/2]
		lte.results.P95Latency = lte.results.Latencies[int(float64(len(lte.results.Latencies))*0.95)]
		lte.results.P99Latency = lte.results.Latencies[int(float64(len(lte.results.Latencies))*0.99)]
	}
}

// ========== Chaos Engineering ==========

// InjectChaos injects faults into the system
func (lte *LoadTestEngine) InjectChaos(scenario *ChaosScenario) {
	lte.activeChaos = scenario
	atomic.StoreInt32(&lte.activeChaosEnabled, 1)
}

// StopChaos stops chaos injection
func (lte *LoadTestEngine) StopChaos() {
	atomic.StoreInt32(&lte.activeChaosEnabled, 0)
	lte.activeChaos = nil
}

// ShouldInjectFault returns whether a fault should be injected
func (lte *LoadTestEngine) ShouldInjectFault() bool {
	if atomic.LoadInt32(&lte.activeChaosEnabled) == 0 {
		return false
	}

	return time.Since(time.Now().Add(-lte.activeChaos.Duration)) < lte.activeChaos.Duration
}

// ========== Rate Limiter ==========

func NewRateLimiter(maxRPS int) *RateLimiter {
	return &RateLimiter{
		maxRPS:    maxRPS,
		allowance: float64(maxRPS),
		lastTime:  time.Now(),
	}
}

func (rl *RateLimiter) Wait() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(rl.lastTime).Seconds()
	rl.lastTime = now

	rl.allowance += elapsed * float64(rl.maxRPS)
	if rl.allowance > float64(rl.maxRPS) {
		rl.allowance = float64(rl.maxRPS)
	}

	if rl.allowance < 1.0 {
		time.Sleep(time.Duration((1.0 - rl.allowance) / float64(rl.maxRPS) * 1e9) * time.Nanosecond)
	}

	rl.allowance--
}

// ========== Circuit Breaker ==========

func NewCircuitBreaker(failureThreshold, successThreshold int, timeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		state:            "closed",
		failureThreshold: failureThreshold,
		successThreshold: successThreshold,
		timeout:          timeout,
	}
}

func (cb *CircuitBreaker) IsHealthy() bool {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	return cb.state == "closed" || (cb.state == "half-open")
}

func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failureCount++
	cb.lastFailureTime = time.Now()
	cb.successCount = 0

	if cb.failureCount >= cb.failureThreshold && cb.state == "closed" {
		cb.state = "open"
	}
}

func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.successCount++
	cb.failureCount = 0

	if cb.state == "half-open" && cb.successCount >= cb.successThreshold {
		cb.state = "closed"
	}
}

// ========== Metrics Collector ==========

func (mc *MetricsCollector) addLatency(latency time.Duration) {
	mc.latenciesMu.Lock()
	defer mc.latenciesMu.Unlock()
	mc.latencies = append(mc.latencies, latency)
}

func (mc *MetricsCollector) addError(err error) {
	mc.errorsMu.Lock()
	defer mc.errorsMu.Unlock()
	mc.errors = append(mc.errors, err)
}

func (mc *MetricsCollector) GetMetrics() map[string]interface{} {
	mc.latenciesMu.RLock()
	latencyCount := len(mc.latencies)
	mc.latenciesMu.RUnlock()

	mc.errorsMu.RLock()
	errorCount := len(mc.errors)
	mc.errorsMu.RUnlock()

	return map[string]interface{}{
		"total_measurements": latencyCount,
		"total_errors":       errorCount,
		"collection_time":    time.Since(mc.startTime),
	}
}

// ========== Helper Functions ==========

func (result *LoadTestResult) GetSummary() map[string]interface{} {
	return map[string]interface{}{
		"total_requests":      result.TotalRequests,
		"successful_requests": result.SuccessfulRequests,
		"failed_requests":     result.FailedRequests,
		"error_rate":          fmt.Sprintf("%.2f%%", result.ErrorRate),
		"throughput":          fmt.Sprintf("%.2f req/s", result.Throughput),
		"avg_latency":         result.AvgLatency.String(),
		"min_latency":         result.MinLatency.String(),
		"max_latency":         result.MaxLatency.String(),
		"p50_latency":         result.P50Latency.String(),
		"p95_latency":         result.P95Latency.String(),
		"p99_latency":         result.P99Latency.String(),
	}
}

func main() {
	// Example load test
	config := &LoadTestConfig{
		Name:           "Example Load Test",
		Duration:       1 * time.Second,
		TargetRPS:      10,
		MaxConcurrency: 2,
		Timeout:        5 * time.Second,
	}

	engine := NewLoadTestEngine(config)
	_, _ = engine.RunLoadTest(context.Background(), func(ctx context.Context) error {
		return nil
	})
}
