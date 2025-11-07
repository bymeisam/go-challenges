# Challenge 04: Structs

**Difficulty:** ⭐⭐ Medium | **Topic:** Fundamentals | **Time:** 20 min

## 🎯 Learning Goals
- Create and use structs (Go's way of defining custom types)
- Understand struct field access
- Learn struct initialization patterns
- Compare to JS/TS objects and classes

## 🔨 Tasks
1. Define a `Person` struct with `Name` (string) and `Age` (int)
2. `NewPerson(name string, age int) Person` - constructor function
3. `UpdateAge(p *Person, newAge int)` - modify struct via pointer
4. `GetInfo(p Person) string` - return formatted string

## 💡 JS vs Go
JS classes/objects vs Go structs (no methods yet, just data)

## 🧪 Testing
```bash
go test -v
```
