package main

import (
	"errors"
	"testing"
)

func TestProcessInput(t *testing.T) {
	if err := ProcessInput("test"); err != nil {
		t.Error("Should not error with valid input")
	}
	if !errors.Is(ProcessInput(""), ErrInvalid) {
		t.Error("Should return ErrInvalid for empty string")
	}
	t.Log("✓ Errors package works!")
}
