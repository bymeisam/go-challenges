# Hints for Challenge 02: Slices and Arrays

## Hint 1: CreateArray

<details>
<summary>Click to reveal hint</summary>

Arrays in Go are declared with a fixed size in square brackets:

```go
arr := [5]int{1, 2, 3, 4, 5}
```

The size `[5]` is part of the type. An array of 5 ints is a different type from an array of 10 ints!

</details>

## Hint 2: CreateSlice

<details>
<summary>Click to reveal hint</summary>

Slices are declared without a size:

```go
slice := []int{1, 2, 3, 4, 5}
```

Notice: `[]int` (no size) vs `[5]int` (with size)

You can also use `make()`:
```go
slice := make([]int, 5)  // length 5, zero-valued
```

</details>

## Hint 3: AppendToSlice

<details>
<summary>Click to reveal hint</summary>

Use the built-in `append()` function:

```go
result := append(slice, value)
return result
```

**Important:** `append()` returns a new slice. Always capture the return value!

In JavaScript:
```javascript
arr.push(value)  // mutates arr
```

In Go:
```go
slice = append(slice, value)  // returns new slice
```

</details>

## Hint 4: SliceCapacity

<details>
<summary>Click to reveal hint</summary>

Go has two built-in functions for slices:
- `len(slice)` - returns the number of elements
- `cap(slice)` - returns the capacity (size of underlying array)

```go
length := len(slice)
capacity := cap(slice)
return length, capacity
```

</details>

## Hint 5: SlicePortions

<details>
<summary>Click to reveal hint</summary>

Go uses slice notation `[start:end]` (similar to Python):

```go
first := slice[0:2]   // elements at index 0 and 1
second := slice[2:]   // elements from index 2 to end
```

**JavaScript comparison:**
```javascript
const first = arr.slice(0, 2);    // Similar!
const second = arr.slice(2);
```

**Key difference:** Go's slice notation creates a view of the underlying array (not a copy), while JS's `.slice()` creates a copy.

</details>

## Still Stuck?

Check `SOLUTION.md` for the complete solution!

## Quick Reference

```go
// Array
var arr [5]int = [5]int{1, 2, 3, 4, 5}

// Slice
var slice []int = []int{1, 2, 3, 4, 5}

// Append
slice = append(slice, 6)

// Length and capacity
len(slice)  // number of elements
cap(slice)  // underlying array size

// Slicing
slice[start:end]  // from start to end-1
slice[start:]     // from start to end of slice
slice[:end]       // from beginning to end-1
```
