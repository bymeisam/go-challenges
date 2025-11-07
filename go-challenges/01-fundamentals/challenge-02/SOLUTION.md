# Solution for Challenge 02: Slices and Arrays

```go
package main

// CreateArray creates and returns an array of 5 integers: [1, 2, 3, 4, 5]
func CreateArray() [5]int {
	return [5]int{1, 2, 3, 4, 5}
}

// CreateSlice creates and returns a slice of integers: [1, 2, 3, 4, 5]
func CreateSlice() []int {
	return []int{1, 2, 3, 4, 5}
}

// AppendToSlice appends a value to the slice and returns the result
func AppendToSlice(slice []int, value int) []int {
	return append(slice, value)
}

// SliceCapacity returns the length and capacity of a slice
func SliceCapacity(slice []int) (int, int) {
	return len(slice), cap(slice)
}

// SlicePortions returns two slices:
// - first: elements from index 0 to 2 (exclusive)
// - second: elements from index 2 to end
func SlicePortions(slice []int) ([]int, []int) {
	return slice[0:2], slice[2:]
}
```

## 📖 Detailed Explanation

### Function 1: CreateArray

```go
func CreateArray() [5]int {
	return [5]int{1, 2, 3, 4, 5}
}
```

**What's happening:**
- `[5]int` specifies an array of exactly 5 integers
- The size is part of the type
- Arrays are **value types** (copied when passed around)

**Alternate syntax:**
```go
// Explicit declaration
var arr [5]int = [5]int{1, 2, 3, 4, 5}

// Let Go infer the size
arr := [...]int{1, 2, 3, 4, 5}  // Size inferred as 5
```

**When to use arrays:**
- Rarely! Most Go code uses slices
- Only when you need fixed-size collections
- Example: cryptographic keys, fixed buffers

### Function 2: CreateSlice

```go
func CreateSlice() []int {
	return []int{1, 2, 3, 4, 5}
}
```

**What's happening:**
- `[]int` (no size) defines a slice
- Slices are **reference types** (point to underlying array)
- Most common collection type in Go

**Other ways to create slices:**
```go
// Using make (preallocate with length and capacity)
slice := make([]int, 5)      // length 5, capacity 5, zero-valued
slice := make([]int, 5, 10)  // length 5, capacity 10

// From an array
arr := [5]int{1, 2, 3, 4, 5}
slice := arr[:]  // Create slice from entire array

// Empty slice
var slice []int  // nil slice
slice := []int{} // empty but non-nil
```

**JS/TS Comparison:**
```javascript
// JavaScript - everything is dynamic (like Go slices)
const arr = [1, 2, 3, 4, 5];

// TypeScript
const arr: number[] = [1, 2, 3, 4, 5];
```

### Function 3: AppendToSlice

```go
func AppendToSlice(slice []int, value int) []int {
	return append(slice, value)
}
```

**What's happening:**
- `append()` adds elements to a slice
- Returns a new slice (may point to new array if capacity exceeded)
- **Must assign the result back**

**How append works internally:**
1. Check if capacity is sufficient
2. If yes: add element to underlying array, increment length
3. If no: allocate new, larger array, copy elements, add new element

**JS/TS Comparison:**
```javascript
// JavaScript - mutates array
arr.push(value);

// Go - returns new slice
slice = append(slice, value)

// To append multiple values:
slice = append(slice, 1, 2, 3)

// To append another slice:
slice = append(slice, anotherSlice...)
```

**Common mistake:**
```go
// WRONG - discarding return value
append(slice, value)

// CORRECT
slice = append(slice, value)
```

### Function 4: SliceCapacity

```go
func SliceCapacity(slice []int) (int, int) {
	return len(slice), cap(slice)
}
```

**What's happening:**
- `len()` returns current number of elements
- `cap()` returns size of underlying array
- Capacity ≥ length always

**Example:**
```go
slice := make([]int, 3, 5)
// Length: 3 (can access indices 0, 1, 2)
// Capacity: 5 (underlying array has space for 5 elements)

slice = append(slice, 10)  // Length becomes 4
slice = append(slice, 20)  // Length becomes 5
slice = append(slice, 30)  // Length becomes 6, NEW array allocated!
```

**Growth strategy:**
When capacity is exceeded, Go typically doubles the capacity (up to a point), then grows by ~25%.

**JS Comparison:**
```javascript
// JavaScript has no capacity concept
arr.length  // Only length

// Arrays grow automatically without explicit capacity management
```

### Function 5: SlicePortions

```go
func SlicePortions(slice []int) ([]int, []int) {
	return slice[0:2], slice[2:]
}
```

**What's happening:**
- `slice[start:end]` creates a new slice view
- `start` is inclusive, `end` is exclusive
- Both slices share the underlying array!

**Slice notation:**
```go
slice[start:end]  // from start to end-1
slice[start:]     // from start to end
slice[:end]       // from beginning to end-1
slice[:]          // entire slice (creates new view)
```

**Examples:**
```go
s := []int{1, 2, 3, 4, 5}

s[1:3]   // [2, 3]
s[:2]    // [1, 2]
s[2:]    // [3, 4, 5]
s[:]     // [1, 2, 3, 4, 5]
```

**Important: Shared underlying array**
```go
original := []int{1, 2, 3, 4, 5}
portion := original[1:3]  // [2, 3]

portion[0] = 99  // Modifies underlying array!
// Now original is [1, 99, 3, 4, 5]
```

**JS Comparison:**
```javascript
// JavaScript slice() creates a COPY
const original = [1, 2, 3, 4, 5];
const portion = original.slice(1, 3);  // [2, 3]

portion[0] = 99;
// original is still [1, 2, 3, 4, 5] - NOT modified!
```

**To create a true copy in Go:**
```go
original := []int{1, 2, 3, 4, 5}
copy_slice := make([]int, len(original))
copy(copy_slice, original)  // Built-in copy function
```

## 🎓 Key Takeaways

1. **Arrays are rarely used** - Use slices instead
2. **Slices are references** - Multiple slices can share data
3. **Always assign append result** - `slice = append(slice, val)`
4. **Understand len vs cap** - Important for performance
5. **Slice notation creates views** - Not copies (unlike JS)

## 🚀 Common Patterns

### Preallocation for performance
```go
// Bad - many reallocations
var slice []int
for i := 0; i < 1000; i++ {
	slice = append(slice, i)
}

// Good - preallocate capacity
slice := make([]int, 0, 1000)
for i := 0; i < 1000; i++ {
	slice = append(slice, i)
}
```

### Removing an element
```go
// Remove element at index i
slice = append(slice[:i], slice[i+1:]...)
```

### Filtering
```go
result := make([]int, 0, len(slice))
for _, v := range slice {
	if condition(v) {
		result = append(result, v)
	}
}
```

## 💡 Pro Tips

1. Use `make()` with capacity when you know the size upfront
2. Be careful with slice sharing - modify copies if needed
3. Slices are more efficient than arrays for most use cases
4. `nil` slice vs empty slice: both have length 0, but nil is preferred

Great work! You now understand one of Go's most important data structures! 🎉
