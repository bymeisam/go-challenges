package main

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
)

type User struct {
	ID    int
	Name  string
	Email string
	Age   int
}

type Database struct {
	db *sql.DB
}

// NewDatabase creates a new database connection
func NewDatabase(dataSourceName string) (*Database, error) {
	db, err := sql.Open("sqlite3", dataSourceName)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &Database{db: db}, nil
}

// Close closes the database connection
func (d *Database) Close() error {
	return d.db.Close()
}

// CreateTable creates the users table
func (d *Database) CreateTable() error {
	query := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		email TEXT NOT NULL UNIQUE,
		age INTEGER NOT NULL
	)`

	_, err := d.db.Exec(query)
	return err
}

// InsertUser inserts a new user and returns the inserted user with ID
func (d *Database) InsertUser(name, email string, age int) (*User, error) {
	query := `INSERT INTO users (name, email, age) VALUES (?, ?, ?)`

	result, err := d.db.Exec(query, name, email, age)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return &User{
		ID:    int(id),
		Name:  name,
		Email: email,
		Age:   age,
	}, nil
}

// GetUser retrieves a user by ID
func (d *Database) GetUser(id int) (*User, error) {
	query := `SELECT id, name, email, age FROM users WHERE id = ?`

	user := &User{}
	err := d.db.QueryRow(query, id).Scan(&user.ID, &user.Name, &user.Email, &user.Age)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// GetAllUsers retrieves all users
func (d *Database) GetAllUsers() ([]User, error) {
	query := `SELECT id, name, email, age FROM users ORDER BY id`

	rows, err := d.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		if err := rows.Scan(&user.ID, &user.Name, &user.Email, &user.Age); err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

// UpdateUser updates an existing user
func (d *Database) UpdateUser(id int, name, email string, age int) error {
	query := `UPDATE users SET name = ?, email = ?, age = ? WHERE id = ?`

	result, err := d.db.Exec(query, name, email, age, id)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// DeleteUser deletes a user by ID
func (d *Database) DeleteUser(id int) error {
	query := `DELETE FROM users WHERE id = ?`

	result, err := d.db.Exec(query, id)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func main() {
	db, err := NewDatabase(":memory:")
	if err != nil {
		fmt.Printf("Error creating database: %v\n", err)
		return
	}
	defer db.Close()

	if err := db.CreateTable(); err != nil {
		fmt.Printf("Error creating table: %v\n", err)
		return
	}

	// Insert users
	user1, _ := db.InsertUser("Alice", "alice@example.com", 25)
	fmt.Printf("Inserted user: %+v\n", user1)

	// Get user
	user, _ := db.GetUser(user1.ID)
	fmt.Printf("Retrieved user: %+v\n", user)

	// Update user
	db.UpdateUser(user1.ID, "Alice Smith", "alice.smith@example.com", 26)

	// Get all users
	users, _ := db.GetAllUsers()
	fmt.Printf("All users: %+v\n", users)
}
