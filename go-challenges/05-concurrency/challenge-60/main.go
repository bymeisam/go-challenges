package main

import (
	"context"
	"time"
)

func WorkWithContext(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(10 * time.Millisecond):
			// Do work
		}
	}
}

func CancellableOperation(ctx context.Context) (string, error) {
	result := make(chan string, 1)
	
	go func() {
		time.Sleep(50 * time.Millisecond)
		result <- "completed"
	}()
	
	select {
	case res := <-result:
		return res, nil
	case <-ctx.Done():
		return "", ctx.Err()
	}
}

func main() {}
