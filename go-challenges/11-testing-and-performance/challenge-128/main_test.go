package main

import (
	"os"
	"path/filepath"
	"testing"
)

var testDB *Database

// TestMain provides suite-level setup and teardown
func TestMain(m *testing.M) {
	// Setup: create a test database connection
	db, err := Connect("test://localhost/testdb")
	if err != nil {
		panic("failed to connect to test database: " + err.Error())
	}
	testDB = db

	// Run all tests
	code := m.Run()

	// Teardown: close database connection
	testDB.Close()

	// Exit with test result code
	os.Exit(code)
}

func TestFileStorage(t *testing.T) {
	// Create temporary directory for this test
	// t.TempDir() automatically cleans up after the test
	tempDir := t.TempDir()

	storage := New(tempDir)

	t.Run("Save and Load", func(t *testing.T) {
		filename := "test.txt"
		content := "Hello, World!"

		// Save file
		err := storage.Save(filename, content)
		if err != nil {
			t.Fatalf("Save failed: %v", err)
		}

		// Load file
		loaded, err := storage.Load(filename)
		if err != nil {
			t.Fatalf("Load failed: %v", err)
		}

		if loaded != content {
			t.Errorf("loaded content = %q; want %q", loaded, content)
		}
	})

	t.Run("List files", func(t *testing.T) {
		// Save multiple files
		files := map[string]string{
			"file1.txt": "content1",
			"file2.txt": "content2",
			"file3.txt": "content3",
		}

		for name, content := range files {
			if err := storage.Save(name, content); err != nil {
				t.Fatalf("Save failed: %v", err)
			}
		}

		// List files
		list, err := storage.List()
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}

		// At least these 3 files should exist (may have more from other subtests)
		if len(list) < 3 {
			t.Errorf("expected at least 3 files, got %d", len(list))
		}
	})

	t.Run("Delete file", func(t *testing.T) {
		filename := "to-delete.txt"
		content := "delete me"

		// Save file
		err := storage.Save(filename, content)
		if err != nil {
			t.Fatalf("Save failed: %v", err)
		}

		// Delete file
		err = storage.Delete(filename)
		if err != nil {
			t.Fatalf("Delete failed: %v", err)
		}

		// Try to load deleted file
		_, err = storage.Load(filename)
		if err == nil {
			t.Error("expected error when loading deleted file")
		}
	})

	t.Run("Load non-existent file", func(t *testing.T) {
		_, err := storage.Load("nonexistent.txt")
		if err == nil {
			t.Error("expected error when loading non-existent file")
		}
	})

	t.Log("✓ All FileStorage tests passed!")
}

func TestFileStorageWithCleanup(t *testing.T) {
	// Alternative pattern: manual temp dir with t.Cleanup()
	tempDir, err := os.MkdirTemp("", "test-storage-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	// Register cleanup function
	t.Cleanup(func() {
		os.RemoveAll(tempDir)
	})

	storage := New(tempDir)

	// Test operations
	err = storage.Save("test.txt", "test content")
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Verify file exists
	path := filepath.Join(tempDir, "test.txt")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("file should exist")
	}

	t.Log("✓ FileStorage with cleanup test passed!")
}

func TestDatabase(t *testing.T) {
	// Use the database from TestMain

	t.Run("Insert and Get User", func(t *testing.T) {
		// Insert user
		id, err := testDB.InsertUser("Alice", "alice@example.com")
		if err != nil {
			t.Fatalf("InsertUser failed: %v", err)
		}

		// Cleanup: delete user after test
		t.Cleanup(func() {
			testDB.DeleteUser(id)
		})

		// Get user
		user, err := testDB.GetUser(id)
		if err != nil {
			t.Fatalf("GetUser failed: %v", err)
		}

		if user.Name != "Alice" {
			t.Errorf("user name = %q; want %q", user.Name, "Alice")
		}
		if user.Email != "alice@example.com" {
			t.Errorf("user email = %q; want %q", user.Email, "alice@example.com")
		}
	})

	t.Run("Delete User", func(t *testing.T) {
		// Insert user
		id, err := testDB.InsertUser("Bob", "bob@example.com")
		if err != nil {
			t.Fatalf("InsertUser failed: %v", err)
		}

		// Delete user
		err = testDB.DeleteUser(id)
		if err != nil {
			t.Fatalf("DeleteUser failed: %v", err)
		}

		// Try to get deleted user
		_, err = testDB.GetUser(id)
		if err == nil {
			t.Error("expected error when getting deleted user")
		}
	})

	t.Run("Insert User with empty fields", func(t *testing.T) {
		_, err := testDB.InsertUser("", "")
		if err == nil {
			t.Error("expected error when inserting user with empty fields")
		}
	})

	t.Run("Get non-existent User", func(t *testing.T) {
		_, err := testDB.GetUser(99999)
		if err == nil {
			t.Error("expected error when getting non-existent user")
		}
	})

	t.Log("✓ All Database tests passed!")
}

func TestDatabaseLifecycle(t *testing.T) {
	// Test database connection lifecycle
	db, err := Connect("test://localhost/lifecycle")
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}

	// Use defer for cleanup
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("Close failed: %v", err)
		}
	}()

	// Insert user
	id, err := db.InsertUser("Charlie", "charlie@example.com")
	if err != nil {
		t.Fatalf("InsertUser failed: %v", err)
	}

	// Verify user exists
	user, err := db.GetUser(id)
	if err != nil {
		t.Fatalf("GetUser failed: %v", err)
	}

	if user.Name != "Charlie" {
		t.Errorf("user name = %q; want %q", user.Name, "Charlie")
	}

	t.Log("✓ Database lifecycle test passed!")
}

func TestMultipleStorageInstances(t *testing.T) {
	// Test that multiple storage instances are isolated

	dir1 := t.TempDir()
	dir2 := t.TempDir()

	storage1 := New(dir1)
	storage2 := New(dir2)

	// Save to storage1
	err := storage1.Save("file.txt", "content1")
	if err != nil {
		t.Fatalf("storage1.Save failed: %v", err)
	}

	// Save to storage2
	err = storage2.Save("file.txt", "content2")
	if err != nil {
		t.Fatalf("storage2.Save failed: %v", err)
	}

	// Load from storage1
	content1, err := storage1.Load("file.txt")
	if err != nil {
		t.Fatalf("storage1.Load failed: %v", err)
	}

	// Load from storage2
	content2, err := storage2.Load("file.txt")
	if err != nil {
		t.Fatalf("storage2.Load failed: %v", err)
	}

	// Verify isolation
	if content1 != "content1" {
		t.Errorf("storage1 content = %q; want %q", content1, "content1")
	}
	if content2 != "content2" {
		t.Errorf("storage2 content = %q; want %q", content2, "content2")
	}

	t.Log("✓ Multiple storage instances test passed!")
}
