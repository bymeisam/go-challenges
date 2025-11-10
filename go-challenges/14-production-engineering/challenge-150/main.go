package main

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"sync"
	"time"
)

// MetricType represents the type of metric
type MetricType string

const (
	MetricTypeCounter   MetricType = "counter"
	MetricTypeGauge     MetricType = "gauge"
	MetricTypeHistogram MetricType = "histogram"
)

// Metric interface defines all metrics
type Metric interface {
	Name() string
	Type() MetricType
	Help() string
	Collect() string
}

// ========== Counter ==========

// Counter represents a monotonically increasing metric
type Counter struct {
	name  string
	help  string
	value int64
	mu    sync.RWMutex
}

// NewCounter creates a new counter metric
func NewCounter(name, help string) *Counter {
	return &Counter{
		name: name,
		help: help,
		value: 0,
	}
}

func (c *Counter) Name() string       { return c.name }
func (c *Counter) Type() MetricType   { return MetricTypeCounter }
func (c *Counter) Help() string       { return c.help }

// Inc increments the counter by 1
func (c *Counter) Inc() {
	c.Add(1)
}

// Add adds a value to the counter
func (c *Counter) Add(value int64) {
	if value < 0 {
		return // Counters can't go down
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.value += value
}

// Value returns current counter value
func (c *Counter) Value() int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.value
}

// Collect returns Prometheus format output
func (c *Counter) Collect() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return fmt.Sprintf("# TYPE %s counter\n# HELP %s %s\n%s %d\n",
		c.name, c.name, c.help, c.name, c.value)
}

// ========== Gauge ==========

// Gauge represents a metric that can go up and down
type Gauge struct {
	name  string
	help  string
	value float64
	mu    sync.RWMutex
}

// NewGauge creates a new gauge metric
func NewGauge(name, help string) *Gauge {
	return &Gauge{
		name: name,
		help: help,
		value: 0,
	}
}

func (g *Gauge) Name() string       { return g.name }
func (g *Gauge) Type() MetricType   { return MetricTypeGauge }
func (g *Gauge) Help() string       { return g.help }

// Set sets the gauge to a specific value
func (g *Gauge) Set(value float64) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.value = value
}

// Inc increments gauge by 1
func (g *Gauge) Inc() {
	g.Add(1)
}

// Dec decrements gauge by 1
func (g *Gauge) Dec() {
	g.Sub(1)
}

// Add adds a value to the gauge
func (g *Gauge) Add(value float64) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.value += value
}

// Sub subtracts a value from the gauge
func (g *Gauge) Sub(value float64) {
	g.Add(-value)
}

// Value returns current gauge value
func (g *Gauge) Value() float64 {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.value
}

// Collect returns Prometheus format output
func (g *Gauge) Collect() string {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return fmt.Sprintf("# TYPE %s gauge\n# HELP %s %s\n%s %g\n",
		g.name, g.name, g.help, g.name, g.value)
}

// ========== Histogram ==========

// HistogramBucket represents a bucket in a histogram
type HistogramBucket struct {
	Le    float64 // Less than or equal
	Count int64
}

// Histogram represents distribution of values
type Histogram struct {
	name    string
	help    string
	buckets []float64
	counts  map[float64]int64
	sum     float64
	count   int64
	mu      sync.RWMutex
}

// NewHistogram creates a new histogram with specified buckets
func NewHistogram(name, help string, buckets []float64) *Histogram {
	// Add +Inf bucket
	buckets = append(buckets, math.Inf(1))

	// Initialize counts
	counts := make(map[float64]int64)
	for _, b := range buckets {
		counts[b] = 0
	}

	return &Histogram{
		name:    name,
		help:    help,
		buckets: buckets,
		counts:  counts,
	}
}

func (h *Histogram) Name() string       { return h.name }
func (h *Histogram) Type() MetricType   { return MetricTypeHistogram }
func (h *Histogram) Help() string       { return h.help }

// Observe records a value in the histogram
func (h *Histogram) Observe(value float64) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.sum += value
	h.count++

	// Update bucket counts
	for _, bucket := range h.buckets {
		if value <= bucket {
			h.counts[bucket]++
		}
	}
}

// Value returns summary statistics
func (h *Histogram) Value() (sum float64, count int64) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.sum, h.count
}

// Collect returns Prometheus format output
func (h *Histogram) Collect() string {
	h.mu.RLock()
	defer h.mu.RUnlock()

	output := fmt.Sprintf("# TYPE %s histogram\n# HELP %s %s\n",
		h.name, h.name, h.help)

	for _, bucket := range h.buckets {
		output += fmt.Sprintf("%s_bucket{le=\"%g\"} %d\n",
			h.name, bucket, h.counts[bucket])
	}

	output += fmt.Sprintf("%s_sum %g\n", h.name, h.sum)
	output += fmt.Sprintf("%s_count %d\n", h.name, h.count)

	return output
}

// ========== Registry ==========

// Registry manages all metrics
type Registry struct {
	metrics map[string]Metric
	mu      sync.RWMutex
}

// NewRegistry creates a new metric registry
func NewRegistry() *Registry {
	return &Registry{
		metrics: make(map[string]Metric),
	}
}

// Register registers a metric in the registry
func (r *Registry) Register(metric Metric) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.metrics[metric.Name()]; exists {
		return fmt.Errorf("metric %s already registered", metric.Name())
	}

	r.metrics[metric.Name()] = metric
	return nil
}

// Gather collects all metrics in Prometheus format
func (r *Registry) Gather() string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	output := ""
	for _, metric := range r.metrics {
		output += metric.Collect()
	}
	return output
}

// ========== HTTP Instrumentation ==========

// HTTPMetrics holds HTTP-related metrics
type HTTPMetrics struct {
	requestsTotal   *Counter
	requestDuration *Histogram
	activeRequests  *Gauge
}

// NewHTTPMetrics creates HTTP metrics and registers them
func NewHTTPMetrics(registry *Registry) (*HTTPMetrics, error) {
	m := &HTTPMetrics{
		requestsTotal: NewCounter(
			"http_requests_total",
			"Total number of HTTP requests",
		),
		requestDuration: NewHistogram(
			"http_request_duration_seconds",
			"HTTP request duration in seconds",
			[]float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1},
		),
		activeRequests: NewGauge(
			"http_active_requests",
			"Number of active HTTP requests",
		),
	}

	if err := registry.Register(m.requestsTotal); err != nil {
		return nil, err
	}
	if err := registry.Register(m.requestDuration); err != nil {
		return nil, err
	}
	if err := registry.Register(m.activeRequests); err != nil {
		return nil, err
	}

	return m, nil
}

// Middleware wraps HTTP handlers with instrumentation
func (m *HTTPMetrics) Middleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m.activeRequests.Inc()
		defer m.activeRequests.Dec()

		start := time.Now()
		next(w, r)
		duration := time.Since(start).Seconds()

		m.requestsTotal.Inc()
		m.requestDuration.Observe(duration)
	}
}

// ========== Custom Application Metrics ==========

// AppMetrics holds application-specific metrics
type AppMetrics struct {
	dbConnections    *Gauge
	cacheHits        *Counter
	cacheMisses      *Counter
	processingTime   *Histogram
	errorCount       *Counter
}

// NewAppMetrics creates application metrics
func NewAppMetrics(registry *Registry) (*AppMetrics, error) {
	m := &AppMetrics{
		dbConnections: NewGauge(
			"app_db_connections",
			"Number of active database connections",
		),
		cacheHits: NewCounter(
			"app_cache_hits_total",
			"Total cache hits",
		),
		cacheMisses: NewCounter(
			"app_cache_misses_total",
			"Total cache misses",
		),
		processingTime: NewHistogram(
			"app_processing_time_seconds",
			"Application processing time in seconds",
			[]float64{0.01, 0.05, 0.1, 0.5, 1, 2},
		),
		errorCount: NewCounter(
			"app_errors_total",
			"Total application errors",
		),
	}

	for _, metric := range []Metric{
		m.dbConnections,
		m.cacheHits,
		m.cacheMisses,
		m.processingTime,
		m.errorCount,
	} {
		if err := registry.Register(metric); err != nil {
			return nil, err
		}
	}

	return m, nil
}

// RecordDBConnection updates database connection count
func (m *AppMetrics) RecordDBConnection(count int) {
	m.dbConnections.Set(float64(count))
}

// RecordCacheHit records a cache hit
func (m *AppMetrics) RecordCacheHit() {
	m.cacheHits.Inc()
}

// RecordCacheMiss records a cache miss
func (m *AppMetrics) RecordCacheMiss() {
	m.cacheMisses.Inc()
}

// RecordProcessingTime records processing time
func (m *AppMetrics) RecordProcessingTime(duration time.Duration) {
	m.processingTime.Observe(duration.Seconds())
}

// RecordError records an error
func (m *AppMetrics) RecordError() {
	m.errorCount.Inc()
}

// ========== Tracer (Mock Implementation) ==========

// Span represents a single operation in a trace
type Span struct {
	Name      string
	StartTime time.Time
	EndTime   time.Time
	Tags      map[string]interface{}
	Logs      []string
}

// Tracer is a mock tracer for distributed tracing
type Tracer struct {
	traces map[string][]*Span
	mu     sync.RWMutex
}

// NewTracer creates a new tracer
func NewTracer() *Tracer {
	return &Tracer{
		traces: make(map[string][]*Span),
	}
}

// StartSpan begins a new span
func (t *Tracer) StartSpan(traceID, spanName string) *Span {
	return &Span{
		Name:      spanName,
		StartTime: time.Now(),
		Tags:      make(map[string]interface{}),
		Logs:      []string{},
	}
}

// FinishSpan ends a span and records it
func (t *Tracer) FinishSpan(traceID string, span *Span) {
	span.EndTime = time.Now()

	t.mu.Lock()
	defer t.mu.Unlock()

	t.traces[traceID] = append(t.traces[traceID], span)
}

// GetTrace retrieves recorded trace spans
func (t *Tracer) GetTrace(traceID string) []*Span {
	t.mu.RLock()
	defer t.mu.RUnlock()

	return t.traces[traceID]
}

// ========== Context Tracing ==========

// WithTrace adds trace information to context
func WithTrace(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, "trace_id", traceID)
}

// GetTraceID extracts trace ID from context
func GetTraceID(ctx context.Context) string {
	if traceID, ok := ctx.Value("trace_id").(string); ok {
		return traceID
	}
	return ""
}

// ========== Example Service ==========

// Service demonstrates metric and trace usage
type Service struct {
	registry  *Registry
	http      *HTTPMetrics
	app       *AppMetrics
	tracer    *Tracer
}

// NewService creates a service with metrics
func NewService() (*Service, error) {
	registry := NewRegistry()

	httpMetrics, err := NewHTTPMetrics(registry)
	if err != nil {
		return nil, err
	}

	appMetrics, err := NewAppMetrics(registry)
	if err != nil {
		return nil, err
	}

	return &Service{
		registry: registry,
		http:     httpMetrics,
		app:      appMetrics,
		tracer:   NewTracer(),
	}, nil
}

// ProcessRequest simulates request processing with metrics
func (s *Service) ProcessRequest(ctx context.Context, duration time.Duration) error {
	traceID := GetTraceID(ctx)

	// Create trace span
	span := s.tracer.StartSpan(traceID, "process_request")
	defer s.tracer.FinishSpan(traceID, span)

	// Simulate work
	time.Sleep(duration)

	// Record metrics
	s.app.RecordProcessingTime(duration)

	// Simulate caching
	if time.Now().Unix()%2 == 0 {
		s.app.RecordCacheHit()
	} else {
		s.app.RecordCacheMiss()
	}

	span.Tags["status"] = "success"
	return nil
}


// GetMetrics returns metrics in Prometheus format
func (s *Service) GetMetrics() string {
	return s.registry.Gather()
}

func main() {}
