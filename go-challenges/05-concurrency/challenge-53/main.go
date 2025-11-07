package main

import (
	"context"
	"time"
)

func OperationWithTimeout(timeout time.Duration) (string, error) {
	result := make(chan string, 1)
	
	go func() {
		time.Sleep(50 * time.Millisecond)
		result <- "completed"
	}()
	
	select {
	case res := <-result:
		return res, nil
	case <-time.After(timeout):
		return "", context.DeadlineExceeded
	}
}

func ContextTimeout(ctx context.Context) error {
	select {
	case <-time.After(50 * time.Millisecond):
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func main() {}
