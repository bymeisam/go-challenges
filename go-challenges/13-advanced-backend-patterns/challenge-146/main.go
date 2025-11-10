package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"
)

// Event represents a domain event
type Event struct {
	ID            string
	AggregateID   string
	AggregateType string
	EventType     string
	Data          interface{}
	Version       int
	Timestamp     time.Time
	Metadata      map[string]string
}

// Snapshot represents an aggregate snapshot
type Snapshot struct {
	AggregateID   string
	AggregateType string
	Version       int
	State         interface{}
	Timestamp     time.Time
}

// EventStore stores events in append-only fashion
type EventStore struct {
	mu               sync.RWMutex
	events           map[string][]Event // aggregateID -> events
	snapshots        map[string]Snapshot
	snapshotInterval int
	globalVersion    int
}

func NewEventStore() *EventStore {
	return &EventStore{
		events:           make(map[string][]Event),
		snapshots:        make(map[string]Snapshot),
		snapshotInterval: 10,
	}
}

func (es *EventStore) AppendEvent(event Event) error {
	es.mu.Lock()
	defer es.mu.Unlock()

	// Set event metadata
	if event.ID == "" {
		event.ID = fmt.Sprintf("evt-%d", time.Now().UnixNano())
	}
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	es.globalVersion++
	event.Version = es.globalVersion

	// Get current events for aggregate
	events := es.events[event.AggregateID]

	// Optimistic concurrency check
	expectedVersion := len(events)
	if event.Version != 0 && expectedVersion != event.Version-1 {
		return fmt.Errorf("concurrency conflict: expected version %d, got %d", expectedVersion, event.Version-1)
	}

	// Append event
	es.events[event.AggregateID] = append(events, event)

	// Check if snapshot needed
	if len(es.events[event.AggregateID])%es.snapshotInterval == 0 {
		log.Printf("Snapshot threshold reached for aggregate %s", event.AggregateID)
	}

	return nil
}

func (es *EventStore) GetEvents(aggregateID string) []Event {
	es.mu.RLock()
	defer es.mu.RUnlock()

	events := es.events[aggregateID]
	result := make([]Event, len(events))
	copy(result, events)
	return result
}

func (es *EventStore) GetEventsSince(aggregateID string, version int) []Event {
	es.mu.RLock()
	defer es.mu.RUnlock()

	events := es.events[aggregateID]
	var result []Event
	for _, event := range events {
		if event.Version > version {
			result = append(result, event)
		}
	}
	return result
}

func (es *EventStore) SaveSnapshot(snapshot Snapshot) error {
	es.mu.Lock()
	defer es.mu.Unlock()

	snapshot.Timestamp = time.Now()
	es.snapshots[snapshot.AggregateID] = snapshot
	return nil
}

func (es *EventStore) GetSnapshot(aggregateID string) (Snapshot, bool) {
	es.mu.RLock()
	defer es.mu.RUnlock()

	snapshot, exists := es.snapshots[aggregateID]
	return snapshot, exists
}

func (es *EventStore) GetAllEvents() []Event {
	es.mu.RLock()
	defer es.mu.RUnlock()

	var allEvents []Event
	for _, events := range es.events {
		allEvents = append(allEvents, events...)
	}
	return allEvents
}

// Aggregate base interface
type Aggregate interface {
	GetID() string
	GetVersion() int
	Apply(event Event)
}

// BankAccount aggregate
type BankAccount struct {
	ID      string
	Balance int
	Status  string
	Version int
	Changes []Event
}

func NewBankAccount(id string) *BankAccount {
	return &BankAccount{
		ID:      id,
		Status:  "active",
		Changes: []Event{},
	}
}

func (a *BankAccount) GetID() string {
	return a.ID
}

func (a *BankAccount) GetVersion() int {
	return a.Version
}

func (a *BankAccount) Apply(event Event) {
	switch event.EventType {
	case "AccountCreated":
		data := event.Data.(map[string]interface{})
		a.Balance = int(data["initial_balance"].(float64))
		a.Status = "active"
	case "MoneyDeposited":
		data := event.Data.(map[string]interface{})
		a.Balance += int(data["amount"].(float64))
	case "MoneyWithdrawn":
		data := event.Data.(map[string]interface{})
		a.Balance -= int(data["amount"].(float64))
	case "AccountClosed":
		a.Status = "closed"
	}
	a.Version = event.Version
}

func (a *BankAccount) RecordChange(eventType string, data interface{}) {
	event := Event{
		AggregateID:   a.ID,
		AggregateType: "BankAccount",
		EventType:     eventType,
		Data:          data,
	}
	a.Changes = append(a.Changes, event)
	a.Apply(event)
}

func (a *BankAccount) CreateAccount(initialBalance int) error {
	if a.Version != 0 {
		return errors.New("account already exists")
	}
	a.RecordChange("AccountCreated", map[string]interface{}{
		"initial_balance": initialBalance,
	})
	return nil
}

func (a *BankAccount) Deposit(amount int) error {
	if amount <= 0 {
		return errors.New("amount must be positive")
	}
	if a.Status != "active" {
		return errors.New("account is not active")
	}
	a.RecordChange("MoneyDeposited", map[string]interface{}{
		"amount": amount,
	})
	return nil
}

func (a *BankAccount) Withdraw(amount int) error {
	if amount <= 0 {
		return errors.New("amount must be positive")
	}
	if a.Status != "active" {
		return errors.New("account is not active")
	}
	if a.Balance < amount {
		return errors.New("insufficient funds")
	}
	a.RecordChange("MoneyWithdrawn", map[string]interface{}{
		"amount": amount,
	})
	return nil
}

func (a *BankAccount) Close() error {
	if a.Status != "active" {
		return errors.New("account is not active")
	}
	a.RecordChange("AccountClosed", map[string]interface{}{})
	return nil
}

func (a *BankAccount) GetUncommittedChanges() []Event {
	return a.Changes
}

func (a *BankAccount) MarkChangesAsCommitted() {
	a.Changes = []Event{}
}

// Repository with event sourcing
type AccountRepository struct {
	eventStore *EventStore
}

func NewAccountRepository(eventStore *EventStore) *AccountRepository {
	return &AccountRepository{
		eventStore: eventStore,
	}
}

func (r *AccountRepository) Save(account *BankAccount) error {
	changes := account.GetUncommittedChanges()

	for _, event := range changes {
		if err := r.eventStore.AppendEvent(event); err != nil {
			return err
		}
	}

	account.MarkChangesAsCommitted()

	// Create snapshot if needed
	if account.Version%r.eventStore.snapshotInterval == 0 {
		snapshot := Snapshot{
			AggregateID:   account.ID,
			AggregateType: "BankAccount",
			Version:       account.Version,
			State:         account,
		}
		r.eventStore.SaveSnapshot(snapshot)
	}

	return nil
}

func (r *AccountRepository) GetByID(id string) (*BankAccount, error) {
	account := NewBankAccount(id)

	// Try to load from snapshot
	snapshot, exists := r.eventStore.GetSnapshot(id)
	if exists {
		snapshotAccount := snapshot.State.(*BankAccount)
		account.Balance = snapshotAccount.Balance
		account.Status = snapshotAccount.Status
		account.Version = snapshot.Version

		// Load events after snapshot
		events := r.eventStore.GetEventsSince(id, snapshot.Version)
		for _, event := range events {
			account.Apply(event)
		}
	} else {
		// Load all events
		events := r.eventStore.GetEvents(id)
		if len(events) == 0 {
			return nil, errors.New("account not found")
		}

		for _, event := range events {
			account.Apply(event)
		}
	}

	return account, nil
}

// Projection for read model
type AccountBalance struct {
	AccountID string
	Balance   int
	LastEvent string
	UpdatedAt time.Time
}

type BalanceProjection struct {
	mu       sync.RWMutex
	balances map[string]*AccountBalance
}

func NewBalanceProjection() *BalanceProjection {
	return &BalanceProjection{
		balances: make(map[string]*AccountBalance),
	}
}

func (p *BalanceProjection) Project(event Event) {
	p.mu.Lock()
	defer p.mu.Unlock()

	balance, exists := p.balances[event.AggregateID]
	if !exists {
		balance = &AccountBalance{
			AccountID: event.AggregateID,
		}
		p.balances[event.AggregateID] = balance
	}

	data := event.Data.(map[string]interface{})

	switch event.EventType {
	case "AccountCreated":
		balance.Balance = int(data["initial_balance"].(float64))
	case "MoneyDeposited":
		balance.Balance += int(data["amount"].(float64))
	case "MoneyWithdrawn":
		balance.Balance -= int(data["amount"].(float64))
	}

	balance.LastEvent = event.EventType
	balance.UpdatedAt = event.Timestamp
}

func (p *BalanceProjection) GetBalance(accountID string) (*AccountBalance, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	balance, exists := p.balances[accountID]
	if !exists {
		return nil, errors.New("balance not found")
	}
	return balance, nil
}

func main() {
	// Create event store
	eventStore := NewEventStore()
	repository := NewAccountRepository(eventStore)

	// Create projection
	projection := NewBalanceProjection()

	// Create account
	account := NewBankAccount("acc-123")
	account.CreateAccount(1000)

	// Deposit money
	account.Deposit(500)
	account.Deposit(200)

	// Withdraw money
	account.Withdraw(300)

	// Save to event store
	if err := repository.Save(account); err != nil {
		log.Fatalf("Failed to save account: %v", err)
	}

	fmt.Println("Events stored:")
	events := eventStore.GetEvents("acc-123")
	for _, event := range events {
		data, _ := json.MarshalIndent(event, "", "  ")
		fmt.Println(string(data))
	}

	// Rebuild from events
	fmt.Println("\nRebuilding account from events...")
	rebuilt, err := repository.GetByID("acc-123")
	if err != nil {
		log.Fatalf("Failed to rebuild account: %v", err)
	}

	fmt.Printf("Account ID: %s\n", rebuilt.ID)
	fmt.Printf("Balance: %d\n", rebuilt.Balance)
	fmt.Printf("Status: %s\n", rebuilt.Status)
	fmt.Printf("Version: %d\n", rebuilt.Version)

	// Build projection
	fmt.Println("\nBuilding projection...")
	for _, event := range events {
		projection.Project(event)
	}

	balance, _ := projection.GetBalance("acc-123")
	fmt.Printf("Projected balance: %d\n", balance.Balance)
	fmt.Printf("Last event: %s\n", balance.LastEvent)

	// Demonstrate time-travel
	fmt.Println("\nTime travel - state after 2 events:")
	timeTravelAccount := NewBankAccount("acc-123")
	for i := 0; i < 2 && i < len(events); i++ {
		timeTravelAccount.Apply(events[i])
	}
	fmt.Printf("Balance at version 2: %d\n", timeTravelAccount.Balance)
}
