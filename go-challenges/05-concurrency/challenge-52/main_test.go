package main

import "testing"

func TestNonBlockingSend(t *testing.T) {
	ch := make(chan int, 1)
	
	if !NonBlockingSend(ch, 42) {
		t.Error("Should send to buffered channel")
	}
	
	if NonBlockingSend(ch, 43) {
		t.Error("Should not block when full")
	}
	
	t.Log("✓ Non-blocking send works!")
}

func TestNonBlockingReceive(t *testing.T) {
	ch := make(chan int, 1)
	ch <- 42
	
	val, ok := NonBlockingReceive(ch)
	if !ok || val != 42 {
		t.Error("Should receive value")
	}
	
	_, ok = NonBlockingReceive(ch)
	if ok {
		t.Error("Should not block on empty channel")
	}
	
	t.Log("✓ Non-blocking receive works!")
}
