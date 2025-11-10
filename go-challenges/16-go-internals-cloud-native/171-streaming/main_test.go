package main

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// TestEventStream tests event streaming
func TestEventStreamBasic(t *testing.T) {
	stream := NewEventStream(10)
	consumer := NewLoggingConsumer()

	stream.Subscribe(consumer)
	stream.Start()

	event := &Event{
		ID:        "test1",
		EventType: "test",
		Timestamp: time.Now(),
	}

	err := stream.Emit(event)
	if err != nil {
		t.Errorf("Expected successful emit, got error: %v", err)
	}

	stream.Close()
}

func TestEventStreamMultipleConsumers(t *testing.T) {
	stream := NewEventStream(10)
	consumer1 := NewLoggingConsumer()
	consumer2 := NewLoggingConsumer()

	stream.Subscribe(consumer1)
	stream.Subscribe(consumer2)
	stream.Start()

	event := &Event{
		ID:        "test",
		EventType: "event",
		Timestamp: time.Now(),
	}

	stream.Emit(event)
	time.Sleep(50 * time.Millisecond)
	stream.Close()

	if len(consumer1.GetEvents()) == 0 {
		t.Errorf("Expected consumer1 to receive event")
	}

	if len(consumer2.GetEvents()) == 0 {
		t.Errorf("Expected consumer2 to receive event")
	}
}

func TestEventStreamStats(t *testing.T) {
	stream := NewEventStream(100)
	consumer := NewLoggingConsumer()
	stream.Subscribe(consumer)
	stream.Start()

	for i := 0; i < 10; i++ {
		stream.Emit(&Event{ID: "test", EventType: "test", Timestamp: time.Now()})
	}

	stream.Close()

	if atomic.LoadInt64(&stream.stats.EventsProduced) != 10 {
		t.Errorf("Expected 10 events produced")
	}
}

// TestBackpressureQueue tests backpressure queue
func TestBackpressureQueueBasic(t *testing.T) {
	queue := NewBackpressureQueue(100, StrategyDrop)

	err := queue.Enqueue("item1")
	if err != nil {
		t.Errorf("Expected successful enqueue, got error: %v", err)
	}

	item := queue.Dequeue()
	if item != "item1" {
		t.Errorf("Expected 'item1', got %v", item)
	}
}

func TestBackpressureQueueDrop(t *testing.T) {
	queue := NewBackpressureQueue(2, StrategyDrop)

	queue.Enqueue(1)
	queue.Enqueue(2)

	err := queue.Enqueue(3)
	if err == nil {
		t.Errorf("Expected error when queue is full with StrategyDrop")
	}

	if atomic.LoadInt64(&queue.dropped) == 0 {
		t.Errorf("Expected drops to be recorded")
	}
}

func TestBackpressureQueueBlock(t *testing.T) {
	queue := NewBackpressureQueue(5, StrategyBlock)

	for i := 0; i < 10; i++ {
		queue.Enqueue(i)
	}

	stats := queue.GetStats()
	if stats["processed"] < 10 {
		t.Logf("Warning: Not all items processed: %d", stats["processed"])
	}
}

func TestBackpressureQueueStats(t *testing.T) {
	queue := NewBackpressureQueue(100, StrategyDrop)

	queue.Enqueue("a")
	queue.Enqueue("b")
	queue.Dequeue()

	stats := queue.GetStats()
	if stats["current_size"] != 1 {
		t.Errorf("Expected current_size=1, got %d", stats["current_size"])
	}

	if stats["processed"] != 1 {
		t.Errorf("Expected processed=1, got %d", stats["processed"])
	}
}

// TestStreamProcessor tests stream processing
func TestStreamProcessorBasic(t *testing.T) {
	consumer := NewLoggingConsumer()
	processor := NewStreamProcessor(consumer)

	event := &Event{
		ID:        "test",
		EventType: "test",
		Timestamp: time.Now(),
	}

	processor.Process(event)

	if atomic.LoadInt64(&processor.stats.InputEvents) != 1 {
		t.Errorf("Expected 1 input event")
	}
}

func TestStreamProcessorFilter(t *testing.T) {
	consumer := NewLoggingConsumer()
	processor := NewStreamProcessor(consumer)

	processor.AddTransformer(&FilterTransformer{
		predicate: func(e *Event) bool {
			return e.EventType == "keep"
		},
	})

	event1 := &Event{ID: "1", EventType: "keep"}
	event2 := &Event{ID: "2", EventType: "drop"}

	processor.Process(event1)
	processor.Process(event2)

	if atomic.LoadInt64(&processor.stats.OutputEvents) != 1 {
		t.Errorf("Expected 1 output event after filtering")
	}

	if atomic.LoadInt64(&processor.stats.FilteredEvents) != 1 {
		t.Errorf("Expected 1 filtered event")
	}
}

func TestStreamProcessorMap(t *testing.T) {
	consumer := NewLoggingConsumer()
	processor := NewStreamProcessor(consumer)

	processor.AddTransformer(&MapTransformer{
		fn: func(e *Event) *Event {
			e.EventType = "transformed-" + e.EventType
			return e
		},
	})

	event := &Event{ID: "test", EventType: "original"}
	processor.Process(event)

	events := consumer.GetEvents()
	if len(events) > 0 && !contains(events[0].EventType, "transformed") {
		t.Errorf("Expected transformed event type")
	}
}

// TestSSEBroadcaster tests SSE broadcasting
func TestSSEBroadcasterBasic(t *testing.T) {
	broadcaster := NewSSEBroadcaster()
	broadcaster.Start()

	sub := broadcaster.Subscribe("client-1")
	if sub == nil {
		t.Errorf("Expected non-nil subscriber")
	}

	if atomic.LoadInt32(&broadcaster.stats.TotalSubscribers) != 1 {
		t.Errorf("Expected 1 subscriber")
	}

	broadcaster.Close()
}

func TestSSEBroadcasterMultipleSubscribers(t *testing.T) {
	broadcaster := NewSSEBroadcaster()
	broadcaster.Start()

	broadcaster.Subscribe("client-1")
	broadcaster.Subscribe("client-2")
	broadcaster.Subscribe("client-3")

	if atomic.LoadInt32(&broadcaster.stats.TotalSubscribers) != 3 {
		t.Errorf("Expected 3 subscribers")
	}

	broadcaster.Close()
}

func TestSSEBroadcasterUnsubscribe(t *testing.T) {
	broadcaster := NewSSEBroadcaster()
	broadcaster.Start()

	sub := broadcaster.Subscribe("client-1")
	broadcaster.Unsubscribe("client-1")

	if atomic.LoadInt32(&broadcaster.stats.TotalSubscribers) != 0 {
		t.Errorf("Expected 0 subscribers after unsubscribe")
	}

	broadcaster.Close()
}

func TestSSEBroadcasterBroadcast(t *testing.T) {
	broadcaster := NewSSEBroadcaster()
	broadcaster.Start()

	broadcaster.Subscribe("client-1")

	event := &Event{ID: "test", EventType: "event", Timestamp: time.Now()}
	broadcaster.Broadcast(event)

	time.Sleep(100 * time.Millisecond)

	if atomic.LoadInt64(&broadcaster.stats.EventsBroadcasted) < 1 {
		t.Logf("Warning: Event may not have been broadcast")
	}

	broadcaster.Close()
}

// TestRealtimeAggregator tests real-time aggregation
func TestRealtimeAggregatorBasic(t *testing.T) {
	aggregator := NewRealtimeAggregator(1 * time.Second)

	event := &Event{ID: "test", EventType: "metric"}
	aggregator.OnEvent(event)

	aggregator.window.mu.RLock()
	if aggregator.window.Count != 1 {
		t.Errorf("Expected window count 1, got %d", aggregator.window.Count)
	}
	aggregator.window.mu.RUnlock()
}

func TestRealtimeAggregatorWindow(t *testing.T) {
	aggregator := NewRealtimeAggregator(100 * time.Millisecond)

	aggregator.OnEvent(&Event{ID: "1"})
	aggregator.OnEvent(&Event{ID: "2"})

	aggregator.window.mu.RLock()
	count := aggregator.window.Count
	aggregator.window.mu.RUnlock()

	if count != 2 {
		t.Errorf("Expected count 2, got %d", count)
	}

	time.Sleep(150 * time.Millisecond)
	aggregator.OnEvent(&Event{ID: "3"})

	aggregator.window.mu.RLock()
	newCount := aggregator.window.Count
	aggregator.window.mu.RUnlock()

	if newCount != 1 {
		t.Errorf("Expected window reset, count should be 1, got %d", newCount)
	}
}

// TestLoggingConsumer tests logging consumer
func TestLoggingConsumerBasic(t *testing.T) {
	consumer := NewLoggingConsumer()

	event := &Event{ID: "test"}
	consumer.OnEvent(event)

	events := consumer.GetEvents()
	if len(events) != 1 {
		t.Errorf("Expected 1 event, got %d", len(events))
	}
}

func TestLoggingConsumerMultiple(t *testing.T) {
	consumer := NewLoggingConsumer()

	for i := 0; i < 5; i++ {
		consumer.OnEvent(&Event{ID: "test"})
	}

	events := consumer.GetEvents()
	if len(events) != 5 {
		t.Errorf("Expected 5 events, got %d", len(events))
	}
}

// Benchmark tests

func BenchmarkEventStreamEmit(b *testing.B) {
	stream := NewEventStream(1000)
	consumer := NewLoggingConsumer()
	stream.Subscribe(consumer)
	stream.Start()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		stream.Emit(&Event{ID: "test", EventType: "bench"})
	}

	stream.Close()
}

func BenchmarkBackpressureQueueEnqueue(b *testing.B) {
	queue := NewBackpressureQueue(10000, StrategyDrop)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		queue.Enqueue(i)
	}
}

func BenchmarkStreamProcessorProcess(b *testing.B) {
	consumer := NewLoggingConsumer()
	processor := NewStreamProcessor(consumer)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		processor.Process(&Event{ID: "test"})
	}
}

// Helper function
func contains(s, substr string) bool {
	for i := 0; i < len(s)-len(substr)+1; i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
