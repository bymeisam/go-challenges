package main

import "testing"

func TestBufferedChannel(t *testing.T) {
	ch := BufferedChannel(2)
	
	// Can send 2 without blocking
	if !TrySend(ch, "a") {
		t.Error("Should send to buffered channel")
	}
	if !TrySend(ch, "b") {
		t.Error("Should send second item")
	}
	
	// Third send should fail (buffer full)
	if TrySend(ch, "c") {
		t.Error("Should not send when buffer full")
	}
	
	t.Log("✓ Buffered channels work!")
}
