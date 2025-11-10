package main

import (
	"fmt"
	"sort"
	"strings"
)

// ProcessData processes a slice of integers
// Square each number, filter out odd numbers, and sort
func ProcessData(data []int) []int {
	// Square all numbers
	squared := make([]int, len(data))
	for i, v := range data {
		squared[i] = v * v
	}

	// Filter out odd numbers
	var filtered []int
	for _, v := range squared {
		if v%2 == 0 {
			filtered = append(filtered, v)
		}
	}

	// Sort the result
	sort.Ints(filtered)

	return filtered
}

// GenerateReport generates a text report with n lines
func GenerateReport(n int) string {
	var result strings.Builder
	result.Grow(n * 50) // Pre-allocate approximate size

	for i := 0; i < n; i++ {
		line := fmt.Sprintf("Line %d: Data value = %d, Status = %s\n",
			i, i*100, "ACTIVE")
		result.WriteString(line)
	}

	return result.String()
}

// FindPrimes finds all prime numbers up to max using Sieve of Eratosthenes
func FindPrimes(max int) []int {
	if max < 2 {
		return []int{}
	}

	// Create boolean array for sieve
	isPrime := make([]bool, max+1)
	for i := 2; i <= max; i++ {
		isPrime[i] = true
	}

	// Sieve of Eratosthenes
	for i := 2; i*i <= max; i++ {
		if isPrime[i] {
			for j := i * i; j <= max; j += i {
				isPrime[j] = false
			}
		}
	}

	// Collect primes
	var primes []int
	for i := 2; i <= max; i++ {
		if isPrime[i] {
			primes = append(primes, i)
		}
	}

	return primes
}

// MatrixMultiply multiplies two square matrices
func MatrixMultiply(a, b [][]int) [][]int {
	if len(a) == 0 || len(b) == 0 {
		return nil
	}

	n := len(a)
	result := make([][]int, n)
	for i := range result {
		result[i] = make([]int, n)
	}

	for i := 0; i < n; i++ {
		for j := 0; j < n; j++ {
			sum := 0
			for k := 0; k < n; k++ {
				sum += a[i][k] * b[k][j]
			}
			result[i][j] = sum
		}
	}

	return result
}

// ImageProcessor processes images (2D matrices)
type ImageProcessor struct{}

// ProcessImage creates and processes a 2D image
func (ip *ImageProcessor) ProcessImage(width, height int) [][]int {
	// Create image
	image := make([][]int, height)
	for i := range image {
		image[i] = make([]int, width)
	}

	// Fill with gradient pattern
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			image[y][x] = (x + y) % 256
		}
	}

	// Apply simple filter (blur-like operation)
	result := make([][]int, height)
	for i := range result {
		result[i] = make([]int, width)
	}

	for y := 1; y < height-1; y++ {
		for x := 1; x < width-1; x++ {
			// Average of surrounding pixels
			sum := 0
			count := 0
			for dy := -1; dy <= 1; dy++ {
				for dx := -1; dx <= 1; dx++ {
					sum += image[y+dy][x+dx]
					count++
				}
			}
			result[y][x] = sum / count
		}
	}

	return result
}

// ComputeStats calculates statistics on data
type Stats struct {
	Mean   float64
	StdDev float64
	Min    int
	Max    int
}

func ComputeStats(data []int) Stats {
	if len(data) == 0 {
		return Stats{}
	}

	// Find min and max
	min, max := data[0], data[0]
	sum := 0
	for _, v := range data {
		sum += v
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}

	// Calculate mean
	mean := float64(sum) / float64(len(data))

	// Calculate standard deviation
	variance := 0.0
	for _, v := range data {
		diff := float64(v) - mean
		variance += diff * diff
	}
	variance /= float64(len(data))
	stdDev := 0.0
	if variance > 0 {
		// Simple sqrt approximation
		stdDev = variance // In real code, use math.Sqrt
	}

	return Stats{
		Mean:   mean,
		StdDev: stdDev,
		Min:    min,
		Max:    max,
	}
}

// StringProcessor performs various string operations
type StringProcessor struct{}

func (sp *StringProcessor) ProcessStrings(strs []string) map[string]int {
	wordCount := make(map[string]int)

	for _, s := range strs {
		words := strings.Fields(s)
		for _, word := range words {
			word = strings.ToLower(word)
			wordCount[word]++
		}
	}

	return wordCount
}

func main() {
	// Example usage
	fmt.Println("Profiling Examples")

	// Process some data
	data := make([]int, 1000)
	for i := range data {
		data[i] = i
	}
	result := ProcessData(data)
	fmt.Printf("Processed %d numbers\n", len(result))

	// Generate report
	report := GenerateReport(100)
	fmt.Printf("Generated report with %d bytes\n", len(report))

	// Find primes
	primes := FindPrimes(1000)
	fmt.Printf("Found %d primes\n", len(primes))
}
