package main

import "testing"

func TestAssertString(t *testing.T) {
	result := AssertString("hello")
	if result != "hello" {
		t.Error("Assertion failed")
	}
	t.Log("✓ Type assertion works!")
}

func TestSafeAssert(t *testing.T) {
	s, err := SafeAssert("test")
	if err != nil || s != "test" {
		t.Error("Safe assertion failed")
	}

	_, err = SafeAssert(42)
	if err == nil {
		t.Error("Should return error for non-string")
	}
	t.Log("✓ Safe assertion works!")
}

func TestCheckCapability(t *testing.T) {
	if !CheckCapability(MyCloser{}) {
		t.Error("Should detect Closer capability")
	}
	if CheckCapability("not a closer") {
		t.Error("Should not detect capability")
	}
	t.Log("✓ Capability check works!")
}
