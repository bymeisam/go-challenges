package main

import (
	"fmt"
	"time"
)

func RunConcurrently(n int) {
	// TODO: Launch n goroutines
	// Each should print its number
	// Use time.Sleep to see concurrent execution
	for i := 0; i < n; i++ {
		go func(num int) {
			fmt.Printf("Goroutine %d\n", num)
		}(i)
	}
	time.Sleep(100 * time.Millisecond) // Wait for goroutines
}

func ConcurrentSum(numbers []int) int {
	// TODO: Split work across goroutines
	// Use a channel to collect results
	result := make(chan int)
	
	go func() {
		sum := 0
		for _, n := range numbers {
			sum += n
		}
		result <- sum
	}()
	
	return <-result
}

var counter int

func Race() {
	// TODO: Create a race condition
	// Multiple goroutines incrementing counter
	for i := 0; i < 1000; i++ {
		go func() {
			counter++ // Race condition!
		}()
	}
	time.Sleep(100 * time.Millisecond)
}

func main() {}
