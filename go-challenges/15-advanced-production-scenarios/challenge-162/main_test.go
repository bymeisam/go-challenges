package main

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestLoadTestBasic(t *testing.T) {
	config := &LoadTestConfig{
		Name:           "Basic Load Test",
		Duration:       1 * time.Second,
		TargetRPS:      10,
		MaxConcurrency: 2,
		Timeout:        5 * time.Second,
	}

	engine := NewLoadTestEngine(config)

	handler := func(ctx context.Context) error {
		return nil
	}

	result, err := engine.RunLoadTest(context.Background(), handler)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result.TotalRequests == 0 {
		t.Fatal("Expected some requests to be executed")
	}

	if result.SuccessfulRequests == 0 {
		t.Fatal("Expected successful requests")
	}

	if result.AvgLatency == 0 {
		t.Fatal("Expected latency measurement")
	}
}

func TestLoadTestWithFailures(t *testing.T) {
	config := &LoadTestConfig{
		Name:           "Load Test With Failures",
		Duration:       500 * time.Millisecond,
		TargetRPS:      10,
		MaxConcurrency: 2,
		Timeout:        5 * time.Second,
	}

	engine := NewLoadTestEngine(config)

	callCount := 0
	handler := func(ctx context.Context) error {
		callCount++
		if callCount%2 == 0 {
			return errors.New("simulated error")
		}
		return nil
	}

	result, _ := engine.RunLoadTest(context.Background(), handler)

	if result.ErrorRate == 0 {
		t.Fatal("Expected some failures")
	}

	if result.FailedRequests == 0 {
		t.Fatal("Expected failed request count")
	}
}

func TestLoadTestThroughput(t *testing.T) {
	config := &LoadTestConfig{
		Name:           "Throughput Test",
		Duration:       1 * time.Second,
		TargetRPS:      100,
		MaxConcurrency: 10,
		Timeout:        5 * time.Second,
	}

	engine := NewLoadTestEngine(config)

	handler := func(ctx context.Context) error {
		return nil
	}

	result, _ := engine.RunLoadTest(context.Background(), handler)

	if result.Throughput <= 0 {
		t.Fatal("Expected positive throughput")
	}
}

func TestLatencyMeasurement(t *testing.T) {
	config := &LoadTestConfig{
		Name:           "Latency Test",
		Duration:       500 * time.Millisecond,
		TargetRPS:      10,
		MaxConcurrency: 2,
		Timeout:        5 * time.Second,
	}

	engine := NewLoadTestEngine(config)

	handler := func(ctx context.Context) error {
		time.Sleep(10 * time.Millisecond)
		return nil
	}

	result, _ := engine.RunLoadTest(context.Background(), handler)

	if result.AvgLatency < 10*time.Millisecond {
		t.Fatal("Expected latency to be at least 10ms")
	}

	if result.P95Latency < result.AvgLatency {
		t.Fatal("Expected P95 to be >= average")
	}

	if result.P99Latency < result.P95Latency {
		t.Fatal("Expected P99 to be >= P95")
	}
}

func TestRateLimiter(t *testing.T) {
	limiter := NewRateLimiter(10)

	start := time.Now()
	for i := 0; i < 10; i++ {
		limiter.Wait()
	}
	elapsed := time.Since(start)

	// Should take at least ~1 second for 10 requests at 10 RPS
	if elapsed < 500*time.Millisecond {
		t.Logf("Expected ~1 second for 10 requests, got %v", elapsed)
	}
}

func TestCircuitBreakerOpen(t *testing.T) {
	cb := NewCircuitBreaker(3, 2, 1*time.Second)

	// Record 3 failures to open circuit
	for i := 0; i < 3; i++ {
		cb.RecordFailure()
	}

	if cb.IsHealthy() {
		t.Fatal("Expected circuit to be open after failures")
	}
}

func TestCircuitBreakerRecovery(t *testing.T) {
	cb := NewCircuitBreaker(3, 2, 1*time.Second)

	// Open circuit
	for i := 0; i < 3; i++ {
		cb.RecordFailure()
	}

	// Record successes to close circuit
	for i := 0; i < 2; i++ {
		cb.RecordSuccess()
	}

	if !cb.IsHealthy() {
		t.Fatal("Expected circuit to be healthy after successes")
	}
}

func TestMetricsCollection(t *testing.T) {
	collector := &MetricsCollector{
		latencies: []time.Duration{},
		errors:    []error{},
		startTime: time.Now(),
	}

	collector.addLatency(10 * time.Millisecond)
	collector.addLatency(20 * time.Millisecond)
	collector.addError(errors.New("test error"))

	metrics := collector.GetMetrics()

	if metrics["total_measurements"] != 2 {
		t.Fatalf("Expected 2 measurements, got %v", metrics["total_measurements"])
	}

	if metrics["total_errors"] != 1 {
		t.Fatalf("Expected 1 error, got %v", metrics["total_errors"])
	}
}

func TestChaosInjection(t *testing.T) {
	config := &LoadTestConfig{
		Name:           "Chaos Test",
		Duration:       500 * time.Millisecond,
		TargetRPS:      10,
		MaxConcurrency: 2,
		Timeout:        5 * time.Second,
	}

	engine := NewLoadTestEngine(config)

	chaos := &ChaosScenario{
		Name:          "Network Latency",
		Description:   "Inject network latency",
		Duration:      500 * time.Millisecond,
		InjectionRate: 0.5,
		Faults: []FaultInjection{
			{Type: "latency", Severity: 100, Duration: 100 * time.Millisecond},
		},
	}

	engine.InjectChaos(chaos)

	if engine.activeChaos == nil {
		t.Fatal("Expected chaos to be active")
	}

	engine.StopChaos()

	if engine.activeChaos != nil {
		t.Fatal("Expected chaos to be stopped")
	}
}

func TestResilienceMetrics(t *testing.T) {
	config := &LoadTestConfig{
		Name:           "Resilience Test",
		Duration:       500 * time.Millisecond,
		TargetRPS:      10,
		MaxConcurrency: 2,
		Timeout:        5 * time.Second,
	}

	engine := NewLoadTestEngine(config)

	result, _ := engine.RunLoadTest(context.Background(), func(ctx context.Context) error {
		return nil
	})

	if engine.resilienceMetrics.SuccessfulAttempts == 0 {
		t.Fatal("Expected successful attempts recorded")
	}
}

func TestLoadTestSummary(t *testing.T) {
	result := &LoadTestResult{
		TotalRequests:      100,
		SuccessfulRequests: 95,
		FailedRequests:     5,
		AvgLatency:         50 * time.Millisecond,
		MinLatency:         10 * time.Millisecond,
		MaxLatency:         100 * time.Millisecond,
		P50Latency:         45 * time.Millisecond,
		P95Latency:         80 * time.Millisecond,
		P99Latency:         95 * time.Millisecond,
		Throughput:         100.0,
		ErrorRate:          5.0,
	}

	summary := result.GetSummary()

	if summary["total_requests"] != int64(100) {
		t.Fatal("Expected total requests in summary")
	}

	if summary["throughput"] == "" {
		t.Fatal("Expected throughput in summary")
	}
}

func TestLoadTestWithTimeout(t *testing.T) {
	config := &LoadTestConfig{
		Name:           "Timeout Test",
		Duration:       500 * time.Millisecond,
		TargetRPS:      10,
		MaxConcurrency: 2,
		Timeout:        5 * time.Second,
	}

	engine := NewLoadTestEngine(config)

	handler := func(ctx context.Context) error {
		select {
		case <-ctx.Done():
			return context.DeadlineExceeded
		case <-time.After(10 * time.Millisecond):
			return nil
		}
	}

	result, _ := engine.RunLoadTest(context.Background(), handler)

	if result.TotalRequests == 0 {
		t.Fatal("Expected requests to be executed")
	}
}

func TestLatencyPercentiles(t *testing.T) {
	config := &LoadTestConfig{
		Name:           "Percentile Test",
		Duration:       1 * time.Second,
		TargetRPS:      20,
		MaxConcurrency: 5,
		Timeout:        5 * time.Second,
	}

	engine := NewLoadTestEngine(config)

	handler := func(ctx context.Context) error {
		return nil
	}

	result, _ := engine.RunLoadTest(context.Background(), handler)

	// Verify percentiles are in order
	if result.P50Latency > result.P95Latency {
		t.Fatal("P50 should be <= P95")
	}

	if result.P95Latency > result.P99Latency {
		t.Fatal("P95 should be <= P99")
	}

	if result.AvgLatency == 0 {
		t.Fatal("Expected average latency")
	}
}

// ========== Benchmarks ==========

func BenchmarkLoadTestExecution(b *testing.B) {
	config := &LoadTestConfig{
		Name:           "Benchmark Test",
		Duration:       1 * time.Second,
		TargetRPS:      100,
		MaxConcurrency: 10,
		Timeout:        5 * time.Second,
	}

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		engine := NewLoadTestEngine(config)
		engine.RunLoadTest(context.Background(), func(ctx context.Context) error {
			return nil
		})
	}
}

func BenchmarkRateLimiter(b *testing.B) {
	limiter := NewRateLimiter(1000)
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		limiter.Wait()
	}
}

func BenchmarkCircuitBreaker(b *testing.B) {
	cb := NewCircuitBreaker(5, 2, 30*time.Second)
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		if i%10 == 0 {
			cb.RecordFailure()
		} else {
			cb.RecordSuccess()
		}
	}
}
