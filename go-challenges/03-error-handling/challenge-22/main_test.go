package main

import (
	"errors"
	"os"
	"testing"
)

func TestReadFileWrapping(t *testing.T) {
	err := ReadFile("nonexistent.txt")
	if err == nil {
		t.Fatal("Should return error for nonexistent file")
	}
	if !errors.Is(err, os.ErrNotExist) {
		t.Error("Should wrap os.ErrNotExist")
	}
	t.Log("✓ Error wrapping works!")
}

func TestErrorsIs(t *testing.T) {
	err := FindItem(0)
	if !errors.Is(err, ErrNotFound) {
		t.Error("errors.Is should detect wrapped ErrNotFound")
	}
	t.Log("✓ errors.Is works!")
}
