package main

import "testing"

func TestBehaviorComposition(t *testing.T) {
	dog := Dog{}
	car := Car{}

	// Both can move
	if dog.Move() != "walking" {
		t.Error("Dog should walk")
	}
	if car.Move() != "walking" {
		t.Error("Car should move")
	}

	// Only dog can speak
	if dog.Speak() != "talking" {
		t.Error("Dog should talk")
	}

	t.Log("✓ Behavior composition works!")
}
