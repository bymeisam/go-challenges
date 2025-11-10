package main

import (
	"fmt"
	"math"
	"math/rand"
	"sync"
	"time"
)

// ========== Circuit Breaker Pattern ==========

// CircuitState represents the state of a circuit breaker
type CircuitState string

const (
	StateClosed CircuitState = "closed"  // Normal operation
	StateOpen   CircuitState = "open"    // Failing, reject requests
	StateHalf   CircuitState = "half"    // Testing recovery
)

// CircuitBreaker implements the circuit breaker pattern
type CircuitBreaker struct {
	name              string
	state             CircuitState
	failureCount      int
	successCount      int
	failureThreshold  int
	successThreshold  int
	timeout           time.Duration
	lastFailureTime   time.Time
	lastAttemptTime   time.Time
	mu                sync.RWMutex
}

// NewCircuitBreaker creates a circuit breaker
func NewCircuitBreaker(name string, failureThreshold, successThreshold int, timeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		name:             name,
		state:            StateClosed,
		failureThreshold: failureThreshold,
		successThreshold: successThreshold,
		timeout:          timeout,
	}
}

// Call executes a function with circuit breaker protection
func (cb *CircuitBreaker) Call(fn func() error) error {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	// Check if we should transition from Open to Half
	if cb.state == StateOpen {
		if time.Since(cb.lastFailureTime) > cb.timeout {
			cb.state = StateHalf
			cb.successCount = 0
		} else {
			return fmt.Errorf("circuit breaker open: %s", cb.name)
		}
	}

	// Execute the function
	err := fn()
	cb.lastAttemptTime = time.Now()

	if err != nil {
		cb.onFailure()
		return err
	}

	cb.onSuccess()
	return nil
}

// onSuccess handles successful calls
func (cb *CircuitBreaker) onSuccess() {
	cb.failureCount = 0

	if cb.state == StateHalf {
		cb.successCount++
		if cb.successCount >= cb.successThreshold {
			cb.state = StateClosed
			cb.successCount = 0
		}
	}
}

// onFailure handles failed calls
func (cb *CircuitBreaker) onFailure() {
	cb.failureCount++
	cb.lastFailureTime = time.Now()

	if cb.state == StateClosed && cb.failureCount >= cb.failureThreshold {
		cb.state = StateOpen
	} else if cb.state == StateHalf {
		cb.state = StateOpen
		cb.successCount = 0
	}
}

// GetState returns current circuit breaker state
func (cb *CircuitBreaker) GetState() CircuitState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// ========== Retry with Exponential Backoff ==========

// RetryConfig configuration for retry strategy
type RetryConfig struct {
	MaxAttempts      int
	InitialBackoff   time.Duration
	MaxBackoff       time.Duration
	BackoffMultiplier float64
}

// DefaultRetryConfig returns default retry configuration
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts:       3,
		InitialBackoff:    100 * time.Millisecond,
		MaxBackoff:        10 * time.Second,
		BackoffMultiplier: 2.0,
	}
}

// ExecuteWithRetry executes function with exponential backoff
func ExecuteWithRetry(config RetryConfig, fn func() error) error {
	var lastErr error
	backoff := config.InitialBackoff

	for attempt := 0; attempt < config.MaxAttempts; attempt++ {
		err := fn()
		if err == nil {
			return nil
		}

		lastErr = err

		if attempt < config.MaxAttempts-1 {
			// Add jitter to prevent thundering herd
			jitter := time.Duration(rand.Int63n(int64(backoff) / 2))
			time.Sleep(backoff + jitter)

			// Calculate next backoff
			backoff = time.Duration(float64(backoff) * config.BackoffMultiplier)
			if backoff > config.MaxBackoff {
				backoff = config.MaxBackoff
			}
		}
	}

	return lastErr
}

// ========== Service Discovery ==========

// ServiceInstance represents a registered service
type ServiceInstance struct {
	ID       string
	Name     string
	Host     string
	Port     int
	Metadata map[string]string
	Healthy  bool
	Weight   int // For load balancing
}

// ServiceRegistry manages service discovery
type ServiceRegistry struct {
	services map[string][]*ServiceInstance
	mu       sync.RWMutex
}

// NewServiceRegistry creates a service registry
func NewServiceRegistry() *ServiceRegistry {
	return &ServiceRegistry{
		services: make(map[string][]*ServiceInstance),
	}
}

// Register registers a service instance
func (sr *ServiceRegistry) Register(instance *ServiceInstance) error {
	sr.mu.Lock()
	defer sr.mu.Unlock()

	if _, exists := sr.services[instance.Name]; !exists {
		sr.services[instance.Name] = []*ServiceInstance{}
	}

	sr.services[instance.Name] = append(sr.services[instance.Name], instance)
	return nil
}

// Deregister removes a service instance
func (sr *ServiceRegistry) Deregister(serviceName, instanceID string) error {
	sr.mu.Lock()
	defer sr.mu.Unlock()

	instances, exists := sr.services[serviceName]
	if !exists {
		return fmt.Errorf("service not found: %s", serviceName)
	}

	// Remove instance
	for i, inst := range instances {
		if inst.ID == instanceID {
			sr.services[serviceName] = append(instances[:i], instances[i+1:]...)
			return nil
		}
	}

	return fmt.Errorf("instance not found: %s", instanceID)
}

// GetHealthyInstances returns healthy instances of a service
func (sr *ServiceRegistry) GetHealthyInstances(serviceName string) []*ServiceInstance {
	sr.mu.RLock()
	defer sr.mu.RUnlock()

	instances, exists := sr.services[serviceName]
	if !exists {
		return []*ServiceInstance{}
	}

	healthy := make([]*ServiceInstance, 0)
	for _, inst := range instances {
		if inst.Healthy {
			healthy = append(healthy, inst)
		}
	}

	return healthy
}

// ========== Load Balancer ==========

// LoadBalancingStrategy defines balancing algorithm
type LoadBalancingStrategy string

const (
	StrategyRoundRobin LoadBalancingStrategy = "round_robin"
	StrategyRandom     LoadBalancingStrategy = "random"
	StrategyLeastConn  LoadBalancingStrategy = "least_conn"
	StrategyWeighted   LoadBalancingStrategy = "weighted"
)

// LoadBalancer distributes requests across instances
type LoadBalancer struct {
	registry  *ServiceRegistry
	strategy  LoadBalancingStrategy
	counter   int
	mu        sync.RWMutex
	connCounts map[string]int
}

// NewLoadBalancer creates a load balancer
func NewLoadBalancer(registry *ServiceRegistry, strategy LoadBalancingStrategy) *LoadBalancer {
	return &LoadBalancer{
		registry:   registry,
		strategy:   strategy,
		connCounts: make(map[string]int),
	}
}

// GetInstance selects an instance based on strategy
func (lb *LoadBalancer) GetInstance(serviceName string) (*ServiceInstance, error) {
	instances := lb.registry.GetHealthyInstances(serviceName)

	if len(instances) == 0 {
		return nil, fmt.Errorf("no healthy instances for service: %s", serviceName)
	}

	switch lb.strategy {
	case StrategyRoundRobin:
		return lb.roundRobin(instances)
	case StrategyRandom:
		return instances[rand.Intn(len(instances))], nil
	case StrategyLeastConn:
		return lb.leastConnections(instances), nil
	case StrategyWeighted:
		return lb.weightedSelection(instances), nil
	default:
		return instances[0], nil
	}
}

// roundRobin implements round-robin selection
func (lb *LoadBalancer) roundRobin(instances []*ServiceInstance) (*ServiceInstance, error) {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	instance := instances[lb.counter%len(instances)]
	lb.counter++

	return instance, nil
}

// leastConnections implements least connections selection
func (lb *LoadBalancer) leastConnections(instances []*ServiceInstance) *ServiceInstance {
	lb.mu.RLock()
	defer lb.mu.RUnlock()

	var selected *ServiceInstance
	minConns := math.MaxInt

	for _, inst := range instances {
		conns := lb.connCounts[inst.ID]
		if conns < minConns {
			minConns = conns
			selected = inst
		}
	}

	return selected
}

// weightedSelection implements weighted load balancing
func (lb *LoadBalancer) weightedSelection(instances []*ServiceInstance) *ServiceInstance {
	totalWeight := 0
	for _, inst := range instances {
		totalWeight += inst.Weight
	}

	if totalWeight == 0 {
		return instances[0]
	}

	choice := rand.Intn(totalWeight)
	current := 0

	for _, inst := range instances {
		current += inst.Weight
		if choice < current {
			return inst
		}
	}

	return instances[len(instances)-1]
}

// ========== API Gateway Pattern ==========

// APIGateway routes requests to appropriate services
type APIGateway struct {
	routes    map[string]string // path -> serviceName
	registry  *ServiceRegistry
	loadBalancer *LoadBalancer
	circuitBreakers map[string]*CircuitBreaker
	mu        sync.RWMutex
}

// NewAPIGateway creates an API gateway
func NewAPIGateway(registry *ServiceRegistry) *APIGateway {
	return &APIGateway{
		routes:          make(map[string]string),
		registry:        registry,
		loadBalancer:    NewLoadBalancer(registry, StrategyRoundRobin),
		circuitBreakers: make(map[string]*CircuitBreaker),
	}
}

// RegisterRoute maps a path to a service
func (gw *APIGateway) RegisterRoute(path, serviceName string) {
	gw.mu.Lock()
	defer gw.mu.Unlock()
	gw.routes[path] = serviceName
}

// RouteRequest routes a request to appropriate service
func (gw *APIGateway) RouteRequest(path string, fn func(*ServiceInstance) error) error {
	gw.mu.RLock()
	serviceName, exists := gw.routes[path]
	gw.mu.RUnlock()

	if !exists {
		return fmt.Errorf("no route for path: %s", path)
	}

	// Get or create circuit breaker
	gw.mu.Lock()
	cb, exists := gw.circuitBreakers[serviceName]
	if !exists {
		cb = NewCircuitBreaker(serviceName, 5, 2, 30*time.Second)
		gw.circuitBreakers[serviceName] = cb
	}
	gw.mu.Unlock()

	// Get instance
	instance, err := gw.loadBalancer.GetInstance(serviceName)
	if err != nil {
		return err
	}

	// Execute with circuit breaker
	return cb.Call(func() error {
		return fn(instance)
	})
}

// ========== Health Check ==========

// HealthChecker performs health checks on services
type HealthChecker struct {
	registry *ServiceRegistry
	interval time.Duration
	timeout  time.Duration
	mu       sync.RWMutex
	running  bool
}

// NewHealthChecker creates a health checker
func NewHealthChecker(registry *ServiceRegistry, interval, timeout time.Duration) *HealthChecker {
	return &HealthChecker{
		registry: registry,
		interval: interval,
		timeout:  timeout,
	}
}

// Start begins periodic health checks
func (hc *HealthChecker) Start() {
	hc.mu.Lock()
	if hc.running {
		hc.mu.Unlock()
		return
	}
	hc.running = true
	hc.mu.Unlock()

	go hc.checkLoop()
}

// Stop stops health checks
func (hc *HealthChecker) Stop() {
	hc.mu.Lock()
	hc.running = false
	hc.mu.Unlock()
}

// checkLoop runs periodic health checks
func (hc *HealthChecker) checkLoop() {
	ticker := time.NewTicker(hc.interval)
	defer ticker.Stop()

	for {
		hc.mu.RLock()
		if !hc.running {
			hc.mu.RUnlock()
			return
		}
		hc.mu.RUnlock()

		hc.checkAllServices()
		<-ticker.C
	}
}

// checkAllServices checks all registered services
func (hc *HealthChecker) checkAllServices() {
	// In production, would make actual HTTP calls
	// For now, simulate by marking as healthy
	for _, instances := range hc.registry.services {
		for _, inst := range instances {
			// Simulate health check (90% healthy)
			inst.Healthy = rand.Float32() > 0.1
		}
	}
}

func main() {}
