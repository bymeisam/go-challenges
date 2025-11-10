package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisClient struct {
	client *redis.Client
	ctx    context.Context
}

type User struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// NewRedisClient creates a new Redis client
func NewRedisClient(addr string) (*RedisClient, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: "",
		DB:       0,
	})

	ctx := context.Background()

	// Test connection
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return &RedisClient{
		client: client,
		ctx:    ctx,
	}, nil
}

// Close closes the Redis connection
func (r *RedisClient) Close() error {
	return r.client.Close()
}

// Set sets a key-value pair
func (r *RedisClient) Set(key string, value interface{}, expiration time.Duration) error {
	return r.client.Set(r.ctx, key, value, expiration).Err()
}

// Get retrieves a value by key
func (r *RedisClient) Get(key string) (string, error) {
	return r.client.Get(r.ctx, key).Result()
}

// Delete deletes a key
func (r *RedisClient) Delete(key string) error {
	return r.client.Del(r.ctx, key).Err()
}

// Exists checks if a key exists
func (r *RedisClient) Exists(key string) (bool, error) {
	result, err := r.client.Exists(r.ctx, key).Result()
	if err != nil {
		return false, err
	}
	return result > 0, nil
}

// GetTTL gets the time-to-live for a key
func (r *RedisClient) GetTTL(key string) (time.Duration, error) {
	return r.client.TTL(r.ctx, key).Result()
}

// Expire sets an expiration time for a key
func (r *RedisClient) Expire(key string, expiration time.Duration) error {
	return r.client.Expire(r.ctx, key, expiration).Err()
}

// SetJSON sets a JSON-encoded value
func (r *RedisClient) SetJSON(key string, value interface{}, expiration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return r.client.Set(r.ctx, key, data, expiration).Err()
}

// GetJSON retrieves and decodes a JSON value
func (r *RedisClient) GetJSON(key string, dest interface{}) error {
	data, err := r.client.Get(r.ctx, key).Result()
	if err != nil {
		return err
	}

	return json.Unmarshal([]byte(data), dest)
}

// Increment increments a numeric value
func (r *RedisClient) Increment(key string) (int64, error) {
	return r.client.Incr(r.ctx, key).Result()
}

// IncrementBy increments a numeric value by delta
func (r *RedisClient) IncrementBy(key string, delta int64) (int64, error) {
	return r.client.IncrBy(r.ctx, key, delta).Result()
}

// Decrement decrements a numeric value
func (r *RedisClient) Decrement(key string) (int64, error) {
	return r.client.Decr(r.ctx, key).Result()
}

// ListPush pushes values to a list (left push)
func (r *RedisClient) ListPush(key string, values ...interface{}) error {
	return r.client.LPush(r.ctx, key, values...).Err()
}

// ListPop pops a value from a list (left pop)
func (r *RedisClient) ListPop(key string) (string, error) {
	return r.client.LPop(r.ctx, key).Result()
}

// ListRange gets a range of elements from a list
func (r *RedisClient) ListRange(key string, start, stop int64) ([]string, error) {
	return r.client.LRange(r.ctx, key, start, stop).Result()
}

// ListLength gets the length of a list
func (r *RedisClient) ListLength(key string) (int64, error) {
	return r.client.LLen(r.ctx, key).Result()
}

// HashSet sets a field in a hash
func (r *RedisClient) HashSet(key, field string, value interface{}) error {
	return r.client.HSet(r.ctx, key, field, value).Err()
}

// HashGet gets a field from a hash
func (r *RedisClient) HashGet(key, field string) (string, error) {
	return r.client.HGet(r.ctx, key, field).Result()
}

// HashGetAll gets all fields from a hash
func (r *RedisClient) HashGetAll(key string) (map[string]string, error) {
	return r.client.HGetAll(r.ctx, key).Result()
}

// HashDelete deletes a field from a hash
func (r *RedisClient) HashDelete(key, field string) error {
	return r.client.HDel(r.ctx, key, field).Err()
}

// HashExists checks if a field exists in a hash
func (r *RedisClient) HashExists(key, field string) (bool, error) {
	return r.client.HExists(r.ctx, key, field).Result()
}

// SetAdd adds members to a set
func (r *RedisClient) SetAdd(key string, members ...interface{}) error {
	return r.client.SAdd(r.ctx, key, members...).Err()
}

// SetMembers gets all members of a set
func (r *RedisClient) SetMembers(key string) ([]string, error) {
	return r.client.SMembers(r.ctx, key).Result()
}

// SetIsMember checks if a value is a member of a set
func (r *RedisClient) SetIsMember(key string, member interface{}) (bool, error) {
	return r.client.SIsMember(r.ctx, key, member).Result()
}

// SetRemove removes members from a set
func (r *RedisClient) SetRemove(key string, members ...interface{}) error {
	return r.client.SRem(r.ctx, key, members...).Err()
}

// FlushAll removes all keys from the current database
func (r *RedisClient) FlushAll() error {
	return r.client.FlushDB(r.ctx).Err()
}

func main() {
	client, err := NewRedisClient("localhost:6379")
	if err != nil {
		fmt.Printf("Error connecting to Redis: %v\n", err)
		return
	}
	defer client.Close()

	// Basic key-value operations
	client.Set("greeting", "Hello, Redis!", 0)
	value, _ := client.Get("greeting")
	fmt.Printf("Retrieved: %s\n", value)

	// JSON operations
	user := User{
		ID:    "1",
		Name:  "John Doe",
		Email: "john@example.com",
	}
	client.SetJSON("user:1", user, 5*time.Minute)

	var retrieved User
	client.GetJSON("user:1", &retrieved)
	fmt.Printf("User: %+v\n", retrieved)

	// Counter operations
	client.Increment("visits")
	count, _ := client.Get("visits")
	fmt.Printf("Visits: %s\n", count)

	// List operations
	client.ListPush("tasks", "task1", "task2", "task3")
	tasks, _ := client.ListRange("tasks", 0, -1)
	fmt.Printf("Tasks: %v\n", tasks)

	// Hash operations
	client.HashSet("user:info", "name", "Alice")
	client.HashSet("user:info", "email", "alice@example.com")
	info, _ := client.HashGetAll("user:info")
	fmt.Printf("User Info: %v\n", info)
}
