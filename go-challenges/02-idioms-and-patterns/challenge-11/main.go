package main

import (
	"os"
)

// OpenAndProcess opens a file and processes it, using defer to close
func OpenAndProcess(filename string) error {
	// TODO: Open file, defer close, process
	// Create a temp file for testing
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	// TODO: Add defer f.Close() here

	// Simulate processing
	_, err = f.WriteString("test")
	return err
}

// MeasureTime returns a cleanup function that should be deferred
func MeasureTime(operation string) func() {
	// TODO: Return a function that will be called via defer
	// For now, just return a no-op function
	return func() {}
}

// DeferOrder demonstrates LIFO execution order of defers
func DeferOrder() []int {
	result := []int{}
	// TODO: Use 3 defers to append 1, 2, 3 to result
	// They should execute in reverse order (LIFO)
	// Hint: defer func() { result = append(result, n) }()
	return result
}

func main() {}
