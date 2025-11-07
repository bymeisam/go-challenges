# Solution for Challenge 07: Methods

```go
package main

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

func (r *Rectangle) Scale(factor float64) {
	r.Width *= factor
	r.Height *= factor
}
```

## Key Points
- Methods are functions with a receiver
- Value receiver: `(r Rectangle)` - read-only
- Pointer receiver: `(r *Rectangle)` - can modify
- Similar to JS class methods, but defined outside the struct
