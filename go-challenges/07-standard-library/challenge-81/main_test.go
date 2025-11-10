package main

import (
	"reflect"
	"testing"
)

func TestStrings(t *testing.T) {
	if !ContainsWord("hello world", "world") {
		t.Error("Contains failed")
	}
	
	words := SplitSentence("hello world test")
	if !reflect.DeepEqual(words, []string{"hello", "world", "test"}) {
		t.Error("Split failed")
	}
	
	joined := JoinWords([]string{"a", "b", "c"})
	if joined != "a b c" {
		t.Error("Join failed")
	}
	
	if TrimSpaces("  hello  ") != "hello" {
		t.Error("Trim failed")
	}
	
	if ReplaceAll("hello hello", "hello", "hi") != "hi hi" {
		t.Error("Replace failed")
	}
	
	if ToUpperCase("hello") != "HELLO" {
		t.Error("ToUpper failed")
	}
	
	t.Log("✓ strings package works!")
}
