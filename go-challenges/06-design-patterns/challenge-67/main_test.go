package main

import "testing"

func TestFactory(t *testing.T) {
	car := VehicleFactory("car")
	if car == nil || car.Drive() != "driving a car" {
		t.Error("Factory should create car")
	}
	
	bike := VehicleFactory("bike")
	if bike == nil || bike.Drive() != "riding a bike" {
		t.Error("Factory should create bike")
	}
	
	t.Log("✓ Factory pattern works!")
}
