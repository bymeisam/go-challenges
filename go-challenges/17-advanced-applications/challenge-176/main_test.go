package main

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

func TestAPIGatewayCreation(t *testing.T) {
	gateway := NewAPIGateway()

	if gateway.routes == nil {
		t.Error("routes map should be initialized")
	}
	if gateway.authenticator == nil {
		t.Error("authenticator should be initialized")
	}
	if gateway.cache == nil {
		t.Error("cache should be initialized")
	}
	if gateway.metrics == nil {
		t.Error("metrics should be initialized")
	}
}

func TestRouteRegistration(t *testing.T) {
	gateway := NewAPIGateway()
	backends := []*Backend{
		{URL: parseURL("http://localhost:8081"), Weight: 1},
	}

	route := &Route{
		Pattern:         "/api/test",
		Methods:         []string{"GET"},
		Backends:        backends,
		RateLimitPerMin: 100,
	}

	err := gateway.RegisterRoute("/api/test", route)
	if err != nil {
		t.Fatalf("RegisterRoute failed: %v", err)
	}

	if _, ok := gateway.routes["/api/test"]; !ok {
		t.Error("route not registered")
	}
	if _, ok := gateway.rateLimiters["/api/test"]; !ok {
		t.Error("rate limiter not initialized")
	}
	if _, ok := gateway.circuitBreakers["/api/test"]; !ok {
		t.Error("circuit breaker not initialized")
	}
	if _, ok := gateway.loadBalancers["/api/test"]; !ok {
		t.Error("load balancer not initialized")
	}
}

func TestRouteRegistrationErrors(t *testing.T) {
	gateway := NewAPIGateway()
	backends := []*Backend{
		{URL: parseURL("http://localhost:8081")},
	}

	tests := []struct {
		name    string
		pattern string
		route   *Route
	}{
		{
			name:    "empty pattern",
			pattern: "",
			route:   &Route{Methods: []string{"GET"}, Backends: backends},
		},
		{
			name:    "no methods",
			pattern: "/api/test",
			route:   &Route{Methods: []string{}, Backends: backends},
		},
		{
			name:    "no backends",
			pattern: "/api/test",
			route:   &Route{Methods: []string{"GET"}, Backends: []*Backend{}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := gateway.RegisterRoute(tt.pattern, tt.route)
			if err == nil {
				t.Error("expected error")
			}
		})
	}
}

func TestRateLimiterBasic(t *testing.T) {
	rl := NewRateLimiter(5)
	clientID := "test-client"

	// Should allow first 5 requests
	for i := 0; i < 5; i++ {
		if !rl.AllowRequest(clientID) {
			t.Errorf("request %d should be allowed", i+1)
		}
	}

	// 6th request should be denied
	if rl.AllowRequest(clientID) {
		t.Error("6th request should be denied")
	}
}

func TestRateLimiterMultipleClients(t *testing.T) {
	rl := NewRateLimiter(3)

	client1 := "client1"
	client2 := "client2"

	// Each client should have independent limits
	for i := 0; i < 3; i++ {
		if !rl.AllowRequest(client1) {
			t.Errorf("client1 request %d should be allowed", i+1)
		}
		if !rl.AllowRequest(client2) {
			t.Errorf("client2 request %d should be allowed", i+1)
		}
	}

	if rl.AllowRequest(client1) {
		t.Error("client1 should be rate limited")
	}
	if rl.AllowRequest(client2) {
		t.Error("client2 should be rate limited")
	}
}

func TestRateLimiterZeroLimit(t *testing.T) {
	rl := NewRateLimiter(0)
	if !rl.AllowRequest("any-client") {
		t.Error("zero limit should allow all requests")
	}
}

func TestAuthenticator(t *testing.T) {
	auth := NewAuthenticator()
	auth.RegisterAPIKey("valid-key")

	tests := []struct {
		name      string
		headerKey string
		value     string
		expected  bool
	}{
		{
			name:      "valid api key",
			headerKey: "X-API-Key",
			value:     "valid-key",
			expected:  true,
		},
		{
			name:      "invalid api key",
			headerKey: "X-API-Key",
			value:     "invalid-key",
			expected:  false,
		},
		{
			name:      "valid jwt",
			headerKey: "Authorization",
			value:     "Bearer eyJhbGc...",
			expected:  true,
		},
		{
			name:      "no auth",
			headerKey: "",
			value:     "",
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/api/test", nil)
			if tt.headerKey != "" {
				req.Header.Set(tt.headerKey, tt.value)
			}

			result := auth.ValidateRequest(req)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestResponseCacheBasic(t *testing.T) {
	cache := NewResponseCache()
	key := "test-key"
	data := []byte("cached data")
	ttl := 10 * time.Second

	cache.Set(key, data, ttl)

	cached := cache.Get(key)
	if !bytes.Equal(cached, data) {
		t.Error("cached data doesn't match")
	}
}

func TestResponseCacheExpiration(t *testing.T) {
	cache := NewResponseCache()
	key := "test-key"
	data := []byte("cached data")
	ttl := 100 * time.Millisecond

	cache.Set(key, data, ttl)

	// Should be cached immediately
	if cached := cache.Get(key); cached == nil {
		t.Error("data should be cached")
	}

	// Wait for expiration
	time.Sleep(150 * time.Millisecond)

	// Should be expired
	if cached := cache.Get(key); cached != nil {
		t.Error("cached data should be expired")
	}
}

func TestResponseCacheInvalidation(t *testing.T) {
	cache := NewResponseCache()
	key := "test-key"
	data := []byte("cached data")

	cache.Set(key, data, 10*time.Second)
	if cache.Get(key) == nil {
		t.Error("data should be cached")
	}

	cache.Invalidate(key)
	if cache.Get(key) != nil {
		t.Error("data should be invalidated")
	}
}

func TestCircuitBreakerBasic(t *testing.T) {
	cb := NewCircuitBreaker(3, 2, 100*time.Millisecond)

	// Circuit should start closed
	if !cb.AllowRequest() {
		t.Error("circuit should be closed initially")
	}

	// Record failures
	for i := 0; i < 3; i++ {
		cb.RecordFailure()
	}

	// Circuit should be open
	if cb.AllowRequest() {
		t.Error("circuit should be open after failures")
	}

	// Wait for timeout
	time.Sleep(150 * time.Millisecond)

	// Circuit should be half-open
	if !cb.AllowRequest() {
		t.Error("circuit should be half-open after timeout")
	}

	// Record successes
	for i := 0; i < 2; i++ {
		cb.RecordSuccess()
	}

	// Circuit should be closed again
	if !cb.AllowRequest() {
		t.Error("circuit should be closed after successes")
	}
}

func TestCircuitBreakerHalfOpen(t *testing.T) {
	cb := NewCircuitBreaker(2, 1, 50*time.Millisecond)

	// Open the circuit
	cb.RecordFailure()
	cb.RecordFailure()

	// Wait for timeout
	time.Sleep(100 * time.Millisecond)

	// Should be half-open
	if !cb.AllowRequest() {
		t.Error("should allow request in half-open state")
	}

	// Single success should close it
	cb.RecordSuccess()
	if !cb.AllowRequest() {
		t.Error("should be closed after success threshold")
	}
}

func TestLoadBalancerRoundRobin(t *testing.T) {
	backends := []*Backend{
		{URL: parseURL("http://localhost:8081")},
		{URL: parseURL("http://localhost:8082")},
		{URL: parseURL("http://localhost:8083")},
	}

	lb := NewLoadBalancer(backends, "round-robin")

	// Should cycle through backends
	selected := make(map[*Backend]int)
	for i := 0; i < 9; i++ {
		b := lb.SelectBackend()
		selected[b]++
	}

	// Each backend should be selected 3 times
	for _, count := range selected {
		if count != 3 {
			t.Errorf("expected each backend selected 3 times, got %d", count)
		}
	}
}

func TestLoadBalancerLeastConnections(t *testing.T) {
	backends := []*Backend{
		{URL: parseURL("http://localhost:8081")},
		{URL: parseURL("http://localhost:8082")},
		{URL: parseURL("http://localhost:8083")},
	}

	// Set up connection counts
	backends[0].TotalRequests.Store(10)
	backends[1].TotalRequests.Store(5)
	backends[2].TotalRequests.Store(15)

	lb := NewLoadBalancer(backends, "least-connections")

	selected := lb.SelectBackend()
	if selected != backends[1] {
		t.Error("should select backend with least connections")
	}
}

func TestHealthChecker(t *testing.T) {
	hc := NewHealthChecker(1 * time.Second)

	backends := []*Backend{
		{
			URL:       parseURL("http://localhost:8081"),
			HealthURL: "http://localhost:8081/health",
		},
	}

	hc.Register(backends)

	if len(hc.backends) != 1 {
		t.Error("backend not registered")
	}

	// New backend should be healthy by default
	if !hc.IsHealthy(backends[0]) {
		t.Error("new backend should be healthy")
	}
}

func TestMetricsCollection(t *testing.T) {
	metrics := NewMetrics()

	// Record some metrics
	metrics.recordSuccess(100*time.Millisecond, http.StatusOK)
	metrics.recordSuccess(150*time.Millisecond, http.StatusOK)
	metrics.recordError()

	m := metrics.GetMetrics()

	if m["total_requests"] != int64(3) {
		t.Errorf("expected 3 total requests, got %v", m["total_requests"])
	}

	if m["total_errors"] != int64(1) {
		t.Errorf("expected 1 error, got %v", m["total_errors"])
	}

	errorRate := m["error_rate"].(float64)
	expected := 1.0 / 3.0
	if errorRate < expected-0.01 || errorRate > expected+0.01 {
		t.Errorf("expected error rate ~%.4f, got %.4f", expected, errorRate)
	}
}

func TestSimpleTransformer(t *testing.T) {
	transformer := &SimpleTransformer{addHeader: "test-value"}
	req := httptest.NewRequest("GET", "/test", nil)

	err := transformer.Transform(req)
	if err != nil {
		t.Fatalf("Transform failed: %v", err)
	}

	if req.Header.Get("X-Transformed") != "test-value" {
		t.Error("header not set by transformer")
	}
}

func TestPathMatching(t *testing.T) {
	tests := []struct {
		pattern string
		path    string
		matches bool
	}{
		{"/api/users", "/api/users", true},
		{"/api/users", "/api/posts", false},
		{"/api/users/*", "/api/users/123", true},
		{"/api/users/*", "/api/users/123/profile", true},
		{"/api/users/*", "/api/posts/123", false},
		{"/", "/anything", true},
	}

	for _, tt := range tests {
		t.Run(tt.pattern+":"+tt.path, func(t *testing.T) {
			result := matchPath(tt.pattern, tt.path)
			if result != tt.matches {
				t.Errorf("expected %v, got %v", tt.matches, result)
			}
		})
	}
}

func TestMethodAllowed(t *testing.T) {
	tests := []struct {
		methods  []string
		method   string
		expected bool
	}{
		{[]string{"GET"}, "GET", true},
		{[]string{"GET"}, "POST", false},
		{[]string{"GET", "POST"}, "POST", true},
		{[]string{"*"}, "DELETE", true},
	}

	for _, tt := range tests {
		t.Run(tt.method, func(t *testing.T) {
			result := methodAllowed(tt.methods, tt.method)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestGatewayWithMockedBackend(t *testing.T) {
	// Create mock backend server
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok"}`))
	}))
	defer backend.Close()

	// Create gateway
	gateway := NewAPIGateway()
	gateway.authenticator.RegisterAPIKey("test-key")

	backendURL, _ := url.Parse(backend.URL)
	route := &Route{
		Pattern:         "/api/test",
		Methods:         []string{"GET"},
		Backends:        []*Backend{{URL: backendURL}},
		RateLimitPerMin: 100,
		RequireAuth:     true,
	}

	gateway.RegisterRoute("/api/test", route)

	// Make request
	req := httptest.NewRequest("GET", "/api/test", nil)
	req.Header.Set("X-API-Key", "test-key")

	w := httptest.NewRecorder()
	gateway.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	body, _ := io.ReadAll(w.Body)
	if !bytes.Contains(body, []byte("status")) {
		t.Error("response body not correct")
	}
}

func TestGatewayUnauthorized(t *testing.T) {
	gateway := NewAPIGateway()
	gateway.authenticator.RegisterAPIKey("test-key")

	backends := []*Backend{
		{URL: parseURL("http://localhost:8081")},
	}

	route := &Route{
		Pattern:         "/api/protected",
		Methods:         []string{"GET"},
		Backends:        backends,
		RateLimitPerMin: 100,
		RequireAuth:     true,
	}

	gateway.RegisterRoute("/api/protected", route)

	// Request without API key
	req := httptest.NewRequest("GET", "/api/protected", nil)
	w := httptest.NewRecorder()
	gateway.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", w.Code)
	}
}

func TestGatewayRateLimited(t *testing.T) {
	gateway := NewAPIGateway()
	gateway.authenticator.RegisterAPIKey("test-key")

	backends := []*Backend{
		{URL: parseURL("http://localhost:8081")},
	}

	route := &Route{
		Pattern:         "/api/limited",
		Methods:         []string{"GET"},
		Backends:        backends,
		RateLimitPerMin: 2,
	}

	gateway.RegisterRoute("/api/limited", route)

	clientID := "192.168.1.1"

	// Make 3 requests from same client
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest("GET", "/api/limited", nil)
		req.RemoteAddr = clientID

		w := httptest.NewRecorder()
		gateway.ServeHTTP(w, req)

		if i < 2 && w.Code != http.StatusNotFound {
			// First 2 should get not found (no backend)
		} else if i == 2 && w.Code != http.StatusTooManyRequests {
			t.Errorf("expected 429, got %d", w.Code)
		}
	}
}

func TestGatewayNotFound(t *testing.T) {
	gateway := NewAPIGateway()

	req := httptest.NewRequest("GET", "/api/nonexistent", nil)
	w := httptest.NewRecorder()
	gateway.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}
}

func TestGatewayWithTransformer(t *testing.T) {
	// Create mock backend
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if transformer added header
		if r.Header.Get("X-Transformed") == "" {
			t.Error("transformer didn't add header")
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}))
	defer backend.Close()

	gateway := NewAPIGateway()
	backendURL, _ := url.Parse(backend.URL)

	transformer := &SimpleTransformer{addHeader: "value"}
	route := &Route{
		Pattern:         "/api/test",
		Methods:         []string{"GET"},
		Backends:        []*Backend{{URL: backendURL}},
		RateLimitPerMin: 100,
		Transform:       transformer,
	}

	gateway.RegisterRoute("/api/test", route)

	req := httptest.NewRequest("GET", "/api/test", nil)
	w := httptest.NewRecorder()
	gateway.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestCacheKeyGeneration(t *testing.T) {
	gateway := NewAPIGateway()

	req1 := httptest.NewRequest("GET", "/api/test", nil)
	req2 := httptest.NewRequest("GET", "/api/test", nil)
	req3 := httptest.NewRequest("POST", "/api/test", nil)

	key1 := gateway.generateCacheKey(req1)
	key2 := gateway.generateCacheKey(req2)
	key3 := gateway.generateCacheKey(req3)

	if key1 != key2 {
		t.Error("same requests should generate same cache key")
	}

	if key1 == key3 {
		t.Error("different methods should generate different cache key")
	}
}

func TestGetClientID(t *testing.T) {
	gateway := NewAPIGateway()

	// Test with API key
	req1 := httptest.NewRequest("GET", "/test", nil)
	req1.Header.Set("X-API-Key", "my-key")
	if gateway.getClientID(req1) != "my-key" {
		t.Error("should use API key as client ID")
	}

	// Test with remote address
	req2 := httptest.NewRequest("GET", "/test", nil)
	req2.RemoteAddr = "192.168.1.1:5000"
	if gateway.getClientID(req2) != "192.168.1.1:5000" {
		t.Error("should use remote address as fallback")
	}
}

func TestResponseWriterWrapper(t *testing.T) {
	recorder := httptest.NewRecorder()
	wrapped := &responseWriterWrapper{ResponseWriter: recorder}

	wrapped.WriteHeader(http.StatusOK)
	wrapped.Write([]byte("test"))

	if wrapped.statusCode != http.StatusOK {
		t.Error("status code not recorded")
	}

	if !bytes.Equal(wrapped.body, []byte("test")) {
		t.Error("body not recorded")
	}
}

func BenchmarkRateLimiterAllowRequest(b *testing.B) {
	rl := NewRateLimiter(1000)
	clientID := "test-client"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rl.AllowRequest(clientID)
	}
}

func BenchmarkCircuitBreakerAllowRequest(b *testing.B) {
	cb := NewCircuitBreaker(5, 2, 10*time.Second)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cb.AllowRequest()
	}
}

func BenchmarkLoadBalancerSelect(b *testing.B) {
	backends := []*Backend{
		{URL: parseURL("http://localhost:8081")},
		{URL: parseURL("http://localhost:8082")},
		{URL: parseURL("http://localhost:8083")},
	}
	lb := NewLoadBalancer(backends, "round-robin")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		lb.SelectBackend()
	}
}

func BenchmarkCacheGetSet(b *testing.B) {
	cache := NewResponseCache()
	key := "bench-key"
	data := []byte("benchmark data")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Set(key, data, 10*time.Second)
		cache.Get(key)
	}
}
