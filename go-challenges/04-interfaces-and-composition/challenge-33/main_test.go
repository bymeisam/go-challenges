package main

import "testing"

func TestGetType(t *testing.T) {
	if GetType(42) != "int" {
		t.Error("Should detect int")
	}
	if GetType("test") != "string" {
		t.Error("Should detect string")
	}
	t.Log("✓ Empty interface works!")
}

func TestConvertToInt(t *testing.T) {
	val, ok := ConvertToInt(42)
	if !ok || val != 42 {
		t.Error("Should convert int")
	}
	_, ok = ConvertToInt("not an int")
	if ok {
		t.Error("Should fail for non-int")
	}
	t.Log("✓ Type assertion works!")
}
