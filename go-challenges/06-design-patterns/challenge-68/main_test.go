package main

import "testing"

func TestBuilder(t *testing.T) {
	house := NewHouseBuilder().
		Windows(10).
		Doors(2).
		Floors(2).
		Garage(true).
		Build()
	
	if house.Windows != 10 || house.Doors != 2 {
		t.Error("Builder failed")
	}
	
	if !house.HasGarage {
		t.Error("House should have garage")
	}
	
	t.Log("✓ Builder pattern works!")
}
