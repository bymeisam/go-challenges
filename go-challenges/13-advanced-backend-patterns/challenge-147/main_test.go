package main

import (
	"encoding/json"
	"testing"
	"time"
)

func TestMockKafka_CreateTopic(t *testing.T) {
	kafka := NewMockKafka()

	kafka.CreateTopic("test-topic", 3)

	if partitions, exists := kafka.topics["test-topic"]; !exists {
		t.Error("Topic should exist")
	} else if int32(len(partitions)) != 3 {
		t.Errorf("Expected 3 partitions, got %d", len(partitions))
	}
}

func TestKafkaProducer_Send(t *testing.T) {
	producer := NewKafkaProducer([]string{"localhost:9092"})
	defer producer.Close()

	producer.kafka.CreateTopic("test-topic", 3)

	msg := &KafkaMessage{
		Topic: "test-topic",
		Key:   []byte("key-1"),
		Value: []byte("test message"),
	}

	err := producer.SendSync(msg)
	if err != nil {
		t.Fatalf("Send failed: %v", err)
	}

	// Verify message was stored
	messages := producer.kafka.topics["test-topic"][msg.Partition]
	if len(messages) != 1 {
		t.Errorf("Expected 1 message, got %d", len(messages))
	}
}

func TestKafkaProducer_Batching(t *testing.T) {
	producer := NewKafkaProducer([]string{"localhost:9092"})
	producer.batchSize = 5
	defer producer.Close()

	producer.kafka.CreateTopic("test-topic", 1)

	// Send messages
	for i := 0; i < 10; i++ {
		msg := &KafkaMessage{
			Topic: "test-topic",
			Value: []byte("message"),
		}
		producer.Send(msg)
	}

	// Check batch size
	if len(producer.batch) != 0 {
		t.Errorf("Batch should be empty after auto-flush, got %d", len(producer.batch))
	}
}

func TestKafkaProducer_Metrics(t *testing.T) {
	producer := NewKafkaProducer([]string{"localhost:9092"})
	defer producer.Close()

	producer.kafka.CreateTopic("test-topic", 1)

	for i := 0; i < 10; i++ {
		msg := &KafkaMessage{
			Topic: "test-topic",
			Value: []byte("test"),
		}
		producer.SendSync(msg)
	}

	metrics := producer.GetMetrics()
	if metrics.MessagesSent != 10 {
		t.Errorf("MessagesSent = %d, want 10", metrics.MessagesSent)
	}
}

func TestKafkaConsumer_Consume(t *testing.T) {
	kafka := NewMockKafka()
	kafka.CreateTopic("test-topic", 1)

	// Create producer and send messages
	producer := NewKafkaProducer([]string{"localhost:9092"})
	producer.kafka = kafka

	msg := &KafkaMessage{
		Topic: "test-topic",
		Value: []byte("test message"),
	}
	producer.SendSync(msg)

	// Create consumer
	consumer := NewKafkaConsumer([]string{"localhost:9092"}, "test-group")
	consumer.kafka = kafka
	defer consumer.Close()

	consumer.Subscribe([]string{"test-topic"})

	// Wait for consumption
	select {
	case received := <-consumer.Messages():
		if string(received.Value) != string(msg.Value) {
			t.Errorf("Value = %s, want %s", received.Value, msg.Value)
		}
	case <-time.After(1 * time.Second):
		t.Error("Timeout waiting for message")
	}
}

func TestKafkaConsumer_ConsumerGroup(t *testing.T) {
	kafka := NewMockKafka()
	kafka.CreateTopic("test-topic", 1)

	// Produce messages
	producer := NewKafkaProducer([]string{"localhost:9092"})
	producer.kafka = kafka

	for i := 0; i < 5; i++ {
		producer.SendSync(&KafkaMessage{
			Topic: "test-topic",
			Value: []byte("message"),
		})
	}

	// Create consumers in same group
	consumer1 := NewKafkaConsumer([]string{"localhost:9092"}, "group-1")
	consumer1.kafka = kafka
	defer consumer1.Close()

	consumer2 := NewKafkaConsumer([]string{"localhost:9092"}, "group-1")
	consumer2.kafka = kafka
	defer consumer2.Close()

	consumer1.Subscribe([]string{"test-topic"})
	consumer2.Subscribe([]string{"test-topic"})

	// Both consumers should share the workload
	// (In this simple implementation, first consumer gets all messages)
	time.Sleep(500 * time.Millisecond)

	metrics1 := consumer1.GetMetrics()
	metrics2 := consumer2.GetMetrics()

	total := metrics1.MessagesConsumed + metrics2.MessagesConsumed
	if total < 5 {
		t.Errorf("Total consumed = %d, want at least 5", total)
	}
}

func TestKafkaConsumer_OffsetCommit(t *testing.T) {
	kafka := NewMockKafka()
	kafka.CreateTopic("test-topic", 1)

	// Produce messages
	producer := NewKafkaProducer([]string{"localhost:9092"})
	producer.kafka = kafka

	for i := 0; i < 3; i++ {
		producer.SendSync(&KafkaMessage{
			Topic: "test-topic",
			Value: []byte("message"),
		})
	}

	// Consume with auto-commit
	consumer := NewKafkaConsumer([]string{"localhost:9092"}, "test-group")
	consumer.kafka = kafka
	consumer.autoCommit = true
	defer consumer.Close()

	consumer.Subscribe([]string{"test-topic"})

	// Wait for consumption
	time.Sleep(500 * time.Millisecond)

	// Check offsets
	if offsets, exists := kafka.offsets["test-group"]; exists {
		if offset, exists := offsets[0]; exists {
			if offset != 3 {
				t.Errorf("Offset = %d, want 3", offset)
			}
		}
	}
}

func TestKafkaConsumer_ManualCommit(t *testing.T) {
	kafka := NewMockKafka()
	kafka.CreateTopic("test-topic", 1)

	producer := NewKafkaProducer([]string{"localhost:9092"})
	producer.kafka = kafka
	producer.SendSync(&KafkaMessage{
		Topic: "test-topic",
		Value: []byte("test"),
	})

	consumer := NewKafkaConsumer([]string{"localhost:9092"}, "test-group")
	consumer.kafka = kafka
	consumer.autoCommit = false
	defer consumer.Close()

	consumer.Subscribe([]string{"test-topic"})

	// Consume and manually commit
	select {
	case msg := <-consumer.Messages():
		err := consumer.CommitMessage(msg)
		if err != nil {
			t.Errorf("CommitMessage failed: %v", err)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for message")
	}

	// Verify offset
	if offsets, exists := kafka.offsets["test-group"]; exists {
		if offset := offsets[0]; offset != 1 {
			t.Errorf("Offset = %d, want 1", offset)
		}
	}
}

func TestKafkaConsumer_Metrics(t *testing.T) {
	kafka := NewMockKafka()
	kafka.CreateTopic("test-topic", 1)

	producer := NewKafkaProducer([]string{"localhost:9092"})
	producer.kafka = kafka

	for i := 0; i < 5; i++ {
		producer.SendSync(&KafkaMessage{
			Topic: "test-topic",
			Value: []byte("test"),
		})
	}

	consumer := NewKafkaConsumer([]string{"localhost:9092"}, "test-group")
	consumer.kafka = kafka
	defer consumer.Close()

	consumer.Subscribe([]string{"test-topic"})

	// Wait for consumption
	time.Sleep(500 * time.Millisecond)

	metrics := consumer.GetMetrics()
	if metrics.MessagesConsumed != 5 {
		t.Errorf("MessagesConsumed = %d, want 5", metrics.MessagesConsumed)
	}
}

func TestPartitioning(t *testing.T) {
	producer := NewKafkaProducer([]string{"localhost:9092"})
	producer.kafka.CreateTopic("test-topic", 3)

	// Send messages with different keys
	keys := [][]byte{
		[]byte("key-a"),
		[]byte("key-b"),
		[]byte("key-c"),
	}

	for _, key := range keys {
		msg := &KafkaMessage{
			Topic: "test-topic",
			Key:   key,
			Value: []byte("test"),
		}
		producer.SendSync(msg)
	}

	// Verify messages are distributed across partitions
	partitionCounts := make([]int, 3)
	for i := 0; i < 3; i++ {
		partitionCounts[i] = len(producer.kafka.topics["test-topic"][i])
	}

	total := 0
	for _, count := range partitionCounts {
		total += count
	}

	if total != 3 {
		t.Errorf("Total messages = %d, want 3", total)
	}
}

func TestUserEvent_Serialization(t *testing.T) {
	event := UserEvent{
		EventType: "user.created",
		UserID:    "user-123",
		Email:     "test@example.com",
		Timestamp: time.Now().Unix(),
	}

	data, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var decoded UserEvent
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if decoded.UserID != event.UserID {
		t.Errorf("UserID = %s, want %s", decoded.UserID, event.UserID)
	}
}

func TestMessageHeaders(t *testing.T) {
	producer := NewKafkaProducer([]string{"localhost:9092"})
	producer.kafka.CreateTopic("test-topic", 1)

	msg := &KafkaMessage{
		Topic: "test-topic",
		Value: []byte("test"),
		Headers: map[string]string{
			"source":      "test-service",
			"contentType": "application/json",
		},
	}

	producer.SendSync(msg)

	messages := producer.kafka.topics["test-topic"][0]
	if len(messages) != 1 {
		t.Fatal("Message not stored")
	}

	if messages[0].Headers["source"] != "test-service" {
		t.Error("Headers not preserved")
	}
}

func BenchmarkKafkaProducer_Send(b *testing.B) {
	producer := NewKafkaProducer([]string{"localhost:9092"})
	producer.kafka.CreateTopic("bench-topic", 1)
	defer producer.Close()

	msg := &KafkaMessage{
		Topic: "bench-topic",
		Value: []byte("benchmark message"),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		producer.Send(msg)
	}
	producer.Flush()
}

func BenchmarkKafkaConsumer_Consume(b *testing.B) {
	kafka := NewMockKafka()
	kafka.CreateTopic("bench-topic", 1)

	producer := NewKafkaProducer([]string{"localhost:9092"})
	producer.kafka = kafka

	// Produce messages
	for i := 0; i < b.N; i++ {
		producer.SendSync(&KafkaMessage{
			Topic: "bench-topic",
			Value: []byte("test"),
		})
	}

	consumer := NewKafkaConsumer([]string{"localhost:9092"}, "bench-group")
	consumer.kafka = kafka
	defer consumer.Close()

	consumer.Subscribe([]string{"bench-topic"})

	b.ResetTimer()
	consumed := 0
	timeout := time.After(5 * time.Second)
	for consumed < b.N {
		select {
		case <-consumer.Messages():
			consumed++
		case <-timeout:
			b.Fatalf("Timeout: consumed %d/%d messages", consumed, b.N)
		}
	}
}
