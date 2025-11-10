package main

import (
	"testing"
)

func TestProcessData(t *testing.T) {
	data := []int{1, 2, 3, 4, 5}
	result := ProcessData(data)

	// Expected: square (1,4,9,16,25), filter even (4,16), sort (4,16)
	expected := []int{4, 16}

	if len(result) != len(expected) {
		t.Errorf("length = %d; want %d", len(result), len(expected))
	}

	for i, v := range result {
		if v != expected[i] {
			t.Errorf("result[%d] = %d; want %d", i, v, expected[i])
		}
	}

	t.Log("✓ ProcessData test passed!")
}

func TestGenerateReport(t *testing.T) {
	report := GenerateReport(10)

	if len(report) == 0 {
		t.Error("report should not be empty")
	}

	if len(report) < 100 {
		t.Errorf("report seems too short: %d bytes", len(report))
	}

	t.Log("✓ GenerateReport test passed!")
}

func TestFindPrimes(t *testing.T) {
	primes := FindPrimes(20)
	expected := []int{2, 3, 5, 7, 11, 13, 17, 19}

	if len(primes) != len(expected) {
		t.Errorf("found %d primes; want %d", len(primes), len(expected))
	}

	for i, p := range primes {
		if p != expected[i] {
			t.Errorf("primes[%d] = %d; want %d", i, p, expected[i])
		}
	}

	t.Log("✓ FindPrimes test passed!")
}

func TestMatrixMultiply(t *testing.T) {
	a := [][]int{
		{1, 2},
		{3, 4},
	}
	b := [][]int{
		{5, 6},
		{7, 8},
	}

	result := MatrixMultiply(a, b)

	// Expected result:
	// [1*5+2*7, 1*6+2*8]   [19, 22]
	// [3*5+4*7, 3*6+4*8] = [43, 50]

	if result[0][0] != 19 || result[0][1] != 22 ||
		result[1][0] != 43 || result[1][1] != 50 {
		t.Errorf("matrix multiply incorrect: %v", result)
	}

	t.Log("✓ MatrixMultiply test passed!")
}

func TestImageProcessor(t *testing.T) {
	ip := &ImageProcessor{}
	image := ip.ProcessImage(10, 10)

	if len(image) != 10 {
		t.Errorf("image height = %d; want 10", len(image))
	}

	if len(image[0]) != 10 {
		t.Errorf("image width = %d; want 10", len(image[0]))
	}

	t.Log("✓ ImageProcessor test passed!")
}

// Benchmarks for profiling

func BenchmarkProcessData(b *testing.B) {
	data := make([]int, 1000)
	for i := range data {
		data[i] = i
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = ProcessData(data)
	}
}

func BenchmarkGenerateReport(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = GenerateReport(1000)
	}
}

func BenchmarkFindPrimes(b *testing.B) {
	b.Run("1000", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = FindPrimes(1000)
		}
	})

	b.Run("10000", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = FindPrimes(10000)
		}
	})

	b.Run("100000", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = FindPrimes(100000)
		}
	})
}

func BenchmarkMatrixMultiply(b *testing.B) {
	sizes := []int{10, 50, 100}

	for _, size := range sizes {
		b.Run("", func(b *testing.B) {
			// Create test matrices
			a := make([][]int, size)
			bMatrix := make([][]int, size)
			for i := 0; i < size; i++ {
				a[i] = make([]int, size)
				bMatrix[i] = make([]int, size)
				for j := 0; j < size; j++ {
					a[i][j] = i + j
					bMatrix[i][j] = i - j
				}
			}

			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				_ = MatrixMultiply(a, bMatrix)
			}
		})
	}
}

func BenchmarkImageProcessor(b *testing.B) {
	ip := &ImageProcessor{}

	sizes := []struct {
		width, height int
	}{
		{100, 100},
		{500, 500},
		{1000, 1000},
	}

	for _, size := range sizes {
		b.Run("", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = ip.ProcessImage(size.width, size.height)
			}
		})
	}
}

func BenchmarkComputeStats(b *testing.B) {
	data := make([]int, 10000)
	for i := range data {
		data[i] = i
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = ComputeStats(data)
	}
}

func BenchmarkStringProcessor(b *testing.B) {
	sp := &StringProcessor{}
	strs := make([]string, 1000)
	for i := range strs {
		strs[i] = "the quick brown fox jumps over the lazy dog"
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = sp.ProcessStrings(strs)
	}
}

// Memory-intensive benchmarks

func BenchmarkMemoryAllocation(b *testing.B) {
	b.Run("SliceAppend", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			var s []int
			for j := 0; j < 1000; j++ {
				s = append(s, j)
			}
		}
	})

	b.Run("SlicePreallocate", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			s := make([]int, 0, 1000)
			for j := 0; j < 1000; j++ {
				s = append(s, j)
			}
		}
	})

	b.Run("MapNoHint", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			m := make(map[int]int)
			for j := 0; j < 1000; j++ {
				m[j] = j
			}
		}
	})

	b.Run("MapWithHint", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			m := make(map[int]int, 1000)
			for j := 0; j < 1000; j++ {
				m[j] = j
			}
		}
	})
}

// CPU-intensive benchmark

func BenchmarkCPUIntensive(b *testing.B) {
	b.Run("Factorial", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			factorial(20)
		}
	})

	b.Run("Fibonacci", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			fibonacci(30)
		}
	})
}

func factorial(n int) int {
	if n <= 1 {
		return 1
	}
	return n * factorial(n-1)
}

func fibonacci(n int) int {
	if n <= 1 {
		return n
	}
	return fibonacci(n-1) + fibonacci(n-2)
}
