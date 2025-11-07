package main

import (
	"testing"
	"time"
)

func TestSelect(t *testing.T) {
	ch1 := make(chan string, 1)
	ch2 := make(chan string, 1)
	
	ch1 <- "hello"
	result := SelectFromMultiple(ch1, ch2)
	
	if result != "ch1: hello" {
		t.Errorf("Select failed: %s", result)
	}
	t.Log("✓ Select works!")
}

func TestSelectTimeout(t *testing.T) {
	ch := make(chan string)
	
	_, ok := SelectWithTimeout(ch, 10*time.Millisecond)
	if ok {
		t.Error("Should timeout")
	}
	
	ch2 := make(chan string, 1)
	ch2 <- "fast"
	msg, ok := SelectWithTimeout(ch2, 100*time.Millisecond)
	if !ok || msg != "fast" {
		t.Error("Should receive message")
	}
	
	t.Log("✓ Select with timeout works!")
}
