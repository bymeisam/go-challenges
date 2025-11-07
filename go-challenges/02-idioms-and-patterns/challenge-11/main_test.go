package main

import (
	"os"
	"reflect"
	"testing"
)

func TestOpenAndProcess(t *testing.T) {
	filename := "test_defer.txt"
	defer os.Remove(filename)

	err := OpenAndProcess(filename)
	if err != nil {
		t.Fatalf("OpenAndProcess failed: %v", err)
	}
	t.Log("✓ OpenAndProcess works!")
}

func TestDeferOrder(t *testing.T) {
	result := DeferOrder()
	expected := []int{3, 2, 1}  // LIFO order
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("DeferOrder() = %v; want %v (LIFO)", result, expected)
	}
	t.Log("✓ DeferOrder works!")
}
