package main

import "sync"

func ProcessWithWaitGroup(items []string) []string {
	var wg sync.WaitGroup
	results := make([]string, len(items))
	
	for i, item := range items {
		wg.Add(1)
		go func(index int, value string) {
			defer wg.Done()
			results[index] = "processed: " + value
		}(i, item)
	}
	
	wg.Wait()
	return results
}

func main() {}
