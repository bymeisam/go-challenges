package main

import (
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"gorm.io/gorm"
)

func setupTestStore(t *testing.T) (*CachedStore, *miniredis.Miniredis) {
	// Create miniredis server
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("Failed to start miniredis: %v", err)
	}

	// Create cached store
	store, err := NewCachedStore(":memory:", mr.Addr(), 5*time.Minute)
	if err != nil {
		mr.Close()
		t.Fatalf("Failed to create store: %v", err)
	}

	return store, mr
}

func TestNewCachedStore(t *testing.T) {
	store, mr := setupTestStore(t)
	defer mr.Close()
	defer store.Close()

	if store.db == nil {
		t.Error("Expected database to be initialized")
	}

	if store.cache == nil {
		t.Error("Expected cache to be initialized")
	}
}

func TestCreateProduct(t *testing.T) {
	store, mr := setupTestStore(t)
	defer mr.Close()
	defer store.Close()

	product := &Product{
		Name:        "Test Product",
		Description: "Test Description",
		Price:       99.99,
		Stock:       10,
	}

	err := store.CreateProduct(product)
	if err != nil {
		t.Fatalf("Failed to create product: %v", err)
	}

	if product.ID == 0 {
		t.Error("Expected ID to be set")
	}
}

func TestGetProductCacheHit(t *testing.T) {
	store, mr := setupTestStore(t)
	defer mr.Close()
	defer store.Close()

	// Create product
	product := &Product{
		Name:  "Cached Product",
		Price: 50.0,
		Stock: 5,
	}
	store.CreateProduct(product)

	// First retrieval - cache miss
	store.ResetStats()
	retrieved1, err := store.GetProduct(product.ID)
	if err != nil {
		t.Fatalf("Failed to get product: %v", err)
	}

	stats1 := store.GetCacheStats()
	if stats1.Misses != 1 {
		t.Errorf("Expected 1 cache miss, got %d", stats1.Misses)
	}

	// Second retrieval - cache hit
	retrieved2, err := store.GetProduct(product.ID)
	if err != nil {
		t.Fatalf("Failed to get product: %v", err)
	}

	stats2 := store.GetCacheStats()
	if stats2.Hits != 1 {
		t.Errorf("Expected 1 cache hit, got %d", stats2.Hits)
	}

	if retrieved1.Name != retrieved2.Name {
		t.Error("Retrieved products should be identical")
	}
}

func TestGetProductNotFound(t *testing.T) {
	store, mr := setupTestStore(t)
	defer mr.Close()
	defer store.Close()

	_, err := store.GetProduct(999)
	if err != gorm.ErrRecordNotFound {
		t.Errorf("Expected gorm.ErrRecordNotFound, got %v", err)
	}

	stats := store.GetCacheStats()
	if stats.Misses != 1 {
		t.Errorf("Expected 1 cache miss, got %d", stats.Misses)
	}
}

func TestUpdateProductInvalidatesCache(t *testing.T) {
	store, mr := setupTestStore(t)
	defer mr.Close()
	defer store.Close()

	// Create product
	product := &Product{
		Name:  "Original",
		Price: 100.0,
		Stock: 10,
	}
	store.CreateProduct(product)

	// Get product to cache it
	store.GetProduct(product.ID)

	// Update product
	product.Name = "Updated"
	product.Price = 120.0
	err := store.UpdateProduct(product)
	if err != nil {
		t.Fatalf("Failed to update product: %v", err)
	}

	// Get product again - should retrieve updated version
	store.ResetStats()
	retrieved, _ := store.GetProduct(product.ID)

	if retrieved.Name != "Updated" {
		t.Errorf("Expected name 'Updated', got %s", retrieved.Name)
	}

	if retrieved.Price != 120.0 {
		t.Errorf("Expected price 120.0, got %f", retrieved.Price)
	}

	// Should be a cache hit since we re-cache after update
	stats := store.GetCacheStats()
	if stats.Hits != 1 {
		t.Errorf("Expected 1 cache hit after update, got %d", stats.Hits)
	}
}

func TestDeleteProductInvalidatesCache(t *testing.T) {
	store, mr := setupTestStore(t)
	defer mr.Close()
	defer store.Close()

	// Create product
	product := &Product{
		Name:  "ToDelete",
		Price: 50.0,
		Stock: 5,
	}
	store.CreateProduct(product)

	// Get product to cache it
	store.GetProduct(product.ID)

	// Delete product
	err := store.DeleteProduct(product.ID)
	if err != nil {
		t.Fatalf("Failed to delete product: %v", err)
	}

	// Try to get product - should fail
	_, err = store.GetProduct(product.ID)
	if err != gorm.ErrRecordNotFound {
		t.Error("Product should be deleted")
	}
}

func TestGetAllProducts(t *testing.T) {
	store, mr := setupTestStore(t)
	defer mr.Close()
	defer store.Close()

	// Create multiple products
	for i := 0; i < 5; i++ {
		product := &Product{
			Name:  "Product",
			Price: 10.0,
			Stock: 1,
		}
		store.CreateProduct(product)
	}

	products, err := store.GetAllProducts()
	if err != nil {
		t.Fatalf("Failed to get all products: %v", err)
	}

	if len(products) != 5 {
		t.Errorf("Expected 5 products, got %d", len(products))
	}
}

func TestSearchProducts(t *testing.T) {
	store, mr := setupTestStore(t)
	defer mr.Close()
	defer store.Close()

	store.CreateProduct(&Product{Name: "Laptop Dell", Price: 999, Stock: 5})
	store.CreateProduct(&Product{Name: "Laptop HP", Price: 899, Stock: 3})
	store.CreateProduct(&Product{Name: "Mouse", Price: 29, Stock: 10})

	results, err := store.SearchProducts("Laptop")
	if err != nil {
		t.Fatalf("Failed to search products: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}
}

func TestInvalidateCache(t *testing.T) {
	store, mr := setupTestStore(t)
	defer mr.Close()
	defer store.Close()

	// Create and cache product
	product := &Product{Name: "Test", Price: 50, Stock: 5}
	store.CreateProduct(product)
	store.GetProduct(product.ID)

	// Invalidate cache
	err := store.InvalidateCache(product.ID)
	if err != nil {
		t.Fatalf("Failed to invalidate cache: %v", err)
	}

	// Next retrieval should be a cache miss
	store.ResetStats()
	store.GetProduct(product.ID)

	stats := store.GetCacheStats()
	if stats.Misses != 1 {
		t.Errorf("Expected 1 cache miss after invalidation, got %d", stats.Misses)
	}
}

func TestInvalidateAllCache(t *testing.T) {
	store, mr := setupTestStore(t)
	defer mr.Close()
	defer store.Close()

	// Create and cache multiple products
	product1 := &Product{Name: "Product 1", Price: 10, Stock: 1}
	product2 := &Product{Name: "Product 2", Price: 20, Stock: 2}

	store.CreateProduct(product1)
	store.CreateProduct(product2)

	store.GetProduct(product1.ID)
	store.GetProduct(product2.ID)

	// Invalidate all
	err := store.InvalidateAllCache()
	if err != nil {
		t.Fatalf("Failed to invalidate all cache: %v", err)
	}

	// Next retrievals should be cache misses
	store.ResetStats()
	store.GetProduct(product1.ID)
	store.GetProduct(product2.ID)

	stats := store.GetCacheStats()
	if stats.Misses != 2 {
		t.Errorf("Expected 2 cache misses, got %d", stats.Misses)
	}
}

func TestGetCacheStats(t *testing.T) {
	store, mr := setupTestStore(t)
	defer mr.Close()
	defer store.Close()

	product := &Product{Name: "Test", Price: 50, Stock: 5}
	store.CreateProduct(product)

	store.ResetStats()

	// Generate some cache activity
	store.GetProduct(product.ID)  // Miss
	store.GetProduct(product.ID)  // Hit
	store.GetProduct(product.ID)  // Hit
	store.GetProduct(999)         // Miss (not found)

	stats := store.GetCacheStats()

	if stats.Hits != 2 {
		t.Errorf("Expected 2 hits, got %d", stats.Hits)
	}

	if stats.Misses != 2 {
		t.Errorf("Expected 2 misses, got %d", stats.Misses)
	}
}

func TestResetStats(t *testing.T) {
	store, mr := setupTestStore(t)
	defer mr.Close()
	defer store.Close()

	product := &Product{Name: "Test", Price: 50, Stock: 5}
	store.CreateProduct(product)

	// Generate activity
	store.GetProduct(product.ID)
	store.GetProduct(product.ID)

	// Reset stats
	store.ResetStats()

	stats := store.GetCacheStats()
	if stats.Hits != 0 || stats.Misses != 0 {
		t.Error("Stats should be reset to zero")
	}
}

func TestGetProductsWithCache(t *testing.T) {
	store, mr := setupTestStore(t)
	defer mr.Close()
	defer store.Close()

	// Create products
	ids := make([]uint, 5)
	for i := 0; i < 5; i++ {
		product := &Product{Name: "Product", Price: 10, Stock: 1}
		store.CreateProduct(product)
		ids[i] = product.ID
	}

	store.ResetStats()

	// First retrieval - all misses
	products1, err := store.GetProductsWithCache(ids)
	if err != nil {
		t.Fatalf("Failed to get products: %v", err)
	}

	if len(products1) != 5 {
		t.Errorf("Expected 5 products, got %d", len(products1))
	}

	stats1 := store.GetCacheStats()
	if stats1.Misses != 5 {
		t.Errorf("Expected 5 cache misses, got %d", stats1.Misses)
	}

	// Second retrieval - all hits
	products2, err := store.GetProductsWithCache(ids)
	if err != nil {
		t.Fatalf("Failed to get products: %v", err)
	}

	if len(products2) != 5 {
		t.Errorf("Expected 5 products, got %d", len(products2))
	}

	stats2 := store.GetCacheStats()
	if stats2.Hits != 5 {
		t.Errorf("Expected 5 cache hits, got %d", stats2.Hits)
	}
}

func TestWarmCache(t *testing.T) {
	store, mr := setupTestStore(t)
	defer mr.Close()
	defer store.Close()

	// Create products
	ids := make([]uint, 3)
	for i := 0; i < 3; i++ {
		product := &Product{Name: "Product", Price: 10, Stock: 1}
		store.CreateProduct(product)
		ids[i] = product.ID
	}

	// Clear cache
	store.InvalidateAllCache()
	store.ResetStats()

	// Warm cache
	err := store.WarmCache(ids)
	if err != nil {
		t.Fatalf("Failed to warm cache: %v", err)
	}

	// Retrievals should all be cache hits
	for _, id := range ids {
		store.GetProduct(id)
	}

	stats := store.GetCacheStats()
	if stats.Hits != 3 {
		t.Errorf("Expected 3 cache hits after warming, got %d", stats.Hits)
	}

	if stats.Misses != 0 {
		t.Errorf("Expected 0 cache misses after warming, got %d", stats.Misses)
	}
}

func TestCacheTTL(t *testing.T) {
	// Create store with very short TTL for testing
	mr, _ := miniredis.Run()
	defer mr.Close()

	store, _ := NewCachedStore(":memory:", mr.Addr(), 100*time.Millisecond)
	defer store.Close()

	product := &Product{Name: "Test", Price: 50, Stock: 5}
	store.CreateProduct(product)

	// Get product to cache it
	store.GetProduct(product.ID)

	// Fast forward time in miniredis
	mr.FastForward(150 * time.Millisecond)

	// Next retrieval should be a cache miss due to expiration
	store.ResetStats()
	store.GetProduct(product.ID)

	stats := store.GetCacheStats()
	if stats.Misses != 1 {
		t.Errorf("Expected cache miss due to TTL expiration, got %d misses", stats.Misses)
	}
}

func TestConcurrentAccess(t *testing.T) {
	store, mr := setupTestStore(t)
	defer mr.Close()
	defer store.Close()

	product := &Product{Name: "Concurrent", Price: 100, Stock: 10}
	store.CreateProduct(product)

	// Multiple concurrent reads
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			_, err := store.GetProduct(product.ID)
			if err != nil {
				t.Errorf("Concurrent read failed: %v", err)
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
}
