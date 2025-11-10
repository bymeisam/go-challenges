package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Product struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"type:varchar(200);not null" json:"name"`
	Description string    `gorm:"type:text" json:"description"`
	Price       float64   `gorm:"type:decimal(10,2);not null" json:"price"`
	Stock       int       `gorm:"type:int;default:0" json:"stock"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CacheStats struct {
	Hits   int64
	Misses int64
}

type CachedStore struct {
	db         *gorm.DB
	cache      *redis.Client
	ctx        context.Context
	cacheTTL   time.Duration
	stats      CacheStats
}

// NewCachedStore creates a new store with caching
func NewCachedStore(dbDSN string, redisAddr string, cacheTTL time.Duration) (*CachedStore, error) {
	// Initialize database
	db, err := gorm.Open(sqlite.Open(dbDSN), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, err
	}

	// Auto migrate
	if err := db.AutoMigrate(&Product{}); err != nil {
		return nil, err
	}

	// Initialize Redis cache
	redisClient := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: "",
		DB:       0,
	})

	ctx := context.Background()

	// Test Redis connection
	if err := redisClient.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return &CachedStore{
		db:       db,
		cache:    redisClient,
		ctx:      ctx,
		cacheTTL: cacheTTL,
		stats:    CacheStats{},
	}, nil
}

// Close closes all connections
func (s *CachedStore) Close() error {
	sqlDB, err := s.db.DB()
	if err != nil {
		return err
	}
	sqlDB.Close()
	return s.cache.Close()
}

// getCacheKey generates a cache key for a product
func (s *CachedStore) getCacheKey(id uint) string {
	return fmt.Sprintf("product:%d", id)
}

// CreateProduct creates a new product
func (s *CachedStore) CreateProduct(product *Product) error {
	// Create in database
	if err := s.db.Create(product).Error; err != nil {
		return err
	}

	// Cache the newly created product
	return s.cacheProduct(product)
}

// GetProduct retrieves a product with cache-aside pattern
func (s *CachedStore) GetProduct(id uint) (*Product, error) {
	cacheKey := s.getCacheKey(id)

	// Try to get from cache first
	cachedData, err := s.cache.Get(s.ctx, cacheKey).Result()
	if err == nil {
		// Cache hit
		s.stats.Hits++
		var product Product
		if err := json.Unmarshal([]byte(cachedData), &product); err != nil {
			return nil, err
		}
		return &product, nil
	}

	// Cache miss - get from database
	s.stats.Misses++

	var product Product
	if err := s.db.First(&product, id).Error; err != nil {
		return nil, err
	}

	// Store in cache for next time
	if err := s.cacheProduct(&product); err != nil {
		// Log error but don't fail the request
		fmt.Printf("Failed to cache product: %v\n", err)
	}

	return &product, nil
}

// cacheProduct caches a product
func (s *CachedStore) cacheProduct(product *Product) error {
	cacheKey := s.getCacheKey(product.ID)
	data, err := json.Marshal(product)
	if err != nil {
		return err
	}

	return s.cache.Set(s.ctx, cacheKey, data, s.cacheTTL).Err()
}

// UpdateProduct updates a product and invalidates cache
func (s *CachedStore) UpdateProduct(product *Product) error {
	// Update in database
	if err := s.db.Save(product).Error; err != nil {
		return err
	}

	// Invalidate cache
	cacheKey := s.getCacheKey(product.ID)
	s.cache.Del(s.ctx, cacheKey)

	// Optionally, re-cache the updated product
	return s.cacheProduct(product)
}

// DeleteProduct deletes a product and invalidates cache
func (s *CachedStore) DeleteProduct(id uint) error {
	// Delete from database
	if err := s.db.Delete(&Product{}, id).Error; err != nil {
		return err
	}

	// Invalidate cache
	cacheKey := s.getCacheKey(id)
	return s.cache.Del(s.ctx, cacheKey).Err()
}

// GetAllProducts retrieves all products (bypasses cache)
func (s *CachedStore) GetAllProducts() ([]Product, error) {
	var products []Product
	err := s.db.Find(&products).Error
	return products, err
}

// SearchProducts searches products by name (bypasses cache)
func (s *CachedStore) SearchProducts(query string) ([]Product, error) {
	var products []Product
	err := s.db.Where("name LIKE ?", "%"+query+"%").Find(&products).Error
	return products, err
}

// InvalidateCache removes a product from cache
func (s *CachedStore) InvalidateCache(id uint) error {
	cacheKey := s.getCacheKey(id)
	return s.cache.Del(s.ctx, cacheKey).Err()
}

// InvalidateAllCache clears all cached products
func (s *CachedStore) InvalidateAllCache() error {
	// In a real application, you'd use a pattern match
	// For simplicity, we'll just flush the DB
	return s.cache.FlushDB(s.ctx).Err()
}

// GetCacheStats returns cache statistics
func (s *CachedStore) GetCacheStats() CacheStats {
	return s.stats
}

// ResetStats resets cache statistics
func (s *CachedStore) ResetStats() {
	s.stats = CacheStats{}
}

// GetProductsWithCache retrieves multiple products using cache
func (s *CachedStore) GetProductsWithCache(ids []uint) ([]Product, error) {
	products := make([]Product, 0, len(ids))

	for _, id := range ids {
		product, err := s.GetProduct(id)
		if err != nil {
			// Skip products that don't exist
			continue
		}
		products = append(products, *product)
	}

	return products, nil
}

// WarmCache pre-loads products into cache
func (s *CachedStore) WarmCache(ids []uint) error {
	for _, id := range ids {
		var product Product
		if err := s.db.First(&product, id).Error; err != nil {
			continue
		}

		if err := s.cacheProduct(&product); err != nil {
			return err
		}
	}

	return nil
}

func main() {
	store, err := NewCachedStore(":memory:", "localhost:6379", 5*time.Minute)
	if err != nil {
		fmt.Printf("Error creating store: %v\n", err)
		return
	}
	defer store.Close()

	// Create products
	product1 := &Product{
		Name:        "Laptop",
		Description: "High-performance laptop",
		Price:       999.99,
		Stock:       10,
	}

	store.CreateProduct(product1)
	fmt.Printf("Created product: %+v\n", product1)

	// First retrieval - cache miss
	retrieved1, _ := store.GetProduct(product1.ID)
	fmt.Printf("Retrieved (miss): %+v\n", retrieved1)

	// Second retrieval - cache hit
	retrieved2, _ := store.GetProduct(product1.ID)
	fmt.Printf("Retrieved (hit): %+v\n", retrieved2)

	// Get stats
	stats := store.GetCacheStats()
	fmt.Printf("Cache Stats: Hits=%d, Misses=%d\n", stats.Hits, stats.Misses)

	// Update product (invalidates cache)
	product1.Price = 899.99
	store.UpdateProduct(product1)
	fmt.Println("Product updated")

	// Retrieve again - cache miss after invalidation
	retrieved3, _ := store.GetProduct(product1.ID)
	fmt.Printf("Retrieved (after update): %+v\n", retrieved3)

	// Final stats
	finalStats := store.GetCacheStats()
	fmt.Printf("Final Stats: Hits=%d, Misses=%d\n", finalStats.Hits, finalStats.Misses)
}
