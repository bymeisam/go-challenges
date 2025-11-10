package main

import (
	"sync"
	"testing"
)

func setupTestDB(t *testing.T) *Database {
	db, err := NewDatabase(":memory:")
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	return db
}

func TestCreateAccount(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	id, err := db.CreateAccount("John Doe", 1000.0)
	if err != nil {
		t.Fatalf("Failed to create account: %v", err)
	}

	if id == 0 {
		t.Error("Expected ID to be set")
	}

	account, err := db.GetAccount(id)
	if err != nil {
		t.Fatalf("Failed to get account: %v", err)
	}

	if account.Name != "John Doe" {
		t.Errorf("Expected name 'John Doe', got %s", account.Name)
	}

	if account.Balance != 1000.0 {
		t.Errorf("Expected balance 1000.0, got %f", account.Balance)
	}
}

func TestCreateAccountNegativeBalance(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	_, err := db.CreateAccount("Invalid", -100)
	if err != ErrInvalidAmount {
		t.Errorf("Expected ErrInvalidAmount, got %v", err)
	}
}

func TestGetAccount(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	id, _ := db.CreateAccount("Test User", 500.0)

	account, err := db.GetAccount(id)
	if err != nil {
		t.Fatalf("Failed to get account: %v", err)
	}

	if account.ID != id {
		t.Errorf("Expected ID %d, got %d", id, account.ID)
	}
}

func TestGetAccountNotFound(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	_, err := db.GetAccount(999)
	if err != ErrAccountNotFound {
		t.Errorf("Expected ErrAccountNotFound, got %v", err)
	}
}

func TestTransferSuccess(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Create accounts
	id1, _ := db.CreateAccount("Alice", 1000.0)
	id2, _ := db.CreateAccount("Bob", 500.0)

	// Perform transfer
	err := db.Transfer(id1, id2, 200.0)
	if err != nil {
		t.Fatalf("Transfer failed: %v", err)
	}

	// Check balances
	alice, _ := db.GetAccount(id1)
	if alice.Balance != 800.0 {
		t.Errorf("Expected Alice balance 800.0, got %f", alice.Balance)
	}

	bob, _ := db.GetAccount(id2)
	if bob.Balance != 700.0 {
		t.Errorf("Expected Bob balance 700.0, got %f", bob.Balance)
	}
}

func TestTransferInsufficientBalance(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	id1, _ := db.CreateAccount("Alice", 100.0)
	id2, _ := db.CreateAccount("Bob", 500.0)

	// Try to transfer more than available
	err := db.Transfer(id1, id2, 200.0)
	if err != ErrInsufficientBalance {
		t.Errorf("Expected ErrInsufficientBalance, got %v", err)
	}

	// Balances should remain unchanged
	alice, _ := db.GetAccount(id1)
	if alice.Balance != 100.0 {
		t.Errorf("Expected Alice balance unchanged at 100.0, got %f", alice.Balance)
	}

	bob, _ := db.GetAccount(id2)
	if bob.Balance != 500.0 {
		t.Errorf("Expected Bob balance unchanged at 500.0, got %f", bob.Balance)
	}
}

func TestTransferInvalidAmount(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	id1, _ := db.CreateAccount("Alice", 1000.0)
	id2, _ := db.CreateAccount("Bob", 500.0)

	// Test negative amount
	err := db.Transfer(id1, id2, -100.0)
	if err != ErrInvalidAmount {
		t.Errorf("Expected ErrInvalidAmount for negative amount, got %v", err)
	}

	// Test zero amount
	err = db.Transfer(id1, id2, 0)
	if err != ErrInvalidAmount {
		t.Errorf("Expected ErrInvalidAmount for zero amount, got %v", err)
	}
}

func TestTransferAccountNotFound(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	id1, _ := db.CreateAccount("Alice", 1000.0)

	// Transfer from non-existent account
	err := db.Transfer(999, id1, 100.0)
	if err != ErrAccountNotFound {
		t.Errorf("Expected ErrAccountNotFound, got %v", err)
	}

	// Transfer to non-existent account
	err = db.Transfer(id1, 999, 100.0)
	if err != ErrAccountNotFound {
		t.Errorf("Expected ErrAccountNotFound, got %v", err)
	}
}

func TestTransferRollback(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	id1, _ := db.CreateAccount("Alice", 1000.0)
	id2, _ := db.CreateAccount("Bob", 500.0)

	initialAlice, _ := db.GetAccount(id1)
	initialBob, _ := db.GetAccount(id2)

	// Attempt transfer with insufficient funds
	db.Transfer(id1, id2, 2000.0)

	// Balances should be unchanged due to rollback
	alice, _ := db.GetAccount(id1)
	if alice.Balance != initialAlice.Balance {
		t.Error("Alice balance changed after failed transfer")
	}

	bob, _ := db.GetAccount(id2)
	if bob.Balance != initialBob.Balance {
		t.Error("Bob balance changed after failed transfer")
	}
}

func TestBatchCreateAccounts(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	accounts := []Account{
		{Name: "Account1", Balance: 100},
		{Name: "Account2", Balance: 200},
		{Name: "Account3", Balance: 300},
	}

	err := db.BatchCreateAccounts(accounts)
	if err != nil {
		t.Fatalf("Failed to batch create accounts: %v", err)
	}

	// Verify accounts were created
	account1, _ := db.GetAccount(1)
	if account1.Name != "Account1" || account1.Balance != 100 {
		t.Error("Account1 not created correctly")
	}

	account2, _ := db.GetAccount(2)
	if account2.Name != "Account2" || account2.Balance != 200 {
		t.Error("Account2 not created correctly")
	}
}

func TestBatchCreateAccountsRollback(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	accounts := []Account{
		{Name: "Account1", Balance: 100},
		{Name: "Account2", Balance: -200}, // Invalid
		{Name: "Account3", Balance: 300},
	}

	err := db.BatchCreateAccounts(accounts)
	if err != ErrInvalidAmount {
		t.Errorf("Expected ErrInvalidAmount, got %v", err)
	}

	// No accounts should be created due to rollback
	_, err = db.GetAccount(1)
	if err != ErrAccountNotFound {
		t.Error("Expected no accounts to be created")
	}
}

func TestGetTransactionHistory(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	id1, _ := db.CreateAccount("Alice", 1000.0)
	id2, _ := db.CreateAccount("Bob", 500.0)
	id3, _ := db.CreateAccount("Charlie", 300.0)

	// Perform multiple transfers
	db.Transfer(id1, id2, 100.0)
	db.Transfer(id1, id3, 50.0)
	db.Transfer(id2, id1, 75.0)

	// Get Alice's transaction history
	transactions, err := db.GetTransactionHistory(id1)
	if err != nil {
		t.Fatalf("Failed to get transaction history: %v", err)
	}

	if len(transactions) != 3 {
		t.Errorf("Expected 3 transactions, got %d", len(transactions))
	}

	// Verify first transaction
	if transactions[0].FromID != id1 || transactions[0].ToID != id2 || transactions[0].Amount != 100.0 {
		t.Error("First transaction details incorrect")
	}
}

func TestMultiTransferSuccess(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	id1, _ := db.CreateAccount("Alice", 1000.0)
	id2, _ := db.CreateAccount("Bob", 500.0)
	id3, _ := db.CreateAccount("Charlie", 300.0)

	transfers := []struct {
		FromID int
		ToID   int
		Amount float64
	}{
		{FromID: id1, ToID: id2, Amount: 100.0},
		{FromID: id1, ToID: id3, Amount: 50.0},
	}

	err := db.MultiTransfer(transfers)
	if err != nil {
		t.Fatalf("Multi-transfer failed: %v", err)
	}

	alice, _ := db.GetAccount(id1)
	if alice.Balance != 850.0 {
		t.Errorf("Expected Alice balance 850.0, got %f", alice.Balance)
	}

	bob, _ := db.GetAccount(id2)
	if bob.Balance != 600.0 {
		t.Errorf("Expected Bob balance 600.0, got %f", bob.Balance)
	}

	charlie, _ := db.GetAccount(id3)
	if charlie.Balance != 350.0 {
		t.Errorf("Expected Charlie balance 350.0, got %f", charlie.Balance)
	}
}

func TestMultiTransferRollback(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	id1, _ := db.CreateAccount("Alice", 1000.0)
	id2, _ := db.CreateAccount("Bob", 500.0)
	id3, _ := db.CreateAccount("Charlie", 300.0)

	transfers := []struct {
		FromID int
		ToID   int
		Amount float64
	}{
		{FromID: id1, ToID: id2, Amount: 100.0},
		{FromID: id1, ToID: id3, Amount: 2000.0}, // This will fail
	}

	err := db.MultiTransfer(transfers)
	if err != ErrInsufficientBalance {
		t.Errorf("Expected ErrInsufficientBalance, got %v", err)
	}

	// All balances should remain unchanged
	alice, _ := db.GetAccount(id1)
	if alice.Balance != 1000.0 {
		t.Errorf("Expected Alice balance unchanged at 1000.0, got %f", alice.Balance)
	}

	bob, _ := db.GetAccount(id2)
	if bob.Balance != 500.0 {
		t.Errorf("Expected Bob balance unchanged at 500.0, got %f", bob.Balance)
	}
}

func TestConcurrentTransfers(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	id1, _ := db.CreateAccount("Alice", 1000.0)
	id2, _ := db.CreateAccount("Bob", 1000.0)

	var wg sync.WaitGroup
	errors := make(chan error, 10)

	// Perform 10 concurrent transfers
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := db.Transfer(id1, id2, 50.0)
			if err != nil {
				errors <- err
			}
		}()
	}

	wg.Wait()
	close(errors)

	// Check for errors
	for err := range errors {
		t.Errorf("Concurrent transfer error: %v", err)
	}

	// Verify final balances
	alice, _ := db.GetAccount(id1)
	bob, _ := db.GetAccount(id2)

	if alice.Balance != 500.0 {
		t.Errorf("Expected Alice balance 500.0, got %f", alice.Balance)
	}

	if bob.Balance != 1500.0 {
		t.Errorf("Expected Bob balance 1500.0, got %f", bob.Balance)
	}
}
