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

	if err := db.CreateTable(); err != nil {
		t.Fatalf("Failed to create table: %v", err)
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
		t.Fatal("Expected db connection to be initialized")
	}
}

func TestCreateTable(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Verify table exists by trying to query it
	_, err := db.db.Exec("SELECT * FROM users")
	if err != nil {
		t.Fatalf("Table was not created: %v", err)
	}
}

func TestInsertUser(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	user, err := db.InsertUser("John Doe", "john@example.com", 30)
	if err != nil {
		t.Fatalf("Failed to insert user: %v", err)
	}

	if user.ID == 0 {
		t.Error("Expected ID to be set")
	}

	if user.Name != "John Doe" {
		t.Errorf("Expected name 'John Doe', got %s", user.Name)
	}

	if user.Email != "john@example.com" {
		t.Errorf("Expected email 'john@example.com', got %s", user.Email)
	}

	if user.Age != 30 {
		t.Errorf("Expected age 30, got %d", user.Age)
	}
}

func TestInsertUserDuplicate(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	_, err := db.InsertUser("John Doe", "john@example.com", 30)
	if err != nil {
		t.Fatalf("Failed to insert user: %v", err)
	}

	// Try to insert duplicate email
	_, err = db.InsertUser("Jane Doe", "john@example.com", 25)
	if err == nil {
		t.Error("Expected error when inserting duplicate email")
	}
}

func TestGetUser(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	inserted, _ := db.InsertUser("Alice", "alice@example.com", 28)

	user, err := db.GetUser(inserted.ID)
	if err != nil {
		t.Fatalf("Failed to get user: %v", err)
	}

	if user.ID != inserted.ID {
		t.Errorf("Expected ID %d, got %d", inserted.ID, user.ID)
	}

	if user.Name != "Alice" {
		t.Errorf("Expected name 'Alice', got %s", user.Name)
	}

	if user.Email != "alice@example.com" {
		t.Errorf("Expected email 'alice@example.com', got %s", user.Email)
	}

	if user.Age != 28 {
		t.Errorf("Expected age 28, got %d", user.Age)
	}
}

func TestGetUserNotFound(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	_, err := db.GetUser(999)
	if err != sql.ErrNoRows {
		t.Errorf("Expected sql.ErrNoRows, got %v", err)
	}
}

func TestGetAllUsers(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Insert multiple users
	db.InsertUser("Alice", "alice@example.com", 28)
	db.InsertUser("Bob", "bob@example.com", 32)
	db.InsertUser("Charlie", "charlie@example.com", 25)

	users, err := db.GetAllUsers()
	if err != nil {
		t.Fatalf("Failed to get all users: %v", err)
	}

	if len(users) != 3 {
		t.Errorf("Expected 3 users, got %d", len(users))
	}

	// Verify order
	if users[0].Name != "Alice" {
		t.Errorf("Expected first user to be Alice, got %s", users[0].Name)
	}
}

func TestGetAllUsersEmpty(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	users, err := db.GetAllUsers()
	if err != nil {
		t.Fatalf("Failed to get all users: %v", err)
	}

	if len(users) != 0 {
		t.Errorf("Expected 0 users, got %d", len(users))
	}
}

func TestUpdateUser(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	user, _ := db.InsertUser("Alice", "alice@example.com", 28)

	err := db.UpdateUser(user.ID, "Alice Smith", "alice.smith@example.com", 29)
	if err != nil {
		t.Fatalf("Failed to update user: %v", err)
	}

	updated, _ := db.GetUser(user.ID)
	if updated.Name != "Alice Smith" {
		t.Errorf("Expected name 'Alice Smith', got %s", updated.Name)
	}

	if updated.Email != "alice.smith@example.com" {
		t.Errorf("Expected email 'alice.smith@example.com', got %s", updated.Email)
	}

	if updated.Age != 29 {
		t.Errorf("Expected age 29, got %d", updated.Age)
	}
}

func TestUpdateUserNotFound(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	err := db.UpdateUser(999, "Nobody", "nobody@example.com", 0)
	if err != sql.ErrNoRows {
		t.Errorf("Expected sql.ErrNoRows, got %v", err)
	}
}

func TestDeleteUser(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	user, _ := db.InsertUser("Alice", "alice@example.com", 28)

	err := db.DeleteUser(user.ID)
	if err != nil {
		t.Fatalf("Failed to delete user: %v", err)
	}

	_, err = db.GetUser(user.ID)
	if err != sql.ErrNoRows {
		t.Error("Expected user to be deleted")
	}
}

func TestDeleteUserNotFound(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	err := db.DeleteUser(999)
	if err != sql.ErrNoRows {
		t.Errorf("Expected sql.ErrNoRows, got %v", err)
	}
}

func TestCRUDWorkflow(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Create
	user, err := db.InsertUser("Test User", "test@example.com", 30)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Read
	retrieved, err := db.GetUser(user.ID)
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}
	if retrieved.Name != "Test User" {
		t.Error("Read returned wrong user")
	}

	// Update
	err = db.UpdateUser(user.ID, "Updated User", "updated@example.com", 31)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	updated, _ := db.GetUser(user.ID)
	if updated.Name != "Updated User" {
		t.Error("Update did not change name")
	}

	// Delete
	err = db.DeleteUser(user.ID)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	_, err = db.GetUser(user.ID)
	if err != sql.ErrNoRows {
		t.Error("User was not deleted")
	}
}
