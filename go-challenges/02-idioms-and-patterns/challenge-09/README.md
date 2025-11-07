# Challenge 09: Zero Values

**Difficulty:** ⭐⭐ Medium | **Topic:** Idioms & Patterns | **Time:** 20 min

## 🎯 Learning Goals
- Understand Go's zero value philosophy
- Learn when to leverage zero values vs explicit initialization
- Compare with JavaScript's undefined/null behavior

## 📝 Description

In JavaScript, uninitialized variables are `undefined`. Go takes a different approach: **every type has a zero value** that's immediately usable.

**Zero values:**
- `0` for numeric types
- `false` for booleans
- `""` for strings
- `nil` for pointers, slices, maps, channels, functions, interfaces

This makes Go code more predictable and reduces nil/undefined bugs!

## 🔨 Your Task

### 1. `SafeCounter() *Counter`

Create a `Counter` struct with an `int` field called `Count`. Write a constructor that relies on zero values (no explicit initialization needed).

### 2. `IsZeroValue(v interface{}) bool`

Check if a value is its zero value. Handle int, string, bool types.

### 3. `InitializeSlice() []int`

Return a non-nil empty slice (different from nil slice). Use `make()` or literal.

### 4. `SafeStringOperation(s *string) string`

Accept a pointer to string. If nil, return "default". Otherwise return the value.

## 💡 Why This Matters

```go
// JavaScript - risky!
let x;
console.log(x.length);  // TypeError: Cannot read property 'length' of undefined

// Go - safe!
var s string
fmt.Println(len(s))  // 0 (zero value "" has length 0)
```

## 🧪 Testing
```bash
cd go-challenges/02-idioms-and-patterns/challenge-09
go test -v
```

## 📚 Resources
- [Effective Go: Zero Values](https://go.dev/doc/effective_go#allocation_new)

---
**Ready?** Open `main.go` and start coding!
