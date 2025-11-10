# Challenge 147: Kafka Producer/Consumer

**Difficulty:** ⭐⭐⭐⭐ Hard
**Time Estimate:** 70 minutes

## Description

Implement Apache Kafka integration with producers and consumers using a Kafka client library. This project demonstrates high-throughput message streaming, consumer groups, partitioning, and offset management.

## Features

- **Producer**: Send messages to Kafka topics
- **Consumer**: Read messages from topics
- **Consumer Groups**: Distributed consumption
- **Partitioning**: Message distribution across partitions
- **Offset Management**: Manual and automatic commits
- **Batch Processing**: Efficient bulk operations
- **Compression**: Message compression (gzip, snappy)
- **Idempotence**: Exactly-once semantics
- **Transactions**: Atomic multi-message operations
- **Error Handling**: Retry and dead letter topics
- **Metrics**: Producer/consumer metrics
- **Rebalancing**: Handle consumer group rebalancing

## Requirements

1. Implement Kafka producer with batching
2. Create consumer with offset management
3. Support consumer groups
4. Handle partition assignment
5. Implement retry logic
6. Support message headers
7. Add compression support
8. Implement graceful shutdown
9. Handle rebalancing
10. Write comprehensive tests with mock Kafka

## Example Usage

```go
// Create producer
producer := NewKafkaProducer([]string{"localhost:9092"})
defer producer.Close()

// Produce message
msg := &Message{
    Topic: "user-events",
    Key:   []byte("user-123"),
    Value: []byte(`{"event": "created"}`),
    Headers: map[string]string{
        "source": "user-service",
    },
}
err := producer.Send(msg)

// Create consumer
consumer := NewKafkaConsumer([]string{"localhost:9092"}, "my-group")
defer consumer.Close()

// Subscribe to topics
consumer.Subscribe([]string{"user-events"})

// Consume messages
for msg := range consumer.Messages() {
    fmt.Printf("Received: %s\n", msg.Value)
    consumer.CommitMessage(msg)
}
```

## Learning Objectives

- Kafka architecture and concepts
- Producer/consumer patterns
- Partitioning strategies
- Consumer groups and rebalancing
- Offset management
- Exactly-once semantics
- Performance optimization
- Error handling strategies
- Monitoring and metrics

## Testing Focus

- Test message production
- Test message consumption
- Test consumer groups
- Test offset commits
- Test error handling
- Test rebalancing
- Mock Kafka for unit tests
- Benchmark throughput
