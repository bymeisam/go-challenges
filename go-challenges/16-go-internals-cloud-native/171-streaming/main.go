package main

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// Challenge 171: Streaming & Real-time Data Processing
// SSE, Backpressure, Stream Processing, Reactive Patterns

// ===== 1. Event Stream =====

type Event struct {
	ID        string
	EventType string
	Data      interface{}
	Timestamp time.Time
}

type StreamConsumer interface {
	OnEvent(event *Event)
	OnError(err error)
	OnComplete()
}

type EventStream struct {
	consumers  []StreamConsumer
	eventChan  chan *Event
	mu         sync.RWMutex
	closed     bool
	stats      *StreamStats
	stopChan   chan struct{}
}

type StreamStats struct {
	EventsProduced int64
	EventsConsumed int64
	Errors         int64
}

func NewEventStream(bufferSize int) *EventStream {
	return &EventStream{
		consumers: make([]StreamConsumer, 0),
		eventChan: make(chan *Event, bufferSize),
		stats:     &StreamStats{},
		stopChan:  make(chan struct{}),
	}
}

func (es *EventStream) Subscribe(consumer StreamConsumer) {
	es.mu.Lock()
	defer es.mu.Unlock()

	es.consumers = append(es.consumers, consumer)
}

func (es *EventStream) Emit(event *Event) error {
	es.mu.RLock()
	closed := es.closed
	es.mu.RUnlock()

	if closed {
		return fmt.Errorf("stream closed")
	}

	select {
	case es.eventChan <- event:
		atomic.AddInt64(&es.stats.EventsProduced, 1)
		return nil
	case <-es.stopChan:
		return fmt.Errorf("stream closed")
	default:
		atomic.AddInt64(&es.stats.Errors, 1)
		return fmt.Errorf("event buffer full")
	}
}

func (es *EventStream) Start() {
	go func() {
		for {
			select {
			case event := <-es.eventChan:
				es.mu.RLock()
				consumers := make([]StreamConsumer, len(es.consumers))
				copy(consumers, es.consumers)
				es.mu.RUnlock()

				for _, consumer := range consumers {
					consumer.OnEvent(event)
				}
				atomic.AddInt64(&es.stats.EventsConsumed, 1)

			case <-es.stopChan:
				return
			}
		}
	}()
}

func (es *EventStream) Close() {
	es.mu.Lock()
	defer es.mu.Unlock()

	if !es.closed {
		es.closed = true
		close(es.stopChan)

		for _, consumer := range es.consumers {
			consumer.OnComplete()
		}
	}
}

// ===== 2. Backpressure Queue =====

type BackpressureQueue struct {
	queue      chan interface{}
	maxSize    int64
	currentSize int64
	dropped    int64
	processed  int64
	mu         sync.RWMutex
	strategy   BackpressureStrategy
}

type BackpressureStrategy int

const (
	StrategyDrop BackpressureStrategy = iota
	StrategyBlock
	StrategyAdaptive
)

func NewBackpressureQueue(maxSize int64, strategy BackpressureStrategy) *BackpressureQueue {
	return &BackpressureQueue{
		queue:      make(chan interface{}, maxSize/2),
		maxSize:    maxSize,
		strategy:   strategy,
	}
}

func (bq *BackpressureQueue) Enqueue(item interface{}) error {
	currentSize := atomic.LoadInt64(&bq.currentSize)

	switch bq.strategy {
	case StrategyDrop:
		if currentSize >= bq.maxSize {
			atomic.AddInt64(&bq.dropped, 1)
			return fmt.Errorf("queue full, item dropped")
		}

	case StrategyBlock:
		for atomic.LoadInt64(&bq.currentSize) >= bq.maxSize {
			time.Sleep(10 * time.Millisecond)
		}

	case StrategyAdaptive:
		waitTime := time.Duration(0)
		for atomic.LoadInt64(&bq.currentSize) >= bq.maxSize {
			time.Sleep(waitTime)
			waitTime += 5 * time.Millisecond
			if waitTime > 100*time.Millisecond {
				waitTime = 100 * time.Millisecond
			}
		}
	}

	select {
	case bq.queue <- item:
		atomic.AddInt64(&bq.currentSize, 1)
		return nil
	default:
		atomic.AddInt64(&bq.dropped, 1)
		return fmt.Errorf("queue error")
	}
}

func (bq *BackpressureQueue) Dequeue() interface{} {
	item := <-bq.queue
	atomic.AddInt64(&bq.currentSize, -1)
	atomic.AddInt64(&bq.processed, 1)
	return item
}

func (bq *BackpressureQueue) GetStats() map[string]int64 {
	return map[string]int64{
		"current_size": atomic.LoadInt64(&bq.currentSize),
		"max_size":     bq.maxSize,
		"dropped":      atomic.LoadInt64(&bq.dropped),
		"processed":    atomic.LoadInt64(&bq.processed),
	}
}

// ===== 3. Stream Processor =====

type StreamProcessor struct {
	transformers []Transformer
	sink         StreamConsumer
	mu           sync.RWMutex
	stats        *ProcessorStats
}

type Transformer interface {
	Transform(event *Event) *Event
}

type ProcessorStats struct {
	InputEvents   int64
	OutputEvents  int64
	FilteredEvents int64
}

type MapTransformer struct {
	fn func(*Event) *Event
}

func (mt *MapTransformer) Transform(event *Event) *Event {
	return mt.fn(event)
}

type FilterTransformer struct {
	predicate func(*Event) bool
}

func (ft *FilterTransformer) Transform(event *Event) *Event {
	if ft.predicate(event) {
		return event
	}
	return nil
}

func NewStreamProcessor(sink StreamConsumer) *StreamProcessor {
	return &StreamProcessor{
		transformers: make([]Transformer, 0),
		sink:         sink,
		stats:        &ProcessorStats{},
	}
}

func (sp *StreamProcessor) AddTransformer(transformer Transformer) {
	sp.mu.Lock()
	defer sp.mu.Unlock()
	sp.transformers = append(sp.transformers, transformer)
}

func (sp *StreamProcessor) Process(event *Event) {
	atomic.AddInt64(&sp.stats.InputEvents, 1)

	sp.mu.RLock()
	transformers := make([]Transformer, len(sp.transformers))
	copy(transformers, sp.transformers)
	sp.mu.RUnlock()

	current := event
	for _, transformer := range transformers {
		if current == nil {
			atomic.AddInt64(&sp.stats.FilteredEvents, 1)
			return
		}
		current = transformer.Transform(current)
	}

	if current != nil {
		sp.sink.OnEvent(current)
		atomic.AddInt64(&sp.stats.OutputEvents, 1)
	}
}

// ===== 4. SSE Server Simulation =====

type SSEBroadcaster struct {
	subscribers map[string]*SSESubscriber
	mu          sync.RWMutex
	eventChan   chan *Event
	stats       *SSEStats
	stopChan    chan struct{}
}

type SSESubscriber struct {
	ID       string
	EventChan chan *Event
	Done     chan struct{}
}

type SSEStats struct {
	TotalSubscribers   int32
	EventsBroadcasted  int64
	SubscriberErrors   int64
}

func NewSSEBroadcaster() *SSEBroadcaster {
	return &SSEBroadcaster{
		subscribers: make(map[string]*SSESubscriber),
		eventChan:   make(chan *Event, 100),
		stats:       &SSEStats{},
		stopChan:    make(chan struct{}),
	}
}

func (sb *SSEBroadcaster) Subscribe(subscriberID string) *SSESubscriber {
	sb.mu.Lock()
	defer sb.mu.Unlock()

	sub := &SSESubscriber{
		ID:        subscriberID,
		EventChan: make(chan *Event, 10),
		Done:      make(chan struct{}),
	}

	sb.subscribers[subscriberID] = sub
	atomic.AddInt32(&sb.stats.TotalSubscribers, 1)

	return sub
}

func (sb *SSEBroadcaster) Unsubscribe(subscriberID string) {
	sb.mu.Lock()
	defer sb.mu.Unlock()

	if sub, exists := sb.subscribers[subscriberID]; exists {
		close(sub.Done)
		delete(sb.subscribers, subscriberID)
		atomic.AddInt32(&sb.stats.TotalSubscribers, -1)
	}
}

func (sb *SSEBroadcaster) Broadcast(event *Event) {
	select {
	case sb.eventChan <- event:
	case <-sb.stopChan:
	}
}

func (sb *SSEBroadcaster) Start() {
	go func() {
		for {
			select {
			case event := <-sb.eventChan:
				sb.mu.RLock()
				subscribers := make(map[string]*SSESubscriber)
				for k, v := range sb.subscribers {
					subscribers[k] = v
				}
				sb.mu.RUnlock()

				for _, sub := range subscribers {
					select {
					case sub.EventChan <- event:
						atomic.AddInt64(&sb.stats.EventsBroadcasted, 1)
					default:
						atomic.AddInt64(&sb.stats.SubscriberErrors, 1)
					}
				}

			case <-sb.stopChan:
				return
			}
		}
	}()
}

func (sb *SSEBroadcaster) Close() {
	close(sb.stopChan)
}

// ===== 5. Real-time Aggregator =====

type AggregationWindow struct {
	Duration    time.Duration
	LastWindow  time.Time
	Aggregation map[string]interface{}
	Count       int64
	mu          sync.RWMutex
}

type RealtimeAggregator struct {
	window     *AggregationWindow
	aggregator func([]*Event) map[string]interface{}
	output     chan map[string]interface{}
	mu         sync.RWMutex
	stopChan   chan struct{}
}

func NewRealtimeAggregator(duration time.Duration) *RealtimeAggregator {
	return &RealtimeAggregator{
		window: &AggregationWindow{
			Duration:    duration,
			LastWindow:  time.Now(),
			Aggregation: make(map[string]interface{}),
		},
		output:   make(chan map[string]interface{}, 10),
		stopChan: make(chan struct{}),
	}
}

func (ra *RealtimeAggregator) SetAggregator(fn func([]*Event) map[string]interface{}) {
	ra.mu.Lock()
	defer ra.mu.Unlock()
	ra.aggregator = fn
}

func (ra *RealtimeAggregator) OnEvent(event *Event) {
	ra.window.mu.Lock()
	defer ra.window.mu.Unlock()

	now := time.Now()
	if now.Sub(ra.window.LastWindow) > ra.window.Duration {
		ra.window.Aggregation = make(map[string]interface{})
		ra.window.LastWindow = now
		ra.window.Count = 0
	}

	ra.window.Count++
}

func (ra *RealtimeAggregator) OnError(err error) {}

func (ra *RealtimeAggregator) OnComplete() {}

// ===== 6. Consumer Implementation =====

type LoggingConsumer struct {
	events []*Event
	mu     sync.RWMutex
}

func NewLoggingConsumer() *LoggingConsumer {
	return &LoggingConsumer{
		events: make([]*Event, 0),
	}
}

func (lc *LoggingConsumer) OnEvent(event *Event) {
	lc.mu.Lock()
	defer lc.mu.Unlock()
	lc.events = append(lc.events, event)
}

func (lc *LoggingConsumer) OnError(err error) {
	fmt.Printf("Error: %v\n", err)
}

func (lc *LoggingConsumer) OnComplete() {
	fmt.Println("Stream completed")
}

func (lc *LoggingConsumer) GetEvents() []*Event {
	lc.mu.RLock()
	defer lc.mu.RUnlock()

	events := make([]*Event, len(lc.events))
	copy(events, lc.events)
	return events
}

// ===== Main Demo =====

func main() {
	fmt.Println("=== Streaming & Real-time Data Processing ===\n")

	// 1. Event Stream
	fmt.Println("1. Event Stream")
	stream := NewEventStream(100)
	consumer := NewLoggingConsumer()
	stream.Subscribe(consumer)
	stream.Start()

	for i := 0; i < 5; i++ {
		event := &Event{
			ID:        fmt.Sprintf("evt-%d", i),
			EventType: "test",
			Data:      fmt.Sprintf("data-%d", i),
			Timestamp: time.Now(),
		}
		stream.Emit(event)
	}

	stream.Close()
	fmt.Printf("Stream stats: Produced=%d, Consumed=%d, Errors=%d\n\n",
		atomic.LoadInt64(&stream.stats.EventsProduced),
		atomic.LoadInt64(&stream.stats.EventsConsumed),
		atomic.LoadInt64(&stream.stats.Errors))

	// 2. Backpressure Queue
	fmt.Println("2. Backpressure Queue")
	queue := NewBackpressureQueue(1000, StrategyDrop)

	for i := 0; i < 100; i++ {
		queue.Enqueue(i)
	}

	stats := queue.GetStats()
	fmt.Printf("Queue stats: CurrentSize=%d, Dropped=%d, Processed=%d\n\n",
		stats["current_size"], stats["dropped"], stats["processed"])

	// 3. Stream Processor
	fmt.Println("3. Stream Processor")
	processor := NewStreamProcessor(consumer)

	processor.AddTransformer(&FilterTransformer{
		predicate: func(e *Event) bool {
			return e.EventType == "important"
		},
	})

	processor.AddTransformer(&MapTransformer{
		fn: func(e *Event) *Event {
			e.EventType = "processed-" + e.EventType
			return e
		},
	})

	event := &Event{
		ID:        "test",
		EventType: "important",
		Data:      "data",
	}
	processor.Process(event)

	fmt.Printf("Processor stats: Input=%d, Output=%d, Filtered=%d\n\n",
		atomic.LoadInt64(&processor.stats.InputEvents),
		atomic.LoadInt64(&processor.stats.OutputEvents),
		atomic.LoadInt64(&processor.stats.FilteredEvents))

	// 4. SSE Broadcaster
	fmt.Println("4. SSE Broadcaster")
	broadcaster := NewSSEBroadcaster()
	broadcaster.Start()

	sub1 := broadcaster.Subscribe("client-1")
	sub2 := broadcaster.Subscribe("client-2")

	for i := 0; i < 3; i++ {
		broadcaster.Broadcast(&Event{
			ID:        fmt.Sprintf("sse-%d", i),
			EventType: "update",
			Timestamp: time.Now(),
		})
	}

	broadcaster.Unsubscribe("client-1")
	broadcaster.Unsubscribe("client-2")
	broadcaster.Close()

	fmt.Printf("SSE stats: Subscribers=%d, Broadcasted=%d, Errors=%d\n\n",
		atomic.LoadInt32(&broadcaster.stats.TotalSubscribers),
		atomic.LoadInt64(&broadcaster.stats.EventsBroadcasted),
		atomic.LoadInt64(&broadcaster.stats.SubscriberErrors))

	// 5. Real-time Aggregator
	fmt.Println("5. Real-time Aggregator")
	aggregator := NewRealtimeAggregator(1 * time.Second)

	for i := 0; i < 10; i++ {
		aggregator.OnEvent(&Event{
			ID:        fmt.Sprintf("agg-%d", i),
			EventType: "metric",
		})
	}

	aggregator.window.mu.RLock()
	count := aggregator.window.Count
	aggregator.window.mu.RUnlock()

	fmt.Printf("Aggregator window: Count=%d, Duration=%v\n\n", count, aggregator.window.Duration)

	// 6. Features Summary
	fmt.Println("6. Streaming Features")
	fmt.Println("  - Event-driven streaming")
	fmt.Println("  - Backpressure handling (drop/block/adaptive)")
	fmt.Println("  - Stream transformations (map/filter)")
	fmt.Println("  - SSE broadcasting")
	fmt.Println("  - Real-time aggregation")
	fmt.Println("  - Consumer pattern")
	fmt.Println("  - Graceful shutdown")

	fmt.Println("\n=== Complete ===")
}
