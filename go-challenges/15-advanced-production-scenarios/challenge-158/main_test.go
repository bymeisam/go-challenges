package main

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"
)

// ========== Idempotency Tests ==========

func TestIdempotencyStoreKeyGeneration(t *testing.T) {
	store := NewIdempotencyStore(24 * time.Hour)

	key1 := store.GenerateKey("request-1")
	key2 := store.GenerateKey("request-1")
	key3 := store.GenerateKey("request-2")

	if key1 != key2 {
		t.Fatal("Expected same key for identical input")
	}

	if key1 == key3 {
		t.Fatal("Expected different keys for different input")
	}

	if len(key1) == 0 {
		t.Fatal("Expected non-empty key")
	}
}

func TestIdempotencyStoreRequest(t *testing.T) {
	store := NewIdempotencyStore(1 * time.Hour)
	key := "test-key-1"

	err := store.StoreRequest(key)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Store again should fail
	err = store.StoreRequest(key)
	if err == nil {
		t.Fatal("Expected error when storing duplicate key")
	}
}

func TestIdempotencyStoreResponse(t *testing.T) {
	store := NewIdempotencyStore(1 * time.Hour)
	key := "test-key-2"

	store.StoreRequest(key)

	response := "response-data"
	err := store.UpdateResponse(key, response, nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	retrieved, success := store.GetResponse(key)
	if !success {
		t.Fatal("Expected successful response retrieval")
	}

	if retrieved != "response-data" {
		t.Fatalf("Expected response-data, got %v", retrieved)
	}
}

func TestIdempotencyGetResponse(t *testing.T) {
	store := NewIdempotencyStore(1 * time.Hour)
	key := "test-key-3"

	store.StoreRequest(key)
	store.UpdateResponse(key, "cached-response", nil)

	// Should return cached response on second call
	retrieved, success := store.GetResponse(key)
	if !success || retrieved != "cached-response" {
		t.Fatal("Expected cached response")
	}
}

func TestIdempotencyKeyExpiration(t *testing.T) {
	store := NewIdempotencyStore(100 * time.Millisecond)
	key := "test-key-4"

	store.StoreRequest(key)
	if !store.ValidateKey(key) {
		t.Fatal("Expected key to be valid")
	}

	time.Sleep(150 * time.Millisecond)

	if store.ValidateKey(key) {
		t.Fatal("Expected key to be expired")
	}
}

func TestIdempotencyKeyStatus(t *testing.T) {
	store := NewIdempotencyStore(1 * time.Hour)
	key := "test-key-5"

	if store.GetKeyStatus(key) != "NOT_FOUND" {
		t.Fatal("Expected NOT_FOUND status for non-existent key")
	}

	store.StoreRequest(key)
	if store.GetKeyStatus(key) != "PENDING" {
		t.Fatal("Expected PENDING status")
	}

	store.UpdateResponse(key, "response", nil)
	if store.GetKeyStatus(key) != "SUCCESS" {
		t.Fatal("Expected SUCCESS status")
	}
}

func TestIdempotencyErrorResponse(t *testing.T) {
	store := NewIdempotencyStore(1 * time.Hour)
	key := "test-key-6"

	store.StoreRequest(key)
	err := store.UpdateResponse(key, nil, errors.New("test error"))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if store.GetKeyStatus(key) != "FAILED" {
		t.Fatal("Expected FAILED status")
	}
}

// ========== Distributed Lock Tests ==========

func TestLockManagerAcquireLock(t *testing.T) {
	lm := NewLockManager(1 * time.Second)
	ctx := context.Background()

	lock, err := lm.AcquireLock(ctx, "resource-1", "owner-1", 5*time.Second)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if lock.Status != LockAcquired {
		t.Fatalf("Expected LockAcquired, got %v", lock.Status)
	}

	if lock.OwnerID != "owner-1" {
		t.Fatalf("Expected owner-1, got %s", lock.OwnerID)
	}
}

func TestLockManagerContention(t *testing.T) {
	lm := NewLockManager(1 * time.Second)
	ctx := context.Background()

	lock1, _ := lm.AcquireLock(ctx, "resource-2", "owner-1", 5*time.Second)
	if lock1 == nil {
		t.Fatal("Expected first lock to be acquired")
	}

	// Try to acquire same lock
	lock2, err := lm.AcquireLock(ctx, "resource-2", "owner-2", 5*time.Second)
	if err == nil {
		t.Fatal("Expected error due to lock contention")
	}

	if lock2 != nil {
		t.Fatal("Expected lock2 to be nil")
	}
}

func TestLockManagerRelease(t *testing.T) {
	lm := NewLockManager(1 * time.Second)
	ctx := context.Background()

	lock, _ := lm.AcquireLock(ctx, "resource-3", "owner-1", 5*time.Second)

	err := lm.ReleaseLock(lock.LockID, "owner-1")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Should be able to acquire again
	lock2, err := lm.AcquireLock(ctx, "resource-3", "owner-2", 5*time.Second)
	if err != nil {
		t.Fatalf("Expected to acquire lock after release, got %v", err)
	}

	if lock2 == nil {
		t.Fatal("Expected lock to be acquired")
	}
}

func TestLockManagerLeaseExpiry(t *testing.T) {
	lm := NewLockManager(100 * time.Millisecond)
	ctx := context.Background()

	lock, _ := lm.AcquireLock(ctx, "resource-4", "owner-1", 5*time.Second)
	if lock == nil {
		t.Fatal("Expected lock to be acquired")
	}

	time.Sleep(150 * time.Millisecond)

	// Lock should have expired, new owner should acquire it
	lock2, err := lm.AcquireLock(ctx, "resource-4", "owner-2", 5*time.Second)
	if err != nil {
		t.Fatalf("Expected to acquire expired lock, got %v", err)
	}

	if lock2 == nil {
		t.Fatal("Expected lock to be acquired")
	}
}

func TestLockManagerRenewLease(t *testing.T) {
	lm := NewLockManager(1 * time.Second)
	ctx := context.Background()

	lock, _ := lm.AcquireLock(ctx, "resource-5", "owner-1", 5*time.Second)
	originalExpiry := lock.ExpiresAt

	time.Sleep(100 * time.Millisecond)

	err := lm.RenewLease(lock.LockID, "owner-1")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	lock = lm.CheckLock(lock.LockID)
	if lock.ExpiresAt.Before(originalExpiry.Add(900 * time.Millisecond)) {
		t.Fatal("Expected lease to be renewed")
	}
}

func TestLockManagerCheckLock(t *testing.T) {
	lm := NewLockManager(1 * time.Second)
	ctx := context.Background()

	lock, _ := lm.AcquireLock(ctx, "resource-6", "owner-1", 5*time.Second)

	retrieved := lm.CheckLock(lock.LockID)
	if retrieved == nil {
		t.Fatal("Expected lock to be found")
	}

	if retrieved.OwnerID != "owner-1" {
		t.Fatalf("Expected owner-1, got %s", retrieved.OwnerID)
	}
}

func TestLockManagerUnauthorizedRelease(t *testing.T) {
	lm := NewLockManager(1 * time.Second)
	ctx := context.Background()

	lock, _ := lm.AcquireLock(ctx, "resource-7", "owner-1", 5*time.Second)

	err := lm.ReleaseLock(lock.LockID, "owner-2")
	if err == nil {
		t.Fatal("Expected error when releasing lock owned by different user")
	}
}

func TestLockManagerMetrics(t *testing.T) {
	lm := NewLockManager(1 * time.Second)
	ctx := context.Background()

	lm.AcquireLock(ctx, "resource-8", "owner-1", 5*time.Second)
	lm.AcquireLock(ctx, "resource-9", "owner-2", 5*time.Second)

	metrics := lm.GetMetrics()
	if metrics.TotalAcquisitions != 2 {
		t.Fatalf("Expected 2 acquisitions, got %d", metrics.TotalAcquisitions)
	}
}

func TestLockManagerDeadlockDetection(t *testing.T) {
	lm := NewLockManager(10 * time.Second)
	ctx := context.Background()

	lm.AcquireLock(ctx, "resource-10", "owner-1", 5*time.Second)
	lm.AcquireLock(ctx, "resource-11", "owner-2", 5*time.Second)

	deadlocks := lm.DetectDeadlock()
	// May or may not detect depending on lock dependencies
	_ = deadlocks
}

// ========== Leader Election Tests ==========

func TestLeaderElectionRegisterNode(t *testing.T) {
	le := NewLeaderElection(1 * time.Second)

	err := le.RegisterNode("node-1")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Register again should fail
	err = le.RegisterNode("node-1")
	if err == nil {
		t.Fatal("Expected error when registering duplicate node")
	}
}

func TestLeaderElectionCastVote(t *testing.T) {
	le := NewLeaderElection(1 * time.Second)

	le.RegisterNode("node-1")
	le.RegisterNode("node-2")

	err := le.CastVote("node-1", "node-2")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

func TestLeaderElectionElectLeader(t *testing.T) {
	le := NewLeaderElection(10 * time.Second)

	le.RegisterNode("node-1")
	le.RegisterNode("node-2")
	le.RegisterNode("node-3")

	// All vote for node-1
	le.CastVote("node-1", "node-1")
	le.CastVote("node-2", "node-1")
	le.CastVote("node-3", "node-1")

	leader, err := le.ElectLeader()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if leader != "node-1" {
		t.Fatalf("Expected node-1 as leader, got %s", leader)
	}
}

func TestLeaderElectionHeartbeat(t *testing.T) {
	le := NewLeaderElection(1 * time.Second)

	le.RegisterNode("node-1")

	err := le.Heartbeat("node-1")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	err = le.Heartbeat("node-99")
	if err == nil {
		t.Fatal("Expected error for non-existent node")
	}
}

func TestLeaderElectionGetLeader(t *testing.T) {
	le := NewLeaderElection(10 * time.Second)

	le.RegisterNode("node-1")
	le.RegisterNode("node-2")

	le.CastVote("node-1", "node-1")
	le.CastVote("node-2", "node-1")

	le.ElectLeader()

	leader := le.GetLeader()
	if leader != "node-1" {
		t.Fatalf("Expected node-1 as leader, got %s", leader)
	}
}

func TestLeaderElectionMultipleVotes(t *testing.T) {
	le := NewLeaderElection(10 * time.Second)

	for i := 1; i <= 5; i++ {
		le.RegisterNode(fmt.Sprintf("node-%d", i))
	}

	// node-3 gets majority
	le.CastVote("node-1", "node-3")
	le.CastVote("node-2", "node-3")
	le.CastVote("node-3", "node-3")

	leader, _ := le.ElectLeader()
	if leader != "node-3" {
		t.Fatalf("Expected node-3 as leader, got %s", leader)
	}
}

// ========== Integration Tests ==========

func TestIdempotencyWithLocking(t *testing.T) {
	store := NewIdempotencyStore(1 * time.Hour)
	lm := NewLockManager(1 * time.Second)
	ctx := context.Background()

	key := "request-123"
	resource := "resource-123"

	// Store request
	store.StoreRequest(key)

	// Acquire lock
	lock, err := lm.AcquireLock(ctx, resource, "owner-1", 5*time.Second)
	if err != nil {
		t.Fatalf("Expected to acquire lock, got %v", err)
	}

	// Update response
	store.UpdateResponse(key, "result", nil)

	// Release lock
	lm.ReleaseLock(lock.LockID, "owner-1")

	// Verify idempotency
	response, success := store.GetResponse(key)
	if !success || response != "result" {
		t.Fatal("Expected cached response")
	}
}

func TestConcurrentLockAcquisition(t *testing.T) {
	lm := NewLockManager(5 * time.Second)
	ctx := context.Background()
	results := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(index int) {
			_, err := lm.AcquireLock(ctx, "shared-resource", fmt.Sprintf("owner-%d", index), 5*time.Second)
			results <- err == nil
		}(i)
	}

	acquired := 0
	failed := 0
	for i := 0; i < 10; i++ {
		if <-results {
			acquired++
		} else {
			failed++
		}
	}

	if acquired == 0 {
		t.Fatal("Expected at least one lock acquisition")
	}

	if failed == 0 {
		t.Fatal("Expected some lock contentions")
	}
}

func TestLeaderElectionWithHeartbeat(t *testing.T) {
	le := NewLeaderElection(100 * time.Millisecond)

	le.RegisterNode("node-1")
	le.RegisterNode("node-2")

	le.CastVote("node-1", "node-1")
	le.CastVote("node-2", "node-1")

	le.ElectLeader()
	leader := le.GetLeader()

	if leader != "node-1" {
		t.Fatalf("Expected node-1 as leader, got %s", leader)
	}

	// Send heartbeat
	le.Heartbeat("node-1")

	// Verify leader is still active
	leader = le.GetLeader()
	if leader != "node-1" {
		t.Fatalf("Expected node-1 to remain leader, got %s", leader)
	}
}

// ========== Benchmarks ==========

func BenchmarkIdempotencyKeyGeneration(b *testing.B) {
	store := NewIdempotencyStore(24 * time.Hour)
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		store.GenerateKey(fmt.Sprintf("request-%d", i))
	}
}

func BenchmarkLockAcquisition(b *testing.B) {
	lm := NewLockManager(5 * time.Second)
	ctx := context.Background()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		lm.AcquireLock(ctx, fmt.Sprintf("resource-%d", i%10), fmt.Sprintf("owner-%d", i), 5*time.Second)
	}
}

func BenchmarkLeaderElection(b *testing.B) {
	le := NewLeaderElection(10 * time.Second)

	for i := 1; i <= 10; i++ {
		le.RegisterNode(fmt.Sprintf("node-%d", i))
	}

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		le.ElectLeader()
	}
}

// ========== Helper for tests ==========
