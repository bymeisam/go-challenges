package main

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Product struct {
	ID          int
	Name        string
	Description string
	Price       float64
	Stock       int
}

type Database struct {
	db               *sql.DB
	insertStmt       *sql.Stmt
	getStmt          *sql.Stmt
	updateStockStmt  *sql.Stmt
	deleteStmt       *sql.Stmt
	searchByNameStmt *sql.Stmt
}

// NewDatabase creates a new database with prepared statements
func NewDatabase(dataSourceName string) (*Database, error) {
	db, err := sql.Open("sqlite3", dataSourceName)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	database := &Database{db: db}

	// Create table
	if err := database.createTable(); err != nil {
		db.Close()
		return nil, err
	}

	// Prepare statements
	if err := database.prepareStatements(); err != nil {
		db.Close()
		return nil, err
	}

	return database, nil
}

func (d *Database) createTable() error {
	query := `
	CREATE TABLE IF NOT EXISTS products (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		description TEXT,
		price REAL NOT NULL,
		stock INTEGER NOT NULL DEFAULT 0
	)`

	_, err := d.db.Exec(query)
	return err
}

// prepareStatements prepares all commonly used statements
func (d *Database) prepareStatements() error {
	var err error

	// Prepare insert statement
	d.insertStmt, err = d.db.Prepare(`
		INSERT INTO products (name, description, price, stock)
		VALUES (?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare insert statement: %w", err)
	}

	// Prepare get statement
	d.getStmt, err = d.db.Prepare(`
		SELECT id, name, description, price, stock
		FROM products WHERE id = ?
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare get statement: %w", err)
	}

	// Prepare update stock statement
	d.updateStockStmt, err = d.db.Prepare(`
		UPDATE products SET stock = stock + ? WHERE id = ?
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare update stock statement: %w", err)
	}

	// Prepare delete statement
	d.deleteStmt, err = d.db.Prepare(`
		DELETE FROM products WHERE id = ?
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare delete statement: %w", err)
	}

	// Prepare search by name statement
	d.searchByNameStmt, err = d.db.Prepare(`
		SELECT id, name, description, price, stock
		FROM products WHERE name LIKE ? ORDER BY name
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare search statement: %w", err)
	}

	return nil
}

// Close closes all prepared statements and the database connection
func (d *Database) Close() error {
	// Close all prepared statements
	if d.insertStmt != nil {
		d.insertStmt.Close()
	}
	if d.getStmt != nil {
		d.getStmt.Close()
	}
	if d.updateStockStmt != nil {
		d.updateStockStmt.Close()
	}
	if d.deleteStmt != nil {
		d.deleteStmt.Close()
	}
	if d.searchByNameStmt != nil {
		d.searchByNameStmt.Close()
	}

	return d.db.Close()
}

// InsertProduct inserts a product using prepared statement
func (d *Database) InsertProduct(product *Product) (int, error) {
	result, err := d.insertStmt.Exec(
		product.Name,
		product.Description,
		product.Price,
		product.Stock,
	)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(id), nil
}

// GetProduct retrieves a product using prepared statement
func (d *Database) GetProduct(id int) (*Product, error) {
	product := &Product{}
	err := d.getStmt.QueryRow(id).Scan(
		&product.ID,
		&product.Name,
		&product.Description,
		&product.Price,
		&product.Stock,
	)
	if err != nil {
		return nil, err
	}

	return product, nil
}

// UpdateStock updates product stock using prepared statement
func (d *Database) UpdateStock(id int, delta int) error {
	result, err := d.updateStockStmt.Exec(delta, id)
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

// DeleteProduct deletes a product using prepared statement
func (d *Database) DeleteProduct(id int) error {
	result, err := d.deleteStmt.Exec(id)
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

// SearchByName searches products by name pattern using prepared statement
func (d *Database) SearchByName(pattern string) ([]Product, error) {
	rows, err := d.searchByNameStmt.Query("%" + pattern + "%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []Product
	for rows.Next() {
		var product Product
		if err := rows.Scan(
			&product.ID,
			&product.Name,
			&product.Description,
			&product.Price,
			&product.Stock,
		); err != nil {
			return nil, err
		}
		products = append(products, product)
	}

	return products, rows.Err()
}

// BenchmarkInserts performs multiple inserts for performance comparison
func (d *Database) BenchmarkInserts(count int, usePrepared bool) time.Duration {
	start := time.Now()

	if usePrepared {
		for i := 0; i < count; i++ {
			d.insertStmt.Exec(
				fmt.Sprintf("Product %d", i),
				"Description",
				9.99,
				10,
			)
		}
	} else {
		for i := 0; i < count; i++ {
			d.db.Exec(
				"INSERT INTO products (name, description, price, stock) VALUES (?, ?, ?, ?)",
				fmt.Sprintf("Product %d", i),
				"Description",
				9.99,
				10,
			)
		}
	}

	return time.Since(start)
}

func main() {
	db, err := NewDatabase(":memory:")
	if err != nil {
		fmt.Printf("Error creating database: %v\n", err)
		return
	}
	defer db.Close()

	// Insert products
	product := &Product{
		Name:        "Laptop",
		Description: "High-performance laptop",
		Price:       999.99,
		Stock:       10,
	}

	id, _ := db.InsertProduct(product)
	fmt.Printf("Inserted product with ID: %d\n", id)

	// Get product
	retrieved, _ := db.GetProduct(id)
	fmt.Printf("Retrieved product: %+v\n", retrieved)

	// Update stock
	db.UpdateStock(id, 5)
	fmt.Println("Stock updated")

	// Search by name
	products, _ := db.SearchByName("Laptop")
	fmt.Printf("Found products: %+v\n", products)

	// Performance comparison
	preparedTime := db.BenchmarkInserts(100, true)
	fmt.Printf("Prepared statements: %v\n", preparedTime)
}
