package main

import (
	"math"
	"testing"
)

func TestRectangle(t *testing.T) {
	r := Rectangle{Width: 5, Height: 3}

	area := r.Area()
	if area != 15 {
		t.Errorf("Rectangle.Area() = %v; want 15", area)
	}

	perimeter := r.Perimeter()
	if perimeter != 16 {
		t.Errorf("Rectangle.Perimeter() = %v; want 16", perimeter)
	}

	t.Log("✓ Rectangle implements Shape!")
}

func TestCircle(t *testing.T) {
	c := Circle{Radius: 2}

	expectedArea := math.Pi * 4
	area := c.Area()
	if math.Abs(area-expectedArea) > 0.01 {
		t.Errorf("Circle.Area() = %v; want %v", area, expectedArea)
	}

	expectedPerimeter := 2 * math.Pi * 2
	perimeter := c.Perimeter()
	if math.Abs(perimeter-expectedPerimeter) > 0.01 {
		t.Errorf("Circle.Perimeter() = %v; want %v", perimeter, expectedPerimeter)
	}

	t.Log("✓ Circle implements Shape!")
}

func TestTotalArea(t *testing.T) {
	shapes := []Shape{
		Rectangle{Width: 5, Height: 3},
		Circle{Radius: 2},
	}

	total := TotalArea(shapes)
	expectedTotal := 15 + math.Pi*4

	if math.Abs(total-expectedTotal) > 0.01 {
		t.Errorf("TotalArea() = %v; want %v", total, expectedTotal)
	}

	t.Log("✓ Interface polymorphism works!")
}

func TestInterfaceSatisfaction(t *testing.T) {
	// Test that types satisfy the interface
	var _ Shape = Rectangle{}
	var _ Shape = Circle{}

	t.Log("✓ Both types satisfy Shape interface!")
}
