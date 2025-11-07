package main

import "errors"

// Divide returns a/b and an error if b is zero
func Divide(a, b float64) (float64, error) {
	// TODO: Check if b==0, return error. Otherwise return a/b
	if b == 0 {
		return 0, errors.New("division by zero")
	}
	return a / b, nil
}

// MakeMultiplier returns a closure that multiplies by factor
func MakeMultiplier(factor int) func(int) int {
	// TODO: Return a function that multiplies its input by factor
	return func(n int) int {
		return n * factor
	}
}

// ApplyOperation applies op to each element in nums
func ApplyOperation(nums []int, op func(int) int) []int {
	// TODO: Apply op to each element and return new slice
	result := make([]int, len(nums))
	for i, n := range nums {
		result[i] = op(n)
	}
	return result
}

func main() {}
