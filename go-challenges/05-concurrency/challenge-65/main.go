package main

import "sync"

type Task struct {
	ID   int
	Data string
}

type Result struct {
	TaskID int
	Output string
}

func ProducerConsumer(numProducers, numConsumers, numTasks int) []Result {
	tasks := make(chan Task, 10)
	results := make(chan Result, 10)
	
	var producerWg sync.WaitGroup
	var consumerWg sync.WaitGroup
	
	// Producers
	for i := 0; i < numProducers; i++ {
		producerWg.Add(1)
		go func(producerID int) {
			defer producerWg.Done()
			for j := 0; j < numTasks/numProducers; j++ {
				taskID := producerID*100 + j
				tasks <- Task{ID: taskID, Data: "task"}
			}
		}(i)
	}
	
	// Close tasks when all producers done
	go func() {
		producerWg.Wait()
		close(tasks)
	}()
	
	// Consumers
	for i := 0; i < numConsumers; i++ {
		consumerWg.Add(1)
		go func() {
			defer consumerWg.Done()
			for task := range tasks {
				results <- Result{
					TaskID: task.ID,
					Output: "processed: " + task.Data,
				}
			}
		}()
	}
	
	// Close results when all consumers done
	go func() {
		consumerWg.Wait()
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
