package main

import "testing"

func TestChannels(t *testing.T) {
	sum := Pipeline()
	if sum != 15 {
		t.Errorf("Pipeline sum = %d; want 15", sum)
	}
	t.Log("✓ Channels work!")
}
