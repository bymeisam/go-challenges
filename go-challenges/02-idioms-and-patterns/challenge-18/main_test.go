package main

import "testing"

func TestEmbedding(t *testing.T) {
	d := Dog{Animal: Animal{Name: "Buddy"}, Breed: "Golden"}
	if d.Name != "Buddy" {
		t.Error("Dog should have access to embedded Animal.Name")
	}
	if d.Speak() != "Woof!" {
		t.Error("Dog.Speak() should be overridden")
	}
	t.Log("✓ Embedding works!")
}
