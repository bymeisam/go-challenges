package main

import (
	"testing"
)

// Test functions to ensure correctness

func TestConcatenation(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected string
	}{
		{"empty", []string{}, ""},
		{"single", []string{"hello"}, "hello"},
		{"multiple", []string{"hello", " ", "world"}, "hello world"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test all three implementations
			if result := ConcatWithPlus(tt.input); result != tt.expected {
				t.Errorf("ConcatWithPlus = %q; want %q", result, tt.expected)
			}
			if result := ConcatWithBuilder(tt.input); result != tt.expected {
				t.Errorf("ConcatWithBuilder = %q; want %q", result, tt.expected)
			}
			if result := ConcatWithJoin(tt.input); result != tt.expected {
				t.Errorf("ConcatWithJoin = %q; want %q", result, tt.expected)
			}
		})
	}

	t.Log("✓ All concatenation tests passed!")
}

func TestFibonacci(t *testing.T) {
	tests := []struct {
		n        int
		expected int
	}{
		{0, 0},
		{1, 1},
		{2, 1},
		{5, 5},
		{10, 55},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			if result := FibonacciRecursive(tt.n); result != tt.expected {
				t.Errorf("FibonacciRecursive(%d) = %d; want %d", tt.n, result, tt.expected)
			}
			if result := FibonacciIterative(tt.n); result != tt.expected {
				t.Errorf("FibonacciIterative(%d) = %d; want %d", tt.n, result, tt.expected)
			}
		})
	}

	t.Log("✓ All Fibonacci tests passed!")
}

func TestSearch(t *testing.T) {
	arr := []int{1, 3, 5, 7, 9, 11, 13, 15}

	tests := []struct {
		target   int
		expected int
	}{
		{1, 0},
		{7, 3},
		{15, 7},
		{2, -1},
		{20, -1},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			if result := LinearSearch(arr, tt.target); result != tt.expected {
				t.Errorf("LinearSearch(%d) = %d; want %d", tt.target, result, tt.expected)
			}
			if result := BinarySearch(arr, tt.target); result != tt.expected {
				t.Errorf("BinarySearch(%d) = %d; want %d", tt.target, result, tt.expected)
			}
		})
	}

	t.Log("✓ All search tests passed!")
}

func TestMapOperations(t *testing.T) {
	sizes := []int{10, 100}

	for _, size := range sizes {
		t.Run("", func(t *testing.T) {
			m1 := MapWithMake(size)
			m2 := MapWithoutMake(size)

			if len(m1) != size {
				t.Errorf("MapWithMake length = %d; want %d", len(m1), size)
			}
			if len(m2) != size {
				t.Errorf("MapWithoutMake length = %d; want %d", len(m2), size)
			}

			// Verify contents are the same
			for i := 0; i < size; i++ {
				if m1[i] != m2[i] {
					t.Errorf("maps differ at key %d", i)
				}
			}
		})
	}

	t.Log("✓ All map operation tests passed!")
}

// Benchmark functions

func BenchmarkConcatWithPlus(b *testing.B) {
	strs := []string{"hello", " ", "world", " ", "from", " ", "Go"}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = ConcatWithPlus(strs)
	}
}

func BenchmarkConcatWithBuilder(b *testing.B) {
	strs := []string{"hello", " ", "world", " ", "from", " ", "Go"}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = ConcatWithBuilder(strs)
	}
}

func BenchmarkConcatWithJoin(b *testing.B) {
	strs := []string{"hello", " ", "world", " ", "from", " ", "Go"}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = ConcatWithJoin(strs)
	}
}

func BenchmarkConcatLarge(b *testing.B) {
	// Create a large slice of strings
	strs := make([]string, 100)
	for i := range strs {
		strs[i] = "word"
	}

	b.Run("Plus", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = ConcatWithPlus(strs)
		}
	})

	b.Run("Builder", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = ConcatWithBuilder(strs)
		}
	})

	b.Run("Join", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = ConcatWithJoin(strs)
		}
	})
}

func BenchmarkFibonacciRecursive(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = FibonacciRecursive(20)
	}
}

func BenchmarkFibonacciIterative(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = FibonacciIterative(20)
	}
}

func BenchmarkFibonacciComparison(b *testing.B) {
	b.Run("Recursive/10", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = FibonacciRecursive(10)
		}
	})

	b.Run("Iterative/10", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = FibonacciIterative(10)
		}
	})

	b.Run("Recursive/20", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = FibonacciRecursive(20)
		}
	})

	b.Run("Iterative/20", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = FibonacciIterative(20)
		}
	})
}

func BenchmarkLinearSearch(b *testing.B) {
	// Create test array
	arr := make([]int, 1000)
	for i := range arr {
		arr[i] = i * 2
	}
	target := 500

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = LinearSearch(arr, target)
	}
}

func BenchmarkBinarySearch(b *testing.B) {
	// Create test array
	arr := make([]int, 1000)
	for i := range arr {
		arr[i] = i * 2
	}
	target := 500

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = BinarySearch(arr, target)
	}
}

func BenchmarkSearchComparison(b *testing.B) {
	// Create different sized arrays
	sizes := []int{100, 1000, 10000}

	for _, size := range sizes {
		arr := make([]int, size)
		for i := range arr {
			arr[i] = i * 2
		}
		target := size / 2 * 2 // middle element

		b.Run("Linear/"+string(rune(size)), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = LinearSearch(arr, target)
			}
		})

		b.Run("Binary/"+string(rune(size)), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = BinarySearch(arr, target)
			}
		})
	}
}

func BenchmarkMapWithMake(b *testing.B) {
	const size = 1000
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = MapWithMake(size)
	}
}

func BenchmarkMapWithoutMake(b *testing.B) {
	const size = 1000
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = MapWithoutMake(size)
	}
}

func BenchmarkMapComparison(b *testing.B) {
	sizes := []int{10, 100, 1000}

	for _, size := range sizes {
		b.Run("WithMake", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = MapWithMake(size)
			}
		})

		b.Run("WithoutMake", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = MapWithoutMake(size)
			}
		})
	}
}
