# Solution for Challenge 31: Basic Interfaces

```go
package main

import "math"

type Shape interface {
	Area() float64
	Perimeter() float64
}

type Rectangle struct {
	Width  float64
	Height float64
}

func (r Rectangle) Area() float64 {
	return r.Width * r.Height
}

func (r Rectangle) Perimeter() float64 {
	return 2 * (r.Width + r.Height)
}

type Circle struct {
	Radius float64
}

func (c Circle) Area() float64 {
	return math.Pi * c.Radius * c.Radius
}

func (c Circle) Perimeter() float64 {
	return 2 * math.Pi * c.Radius
}

func TotalArea(shapes []Shape) float64 {
	total := 0.0
	for _, shape := range shapes {
		total += shape.Area()
	}
	return total
}
```

## Key Points

1. **Implicit implementation**: No `implements` keyword - if a type has the methods, it satisfies the interface
2. **Polymorphism**: `[]Shape` can hold both Rectangle and Circle
3. **Small interfaces**: Go favors small, focused interfaces (often just 1-2 methods)
4. **Duck typing**: "If it walks like a duck and quacks like a duck, it's a duck"

## Why This Matters

Interfaces enable loose coupling and testability. You can swap implementations without changing code that uses the interface.
