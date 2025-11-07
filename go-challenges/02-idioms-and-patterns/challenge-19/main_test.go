package main

import "testing"

func TestGetString(t *testing.T) {
	s, ok := GetString("hello")
	if !ok || s != "hello" {
		t.Error("GetString should return (\"hello\", true)")
	}
	_, ok = GetString(42)
	if ok {
		t.Error("GetString(42) should return ok=false")
	}
	t.Log("✓ GetString works!")
}

func TestTypeSwitch(t *testing.T) {
	if TypeSwitch("test") != "string" {
		t.Error("TypeSwitch(\"test\") should return \"string\"")
	}
	if TypeSwitch(42) != "int" {
		t.Error("TypeSwitch(42) should return \"int\"")
	}
	t.Log("✓ TypeSwitch works!")
}
