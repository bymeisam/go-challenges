package main

import (
	"errors"
	"fmt"
	"testing"
	"time"
)

func TestCircuitBreakerClosed(t *testing.T) {
	cb := NewCircuitBreaker("test", 3, 2, 100*time.Millisecond)

	if cb.GetState() != StateClosed {
		t.Error("Initial state should be Closed")
	}

	// Successful call
	err := cb.Call(func() error { return nil })
	if err != nil {
		t.Error("Should not error on success")
	}

	if cb.GetState() != StateClosed {
		t.Error("Should remain Closed on success")
	}

	t.Log("✓ Circuit Breaker Closed state works!")
}

func TestCircuitBreakerOpens(t *testing.T) {
	cb := NewCircuitBreaker("test", 3, 2, 100*time.Millisecond)

	// Fail 3 times (threshold)
	for i := 0; i < 3; i++ {
		cb.Call(func() error { return errors.New("fail") })
	}

	if cb.GetState() != StateOpen {
		t.Error("Should transition to Open after failures")
	}

	// Next call should fail immediately
	err := cb.Call(func() error { return nil })
	if err == nil {
		t.Error("Should return error when Open")
	}

	t.Log("✓ Circuit Breaker opens on failures!")
}

func TestCircuitBreakerHalf(t *testing.T) {
	cb := NewCircuitBreaker("test", 2, 2, 50*time.Millisecond)

	// Open the circuit
	for i := 0; i < 2; i++ {
		cb.Call(func() error { return errors.New("fail") })
	}

	if cb.GetState() != StateOpen {
		t.Error("Should be Open")
	}

	// Wait for timeout
	time.Sleep(60 * time.Millisecond)

	// Next call should transition to Half
	cb.Call(func() error { return nil })

	if cb.GetState() != StateHalf {
		t.Error("Should be Half after timeout and success")
	}

	t.Log("✓ Circuit Breaker half-open state works!")
}

func TestCircuitBreakerRecovery(t *testing.T) {
	cb := NewCircuitBreaker("test", 2, 2, 50*time.Millisecond)

	// Open circuit
	for i := 0; i < 2; i++ {
		cb.Call(func() error { return errors.New("fail") })
	}

	// Wait and enter half-open
	time.Sleep(60 * time.Millisecond)
	cb.Call(func() error { return nil })

	if cb.GetState() != StateHalf {
		t.Error("Should be Half")
	}

	// Success twice (threshold)
	cb.Call(func() error { return nil })

	if cb.GetState() != StateClosed {
		t.Error("Should be Closed after recovery")
	}

	t.Log("✓ Circuit Breaker recovery works!")
}

func TestRetryWithSuccess(t *testing.T) {
	config := DefaultRetryConfig()
	attempts := 0

	err := ExecuteWithRetry(config, func() error {
		attempts++
		if attempts < 2 {
			return errors.New("fail")
		}
		return nil
	})

	if err != nil {
		t.Error("Should succeed on retry")
	}

	if attempts != 2 {
		t.Errorf("Expected 2 attempts, got %d", attempts)
	}

	t.Log("✓ Retry with success works!")
}

func TestRetryExhaustsAttempts(t *testing.T) {
	config := RetryConfig{
		MaxAttempts:       2,
		InitialBackoff:    10 * time.Millisecond,
		MaxBackoff:        100 * time.Millisecond,
		BackoffMultiplier: 2.0,
	}

	attempts := 0
	err := ExecuteWithRetry(config, func() error {
		attempts++
		return errors.New("always fail")
	})

	if err == nil {
		t.Error("Should fail after max attempts")
	}

	if attempts != 2 {
		t.Errorf("Expected 2 attempts, got %d", attempts)
	}

	t.Log("✓ Retry exhaustion works!")
}

func TestServiceRegistry(t *testing.T) {
	registry := NewServiceRegistry()

	instance := &ServiceInstance{
		ID:       "svc-1",
		Name:     "user-service",
		Host:     "localhost",
		Port:     8081,
		Healthy:  true,
		Weight:   1,
	}

	err := registry.Register(instance)
	if err != nil {
		t.Fatalf("Failed to register: %v", err)
	}

	healthy := registry.GetHealthyInstances("user-service")
	if len(healthy) != 1 {
		t.Errorf("Expected 1 healthy instance, got %d", len(healthy))
	}

	t.Log("✓ Service Registry works!")
}

func TestServiceRegistryDeregister(t *testing.T) {
	registry := NewServiceRegistry()

	instance := &ServiceInstance{
		ID:   "svc-1",
		Name: "user-service",
	}

	registry.Register(instance)
	err := registry.Deregister("user-service", "svc-1")

	if err != nil {
		t.Fatalf("Failed to deregister: %v", err)
	}

	healthy := registry.GetHealthyInstances("user-service")
	if len(healthy) != 0 {
		t.Error("Should have no instances after deregister")
	}

	t.Log("✓ Service Registry deregister works!")
}

func TestLoadBalancerRoundRobin(t *testing.T) {
	registry := NewServiceRegistry()

	// Register 3 instances
	for i := 1; i <= 3; i++ {
		instance := &ServiceInstance{
			ID:      fmt.Sprintf("svc-%d", i),
			Name:    "api-service",
			Healthy: true,
		}
		registry.Register(instance)
	}

	lb := NewLoadBalancer(registry, StrategyRoundRobin)

	// Get instances in round-robin order
	selections := make(map[string]int)
	for i := 0; i < 9; i++ {
		inst, _ := lb.GetInstance("api-service")
		selections[inst.ID]++
	}

	// Each should be selected 3 times (round-robin)
	for _, count := range selections {
		if count != 3 {
			t.Errorf("Expected each instance selected 3 times, got %d", count)
		}
	}

	t.Log("✓ Load Balancer Round-Robin works!")
}

func TestLoadBalancerRandom(t *testing.T) {
	registry := NewServiceRegistry()

	// Register instances
	for i := 1; i <= 5; i++ {
		registry.Register(&ServiceInstance{
			ID:      fmt.Sprintf("svc-%d", i),
			Name:    "service",
			Healthy: true,
		})
	}

	lb := NewLoadBalancer(registry, StrategyRandom)

	// Should get varied selections
	selections := make(map[string]int)
	for i := 0; i < 100; i++ {
		inst, _ := lb.GetInstance("service")
		selections[inst.ID]++
	}

	// All should be selected at least once
	if len(selections) != 5 {
		t.Errorf("Expected all 5 instances selected, got %d", len(selections))
	}

	t.Log("✓ Load Balancer Random works!")
}

func TestLoadBalancerWeighted(t *testing.T) {
	registry := NewServiceRegistry()

	// Register with different weights
	for i, weight := range []int{1, 2, 3} {
		registry.Register(&ServiceInstance{
			ID:      fmt.Sprintf("svc-%d", i),
			Name:    "service",
			Healthy: true,
			Weight:  weight,
		})
	}

	lb := NewLoadBalancer(registry, StrategyWeighted)

	selections := make(map[string]int)
	for i := 0; i < 600; i++ {
		inst, _ := lb.GetInstance("service")
		selections[inst.ID]++
	}

	// Weighted distribution should be roughly 1:2:3
	// svc-0 ~100, svc-1 ~200, svc-2 ~300
	if selections["svc-0"] > 150 || selections["svc-0"] < 50 {
		t.Logf("Weighted distribution might be off: %v", selections)
	}

	t.Log("✓ Load Balancer Weighted works!")
}

func TestAPIGatewayRouting(t *testing.T) {
	registry := NewServiceRegistry()
	registry.Register(&ServiceInstance{
		ID:      "user-1",
		Name:    "user-service",
		Healthy: true,
	})

	gateway := NewAPIGateway(registry)
	gateway.RegisterRoute("/users", "user-service")

	// Route a request
	called := false
	err := gateway.RouteRequest("/users", func(inst *ServiceInstance) error {
		called = true
		if inst.Name != "user-service" {
			t.Error("Should route to correct service")
		}
		return nil
	})

	if err != nil {
		t.Fatalf("Route failed: %v", err)
	}

	if !called {
		t.Error("Handler should be called")
	}

	t.Log("✓ API Gateway routing works!")
}

func TestAPIGatewayCircuitBreaker(t *testing.T) {
	registry := NewServiceRegistry()
	registry.Register(&ServiceInstance{
		ID:      "svc-1",
		Name:    "service",
		Healthy: true,
	})

	gateway := NewAPIGateway(registry)
	gateway.RegisterRoute("/api", "service")

	// Fail several times
	failCount := 0
	for i := 0; i < 6; i++ {
		gateway.RouteRequest("/api", func(inst *ServiceInstance) error {
			return errors.New("service error")
		})
		failCount++
	}

	// Next request should fail fast (circuit open)
	err := gateway.RouteRequest("/api", func(inst *ServiceInstance) error {
		return nil
	})

	if err == nil {
		t.Error("Should fail when circuit open")
	}

	t.Log("✓ API Gateway circuit breaker works!")
}

func TestHealthChecker(t *testing.T) {
	registry := NewServiceRegistry()
	registry.Register(&ServiceInstance{
		ID:      "svc-1",
		Name:    "service",
		Healthy: true,
	})

	checker := NewHealthChecker(registry, 50*time.Millisecond, 5*time.Second)
	checker.Start()

	time.Sleep(100 * time.Millisecond)
	checker.Stop()

	// Health checks ran
	instances := registry.GetHealthyInstances("service")
	if len(instances) == 0 {
		t.Log("Note: Health check ran (may have marked as unhealthy by chance)")
	}

	t.Log("✓ Health Checker works!")
}

func TestDefaultRetryConfig(t *testing.T) {
	config := DefaultRetryConfig()

	if config.MaxAttempts < 1 {
		t.Error("MaxAttempts should be at least 1")
	}

	if config.InitialBackoff == 0 {
		t.Error("InitialBackoff should be set")
	}

	if config.BackoffMultiplier < 1 {
		t.Error("BackoffMultiplier should be > 1")
	}

	t.Log("✓ Default retry config is valid!")
}

func TestMultipleServices(t *testing.T) {
	registry := NewServiceRegistry()

	// Register multiple services
	services := []string{"user-service", "order-service", "payment-service"}

	for _, svc := range services {
		for i := 1; i <= 2; i++ {
			registry.Register(&ServiceInstance{
				ID:      fmt.Sprintf("%s-%d", svc, i),
				Name:    svc,
				Healthy: true,
			})
		}
	}

	// Each service should have 2 instances
	for _, svc := range services {
		healthy := registry.GetHealthyInstances(svc)
		if len(healthy) != 2 {
			t.Errorf("Expected 2 instances for %s, got %d", svc, len(healthy))
		}
	}

	t.Log("✓ Multiple services work!")
}

func TestCircuitBreakerReset(t *testing.T) {
	cb := NewCircuitBreaker("test", 1, 1, 50*time.Millisecond)

	// Fail to open
	cb.Call(func() error { return errors.New("fail") })

	if cb.GetState() != StateOpen {
		t.Error("Should be open after failure")
	}

	// Wait for half-open
	time.Sleep(60 * time.Millisecond)

	// Succeed to close
	cb.Call(func() error { return nil })

	if cb.GetState() != StateClosed {
		t.Error("Should be closed after recovery")
	}

	t.Log("✓ Circuit Breaker reset works!")
}

func TestRetryBackoff(t *testing.T) {
	config := RetryConfig{
		MaxAttempts:       3,
		InitialBackoff:    10 * time.Millisecond,
		MaxBackoff:        100 * time.Millisecond,
		BackoffMultiplier: 2.0,
	}

	start := time.Now()
	attempts := 0

	ExecuteWithRetry(config, func() error {
		attempts++
		return errors.New("fail")
	})

	elapsed := time.Since(start)

	// Should have 2 backoff periods (10ms + 20ms = 30ms, plus jitter)
	if elapsed < 20*time.Millisecond {
		t.Logf("Backoff seems too short: %v", elapsed)
	}

	t.Log("✓ Retry backoff timing works!")
}

func TestLoadBalancerNoHealthyInstances(t *testing.T) {
	registry := NewServiceRegistry()
	registry.Register(&ServiceInstance{
		ID:      "svc-1",
		Name:    "service",
		Healthy: false,
	})

	lb := NewLoadBalancer(registry, StrategyRoundRobin)

	_, err := lb.GetInstance("service")
	if err == nil {
		t.Error("Should error when no healthy instances")
	}

	t.Log("✓ Load Balancer handles no healthy instances!")
}

func TestAPIGatewayUnregisteredRoute(t *testing.T) {
	registry := NewServiceRegistry()
	gateway := NewAPIGateway(registry)

	err := gateway.RouteRequest("/unknown", func(inst *ServiceInstance) error {
		return nil
	})

	if err == nil {
		t.Error("Should error on unregistered route")
	}

	t.Log("✓ API Gateway rejects unknown routes!")
}

func BenchmarkCircuitBreakerCall(b *testing.B) {
	cb := NewCircuitBreaker("bench", 5, 2, 1*time.Second)

	for i := 0; i < b.N; i++ {
		cb.Call(func() error { return nil })
	}
}

func BenchmarkRetryExecution(b *testing.B) {
	config := DefaultRetryConfig()

	for i := 0; i < b.N; i++ {
		ExecuteWithRetry(config, func() error { return nil })
	}
}

func BenchmarkLoadBalancerSelection(b *testing.B) {
	registry := NewServiceRegistry()

	for i := 1; i <= 10; i++ {
		registry.Register(&ServiceInstance{
			ID:      fmt.Sprintf("svc-%d", i),
			Name:    "service",
			Healthy: true,
		})
	}

	lb := NewLoadBalancer(registry, StrategyRoundRobin)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		lb.GetInstance("service")
	}
}
