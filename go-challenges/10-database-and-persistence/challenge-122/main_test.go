package main

import (
	"testing"

	"gorm.io/gorm"
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
}

func TestCreateUser(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	user := &User{
		Name:   "Test User",
		Email:  "test@example.com",
		Age:    25,
		Active: true,
	}

	err := db.CreateUser(user)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	if user.ID == 0 {
		t.Error("Expected ID to be set")
	}

	if user.CreatedAt.IsZero() {
		t.Error("Expected CreatedAt to be set")
	}

	if user.UpdatedAt.IsZero() {
		t.Error("Expected UpdatedAt to be set")
	}
}

func TestGetUser(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	user := &User{
		Name:   "Alice",
		Email:  "alice@example.com",
		Age:    28,
		Active: true,
	}

	db.CreateUser(user)

	retrieved, err := db.GetUser(user.ID)
	if err != nil {
		t.Fatalf("Failed to get user: %v", err)
	}

	if retrieved.ID != user.ID {
		t.Errorf("Expected ID %d, got %d", user.ID, retrieved.ID)
	}

	if retrieved.Name != "Alice" {
		t.Errorf("Expected name 'Alice', got %s", retrieved.Name)
	}

	if retrieved.Email != "alice@example.com" {
		t.Errorf("Expected email 'alice@example.com', got %s", retrieved.Email)
	}

	if retrieved.Age != 28 {
		t.Errorf("Expected age 28, got %d", retrieved.Age)
	}
}

func TestGetUserNotFound(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	_, err := db.GetUser(999)
	if err != gorm.ErrRecordNotFound {
		t.Errorf("Expected gorm.ErrRecordNotFound, got %v", err)
	}
}

func TestGetUserByEmail(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	user := &User{
		Name:  "Bob",
		Email: "bob@example.com",
		Age:   35,
	}

	db.CreateUser(user)

	retrieved, err := db.GetUserByEmail("bob@example.com")
	if err != nil {
		t.Fatalf("Failed to get user by email: %v", err)
	}

	if retrieved.Name != "Bob" {
		t.Errorf("Expected name 'Bob', got %s", retrieved.Name)
	}
}

func TestGetUserByEmailNotFound(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	_, err := db.GetUserByEmail("nonexistent@example.com")
	if err != gorm.ErrRecordNotFound {
		t.Errorf("Expected gorm.ErrRecordNotFound, got %v", err)
	}
}

func TestGetAllUsers(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	users := []User{
		{Name: "User1", Email: "user1@example.com", Age: 20},
		{Name: "User2", Email: "user2@example.com", Age: 25},
		{Name: "User3", Email: "user3@example.com", Age: 30},
	}

	for _, u := range users {
		user := u
		db.CreateUser(&user)
	}

	retrieved, err := db.GetAllUsers()
	if err != nil {
		t.Fatalf("Failed to get all users: %v", err)
	}

	if len(retrieved) != 3 {
		t.Errorf("Expected 3 users, got %d", len(retrieved))
	}
}

func TestGetActiveUsers(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	db.CreateUser(&User{Name: "Active1", Email: "active1@example.com", Age: 20, Active: true})
	db.CreateUser(&User{Name: "Active2", Email: "active2@example.com", Age: 25, Active: true})
	db.CreateUser(&User{Name: "Inactive", Email: "inactive@example.com", Age: 30, Active: false})

	active, err := db.GetActiveUsers()
	if err != nil {
		t.Fatalf("Failed to get active users: %v", err)
	}

	if len(active) != 2 {
		t.Errorf("Expected 2 active users, got %d", len(active))
	}

	for _, user := range active {
		if !user.Active {
			t.Error("Retrieved inactive user")
		}
	}
}

func TestUpdateUser(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	user := &User{
		Name:  "Original",
		Email: "original@example.com",
		Age:   25,
	}

	db.CreateUser(user)

	user.Name = "Updated"
	user.Age = 26

	err := db.UpdateUser(user)
	if err != nil {
		t.Fatalf("Failed to update user: %v", err)
	}

	retrieved, _ := db.GetUser(user.ID)
	if retrieved.Name != "Updated" {
		t.Errorf("Expected name 'Updated', got %s", retrieved.Name)
	}

	if retrieved.Age != 26 {
		t.Errorf("Expected age 26, got %d", retrieved.Age)
	}
}

func TestUpdateUserFields(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	user := &User{
		Name:  "Alice",
		Email: "alice@example.com",
		Age:   25,
	}

	db.CreateUser(user)

	updates := map[string]interface{}{
		"age": 30,
	}

	err := db.UpdateUserFields(user.ID, updates)
	if err != nil {
		t.Fatalf("Failed to update user fields: %v", err)
	}

	retrieved, _ := db.GetUser(user.ID)
	if retrieved.Age != 30 {
		t.Errorf("Expected age 30, got %d", retrieved.Age)
	}

	if retrieved.Name != "Alice" {
		t.Error("Name should not have changed")
	}
}

func TestDeleteUser(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	user := &User{
		Name:  "ToDelete",
		Email: "delete@example.com",
		Age:   25,
	}

	db.CreateUser(user)

	err := db.DeleteUser(user.ID)
	if err != nil {
		t.Fatalf("Failed to delete user: %v", err)
	}

	// Soft delete - should not be found
	_, err = db.GetUser(user.ID)
	if err != gorm.ErrRecordNotFound {
		t.Error("User should be soft deleted")
	}
}

func TestHardDeleteUser(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	user := &User{
		Name:  "ToHardDelete",
		Email: "harddelete@example.com",
		Age:   25,
	}

	db.CreateUser(user)

	err := db.HardDeleteUser(user.ID)
	if err != nil {
		t.Fatalf("Failed to hard delete user: %v", err)
	}

	_, err = db.GetUser(user.ID)
	if err != gorm.ErrRecordNotFound {
		t.Error("User should be permanently deleted")
	}
}

func TestCreateProduct(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	product := &Product{
		Name:        "Test Product",
		Description: "Test Description",
		Price:       99.99,
		Stock:       10,
	}

	err := db.CreateProduct(product)
	if err != nil {
		t.Fatalf("Failed to create product: %v", err)
	}

	if product.ID == 0 {
		t.Error("Expected ID to be set")
	}
}

func TestGetProduct(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	product := &Product{
		Name:  "Laptop",
		Price: 999.99,
		Stock: 5,
	}

	db.CreateProduct(product)

	retrieved, err := db.GetProduct(product.ID)
	if err != nil {
		t.Fatalf("Failed to get product: %v", err)
	}

	if retrieved.Name != "Laptop" {
		t.Errorf("Expected name 'Laptop', got %s", retrieved.Name)
	}

	if retrieved.Price != 999.99 {
		t.Errorf("Expected price 999.99, got %f", retrieved.Price)
	}
}

func TestGetProductsByPriceRange(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	products := []Product{
		{Name: "Cheap", Price: 10.0, Stock: 10},
		{Name: "Medium", Price: 50.0, Stock: 5},
		{Name: "Expensive", Price: 100.0, Stock: 2},
	}

	for _, p := range products {
		product := p
		db.CreateProduct(&product)
	}

	results, err := db.GetProductsByPriceRange(20.0, 80.0)
	if err != nil {
		t.Fatalf("Failed to get products by price range: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("Expected 1 product, got %d", len(results))
	}

	if len(results) > 0 && results[0].Name != "Medium" {
		t.Errorf("Expected product 'Medium', got %s", results[0].Name)
	}
}

func TestSearchProducts(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	db.CreateProduct(&Product{Name: "Laptop Dell", Price: 999.99, Stock: 5})
	db.CreateProduct(&Product{Name: "Laptop HP", Price: 899.99, Stock: 3})
	db.CreateProduct(&Product{Name: "Mouse", Price: 29.99, Stock: 20})

	results, err := db.SearchProducts("Laptop")
	if err != nil {
		t.Fatalf("Failed to search products: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 products, got %d", len(results))
	}
}

func TestUpdateProductStock(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	product := &Product{
		Name:  "Test Product",
		Price: 50.0,
		Stock: 10,
	}

	db.CreateProduct(product)

	err := db.UpdateProductStock(product.ID, 20)
	if err != nil {
		t.Fatalf("Failed to update product stock: %v", err)
	}

	retrieved, _ := db.GetProduct(product.ID)
	if retrieved.Stock != 20 {
		t.Errorf("Expected stock 20, got %d", retrieved.Stock)
	}
}

func TestDeleteProduct(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	product := &Product{
		Name:  "ToDelete",
		Price: 50.0,
		Stock: 5,
	}

	db.CreateProduct(product)

	err := db.DeleteProduct(product.ID)
	if err != nil {
		t.Fatalf("Failed to delete product: %v", err)
	}

	_, err = db.GetProduct(product.ID)
	if err != gorm.ErrRecordNotFound {
		t.Error("Product should be soft deleted")
	}
}

func TestBatchCreateUsers(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	users := []User{
		{Name: "Batch1", Email: "batch1@example.com", Age: 20},
		{Name: "Batch2", Email: "batch2@example.com", Age: 25},
		{Name: "Batch3", Email: "batch3@example.com", Age: 30},
	}

	err := db.BatchCreateUsers(users)
	if err != nil {
		t.Fatalf("Failed to batch create users: %v", err)
	}

	all, _ := db.GetAllUsers()
	if len(all) != 3 {
		t.Errorf("Expected 3 users, got %d", len(all))
	}
}

func TestCountUsers(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	for i := 0; i < 5; i++ {
		db.CreateUser(&User{
			Name:  "User",
			Email: fmt.Sprintf("user%d@example.com", i),
			Age:   20,
		})
	}

	count, err := db.CountUsers()
	if err != nil {
		t.Fatalf("Failed to count users: %v", err)
	}

	if count != 5 {
		t.Errorf("Expected count 5, got %d", count)
	}
}

func TestGetUsersPaginated(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	for i := 0; i < 10; i++ {
		db.CreateUser(&User{
			Name:  "User",
			Email: fmt.Sprintf("user%d@example.com", i),
			Age:   20,
		})
	}

	// Get first page
	page1, err := db.GetUsersPaginated(1, 3)
	if err != nil {
		t.Fatalf("Failed to get page 1: %v", err)
	}

	if len(page1) != 3 {
		t.Errorf("Expected 3 users on page 1, got %d", len(page1))
	}

	// Get second page
	page2, err := db.GetUsersPaginated(2, 3)
	if err != nil {
		t.Fatalf("Failed to get page 2: %v", err)
	}

	if len(page2) != 3 {
		t.Errorf("Expected 3 users on page 2, got %d", len(page2))
	}

	// Verify different users
	if page1[0].ID == page2[0].ID {
		t.Error("Page 1 and Page 2 should contain different users")
	}
}

func TestGetUsersOrderedByAge(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	db.CreateUser(&User{Name: "Young", Email: "young@example.com", Age: 20})
	db.CreateUser(&User{Name: "Old", Email: "old@example.com", Age: 50})
	db.CreateUser(&User{Name: "Middle", Email: "middle@example.com", Age: 35})

	// Ascending order
	ascending, err := db.GetUsersOrderedByAge(true)
	if err != nil {
		t.Fatalf("Failed to get users ordered ascending: %v", err)
	}

	if len(ascending) != 3 {
		t.Errorf("Expected 3 users, got %d", len(ascending))
	}

	if ascending[0].Age != 20 || ascending[2].Age != 50 {
		t.Error("Users not ordered correctly in ascending order")
	}

	// Descending order
	descending, err := db.GetUsersOrderedByAge(false)
	if err != nil {
		t.Fatalf("Failed to get users ordered descending: %v", err)
	}

	if descending[0].Age != 50 || descending[2].Age != 20 {
		t.Error("Users not ordered correctly in descending order")
	}
}
