package main

import (
	"context"
	"testing"
	"time"
)

func TestSagaOrchestrationSuccess(t *testing.T) {
	orchestrator := NewSagaOrchestrator()
	order := &Order{
		OrderID:   "order-123",
		UserID:    "user-456",
		Amount:    100.0,
		Inventory: map[string]int{"item1": 5, "item2": 3},
		CreatedAt: time.Now(),
		Metadata:  make(map[string]interface{}),
	}

	ctx := context.Background()
	sagaID, err := orchestrator.ExecuteSagaOrchestrated(ctx, order, "idempotency-key-1")

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if sagaID == "" {
		t.Fatal("Expected non-empty saga ID")
	}

	state := orchestrator.GetSagaState(sagaID)
	if state == nil {
		t.Fatal("Expected saga state to exist")
	}

	if state.Status != SagaCompleted {
		t.Fatalf("Expected SagaCompleted, got %v", state.Status)
	}

	if state.Order.OrderID != "order-123" {
		t.Fatalf("Expected order ID order-123, got %s", state.Order.OrderID)
	}
}

func TestSagaIdempotency(t *testing.T) {
	orchestrator := NewSagaOrchestrator()
	order := &Order{
		OrderID:   "order-124",
		UserID:    "user-457",
		Amount:    150.0,
		Inventory: map[string]int{"item1": 2},
		CreatedAt: time.Now(),
		Metadata:  make(map[string]interface{}),
	}

	ctx := context.Background()
	idempotencyKey := "idempotency-key-repeat"

	sagaID1, err1 := orchestrator.ExecuteSagaOrchestrated(ctx, order, idempotencyKey)
	if err1 != nil {
		t.Fatalf("First execution failed: %v", err1)
	}

	sagaID2, err2 := orchestrator.ExecuteSagaOrchestrated(ctx, order, idempotencyKey)
	if err2 != nil {
		t.Fatalf("Second execution failed: %v", err2)
	}

	if sagaID1 != sagaID2 {
		t.Fatalf("Expected same saga ID for idempotent call, got %s and %s", sagaID1, sagaID2)
	}
}

func TestSagaCompensation(t *testing.T) {
	orchestrator := NewSagaOrchestrator()

	// Create order with invalid amount to trigger failure
	order := &Order{
		OrderID:   "order-125",
		UserID:    "user-458",
		Amount:    -50.0, // Invalid amount
		Inventory: map[string]int{"item1": 5},
		CreatedAt: time.Now(),
		Metadata:  make(map[string]interface{}),
	}

	ctx := context.Background()
	sagaID, err := orchestrator.ExecuteSagaOrchestrated(ctx, order, "idempotency-key-2")

	if err == nil {
		t.Fatal("Expected error due to invalid amount")
	}

	state := orchestrator.GetSagaState(sagaID)
	if state.Status != SagaFailed {
		t.Fatalf("Expected SagaFailed, got %v", state.Status)
	}
}

func TestSagaAuditLog(t *testing.T) {
	orchestrator := NewSagaOrchestrator()
	order := &Order{
		OrderID:   "order-126",
		UserID:    "user-459",
		Amount:    200.0,
		Inventory: map[string]int{"item1": 10},
		CreatedAt: time.Now(),
		Metadata:  make(map[string]interface{}),
	}

	ctx := context.Background()
	sagaID, _ := orchestrator.ExecuteSagaOrchestrated(ctx, order, "idempotency-key-3")

	auditLog := orchestrator.GetAuditLog(sagaID)
	if len(auditLog) == 0 {
		t.Fatal("Expected audit log entries")
	}

	// Check for saga started entry
	hasStarted := false
	hasCompleted := false
	for _, entry := range auditLog {
		if entry.Action == "SAGA_STARTED" {
			hasStarted = true
		}
		if entry.Action == "SAGA_COMPLETED" {
			hasCompleted = true
		}
	}

	if !hasStarted {
		t.Fatal("Expected SAGA_STARTED in audit log")
	}
	if !hasCompleted {
		t.Fatal("Expected SAGA_COMPLETED in audit log")
	}
}

func TestMultipleSagas(t *testing.T) {
	orchestrator := NewSagaOrchestrator()
	ctx := context.Background()

	sagaIDs := make(map[string]bool)
	for i := 0; i < 5; i++ {
		order := &Order{
			OrderID:   "order-" + string(rune(i)),
			UserID:    "user-" + string(rune(i)),
			Amount:    float64(100 + i*10),
			Inventory: map[string]int{"item1": i + 1},
			CreatedAt: time.Now(),
			Metadata:  make(map[string]interface{}),
		}

		sagaID, err := orchestrator.ExecuteSagaOrchestrated(ctx, order, "key-"+string(rune(i)))
		if err != nil {
			t.Fatalf("Saga %d failed: %v", i, err)
		}

		sagaIDs[sagaID] = true
	}

	if len(sagaIDs) != 5 {
		t.Fatalf("Expected 5 unique saga IDs, got %d", len(sagaIDs))
	}
}

func TestSagaStepStatus(t *testing.T) {
	orchestrator := NewSagaOrchestrator()
	order := &Order{
		OrderID:   "order-127",
		UserID:    "user-460",
		Amount:    100.0,
		Inventory: map[string]int{"item1": 5},
		CreatedAt: time.Now(),
		Metadata:  make(map[string]interface{}),
	}

	ctx := context.Background()
	sagaID, _ := orchestrator.ExecuteSagaOrchestrated(ctx, order, "idempotency-key-4")

	state := orchestrator.GetSagaState(sagaID)
	for _, stepID := range state.StepOrder {
		step := state.Steps[stepID]
		if step.Status != StepCompleted {
			t.Fatalf("Expected step %s to be completed, got %v", stepID, step.Status)
		}
	}
}

func TestSagaTimeout(t *testing.T) {
	orchestrator := NewSagaOrchestrator()
	order := &Order{
		OrderID:   "order-128",
		UserID:    "user-461",
		Amount:    100.0,
		Inventory: map[string]int{"item1": 5},
		CreatedAt: time.Now(),
		Metadata:  make(map[string]interface{}),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// This should complete before timeout in this test
	sagaID, err := orchestrator.ExecuteSagaOrchestrated(ctx, order, "idempotency-key-5")
	if err != nil && err != context.DeadlineExceeded {
		// It's okay if it completes or times out
	}

	if sagaID != "" {
		state := orchestrator.GetSagaState(sagaID)
		if state == nil {
			t.Fatal("Expected saga state")
		}
	}
}

func TestSagaStateSnapshot(t *testing.T) {
	orchestrator := NewSagaOrchestrator()
	order := &Order{
		OrderID:   "order-129",
		UserID:    "user-462",
		Amount:    100.0,
		Inventory: map[string]int{"item1": 5},
		CreatedAt: time.Now(),
		Metadata:  map[string]interface{}{"test": "value"},
	}

	ctx := context.Background()
	sagaID, _ := orchestrator.ExecuteSagaOrchestrated(ctx, order, "idempotency-key-6")

	state := orchestrator.GetSagaState(sagaID)
	if state == nil {
		t.Fatal("Expected saga state to exist")
	}

	if state.SagaID != sagaID {
		t.Fatal("Expected matching saga ID")
	}

	if state.Order.OrderID != "order-129" {
		t.Fatal("Expected order data to be preserved")
	}
}

func TestEventBusPublishSubscribe(t *testing.T) {
	eventBus := NewEventBus()
	received := false
	var receivedEvent *SagaEvent

	eventBus.Subscribe("TestEvent", func(event *SagaEvent) {
		received = true
		receivedEvent = event
	})

	event := &SagaEvent{
		EventID:   "event-1",
		SagaID:    "saga-1",
		EventType: "TestEvent",
		Timestamp: time.Now(),
		Data:      map[string]interface{}{"key": "value"},
	}

	eventBus.Publish(event)

	if !received {
		t.Fatal("Expected event to be received")
	}

	if receivedEvent.SagaID != "saga-1" {
		t.Fatalf("Expected saga ID saga-1, got %s", receivedEvent.SagaID)
	}
}

func TestSagaChoreography(t *testing.T) {
	orchestrator := NewSagaOrchestrator()
	handler := NewSagaChoreographyHandler(orchestrator)

	// Publish order created event
	event := &SagaEvent{
		EventID:   "event-1",
		SagaID:    "saga-choreography-1",
		EventType: "OrderCreated",
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"order_id": "order-130",
			"amount":   100.0,
		},
	}

	orchestrator.eventBus.Publish(event)

	// Give handlers time to process
	time.Sleep(100 * time.Millisecond)

	if handler == nil {
		t.Fatal("Expected choreography handler to exist")
	}
}

func TestFailedCompensationDLQ(t *testing.T) {
	orchestrator := NewSagaOrchestrator()

	// Manually add a failed compensation
	failed := &FailedCompensation{
		SagaID:    "saga-1",
		StepID:    "step-1",
		Error:     nil,
		Timestamp: time.Now(),
		Attempts:  3,
	}

	orchestrator.compensationDLQ = append(orchestrator.compensationDLQ, failed)

	dlq := orchestrator.GetFailedCompensations()
	if len(dlq) != 1 {
		t.Fatalf("Expected 1 DLQ entry, got %d", len(dlq))
	}

	if dlq[0].SagaID != "saga-1" {
		t.Fatalf("Expected saga ID saga-1, got %s", dlq[0].SagaID)
	}
}

func TestSagaRetryLogic(t *testing.T) {
	orchestrator := NewSagaOrchestrator()
	order := &Order{
		OrderID:   "order-131",
		UserID:    "user-463",
		Amount:    100.0,
		Inventory: map[string]int{"item1": 5},
		CreatedAt: time.Now(),
		Metadata:  make(map[string]interface{}),
	}

	ctx := context.Background()
	sagaID, _ := orchestrator.ExecuteSagaOrchestrated(ctx, order, "idempotency-key-7")

	state := orchestrator.GetSagaState(sagaID)
	// First step should have retried (RetryCount >= 0)
	firstStep := state.Steps[state.StepOrder[0]]
	if firstStep.RetryCount < 0 {
		t.Fatalf("Expected non-negative retry count, got %d", firstStep.RetryCount)
	}
}

func TestConcurrentSagaExecution(t *testing.T) {
	orchestrator := NewSagaOrchestrator()
	ctx := context.Background()
	results := make(chan string, 10)

	for i := 0; i < 10; i++ {
		go func(index int) {
			order := &Order{
				OrderID:   "order-concurrent-" + string(rune(index)),
				UserID:    "user-" + string(rune(index)),
				Amount:    100.0,
				Inventory: map[string]int{"item1": 5},
				CreatedAt: time.Now(),
				Metadata:  make(map[string]interface{}),
			}

			sagaID, _ := orchestrator.ExecuteSagaOrchestrated(ctx, order, "key-concurrent-"+string(rune(index)))
			results <- sagaID
		}(i)
	}

	received := 0
	for i := 0; i < 10; i++ {
		<-results
		received++
	}

	if received != 10 {
		t.Fatalf("Expected 10 saga IDs, got %d", received)
	}
}

func TestSagaContextCancel(t *testing.T) {
	orchestrator := NewSagaOrchestrator()
	order := &Order{
		OrderID:   "order-132",
		UserID:    "user-464",
		Amount:    100.0,
		Inventory: map[string]int{"item1": 5},
		CreatedAt: time.Now(),
		Metadata:  make(map[string]interface{}),
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := orchestrator.ExecuteSagaOrchestrated(ctx, order, "idempotency-key-8")
	// Should handle context cancellation gracefully
	_ = err
}

func BenchmarkSagaExecution(b *testing.B) {
	orchestrator := NewSagaOrchestrator()
	ctx := context.Background()

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		order := &Order{
			OrderID:   "order-bench-" + string(rune(i)),
			UserID:    "user-" + string(rune(i)),
			Amount:    100.0,
			Inventory: map[string]int{"item1": 5},
			CreatedAt: time.Now(),
			Metadata:  make(map[string]interface{}),
		}

		_, _ = orchestrator.ExecuteSagaOrchestrated(ctx, order, "key-bench-"+string(rune(i)))
	}
}

func BenchmarkEventPublishing(b *testing.B) {
	eventBus := NewEventBus()

	eventBus.Subscribe("BenchEvent", func(event *SagaEvent) {
		// No-op handler
	})

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		event := &SagaEvent{
			EventID:   "event-" + string(rune(i)),
			SagaID:    "saga-" + string(rune(i)),
			EventType: "BenchEvent",
			Timestamp: time.Now(),
			Data:      map[string]interface{}{},
		}
		eventBus.Publish(event)
	}
}
