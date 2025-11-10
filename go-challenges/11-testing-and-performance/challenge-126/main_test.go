package main

import "testing"

func TestAdd(t *testing.T) {
	tests := []struct {
		name     string
		a, b     int
		expected int
	}{
		{"positive numbers", 2, 3, 5},
		{"negative numbers", -1, -1, -2},
		{"zero", 0, 0, 0},
		{"positive and negative", 10, -5, 5},
		{"large numbers", 1000000, 2000000, 3000000},
	}

	for _, tt := range tests {
		result := Add(tt.a, tt.b)
		if result != tt.expected {
			t.Errorf("%s: Add(%d, %d) = %d; want %d",
				tt.name, tt.a, tt.b, result, tt.expected)
		}
	}

	t.Log("✓ All Add tests passed!")
}

func TestIsPalindrome(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"simple palindrome", "racecar", true},
		{"palindrome with capital", "Racecar", true},
		{"palindrome with spaces", "A man a plan a canal Panama", true},
		{"not a palindrome", "hello", false},
		{"single character", "a", true},
		{"empty string", "", true},
		{"palindrome phrase", "Was it a car or a cat I saw", true},
		{"numbers", "12321", true},
		{"non-palindrome numbers", "12345", false},
	}

	for _, tt := range tests {
		result := IsPalindrome(tt.input)
		if result != tt.expected {
			t.Errorf("%s: IsPalindrome(%q) = %v; want %v",
				tt.name, tt.input, result, tt.expected)
		}
	}

	t.Log("✓ All IsPalindrome tests passed!")
}

func TestValidateEmail(t *testing.T) {
	tests := []struct {
		name     string
		email    string
		expected bool
	}{
		{"valid email", "user@example.com", true},
		{"valid email with subdomain", "user@mail.example.com", true},
		{"no @ symbol", "userexample.com", false},
		{"multiple @ symbols", "user@@example.com", false},
		{"no domain", "user@", false},
		{"no username", "@example.com", false},
		{"no dot after @", "user@example", false},
		{"dot at the end", "user@example.", false},
		{"valid complex email", "john.doe@company.co.uk", true},
		{"empty string", "", false},
	}

	for _, tt := range tests {
		result := ValidateEmail(tt.email)
		if result != tt.expected {
			t.Errorf("%s: ValidateEmail(%q) = %v; want %v",
				tt.name, tt.email, result, tt.expected)
		}
	}

	t.Log("✓ All ValidateEmail tests passed!")
}

func TestFibonacci(t *testing.T) {
	tests := []struct {
		name     string
		n        int
		expected int
	}{
		{"F(0)", 0, 0},
		{"F(1)", 1, 1},
		{"F(2)", 2, 1},
		{"F(3)", 3, 2},
		{"F(4)", 4, 3},
		{"F(5)", 5, 5},
		{"F(6)", 6, 8},
		{"F(7)", 7, 13},
		{"F(10)", 10, 55},
		{"F(15)", 15, 610},
	}

	for _, tt := range tests {
		result := Fibonacci(tt.n)
		if result != tt.expected {
			t.Errorf("%s: Fibonacci(%d) = %d; want %d",
				tt.name, tt.n, result, tt.expected)
		}
	}

	t.Log("✓ All Fibonacci tests passed!")
}
