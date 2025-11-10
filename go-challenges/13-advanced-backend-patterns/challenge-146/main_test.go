package main

import (
	"testing"
	"time"
)

func TestEventStore_AppendEvent(t *testing.T) {
	store := NewEventStore()

	event := Event{
		AggregateID:   "agg-1",
		AggregateType: "TestAggregate",
		EventType:     "TestEvent",
		Data:          map[string]interface{}{"key": "value"},
	}

	err := store.AppendEvent(event)
	if err != nil {
		t.Fatalf("AppendEvent failed: %v", err)
	}

	events := store.GetEvents("agg-1")
	if len(events) != 1 {
		t.Errorf("Expected 1 event, got %d", len(events))
	}

	if events[0].EventType != "TestEvent" {
		t.Errorf("EventType = %s, want TestEvent", events[0].EventType)
	}
}

func TestEventStore_GetEvents(t *testing.T) {
	store := NewEventStore()

	for i := 0; i < 5; i++ {
		store.AppendEvent(Event{
			AggregateID:   "agg-1",
			AggregateType: "TestAggregate",
			EventType:     "TestEvent",
			Data:          map[string]interface{}{"index": float64(i)},
		})
	}

	events := store.GetEvents("agg-1")
	if len(events) != 5 {
		t.Errorf("Expected 5 events, got %d", len(events))
	}
}

func TestEventStore_GetEventsSince(t *testing.T) {
	store := NewEventStore()

	for i := 0; i < 5; i++ {
		store.AppendEvent(Event{
			AggregateID:   "agg-1",
			AggregateType: "TestAggregate",
			EventType:     "TestEvent",
			Data:          map[string]interface{}{"index": float64(i)},
		})
	}

	events := store.GetEventsSince("agg-1", 2)
	if len(events) != 3 {
		t.Errorf("Expected 3 events after version 2, got %d", len(events))
	}
}

func TestEventStore_Snapshot(t *testing.T) {
	store := NewEventStore()

	snapshot := Snapshot{
		AggregateID:   "agg-1",
		AggregateType: "TestAggregate",
		Version:       10,
		State:         map[string]interface{}{"balance": float64(1000)},
	}

	err := store.SaveSnapshot(snapshot)
	if err != nil {
		t.Fatalf("SaveSnapshot failed: %v", err)
	}

	loaded, exists := store.GetSnapshot("agg-1")
	if !exists {
		t.Fatal("Snapshot should exist")
	}

	if loaded.Version != 10 {
		t.Errorf("Version = %d, want 10", loaded.Version)
	}
}

func TestBankAccount_CreateAccount(t *testing.T) {
	account := NewBankAccount("acc-1")

	err := account.CreateAccount(1000)
	if err != nil {
		t.Fatalf("CreateAccount failed: %v", err)
	}

	if account.Balance != 1000 {
		t.Errorf("Balance = %d, want 1000", account.Balance)
	}

	if account.Status != "active" {
		t.Errorf("Status = %s, want active", account.Status)
	}

	changes := account.GetUncommittedChanges()
	if len(changes) != 1 {
		t.Errorf("Expected 1 uncommitted change, got %d", len(changes))
	}
}

func TestBankAccount_Deposit(t *testing.T) {
	account := NewBankAccount("acc-1")
	account.CreateAccount(1000)

	err := account.Deposit(500)
	if err != nil {
		t.Fatalf("Deposit failed: %v", err)
	}

	if account.Balance != 1500 {
		t.Errorf("Balance = %d, want 1500", account.Balance)
	}
}

func TestBankAccount_Withdraw(t *testing.T) {
	account := NewBankAccount("acc-1")
	account.CreateAccount(1000)

	err := account.Withdraw(300)
	if err != nil {
		t.Fatalf("Withdraw failed: %v", err)
	}

	if account.Balance != 700 {
		t.Errorf("Balance = %d, want 700", account.Balance)
	}
}

func TestBankAccount_InsufficientFunds(t *testing.T) {
	account := NewBankAccount("acc-1")
	account.CreateAccount(100)

	err := account.Withdraw(200)
	if err == nil {
		t.Error("Withdraw should fail with insufficient funds")
	}
}

func TestBankAccount_CloseAccount(t *testing.T) {
	account := NewBankAccount("acc-1")
	account.CreateAccount(1000)

	err := account.Close()
	if err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	if account.Status != "closed" {
		t.Errorf("Status = %s, want closed", account.Status)
	}

	// Operations on closed account should fail
	err = account.Deposit(100)
	if err == nil {
		t.Error("Deposit should fail on closed account")
	}
}

func TestAccountRepository_SaveAndLoad(t *testing.T) {
	store := NewEventStore()
	repo := NewAccountRepository(store)

	// Create and save account
	account := NewBankAccount("acc-1")
	account.CreateAccount(1000)
	account.Deposit(500)
	account.Withdraw(200)

	err := repo.Save(account)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Load account
	loaded, err := repo.GetByID("acc-1")
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}

	if loaded.Balance != 1300 {
		t.Errorf("Balance = %d, want 1300", loaded.Balance)
	}

	if loaded.Version != account.Version {
		t.Errorf("Version = %d, want %d", loaded.Version, account.Version)
	}
}

func TestAccountRepository_Snapshot(t *testing.T) {
	store := NewEventStore()
	store.snapshotInterval = 3 // Set small interval for testing
	repo := NewAccountRepository(store)

	account := NewBankAccount("acc-1")
	account.CreateAccount(1000)

	// Generate enough events to trigger snapshot
	for i := 0; i < 5; i++ {
		account.Deposit(100)
	}

	repo.Save(account)

	// Check if snapshot was created
	_, exists := store.GetSnapshot("acc-1")
	if !exists {
		t.Error("Snapshot should have been created")
	}

	// Verify can rebuild from snapshot
	loaded, _ := repo.GetByID("acc-1")
	if loaded.Balance != account.Balance {
		t.Errorf("Balance = %d, want %d", loaded.Balance, account.Balance)
	}
}

func TestBalanceProjection(t *testing.T) {
	projection := NewBalanceProjection()

	// Project events
	events := []Event{
		{
			AggregateID: "acc-1",
			EventType:   "AccountCreated",
			Data:        map[string]interface{}{"initial_balance": float64(1000)},
			Timestamp:   time.Now(),
		},
		{
			AggregateID: "acc-1",
			EventType:   "MoneyDeposited",
			Data:        map[string]interface{}{"amount": float64(500)},
			Timestamp:   time.Now(),
		},
		{
			AggregateID: "acc-1",
			EventType:   "MoneyWithdrawn",
			Data:        map[string]interface{}{"amount": float64(200)},
			Timestamp:   time.Now(),
		},
	}

	for _, event := range events {
		projection.Project(event)
	}

	balance, err := projection.GetBalance("acc-1")
	if err != nil {
		t.Fatalf("GetBalance failed: %v", err)
	}

	if balance.Balance != 1300 {
		t.Errorf("Balance = %d, want 1300", balance.Balance)
	}

	if balance.LastEvent != "MoneyWithdrawn" {
		t.Errorf("LastEvent = %s, want MoneyWithdrawn", balance.LastEvent)
	}
}

func TestEventVersioning(t *testing.T) {
	store := NewEventStore()

	event1 := Event{
		AggregateID: "agg-1",
		EventType:   "Event1",
		Data:        map[string]interface{}{},
	}

	event2 := Event{
		AggregateID: "agg-1",
		EventType:   "Event2",
		Data:        map[string]interface{}{},
	}

	store.AppendEvent(event1)
	store.AppendEvent(event2)

	events := store.GetEvents("agg-1")
	if events[0].Version >= events[1].Version {
		t.Error("Events should have increasing versions")
	}
}

func TestTimeTravel(t *testing.T) {
	account := NewBankAccount("acc-1")
	account.CreateAccount(1000)
	account.Deposit(500)
	account.Withdraw(300)

	changes := account.GetUncommittedChanges()

	// Replay only first 2 events
	timeTravelAccount := NewBankAccount("acc-1")
	for i := 0; i < 2; i++ {
		timeTravelAccount.Apply(changes[i])
	}

	// Should have balance after creation and first deposit
	if timeTravelAccount.Balance != 1500 {
		t.Errorf("Time travel balance = %d, want 1500", timeTravelAccount.Balance)
	}
}

func TestConcurrentEventAppend(t *testing.T) {
	store := NewEventStore()

	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(id int) {
			for j := 0; j < 10; j++ {
				store.AppendEvent(Event{
					AggregateID: "agg-1",
					EventType:   "TestEvent",
					Data:        map[string]interface{}{"id": float64(id)},
				})
			}
			done <- true
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	events := store.GetEvents("agg-1")
	if len(events) != 100 {
		t.Errorf("Expected 100 events, got %d", len(events))
	}
}

func BenchmarkEventAppend(b *testing.B) {
	store := NewEventStore()

	event := Event{
		AggregateID: "agg-1",
		EventType:   "BenchEvent",
		Data:        map[string]interface{}{"value": float64(100)},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		store.AppendEvent(event)
	}
}

func BenchmarkEventReplay(b *testing.B) {
	store := NewEventStore()

	// Create 1000 events
	for i := 0; i < 1000; i++ {
		store.AppendEvent(Event{
			AggregateID: "agg-1",
			EventType:   "MoneyDeposited",
			Data:        map[string]interface{}{"amount": float64(100)},
		})
	}

	events := store.GetEvents("agg-1")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		account := NewBankAccount("agg-1")
		for _, event := range events {
			account.Apply(event)
		}
	}
}

func BenchmarkProjection(b *testing.B) {
	projection := NewBalanceProjection()

	event := Event{
		AggregateID: "agg-1",
		EventType:   "MoneyDeposited",
		Data:        map[string]interface{}{"amount": float64(100)},
		Timestamp:   time.Now(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		projection.Project(event)
	}
}
