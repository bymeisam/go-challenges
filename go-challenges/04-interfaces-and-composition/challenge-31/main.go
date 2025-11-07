package main

import "math"

// TODO: Define Shape interface with Area() and Perimeter() methods

// Rectangle represents a rectangle shape
type Rectangle struct {
	Width  float64
	Height float64
}

// TODO: Implement Area() for Rectangle
func (r Rectangle) Area() float64 {
	return 0 // TODO: Calculate area
}

// TODO: Implement Perimeter() for Rectangle
func (r Rectangle) Perimeter() float64 {
	return 0 // TODO: Calculate perimeter
}

// Circle represents a circle shape
type Circle struct {
	Radius float64
}

// TODO: Implement Area() for Circle
func (c Circle) Area() float64 {
	return 0 // TODO: Calculate area using math.Pi
}

// TODO: Implement Perimeter() for Circle
func (c Circle) Perimeter() float64 {
	return 0 // TODO: Calculate circumference
}

// TotalArea calculates the total area of all shapes
func TotalArea(shapes []Shape) float64 {
	// TODO: Sum up areas of all shapes
	return 0
}

func main() {}
