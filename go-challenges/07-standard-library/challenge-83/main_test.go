package main

import (
	"strings"
	"testing"
)

func TestFmt(t *testing.T) {
	p := Person{Name: "Alice", Age: 30}
	
	result := FormatPerson(p)
	if result != "Alice is 30 years old" {
		t.Errorf("FormatPerson failed: %s", result)
	}
	
	result = FormatWithTypes("test", 42, 3.14159)
	if result != "String: test, Int: 42, Float: 3.14" {
		t.Errorf("FormatWithTypes failed: %s", result)
	}
	
	result = FormatStruct(p)
	if !strings.Contains(result, "Alice") || !strings.Contains(result, "30") {
		t.Error("FormatStruct failed")
	}
	
	t.Log("✓ fmt package works!")
}
