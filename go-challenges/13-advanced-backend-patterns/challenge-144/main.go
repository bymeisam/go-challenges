package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"
)

// Event represents a domain event
type Event struct {
	ID        string
	Type      string
	Data      interface{}
	Timestamp time.Time
	Metadata  map[string]string
}

// EventHandler processes events
type EventHandler func(Event) error

// Subscriber represents an event subscriber
type Subscriber struct {
	ID      string
	Pattern string
	Handler EventHandler
	Filter  func(Event) bool
}

// EventBus manages pub/sub for events
type EventBus struct {
	mu          sync.RWMutex
	subscribers map[string][]*Subscriber
	history     []Event
	maxHistory  int
	dlq         chan Event
	metrics     *Metrics
	ctx         context.Context
	cancel      context.CancelFunc
	wg          sync.WaitGroup
}

// Metrics tracks event bus performance
type Metrics struct {
	mu              sync.RWMutex
	PublishedEvents int64
	DeliveredEvents int64
	FailedEvents    int64
	Subscribers     int64
}

func (m *Metrics) IncrementPublished() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.PublishedEvents++
}

func (m *Metrics) IncrementDelivered() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.DeliveredEvents++
}

func (m *Metrics) IncrementFailed() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.FailedEvents++
}

func (m *Metrics) GetStats() (published, delivered, failed int64) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.PublishedEvents, m.DeliveredEvents, m.FailedEvents
}

// NewEventBus creates a new event bus
func NewEventBus() *EventBus {
	ctx, cancel := context.WithCancel(context.Background())
	bus := &EventBus{
		subscribers: make(map[string][]*Subscriber),
		history:     make([]Event, 0),
		maxHistory:  1000,
		dlq:         make(chan Event, 100),
		metrics:     &Metrics{},
		ctx:         ctx,
		cancel:      cancel,
	}

	// Start DLQ processor
	bus.wg.Add(1)
	go bus.processDLQ()

	return bus
}

// Subscribe registers a handler for events matching the pattern
func (eb *EventBus) Subscribe(pattern string, handler EventHandler) (string, error) {
	return eb.SubscribeWithFilter(pattern, handler, nil)
}

// SubscribeWithFilter registers a handler with a custom filter
func (eb *EventBus) SubscribeWithFilter(pattern string, handler EventHandler, filter func(Event) bool) (string, error) {
	if handler == nil {
		return "", errors.New("handler cannot be nil")
	}

	eb.mu.Lock()
	defer eb.mu.Unlock()

	sub := &Subscriber{
		ID:      fmt.Sprintf("sub-%d", time.Now().UnixNano()),
		Pattern: pattern,
		Handler: handler,
		Filter:  filter,
	}

	eb.subscribers[pattern] = append(eb.subscribers[pattern], sub)
	eb.metrics.Subscribers++

	return sub.ID, nil
}

// Unsubscribe removes a subscriber
func (eb *EventBus) Unsubscribe(subscriberID string) error {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	for pattern, subs := range eb.subscribers {
		for i, sub := range subs {
			if sub.ID == subscriberID {
				eb.subscribers[pattern] = append(subs[:i], subs[i+1:]...)
				eb.metrics.Subscribers--
				return nil
			}
		}
	}

	return errors.New("subscriber not found")
}

// Publish sends an event to all matching subscribers
func (eb *EventBus) Publish(event Event) error {
	// Set timestamp if not provided
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	// Set ID if not provided
	if event.ID == "" {
		event.ID = fmt.Sprintf("evt-%d", time.Now().UnixNano())
	}

	eb.mu.Lock()
	// Add to history
	eb.history = append(eb.history, event)
	if len(eb.history) > eb.maxHistory {
		eb.history = eb.history[1:]
	}
	eb.mu.Unlock()

	eb.metrics.IncrementPublished()

	// Find matching subscribers
	subscribers := eb.findMatchingSubscribers(event.Type)

	// Deliver to subscribers asynchronously
	for _, sub := range subscribers {
		// Apply filter if present
		if sub.Filter != nil && !sub.Filter(event) {
			continue
		}

		eb.wg.Add(1)
		go func(s *Subscriber) {
			defer eb.wg.Done()
			if err := s.Handler(event); err != nil {
				log.Printf("Handler error for %s: %v", event.Type, err)
				eb.metrics.IncrementFailed()
				// Send to DLQ
				select {
				case eb.dlq <- event:
				default:
					log.Printf("DLQ full, dropping event: %s", event.ID)
				}
			} else {
				eb.metrics.IncrementDelivered()
			}
		}(sub)
	}

	return nil
}

// PublishSync publishes an event and waits for all handlers to complete
func (eb *EventBus) PublishSync(event Event) error {
	// Set timestamp if not provided
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	// Set ID if not provided
	if event.ID == "" {
		event.ID = fmt.Sprintf("evt-%d", time.Now().UnixNano())
	}

	eb.mu.Lock()
	eb.history = append(eb.history, event)
	if len(eb.history) > eb.maxHistory {
		eb.history = eb.history[1:]
	}
	eb.mu.Unlock()

	eb.metrics.IncrementPublished()

	subscribers := eb.findMatchingSubscribers(event.Type)

	var errs []error
	for _, sub := range subscribers {
		if sub.Filter != nil && !sub.Filter(event) {
			continue
		}

		if err := sub.Handler(event); err != nil {
			errs = append(errs, err)
			eb.metrics.IncrementFailed()
		} else {
			eb.metrics.IncrementDelivered()
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("handler errors: %v", errs)
	}

	return nil
}

// findMatchingSubscribers finds subscribers matching the event type
func (eb *EventBus) findMatchingSubscribers(eventType string) []*Subscriber {
	eb.mu.RLock()
	defer eb.mu.RUnlock()

	var matched []*Subscriber

	for pattern, subs := range eb.subscribers {
		if matchPattern(pattern, eventType) {
			matched = append(matched, subs...)
		}
	}

	return matched
}

// matchPattern checks if an event type matches a subscription pattern
func matchPattern(pattern, eventType string) bool {
	// Exact match
	if pattern == eventType {
		return true
	}

	// Wildcard match (simple implementation)
	// Supports: "user.*", "*.created", "*"
	if pattern == "*" {
		return true
	}

	parts := strings.Split(pattern, ".")
	eventParts := strings.Split(eventType, ".")

	if len(parts) != len(eventParts) {
		return false
	}

	for i, part := range parts {
		if part != "*" && part != eventParts[i] {
			return false
		}
	}

	return true
}

// GetHistory returns event history
func (eb *EventBus) GetHistory(limit int) []Event {
	eb.mu.RLock()
	defer eb.mu.RUnlock()

	if limit <= 0 || limit > len(eb.history) {
		limit = len(eb.history)
	}

	start := len(eb.history) - limit
	history := make([]Event, limit)
	copy(history, eb.history[start:])

	return history
}

// Replay replays events from history
func (eb *EventBus) Replay(fromTime time.Time) error {
	eb.mu.RLock()
	var eventsToReplay []Event
	for _, event := range eb.history {
		if event.Timestamp.After(fromTime) {
			eventsToReplay = append(eventsToReplay, event)
		}
	}
	eb.mu.RUnlock()

	for _, event := range eventsToReplay {
		if err := eb.Publish(event); err != nil {
			return err
		}
	}

	return nil
}

// processDLQ handles failed events
func (eb *EventBus) processDLQ() {
	defer eb.wg.Done()

	for {
		select {
		case <-eb.ctx.Done():
			return
		case event := <-eb.dlq:
			log.Printf("DLQ: Processing failed event: %s (type: %s)", event.ID, event.Type)
			// In a real system, you might retry, log to storage, send alerts, etc.
		}
	}
}

// GetMetrics returns current metrics
func (eb *EventBus) GetMetrics() *Metrics {
	return eb.metrics
}

// Close shuts down the event bus gracefully
func (eb *EventBus) Close() error {
	eb.cancel()

	// Wait for all handlers to complete (with timeout)
	done := make(chan struct{})
	go func() {
		eb.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Println("EventBus closed gracefully")
	case <-time.After(5 * time.Second):
		log.Println("EventBus shutdown timeout")
	}

	return nil
}

// Example event types
type UserCreatedEvent struct {
	UserID string
	Email  string
	Name   string
}

type UserUpdatedEvent struct {
	UserID string
	Fields map[string]string
}

type OrderPlacedEvent struct {
	OrderID    string
	CustomerID string
	Total      float64
	Items      []string
}

type OrderShippedEvent struct {
	OrderID      string
	TrackingCode string
}

func main() {
	// Create event bus
	bus := NewEventBus()
	defer bus.Close()

	// Subscribe to user events
	bus.Subscribe("user.created", func(event Event) error {
		e := event.Data.(UserCreatedEvent)
		fmt.Printf("[Email Service] Sending welcome email to: %s\n", e.Email)
		return nil
	})

	bus.Subscribe("user.created", func(event Event) error {
		e := event.Data.(UserCreatedEvent)
		fmt.Printf("[Analytics] New user registered: %s\n", e.UserID)
		return nil
	})

	// Subscribe to all user events with wildcard
	bus.Subscribe("user.*", func(event Event) error {
		fmt.Printf("[Audit] User event: %s\n", event.Type)
		return nil
	})

	// Subscribe to order events
	bus.Subscribe("order.placed", func(event Event) error {
		e := event.Data.(OrderPlacedEvent)
		fmt.Printf("[Payment] Processing payment for order: %s ($%.2f)\n", e.OrderID, e.Total)
		return nil
	})

	bus.Subscribe("order.shipped", func(event Event) error {
		e := event.Data.(OrderShippedEvent)
		fmt.Printf("[Notification] Order %s shipped. Tracking: %s\n", e.OrderID, e.TrackingCode)
		return nil
	})

	// Subscribe with filter
	bus.SubscribeWithFilter("order.placed", func(event Event) error {
		e := event.Data.(OrderPlacedEvent)
		fmt.Printf("[VIP Alert] High-value order: %s\n", e.OrderID)
		return nil
	}, func(event Event) bool {
		e := event.Data.(OrderPlacedEvent)
		return e.Total > 100.0 // Only high-value orders
	})

	// Publish events
	fmt.Println("\n--- Publishing Events ---\n")

	bus.Publish(Event{
		Type: "user.created",
		Data: UserCreatedEvent{
			UserID: "user-123",
			Email:  "john@example.com",
			Name:   "John Doe",
		},
	})

	time.Sleep(100 * time.Millisecond)

	bus.Publish(Event{
		Type: "order.placed",
		Data: OrderPlacedEvent{
			OrderID:    "order-456",
			CustomerID: "user-123",
			Total:      49.99,
			Items:      []string{"item1", "item2"},
		},
	})

	time.Sleep(100 * time.Millisecond)

	bus.Publish(Event{
		Type: "order.placed",
		Data: OrderPlacedEvent{
			OrderID:    "order-789",
			CustomerID: "user-123",
			Total:      199.99,
			Items:      []string{"item3"},
		},
	})

	time.Sleep(100 * time.Millisecond)

	bus.Publish(Event{
		Type: "order.shipped",
		Data: OrderShippedEvent{
			OrderID:      "order-789",
			TrackingCode: "TRACK123456",
		},
	})

	// Wait for async processing
	time.Sleep(200 * time.Millisecond)

	// Show metrics
	fmt.Println("\n--- Metrics ---")
	published, delivered, failed := bus.GetMetrics().GetStats()
	fmt.Printf("Published: %d, Delivered: %d, Failed: %d\n", published, delivered, failed)

	// Show history
	fmt.Println("\n--- Event History ---")
	history := bus.GetHistory(10)
	for _, event := range history {
		fmt.Printf("%s: %s at %s\n", event.ID, event.Type, event.Timestamp.Format(time.RFC3339))
	}
}
