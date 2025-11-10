package main

import "time"

func RateLimiter(requests int, duration time.Duration) <-chan time.Time {
	limiter := make(chan time.Time, requests)
	
	go func() {
		ticker := time.NewTicker(duration / time.Duration(requests))
		defer ticker.Stop()
		
		for t := range ticker.C {
			limiter <- t
		}
	}()
	
	return limiter
}

func ProcessWithRateLimit(items []string, rps int) []string {
	limiter := time.Tick(time.Second / time.Duration(rps))
	
	results := make([]string, 0, len(items))
	for _, item := range items {
		<-limiter
		results = append(results, "processed: "+item)
	}
	
	return results
}

func main() {}
