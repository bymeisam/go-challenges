# Challenge 144: Event-Driven Architecture

**Difficulty:** ⭐⭐⭐⭐ Hard
**Time Estimate:** 60 minutes

## Description

Implement an event-driven architecture with an event bus using the pub/sub pattern. This project demonstrates event sourcing concepts, asynchronous communication, event handlers, and message passing using Go channels.

## Features

- **Event Bus**: Central message broker using channels
- **Pub/Sub Pattern**: Publishers and subscribers
- **Event Types**: Strongly typed event structures
- **Event Handlers**: Asynchronous event processing
- **Topic-based Routing**: Subscribe to specific event types
- **Wildcards**: Pattern matching for topics
- **Event Filtering**: Filter events by criteria
- **Multiple Subscribers**: One event, many handlers
- **Error Handling**: Handle failures gracefully
- **Event History**: Store events for replay
- **Dead Letter Queue**: Handle failed events
- **Metrics**: Track event throughput

## Requirements

1. Implement EventBus with pub/sub capabilities
2. Support multiple event types
3. Allow multiple subscribers per event type
4. Process events asynchronously
5. Handle subscriber errors gracefully
6. Support wildcard subscriptions
7. Implement event filtering
8. Add event history/logging
9. Implement graceful shutdown
10. Write comprehensive tests

## Example Usage

```go
// Create event bus
bus := NewEventBus()
defer bus.Close()

// Define events
type UserCreatedEvent struct {
    UserID string
    Email  string
    Time   time.Time
}

type OrderPlacedEvent struct {
    OrderID    string
    CustomerID string
    Total      float64
    Time       time.Time
}

// Subscribe to events
bus.Subscribe("user.created", func(event Event) error {
    e := event.Data.(UserCreatedEvent)
    fmt.Printf("New user: %s\n", e.Email)
    return nil
})

bus.Subscribe("order.*", func(event Event) error {
    fmt.Printf("Order event: %s\n", event.Type)
    return nil
})

// Publish events
bus.Publish(Event{
    Type: "user.created",
    Data: UserCreatedEvent{
        UserID: "123",
        Email:  "user@example.com",
        Time:   time.Now(),
    },
})

bus.Publish(Event{
    Type: "order.placed",
    Data: OrderPlacedEvent{
        OrderID:    "order-456",
        CustomerID: "123",
        Total:      99.99,
        Time:       time.Now(),
    },
})
```

## Learning Objectives

- Event-driven architecture principles
- Pub/sub pattern implementation
- Asynchronous message processing
- Channel-based communication
- Event routing and filtering
- Error handling in async systems
- Graceful shutdown patterns
- Event replay mechanisms
- Dead letter queue concepts
- Performance optimization

## Testing Focus

- Test event publishing
- Test subscriber registration
- Test event delivery
- Test multiple subscribers
- Test wildcard matching
- Test error handling
- Test concurrent publishing
- Test event history
- Benchmark throughput
