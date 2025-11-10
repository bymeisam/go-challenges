package main

import (
	"reflect"
	"testing"
)

func TestRegexp(t *testing.T) {
	if !MatchEmail("test@example.com") {
		t.Error("Should match valid email")
	}
	
	if MatchEmail("invalid-email") {
		t.Error("Should not match invalid email")
	}
	
	numbers := FindAllNumbers("abc 123 def 456")
	if !reflect.DeepEqual(numbers, []string{"123", "456"}) {
		t.Errorf("FindAllNumbers failed: %v", numbers)
	}
	
	result := ReplaceDigits("abc123")
	if result != "abcXXX" {
		t.Errorf("ReplaceDigits failed: %s", result)
	}
	
	t.Log("✓ regexp package works!")
}
