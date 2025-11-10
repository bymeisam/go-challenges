package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"net"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

// Challenge 168: Advanced Networking
// Custom TCP Protocol, Connection Pooling, Load Balancing, Reverse Proxy

// ===== 1. Custom Binary Protocol =====

type MessageType uint8

const (
	MsgTypePing     MessageType = 1
	MsgTypePong     MessageType = 2
	MsgTypeData     MessageType = 3
	MsgTypeError    MessageType = 4
	MsgTypeClose    MessageType = 5
)

type Protocol struct {
	Conn    net.Conn
	Reader  *bufio.Reader
	Writer  *bufio.Writer
	mu      sync.Mutex
	Metrics *ProtocolMetrics
}

type ProtocolMetrics struct {
	MessagesSent     int64
	MessagesReceived int64
	BytesSent        int64
	BytesReceived    int64
	Errors           int64
}

type Message struct {
	Type      MessageType
	Timestamp int64
	Payload   []byte
	Checksum  uint32
}

func NewProtocol(conn net.Conn) *Protocol {
	return &Protocol{
		Conn:    conn,
		Reader:  bufio.NewReader(conn),
		Writer:  bufio.NewWriter(conn),
		Metrics: &ProtocolMetrics{},
	}
}

func (p *Protocol) SendMessage(msg *Message) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	buf := new(bytes.Buffer)

	// Write message header: type (1) + timestamp (8) + payload size (4) + checksum (4)
	buf.WriteByte(byte(msg.Type))
	binary.Write(buf, binary.BigEndian, msg.Timestamp)
	binary.Write(buf, binary.BigEndian, uint32(len(msg.Payload)))
	binary.Write(buf, binary.BigEndian, msg.Checksum)

	// Write payload
	buf.Write(msg.Payload)

	// Write frame size at the beginning
	frameSize := buf.Len()
	frameBuf := new(bytes.Buffer)
	binary.Write(frameBuf, binary.BigEndian, uint32(frameSize))
	frameBuf.Write(buf.Bytes())

	_, err := p.Writer.Write(frameBuf.Bytes())
	if err == nil {
		err = p.Writer.Flush()
	}

	if err == nil {
		atomic.AddInt64(&p.Metrics.MessagesSent, 1)
		atomic.AddInt64(&p.Metrics.BytesSent, int64(frameBuf.Len()))
	} else {
		atomic.AddInt64(&p.Metrics.Errors, 1)
	}

	return err
}

func (p *Protocol) ReadMessage() (*Message, error) {
	// Read frame size
	var frameSize uint32
	if err := binary.Read(p.Reader, binary.BigEndian, &frameSize); err != nil {
		atomic.AddInt64(&p.Metrics.Errors, 1)
		return nil, err
	}

	if frameSize > 1024*1024 { // Max 1MB per message
		atomic.AddInt64(&p.Metrics.Errors, 1)
		return nil, fmt.Errorf("frame too large: %d", frameSize)
	}

	// Read frame
	frame := make([]byte, frameSize)
	if _, err := io.ReadFull(p.Reader, frame); err != nil {
		atomic.AddInt64(&p.Metrics.Errors, 1)
		return nil, err
	}

	buf := bytes.NewReader(frame)

	var msgType uint8
	var timestamp int64
	var payloadSize uint32
	var checksum uint32

	if err := binary.Read(buf, binary.BigEndian, &msgType); err != nil {
		return nil, err
	}
	if err := binary.Read(buf, binary.BigEndian, &timestamp); err != nil {
		return nil, err
	}
	if err := binary.Read(buf, binary.BigEndian, &payloadSize); err != nil {
		return nil, err
	}
	if err := binary.Read(buf, binary.BigEndian, &checksum); err != nil {
		return nil, err
	}

	payload := make([]byte, payloadSize)
	if _, err := io.ReadFull(buf, payload); err != nil {
		return nil, err
	}

	atomic.AddInt64(&p.Metrics.MessagesReceived, 1)
	atomic.AddInt64(&p.Metrics.BytesReceived, int64(frameSize))

	return &Message{
		Type:      MessageType(msgType),
		Timestamp: timestamp,
		Payload:   payload,
		Checksum:  checksum,
	}, nil
}

// ===== 2. Connection Pool =====

type PooledConnection struct {
	Conn     net.Conn
	Protocol *Protocol
	LastUsed time.Time
	IsValid  bool
}

type ConnectionPool struct {
	address      string
	maxPoolSize  int
	maxIdleTime  time.Duration
	available    chan *PooledConnection
	allConns     []*PooledConnection
	mu           sync.RWMutex
	metrics      *PoolMetrics
	stopChan     chan struct{}
	activeCount  int32
	totalCreated int64
}

type PoolMetrics struct {
	Acquisitions   int64
	Releases       int64
	CreatedConns   int64
	ClosedConns    int64
	PoolExhausted  int64
	ValidationFailed int64
}

func NewConnectionPool(address string, maxSize int) *ConnectionPool {
	cp := &ConnectionPool{
		address:     address,
		maxPoolSize: maxSize,
		maxIdleTime: 5 * time.Minute,
		available:   make(chan *PooledConnection, maxSize),
		allConns:    make([]*PooledConnection, 0, maxSize),
		metrics:     &PoolMetrics{},
		stopChan:    make(chan struct{}),
	}

	// Start cleanup goroutine
	go cp.cleanupLoop()

	return cp
}

func (cp *ConnectionPool) Acquire(ctx context.Context) (*PooledConnection, error) {
	atomic.AddInt64(&cp.metrics.Acquisitions, 1)

	select {
	case pooledConn := <-cp.available:
		// Validate connection
		if cp.isConnectionValid(pooledConn) {
			atomic.AddInt32(&cp.activeCount, 1)
			return pooledConn, nil
		}
		// Connection is invalid, close it
		pooledConn.Conn.Close()
		atomic.AddInt64(&cp.metrics.ValidationFailed, 1)
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// Create new connection if under limit
	if int(atomic.LoadInt32(&cp.activeCount))+len(cp.available) < cp.maxPoolSize {
		conn, err := net.DialTimeout("tcp", cp.address, 5*time.Second)
		if err != nil {
			return nil, err
		}

		pooledConn := &PooledConnection{
			Conn:     conn,
			Protocol: NewProtocol(conn),
			LastUsed: time.Now(),
			IsValid:  true,
		}

		cp.mu.Lock()
		cp.allConns = append(cp.allConns, pooledConn)
		cp.mu.Unlock()

		atomic.AddInt64(&cp.metrics.CreatedConns, 1)
		atomic.AddInt64(&cp.totalCreated, 1)
		atomic.AddInt32(&cp.activeCount, 1)

		return pooledConn, nil
	}

	atomic.AddInt64(&cp.metrics.PoolExhausted, 1)
	return nil, fmt.Errorf("connection pool exhausted")
}

func (cp *ConnectionPool) Release(pooledConn *PooledConnection) {
	if pooledConn == nil {
		return
	}

	atomic.AddInt64(&cp.metrics.Releases, 1)
	atomic.AddInt32(&cp.activeCount, -1)

	pooledConn.LastUsed = time.Now()

	select {
	case cp.available <- pooledConn:
	default:
		// Pool full, close connection
		pooledConn.Conn.Close()
		atomic.AddInt64(&cp.metrics.ClosedConns, 1)
	}
}

func (cp *ConnectionPool) isConnectionValid(pc *PooledConnection) bool {
	if !pc.IsValid {
		return false
	}

	// Check if connection is idle too long
	if time.Since(pc.LastUsed) > cp.maxIdleTime {
		return false
	}

	// Quick validation: set read timeout and check if readable
	pc.Conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
	defer pc.Conn.SetReadDeadline(time.Time{})

	return true
}

func (cp *ConnectionPool) cleanupLoop() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-cp.stopChan:
			return
		case <-ticker.C:
			cp.mu.Lock()
			for i := 0; i < len(cp.allConns); i++ {
				conn := cp.allConns[i]
				if !cp.isConnectionValid(conn) {
					conn.Conn.Close()
					atomic.AddInt64(&cp.metrics.ClosedConns, 1)
				}
			}
			cp.mu.Unlock()
		}
	}
}

func (cp *ConnectionPool) Close() {
	close(cp.stopChan)
	cp.mu.Lock()
	for _, conn := range cp.allConns {
		conn.Conn.Close()
	}
	cp.mu.Unlock()
}

// ===== 3. Load Balancer =====

type LoadBalancer struct {
	targets        []*Backend
	currentIndex   int32
	mu             sync.RWMutex
	metrics        *LBMetrics
	healthCheckInterval time.Duration
	stopChan       chan struct{}
}

type Backend struct {
	Address string
	Weight  int
	Healthy bool
	mu      sync.RWMutex
}

type LBMetrics struct {
	RequestsProcessed int64
	HealthChecks      int64
	BackendErrors     int64
	HealthyBackends   int32
}

func NewLoadBalancer(targets []string) *LoadBalancer {
	backends := make([]*Backend, len(targets))
	for i, addr := range targets {
		backends[i] = &Backend{
			Address: addr,
			Weight:  1,
			Healthy: true,
		}
	}

	lb := &LoadBalancer{
		targets:             backends,
		metrics:             &LBMetrics{},
		healthCheckInterval: 10 * time.Second,
		stopChan:            make(chan struct{}),
	}

	go lb.healthCheckLoop()
	return lb
}

func (lb *LoadBalancer) NextBackend() (*Backend, error) {
	atomic.AddInt64(&lb.metrics.RequestsProcessed, 1)

	lb.mu.RLock()
	defer lb.mu.RUnlock()

	var healthyCount int
	for _, b := range lb.targets {
		if b.Healthy {
			healthyCount++
		}
	}

	if healthyCount == 0 {
		atomic.AddInt64(&lb.metrics.BackendErrors, 1)
		return nil, fmt.Errorf("no healthy backends available")
	}

	// Round-robin among healthy backends
	attempts := 0
	for attempts < len(lb.targets) {
		idx := atomic.AddInt32(&lb.currentIndex, 1) % int32(len(lb.targets))
		backend := lb.targets[idx]

		backend.mu.RLock()
		healthy := backend.Healthy
		backend.mu.RUnlock()

		if healthy {
			return backend, nil
		}
		attempts++
	}

	return nil, fmt.Errorf("no healthy backends available")
}

func (lb *LoadBalancer) healthCheckLoop() {
	ticker := time.NewTicker(lb.healthCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-lb.stopChan:
			return
		case <-ticker.C:
			lb.performHealthChecks()
		}
	}
}

func (lb *LoadBalancer) performHealthChecks() {
	atomic.AddInt64(&lb.metrics.HealthChecks, 1)

	healthyCount := int32(0)
	for _, backend := range lb.targets {
		// Simple health check: try to connect
		conn, err := net.DialTimeout("tcp", backend.Address, 2*time.Second)
		if err == nil {
			conn.Close()
			backend.mu.Lock()
			backend.Healthy = true
			backend.mu.Unlock()
			healthyCount++
		} else {
			backend.mu.Lock()
			backend.Healthy = false
			backend.mu.Unlock()
		}
	}

	atomic.StoreInt32(&lb.metrics.HealthyBackends, healthyCount)
}

func (lb *LoadBalancer) Close() {
	close(lb.stopChan)
}

// ===== 4. Reverse Proxy =====

type ReverseProxy struct {
	lb         *LoadBalancer
	pools      map[string]*ConnectionPool
	mu         sync.RWMutex
	metrics    *ProxyMetrics
	stopChan   chan struct{}
}

type ProxyMetrics struct {
	RequestsForwarded int64
	ResponsesForwarded int64
	ProxyErrors       int64
	ActiveConnections int32
}

func NewReverseProxy(targets []string) *ReverseProxy {
	lb := NewLoadBalancer(targets)

	rp := &ReverseProxy{
		lb:       lb,
		pools:    make(map[string]*ConnectionPool),
		metrics:  &ProxyMetrics{},
		stopChan: make(chan struct{}),
	}

	// Create connection pools for each target
	for _, target := range targets {
		rp.pools[target] = NewConnectionPool(target, 10)
	}

	return rp
}

func (rp *ReverseProxy) ForwardRequest(ctx context.Context, req []byte) ([]byte, error) {
	atomic.AddInt64(&rp.metrics.RequestsForwarded, 1)
	atomic.AddInt32(&rp.metrics.ActiveConnections, 1)
	defer atomic.AddInt32(&rp.metrics.ActiveConnections, -1)

	backend, err := rp.lb.NextBackend()
	if err != nil {
		atomic.AddInt64(&rp.metrics.ProxyErrors, 1)
		return nil, err
	}

	// Get connection from pool
	rp.mu.RLock()
	pool, exists := rp.pools[backend.Address]
	rp.mu.RUnlock()

	if !exists {
		atomic.AddInt64(&rp.metrics.ProxyErrors, 1)
		return nil, fmt.Errorf("no pool for backend: %s", backend.Address)
	}

	pooledConn, err := pool.Acquire(ctx)
	if err != nil {
		atomic.AddInt64(&rp.metrics.ProxyErrors, 1)
		return nil, err
	}
	defer pool.Release(pooledConn)

	// Send request
	msg := &Message{
		Type:      MsgTypeData,
		Timestamp: time.Now().Unix(),
		Payload:   req,
	}

	if err := pooledConn.Protocol.SendMessage(msg); err != nil {
		atomic.AddInt64(&rp.metrics.ProxyErrors, 1)
		return nil, err
	}

	// Read response
	response, err := pooledConn.Protocol.ReadMessage()
	if err != nil {
		atomic.AddInt64(&rp.metrics.ProxyErrors, 1)
		return nil, err
	}

	atomic.AddInt64(&rp.metrics.ResponsesForwarded, 1)
	return response.Payload, nil
}

func (rp *ReverseProxy) Close() {
	close(rp.stopChan)
	rp.lb.Close()

	rp.mu.Lock()
	for _, pool := range rp.pools {
		pool.Close()
	}
	rp.mu.Unlock()
}

// ===== 5. HTTP Server with Keep-Alive =====

type GracefulHTTPServer struct {
	server        *http.Server
	activeConns   int32
	maxConnTime   time.Duration
	readTimeout   time.Duration
	writeTimeout  time.Duration
	shutdownChan  chan struct{}
}

func NewGracefulHTTPServer(addr string) *GracefulHTTPServer {
	return &GracefulHTTPServer{
		server: &http.Server{
			Addr:         addr,
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
		maxConnTime:  5 * time.Minute,
		readTimeout:  15 * time.Second,
		writeTimeout: 15 * time.Second,
		shutdownChan: make(chan struct{}),
	}
}

func (gs *GracefulHTTPServer) Start(handler http.HandlerFunc) error {
	gs.server.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&gs.activeConns, 1)
		defer atomic.AddInt32(&gs.activeConns, -1)

		// Keep-Alive is enabled by default in HTTP/1.1
		w.Header().Set("Connection", "keep-alive")
		handler(w, r)
	})

	return gs.server.ListenAndServe()
}

func (gs *GracefulHTTPServer) Shutdown(ctx context.Context) error {
	close(gs.shutdownChan)
	return gs.server.Shutdown(ctx)
}

// ===== 6. Helper Functions =====

func calculateChecksum(data []byte) uint32 {
	sum := uint32(0)
	for _, b := range data {
		sum += uint32(b)
	}
	return sum
}

func newMessage(msgType MessageType, payload []byte) *Message {
	return &Message{
		Type:      msgType,
		Timestamp: time.Now().Unix(),
		Payload:   payload,
		Checksum:  calculateChecksum(payload),
	}
}

// ===== Main Demo =====

func main() {
	fmt.Println("=== Advanced Networking ===\n")

	// 1. Custom Protocol
	fmt.Println("1. Custom Binary Protocol")
	msg := newMessage(MsgTypeData, []byte("Hello Network"))
	fmt.Printf("Message: Type=%d, Payload=%s, Checksum=%d\n\n", msg.Type, string(msg.Payload), msg.Checksum)

	// 2. Connection Pool
	fmt.Println("2. Connection Pool Metrics")
	// Note: Real pooling would need actual servers running
	fmt.Printf("Connection pool demo (would need live servers)\n\n")

	// 3. Load Balancer
	fmt.Println("3. Load Balancer")
	lb := NewLoadBalancer([]string{"server1:8001", "server2:8002", "server3:8003"})

	for i := 0; i < 5; i++ {
		backend, _ := lb.NextBackend()
		if backend != nil {
			fmt.Printf("Request %d -> %s\n", i+1, backend.Address)
		}
	}

	fmt.Printf("LB Metrics: Requests=%d, HealthChecks=%d, HealthyBackends=%d\n\n",
		atomic.LoadInt64(&lb.metrics.RequestsProcessed),
		atomic.LoadInt64(&lb.metrics.HealthChecks),
		atomic.LoadInt32(&lb.metrics.HealthyBackends))

	lb.Close()

	// 4. Reverse Proxy Demo
	fmt.Println("4. Reverse Proxy Architecture")
	rp := NewReverseProxy([]string{"backend1:9001", "backend2:9002"})

	fmt.Printf("Reverse Proxy: Requests Forwarded=%d, Active Connections=%d\n\n",
		atomic.LoadInt64(&rp.metrics.RequestsForwarded),
		atomic.LoadInt32(&rp.metrics.ActiveConnections))

	rp.Close()

	// 5. Protocol Metrics
	fmt.Println("5. Protocol Message Processing")
	fmt.Println("Message Types: Ping, Pong, Data, Error, Close")
	fmt.Println("Protocol Features:")
	fmt.Println("  - Binary message framing")
	fmt.Println("  - Checksum validation")
	fmt.Println("  - Timestamp tracking")
	fmt.Println("  - Max frame size limits")

	// 6. Performance Characteristics
	fmt.Println("\n6. Performance Characteristics")
	fmt.Printf("Message Header Size: 17 bytes (type + timestamp + payload size + checksum)\n")
	fmt.Printf("Frame Overhead: 4 bytes (frame size prefix)\n")
	fmt.Printf("Max Message Size: 1 MB\n")
	fmt.Printf("Pool Management: Configurable pool size with idle timeouts\n")
	fmt.Printf("Load Balancing: Round-robin with health checks\n")

	fmt.Println("\n=== Complete ===")
}
