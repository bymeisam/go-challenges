package main

import "testing"

func TestSafeCounter(t *testing.T) {
	c := SafeCounter()
	if c == nil {
		t.Fatal("SafeCounter() should not return nil")
	}
	if c.Count != 0 {
		t.Errorf("Counter.Count should be 0 (zero value), got %d", c.Count)
	}
	t.Log("✓ SafeCounter works!")
}

func TestIsZeroValue(t *testing.T) {
	tests := []struct {
		name     string
		value    interface{}
		expected bool
	}{
		{"zero int", 0, true},
		{"non-zero int", 42, false},
		{"empty string", "", true},
		{"non-empty string", "hello", false},
		{"false bool", false, true},
		{"true bool", true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsZeroValue(tt.value)
			if result != tt.expected {
				t.Errorf("IsZeroValue(%v) = %v; want %v", tt.value, result, tt.expected)
			}
		})
	}
	t.Log("✓ IsZeroValue works!")
}

func TestInitializeSlice(t *testing.T) {
	slice := InitializeSlice()
	if slice == nil {
		t.Error("InitializeSlice() should return non-nil slice")
	}
	if len(slice) != 0 {
		t.Errorf("Slice length should be 0, got %d", len(slice))
	}
	t.Log("✓ InitializeSlice works!")
}

func TestSafeStringOperation(t *testing.T) {
	// Test with nil
	result := SafeStringOperation(nil)
	if result != "default" {
		t.Errorf("SafeStringOperation(nil) = %q; want \"default\"", result)
	}

	// Test with value
	val := "test"
	result = SafeStringOperation(&val)
	if result != "test" {
		t.Errorf("SafeStringOperation(&\"test\") = %q; want \"test\"", result)
	}

	t.Log("✓ SafeStringOperation works!")
}
