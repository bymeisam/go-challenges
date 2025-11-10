package main

import (
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func setupTestRedis(t *testing.T) (*RedisClient, *miniredis.Miniredis) {
	// Create a miniredis server
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("Failed to start miniredis: %v", err)
	}

	// Create Redis client
	client, err := NewRedisClient(mr.Addr())
	if err != nil {
		mr.Close()
		t.Fatalf("Failed to create Redis client: %v", err)
	}

	return client, mr
}

func TestNewRedisClient(t *testing.T) {
	client, mr := setupTestRedis(t)
	defer mr.Close()
	defer client.Close()

	if client.client == nil {
		t.Error("Expected client to be initialized")
	}
}

func TestSetAndGet(t *testing.T) {
	client, mr := setupTestRedis(t)
	defer mr.Close()
	defer client.Close()

	key := "test-key"
	value := "test-value"

	err := client.Set(key, value, 0)
	if err != nil {
		t.Fatalf("Failed to set key: %v", err)
	}

	retrieved, err := client.Get(key)
	if err != nil {
		t.Fatalf("Failed to get key: %v", err)
	}

	if retrieved != value {
		t.Errorf("Expected value %s, got %s", value, retrieved)
	}
}

func TestGetNonExistent(t *testing.T) {
	client, mr := setupTestRedis(t)
	defer mr.Close()
	defer client.Close()

	_, err := client.Get("non-existent")
	if err != redis.Nil {
		t.Errorf("Expected redis.Nil error, got %v", err)
	}
}

func TestDelete(t *testing.T) {
	client, mr := setupTestRedis(t)
	defer mr.Close()
	defer client.Close()

	key := "to-delete"
	client.Set(key, "value", 0)

	err := client.Delete(key)
	if err != nil {
		t.Fatalf("Failed to delete key: %v", err)
	}

	_, err = client.Get(key)
	if err != redis.Nil {
		t.Error("Key should be deleted")
	}
}

func TestExists(t *testing.T) {
	client, mr := setupTestRedis(t)
	defer mr.Close()
	defer client.Close()

	key := "exists-test"
	client.Set(key, "value", 0)

	exists, err := client.Exists(key)
	if err != nil {
		t.Fatalf("Failed to check existence: %v", err)
	}

	if !exists {
		t.Error("Key should exist")
	}

	exists, _ = client.Exists("non-existent")
	if exists {
		t.Error("Non-existent key should not exist")
	}
}

func TestExpiration(t *testing.T) {
	client, mr := setupTestRedis(t)
	defer mr.Close()
	defer client.Close()

	key := "expiring-key"
	client.Set(key, "value", 100*time.Millisecond)

	// Fast forward time in miniredis
	mr.FastForward(150 * time.Millisecond)

	_, err := client.Get(key)
	if err != redis.Nil {
		t.Error("Key should have expired")
	}
}

func TestGetTTL(t *testing.T) {
	client, mr := setupTestRedis(t)
	defer mr.Close()
	defer client.Close()

	key := "ttl-test"
	client.Set(key, "value", 10*time.Second)

	ttl, err := client.GetTTL(key)
	if err != nil {
		t.Fatalf("Failed to get TTL: %v", err)
	}

	if ttl <= 0 {
		t.Error("TTL should be positive")
	}

	// Key without expiration should return -1
	client.Set("no-expiry", "value", 0)
	ttl, _ = client.GetTTL("no-expiry")
	if ttl != -1 {
		t.Errorf("Expected TTL -1 for key without expiration, got %v", ttl)
	}
}

func TestExpire(t *testing.T) {
	client, mr := setupTestRedis(t)
	defer mr.Close()
	defer client.Close()

	key := "expire-test"
	client.Set(key, "value", 0)

	err := client.Expire(key, 5*time.Second)
	if err != nil {
		t.Fatalf("Failed to set expiration: %v", err)
	}

	ttl, _ := client.GetTTL(key)
	if ttl <= 0 {
		t.Error("TTL should be set")
	}
}

func TestSetJSON(t *testing.T) {
	client, mr := setupTestRedis(t)
	defer mr.Close()
	defer client.Close()

	user := User{
		ID:    "123",
		Name:  "Test User",
		Email: "test@example.com",
	}

	err := client.SetJSON("user:123", user, 0)
	if err != nil {
		t.Fatalf("Failed to set JSON: %v", err)
	}

	var retrieved User
	err = client.GetJSON("user:123", &retrieved)
	if err != nil {
		t.Fatalf("Failed to get JSON: %v", err)
	}

	if retrieved.ID != user.ID {
		t.Errorf("Expected ID %s, got %s", user.ID, retrieved.ID)
	}

	if retrieved.Name != user.Name {
		t.Errorf("Expected Name %s, got %s", user.Name, retrieved.Name)
	}

	if retrieved.Email != user.Email {
		t.Errorf("Expected Email %s, got %s", user.Email, retrieved.Email)
	}
}

func TestIncrement(t *testing.T) {
	client, mr := setupTestRedis(t)
	defer mr.Close()
	defer client.Close()

	key := "counter"

	count, err := client.Increment(key)
	if err != nil {
		t.Fatalf("Failed to increment: %v", err)
	}

	if count != 1 {
		t.Errorf("Expected count 1, got %d", count)
	}

	count, _ = client.Increment(key)
	if count != 2 {
		t.Errorf("Expected count 2, got %d", count)
	}
}

func TestIncrementBy(t *testing.T) {
	client, mr := setupTestRedis(t)
	defer mr.Close()
	defer client.Close()

	key := "counter"

	count, err := client.IncrementBy(key, 5)
	if err != nil {
		t.Fatalf("Failed to increment by: %v", err)
	}

	if count != 5 {
		t.Errorf("Expected count 5, got %d", count)
	}

	count, _ = client.IncrementBy(key, 10)
	if count != 15 {
		t.Errorf("Expected count 15, got %d", count)
	}
}

func TestDecrement(t *testing.T) {
	client, mr := setupTestRedis(t)
	defer mr.Close()
	defer client.Close()

	key := "counter"
	client.Set(key, "10", 0)

	count, err := client.Decrement(key)
	if err != nil {
		t.Fatalf("Failed to decrement: %v", err)
	}

	if count != 9 {
		t.Errorf("Expected count 9, got %d", count)
	}
}

func TestListPushAndRange(t *testing.T) {
	client, mr := setupTestRedis(t)
	defer mr.Close()
	defer client.Close()

	key := "list"

	err := client.ListPush(key, "item1", "item2", "item3")
	if err != nil {
		t.Fatalf("Failed to push to list: %v", err)
	}

	items, err := client.ListRange(key, 0, -1)
	if err != nil {
		t.Fatalf("Failed to get list range: %v", err)
	}

	if len(items) != 3 {
		t.Errorf("Expected 3 items, got %d", len(items))
	}

	// Items are pushed to the left, so order is reversed
	if items[0] != "item3" {
		t.Errorf("Expected first item 'item3', got %s", items[0])
	}
}

func TestListPop(t *testing.T) {
	client, mr := setupTestRedis(t)
	defer mr.Close()
	defer client.Close()

	key := "list"
	client.ListPush(key, "item1", "item2")

	item, err := client.ListPop(key)
	if err != nil {
		t.Fatalf("Failed to pop from list: %v", err)
	}

	if item != "item2" {
		t.Errorf("Expected 'item2', got %s", item)
	}

	length, _ := client.ListLength(key)
	if length != 1 {
		t.Errorf("Expected list length 1, got %d", length)
	}
}

func TestListLength(t *testing.T) {
	client, mr := setupTestRedis(t)
	defer mr.Close()
	defer client.Close()

	key := "list"
	client.ListPush(key, "a", "b", "c", "d")

	length, err := client.ListLength(key)
	if err != nil {
		t.Fatalf("Failed to get list length: %v", err)
	}

	if length != 4 {
		t.Errorf("Expected length 4, got %d", length)
	}
}

func TestHashSetAndGet(t *testing.T) {
	client, mr := setupTestRedis(t)
	defer mr.Close()
	defer client.Close()

	key := "hash"
	field := "name"
	value := "Alice"

	err := client.HashSet(key, field, value)
	if err != nil {
		t.Fatalf("Failed to set hash field: %v", err)
	}

	retrieved, err := client.HashGet(key, field)
	if err != nil {
		t.Fatalf("Failed to get hash field: %v", err)
	}

	if retrieved != value {
		t.Errorf("Expected value %s, got %s", value, retrieved)
	}
}

func TestHashGetAll(t *testing.T) {
	client, mr := setupTestRedis(t)
	defer mr.Close()
	defer client.Close()

	key := "hash"
	client.HashSet(key, "name", "Bob")
	client.HashSet(key, "email", "bob@example.com")
	client.HashSet(key, "age", "30")

	all, err := client.HashGetAll(key)
	if err != nil {
		t.Fatalf("Failed to get all hash fields: %v", err)
	}

	if len(all) != 3 {
		t.Errorf("Expected 3 fields, got %d", len(all))
	}

	if all["name"] != "Bob" {
		t.Errorf("Expected name 'Bob', got %s", all["name"])
	}
}

func TestHashDelete(t *testing.T) {
	client, mr := setupTestRedis(t)
	defer mr.Close()
	defer client.Close()

	key := "hash"
	client.HashSet(key, "field1", "value1")
	client.HashSet(key, "field2", "value2")

	err := client.HashDelete(key, "field1")
	if err != nil {
		t.Fatalf("Failed to delete hash field: %v", err)
	}

	_, err = client.HashGet(key, "field1")
	if err != redis.Nil {
		t.Error("Field should be deleted")
	}

	// field2 should still exist
	_, err = client.HashGet(key, "field2")
	if err != nil {
		t.Error("field2 should still exist")
	}
}

func TestHashExists(t *testing.T) {
	client, mr := setupTestRedis(t)
	defer mr.Close()
	defer client.Close()

	key := "hash"
	client.HashSet(key, "field", "value")

	exists, err := client.HashExists(key, "field")
	if err != nil {
		t.Fatalf("Failed to check hash field existence: %v", err)
	}

	if !exists {
		t.Error("Field should exist")
	}

	exists, _ = client.HashExists(key, "non-existent")
	if exists {
		t.Error("Non-existent field should not exist")
	}
}

func TestSetOperations(t *testing.T) {
	client, mr := setupTestRedis(t)
	defer mr.Close()
	defer client.Close()

	key := "set"

	err := client.SetAdd(key, "member1", "member2", "member3")
	if err != nil {
		t.Fatalf("Failed to add to set: %v", err)
	}

	members, err := client.SetMembers(key)
	if err != nil {
		t.Fatalf("Failed to get set members: %v", err)
	}

	if len(members) != 3 {
		t.Errorf("Expected 3 members, got %d", len(members))
	}
}

func TestSetIsMember(t *testing.T) {
	client, mr := setupTestRedis(t)
	defer mr.Close()
	defer client.Close()

	key := "set"
	client.SetAdd(key, "apple", "banana", "cherry")

	isMember, err := client.SetIsMember(key, "banana")
	if err != nil {
		t.Fatalf("Failed to check set membership: %v", err)
	}

	if !isMember {
		t.Error("'banana' should be a member")
	}

	isMember, _ = client.SetIsMember(key, "grape")
	if isMember {
		t.Error("'grape' should not be a member")
	}
}

func TestSetRemove(t *testing.T) {
	client, mr := setupTestRedis(t)
	defer mr.Close()
	defer client.Close()

	key := "set"
	client.SetAdd(key, "a", "b", "c")

	err := client.SetRemove(key, "b")
	if err != nil {
		t.Fatalf("Failed to remove from set: %v", err)
	}

	isMember, _ := client.SetIsMember(key, "b")
	if isMember {
		t.Error("'b' should be removed")
	}

	members, _ := client.SetMembers(key)
	if len(members) != 2 {
		t.Errorf("Expected 2 members, got %d", len(members))
	}
}

func TestFlushAll(t *testing.T) {
	client, mr := setupTestRedis(t)
	defer mr.Close()
	defer client.Close()

	client.Set("key1", "value1", 0)
	client.Set("key2", "value2", 0)

	err := client.FlushAll()
	if err != nil {
		t.Fatalf("Failed to flush all: %v", err)
	}

	exists1, _ := client.Exists("key1")
	exists2, _ := client.Exists("key2")

	if exists1 || exists2 {
		t.Error("All keys should be deleted")
	}
}
