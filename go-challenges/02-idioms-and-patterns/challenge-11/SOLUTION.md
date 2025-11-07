# Solution for Challenge 11: Defer Statement

```go
package main

import "os"

func OpenAndProcess(filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()  // Cleanup guaranteed!

	_, err = f.WriteString("test")
	return err
}

func MeasureTime(operation string) func() {
	// start := time.Now()
	return func() {
		// fmt.Printf("%s took %v\n", operation, time.Since(start))
	}
}

func DeferOrder() []int {
	result := []int{}
	defer func() { result = append(result, 1) }()
	defer func() { result = append(result, 2) }()
	defer func() { result = append(result, 3) }()
	return result  // Returns [3, 2, 1] - LIFO!
}
```

## Key: Defer executes in LIFO (Last In, First Out) order!
