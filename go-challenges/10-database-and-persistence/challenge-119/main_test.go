package main

import (
	"database/sql"
	"testing"
)

func setupTestDB(t *testing.T) *Database {
	db, err := NewDatabase(":memory:")
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	return db
}

func TestNewDatabase(t *testing.T) {
	db, err := NewDatabase(":memory:")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	defer db.Close()

	if db.db == nil {
		t.Error("Expected db connection to be initialized")
	}

	if db.insertStmt == nil {
		t.Error("Expected insert statement to be prepared")
	}

	if db.getStmt == nil {
		t.Error("Expected get statement to be prepared")
	}

	if db.updateStockStmt == nil {
		t.Error("Expected update stock statement to be prepared")
	}

	if db.deleteStmt == nil {
		t.Error("Expected delete statement to be prepared")
	}

	if db.searchByNameStmt == nil {
		t.Error("Expected search by name statement to be prepared")
	}
}

func TestInsertProduct(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	product := &Product{
		Name:        "Test Product",
		Description: "Test Description",
		Price:       19.99,
		Stock:       5,
	}

	id, err := db.InsertProduct(product)
	if err != nil {
		t.Fatalf("Failed to insert product: %v", err)
	}

	if id == 0 {
		t.Error("Expected ID to be set")
	}
}

func TestInsertMultipleProducts(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	products := []*Product{
		{Name: "Product 1", Description: "Desc 1", Price: 10.0, Stock: 1},
		{Name: "Product 2", Description: "Desc 2", Price: 20.0, Stock: 2},
		{Name: "Product 3", Description: "Desc 3", Price: 30.0, Stock: 3},
	}

	for _, p := range products {
		id, err := db.InsertProduct(p)
		if err != nil {
			t.Fatalf("Failed to insert product: %v", err)
		}
		if id == 0 {
			t.Error("Expected ID to be set")
		}
	}
}

func TestGetProduct(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	product := &Product{
		Name:        "Laptop",
		Description: "Gaming laptop",
		Price:       1299.99,
		Stock:       3,
	}

	id, _ := db.InsertProduct(product)

	retrieved, err := db.GetProduct(id)
	if err != nil {
		t.Fatalf("Failed to get product: %v", err)
	}

	if retrieved.ID != id {
		t.Errorf("Expected ID %d, got %d", id, retrieved.ID)
	}

	if retrieved.Name != "Laptop" {
		t.Errorf("Expected name 'Laptop', got %s", retrieved.Name)
	}

	if retrieved.Description != "Gaming laptop" {
		t.Errorf("Expected description 'Gaming laptop', got %s", retrieved.Description)
	}

	if retrieved.Price != 1299.99 {
		t.Errorf("Expected price 1299.99, got %f", retrieved.Price)
	}

	if retrieved.Stock != 3 {
		t.Errorf("Expected stock 3, got %d", retrieved.Stock)
	}
}

func TestGetProductNotFound(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	_, err := db.GetProduct(999)
	if err != sql.ErrNoRows {
		t.Errorf("Expected sql.ErrNoRows, got %v", err)
	}
}

func TestUpdateStock(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	product := &Product{
		Name:  "Test Product",
		Price: 10.0,
		Stock: 10,
	}

	id, _ := db.InsertProduct(product)

	// Increase stock
	err := db.UpdateStock(id, 5)
	if err != nil {
		t.Fatalf("Failed to update stock: %v", err)
	}

	updated, _ := db.GetProduct(id)
	if updated.Stock != 15 {
		t.Errorf("Expected stock 15, got %d", updated.Stock)
	}

	// Decrease stock
	err = db.UpdateStock(id, -3)
	if err != nil {
		t.Fatalf("Failed to decrease stock: %v", err)
	}

	updated, _ = db.GetProduct(id)
	if updated.Stock != 12 {
		t.Errorf("Expected stock 12, got %d", updated.Stock)
	}
}

func TestUpdateStockNotFound(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	err := db.UpdateStock(999, 5)
	if err != sql.ErrNoRows {
		t.Errorf("Expected sql.ErrNoRows, got %v", err)
	}
}

func TestDeleteProduct(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	product := &Product{
		Name:  "Test Product",
		Price: 10.0,
		Stock: 5,
	}

	id, _ := db.InsertProduct(product)

	err := db.DeleteProduct(id)
	if err != nil {
		t.Fatalf("Failed to delete product: %v", err)
	}

	_, err = db.GetProduct(id)
	if err != sql.ErrNoRows {
		t.Error("Expected product to be deleted")
	}
}

func TestDeleteProductNotFound(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	err := db.DeleteProduct(999)
	if err != sql.ErrNoRows {
		t.Errorf("Expected sql.ErrNoRows, got %v", err)
	}
}

func TestSearchByName(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Insert test products
	products := []*Product{
		{Name: "Laptop Dell", Description: "Dell laptop", Price: 999.0, Stock: 5},
		{Name: "Laptop HP", Description: "HP laptop", Price: 899.0, Stock: 3},
		{Name: "Mouse", Description: "Wireless mouse", Price: 29.99, Stock: 10},
		{Name: "Keyboard", Description: "Mechanical keyboard", Price: 79.99, Stock: 7},
	}

	for _, p := range products {
		db.InsertProduct(p)
	}

	// Search for laptops
	results, err := db.SearchByName("Laptop")
	if err != nil {
		t.Fatalf("Failed to search: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}

	// Verify results contain "Laptop" in name
	for _, r := range results {
		if r.Name != "Laptop Dell" && r.Name != "Laptop HP" {
			t.Errorf("Unexpected product in results: %s", r.Name)
		}
	}
}

func TestSearchByNameNoResults(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	results, err := db.SearchByName("NonExistent")
	if err != nil {
		t.Fatalf("Failed to search: %v", err)
	}

	if len(results) != 0 {
		t.Errorf("Expected 0 results, got %d", len(results))
	}
}

func TestSearchByNamePartialMatch(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	db.InsertProduct(&Product{Name: "Apple iPhone", Price: 999, Stock: 5})
	db.InsertProduct(&Product{Name: "Apple iPad", Price: 599, Stock: 3})
	db.InsertProduct(&Product{Name: "Samsung Phone", Price: 799, Stock: 4})

	results, err := db.SearchByName("Apple")
	if err != nil {
		t.Fatalf("Failed to search: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}
}

func TestPreparedStatementsReuse(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Insert multiple products using the same prepared statement
	for i := 0; i < 10; i++ {
		product := &Product{
			Name:  "Product",
			Price: 10.0,
			Stock: 1,
		}

		_, err := db.InsertProduct(product)
		if err != nil {
			t.Fatalf("Failed to insert product %d: %v", i, err)
		}
	}

	// Verify all products were inserted
	products, err := db.SearchByName("Product")
	if err != nil {
		t.Fatalf("Failed to search products: %v", err)
	}

	if len(products) != 10 {
		t.Errorf("Expected 10 products, got %d", len(products))
	}
}

func TestBenchmarkInserts(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Benchmark with prepared statements
	preparedTime := db.BenchmarkInserts(50, true)
	if preparedTime <= 0 {
		t.Error("Expected positive duration for prepared statements")
	}

	// Clear table
	db.db.Exec("DELETE FROM products")

	// Benchmark without prepared statements
	unpreparedTime := db.BenchmarkInserts(50, false)
	if unpreparedTime <= 0 {
		t.Error("Expected positive duration for unprepared statements")
	}

	// Prepared statements should generally be faster, but we won't enforce
	// this in the test as it can vary
	t.Logf("Prepared: %v, Unprepared: %v", preparedTime, unpreparedTime)
}

func TestClose(t *testing.T) {
	db := setupTestDB(t)

	err := db.Close()
	if err != nil {
		t.Fatalf("Failed to close database: %v", err)
	}

	// Verify that operations fail after closing
	_, err = db.InsertProduct(&Product{Name: "Test", Price: 10, Stock: 1})
	if err == nil {
		t.Error("Expected error when using closed database")
	}
}
