package main

import "testing"

func TestDescribe(t *testing.T) {
	if Describe(42) != "Integer: 42" {
		t.Error("Should describe int")
	}
	if Describe("test") != "String: test" {
		t.Error("Should describe string")
	}
	t.Log("✓ Type switch works!")
}
