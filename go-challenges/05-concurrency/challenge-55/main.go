package main

import "sync"

func FanOut(input []int, numWorkers int) []chan int {
	channels := make([]chan int, numWorkers)
	
	for i := 0; i < numWorkers; i++ {
		channels[i] = make(chan int)
		
		go func(ch chan int, workerID int) {
			for _, val := range input {
				if val%numWorkers == workerID {
					ch <- val * 2
				}
			}
			close(ch)
		}(channels[i], i)
	}
	
	return channels
}

func FanIn(channels []chan int) []int {
	var wg sync.WaitGroup
	result := make(chan int)
	
	for _, ch := range channels {
		wg.Add(1)
		go func(c chan int) {
			defer wg.Done()
			for val := range c {
				result <- val
			}
		}(ch)
	}
	
	go func() {
		wg.Wait()
		close(result)
	}()
	
	var results []int
	for val := range result {
		results = append(results, val)
	}
	
	return results
}

func main() {}
