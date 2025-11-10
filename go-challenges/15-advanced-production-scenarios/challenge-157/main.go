package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"
)

// ========== Saga Domain Models ==========

// SagaStatus represents the status of a saga
type SagaStatus string

const (
	SagaPending      SagaStatus = "PENDING"
	SagaInProgress   SagaStatus = "IN_PROGRESS"
	SagaCompleted    SagaStatus = "COMPLETED"
	SagaCompensating SagaStatus = "COMPENSATING"
	SagaFailed       SagaStatus = "FAILED"
)

// StepStatus represents the status of a saga step
type StepStatus string

const (
	StepPending      StepStatus = "PENDING"
	StepInProgress   StepStatus = "IN_PROGRESS"
	StepCompleted    StepStatus = "COMPLETED"
	StepFailed       StepStatus = "FAILED"
	StepCompensated  StepStatus = "COMPENSATED"
)

// Order represents the order being processed
type Order struct {
	OrderID      string                 `json:"order_id"`
	UserID       string                 `json:"user_id"`
	Amount       float64                `json:"amount"`
	Inventory    map[string]int         `json:"inventory"`
	CreatedAt    time.Time              `json:"created_at"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// SagaStep represents a single step in the saga
type SagaStep struct {
	StepID        string      `json:"step_id"`
	StepName      string      `json:"step_name"`
	Status        StepStatus  `json:"status"`
	Action        func(context.Context, *Order) error
	Compensation  func(context.Context, *Order) error
	RetryCount    int         `json:"retry_count"`
	MaxRetries    int         `json:"max_retries"`
	Error         string      `json:"error,omitempty"`
	CompletedAt   time.Time   `json:"completed_at,omitempty"`
}

// SagaState represents the complete state of a saga execution
type SagaState struct {
	SagaID       string                 `json:"saga_id"`
	Order        *Order                 `json:"order"`
	Status       SagaStatus             `json:"status"`
	Steps        map[string]*SagaStep   `json:"steps"`
	StepOrder    []string               `json:"step_order"`
	CurrentStep  int                    `json:"current_step"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
	AuditLog     []*AuditEntry          `json:"audit_log"`
	IdempotencyK string                 `json:"idempotency_key"`
	Mu           sync.RWMutex           `json:"-"`
}

// AuditEntry represents an entry in the saga audit log
type AuditEntry struct {
	Timestamp   time.Time              `json:"timestamp"`
	StepID      string                 `json:"step_id"`
	Action      string                 `json:"action"`
	Status      string                 `json:"status"`
	Message     string                 `json:"message"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// ========== Saga Orchestrator ==========

type SagaOrchestrator struct {
	states           map[string]*SagaState
	statesMu         sync.RWMutex
	eventBus         *EventBus
	compensationDLQ  []*FailedCompensation
	compensationMu   sync.RWMutex
	idempotencyCache map[string]*SagaState
	cacheMu          sync.RWMutex
}

type FailedCompensation struct {
	SagaID    string
	StepID    string
	Error     error
	Timestamp time.Time
	Attempts  int
}

// EventBus for choreography-based saga
type EventBus struct {
	handlers map[string][]func(*SagaEvent)
	mu       sync.RWMutex
}

type SagaEvent struct {
	EventID   string                 `json:"event_id"`
	SagaID    string                 `json:"saga_id"`
	EventType string                 `json:"event_type"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
}

// NewSagaOrchestrator creates a new saga orchestrator
func NewSagaOrchestrator() *SagaOrchestrator {
	return &SagaOrchestrator{
		states:           make(map[string]*SagaState),
		eventBus:         NewEventBus(),
		idempotencyCache: make(map[string]*SagaState),
	}
}

// NewEventBus creates a new event bus
func NewEventBus() *EventBus {
	return &EventBus{
		handlers: make(map[string][]func(*SagaEvent)),
	}
}

// Subscribe registers a handler for an event type
func (eb *EventBus) Subscribe(eventType string, handler func(*SagaEvent)) {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	eb.handlers[eventType] = append(eb.handlers[eventType], handler)
}

// Publish publishes an event to all subscribers
func (eb *EventBus) Publish(event *SagaEvent) {
	eb.mu.RLock()
	handlers := eb.handlers[event.EventType]
	eb.mu.RUnlock()

	for _, handler := range handlers {
		handler(event)
	}
}

// ========== Saga Orchestration ==========

// ExecuteSagaOrchestrated executes a saga using orchestration pattern
func (so *SagaOrchestrator) ExecuteSagaOrchestrated(ctx context.Context, order *Order, idempotencyKey string) (string, error) {
	// Check idempotency cache
	so.cacheMu.RLock()
	if cached, exists := so.idempotencyCache[idempotencyKey]; exists {
		so.cacheMu.RUnlock()
		return cached.SagaID, nil
	}
	so.cacheMu.RUnlock()

	sagaID := generateID()
	state := &SagaState{
		SagaID:       sagaID,
		Order:        order,
		Status:       SagaPending,
		Steps:        make(map[string]*SagaStep),
		StepOrder:    []string{},
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		AuditLog:     []*AuditEntry{},
		IdempotencyK: idempotencyKey,
	}

	// Define saga steps
	steps := []*SagaStep{
		{
			StepID:   "reserve_funds",
			StepName: "Reserve Funds",
			Action:   so.reserveFundsAction,
			Compensation: so.compensateReserveFunds,
			MaxRetries: 3,
		},
		{
			StepID:   "reserve_inventory",
			StepName: "Reserve Inventory",
			Action:   so.reserveInventoryAction,
			Compensation: so.compensateReserveInventory,
			MaxRetries: 3,
		},
		{
			StepID:   "create_shipment",
			StepName: "Create Shipment",
			Action:   so.createShipmentAction,
			Compensation: so.compensateCreateShipment,
			MaxRetries: 3,
		},
	}

	for _, step := range steps {
		step.Status = StepPending
		state.Steps[step.StepID] = step
		state.StepOrder = append(state.StepOrder, step.StepID)
	}

	// Store saga state
	so.statesMu.Lock()
	so.states[sagaID] = state
	so.statesMu.Unlock()

	// Store in idempotency cache
	so.cacheMu.Lock()
	so.idempotencyCache[idempotencyKey] = state
	so.cacheMu.Unlock()

	// Log saga start
	so.logAuditEntry(sagaID, "", "SAGA_STARTED", "PENDING", "Saga execution started", nil)
	state.Status = SagaInProgress

	// Execute steps
	for i, stepID := range state.StepOrder {
		state.CurrentStep = i
		step := state.Steps[stepID]

		if err := so.executeStep(ctx, sagaID, step, state); err != nil {
			// Compensation on failure
			so.logAuditEntry(sagaID, stepID, "STEP_FAILED", "FAILED", err.Error(), nil)
			state.Status = SagaCompensating
			so.compensateSaga(ctx, sagaID, state)
			state.Status = SagaFailed
			so.logAuditEntry(sagaID, "", "SAGA_FAILED", "FAILED", "Saga execution failed and compensated", nil)
			return sagaID, err
		}
	}

	state.Status = SagaCompleted
	state.UpdatedAt = time.Now()
	so.logAuditEntry(sagaID, "", "SAGA_COMPLETED", "COMPLETED", "Saga execution completed successfully", nil)

	return sagaID, nil
}

// executeStep executes a single step with retry logic
func (so *SagaOrchestrator) executeStep(ctx context.Context, sagaID string, step *SagaStep, state *SagaState) error {
	step.Status = StepInProgress

	var lastErr error
	for attempt := 0; attempt <= step.MaxRetries; attempt++ {
		step.RetryCount = attempt

		ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		err := step.Action(ctx, state.Order)
		cancel()

		if err == nil {
			step.Status = StepCompleted
			step.CompletedAt = time.Now()
			so.logAuditEntry(sagaID, step.StepID, "STEP_COMPLETED", "COMPLETED",
				fmt.Sprintf("Step completed after %d attempts", attempt+1), nil)
			return nil
		}

		lastErr = err
		step.Error = err.Error()

		if attempt < step.MaxRetries {
			backoff := time.Duration((1<<uint(attempt))*100) * time.Millisecond
			time.Sleep(backoff)
			so.logAuditEntry(sagaID, step.StepID, "STEP_RETRY", "IN_PROGRESS",
				fmt.Sprintf("Retrying step (attempt %d/%d)", attempt+1, step.MaxRetries), nil)
		}
	}

	step.Status = StepFailed
	return lastErr
}

// compensateSaga compensates failed saga in reverse order
func (so *SagaOrchestrator) compensateSaga(ctx context.Context, sagaID string, state *SagaState) {
	so.logAuditEntry(sagaID, "", "COMPENSATION_STARTED", "COMPENSATING", "Starting saga compensation", nil)

	// Compensate in reverse order
	for i := len(state.StepOrder) - 1; i >= 0; i-- {
		stepID := state.StepOrder[i]
		step := state.Steps[stepID]

		if step.Status != StepCompleted {
			continue
		}

		so.logAuditEntry(sagaID, stepID, "COMPENSATION_STARTED", "COMPENSATING", "Starting step compensation", nil)

		// Execute compensation with retry
		var attempts int
		for {
			ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
			err := step.Compensation(ctx, state.Order)
			cancel()

			attempts++

			if err == nil {
				step.Status = StepCompensated
				so.logAuditEntry(sagaID, stepID, "COMPENSATION_COMPLETED", "COMPENSATED",
					fmt.Sprintf("Step compensated after %d attempts", attempts), nil)
				break
			}

			if attempts >= 3 {
				so.compensationDLQ = append(so.compensationDLQ, &FailedCompensation{
					SagaID:    sagaID,
					StepID:    stepID,
					Error:     err,
					Timestamp: time.Now(),
					Attempts:  attempts,
				})
				so.logAuditEntry(sagaID, stepID, "COMPENSATION_FAILED", "FAILED",
					fmt.Sprintf("Compensation failed after %d attempts, moved to DLQ", attempts),
					map[string]interface{}{"error": err.Error()})
				break
			}

			time.Sleep(time.Duration(attempts*100) * time.Millisecond)
		}
	}

	so.logAuditEntry(sagaID, "", "COMPENSATION_COMPLETED", "COMPENSATING", "Saga compensation completed", nil)
}

// ========== Saga Actions (Simulated) ==========

func (so *SagaOrchestrator) reserveFundsAction(ctx context.Context, order *Order) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	// Simulate funds reservation
	if order.Amount <= 0 {
		return errors.New("invalid amount")
	}
	return nil
}

func (so *SagaOrchestrator) compensateReserveFunds(ctx context.Context, order *Order) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	// Simulate funds release
	return nil
}

func (so *SagaOrchestrator) reserveInventoryAction(ctx context.Context, order *Order) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	for _, qty := range order.Inventory {
		if qty <= 0 {
			return errors.New("insufficient inventory")
		}
	}
	return nil
}

func (so *SagaOrchestrator) compensateReserveInventory(ctx context.Context, order *Order) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	// Release inventory
	return nil
}

func (so *SagaOrchestrator) createShipmentAction(ctx context.Context, order *Order) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	// Create shipment
	return nil
}

func (so *SagaOrchestrator) compensateCreateShipment(ctx context.Context, order *Order) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	// Cancel shipment
	return nil
}

// ========== Audit Logging ==========

func (so *SagaOrchestrator) logAuditEntry(sagaID, stepID, action, status, message string, metadata map[string]interface{}) {
	so.statesMu.Lock()
	defer so.statesMu.Unlock()

	if state, exists := so.states[sagaID]; exists {
		entry := &AuditEntry{
			Timestamp: time.Now(),
			StepID:    stepID,
			Action:    action,
			Status:    status,
			Message:   message,
			Metadata:  metadata,
		}
		state.AuditLog = append(state.AuditLog, entry)
		state.UpdatedAt = time.Now()
	}
}

// ========== Saga Choreography (Event-Driven) ==========

// SagaChoreographyHandler coordinates saga steps via events
type SagaChoreographyHandler struct {
	orchestrator *SagaOrchestrator
	eventBus    *EventBus
}

// NewSagaChoreographyHandler creates a new choreography handler
func NewSagaChoreographyHandler(orchestrator *SagaOrchestrator) *SagaChoreographyHandler {
	handler := &SagaChoreographyHandler{
		orchestrator: orchestrator,
		eventBus:     orchestrator.eventBus,
	}

	// Subscribe to events
	handler.eventBus.Subscribe("OrderCreated", handler.handleOrderCreated)
	handler.eventBus.Subscribe("FundsReserved", handler.handleFundsReserved)
	handler.eventBus.Subscribe("InventoryReserved", handler.handleInventoryReserved)
	handler.eventBus.Subscribe("StepFailed", handler.handleStepFailed)

	return handler
}

func (h *SagaChoreographyHandler) handleOrderCreated(event *SagaEvent) {
	// Process order and publish next event
	h.eventBus.Publish(&SagaEvent{
		EventID:   generateID(),
		SagaID:    event.SagaID,
		EventType: "ReserveFundsRequested",
		Timestamp: time.Now(),
		Data:      event.Data,
	})
}

func (h *SagaChoreographyHandler) handleFundsReserved(event *SagaEvent) {
	h.eventBus.Publish(&SagaEvent{
		EventID:   generateID(),
		SagaID:    event.SagaID,
		EventType: "ReserveInventoryRequested",
		Timestamp: time.Now(),
		Data:      event.Data,
	})
}

func (h *SagaChoreographyHandler) handleInventoryReserved(event *SagaEvent) {
	h.eventBus.Publish(&SagaEvent{
		EventID:   generateID(),
		SagaID:    event.SagaID,
		EventType: "SagaCompleted",
		Timestamp: time.Now(),
		Data:      event.Data,
	})
}

func (h *SagaChoreographyHandler) handleStepFailed(event *SagaEvent) {
	h.eventBus.Publish(&SagaEvent{
		EventID:   generateID(),
		SagaID:    event.SagaID,
		EventType: "CompensationRequested",
		Timestamp: time.Now(),
		Data:      event.Data,
	})
}

// ========== Query Methods ==========

// GetSagaState retrieves the state of a saga
func (so *SagaOrchestrator) GetSagaState(sagaID string) *SagaState {
	so.statesMu.RLock()
	defer so.statesMu.RUnlock()
	return so.states[sagaID]
}

// GetAuditLog retrieves the audit log for a saga
func (so *SagaOrchestrator) GetAuditLog(sagaID string) []*AuditEntry {
	so.statesMu.RLock()
	defer so.statesMu.RUnlock()

	if state, exists := so.states[sagaID]; exists {
		entries := make([]*AuditEntry, len(state.AuditLog))
		copy(entries, state.AuditLog)
		return entries
	}
	return nil
}

// GetFailedCompensations retrieves failed compensations
func (so *SagaOrchestrator) GetFailedCompensations() []*FailedCompensation {
	so.compensationMu.RLock()
	defer so.compensationMu.RUnlock()
	return so.compensationDLQ
}

// SerializeSagaState serializes saga state to JSON
func (so *SagaOrchestrator) SerializeSagaState(sagaID string) (string, error) {
	state := so.GetSagaState(sagaID)
	if state == nil {
		return "", errors.New("saga not found")
	}

	data, err := json.MarshalIndent(state, "", "  ")
	return string(data), err
}

// Helper function
func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

func main() {
	// Example saga execution
	orchestrator := NewSagaOrchestrator()
	order := &Order{
		OrderID:   "order-1",
		UserID:    "user-1",
		Amount:    100.0,
		Inventory: map[string]int{"item1": 5},
		CreatedAt: time.Now(),
		Metadata:  make(map[string]interface{}),
	}

	ctx := context.Background()
	_, _ = orchestrator.ExecuteSagaOrchestrated(ctx, order, "idempotency-1")
}
