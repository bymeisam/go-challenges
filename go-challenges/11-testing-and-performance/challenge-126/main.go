package main

import "strings"

// Add returns the sum of two integers
func Add(a, b int) int {
	return a + b
}

// IsPalindrome checks if a string is a palindrome
// ignoring case and spaces
func IsPalindrome(s string) bool {
	// Remove spaces and convert to lowercase
	s = strings.ToLower(strings.ReplaceAll(s, " ", ""))

	// Check if string reads the same forwards and backwards
	for i := 0; i < len(s)/2; i++ {
		if s[i] != s[len(s)-1-i] {
			return false
		}
	}
	return true
}

// ValidateEmail validates an email address
// Must contain exactly one '@' symbol
// Must have at least one character before and after '@'
// Must have at least one '.' after '@'
func ValidateEmail(email string) bool {
	// Count '@' symbols
	atCount := strings.Count(email, "@")
	if atCount != 1 {
		return false
	}

	// Find '@' position
	atIndex := strings.Index(email, "@")
	if atIndex == 0 || atIndex == len(email)-1 {
		return false
	}

	// Check for '.' after '@'
	afterAt := email[atIndex+1:]
	if !strings.Contains(afterAt, ".") {
		return false
	}

	// Check '.' is not the last character
	if strings.HasSuffix(email, ".") {
		return false
	}

	return true
}

// Fibonacci returns the nth Fibonacci number (0-indexed)
func Fibonacci(n int) int {
	if n <= 1 {
		return n
	}

	a, b := 0, 1
	for i := 2; i <= n; i++ {
		a, b = b, a+b
	}
	return b
}

func main() {
	// Example usage
}
