# Challenge 02: Slices and Arrays

**Difficulty:** ⭐ Easy
**Topic:** Fundamentals
**Estimated Time:** 20-25 minutes

## 🎯 Learning Goals

- Understand the difference between arrays and slices
- Learn slice operations (append, len, cap)
- Practice slice manipulation
- Understand that slices are references (unlike arrays)

## 📝 Description

Coming from JavaScript, you're familiar with arrays. Go has **two** concepts:

1. **Arrays**: Fixed-size, rarely used directly
2. **Slices**: Dynamic-size, similar to JS arrays

**Key difference:** In JavaScript, all "arrays" are actually dynamic. In Go, you'll almost always use slices, not arrays.

## 🔨 Your Task

Implement the following functions in `main.go`:

### 1. `CreateArray() [5]int`

Create and return an array of exactly 5 integers with values `[1, 2, 3, 4, 5]`.

**Note:** Arrays have fixed size specified in the type: `[5]int`

### 2. `CreateSlice() []int`

Create and return a slice containing `[1, 2, 3, 4, 5]`.

**Note:** Slices don't specify size in the type: `[]int`

### 3. `AppendToSlice(slice []int, value int) []int`

Append a value to the slice and return the result.

**Hint:** Use the built-in `append()` function.

### 4. `SliceCapacity(slice []int) (int, int)`

Return the length and capacity of the slice.

**Hint:** Use `len()` and `cap()` built-in functions.

### 5. `SlicePortions(slice []int) ([]int, []int)`

Return two slices:
- First: elements from index 0 to 2 (exclusive)
- Second: elements from index 2 to end

**Hint:** Use slice operators: `slice[start:end]`

## 🧪 Testing

```bash
cd go-challenges/01-fundamentals/challenge-02
go test -v
```

## 💡 Key Differences from JavaScript

| Concept | JavaScript | Go |
|---------|-----------|-----|
| Array type | `number[]` (dynamic) | `[]int` (slice) or `[5]int` (array) |
| Fixed size | No concept | Arrays: `[5]int` |
| Dynamic size | Default | Slices: `[]int` |
| Append | `arr.push(x)` | `append(slice, x)` |
| Length | `arr.length` | `len(slice)` |
| Slicing | `arr.slice(0, 2)` | `slice[0:2]` |
| Capacity | No concept | `cap(slice)` |

## 📚 Important Concepts

### Arrays vs Slices

```go
// Array - fixed size, value type
var arr [3]int = [3]int{1, 2, 3}

// Slice - dynamic size, reference type
var slice []int = []int{1, 2, 3}
```

### Slice Internals

A slice is a reference to an underlying array with three components:
- Pointer to the array
- Length (current number of elements)
- Capacity (size of underlying array)

### Append Behavior

```go
slice := []int{1, 2, 3}
slice = append(slice, 4)  // May allocate new array if capacity exceeded
```

**Important:** Always assign the result of `append()` back to the slice!

## 📚 Resources

- [Go by Example: Slices](https://gobyexample.com/slices)
- [Go by Example: Arrays](https://gobyexample.com/arrays)
- [Go Slices: usage and internals](https://go.dev/blog/slices-intro)

## ✨ Bonus Challenge (Optional)

Try creating a slice with `make()`:
```go
slice := make([]int, length, capacity)
```

What's the difference between `make()` and literal initialization?

---

**Ready?** Open `main.go` and start coding! Run `go test -v` when you're done.
