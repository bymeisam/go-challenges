package main

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestEventBus_Subscribe(t *testing.T) {
	bus := NewEventBus()
	defer bus.Close()

	called := false
	handler := func(event Event) error {
		called = true
		return nil
	}

	subID, err := bus.Subscribe("test.event", handler)
	if err != nil {
		t.Fatalf("Subscribe failed: %v", err)
	}

	if subID == "" {
		t.Error("Subscribe should return subscriber ID")
	}

	if len(bus.subscribers["test.event"]) != 1 {
		t.Errorf("Expected 1 subscriber, got %d", len(bus.subscribers["test.event"]))
	}
}

func TestEventBus_Unsubscribe(t *testing.T) {
	bus := NewEventBus()
	defer bus.Close()

	handler := func(event Event) error { return nil }
	subID, _ := bus.Subscribe("test.event", handler)

	err := bus.Unsubscribe(subID)
	if err != nil {
		t.Errorf("Unsubscribe failed: %v", err)
	}

	if len(bus.subscribers["test.event"]) != 0 {
		t.Error("Subscriber should be removed")
	}

	// Try to unsubscribe again
	err = bus.Unsubscribe(subID)
	if err == nil {
		t.Error("Unsubscribe should fail for non-existent subscriber")
	}
}

func TestEventBus_PublishSync(t *testing.T) {
	bus := NewEventBus()
	defer bus.Close()

	var called int32
	handler := func(event Event) error {
		atomic.AddInt32(&called, 1)
		return nil
	}

	bus.Subscribe("test.event", handler)

	event := Event{
		Type: "test.event",
		Data: "test data",
	}

	err := bus.PublishSync(event)
	if err != nil {
		t.Fatalf("PublishSync failed: %v", err)
	}

	if atomic.LoadInt32(&called) != 1 {
		t.Errorf("Handler called %d times, expected 1", called)
	}
}

func TestEventBus_PublishAsync(t *testing.T) {
	bus := NewEventBus()
	defer bus.Close()

	var called int32
	handler := func(event Event) error {
		atomic.AddInt32(&called, 1)
		return nil
	}

	bus.Subscribe("test.event", handler)

	event := Event{
		Type: "test.event",
		Data: "test data",
	}

	err := bus.Publish(event)
	if err != nil {
		t.Fatalf("Publish failed: %v", err)
	}

	// Wait for async processing
	time.Sleep(50 * time.Millisecond)

	if atomic.LoadInt32(&called) != 1 {
		t.Errorf("Handler called %d times, expected 1", called)
	}
}

func TestEventBus_MultipleSubscribers(t *testing.T) {
	bus := NewEventBus()
	defer bus.Close()

	var count int32

	handler1 := func(event Event) error {
		atomic.AddInt32(&count, 1)
		return nil
	}

	handler2 := func(event Event) error {
		atomic.AddInt32(&count, 10)
		return nil
	}

	handler3 := func(event Event) error {
		atomic.AddInt32(&count, 100)
		return nil
	}

	bus.Subscribe("test.event", handler1)
	bus.Subscribe("test.event", handler2)
	bus.Subscribe("test.event", handler3)

	event := Event{
		Type: "test.event",
		Data: "test",
	}

	bus.Publish(event)
	time.Sleep(50 * time.Millisecond)

	expected := int32(111) // 1 + 10 + 100
	if atomic.LoadInt32(&count) != expected {
		t.Errorf("Count = %d, expected %d", count, expected)
	}
}

func TestEventBus_WildcardMatching(t *testing.T) {
	tests := []struct {
		pattern   string
		eventType string
		match     bool
	}{
		{"user.created", "user.created", true},
		{"user.*", "user.created", true},
		{"user.*", "user.updated", true},
		{"user.*", "order.created", false},
		{"*.created", "user.created", true},
		{"*.created", "order.created", true},
		{"*.created", "user.updated", false},
		{"*", "any.event", true},
		{"user.*.event", "user.test.event", true},
		{"user.*.event", "user.created", false},
	}

	for _, tt := range tests {
		t.Run(tt.pattern+":"+tt.eventType, func(t *testing.T) {
			matched := matchPattern(tt.pattern, tt.eventType)
			if matched != tt.match {
				t.Errorf("matchPattern(%s, %s) = %v, want %v",
					tt.pattern, tt.eventType, matched, tt.match)
			}
		})
	}
}

func TestEventBus_WildcardSubscription(t *testing.T) {
	bus := NewEventBus()
	defer bus.Close()

	var events []string
	var mu sync.Mutex

	handler := func(event Event) error {
		mu.Lock()
		events = append(events, event.Type)
		mu.Unlock()
		return nil
	}

	bus.Subscribe("user.*", handler)

	bus.PublishSync(Event{Type: "user.created", Data: nil})
	bus.PublishSync(Event{Type: "user.updated", Data: nil})
	bus.PublishSync(Event{Type: "order.created", Data: nil})

	if len(events) != 2 {
		t.Errorf("Expected 2 events, got %d", len(events))
	}
}

func TestEventBus_ErrorHandling(t *testing.T) {
	bus := NewEventBus()
	defer bus.Close()

	testErr := errors.New("handler error")

	handler := func(event Event) error {
		return testErr
	}

	bus.Subscribe("test.event", handler)

	err := bus.PublishSync(Event{Type: "test.event", Data: nil})
	if err == nil {
		t.Error("PublishSync should return error when handler fails")
	}
}

func TestEventBus_Filter(t *testing.T) {
	bus := NewEventBus()
	defer bus.Close()

	var called int32

	handler := func(event Event) error {
		atomic.AddInt32(&called, 1)
		return nil
	}

	filter := func(event Event) bool {
		data := event.Data.(int)
		return data > 10
	}

	bus.SubscribeWithFilter("test.event", handler, filter)

	// Should not trigger handler (5 <= 10)
	bus.PublishSync(Event{Type: "test.event", Data: 5})

	// Should trigger handler (15 > 10)
	bus.PublishSync(Event{Type: "test.event", Data: 15})

	if atomic.LoadInt32(&called) != 1 {
		t.Errorf("Handler called %d times, expected 1", called)
	}
}

func TestEventBus_History(t *testing.T) {
	bus := NewEventBus()
	defer bus.Close()

	for i := 0; i < 5; i++ {
		bus.Publish(Event{
			Type: "test.event",
			Data: i,
		})
	}

	time.Sleep(50 * time.Millisecond)

	history := bus.GetHistory(10)
	if len(history) != 5 {
		t.Errorf("Expected 5 events in history, got %d", len(history))
	}

	// Test limit
	history = bus.GetHistory(3)
	if len(history) != 3 {
		t.Errorf("Expected 3 events in history, got %d", len(history))
	}
}

func TestEventBus_Replay(t *testing.T) {
	bus := NewEventBus()
	defer bus.Close()

	var count int32
	handler := func(event Event) error {
		atomic.AddInt32(&count, 1)
		return nil
	}

	// Publish some events
	for i := 0; i < 3; i++ {
		bus.Publish(Event{Type: "test.event", Data: i})
	}

	time.Sleep(50 * time.Millisecond)

	// Now subscribe (should not receive past events yet)
	bus.Subscribe("test.event", handler)

	// Replay from beginning
	bus.Replay(time.Time{})

	time.Sleep(100 * time.Millisecond)

	if atomic.LoadInt32(&count) != 3 {
		t.Errorf("Handler called %d times, expected 3", count)
	}
}

func TestEventBus_Metrics(t *testing.T) {
	bus := NewEventBus()
	defer bus.Close()

	handler := func(event Event) error {
		return nil
	}

	bus.Subscribe("test.event", handler)

	for i := 0; i < 10; i++ {
		bus.Publish(Event{Type: "test.event", Data: i})
	}

	time.Sleep(100 * time.Millisecond)

	published, delivered, failed := bus.GetMetrics().GetStats()

	if published != 10 {
		t.Errorf("Published = %d, expected 10", published)
	}

	if delivered != 10 {
		t.Errorf("Delivered = %d, expected 10", delivered)
	}

	if failed != 0 {
		t.Errorf("Failed = %d, expected 0", failed)
	}
}

func TestEventBus_ConcurrentPublish(t *testing.T) {
	bus := NewEventBus()
	defer bus.Close()

	var count int32
	handler := func(event Event) error {
		atomic.AddInt32(&count, 1)
		return nil
	}

	bus.Subscribe("test.event", handler)

	const goroutines = 100
	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()
			bus.Publish(Event{Type: "test.event", Data: id})
		}(i)
	}

	wg.Wait()
	time.Sleep(200 * time.Millisecond)

	if atomic.LoadInt32(&count) != goroutines {
		t.Errorf("Handler called %d times, expected %d", count, goroutines)
	}
}

func TestEventBus_ConcurrentSubscribe(t *testing.T) {
	bus := NewEventBus()
	defer bus.Close()

	const goroutines = 50
	var wg sync.WaitGroup
	wg.Add(goroutines)

	handler := func(event Event) error {
		return nil
	}

	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			bus.Subscribe("test.event", handler)
		}()
	}

	wg.Wait()

	if len(bus.subscribers["test.event"]) != goroutines {
		t.Errorf("Expected %d subscribers, got %d", goroutines, len(bus.subscribers["test.event"]))
	}
}

func TestEventBus_GracefulShutdown(t *testing.T) {
	bus := NewEventBus()

	var count int32
	handler := func(event Event) error {
		time.Sleep(50 * time.Millisecond)
		atomic.AddInt32(&count, 1)
		return nil
	}

	bus.Subscribe("test.event", handler)

	// Publish events
	for i := 0; i < 5; i++ {
		bus.Publish(Event{Type: "test.event", Data: i})
	}

	// Close should wait for handlers
	err := bus.Close()
	if err != nil {
		t.Errorf("Close failed: %v", err)
	}

	// All events should be processed
	if atomic.LoadInt32(&count) != 5 {
		t.Errorf("Handler called %d times, expected 5", count)
	}
}

func BenchmarkPublish(b *testing.B) {
	bus := NewEventBus()
	defer bus.Close()

	handler := func(event Event) error {
		return nil
	}

	bus.Subscribe("test.event", handler)

	event := Event{Type: "test.event", Data: "test"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bus.Publish(event)
	}
}

func BenchmarkPublishSync(b *testing.B) {
	bus := NewEventBus()
	defer bus.Close()

	handler := func(event Event) error {
		return nil
	}

	bus.Subscribe("test.event", handler)

	event := Event{Type: "test.event", Data: "test"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bus.PublishSync(event)
	}
}

func BenchmarkMultipleSubscribers(b *testing.B) {
	bus := NewEventBus()
	defer bus.Close()

	handler := func(event Event) error {
		return nil
	}

	// Add 10 subscribers
	for i := 0; i < 10; i++ {
		bus.Subscribe("test.event", handler)
	}

	event := Event{Type: "test.event", Data: "test"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bus.Publish(event)
	}
}
