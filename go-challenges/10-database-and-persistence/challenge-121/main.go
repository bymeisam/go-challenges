package main

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type PoolConfig struct {
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

type Database struct {
	db     *sql.DB
	config PoolConfig
}

type PoolStats struct {
	MaxOpenConnections int
	OpenConnections    int
	InUse              int
	Idle               int
	WaitCount          int64
	WaitDuration       time.Duration
	MaxIdleClosed      int64
	MaxLifetimeClosed  int64
}

// NewDatabase creates a database with custom connection pool settings
func NewDatabase(dataSourceName string, config PoolConfig) (*Database, error) {
	db, err := sql.Open("sqlite3", dataSourceName)
	if err != nil {
		return nil, err
	}

	// Set connection pool parameters
	db.SetMaxOpenConns(config.MaxOpenConns)
	db.SetMaxIdleConns(config.MaxIdleConns)
	db.SetConnMaxLifetime(config.ConnMaxLifetime)
	db.SetConnMaxIdleTime(config.ConnMaxIdleTime)

	if err := db.Ping(); err != nil {
		return nil, err
	}

	database := &Database{
		db:     db,
		config: config,
	}

	if err := database.createTable(); err != nil {
		db.Close()
		return nil, err
	}

	return database, nil
}

func (d *Database) createTable() error {
	query := `
	CREATE TABLE IF NOT EXISTS items (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		value INTEGER NOT NULL
	)`

	_, err := d.db.Exec(query)
	return err
}

// Close closes the database connection
func (d *Database) Close() error {
	return d.db.Close()
}

// GetStats returns current connection pool statistics
func (d *Database) GetStats() PoolStats {
	stats := d.db.Stats()

	return PoolStats{
		MaxOpenConnections: stats.MaxOpenConnections,
		OpenConnections:    stats.OpenConnections,
		InUse:              stats.InUse,
		Idle:               stats.Idle,
		WaitCount:          stats.WaitCount,
		WaitDuration:       stats.WaitDuration,
		MaxIdleClosed:      stats.MaxIdleClosed,
		MaxLifetimeClosed:  stats.MaxLifetimeClosed,
	}
}

// GetConfig returns the current pool configuration
func (d *Database) GetConfig() PoolConfig {
	return d.config
}

// Insert inserts an item into the database
func (d *Database) Insert(name string, value int) (int, error) {
	result, err := d.db.Exec(
		"INSERT INTO items (name, value) VALUES (?, ?)",
		name, value,
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

// InsertWithContext inserts with a context (useful for timeouts)
func (d *Database) InsertWithContext(ctx context.Context, name string, value int) (int, error) {
	result, err := d.db.ExecContext(
		ctx,
		"INSERT INTO items (name, value) VALUES (?, ?)",
		name, value,
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

// Query performs a query with potential delay to test connection pooling
func (d *Database) Query(id int) (string, int, error) {
	var name string
	var value int

	err := d.db.QueryRow(
		"SELECT name, value FROM items WHERE id = ?",
		id,
	).Scan(&name, &value)

	if err != nil {
		return "", 0, err
	}

	return name, value, nil
}

// QueryWithContext performs a query with context
func (d *Database) QueryWithContext(ctx context.Context, id int) (string, int, error) {
	var name string
	var value int

	err := d.db.QueryRowContext(
		ctx,
		"SELECT name, value FROM items WHERE id = ?",
		id,
	).Scan(&name, &value)

	if err != nil {
		return "", 0, err
	}

	return name, value, nil
}

// PingWithContext checks database connectivity with context
func (d *Database) PingWithContext(ctx context.Context) error {
	return d.db.PingContext(ctx)
}

// GetConnectionWithTimeout acquires a connection with timeout
func (d *Database) GetConnectionWithTimeout(timeout time.Duration) (*sql.Conn, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	return d.db.Conn(ctx)
}

// ExecuteWithConnection executes a query using a specific connection
func (d *Database) ExecuteWithConnection(conn *sql.Conn, name string, value int) error {
	_, err := conn.ExecContext(
		context.Background(),
		"INSERT INTO items (name, value) VALUES (?, ?)",
		name, value,
	)
	return err
}

// SimulateLoad simulates database load for testing pool behavior
func (d *Database) SimulateLoad(numRequests int, delay time.Duration) []error {
	errors := make([]error, 0)
	results := make(chan error, numRequests)

	for i := 0; i < numRequests; i++ {
		go func(index int) {
			_, err := d.Insert(fmt.Sprintf("item-%d", index), index)
			if delay > 0 {
				time.Sleep(delay)
			}
			results <- err
		}(i)
	}

	for i := 0; i < numRequests; i++ {
		if err := <-results; err != nil {
			errors = append(errors, err)
		}
	}

	return errors
}

func main() {
	config := PoolConfig{
		MaxOpenConns:    10,
		MaxIdleConns:    5,
		ConnMaxLifetime: time.Hour,
		ConnMaxIdleTime: 10 * time.Minute,
	}

	db, err := NewDatabase(":memory:", config)
	if err != nil {
		fmt.Printf("Error creating database: %v\n", err)
		return
	}
	defer db.Close()

	// Insert some data
	id, _ := db.Insert("test-item", 42)
	fmt.Printf("Inserted item with ID: %d\n", id)

	// Query data
	name, value, _ := db.Query(id)
	fmt.Printf("Retrieved: %s = %d\n", name, value)

	// Get pool stats
	stats := db.GetStats()
	fmt.Printf("Pool Stats: Open=%d, InUse=%d, Idle=%d\n",
		stats.OpenConnections, stats.InUse, stats.Idle)

	// Simulate load
	fmt.Println("Simulating load...")
	errors := db.SimulateLoad(20, 10*time.Millisecond)
	fmt.Printf("Completed with %d errors\n", len(errors))

	// Final stats
	finalStats := db.GetStats()
	fmt.Printf("Final Stats: Open=%d, InUse=%d, Idle=%d, WaitCount=%d\n",
		finalStats.OpenConnections, finalStats.InUse, finalStats.Idle, finalStats.WaitCount)
}
