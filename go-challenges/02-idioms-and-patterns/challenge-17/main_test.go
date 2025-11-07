package main

import "testing"

func TestGenericMax(t *testing.T) {
	if GenericMax(5, 3) != 5 {
		t.Error("GenericMax(5, 3) should be 5")
	}
	if GenericMax("b", "a") != "b" {
		t.Error("GenericMax(\"b\", \"a\") should be \"b\"")
	}
	t.Log("✓ GenericMax works!")
}

func TestContains(t *testing.T) {
	if !Contains([]int{1, 2, 3}, 2) {
		t.Error("Contains should find 2")
	}
	if Contains([]string{"a", "b"}, "c") {
		t.Error("Contains should not find \"c\"")
	}
	t.Log("✓ Contains works!")
}
