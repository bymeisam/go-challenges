# Solution for Challenge 05: Pointers

```go
package main

func SwapValues(a, b *int) {
	temp := *a
	*a = *b
	*b = temp
}

func DoubleValue(n *int) {
	*n = *n * 2
}

func GetPointer(x int) *int {
	return &x  // Returns pointer to x
}
```

## Key Points
- `&x` gets the address of x
- `*ptr` dereferences a pointer
- Pointers allow functions to modify values
- Use pointers for large structs to avoid copying

## JS Comparison
In JavaScript, objects are passed by reference automatically. In Go, you explicitly use pointers.
