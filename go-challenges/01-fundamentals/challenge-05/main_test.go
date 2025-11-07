package main

import "testing"

func TestSwapValues(t *testing.T) {
	a, b := 10, 20
	SwapValues(&a, &b)
	if a != 20 || b != 10 {
		t.Errorf("SwapValues failed: a=%d, b=%d", a, b)
	}
	t.Log("✓ SwapValues works!")
}

func TestDoubleValue(t *testing.T) {
	n := 5
	DoubleValue(&n)
	if n != 10 {
		t.Errorf("DoubleValue failed: n=%d, want 10", n)
	}
	t.Log("✓ DoubleValue works!")
}

func TestGetPointer(t *testing.T) {
	ptr := GetPointer(42)
	if ptr == nil || *ptr != 42 {
		t.Errorf("GetPointer failed")
	}
	t.Log("✓ GetPointer works!")
}
