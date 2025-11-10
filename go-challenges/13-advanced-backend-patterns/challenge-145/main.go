package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"
)

// Message represents a message to be sent/received
type Message struct {
	ID         string
	RoutingKey string
	Body       []byte
	Headers    map[string]interface{}
	Priority   uint8
	Timestamp  time.Time
	Retries    int
}

// Exchange types
type ExchangeType string

const (
	ExchangeDirect  ExchangeType = "direct"
	ExchangeTopic   ExchangeType = "topic"
	ExchangeFanout  ExchangeType = "fanout"
	ExchangeHeaders ExchangeType = "headers"
)

// Queue configuration
type QueueConfig struct {
	Name       string
	Durable    bool
	AutoDelete bool
	Exclusive  bool
	Args       map[string]interface{}
}

// Exchange configuration
type ExchangeConfig struct {
	Name       string
	Type       ExchangeType
	Durable    bool
	AutoDelete bool
	Args       map[string]interface{}
}

// Mock AMQP for testing (simulates RabbitMQ)
type MockAMQP struct {
	mu         sync.RWMutex
	exchanges  map[string]*ExchangeConfig
	queues     map[string]*QueueConfig
	bindings   map[string][]string // queue -> routing keys
	messages   map[string][]Message
	dlq        map[string][]Message
	consumers  map[string]chan Message
	connected  bool
	reconnects int
}

func NewMockAMQP() *MockAMQP {
	return &MockAMQP{
		exchanges: make(map[string]*ExchangeConfig),
		queues:    make(map[string]*QueueConfig),
		bindings:  make(map[string][]string),
		messages:  make(map[string][]Message),
		dlq:       make(map[string][]Message),
		consumers: make(map[string]chan Message),
		connected: true,
	}
}

func (m *MockAMQP) Connect() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.connected {
		return nil
	}

	m.connected = true
	m.reconnects++
	return nil
}

func (m *MockAMQP) IsConnected() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.connected
}

func (m *MockAMQP) Disconnect() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.connected = false
}

// RabbitMQ wrapper
type RabbitMQ struct {
	amqp          *MockAMQP
	url           string
	reconnectWait time.Duration
	mu            sync.Mutex
	ctx           context.Context
	cancel        context.CancelFunc
}

func NewRabbitMQ(url string) *RabbitMQ {
	ctx, cancel := context.WithCancel(context.Background())
	mq := &RabbitMQ{
		amqp:          NewMockAMQP(),
		url:           url,
		reconnectWait: 5 * time.Second,
		ctx:           ctx,
		cancel:        cancel,
	}

	// Start connection monitor
	go mq.monitorConnection()

	return mq
}

func (mq *RabbitMQ) monitorConnection() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-mq.ctx.Done():
			return
		case <-ticker.C:
			if !mq.amqp.IsConnected() {
				log.Println("Connection lost, reconnecting...")
				if err := mq.amqp.Connect(); err != nil {
					log.Printf("Reconnection failed: %v", err)
				} else {
					log.Println("Reconnected successfully")
				}
			}
		}
	}
}

func (mq *RabbitMQ) DeclareExchange(name string, exchangeType ExchangeType, durable bool) error {
	mq.mu.Lock()
	defer mq.mu.Unlock()

	if !mq.amqp.IsConnected() {
		return errors.New("not connected")
	}

	config := &ExchangeConfig{
		Name:    name,
		Type:    exchangeType,
		Durable: durable,
	}

	mq.amqp.mu.Lock()
	mq.amqp.exchanges[name] = config
	mq.amqp.mu.Unlock()

	return nil
}

func (mq *RabbitMQ) DeclareQueue(name string, durable, autoDelete bool) error {
	mq.mu.Lock()
	defer mq.mu.Unlock()

	if !mq.amqp.IsConnected() {
		return errors.New("not connected")
	}

	config := &QueueConfig{
		Name:       name,
		Durable:    durable,
		AutoDelete: autoDelete,
	}

	mq.amqp.mu.Lock()
	mq.amqp.queues[name] = config
	if _, exists := mq.amqp.messages[name]; !exists {
		mq.amqp.messages[name] = []Message{}
	}
	mq.amqp.mu.Unlock()

	return nil
}

func (mq *RabbitMQ) BindQueue(queueName, routingKey, exchangeName string) error {
	mq.mu.Lock()
	defer mq.mu.Unlock()

	if !mq.amqp.IsConnected() {
		return errors.New("not connected")
	}

	mq.amqp.mu.Lock()
	defer mq.amqp.mu.Unlock()

	if _, exists := mq.amqp.queues[queueName]; !exists {
		return errors.New("queue does not exist")
	}

	if _, exists := mq.amqp.exchanges[exchangeName]; !exists {
		return errors.New("exchange does not exist")
	}

	bindingKey := fmt.Sprintf("%s:%s", exchangeName, routingKey)
	mq.amqp.bindings[queueName] = append(mq.amqp.bindings[queueName], bindingKey)

	return nil
}

func (mq *RabbitMQ) Publish(exchangeName string, msg Message) error {
	mq.mu.Lock()
	defer mq.mu.Unlock()

	if !mq.amqp.IsConnected() {
		return errors.New("not connected")
	}

	// Set message metadata
	if msg.Timestamp.IsZero() {
		msg.Timestamp = time.Now()
	}
	if msg.ID == "" {
		msg.ID = fmt.Sprintf("msg-%d", time.Now().UnixNano())
	}

	mq.amqp.mu.Lock()
	defer mq.amqp.mu.Unlock()

	// Find queues bound to this exchange with matching routing key
	for queueName, bindings := range mq.amqp.bindings {
		for _, binding := range bindings {
			parts := splitBinding(binding)
			if len(parts) == 2 && parts[0] == exchangeName {
				if matchRoutingKey(parts[1], msg.RoutingKey) {
					mq.amqp.messages[queueName] = append(mq.amqp.messages[queueName], msg)

					// Notify consumers
					if ch, exists := mq.amqp.consumers[queueName]; exists {
						select {
						case ch <- msg:
						default:
						}
					}
				}
			}
		}
	}

	return nil
}

func (mq *RabbitMQ) Consume(queueName, consumerTag string) (<-chan Message, error) {
	mq.mu.Lock()
	defer mq.mu.Unlock()

	if !mq.amqp.IsConnected() {
		return nil, errors.New("not connected")
	}

	mq.amqp.mu.Lock()
	defer mq.amqp.mu.Unlock()

	if _, exists := mq.amqp.queues[queueName]; !exists {
		return nil, errors.New("queue does not exist")
	}

	msgChan := make(chan Message, 100)
	mq.amqp.consumers[queueName] = msgChan

	// Send existing messages
	go func() {
		mq.amqp.mu.Lock()
		messages := mq.amqp.messages[queueName]
		mq.amqp.messages[queueName] = []Message{}
		mq.amqp.mu.Unlock()

		for _, msg := range messages {
			msgChan <- msg
		}
	}()

	return msgChan, nil
}

func (mq *RabbitMQ) Ack(msg Message) error {
	// Simulate acknowledgment
	return nil
}

func (mq *RabbitMQ) Nack(msg Message, requeue bool) error {
	if !requeue {
		// Send to DLQ
		mq.amqp.mu.Lock()
		mq.amqp.dlq["dlq"] = append(mq.amqp.dlq["dlq"], msg)
		mq.amqp.mu.Unlock()
	}
	return nil
}

func (mq *RabbitMQ) PublishWithRetry(exchangeName string, msg Message, maxRetries int) error {
	var err error
	for i := 0; i <= maxRetries; i++ {
		err = mq.Publish(exchangeName, msg)
		if err == nil {
			return nil
		}

		if i < maxRetries {
			wait := time.Duration(i+1) * time.Second
			log.Printf("Publish failed, retrying in %v: %v", wait, err)
			time.Sleep(wait)
		}
	}
	return fmt.Errorf("publish failed after %d retries: %w", maxRetries, err)
}

func (mq *RabbitMQ) GetDLQMessages() []Message {
	mq.amqp.mu.RLock()
	defer mq.amqp.mu.RUnlock()
	return mq.amqp.dlq["dlq"]
}

func (mq *RabbitMQ) Close() error {
	mq.cancel()
	mq.amqp.Disconnect()
	return nil
}

// Helper functions
func splitBinding(binding string) []string {
	parts := []string{}
	idx := 0
	for i, ch := range binding {
		if ch == ':' {
			parts = append(parts, binding[idx:i])
			idx = i + 1
		}
	}
	if idx < len(binding) {
		parts = append(parts, binding[idx:])
	}
	return parts
}

func matchRoutingKey(pattern, key string) bool {
	// Simple pattern matching (supports * for wildcards)
	if pattern == "#" || pattern == "*" {
		return true
	}

	patternParts := splitByDot(pattern)
	keyParts := splitByDot(key)

	if len(patternParts) != len(keyParts) {
		return false
	}

	for i, part := range patternParts {
		if part != "*" && part != keyParts[i] {
			return false
		}
	}

	return true
}

func splitByDot(s string) []string {
	parts := []string{}
	idx := 0
	for i, ch := range s {
		if ch == '.' {
			parts = append(parts, s[idx:i])
			idx = i + 1
		}
	}
	if idx < len(s) {
		parts = append(parts, s[idx:])
	}
	return parts
}

// Example domain events
type UserCreatedEvent struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Name   string `json:"name"`
}

type OrderPlacedEvent struct {
	OrderID    string  `json:"order_id"`
	CustomerID string  `json:"customer_id"`
	Total      float64 `json:"total"`
}

func main() {
	// Create RabbitMQ instance
	mq := NewRabbitMQ("amqp://guest:guest@localhost:5672/")
	defer mq.Close()

	// Declare exchange
	if err := mq.DeclareExchange("events", ExchangeTopic, true); err != nil {
		log.Fatalf("Failed to declare exchange: %v", err)
	}

	// Declare queues
	if err := mq.DeclareQueue("user.events", true, false); err != nil {
		log.Fatalf("Failed to declare queue: %v", err)
	}

	if err := mq.DeclareQueue("order.events", true, false); err != nil {
		log.Fatalf("Failed to declare queue: %v", err)
	}

	// Bind queues
	mq.BindQueue("user.events", "user.*", "events")
	mq.BindQueue("order.events", "order.*", "events")

	// Start consumer for user events
	go func() {
		messages, err := mq.Consume("user.events", "user-consumer")
		if err != nil {
			log.Printf("Failed to start consumer: %v", err)
			return
		}

		for msg := range messages {
			fmt.Printf("[User Consumer] Received: %s\n", msg.Body)
			mq.Ack(msg)
		}
	}()

	// Start consumer for order events
	go func() {
		messages, err := mq.Consume("order.events", "order-consumer")
		if err != nil {
			log.Printf("Failed to start consumer: %v", err)
			return
		}

		for msg := range messages {
			fmt.Printf("[Order Consumer] Received: %s\n", msg.Body)
			mq.Ack(msg)
		}
	}()

	// Wait for consumers to start
	time.Sleep(100 * time.Millisecond)

	// Publish user event
	userEvent := UserCreatedEvent{
		UserID: "user-123",
		Email:  "user@example.com",
		Name:   "John Doe",
	}
	userBody, _ := json.Marshal(userEvent)

	mq.Publish("events", Message{
		RoutingKey: "user.created",
		Body:       userBody,
		Priority:   5,
	})

	// Publish order event
	orderEvent := OrderPlacedEvent{
		OrderID:    "order-456",
		CustomerID: "user-123",
		Total:      99.99,
	}
	orderBody, _ := json.Marshal(orderEvent)

	mq.Publish("events", Message{
		RoutingKey: "order.placed",
		Body:       orderBody,
		Priority:   3,
	})

	// Wait for processing
	time.Sleep(200 * time.Millisecond)

	fmt.Println("\nDemo completed!")
}
