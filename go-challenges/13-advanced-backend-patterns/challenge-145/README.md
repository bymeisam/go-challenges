# Challenge 145: RabbitMQ Integration

**Difficulty:** ⭐⭐⭐⭐ Hard
**Time Estimate:** 70 minutes

## Description

Implement RabbitMQ message queue integration with producer/consumer patterns using amqp091-go. This project demonstrates message queuing, reliable delivery, acknowledgments, dead letter exchanges, and various exchange types.

## Features

- **Connection Management**: Robust RabbitMQ connections with reconnection
- **Producer Pattern**: Publish messages to exchanges
- **Consumer Pattern**: Consume messages from queues
- **Exchange Types**: Direct, Topic, Fanout, Headers
- **Message Acknowledgment**: Manual and automatic ack
- **Dead Letter Exchange**: Handle failed messages
- **Message TTL**: Time-to-live for messages
- **Priority Queues**: Message prioritization
- **Prefetch**: Control message flow
- **Publishing Confirms**: Ensure reliable delivery
- **Message Persistence**: Durable queues and messages
- **Retry Mechanism**: Automatic retry with backoff

## Requirements

1. Implement connection wrapper with auto-reconnect
2. Create producer with various exchange types
3. Implement consumer with acknowledgments
4. Support dead letter exchange
5. Handle connection failures gracefully
6. Implement message retries
7. Support priority queues
8. Use publishing confirms
9. Implement worker pools for consumers
10. Write comprehensive tests with mocks

## Example Usage

```go
// Connect to RabbitMQ
mq := NewRabbitMQ("amqp://guest:guest@localhost:5672/")
defer mq.Close()

// Declare exchange and queue
mq.DeclareExchange("events", "topic", true)
mq.DeclareQueue("user.events", true, false)
mq.BindQueue("user.events", "user.*", "events")

// Publish message
message := Message{
	RoutingKey: "user.created",
	Body:       []byte(`{"id": "123", "email": "user@example.com"}`),
	Priority:   5,
}
err := mq.Publish("events", message)

// Consume messages
messages, err := mq.Consume("user.events", "consumer-1")
for msg := range messages {
    fmt.Printf("Received: %s\n", msg.Body)
    msg.Ack(false)
}
```

## Learning Objectives

- RabbitMQ architecture and concepts
- AMQP protocol fundamentals
- Exchange types and routing
- Queue bindings and patterns
- Message acknowledgment strategies
- Dead letter exchange pattern
- Reliability and durability
- Connection management
- Error handling and retries
- Performance optimization

## Testing Focus

- Test connection establishment
- Test message publishing
- Test message consumption
- Test acknowledgments
- Test dead letter handling
- Test reconnection logic
- Test concurrent publishers
- Test concurrent consumers
- Mock RabbitMQ for unit tests
