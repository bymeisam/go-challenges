package main

import (
	"database/sql"
	"errors"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

type Account struct {
	ID      int
	Name    string
	Balance float64
}

type Transaction struct {
	ID     int
	FromID int
	ToID   int
	Amount float64
}

type Database struct {
	db *sql.DB
}

var (
	ErrInsufficientBalance = errors.New("insufficient balance")
	ErrInvalidAmount       = errors.New("invalid amount")
	ErrAccountNotFound     = errors.New("account not found")
)

func NewDatabase(dataSourceName string) (*Database, error) {
	db, err := sql.Open("sqlite3", dataSourceName)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	database := &Database{db: db}

	if err := database.createTables(); err != nil {
		db.Close()
		return nil, err
	}

	return database, nil
}

func (d *Database) createTables() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS accounts (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			balance REAL NOT NULL DEFAULT 0 CHECK(balance >= 0)
		)`,
		`CREATE TABLE IF NOT EXISTS transactions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			from_id INTEGER NOT NULL,
			to_id INTEGER NOT NULL,
			amount REAL NOT NULL,
			FOREIGN KEY (from_id) REFERENCES accounts(id),
			FOREIGN KEY (to_id) REFERENCES accounts(id)
		)`,
	}

	for _, query := range queries {
		if _, err := d.db.Exec(query); err != nil {
			return err
		}
	}

	return nil
}

func (d *Database) Close() error {
	return d.db.Close()
}

// CreateAccount creates a new account with initial balance
func (d *Database) CreateAccount(name string, initialBalance float64) (int, error) {
	if initialBalance < 0 {
		return 0, ErrInvalidAmount
	}

	result, err := d.db.Exec(
		"INSERT INTO accounts (name, balance) VALUES (?, ?)",
		name, initialBalance,
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

// GetAccount retrieves an account by ID
func (d *Database) GetAccount(id int) (*Account, error) {
	account := &Account{}
	err := d.db.QueryRow(
		"SELECT id, name, balance FROM accounts WHERE id = ?",
		id,
	).Scan(&account.ID, &account.Name, &account.Balance)

	if err == sql.ErrNoRows {
		return nil, ErrAccountNotFound
	}
	if err != nil {
		return nil, err
	}

	return account, nil
}

// Transfer performs a money transfer between two accounts using a transaction
func (d *Database) Transfer(fromID, toID int, amount float64) error {
	if amount <= 0 {
		return ErrInvalidAmount
	}

	// Begin transaction
	tx, err := d.db.Begin()
	if err != nil {
		return err
	}

	// Defer rollback in case of error
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// Check if source account exists and has sufficient balance
	var fromBalance float64
	err = tx.QueryRow(
		"SELECT balance FROM accounts WHERE id = ?",
		fromID,
	).Scan(&fromBalance)
	if err == sql.ErrNoRows {
		return ErrAccountNotFound
	}
	if err != nil {
		return err
	}

	if fromBalance < amount {
		return ErrInsufficientBalance
	}

	// Check if destination account exists
	var toBalance float64
	err = tx.QueryRow(
		"SELECT balance FROM accounts WHERE id = ?",
		toID,
	).Scan(&toBalance)
	if err == sql.ErrNoRows {
		return ErrAccountNotFound
	}
	if err != nil {
		return err
	}

	// Deduct from source account
	_, err = tx.Exec(
		"UPDATE accounts SET balance = balance - ? WHERE id = ?",
		amount, fromID,
	)
	if err != nil {
		return err
	}

	// Add to destination account
	_, err = tx.Exec(
		"UPDATE accounts SET balance = balance + ? WHERE id = ?",
		amount, toID,
	)
	if err != nil {
		return err
	}

	// Record transaction
	_, err = tx.Exec(
		"INSERT INTO transactions (from_id, to_id, amount) VALUES (?, ?, ?)",
		fromID, toID, amount,
	)
	if err != nil {
		return err
	}

	// Commit transaction
	return tx.Commit()
}

// BatchCreateAccounts creates multiple accounts in a single transaction
func (d *Database) BatchCreateAccounts(accounts []Account) error {
	tx, err := d.db.Begin()
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	stmt, err := tx.Prepare("INSERT INTO accounts (name, balance) VALUES (?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, account := range accounts {
		if account.Balance < 0 {
			return ErrInvalidAmount
		}

		_, err = stmt.Exec(account.Name, account.Balance)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// GetTransactionHistory retrieves all transactions for an account
func (d *Database) GetTransactionHistory(accountID int) ([]Transaction, error) {
	query := `
		SELECT id, from_id, to_id, amount
		FROM transactions
		WHERE from_id = ? OR to_id = ?
		ORDER BY id
	`

	rows, err := d.db.Query(query, accountID, accountID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []Transaction
	for rows.Next() {
		var t Transaction
		if err := rows.Scan(&t.ID, &t.FromID, &t.ToID, &t.Amount); err != nil {
			return nil, err
		}
		transactions = append(transactions, t)
	}

	return transactions, rows.Err()
}

// MultiTransfer performs multiple transfers atomically
func (d *Database) MultiTransfer(transfers []struct {
	FromID int
	ToID   int
	Amount float64
}) error {
	tx, err := d.db.Begin()
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	for _, transfer := range transfers {
		if transfer.Amount <= 0 {
			return ErrInvalidAmount
		}

		// Check source balance
		var balance float64
		err = tx.QueryRow(
			"SELECT balance FROM accounts WHERE id = ?",
			transfer.FromID,
		).Scan(&balance)
		if err == sql.ErrNoRows {
			return ErrAccountNotFound
		}
		if err != nil {
			return err
		}

		if balance < transfer.Amount {
			return ErrInsufficientBalance
		}

		// Deduct from source
		_, err = tx.Exec(
			"UPDATE accounts SET balance = balance - ? WHERE id = ?",
			transfer.Amount, transfer.FromID,
		)
		if err != nil {
			return err
		}

		// Add to destination
		_, err = tx.Exec(
			"UPDATE accounts SET balance = balance + ? WHERE id = ?",
			transfer.Amount, transfer.ToID,
		)
		if err != nil {
			return err
		}

		// Record transaction
		_, err = tx.Exec(
			"INSERT INTO transactions (from_id, to_id, amount) VALUES (?, ?, ?)",
			transfer.FromID, transfer.ToID, transfer.Amount,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func main() {
	db, err := NewDatabase(":memory:")
	if err != nil {
		fmt.Printf("Error creating database: %v\n", err)
		return
	}
	defer db.Close()

	// Create accounts
	id1, _ := db.CreateAccount("Alice", 1000)
	id2, _ := db.CreateAccount("Bob", 500)

	fmt.Printf("Created accounts: Alice (ID: %d), Bob (ID: %d)\n", id1, id2)

	// Perform transfer
	err = db.Transfer(id1, id2, 200)
	if err != nil {
		fmt.Printf("Transfer failed: %v\n", err)
	} else {
		fmt.Println("Transfer successful")
	}

	// Check balances
	alice, _ := db.GetAccount(id1)
	bob, _ := db.GetAccount(id2)
	fmt.Printf("Alice balance: %.2f\n", alice.Balance)
	fmt.Printf("Bob balance: %.2f\n", bob.Balance)
}
