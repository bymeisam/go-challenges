# Challenge 157: Distributed Transactions - Saga Pattern

**Difficulty:** ⭐⭐⭐⭐⭐ Expert | **Time:** 90 min

Implement comprehensive distributed transaction management using the Saga pattern with both orchestration and choreography approaches.

## Learning Objectives
- Saga pattern orchestration (centralized coordinator)
- Saga pattern choreography (event-driven)
- Compensation logic and rollback strategies
- State management and idempotency
- Handling partial failures
- Retry mechanisms with exponential backoff
- Transaction log and audit trail
- Dead letter queues

## Advanced Topics
1. **Orchestration**: CancelOrderSaga, ReserveFundsStep, ReserveInventoryStep
2. **Choreography**: Event-driven saga using event bus
3. **Compensation**: Automatic rollback on failures
4. **State Management**: Saga state machine and persistence
5. **Idempotency**: Ensuring repeated operations are safe
6. **Observability**: Complete audit trail and monitoring

## Architecture Patterns
- Saga orchestrator pattern
- Event sourcing
- Compensating transactions
- Idempotent operations
- State machines

## Tasks
1. Implement Saga orchestrator with multiple steps
2. Create choreography-based saga with event handlers
3. Implement compensation logic for rollback
4. Add retry mechanism with exponential backoff
5. Create idempotency token support
6. Implement saga state persistence
7. Add comprehensive audit logging
8. Handle distributed timeout scenarios

```bash
go test -v
go test -bench=. -benchmem
```

## Production Considerations
- Saga state must be persisted for recovery
- Compensation must always eventually succeed
- Idempotency is critical for retries
- Monitor compensating transaction failures (DLQ)
- Use saga IDs for tracing across services
- Implement health checks for saga coordinator
- Set appropriate timeouts for each step
- Log every state transition for debugging
