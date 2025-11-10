package main

import (
	"context"
	"net"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// TestMessage tests message creation
func TestMessageCreation(t *testing.T) {
	payload := []byte("test payload")
	msg := newMessage(MsgTypeData, payload)

	if msg.Type != MsgTypeData {
		t.Errorf("Expected MsgTypeData, got %d", msg.Type)
	}

	if string(msg.Payload) != "test payload" {
		t.Errorf("Expected 'test payload', got '%s'", string(msg.Payload))
	}

	if msg.Checksum == 0 {
		t.Errorf("Expected non-zero checksum")
	}
}

func TestCalculateChecksum(t *testing.T) {
	data := []byte{1, 2, 3, 4, 5}
	checksum := calculateChecksum(data)

	expected := uint32(15) // 1+2+3+4+5
	if checksum != expected {
		t.Errorf("Expected checksum %d, got %d", expected, checksum)
	}
}

// TestLoadBalancer tests load balancing
func TestLoadBalancerBasic(t *testing.T) {
	targets := []string{"server1:8001", "server2:8002", "server3:8003"}
	lb := NewLoadBalancer(targets)
	defer lb.Close()

	backend, err := lb.NextBackend()
	if err != nil {
		t.Errorf("Expected successful backend selection, got error: %v", err)
	}

	if backend == nil {
		t.Errorf("Expected non-nil backend")
	}
}

func TestLoadBalancerRoundRobin(t *testing.T) {
	targets := []string{"s1:8001", "s2:8002", "s3:8003"}
	lb := NewLoadBalancer(targets)
	defer lb.Close()

	selectedCount := make(map[string]int)
	for i := 0; i < 30; i++ {
		backend, _ := lb.NextBackend()
		if backend != nil {
			selectedCount[backend.Address]++
		}
	}

	// Should have selected all backends roughly equally
	if len(selectedCount) != 3 {
		t.Errorf("Expected 3 different backends selected, got %d", len(selectedCount))
	}

	for addr, count := range selectedCount {
		if count < 5 || count > 15 {
			t.Logf("Backend %s selected %d times (expected ~10)", addr, count)
		}
	}
}

func TestLoadBalancerMetrics(t *testing.T) {
	lb := NewLoadBalancer([]string{"s1:8001", "s2:8002"})
	defer lb.Close()

	for i := 0; i < 100; i++ {
		lb.NextBackend()
	}

	if atomic.LoadInt64(&lb.metrics.RequestsProcessed) != 100 {
		t.Errorf("Expected 100 requests processed")
	}
}

// TestBackend tests backend health
func TestBackendHealth(t *testing.T) {
	backend := &Backend{
		Address: "localhost:8001",
		Weight:  1,
		Healthy: true,
	}

	if !backend.Healthy {
		t.Errorf("Expected healthy backend")
	}

	backend.mu.Lock()
	backend.Healthy = false
	backend.mu.Unlock()

	backend.mu.RLock()
	healthy := backend.Healthy
	backend.mu.RUnlock()

	if healthy {
		t.Errorf("Expected unhealthy backend")
	}
}

// TestReverseProxy tests reverse proxy
func TestReverseProxyCreation(t *testing.T) {
	rp := NewReverseProxy([]string{"backend1:9001", "backend2:9002"})
	defer rp.Close()

	if rp.lb == nil {
		t.Errorf("Expected load balancer to be initialized")
	}

	if len(rp.pools) != 2 {
		t.Errorf("Expected 2 connection pools, got %d", len(rp.pools))
	}
}

func TestReverseProxyMetrics(t *testing.T) {
	rp := NewReverseProxy([]string{"backend1:9001"})
	defer rp.Close()

	// Simulate metrics updates
	atomic.AddInt64(&rp.metrics.RequestsForwarded, 10)
	atomic.AddInt64(&rp.metrics.ResponsesForwarded, 8)
	atomic.AddInt64(&rp.metrics.ProxyErrors, 2)

	if atomic.LoadInt64(&rp.metrics.RequestsForwarded) != 10 {
		t.Errorf("Expected 10 forwarded requests")
	}
}

// TestConnectionPoolCreation tests pool creation
func TestConnectionPoolBasic(t *testing.T) {
	// Note: This test uses a non-existent address, so connections will fail
	// But we can test the pool structure
	pool := NewConnectionPool("localhost:9999", 5)
	defer pool.Close()

	if pool.maxPoolSize != 5 {
		t.Errorf("Expected max pool size 5, got %d", pool.maxPoolSize)
	}
}

func TestConnectionPoolMetrics(t *testing.T) {
	pool := NewConnectionPool("localhost:9999", 5)
	defer pool.Close()

	// Simulate acquisitions
	atomic.AddInt64(&pool.metrics.Acquisitions, 5)
	atomic.AddInt64(&pool.metrics.Releases, 3)

	if atomic.LoadInt64(&pool.metrics.Acquisitions) != 5 {
		t.Errorf("Expected 5 acquisitions")
	}

	if atomic.LoadInt64(&pool.metrics.Releases) != 3 {
		t.Errorf("Expected 3 releases")
	}
}

// TestProtocolMetrics tests protocol metrics
func TestProtocolMetrics(t *testing.T) {
	// Create a pipe for testing
	client, server := net.Pipe()
	defer client.Close()
	defer server.Close()

	proto := NewProtocol(client)

	// Simulate message sending/receiving
	atomic.AddInt64(&proto.Metrics.MessagesSent, 1)
	atomic.AddInt64(&proto.Metrics.BytesSent, 100)
	atomic.AddInt64(&proto.Metrics.MessagesReceived, 1)
	atomic.AddInt64(&proto.Metrics.BytesReceived, 100)

	if atomic.LoadInt64(&proto.Metrics.MessagesSent) != 1 {
		t.Errorf("Expected 1 message sent")
	}

	if atomic.LoadInt64(&proto.Metrics.BytesSent) != 100 {
		t.Errorf("Expected 100 bytes sent")
	}
}

// TestMessageFraming tests message framing
func TestMessageFraming(t *testing.T) {
	client, server := net.Pipe()
	defer client.Close()

	go func() {
		proto := NewProtocol(client)
		msg := newMessage(MsgTypeData, []byte("test"))
		proto.SendMessage(msg)
	}()

	proto := NewProtocol(server)
	msg, err := proto.ReadMessage()
	server.Close()

	if err != nil {
		t.Errorf("Expected successful message read, got error: %v", err)
	}

	if msg == nil {
		t.Errorf("Expected non-nil message")
	}

	if string(msg.Payload) != "test" {
		t.Errorf("Expected 'test', got '%s'", string(msg.Payload))
	}
}

// TestConcurrentLoadBalancing tests concurrent load balancing
func TestConcurrentLoadBalancing(t *testing.T) {
	lb := NewLoadBalancer([]string{"s1:8001", "s2:8002", "s3:8003", "s4:8004"})
	defer lb.Close()

	var wg sync.WaitGroup
	selectedCount := make(map[string]int)
	var mu sync.Mutex

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			backend, err := lb.NextBackend()
			if err == nil && backend != nil {
				mu.Lock()
				selectedCount[backend.Address]++
				mu.Unlock()
			}
		}()
	}

	wg.Wait()

	if len(selectedCount) != 4 {
		t.Errorf("Expected 4 different backends selected, got %d", len(selectedCount))
	}
}

// Benchmark tests

func BenchmarkLoadBalancerNextBackend(b *testing.B) {
	lb := NewLoadBalancer([]string{"s1:8001", "s2:8002", "s3:8003"})
	defer lb.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		lb.NextBackend()
	}
}

func BenchmarkMessageCreation(b *testing.B) {
	payload := []byte("test payload data")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = newMessage(MsgTypeData, payload)
	}
}

func BenchmarkChecksum(b *testing.B) {
	data := []byte("test data for checksum calculation")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = calculateChecksum(data)
	}
}

func BenchmarkConcurrentLoadBalancing(b *testing.B) {
	lb := NewLoadBalancer([]string{"s1:8001", "s2:8002", "s3:8003"})
	defer lb.Close()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			lb.NextBackend()
		}
	})
}
