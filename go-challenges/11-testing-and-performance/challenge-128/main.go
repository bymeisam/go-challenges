package main

import (
	"errors"
	"os"
	"path/filepath"
)

// FileStorage provides file-based storage operations
type FileStorage struct {
	baseDir string
}

// New creates a new FileStorage instance
func New(baseDir string) *FileStorage {
	return &FileStorage{baseDir: baseDir}
}

// Save saves content to a file
func (fs *FileStorage) Save(filename, content string) error {
	// Ensure base directory exists
	if err := os.MkdirAll(fs.baseDir, 0755); err != nil {
		return err
	}

	path := filepath.Join(fs.baseDir, filename)
	return os.WriteFile(path, []byte(content), 0644)
}

// Load loads content from a file
func (fs *FileStorage) Load(filename string) (string, error) {
	path := filepath.Join(fs.baseDir, filename)
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// Delete deletes a file
func (fs *FileStorage) Delete(filename string) error {
	path := filepath.Join(fs.baseDir, filename)
	return os.Remove(path)
}

// List lists all files in the storage directory
func (fs *FileStorage) List() ([]string, error) {
	entries, err := os.ReadDir(fs.baseDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}

	var files []string
	for _, entry := range entries {
		if !entry.IsDir() {
			files = append(files, entry.Name())
		}
	}
	return files, nil
}

// User represents a user in the database
type User struct {
	ID    int
	Name  string
	Email string
}

// Database is a mock database for testing
type Database struct {
	dsn       string
	users     map[int]User
	nextID    int
	connected bool
}

// Connect creates a new database connection
func Connect(dsn string) (*Database, error) {
	if dsn == "" {
		return nil, errors.New("DSN cannot be empty")
	}

	return &Database{
		dsn:       dsn,
		users:     make(map[int]User),
		nextID:    1,
		connected: true,
	}, nil
}

// Close closes the database connection
func (db *Database) Close() error {
	if !db.connected {
		return errors.New("database not connected")
	}
	db.connected = false
	return nil
}

// InsertUser inserts a new user and returns the ID
func (db *Database) InsertUser(name, email string) (int, error) {
	if !db.connected {
		return 0, errors.New("database not connected")
	}

	if name == "" || email == "" {
		return 0, errors.New("name and email cannot be empty")
	}

	id := db.nextID
	db.users[id] = User{
		ID:    id,
		Name:  name,
		Email: email,
	}
	db.nextID++

	return id, nil
}

// GetUser retrieves a user by ID
func (db *Database) GetUser(id int) (User, error) {
	if !db.connected {
		return User{}, errors.New("database not connected")
	}

	user, exists := db.users[id]
	if !exists {
		return User{}, errors.New("user not found")
	}

	return user, nil
}

// DeleteUser deletes a user by ID
func (db *Database) DeleteUser(id int) error {
	if !db.connected {
		return errors.New("database not connected")
	}

	if _, exists := db.users[id]; !exists {
		return errors.New("user not found")
	}

	delete(db.users, id)
	return nil
}

func main() {
	// Example usage
}
