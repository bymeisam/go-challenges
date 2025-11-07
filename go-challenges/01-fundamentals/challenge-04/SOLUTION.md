# Solution for Challenge 04: Structs

```go
package main

import "fmt"

// Person struct definition
type Person struct {
	Name string
	Age  int
}

func NewPerson(name string, age int) Person {
	return Person{Name: name, Age: age}
}

func UpdateAge(p *Person, newAge int) {
	p.Age = newAge
}

func GetInfo(p Person) string {
	return fmt.Sprintf("Name: %s, Age: %d", p.Name, p.Age)
}
```

## Key Points
- Structs are Go's way to define custom types
- Use pointers (`*Person`) to modify structs
- No classes or constructors - use factory functions like `NewPerson`
- Exported fields start with capital letters

## JS/TS Comparison
```javascript
// TypeScript class
class Person {
  constructor(public name: string, public age: number) {}
}

// Go struct (no methods yet)
type Person struct {
  Name string
  Age  int
}
```
