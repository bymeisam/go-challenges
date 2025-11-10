package main

import "testing"

func TestDecorator(t *testing.T) {
	coffee := SimpleCoffee{}
	
	if coffee.Cost() != 5 {
		t.Error("Simple coffee should cost 5")
	}
	
	withMilk := MilkDecorator{coffee: coffee}
	if withMilk.Cost() != 7 {
		t.Errorf("Coffee with milk should cost 7, got %d", withMilk.Cost())
	}
	
	withMilkAndSugar := SugarDecorator{coffee: withMilk}
	if withMilkAndSugar.Cost() != 8 {
		t.Errorf("Coffee with milk and sugar should cost 8, got %d", withMilkAndSugar.Cost())
	}
	
	desc := withMilkAndSugar.Description()
	if desc != "Simple coffee, milk, sugar" {
		t.Errorf("Description wrong: %s", desc)
	}
	
	t.Log("✓ Decorator pattern works!")
}
