package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sync"
	"time"
)

// ========== Idempotency Models ==========

// IdempotencyKey represents a unique idempotency key
type IdempotencyKey struct {
	Key       string
	CreatedAt time.Time
	ExpiresAt time.Time
	Response  interface{}
	Status    string // PENDING, SUCCESS, FAILED
	Error     string
}

// IdempotencyStore manages idempotent requests
type IdempotencyStore struct {
	keys map[string]*IdempotencyKey
	mu   sync.RWMutex
	ttl  time.Duration
}

// ========== Distributed Lock Models ==========

// LockStatus represents the status of a lock
type LockStatus string

const (
	LockAcquired LockStatus = "ACQUIRED"
	LockWaiting  LockStatus = "WAITING"
	LockFailed   LockStatus = "FAILED"
)

// DistributedLock represents a distributed lock
type DistributedLock struct {
	LockID        string
	ResourceID    string
	OwnerID       string
	AcquiredAt    time.Time
	ExpiresAt     time.Time
	LeaseID       string
	Status        LockStatus
	WaitersCount  int
	ContentionMsg string
}

// LockManager manages distributed locks
type LockManager struct {
	locks       map[string]*DistributedLock
	locksMu     sync.RWMutex
	waitQueues  map[string][]*LockRequest
	queueMu     sync.RWMutex
	metrics     *LockMetrics
	leaseTime   time.Duration
	maxWaiters  int
}

// LockRequest represents a request to acquire a lock
type LockRequest struct {
	RequestID   string
	ResourceID  string
	OwnerID     string
	Timestamp   time.Time
	RetryCount  int
	MaxRetries  int
	WaitTime    time.Duration
}

// LockMetrics tracks lock usage metrics
type LockMetrics struct {
	TotalAcquisitions  int64
	TotalReleases      int64
	TotalTimeouts      int64
	TotalDeadlocks     int64
	AvgWaitTime        time.Duration
	MaxWaitTime        time.Duration
	CurrentContention  int
	mu                 sync.RWMutex
}

// ========== Leader Election Models ==========

// LeaderState represents the state of a node in leader election
type LeaderState struct {
	NodeID       string
	IsLeader     bool
	Term         int64
	LeaderID     string
	LastHeartbeat time.Time
	Votes        map[string]bool
}

// LeaderElection manages leader election across nodes
type LeaderElection struct {
	nodes        map[string]*LeaderState
	nodesMu      sync.RWMutex
	currentTerm  int64
	votedFor     string
	log          []*LogEntry
	electionMu   sync.RWMutex
	heartbeatTTL time.Duration
}

// LogEntry represents an entry in the election log
type LogEntry struct {
	Term      int64
	Index     int64
	NodeID    string
	Action    string
	Timestamp time.Time
}

// ========== Idempotency Store Implementation ==========

// NewIdempotencyStore creates a new idempotency store
func NewIdempotencyStore(ttl time.Duration) *IdempotencyStore {
	store := &IdempotencyStore{
		keys: make(map[string]*IdempotencyKey),
		ttl:  ttl,
	}

	// Cleanup expired keys periodically
	go store.cleanupExpired()

	return store
}

// cleanupExpired removes expired idempotency keys
func (is *IdempotencyStore) cleanupExpired() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		is.mu.Lock()
		now := time.Now()
		for key, ikey := range is.keys {
			if now.After(ikey.ExpiresAt) {
				delete(is.keys, key)
			}
		}
		is.mu.Unlock()
	}
}

// GenerateKey creates a new idempotency key from input
func (is *IdempotencyStore) GenerateKey(input string) string {
	hash := sha256.Sum256([]byte(input))
	return hex.EncodeToString(hash[:])
}

// StoreRequest stores a request for idempotency
func (is *IdempotencyStore) StoreRequest(key string) error {
	is.mu.Lock()
	defer is.mu.Unlock()

	if _, exists := is.keys[key]; exists {
		return errors.New("key already exists")
	}

	is.keys[key] = &IdempotencyKey{
		Key:       key,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(is.ttl),
		Status:    "PENDING",
	}

	return nil
}

// UpdateResponse updates the response for a key
func (is *IdempotencyStore) UpdateResponse(key string, response interface{}, err error) error {
	is.mu.Lock()
	defer is.mu.Unlock()

	ikey, exists := is.keys[key]
	if !exists {
		return errors.New("key not found")
	}

	ikey.Response = response
	ikey.Status = "SUCCESS"

	if err != nil {
		ikey.Status = "FAILED"
		ikey.Error = err.Error()
	}

	return nil
}

// GetResponse retrieves a cached response
func (is *IdempotencyStore) GetResponse(key string) (interface{}, bool) {
	is.mu.RLock()
	defer is.mu.RUnlock()

	if ikey, exists := is.keys[key]; exists && time.Now().Before(ikey.ExpiresAt) {
		return ikey.Response, ikey.Status == "SUCCESS"
	}

	return nil, false
}

// ValidateKey checks if a key is still valid
func (is *IdempotencyStore) ValidateKey(key string) bool {
	is.mu.RLock()
	defer is.mu.RUnlock()

	ikey, exists := is.keys[key]
	return exists && time.Now().Before(ikey.ExpiresAt)
}

// GetKeyStatus retrieves the status of a key
func (is *IdempotencyStore) GetKeyStatus(key string) string {
	is.mu.RLock()
	defer is.mu.RUnlock()

	if ikey, exists := is.keys[key]; exists {
		return ikey.Status
	}
	return "NOT_FOUND"
}

// ========== Lock Manager Implementation ==========

// NewLockManager creates a new lock manager
func NewLockManager(leaseTime time.Duration) *LockManager {
	return &LockManager{
		locks:      make(map[string]*DistributedLock),
		waitQueues: make(map[string][]*LockRequest),
		metrics:    &LockMetrics{},
		leaseTime:  leaseTime,
		maxWaiters: 100,
	}
}

// AcquireLock attempts to acquire a distributed lock
func (lm *LockManager) AcquireLock(ctx context.Context, resourceID, ownerID string, timeout time.Duration) (*DistributedLock, error) {
	lockID := fmt.Sprintf("lock:%s", resourceID)
	leaseID := generateLockID()

	lm.locksMu.Lock()
	defer lm.locksMu.Unlock()

	// Check if lock is available
	if lock, exists := lm.locks[lockID]; exists {
		if time.Now().Before(lock.ExpiresAt) && lock.OwnerID != ownerID {
			// Lock is held by someone else
			lm.addToWaitQueue(resourceID, ownerID)
			lm.metrics.mu.Lock()
			lm.metrics.CurrentContention++
			lm.metrics.mu.Unlock()

			return nil, errors.New("lock already held")
		}

		// Lease expired, take over
		if time.Now().After(lock.ExpiresAt) {
			lm.locks[lockID] = &DistributedLock{
				LockID:       lockID,
				ResourceID:   resourceID,
				OwnerID:      ownerID,
				AcquiredAt:   time.Now(),
				ExpiresAt:    time.Now().Add(lm.leaseTime),
				LeaseID:      leaseID,
				Status:       LockAcquired,
			}

			lm.metrics.recordAcquisition()
			return lm.locks[lockID], nil
		}

		// Same owner, renew lease
		if lock.OwnerID == ownerID {
			lock.ExpiresAt = time.Now().Add(lm.leaseTime)
			lock.LeaseID = leaseID
			return lock, nil
		}

		return nil, errors.New("lock contention")
	}

	// Create new lock
	lock := &DistributedLock{
		LockID:     lockID,
		ResourceID: resourceID,
		OwnerID:    ownerID,
		AcquiredAt: time.Now(),
		ExpiresAt:  time.Now().Add(lm.leaseTime),
		LeaseID:    leaseID,
		Status:     LockAcquired,
	}

	lm.locks[lockID] = lock
	lm.metrics.recordAcquisition()

	return lock, nil
}

// ReleaseLock releases a distributed lock
func (lm *LockManager) ReleaseLock(lockID, ownerID string) error {
	lm.locksMu.Lock()
	defer lm.locksMu.Unlock()

	lock, exists := lm.locks[lockID]
	if !exists {
		return errors.New("lock not found")
	}

	if lock.OwnerID != ownerID {
		return errors.New("lock not owned by requester")
	}

	delete(lm.locks, lockID)
	lm.metrics.recordRelease()

	// Process waiting requests
	lm.queueMu.Lock()
	if queue, exists := lm.waitQueues[lock.ResourceID]; exists && len(queue) > 0 {
		// Notify next waiter
		nextRequest := queue[0]
		lm.waitQueues[lock.ResourceID] = queue[1:]
		lm.metrics.mu.Lock()
		lm.metrics.CurrentContention--
		lm.metrics.mu.Unlock()
		_ = nextRequest // Could trigger callback
	}
	lm.queueMu.Unlock()

	return nil
}

// RenewLease renews an acquired lock's lease
func (lm *LockManager) RenewLease(lockID, ownerID string) error {
	lm.locksMu.Lock()
	defer lm.locksMu.Unlock()

	lock, exists := lm.locks[lockID]
	if !exists {
		return errors.New("lock not found")
	}

	if lock.OwnerID != ownerID {
		return errors.New("lock not owned by requester")
	}

	if time.Now().After(lock.ExpiresAt) {
		return errors.New("lock lease expired")
	}

	lock.ExpiresAt = time.Now().Add(lm.leaseTime)
	return nil
}

// CheckLock checks the status of a lock
func (lm *LockManager) CheckLock(lockID string) *DistributedLock {
	lm.locksMu.RLock()
	defer lm.locksMu.RUnlock()

	if lock, exists := lm.locks[lockID]; exists {
		if time.Now().Before(lock.ExpiresAt) {
			return lock
		}
	}

	return nil
}

// DetectDeadlock detects potential deadlock cycles
func (lm *LockManager) DetectDeadlock() []string {
	lm.locksMu.RLock()
	defer lm.locksMu.RUnlock()

	deadlocks := []string{}
	visited := make(map[string]bool)
	recStack := make(map[string]bool)

	for lockID := range lm.locks {
		if !visited[lockID] {
			if lm.hasCycle(lockID, visited, recStack) {
				deadlocks = append(deadlocks, lockID)
			}
		}
	}

	if len(deadlocks) > 0 {
		lm.metrics.mu.Lock()
		lm.metrics.TotalDeadlocks++
		lm.metrics.mu.Unlock()
	}

	return deadlocks
}

func (lm *LockManager) hasCycle(lockID string, visited, recStack map[string]bool) bool {
	visited[lockID] = true
	recStack[lockID] = true

	// Simplified cycle detection
	lock := lm.locks[lockID]
	for otherID, otherLock := range lm.locks {
		if otherLock.OwnerID == lock.OwnerID && otherID != lockID {
			if recStack[otherID] {
				return true
			}
			if !visited[otherID] && lm.hasCycle(otherID, visited, recStack) {
				return true
			}
		}
	}

	recStack[lockID] = false
	return false
}

// GetMetrics retrieves lock metrics
func (lm *LockManager) GetMetrics() *LockMetrics {
	lm.metrics.mu.RLock()
	defer lm.metrics.mu.RUnlock()
	return lm.metrics
}

func (lm *LockManager) addToWaitQueue(resourceID, ownerID string) {
	lm.queueMu.Lock()
	defer lm.queueMu.Unlock()

	lm.waitQueues[resourceID] = append(lm.waitQueues[resourceID], &LockRequest{
		RequestID:  generateLockID(),
		ResourceID: resourceID,
		OwnerID:    ownerID,
		Timestamp:  time.Now(),
		MaxRetries: 3,
	})
}

// ========== Metrics Recording ==========

func (m *LockMetrics) recordAcquisition() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.TotalAcquisitions++
}

func (m *LockMetrics) recordRelease() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.TotalReleases++
}

func (m *LockMetrics) recordWaitTime(d time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if d > m.MaxWaitTime {
		m.MaxWaitTime = d
	}

	// Simple average calculation
	if m.AvgWaitTime == 0 {
		m.AvgWaitTime = d
	} else {
		m.AvgWaitTime = (m.AvgWaitTime + d) / 2
	}
}

// ========== Leader Election Implementation ==========

// NewLeaderElection creates a new leader election coordinator
func NewLeaderElection(heartbeatTTL time.Duration) *LeaderElection {
	return &LeaderElection{
		nodes:        make(map[string]*LeaderState),
		currentTerm:  0,
		votedFor:     "",
		log:          []*LogEntry{},
		heartbeatTTL: heartbeatTTL,
	}
}

// RegisterNode registers a node for leader election
func (le *LeaderElection) RegisterNode(nodeID string) error {
	le.nodesMu.Lock()
	defer le.nodesMu.Unlock()

	if _, exists := le.nodes[nodeID]; exists {
		return errors.New("node already registered")
	}

	le.nodes[nodeID] = &LeaderState{
		NodeID:        nodeID,
		IsLeader:      false,
		Term:          0,
		LastHeartbeat: time.Now(),
		Votes:         make(map[string]bool),
	}

	return nil
}

// CastVote casts a vote for a candidate
func (le *LeaderElection) CastVote(voterID, candidateID string) error {
	le.electionMu.Lock()
	defer le.electionMu.Unlock()

	le.nodesMu.Lock()
	voter, voterExists := le.nodes[voterID]
	candidate, candidateExists := le.nodes[candidateID]
	le.nodesMu.Unlock()

	if !voterExists || !candidateExists {
		return errors.New("node not found")
	}

	voter.Votes[candidateID] = true
	candidate.Votes[candidateID] = true

	return nil
}

// ElectLeader performs leader election
func (le *LeaderElection) ElectLeader() (string, error) {
	le.nodesMu.Lock()
	defer le.nodesMu.Unlock()

	if len(le.nodes) == 0 {
		return "", errors.New("no nodes registered")
	}

	// Count votes
	voteCount := make(map[string]int)
	for _, node := range le.nodes {
		for votedFor := range node.Votes {
			voteCount[votedFor]++
		}
		if node.IsLeader {
			// Check if leader is still alive
			if time.Now().After(node.LastHeartbeat.Add(le.heartbeatTTL)) {
				node.IsLeader = false
			}
		}
	}

	// Find candidate with most votes
	var maxVotes int
	var leader string
	for candidate, votes := range voteCount {
		if votes > maxVotes && votes > len(le.nodes)/2 {
			maxVotes = votes
			leader = candidate
		}
	}

	if leader == "" {
		// No clear leader, elect first active node
		for nodeID, node := range le.nodes {
			if time.Now().Before(node.LastHeartbeat.Add(le.heartbeatTTL)) {
				leader = nodeID
				break
			}
		}
	}

	if leader != "" {
		le.nodes[leader].IsLeader = true
		le.nodes[leader].LastHeartbeat = time.Now()
		le.logEntry("ELECTION", leader)
	}

	return leader, nil
}

// Heartbeat sends a heartbeat from the leader
func (le *LeaderElection) Heartbeat(nodeID string) error {
	le.nodesMu.Lock()
	defer le.nodesMu.Unlock()

	node, exists := le.nodes[nodeID]
	if !exists {
		return errors.New("node not found")
	}

	node.LastHeartbeat = time.Now()
	return nil
}

// GetLeader returns the current leader
func (le *LeaderElection) GetLeader() string {
	le.nodesMu.RLock()
	defer le.nodesMu.RUnlock()

	for nodeID, node := range le.nodes {
		if node.IsLeader {
			return nodeID
		}
	}

	return ""
}

func (le *LeaderElection) logEntry(action, nodeID string) {
	le.log = append(le.log, &LogEntry{
		Term:      le.currentTerm,
		Index:     int64(len(le.log)),
		NodeID:    nodeID,
		Action:    action,
		Timestamp: time.Now(),
	})
}

// ========== Helper Functions ==========

func generateLockID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

func main() {
	// Example distributed lock
	lm := NewLockManager(5 * time.Second)
	ctx := context.Background()

	_, _ = lm.AcquireLock(ctx, "resource-1", "owner-1", 5*time.Second)
}
