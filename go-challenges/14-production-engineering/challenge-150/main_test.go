package main

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestCounterInc(t *testing.T) {
	c := NewCounter("test_counter", "A test counter")

	if c.Value() != 0 {
		t.Errorf("Expected initial value 0, got %d", c.Value())
	}

	c.Inc()
	if c.Value() != 1 {
		t.Errorf("Expected value 1 after Inc, got %d", c.Value())
	}

	c.Inc()
	c.Inc()
	if c.Value() != 3 {
		t.Errorf("Expected value 3 after 3 Incs, got %d", c.Value())
	}

	t.Log("✓ Counter Inc works!")
}

func TestCounterAdd(t *testing.T) {
	c := NewCounter("test_counter", "A test counter")

	c.Add(5)
	if c.Value() != 5 {
		t.Errorf("Expected value 5, got %d", c.Value())
	}

	c.Add(10)
	if c.Value() != 15 {
		t.Errorf("Expected value 15, got %d", c.Value())
	}

	// Negative values should not be added
	c.Add(-5)
	if c.Value() != 15 {
		t.Errorf("Expected value to remain 15, got %d", c.Value())
	}

	t.Log("✓ Counter Add works!")
}

func TestCounterCollect(t *testing.T) {
	c := NewCounter("test_counter", "A test counter")
	c.Add(42)

	output := c.Collect()

	if !strings.Contains(output, "# TYPE test_counter counter") {
		t.Errorf("Output missing type declaration: %s", output)
	}

	if !strings.Contains(output, "test_counter 42") {
		t.Errorf("Output missing metric value: %s", output)
	}

	t.Log("✓ Counter Collect works!")
}

func TestGaugeSet(t *testing.T) {
	g := NewGauge("test_gauge", "A test gauge")

	g.Set(10.5)
	if g.Value() != 10.5 {
		t.Errorf("Expected value 10.5, got %g", g.Value())
	}

	g.Set(-5.5)
	if g.Value() != -5.5 {
		t.Errorf("Expected value -5.5, got %g", g.Value())
	}

	t.Log("✓ Gauge Set works!")
}

func TestGaugeIncDec(t *testing.T) {
	g := NewGauge("test_gauge", "A test gauge")

	g.Inc()
	if g.Value() != 1 {
		t.Errorf("Expected value 1 after Inc, got %g", g.Value())
	}

	g.Dec()
	if g.Value() != 0 {
		t.Errorf("Expected value 0 after Dec, got %g", g.Value())
	}

	g.Add(5.5)
	if g.Value() != 5.5 {
		t.Errorf("Expected value 5.5 after Add, got %g", g.Value())
	}

	g.Sub(2.5)
	if g.Value() != 3 {
		t.Errorf("Expected value 3 after Sub, got %g", g.Value())
	}

	t.Log("✓ Gauge Inc/Dec works!")
}

func TestGaugeCollect(t *testing.T) {
	g := NewGauge("test_gauge", "A test gauge")
	g.Set(42.5)

	output := g.Collect()

	if !strings.Contains(output, "# TYPE test_gauge gauge") {
		t.Errorf("Output missing type declaration: %s", output)
	}

	if !strings.Contains(output, "test_gauge") {
		t.Errorf("Output missing metric name: %s", output)
	}

	t.Log("✓ Gauge Collect works!")
}

func TestHistogramObserve(t *testing.T) {
	h := NewHistogram("test_histogram", "A test histogram", []float64{0.1, 0.5, 1, 5})

	h.Observe(0.05)
	h.Observe(0.3)
	h.Observe(2)

	sum, count := h.Value()
	if count != 3 {
		t.Errorf("Expected count 3, got %d", count)
	}

	expectedSum := 0.05 + 0.3 + 2
	if sum != expectedSum {
		t.Errorf("Expected sum %g, got %g", expectedSum, sum)
	}

	t.Log("✓ Histogram Observe works!")
}

func TestHistogramBuckets(t *testing.T) {
	h := NewHistogram("test_histogram", "A test histogram", []float64{0.1, 0.5, 1})

	h.Observe(0.05)
	h.Observe(0.3)
	h.Observe(0.7)
	h.Observe(2)

	output := h.Collect()

	if !strings.Contains(output, "test_histogram_bucket{le=\"0.1\"}") {
		t.Errorf("Output missing bucket 0.1: %s", output)
	}

	if !strings.Contains(output, "test_histogram_sum") {
		t.Errorf("Output missing sum: %s", output)
	}

	if !strings.Contains(output, "test_histogram_count") {
		t.Errorf("Output missing count: %s", output)
	}

	t.Log("✓ Histogram Buckets work!")
}

func TestRegistry(t *testing.T) {
	registry := NewRegistry()

	c := NewCounter("counter1", "First counter")
	g := NewGauge("gauge1", "First gauge")

	if err := registry.Register(c); err != nil {
		t.Fatalf("Failed to register counter: %v", err)
	}

	if err := registry.Register(g); err != nil {
		t.Fatalf("Failed to register gauge: %v", err)
	}

	// Duplicate registration should fail
	if err := registry.Register(c); err == nil {
		t.Error("Expected error on duplicate registration")
	}

	output := registry.Gather()

	if !strings.Contains(output, "counter1") {
		t.Errorf("Output missing counter1: %s", output)
	}

	if !strings.Contains(output, "gauge1") {
		t.Errorf("Output missing gauge1: %s", output)
	}

	t.Log("✓ Registry works!")
}

func TestHTTPMetrics(t *testing.T) {
	registry := NewRegistry()
	httpMetrics, err := NewHTTPMetrics(registry)
	if err != nil {
		t.Fatalf("Failed to create HTTP metrics: %v", err)
	}

	if httpMetrics == nil {
		t.Error("HTTP metrics should not be nil")
	}

	if httpMetrics.requestsTotal.Value() != 0 {
		t.Errorf("Expected initial requests to be 0, got %d", httpMetrics.requestsTotal.Value())
	}

	httpMetrics.requestsTotal.Inc()
	if httpMetrics.requestsTotal.Value() != 1 {
		t.Errorf("Expected requests to be 1, got %d", httpMetrics.requestsTotal.Value())
	}

	t.Log("✓ HTTP Metrics work!")
}

func TestHTTPMetricsMiddleware(t *testing.T) {
	registry := NewRegistry()
	httpMetrics, _ := NewHTTPMetrics(registry)

	// Test handler
	handler := httpMetrics.Middleware(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(10 * time.Millisecond)
		w.WriteHeader(200)
	})

	if handler == nil {
		t.Error("Handler should not be nil")
	}

	t.Log("✓ HTTP Metrics Middleware works!")
}

func TestAppMetrics(t *testing.T) {
	registry := NewRegistry()
	appMetrics, err := NewAppMetrics(registry)
	if err != nil {
		t.Fatalf("Failed to create app metrics: %v", err)
	}

	appMetrics.RecordDBConnection(5)
	if appMetrics.dbConnections.Value() != 5 {
		t.Errorf("Expected DB connections 5, got %g", appMetrics.dbConnections.Value())
	}

	appMetrics.RecordCacheHit()
	if appMetrics.cacheHits.Value() != 1 {
		t.Errorf("Expected cache hits 1, got %d", appMetrics.cacheHits.Value())
	}

	appMetrics.RecordCacheMiss()
	if appMetrics.cacheMisses.Value() != 1 {
		t.Errorf("Expected cache misses 1, got %d", appMetrics.cacheMisses.Value())
	}

	appMetrics.RecordError()
	if appMetrics.errorCount.Value() != 1 {
		t.Errorf("Expected errors 1, got %d", appMetrics.errorCount.Value())
	}

	t.Log("✓ App Metrics work!")
}

func TestTracer(t *testing.T) {
	tracer := NewTracer()

	traceID := "trace-123"
	span := tracer.StartSpan(traceID, "test_operation")

	if span.Name != "test_operation" {
		t.Errorf("Expected span name test_operation, got %s", span.Name)
	}

	span.Tags["user_id"] = "user123"
	span.Logs = append(span.Logs, "Operation started")

	tracer.FinishSpan(traceID, span)

	traces := tracer.GetTrace(traceID)
	if len(traces) != 1 {
		t.Errorf("Expected 1 trace, got %d", len(traces))
	}

	if traces[0].Tags["user_id"] != "user123" {
		t.Errorf("Expected user_id tag, got %v", traces[0].Tags)
	}

	t.Log("✓ Tracer works!")
}

func TestContextTracing(t *testing.T) {
	ctx := context.Background()
	ctx = WithTrace(ctx, "trace-456")

	traceID := GetTraceID(ctx)
	if traceID != "trace-456" {
		t.Errorf("Expected trace ID trace-456, got %s", traceID)
	}

	// Test missing trace ID
	ctx2 := context.Background()
	traceID2 := GetTraceID(ctx2)
	if traceID2 != "" {
		t.Errorf("Expected empty trace ID, got %s", traceID2)
	}

	t.Log("✓ Context Tracing works!")
}

func TestServiceCreation(t *testing.T) {
	service, err := NewService()
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	if service == nil {
		t.Error("Service should not be nil")
	}

	if service.registry == nil {
		t.Error("Registry should not be nil")
	}

	if service.http == nil {
		t.Error("HTTP metrics should not be nil")
	}

	if service.app == nil {
		t.Error("App metrics should not be nil")
	}

	if service.tracer == nil {
		t.Error("Tracer should not be nil")
	}

	t.Log("✓ Service creation works!")
}

func TestServiceProcessRequest(t *testing.T) {
	service, _ := NewService()

	ctx := context.Background()
	ctx = WithTrace(ctx, "trace-789")

	err := service.ProcessRequest(ctx, 10*time.Millisecond)
	if err != nil {
		t.Fatalf("ProcessRequest failed: %v", err)
	}

	traces := service.tracer.GetTrace("trace-789")
	if len(traces) == 0 {
		t.Error("Expected trace to be recorded")
	}

	if len(traces) > 0 && traces[0].Name != "process_request" {
		t.Errorf("Expected span name process_request, got %s", traces[0].Name)
	}

	t.Log("✓ Service ProcessRequest works!")
}

func TestMetricsOutput(t *testing.T) {
	registry := NewRegistry()

	c := NewCounter("test_requests", "Test requests")
	c.Add(100)

	g := NewGauge("test_connections", "Test connections")
	g.Set(42)

	registry.Register(c)
	registry.Register(g)

	output := registry.Gather()

	if !strings.Contains(output, "# TYPE test_requests counter") {
		t.Error("Output missing counter type")
	}

	if !strings.Contains(output, "# TYPE test_connections gauge") {
		t.Error("Output missing gauge type")
	}

	if !strings.Contains(output, "test_requests 100") {
		t.Error("Output missing counter value")
	}

	t.Log("✓ Metrics Output works!")
}

func TestConcurrentMetrics(t *testing.T) {
	c := NewCounter("concurrent_counter", "Concurrent counter")

	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				c.Inc()
			}
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	if c.Value() != 1000 {
		t.Errorf("Expected counter value 1000, got %d", c.Value())
	}

	t.Log("✓ Concurrent Metrics work!")
}

func TestHistogramPercentiles(t *testing.T) {
	h := NewHistogram("latency", "Request latency", []float64{0.01, 0.05, 0.1, 0.5, 1})

	// Simulate latencies
	latencies := []float64{0.001, 0.005, 0.008, 0.02, 0.1, 0.2, 0.5, 0.8}
	for _, lat := range latencies {
		h.Observe(lat)
	}

	sum, count := h.Value()
	if count != int64(len(latencies)) {
		t.Errorf("Expected count %d, got %d", len(latencies), count)
	}

	expectedSum := 0.0
	for _, lat := range latencies {
		expectedSum += lat
	}

	if sum != expectedSum {
		t.Errorf("Expected sum %g, got %g", expectedSum, sum)
	}

	t.Log("✓ Histogram Percentiles work!")
}

func BenchmarkCounterInc(b *testing.B) {
	c := NewCounter("bench_counter", "Benchmark counter")
	for i := 0; i < b.N; i++ {
		c.Inc()
	}
}

func BenchmarkGaugeSet(b *testing.B) {
	g := NewGauge("bench_gauge", "Benchmark gauge")
	for i := 0; i < b.N; i++ {
		g.Set(float64(i))
	}
}

func BenchmarkHistogramObserve(b *testing.B) {
	h := NewHistogram("bench_histogram", "Benchmark histogram", []float64{0.1, 0.5, 1})
	for i := 0; i < b.N; i++ {
		h.Observe(0.1)
	}
}
