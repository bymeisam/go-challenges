package main

import "testing"

func TestRectangleArea(t *testing.T) {
	r := Rectangle{Width: 5, Height: 3}
	if r.Area() != 15 {
		t.Errorf("Area() = %v, want 15", r.Area())
	}
	t.Log("✓ Area works!")
}

func TestRectanglePerimeter(t *testing.T) {
	r := Rectangle{Width: 5, Height: 3}
	if r.Perimeter() != 16 {
		t.Errorf("Perimeter() = %v, want 16", r.Perimeter())
	}
	t.Log("✓ Perimeter works!")
}

func TestRectangleScale(t *testing.T) {
	r := Rectangle{Width: 2, Height: 3}
	r.Scale(2)
	if r.Width != 4 || r.Height != 6 {
		t.Errorf("Scale failed: %+v", r)
	}
	t.Log("✓ Scale works!")
}
