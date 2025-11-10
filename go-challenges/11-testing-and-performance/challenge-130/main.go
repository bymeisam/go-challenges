package main

import (
	"strings"
)

// ConcatWithPlus concatenates strings using the + operator
func ConcatWithPlus(strs []string) string {
	result := ""
	for _, s := range strs {
		result += s
	}
	return result
}

// ConcatWithBuilder concatenates strings using strings.Builder
func ConcatWithBuilder(strs []string) string {
	var builder strings.Builder
	for _, s := range strs {
		builder.WriteString(s)
	}
	return builder.String()
}

// ConcatWithJoin concatenates strings using strings.Join
func ConcatWithJoin(strs []string) string {
	return strings.Join(strs, "")
}

// FibonacciRecursive calculates Fibonacci number recursively
func FibonacciRecursive(n int) int {
	if n <= 1 {
		return n
	}
	return FibonacciRecursive(n-1) + FibonacciRecursive(n-2)
}

// FibonacciIterative calculates Fibonacci number iteratively
func FibonacciIterative(n int) int {
	if n <= 1 {
		return n
	}

	a, b := 0, 1
	for i := 2; i <= n; i++ {
		a, b = b, a+b
	}
	return b
}

// LinearSearch performs linear search on an array
func LinearSearch(arr []int, target int) int {
	for i, v := range arr {
		if v == target {
			return i
		}
	}
	return -1
}

// BinarySearch performs binary search on a sorted array
func BinarySearch(arr []int, target int) int {
	left, right := 0, len(arr)-1

	for left <= right {
		mid := left + (right-left)/2

		if arr[mid] == target {
			return mid
		} else if arr[mid] < target {
			left = mid + 1
		} else {
			right = mid - 1
		}
	}

	return -1
}

// MapWithMake creates a map with make and size hint
func MapWithMake(size int) map[int]int {
	m := make(map[int]int, size)
	for i := 0; i < size; i++ {
		m[i] = i * 2
	}
	return m
}

// MapWithoutMake creates a map without size hint
func MapWithoutMake(size int) map[int]int {
	m := make(map[int]int)
	for i := 0; i < size; i++ {
		m[i] = i * 2
	}
	return m
}

func main() {
	// Example usage
}
