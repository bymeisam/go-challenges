package main

import "sync"

type Job struct {
	ID    int
	Value int
}

type Result struct {
	JobID  int
	Result int
}

func WorkerPool(numWorkers int, jobs []Job) []Result {
	jobsChan := make(chan Job, len(jobs))
	results := make(chan Result, len(jobs))
	
	var wg sync.WaitGroup
	
	// Start workers
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for job := range jobsChan {
				results <- Result{
					JobID:  job.ID,
					Result: job.Value * 2,
				}
			}
		}()
	}
	
	// Send jobs
	for _, job := range jobs {
		jobsChan <- job
	}
	close(jobsChan)
	
	// Wait and close results
	go func() {
		wg.Wait()
		close(results)
	}()
	
	// Collect results
	var resultSlice []Result
	for result := range results {
		resultSlice = append(resultSlice, result)
	}
	
	return resultSlice
}

func main() {}
