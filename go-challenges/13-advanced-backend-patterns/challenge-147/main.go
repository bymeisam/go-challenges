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

// Kafka message
type KafkaMessage struct {
	Topic     string
	Partition int32
	Offset    int64
	Key       []byte
	Value     []byte
	Headers   map[string]string
	Timestamp time.Time
}

// Mock Kafka (simulates Kafka broker)
type MockKafka struct {
	mu         sync.RWMutex
	topics     map[string][][]KafkaMessage // topic -> partitions -> messages
	offsets    map[string]map[int32]int64  // consumerGroup -> partition -> offset
	partitions map[string]int32            // topic -> partition count
}

func NewMockKafka() *MockKafka {
	return &MockKafka{
		topics:     make(map[string][][]KafkaMessage),
		offsets:    make(map[string]map[int32]int64),
		partitions: make(map[string]int32),
	}
}

func (mk *MockKafka) CreateTopic(topic string, partitions int32) {
	mk.mu.Lock()
	defer mk.mu.Unlock()

	mk.topics[topic] = make([][]KafkaMessage, partitions)
	for i := int32(0); i < partitions; i++ {
		mk.topics[topic][i] = []KafkaMessage{}
	}
	mk.partitions[topic] = partitions
}

// Producer
type KafkaProducer struct {
	brokers   []string
	kafka     *MockKafka
	mu        sync.Mutex
	metrics   *ProducerMetrics
	batchSize int
	batch     []KafkaMessage
}

type ProducerMetrics struct {
	mu              sync.RWMutex
	MessagesSent    int64
	MessagesErrored int64
	BytesSent       int64
}

func (m *ProducerMetrics) IncrementSent(bytes int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.MessagesSent++
	m.BytesSent += int64(bytes)
}

func (m *ProducerMetrics) IncrementErrored() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.MessagesErrored++
}

func NewKafkaProducer(brokers []string) *KafkaProducer {
	return &KafkaProducer{
		brokers:   brokers,
		kafka:     NewMockKafka(),
		metrics:   &ProducerMetrics{},
		batchSize: 100,
		batch:     []KafkaMessage{},
	}
}

func (p *KafkaProducer) Send(msg *KafkaMessage) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Set timestamp if not provided
	if msg.Timestamp.IsZero() {
		msg.Timestamp = time.Now()
	}

	// Determine partition (simple hash of key)
	partitionCount := p.kafka.partitions[msg.Topic]
	if partitionCount == 0 {
		return errors.New("topic does not exist")
	}

	if msg.Key != nil {
		msg.Partition = int32(hash(msg.Key) % int(partitionCount))
	} else {
		msg.Partition = 0
	}

	// Add to batch
	p.batch = append(p.batch, *msg)

	// Flush if batch is full
	if len(p.batch) >= p.batchSize {
		return p.flush()
	}

	return nil
}

func (p *KafkaProducer) SendSync(msg *KafkaMessage) error {
	if err := p.Send(msg); err != nil {
		return err
	}
	return p.Flush()
}

func (p *KafkaProducer) Flush() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.flush()
}

func (p *KafkaProducer) flush() error {
	if len(p.batch) == 0 {
		return nil
	}

	p.kafka.mu.Lock()
	defer p.kafka.mu.Unlock()

	for _, msg := range p.batch {
		// Set offset
		partitions := p.kafka.topics[msg.Topic]
		if msg.Partition >= int32(len(partitions)) {
			p.metrics.IncrementErrored()
			continue
		}

		msg.Offset = int64(len(partitions[msg.Partition]))
		partitions[msg.Partition] = append(partitions[msg.Partition], msg)
		p.metrics.IncrementSent(len(msg.Value))
	}

	p.batch = []KafkaMessage{}
	return nil
}

func (p *KafkaProducer) GetMetrics() *ProducerMetrics {
	return p.metrics
}

func (p *KafkaProducer) Close() error {
	return p.Flush()
}

// Consumer
type KafkaConsumer struct {
	brokers       []string
	groupID       string
	kafka         *MockKafka
	topics        []string
	messages      chan KafkaMessage
	ctx           context.Context
	cancel        context.CancelFunc
	mu            sync.Mutex
	metrics       *ConsumerMetrics
	autoCommit    bool
	commitOffsets map[int32]int64
}

type ConsumerMetrics struct {
	mu                sync.RWMutex
	MessagesConsumed  int64
	BytesConsumed     int64
	CommittedOffsets  int64
}

func (m *ConsumerMetrics) IncrementConsumed(bytes int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.MessagesConsumed++
	m.BytesConsumed += int64(bytes)
}

func (m *ConsumerMetrics) IncrementCommitted() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.CommittedOffsets++
}

func NewKafkaConsumer(brokers []string, groupID string) *KafkaConsumer {
	ctx, cancel := context.WithCancel(context.Background())
	return &KafkaConsumer{
		brokers:       brokers,
		groupID:       groupID,
		kafka:         NewMockKafka(),
		messages:      make(chan KafkaMessage, 100),
		ctx:           ctx,
		cancel:        cancel,
		metrics:       &ConsumerMetrics{},
		autoCommit:    true,
		commitOffsets: make(map[int32]int64),
	}
}

func (c *KafkaConsumer) Subscribe(topics []string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.topics = topics
	go c.poll()
	return nil
}

func (c *KafkaConsumer) poll() {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-c.ctx.Done():
			return
		case <-ticker.C:
			c.fetchMessages()
		}
	}
}

func (c *KafkaConsumer) fetchMessages() {
	c.kafka.mu.RLock()
	defer c.kafka.mu.RUnlock()

	for _, topic := range c.topics {
		partitions, exists := c.kafka.topics[topic]
		if !exists {
			continue
		}

		for partitionID, messages := range partitions {
			// Get consumer group offset
			offset := c.getOffset(partitionID)

			// Fetch messages after offset
			for i := offset; i < int64(len(messages)); i++ {
				msg := messages[i]
				select {
				case c.messages <- msg:
					c.metrics.IncrementConsumed(len(msg.Value))
					if c.autoCommit {
						c.commitOffset(partitionID, msg.Offset+1)
					}
				default:
					// Channel full, will retry next poll
					return
				}
			}
		}
	}
}

func (c *KafkaConsumer) getOffset(partition int32) int64 {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check consumer group offset
	if groupOffsets, exists := c.kafka.offsets[c.groupID]; exists {
		if offset, exists := groupOffsets[partition]; exists {
			return offset
		}
	}

	return 0
}

func (c *KafkaConsumer) commitOffset(partition int32, offset int64) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.kafka.mu.Lock()
	defer c.kafka.mu.Unlock()

	if _, exists := c.kafka.offsets[c.groupID]; !exists {
		c.kafka.offsets[c.groupID] = make(map[int32]int64)
	}

	c.kafka.offsets[c.groupID][partition] = offset
	c.metrics.IncrementCommitted()
}

func (c *KafkaConsumer) Messages() <-chan KafkaMessage {
	return c.messages
}

func (c *KafkaConsumer) CommitMessage(msg KafkaMessage) error {
	c.commitOffset(msg.Partition, msg.Offset+1)
	return nil
}

func (c *KafkaConsumer) GetMetrics() *ConsumerMetrics {
	return c.metrics
}

func (c *KafkaConsumer) Close() error {
	c.cancel()
	close(c.messages)
	return nil
}

// Helper function
func hash(data []byte) int {
	h := 0
	for _, b := range data {
		h = h*31 + int(b)
	}
	if h < 0 {
		return -h
	}
	return h
}

// Example events
type UserEvent struct {
	EventType string `json:"event_type"`
	UserID    string `json:"user_id"`
	Email     string `json:"email"`
	Timestamp int64  `json:"timestamp"`
}

func main() {
	// Shared Kafka instance for demo
	kafka := NewMockKafka()
	kafka.CreateTopic("user-events", 3)

	// Create producer
	producer := NewKafkaProducer([]string{"localhost:9092"})
	producer.kafka = kafka // Share mock for demo
	defer producer.Close()

	// Create consumer
	consumer := NewKafkaConsumer([]string{"localhost:9092"}, "my-consumer-group")
	consumer.kafka = kafka // Share mock for demo
	defer consumer.Close()

	// Subscribe to topic
	consumer.Subscribe([]string{"user-events"})

	// Start consuming in background
	go func() {
		for msg := range consumer.Messages() {
			var event UserEvent
			json.Unmarshal(msg.Value, &event)
			fmt.Printf("[Consumer] Received: %s for user %s (partition: %d, offset: %d)\n",
				event.EventType, event.UserID, msg.Partition, msg.Offset)
			consumer.CommitMessage(msg)
		}
	}()

	// Wait for consumer to start
	time.Sleep(200 * time.Millisecond)

	// Produce messages
	fmt.Println("Producing messages...")
	for i := 1; i <= 10; i++ {
		event := UserEvent{
			EventType: "user.created",
			UserID:    fmt.Sprintf("user-%d", i),
			Email:     fmt.Sprintf("user%d@example.com", i),
			Timestamp: time.Now().Unix(),
		}

		data, _ := json.Marshal(event)
		msg := &KafkaMessage{
			Topic: "user-events",
			Key:   []byte(event.UserID),
			Value: data,
			Headers: map[string]string{
				"source": "user-service",
			},
		}

		if err := producer.Send(msg); err != nil {
			log.Printf("Failed to send message: %v", err)
		}
	}

	// Flush remaining messages
	producer.Flush()

	// Wait for consumption
	time.Sleep(500 * time.Millisecond)

	// Print metrics
	fmt.Println("\n--- Producer Metrics ---")
	pMetrics := producer.GetMetrics()
	fmt.Printf("Messages sent: %d\n", pMetrics.MessagesSent)
	fmt.Printf("Bytes sent: %d\n", pMetrics.BytesSent)

	fmt.Println("\n--- Consumer Metrics ---")
	cMetrics := consumer.GetMetrics()
	fmt.Printf("Messages consumed: %d\n", cMetrics.MessagesConsumed)
	fmt.Printf("Bytes consumed: %d\n", cMetrics.BytesConsumed)
	fmt.Printf("Committed offsets: %d\n", cMetrics.CommittedOffsets)

	fmt.Println("\nDemo completed!")
}
