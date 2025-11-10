# Challenge 146: Event Sourcing Pattern

**Difficulty:** ⭐⭐⭐⭐ Hard
**Time Estimate:** 70 minutes

## Description

Implement the Event Sourcing pattern with event store, aggregates, and event replay. This project demonstrates storing state as a sequence of events, rebuilding state from events, and implementing CQRS-ready architecture.

## Features

- **Event Store**: Append-only event log
- **Aggregates**: Domain entities reconstructed from events
- **Event Versioning**: Handle schema evolution
- **Snapshots**: Optimize aggregate rebuilding
- **Event Replay**: Rebuild state from events
- **Projections**: Read models from events
- **Idempotency**: Handle duplicate events
- **Optimistic Concurrency**: Version-based concurrency control
- **Event Metadata**: Track causation and correlation
- **Time Travel**: Query state at any point in time
- **Event Upcasting**: Migrate old events
- **Stream Slicing**: Partition event streams

## Requirements

1. Implement event store with append-only log
2. Create aggregate base with event sourcing
3. Support event versioning
4. Implement snapshot mechanism
5. Build state from event stream
6. Handle concurrent modifications
7. Implement projections
8. Support time-travel queries
9. Add event metadata tracking
10. Write comprehensive tests

## Example Usage

```go
// Create event store
store := NewEventStore()

// Define aggregate
type Account struct {
    ID      string
    Balance int
    Version int
}

// Define events
type AccountCreated struct {
    AccountID string
    InitialBalance int
}

type MoneyDeposited struct {
    AccountID string
    Amount int
}

// Create account
account := &Account{ID: "acc-123"}
store.AppendEvent(account.ID, AccountCreated{
    AccountID: "acc-123",
    InitialBalance: 1000,
})

// Deposit money
store.AppendEvent(account.ID, MoneyDeposited{
    AccountID: "acc-123",
    Amount: 500,
})

// Rebuild from events
events := store.GetEvents(account.ID)
rebuilt := ReplayEvents(events)
```

## Learning Objectives

- Event sourcing principles
- Aggregate design patterns
- Event store implementation
- State reconstruction
- Snapshot strategies
- Concurrency control
- Event versioning
- Projection patterns
- Time-travel queries
- Domain-driven design concepts

## Testing Focus

- Test event appending
- Test state reconstruction
- Test snapshots
- Test concurrent modifications
- Test event versioning
- Test projections
- Test time-travel queries
- Benchmark event replay
