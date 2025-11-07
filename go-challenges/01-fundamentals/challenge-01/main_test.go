package main

import (
	"testing"
)

func TestDeclareVariables(t *testing.T) {
	i, s, b, f := DeclareVariables()

	if i != 0 {
		t.Errorf("Expected int zero value 0, got %d", i)
	}
	if s != "" {
		t.Errorf("Expected string zero value \"\", got %q", s)
	}
	if b != false {
		t.Errorf("Expected bool zero value false, got %v", b)
	}
	if f != 0.0 {
		t.Errorf("Expected float64 zero value 0.0, got %f", f)
	}

	t.Log("✓ All zero values are correct!")
}

func TestUseShortDeclaration(t *testing.T) {
	tests := []struct {
		name     string
		x        int
		y        int
		expected int
	}{
		{"positive numbers", 5, 3, 8},
		{"negative numbers", -5, -3, -8},
		{"mixed signs", 10, -3, 7},
		{"zeros", 0, 0, 0},
		{"large numbers", 1000, 2000, 3000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := UseShortDeclaration(tt.x, tt.y)
			if result != tt.expected {
				t.Errorf("UseShortDeclaration(%d, %d) = %d; want %d",
					tt.x, tt.y, result, tt.expected)
			}
		})
	}

	t.Log("✓ Short declaration works correctly!")
}

func TestTypeConversion(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    int
		shouldError bool
	}{
		{"valid positive", "42", 42, false},
		{"valid negative", "-42", -42, false},
		{"valid zero", "0", 0, false},
		{"invalid letters", "abc", 0, true},
		{"invalid mixed", "12abc", 0, true},
		{"empty string", "", 0, true},
		{"large number", "123456", 123456, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := TypeConversion(tt.input)

			if tt.shouldError {
				if err == nil {
					t.Errorf("TypeConversion(%q) should return error, got nil", tt.input)
				}
			} else {
				if err != nil {
					t.Errorf("TypeConversion(%q) unexpected error: %v", tt.input, err)
				}
				if result != tt.expected {
					t.Errorf("TypeConversion(%q) = %d; want %d", tt.input, result, tt.expected)
				}
			}
		})
	}

	t.Log("✓ Type conversion works correctly!")
}

func TestConstantsExample(t *testing.T) {
	name, version := ConstantsExample()

	if name != "GoChallenge" {
		t.Errorf("Expected appName to be \"GoChallenge\", got %q", name)
	}
	if version != 1 {
		t.Errorf("Expected version to be 1, got %d", version)
	}

	t.Log("✓ Constants are correct!")
}

// Benchmark to show how fast operations are
func BenchmarkUseShortDeclaration(b *testing.B) {
	for i := 0; i < b.N; i++ {
		UseShortDeclaration(10, 20)
	}
}
