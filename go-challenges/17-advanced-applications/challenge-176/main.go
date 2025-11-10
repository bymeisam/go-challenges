package main

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"sync/atomic"
	"time"
)

// APIGateway is the main gateway component
type APIGateway struct {
	routes           map[string]*Route
	rateLimiters     map[string]*RateLimiter
	authenticator    *Authenticator
	cache            *ResponseCache
	circuitBreakers  map[string]*CircuitBreaker
	loadBalancers    map[string]*LoadBalancer
	healthChecker    *HealthChecker
	metrics          *Metrics
	mu               sync.RWMutex
	requestIDCounter atomic.Int64
}

// Route represents a route configuration
type Route struct {
	Pattern          string
	Methods          []string
	Backends         []*Backend
	RateLimitPerMin  int
	RequireAuth      bool
	CacheTTL         time.Duration
	Transform        RequestTransformer
	ResponseHandler  ResponseTransformer
	CircuitBreaker   *CircuitBreaker
	LoadBalancer     *LoadBalancer
}

// Backend represents a backend service
type Backend struct {
	URL           *url.URL
	Weight        int
	HealthURL     string
	Timeout       time.Duration
	TotalRequests atomic.Int64
	Errors        atomic.Int64
}

// RateLimiter implements token bucket rate limiting
type RateLimiter struct {
	clientBuckets map[string]*TokenBucket
	perMinute     int
	mu            sync.RWMutex
}

// TokenBucket represents a token bucket for a client
type TokenBucket struct {
	tokens    atomic.Int64
	maxTokens int64
	refillAt  atomic.Int64
}

// Authenticator handles authentication
type Authenticator struct {
	apiKeys   map[string]bool
	jwtSecret string
	mu        sync.RWMutex
}

// ResponseCache caches responses with TTL
type ResponseCache struct {
	entries map[string]*CacheEntry
	mu      sync.RWMutex
}

// CacheEntry represents a cached response
type CacheEntry struct {
	Data      []byte
	ExpiresAt time.Time
	etag      string
}

// CircuitBreaker implements circuit breaker pattern
type CircuitBreaker struct {
	failureThreshold int
	successThreshold int
	timeout          time.Duration
	state            atomic.Value // string: "closed", "open", "half-open"
	failures         atomic.Int32
	successes        atomic.Int32
	lastFailTime     atomic.Value // time.Time
}

// LoadBalancer implements load balancing strategies
type LoadBalancer struct {
	backends       []*Backend
	strategy       string // "round-robin", "least-connections"
	roundRobinIdx  atomic.Int32
	mu             sync.RWMutex
}

// HealthChecker performs health checks on backends
type HealthChecker struct {
	backends  []*Backend
	interval  time.Duration
	mu        sync.RWMutex
	healthy   map[string]bool
	lastCheck map[string]time.Time
}

// RequestTransformer transforms requests
type RequestTransformer interface {
	Transform(req *http.Request) error
}

// ResponseTransformer transforms responses
type ResponseTransformer interface {
	Transform(resp *http.Response) error
}

// Metrics collects gateway metrics
type Metrics struct {
	totalRequests  atomic.Int64
	totalErrors    atomic.Int64
	totalLatency   atomic.Int64
	activeRequests atomic.Int32
	requestsByCode map[int]atomic.Int64
	mu             sync.RWMutex
}

// NewAPIGateway creates a new API gateway
func NewAPIGateway() *APIGateway {
	return &APIGateway{
		routes:          make(map[string]*Route),
		rateLimiters:    make(map[string]*RateLimiter),
		authenticator:   NewAuthenticator(),
		cache:           NewResponseCache(),
		circuitBreakers: make(map[string]*CircuitBreaker),
		loadBalancers:   make(map[string]*LoadBalancer),
		healthChecker:   NewHealthChecker(5 * time.Second),
		metrics:         NewMetrics(),
	}
}

// RegisterRoute registers a new route
func (ag *APIGateway) RegisterRoute(pattern string, route *Route) error {
	if pattern == "" {
		return errors.New("pattern cannot be empty")
	}
	if len(route.Methods) == 0 {
		return errors.New("at least one method must be specified")
	}
	if len(route.Backends) == 0 {
		return errors.New("at least one backend must be specified")
	}

	ag.mu.Lock()
	defer ag.mu.Unlock()

	ag.routes[pattern] = route

	// Initialize rate limiter
	ag.rateLimiters[pattern] = NewRateLimiter(route.RateLimitPerMin)

	// Initialize circuit breaker
	cb := NewCircuitBreaker(5, 2, 10*time.Second)
	ag.circuitBreakers[pattern] = cb
	route.CircuitBreaker = cb

	// Initialize load balancer
	lb := NewLoadBalancer(route.Backends, "round-robin")
	ag.loadBalancers[pattern] = lb
	route.LoadBalancer = lb

	// Register health check
	ag.healthChecker.Register(route.Backends)

	return nil
}

// ServeHTTP handles incoming HTTP requests
func (ag *APIGateway) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	requestID := ag.requestIDCounter.Add(1)
	r.Header.Set("X-Request-ID", fmt.Sprintf("%d", requestID))

	ag.metrics.activeRequests.Add(1)
	defer ag.metrics.activeRequests.Add(-1)

	// Find matching route
	route := ag.findRoute(r)
	if route == nil {
		ag.metrics.recordError()
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "route not found"})
		return
	}

	// Check authentication
	if route.RequireAuth {
		if !ag.authenticator.ValidateRequest(r) {
			ag.metrics.recordError()
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"})
			return
		}
	}

	// Check rate limit
	clientID := ag.getClientID(r)
	rl := ag.rateLimiters[findRouteKey(ag.routes, route)]
	if !rl.AllowRequest(clientID) {
		ag.metrics.recordError()
		w.WriteHeader(http.StatusTooManyRequests)
		json.NewEncoder(w).Encode(map[string]string{"error": "rate limit exceeded"})
		return
	}

	// Check cache
	cacheKey := ag.generateCacheKey(r)
	if cached := ag.cache.Get(cacheKey); cached != nil {
		w.Header().Set("X-Cache", "HIT")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(cached)
		ag.metrics.recordSuccess(time.Since(start), http.StatusOK)
		return
	}

	// Select backend
	backend := route.LoadBalancer.SelectBackend()
	if backend == nil {
		ag.metrics.recordError()
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{"error": "no available backends"})
		return
	}

	// Check circuit breaker
	if !route.CircuitBreaker.AllowRequest() {
		ag.metrics.recordError()
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{"error": "circuit breaker open"})
		return
	}

	// Transform request
	if route.Transform != nil {
		if err := route.Transform.Transform(r); err != nil {
			ag.metrics.recordError()
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}
	}

	// Proxy request
	proxy := httputil.NewSingleHostReverseProxy(backend.URL)
	proxy.Director = func(req *http.Request) {
		req.URL.Scheme = backend.URL.Scheme
		req.URL.Host = backend.URL.Host
		req.Host = backend.URL.Host
		req.RequestURI = ""
		req.Header.Set("X-Forwarded-For", r.RemoteAddr)
		req.Header.Set("X-Request-ID", fmt.Sprintf("%d", requestID))
	}

	// Create response writer wrapper
	responseWriter := &responseWriterWrapper{ResponseWriter: w}
	proxy.ServeHTTP(responseWriter, r)

	// Record metrics
	latency := time.Since(start)
	if responseWriter.statusCode == 0 {
		responseWriter.statusCode = http.StatusOK
	}

	if responseWriter.statusCode >= 400 {
		ag.metrics.recordError()
		route.CircuitBreaker.RecordFailure()
		backend.Errors.Add(1)
	} else {
		route.CircuitBreaker.RecordSuccess()
		backend.TotalRequests.Add(1)
	}

	ag.metrics.recordSuccess(latency, responseWriter.statusCode)

	// Cache successful response
	if route.CacheTTL > 0 && responseWriter.statusCode == http.StatusOK {
		ag.cache.Set(cacheKey, responseWriter.body, route.CacheTTL)
	}
}

// findRoute finds a matching route for the request
func (ag *APIGateway) findRoute(r *http.Request) *Route {
	ag.mu.RLock()
	defer ag.mu.RUnlock()

	for pattern, route := range ag.routes {
		if matchPath(pattern, r.URL.Path) && methodAllowed(route.Methods, r.Method) {
			return route
		}
	}
	return nil
}

// getClientID extracts client ID from request
func (ag *APIGateway) getClientID(r *http.Request) string {
	if apiKey := r.Header.Get("X-API-Key"); apiKey != "" {
		return apiKey
	}
	return r.RemoteAddr
}

// generateCacheKey generates a cache key for the request
func (ag *APIGateway) generateCacheKey(r *http.Request) string {
	h := sha256.New()
	h.Write([]byte(r.Method + ":" + r.URL.String()))
	return fmt.Sprintf("%x", h.Sum(nil))
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(perMinute int) *RateLimiter {
	return &RateLimiter{
		clientBuckets: make(map[string]*TokenBucket),
		perMinute:     perMinute,
	}
}

// AllowRequest checks if a request is allowed
func (rl *RateLimiter) AllowRequest(clientID string) bool {
	if rl.perMinute == 0 {
		return true
	}

	rl.mu.Lock()
	defer rl.mu.Unlock()

	bucket, exists := rl.clientBuckets[clientID]
	if !exists {
		bucket = &TokenBucket{
			maxTokens: int64(rl.perMinute),
		}
		bucket.tokens.Store(int64(rl.perMinute))
		bucket.refillAt.Store(time.Now().Add(time.Minute).UnixNano())
		rl.clientBuckets[clientID] = bucket
	}

	now := time.Now().UnixNano()
	refillAt := bucket.refillAt.Load()
	if now >= refillAt {
		bucket.tokens.Store(bucket.maxTokens)
		bucket.refillAt.Store(time.Now().Add(time.Minute).UnixNano())
	}

	tokens := bucket.tokens.Load()
	if tokens > 0 {
		bucket.tokens.Add(-1)
		return true
	}
	return false
}

// NewAuthenticator creates a new authenticator
func NewAuthenticator() *Authenticator {
	return &Authenticator{
		apiKeys:   make(map[string]bool),
		jwtSecret: "secret-key",
	}
}

// RegisterAPIKey registers an API key
func (a *Authenticator) RegisterAPIKey(key string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.apiKeys[key] = true
}

// ValidateRequest validates the request authentication
func (a *Authenticator) ValidateRequest(r *http.Request) bool {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if apiKey := r.Header.Get("X-API-Key"); apiKey != "" {
		return a.apiKeys[apiKey]
	}

	if authHeader := r.Header.Get("Authorization"); authHeader != "" {
		// Basic JWT validation (simplified)
		return len(authHeader) > 7
	}

	return false
}

// NewResponseCache creates a new response cache
func NewResponseCache() *ResponseCache {
	cache := &ResponseCache{
		entries: make(map[string]*CacheEntry),
	}
	go cache.cleanup()
	return cache
}

// Get retrieves a cached response
func (rc *ResponseCache) Get(key string) []byte {
	rc.mu.RLock()
	defer rc.mu.RUnlock()

	entry, exists := rc.entries[key]
	if !exists {
		return nil
	}
	if time.Now().After(entry.ExpiresAt) {
		return nil
	}
	return entry.Data
}

// Set stores a response in cache
func (rc *ResponseCache) Set(key string, data []byte, ttl time.Duration) {
	rc.mu.Lock()
	defer rc.mu.Unlock()

	rc.entries[key] = &CacheEntry{
		Data:      data,
		ExpiresAt: time.Now().Add(ttl),
	}
}

// Invalidate removes a cache entry
func (rc *ResponseCache) Invalidate(key string) {
	rc.mu.Lock()
	defer rc.mu.Unlock()
	delete(rc.entries, key)
}

// cleanup removes expired entries
func (rc *ResponseCache) cleanup() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		rc.mu.Lock()
		now := time.Now()
		for key, entry := range rc.entries {
			if now.After(entry.ExpiresAt) {
				delete(rc.entries, key)
			}
		}
		rc.mu.Unlock()
	}
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(failureThreshold, successThreshold int, timeout time.Duration) *CircuitBreaker {
	cb := &CircuitBreaker{
		failureThreshold: failureThreshold,
		successThreshold: successThreshold,
		timeout:          timeout,
	}
	cb.state.Store("closed")
	return cb
}

// AllowRequest checks if the circuit breaker allows requests
func (cb *CircuitBreaker) AllowRequest() bool {
	state := cb.state.Load().(string)
	switch state {
	case "closed":
		return true
	case "half-open":
		return true
	case "open":
		if time.Now().After(cb.getLastFailTime().Add(cb.timeout)) {
			cb.state.Store("half-open")
			cb.successes.Store(0)
			return true
		}
		return false
	}
	return false
}

// RecordSuccess records a successful request
func (cb *CircuitBreaker) RecordSuccess() {
	state := cb.state.Load().(string)
	if state == "half-open" {
		if cb.successes.Add(1) >= int32(cb.successThreshold) {
			cb.state.Store("closed")
			cb.failures.Store(0)
		}
	}
}

// RecordFailure records a failed request
func (cb *CircuitBreaker) RecordFailure() {
	state := cb.state.Load().(string)
	cb.failures.Add(1)

	if state == "closed" && cb.failures.Load() >= int32(cb.failureThreshold) {
		cb.state.Store("open")
		cb.lastFailTime.Store(time.Now())
	} else if state == "half-open" {
		cb.state.Store("open")
		cb.lastFailTime.Store(time.Now())
	}
}

func (cb *CircuitBreaker) getLastFailTime() time.Time {
	if t := cb.lastFailTime.Load(); t != nil {
		return t.(time.Time)
	}
	return time.Now()
}

// NewLoadBalancer creates a new load balancer
func NewLoadBalancer(backends []*Backend, strategy string) *LoadBalancer {
	return &LoadBalancer{
		backends: backends,
		strategy: strategy,
	}
}

// SelectBackend selects a backend for the request
func (lb *LoadBalancer) SelectBackend() *Backend {
	if len(lb.backends) == 0 {
		return nil
	}

	switch lb.strategy {
	case "least-connections":
		return lb.selectLeastConnections()
	default:
		return lb.selectRoundRobin()
	}
}

// selectRoundRobin selects a backend using round-robin
func (lb *LoadBalancer) selectRoundRobin() *Backend {
	idx := lb.roundRobinIdx.Add(1)
	return lb.backends[int(idx)%len(lb.backends)]
}

// selectLeastConnections selects a backend with least connections
func (lb *LoadBalancer) selectLeastConnections() *Backend {
	minRequests := lb.backends[0].TotalRequests.Load()
	selected := lb.backends[0]

	for _, backend := range lb.backends[1:] {
		reqs := backend.TotalRequests.Load()
		if reqs < minRequests {
			minRequests = reqs
			selected = backend
		}
	}
	return selected
}

// NewHealthChecker creates a new health checker
func NewHealthChecker(interval time.Duration) *HealthChecker {
	return &HealthChecker{
		backends:  make([]*Backend, 0),
		interval:  interval,
		healthy:   make(map[string]bool),
		lastCheck: make(map[string]time.Time),
	}
}

// Register registers backends for health checking
func (hc *HealthChecker) Register(backends []*Backend) {
	hc.mu.Lock()
	defer hc.mu.Unlock()

	for _, backend := range backends {
		key := backend.URL.String()
		hc.backends = append(hc.backends, backend)
		hc.healthy[key] = true
	}
}

// Check performs a health check
func (hc *HealthChecker) Check(backend *Backend) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", backend.HealthURL, nil)
	if err != nil {
		return false
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}

// IsHealthy returns the health status of a backend
func (hc *HealthChecker) IsHealthy(backend *Backend) bool {
	hc.mu.RLock()
	defer hc.mu.RUnlock()
	return hc.healthy[backend.URL.String()]
}

// NewMetrics creates a new metrics collector
func NewMetrics() *Metrics {
	return &Metrics{
		requestsByCode: make(map[int]atomic.Int64),
	}
}

// recordSuccess records a successful request
func (m *Metrics) recordSuccess(latency time.Duration, code int) {
	m.totalRequests.Add(1)
	m.totalLatency.Add(latency.Milliseconds())
}

// recordError records a failed request
func (m *Metrics) recordError() {
	m.totalErrors.Add(1)
	m.totalRequests.Add(1)
}

// GetMetrics returns current metrics
func (m *Metrics) GetMetrics() map[string]interface{} {
	total := m.totalRequests.Load()
	errors := m.totalErrors.Load()
	latency := m.totalLatency.Load()

	avgLatency := int64(0)
	if total > 0 {
		avgLatency = latency / total
	}

	return map[string]interface{}{
		"total_requests":  total,
		"total_errors":    errors,
		"error_rate":      float64(errors) / float64(total),
		"avg_latency_ms":  avgLatency,
		"active_requests": m.activeRequests.Load(),
	}
}

// Helper functions

func matchPath(pattern, path string) bool {
	if pattern == "/" {
		return true
	}
	if pattern == path {
		return true
	}
	if len(pattern) > 0 && pattern[len(pattern)-1] == '*' {
		return len(path) >= len(pattern)-1 && path[:len(pattern)-1] == pattern[:len(pattern)-1]
	}
	return false
}

func methodAllowed(methods []string, method string) bool {
	for _, m := range methods {
		if m == method || m == "*" {
			return true
		}
	}
	return false
}

func findRouteKey(routes map[string]*Route, target *Route) string {
	for key, route := range routes {
		if route == target {
			return key
		}
	}
	return ""
}

type responseWriterWrapper struct {
	http.ResponseWriter
	statusCode int
	body       []byte
}

func (rw *responseWriterWrapper) WriteHeader(statusCode int) {
	rw.statusCode = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

func (rw *responseWriterWrapper) Write(data []byte) (int, error) {
	rw.body = append(rw.body, data...)
	return rw.ResponseWriter.Write(data)
}

// SimpleTransformer is a basic transformer for testing
type SimpleTransformer struct {
	addHeader string
}

func (st *SimpleTransformer) Transform(req *http.Request) error {
	if st.addHeader != "" {
		req.Header.Set("X-Transformed", st.addHeader)
	}
	return nil
}

func main() {
	gateway := NewAPIGateway()

	// Register API key
	gateway.authenticator.RegisterAPIKey("test-key-123")

	// Create backends
	backends := []*Backend{
		{
			URL:       parseURL("http://localhost:8081"),
			Weight:    1,
			HealthURL: "http://localhost:8081/health",
			Timeout:   5 * time.Second,
		},
		{
			URL:       parseURL("http://localhost:8082"),
			Weight:    1,
			HealthURL: "http://localhost:8082/health",
			Timeout:   5 * time.Second,
		},
	}

	// Register route
	route := &Route{
		Pattern:         "/api/users/*",
		Methods:         []string{"GET", "POST"},
		Backends:        backends,
		RateLimitPerMin: 100,
		RequireAuth:     true,
		CacheTTL:        5 * time.Minute,
		Transform:       &SimpleTransformer{addHeader: "test"},
	}

	gateway.RegisterRoute("/api/users/*", route)

	// Start server
	http.Handle("/", gateway)
	http.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(gateway.metrics.GetMetrics())
	})

	fmt.Println("API Gateway listening on :8080")
	http.ListenAndServe(":8080", nil)
}

func parseURL(urlStr string) *url.URL {
	u, _ := url.Parse(urlStr)
	return u
}
