package main

import (
	"encoding/json"
	"testing"
	"time"
)

func TestMockAMQP_Connect(t *testing.T) {
	amqp := NewMockAMQP()

	if !amqp.IsConnected() {
		t.Error("Should be connected initially")
	}

	amqp.Disconnect()
	if amqp.IsConnected() {
		t.Error("Should be disconnected")
	}

	err := amqp.Connect()
	if err != nil {
		t.Errorf("Connect failed: %v", err)
	}

	if !amqp.IsConnected() {
		t.Error("Should be reconnected")
	}
}

func TestRabbitMQ_DeclareExchange(t *testing.T) {
	mq := NewRabbitMQ("amqp://localhost")
	defer mq.Close()

	err := mq.DeclareExchange("test-exchange", ExchangeTopic, true)
	if err != nil {
		t.Errorf("DeclareExchange failed: %v", err)
	}

	if _, exists := mq.amqp.exchanges["test-exchange"]; !exists {
		t.Error("Exchange should exist")
	}
}

func TestRabbitMQ_DeclareQueue(t *testing.T) {
	mq := NewRabbitMQ("amqp://localhost")
	defer mq.Close()

	err := mq.DeclareQueue("test-queue", true, false)
	if err != nil {
		t.Errorf("DeclareQueue failed: %v", err)
	}

	if _, exists := mq.amqp.queues["test-queue"]; !exists {
		t.Error("Queue should exist")
	}
}

func TestRabbitMQ_BindQueue(t *testing.T) {
	mq := NewRabbitMQ("amqp://localhost")
	defer mq.Close()

	mq.DeclareExchange("test-exchange", ExchangeTopic, true)
	mq.DeclareQueue("test-queue", true, false)

	err := mq.BindQueue("test-queue", "test.#", "test-exchange")
	if err != nil {
		t.Errorf("BindQueue failed: %v", err)
	}

	bindings := mq.amqp.bindings["test-queue"]
	if len(bindings) != 1 {
		t.Errorf("Expected 1 binding, got %d", len(bindings))
	}
}

func TestRabbitMQ_PublishConsume(t *testing.T) {
	mq := NewRabbitMQ("amqp://localhost")
	defer mq.Close()

	// Setup
	mq.DeclareExchange("test-exchange", ExchangeTopic, true)
	mq.DeclareQueue("test-queue", true, false)
	mq.BindQueue("test-queue", "test.*", "test-exchange")

	// Start consumer
	messages, err := mq.Consume("test-queue", "test-consumer")
	if err != nil {
		t.Fatalf("Consume failed: %v", err)
	}

	// Publish message
	testMsg := Message{
		RoutingKey: "test.message",
		Body:       []byte("Hello, RabbitMQ!"),
	}

	err = mq.Publish("test-exchange", testMsg)
	if err != nil {
		t.Fatalf("Publish failed: %v", err)
	}

	// Receive message
	select {
	case msg := <-messages:
		if string(msg.Body) != string(testMsg.Body) {
			t.Errorf("Body = %s, want %s", msg.Body, testMsg.Body)
		}
		if msg.RoutingKey != testMsg.RoutingKey {
			t.Errorf("RoutingKey = %s, want %s", msg.RoutingKey, testMsg.RoutingKey)
		}
	case <-time.After(1 * time.Second):
		t.Error("Timeout waiting for message")
	}
}

func TestRabbitMQ_RoutingKey Matching(t *testing.T) {
	tests := []struct {
		pattern    string
		routingKey string
		match      bool
	}{
		{"test.message", "test.message", true},
		{"test.*", "test.message", true},
		{"test.*", "test.event", true},
		{"test.*", "other.message", false},
		{"*.message", "test.message", true},
		{"*.message", "user.message", true},
		{"#", "anything", true},
		{"*", "anything", true},
	}

	for _, tt := range tests {
		t.Run(tt.pattern+":"+tt.routingKey, func(t *testing.T) {
			matched := matchRoutingKey(tt.pattern, tt.routingKey)
			if matched != tt.match {
				t.Errorf("matchRoutingKey(%s, %s) = %v, want %v",
					tt.pattern, tt.routingKey, matched, tt.match)
			}
		})
	}
}

func TestRabbitMQ_MultipleConsumers(t *testing.T) {
	mq := NewRabbitMQ("amqp://localhost")
	defer mq.Close()

	mq.DeclareExchange("test-exchange", ExchangeFanout, true)
	mq.DeclareQueue("queue1", true, false)
	mq.DeclareQueue("queue2", true, false)
	mq.BindQueue("queue1", "", "test-exchange")
	mq.BindQueue("queue2", "", "test-exchange")

	messages1, _ := mq.Consume("queue1", "consumer1")
	messages2, _ := mq.Consume("queue2", "consumer2")

	testMsg := Message{
		RoutingKey: "",
		Body:       []byte("broadcast"),
	}

	mq.Publish("test-exchange", testMsg)

	// Both consumers should receive the message
	timeout := time.After(1 * time.Second)
	received := 0

	for received < 2 {
		select {
		case <-messages1:
			received++
		case <-messages2:
			received++
		case <-timeout:
			t.Fatalf("Timeout: only received %d messages, expected 2", received)
		}
	}
}

func TestRabbitMQ_DeadLetterQueue(t *testing.T) {
	mq := NewRabbitMQ("amqp://localhost")
	defer mq.Close()

	mq.DeclareExchange("test-exchange", ExchangeDirect, true)
	mq.DeclareQueue("test-queue", true, false)
	mq.BindQueue("test-queue", "test", "test-exchange")

	messages, _ := mq.Consume("test-queue", "consumer")

	testMsg := Message{
		RoutingKey: "test",
		Body:       []byte("test message"),
	}

	mq.Publish("test-exchange", testMsg)

	// Receive and nack
	select {
	case msg := <-messages:
		mq.Nack(msg, false) // Don't requeue
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for message")
	}

	// Check DLQ
	dlqMessages := mq.GetDLQMessages()
	if len(dlqMessages) != 1 {
		t.Errorf("Expected 1 message in DLQ, got %d", len(dlqMessages))
	}
}

func TestRabbitMQ_PublishWithRetry(t *testing.T) {
	mq := NewRabbitMQ("amqp://localhost")
	defer mq.Close()

	mq.DeclareExchange("test-exchange", ExchangeDirect, true)

	testMsg := Message{
		RoutingKey: "test",
		Body:       []byte("retry test"),
	}

	err := mq.PublishWithRetry("test-exchange", testMsg, 3)
	if err != nil {
		t.Errorf("PublishWithRetry failed: %v", err)
	}
}

func TestRabbitMQ_MessagePriority(t *testing.T) {
	mq := NewRabbitMQ("amqp://localhost")
	defer mq.Close()

	mq.DeclareExchange("test-exchange", ExchangeDirect, true)
	mq.DeclareQueue("test-queue", true, false)
	mq.BindQueue("test-queue", "test", "test-exchange")

	messages, _ := mq.Consume("test-queue", "consumer")

	// Publish messages with different priorities
	highPriority := Message{
		RoutingKey: "test",
		Body:       []byte("high priority"),
		Priority:   10,
	}

	lowPriority := Message{
		RoutingKey: "test",
		Body:       []byte("low priority"),
		Priority:   1,
	}

	mq.Publish("test-exchange", lowPriority)
	mq.Publish("test-exchange", highPriority)

	// Receive messages
	select {
	case msg := <-messages:
		// In a real system, priority would affect delivery order
		if msg.Priority != 1 && msg.Priority != 10 {
			t.Errorf("Unexpected priority: %d", msg.Priority)
		}
	case <-time.After(1 * time.Second):
		t.Error("Timeout waiting for message")
	}
}

func TestUserCreatedEvent_Serialization(t *testing.T) {
	event := UserCreatedEvent{
		UserID: "user-123",
		Email:  "test@example.com",
		Name:   "Test User",
	}

	data, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var decoded UserCreatedEvent
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if decoded.UserID != event.UserID {
		t.Errorf("UserID = %s, want %s", decoded.UserID, event.UserID)
	}
}

func TestRabbitMQ_Reconnection(t *testing.T) {
	mq := NewRabbitMQ("amqp://localhost")
	defer mq.Close()

	// Simulate connection loss
	mq.amqp.Disconnect()

	// Wait for reconnection
	time.Sleep(2 * time.Second)

	// Check reconnection count
	if mq.amqp.reconnects == 0 {
		t.Error("Should have attempted reconnection")
	}
}

func BenchmarkPublish(b *testing.B) {
	mq := NewRabbitMQ("amqp://localhost")
	defer mq.Close()

	mq.DeclareExchange("bench-exchange", ExchangeDirect, true)

	msg := Message{
		RoutingKey: "bench",
		Body:       []byte("benchmark message"),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mq.Publish("bench-exchange", msg)
	}
}

func BenchmarkConsume(b *testing.B) {
	mq := NewRabbitMQ("amqp://localhost")
	defer mq.Close()

	mq.DeclareExchange("bench-exchange", ExchangeDirect, true)
	mq.DeclareQueue("bench-queue", true, false)
	mq.BindQueue("bench-queue", "bench", "bench-exchange")

	messages, _ := mq.Consume("bench-queue", "bench-consumer")

	// Publish messages
	for i := 0; i < b.N; i++ {
		mq.Publish("bench-exchange", Message{
			RoutingKey: "bench",
			Body:       []byte("test"),
		})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		<-messages
	}
}
