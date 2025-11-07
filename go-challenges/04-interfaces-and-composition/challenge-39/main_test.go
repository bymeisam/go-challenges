package main

import "testing"

func TestEmbedding(t *testing.T) {
	e := NewExtended(1, "test", "extra")

	if e.ID != 1 {
		t.Error("Should access embedded field directly")
	}
	if e.GetID() != 1 {
		t.Error("Should call embedded method")
	}
	if e.Extra != "extra" {
		t.Error("Extended field failed")
	}
	t.Log("✓ Struct embedding works!")
}
