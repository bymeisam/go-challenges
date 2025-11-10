package main

import (
	"context"
	"database/sql"
	"sync"
	"testing"
	"time"
)

func getDefaultConfig() PoolConfig {
	return PoolConfig{
		MaxOpenConns:    10,
		MaxIdleConns:    5,
		ConnMaxLifetime: time.Hour,
		ConnMaxIdleTime: 10 * time.Minute,
	}
}

func setupTestDB(t *testing.T, config PoolConfig) *Database {
	db, err := NewDatabase(":memory:", config)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	return db
}

func TestNewDatabase(t *testing.T) {
	config := getDefaultConfig()
	db, err := NewDatabase(":memory:", config)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	defer db.Close()

	if db.db == nil {
		t.Error("Expected db connection to be initialized")
	}
}

func TestPoolConfiguration(t *testing.T) {
	config := PoolConfig{
		MaxOpenConns:    20,
		MaxIdleConns:    10,
		ConnMaxLifetime: 2 * time.Hour,
		ConnMaxIdleTime: 30 * time.Minute,
	}

	db := setupTestDB(t, config)
	defer db.Close()

	retrievedConfig := db.GetConfig()

	if retrievedConfig.MaxOpenConns != config.MaxOpenConns {
		t.Errorf("Expected MaxOpenConns %d, got %d",
			config.MaxOpenConns, retrievedConfig.MaxOpenConns)
	}

	if retrievedConfig.MaxIdleConns != config.MaxIdleConns {
		t.Errorf("Expected MaxIdleConns %d, got %d",
			config.MaxIdleConns, retrievedConfig.MaxIdleConns)
	}
}

func TestGetStats(t *testing.T) {
	config := getDefaultConfig()
	db := setupTestDB(t, config)
	defer db.Close()

	stats := db.GetStats()

	if stats.MaxOpenConnections != config.MaxOpenConns {
		t.Errorf("Expected MaxOpenConnections %d, got %d",
			config.MaxOpenConns, stats.MaxOpenConnections)
	}
}

func TestInsert(t *testing.T) {
	config := getDefaultConfig()
	db := setupTestDB(t, config)
	defer db.Close()

	id, err := db.Insert("test-item", 100)
	if err != nil {
		t.Fatalf("Failed to insert: %v", err)
	}

	if id == 0 {
		t.Error("Expected ID to be set")
	}
}

func TestQuery(t *testing.T) {
	config := getDefaultConfig()
	db := setupTestDB(t, config)
	defer db.Close()

	id, _ := db.Insert("test-item", 42)

	name, value, err := db.Query(id)
	if err != nil {
		t.Fatalf("Failed to query: %v", err)
	}

	if name != "test-item" {
		t.Errorf("Expected name 'test-item', got %s", name)
	}

	if value != 42 {
		t.Errorf("Expected value 42, got %d", value)
	}
}

func TestInsertWithContext(t *testing.T) {
	config := getDefaultConfig()
	db := setupTestDB(t, config)
	defer db.Close()

	ctx := context.Background()
	id, err := db.InsertWithContext(ctx, "context-item", 99)
	if err != nil {
		t.Fatalf("Failed to insert with context: %v", err)
	}

	if id == 0 {
		t.Error("Expected ID to be set")
	}
}

func TestInsertWithContextTimeout(t *testing.T) {
	config := getDefaultConfig()
	db := setupTestDB(t, config)
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	time.Sleep(10 * time.Millisecond) // Ensure context is expired

	_, err := db.InsertWithContext(ctx, "timeout-item", 99)
	if err == nil {
		t.Error("Expected error with expired context")
	}
}

func TestQueryWithContext(t *testing.T) {
	config := getDefaultConfig()
	db := setupTestDB(t, config)
	defer db.Close()

	id, _ := db.Insert("test-item", 55)

	ctx := context.Background()
	name, value, err := db.QueryWithContext(ctx, id)
	if err != nil {
		t.Fatalf("Failed to query with context: %v", err)
	}

	if name != "test-item" || value != 55 {
		t.Error("Unexpected query results")
	}
}

func TestPingWithContext(t *testing.T) {
	config := getDefaultConfig()
	db := setupTestDB(t, config)
	defer db.Close()

	ctx := context.Background()
	err := db.PingWithContext(ctx)
	if err != nil {
		t.Fatalf("Ping failed: %v", err)
	}
}

func TestGetConnectionWithTimeout(t *testing.T) {
	config := getDefaultConfig()
	db := setupTestDB(t, config)
	defer db.Close()

	conn, err := db.GetConnectionWithTimeout(5 * time.Second)
	if err != nil {
		t.Fatalf("Failed to get connection: %v", err)
	}
	defer conn.Close()

	if conn == nil {
		t.Error("Expected connection to be returned")
	}
}

func TestExecuteWithConnection(t *testing.T) {
	config := getDefaultConfig()
	db := setupTestDB(t, config)
	defer db.Close()

	conn, err := db.GetConnectionWithTimeout(5 * time.Second)
	if err != nil {
		t.Fatalf("Failed to get connection: %v", err)
	}
	defer conn.Close()

	err = db.ExecuteWithConnection(conn, "conn-item", 77)
	if err != nil {
		t.Fatalf("Failed to execute with connection: %v", err)
	}

	// Verify the insert worked
	_, value, err := db.Query(1)
	if err != nil {
		t.Fatalf("Failed to query inserted item: %v", err)
	}

	if value != 77 {
		t.Errorf("Expected value 77, got %d", value)
	}
}

func TestConcurrentInserts(t *testing.T) {
	config := PoolConfig{
		MaxOpenConns:    5,
		MaxIdleConns:    2,
		ConnMaxLifetime: time.Hour,
		ConnMaxIdleTime: 10 * time.Minute,
	}

	db := setupTestDB(t, config)
	defer db.Close()

	var wg sync.WaitGroup
	numGoroutines := 20
	errors := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			_, err := db.Insert("concurrent-item", index)
			if err != nil {
				errors <- err
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	errorCount := 0
	for err := range errors {
		t.Errorf("Concurrent insert error: %v", err)
		errorCount++
	}

	if errorCount > 0 {
		t.Fatalf("Had %d errors during concurrent inserts", errorCount)
	}

	stats := db.GetStats()
	if stats.OpenConnections > config.MaxOpenConns {
		t.Errorf("Exceeded max open connections: %d > %d",
			stats.OpenConnections, config.MaxOpenConns)
	}
}

func TestPoolStatsUnderLoad(t *testing.T) {
	config := PoolConfig{
		MaxOpenConns:    3,
		MaxIdleConns:    1,
		ConnMaxLifetime: time.Hour,
		ConnMaxIdleTime: 10 * time.Minute,
	}

	db := setupTestDB(t, config)
	defer db.Close()

	// Simulate load
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			db.Insert("load-test", index)
			time.Sleep(10 * time.Millisecond)
		}(i)
	}

	// Check stats while load is running
	time.Sleep(20 * time.Millisecond)
	stats := db.GetStats()

	if stats.OpenConnections > config.MaxOpenConns {
		t.Errorf("Exceeded max open connections: %d > %d",
			stats.OpenConnections, config.MaxOpenConns)
	}

	wg.Wait()

	// Final stats after load
	finalStats := db.GetStats()
	t.Logf("Final stats: Open=%d, InUse=%d, Idle=%d, WaitCount=%d",
		finalStats.OpenConnections, finalStats.InUse,
		finalStats.Idle, finalStats.WaitCount)
}

func TestSimulateLoad(t *testing.T) {
	config := getDefaultConfig()
	db := setupTestDB(t, config)
	defer db.Close()

	numRequests := 30
	errors := db.SimulateLoad(numRequests, 5*time.Millisecond)

	if len(errors) > 0 {
		t.Errorf("Expected no errors, got %d errors", len(errors))
	}

	stats := db.GetStats()
	t.Logf("Stats after load: Open=%d, InUse=%d, Idle=%d, WaitCount=%d",
		stats.OpenConnections, stats.InUse, stats.Idle, stats.WaitCount)
}

func TestConnectionPoolExhaustion(t *testing.T) {
	config := PoolConfig{
		MaxOpenConns:    2,
		MaxIdleConns:    1,
		ConnMaxLifetime: time.Hour,
		ConnMaxIdleTime: 10 * time.Minute,
	}

	db := setupTestDB(t, config)
	defer db.Close()

	// Acquire connections and hold them
	conn1, err := db.GetConnectionWithTimeout(1 * time.Second)
	if err != nil {
		t.Fatalf("Failed to get first connection: %v", err)
	}
	defer conn1.Close()

	conn2, err := db.GetConnectionWithTimeout(1 * time.Second)
	if err != nil {
		t.Fatalf("Failed to get second connection: %v", err)
	}
	defer conn2.Close()

	// Try to get a third connection with short timeout
	_, err = db.GetConnectionWithTimeout(100 * time.Millisecond)
	if err == nil {
		t.Error("Expected error when pool is exhausted")
	}

	stats := db.GetStats()
	if stats.WaitCount == 0 {
		t.Log("Warning: Expected WaitCount > 0, but connection might have been acquired")
	}
}

func TestMaxIdleConnections(t *testing.T) {
	config := PoolConfig{
		MaxOpenConns:    10,
		MaxIdleConns:    2,
		ConnMaxLifetime: time.Hour,
		ConnMaxIdleTime: 10 * time.Minute,
	}

	db := setupTestDB(t, config)
	defer db.Close()

	// Create some activity
	for i := 0; i < 5; i++ {
		db.Insert("test", i)
	}

	// Wait a moment for connections to return to idle
	time.Sleep(100 * time.Millisecond)

	stats := db.GetStats()

	// Idle connections should not exceed MaxIdleConns
	if stats.Idle > config.MaxIdleConns {
		t.Errorf("Idle connections %d exceeds MaxIdleConns %d",
			stats.Idle, config.MaxIdleConns)
	}
}

func TestQueryNotFound(t *testing.T) {
	config := getDefaultConfig()
	db := setupTestDB(t, config)
	defer db.Close()

	_, _, err := db.Query(999)
	if err != sql.ErrNoRows {
		t.Errorf("Expected sql.ErrNoRows, got %v", err)
	}
}
