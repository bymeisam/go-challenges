# Challenge 03: Maps

**Difficulty:** ⭐ Easy | **Topic:** Fundamentals | **Time:** 15-20 min

## 🎯 Learning Goals
- Understand Go maps (similar to JS objects/Maps)
- Learn map operations: create, add, delete, check existence
- Practice map iteration

## 📝 Description
Maps in Go are similar to JavaScript objects or ES6 Maps. Key differences:
- Keys must be comparable types (numbers, strings, etc.)
- Access to non-existent keys returns zero value (not `undefined`)
- Must use `make()` or literal syntax to initialize

## 🔨 Your Task

### 1. `CreateMap() map[string]int`
Create and return a map with: `{"one": 1, "two": 2, "three": 3}`

### 2. `AddToMap(m map[string]int, key string, value int)`
Add a key-value pair to the map (maps are reference types, so no return needed).

### 3. `GetFromMap(m map[string]int, key string) (int, bool)`
Get value for key. Return value and a boolean indicating if key exists.

### 4. `DeleteFromMap(m map[string]int, key string)`
Delete the key from the map.

### 5. `MapKeys(m map[string]int) []string`
Return all keys as a slice.

## 💡 JS vs Go

| Operation | JavaScript | Go |
|-----------|-----------|-----|
| Create | `{}` or `new Map()` | `make(map[K]V)` or `map[K]V{}` |
| Add | `obj[k] = v` | `m[k] = v` |
| Get | `obj[k]` | `v, ok := m[k]` |
| Delete | `delete obj[k]` | `delete(m, k)` |
| Check | `k in obj` | `_, ok := m[k]` |

## 🧪 Testing
```bash
cd go-challenges/01-fundamentals/challenge-03
go test -v
```

## 📚 Resources
- [Go by Example: Maps](https://gobyexample.com/maps)

---
**Ready?** Open `main.go` and start coding!
