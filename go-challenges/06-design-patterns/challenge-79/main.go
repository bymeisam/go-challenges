package main

import (
	"errors"
	"sync"
)

type CircuitBreaker struct {
	mu           sync.Mutex
	failures     int
	threshold    int
	isOpen       bool
}

func NewCircuitBreaker(threshold int) *CircuitBreaker {
	return &CircuitBreaker{
		threshold: threshold,
	}
}

func (cb *CircuitBreaker) Call(fn func() error) error {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	
	if cb.isOpen {
		return errors.New("circuit breaker is open")
	}
	
	err := fn()
	if err != nil {
		cb.failures++
		if cb.failures >= cb.threshold {
			cb.isOpen = true
		}
		return err
	}
	
	cb.failures = 0
	return nil
}

func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.failures = 0
	cb.isOpen = false
}

func main() {}
