# Solution for Challenge 09: Zero Values

```go
package main

type Counter struct {
	Count int
}

func SafeCounter() *Counter {
	return &Counter{}  // Count is automatically 0
}

func IsZeroValue(v interface{}) bool {
	switch val := v.(type) {
	case int:
		return val == 0
	case string:
		return val == ""
	case bool:
		return val == false
	default:
		return false
	}
}

func InitializeSlice() []int {
	return []int{}  // Or: make([]int, 0)
}

func SafeStringOperation(s *string) string {
	if s == nil {
		return "default"
	}
	return *s
}
```

## Key Points

1. **Zero values make code safer**: No need to check for undefined
2. **Structs**: All fields get zero values automatically
3. **Nil vs empty**: `var slice []int` is nil, `[]int{}` is empty but non-nil
4. **Pointer safety**: Always check for nil before dereferencing

## JS vs Go

```javascript
// JavaScript
let obj = {};
console.log(obj.count);  // undefined - need to check!

// Go
var c Counter
fmt.Println(c.Count)  // 0 - always safe!
```

This is a fundamental Go idiom - design types so their zero value is useful!
