# Challenge 31: Basic Interfaces

**Difficulty:** ⭐⭐ Medium | **Topic:** Interfaces & Composition | **Time:** 25 min

## 🎯 Learning Goals
- Understand Go interfaces (implicit implementation)
- Learn interface-based polymorphism
- Compare with TypeScript interfaces

## 📝 Description

Unlike TypeScript where you explicitly implement interfaces, Go uses **implicit implementation**. If a type has all the methods of an interface, it automatically satisfies that interface!

```go
// No "implements" keyword needed!
type Writer interface {
    Write([]byte) (int, error)
}
```

## 🔨 Your Task

### 1. Define `Shape` Interface
Create an interface with two methods:
- `Area() float64`
- `Perimeter() float64`

### 2. Implement `Rectangle` Type
Create a Rectangle struct and implement the Shape interface.

### 3. Implement `Circle` Type
Create a Circle struct and implement the Shape interface.

### 4. `TotalArea(shapes []Shape) float64`
Accept a slice of Shape interfaces and return total area.

## 💡 JS/TS vs Go

```typescript
// TypeScript - explicit
interface Shape {
    area(): number;
}
class Circle implements Shape {  // "implements" keyword
    area(): number { return 3.14 * this.r * this.r; }
}

// Go - implicit
type Shape interface {
    Area() float64
}
type Circle struct { R float64 }
func (c Circle) Area() float64 { return 3.14 * c.R * c.R }
// Circle automatically satisfies Shape!
```

## 🧪 Testing
```bash
cd go-challenges/04-interfaces-and-composition/challenge-31
go test -v
```

## 📚 Resources
- [Effective Go: Interfaces](https://go.dev/doc/effective_go#interfaces)

---
**Ready?** Open `main.go` and start coding!
