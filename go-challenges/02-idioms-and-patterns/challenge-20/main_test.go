package main

import (
	"testing"
)

func TestPersonString(t *testing.T) {
	p := Person{Name: "Alice", Age: 30}
	expected := "Alice (30 years old)"
	if p.String() != expected {
		t.Errorf("Person.String() = %q; want %q", p.String(), expected)
	}
	t.Log("✓ Stringer interface works!")
}
