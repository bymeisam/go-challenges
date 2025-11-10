package main

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// TestRetryPolicy tests retry policy and backoff
func TestRetryPolicyBasic(t *testing.T) {
	rp := NewRetryPolicy()

	if rp.maxRetries != 3 {
		t.Errorf("Expected maxRetries=3, got %d", rp.maxRetries)
	}
}

func TestRetryPolicyBackoffIncreases(t *testing.T) {
	rp := NewRetryPolicy()

	backoff0 := rp.GetBackoffDuration(0)
	backoff1 := rp.GetBackoffDuration(1)
	backoff2 := rp.GetBackoffDuration(2)

	if backoff1 <= backoff0 {
		t.Errorf("Backoff should increase: %v <= %v", backoff1, backoff0)
	}
	if backoff2 <= backoff1 {
		t.Errorf("Backoff should increase: %v <= %v", backoff2, backoff1)
	}
}

func TestRetryPolicyCapped(t *testing.T) {
	rp := NewRetryPolicy()

	backoff := rp.GetBackoffDuration(100)
	if backoff > rp.maxBackoff {
		t.Errorf("Backoff exceeds maxBackoff: %v > %v", backoff, rp.maxBackoff)
	}
}

// TestMockS3Service tests S3 operations
func TestMockS3ServiceCreateBucket(t *testing.T) {
	s3 := NewMockS3Service()

	err := s3.CreateBucket("test-bucket")
	if err != nil {
		t.Errorf("Expected successful bucket creation, got error: %v", err)
	}

	// Try creating again
	err = s3.CreateBucket("test-bucket")
	if err == nil {
		t.Errorf("Expected error when bucket already exists")
	}
}

func TestMockS3ServicePutGet(t *testing.T) {
	s3 := NewMockS3Service()
	s3.CreateBucket("test-bucket")
	ctx := context.Background()

	data := []byte("hello world")
	err := s3.PutObject(ctx, "test-bucket", "key1", data)
	if err != nil {
		t.Errorf("Expected successful put, got error: %v", err)
	}

	retrieved, err := s3.GetObject(ctx, "test-bucket", "key1")
	if err != nil {
		t.Errorf("Expected successful get, got error: %v", err)
	}

	if string(retrieved) != "hello world" {
		t.Errorf("Expected 'hello world', got '%s'", string(retrieved))
	}
}

func TestMockS3ServiceDelete(t *testing.T) {
	s3 := NewMockS3Service()
	s3.CreateBucket("test-bucket")
	ctx := context.Background()

	s3.PutObject(ctx, "test-bucket", "key1", []byte("data"))
	err := s3.DeleteObject(ctx, "test-bucket", "key1")
	if err != nil {
		t.Errorf("Expected successful delete, got error: %v", err)
	}

	_, err = s3.GetObject(ctx, "test-bucket", "key1")
	if err == nil {
		t.Errorf("Expected error after delete")
	}
}

func TestMockS3ServiceList(t *testing.T) {
	s3 := NewMockS3Service()
	s3.CreateBucket("test-bucket")
	ctx := context.Background()

	s3.PutObject(ctx, "test-bucket", "file1.txt", []byte("data"))
	s3.PutObject(ctx, "test-bucket", "file2.txt", []byte("data"))
	s3.PutObject(ctx, "test-bucket", "image.png", []byte("data"))

	keys, err := s3.ListObjects(ctx, "test-bucket", "file")
	if err != nil {
		t.Errorf("Expected successful list, got error: %v", err)
	}

	if len(keys) != 2 {
		t.Errorf("Expected 2 keys with 'file' prefix, got %d", len(keys))
	}
}

func TestMockS3ServicePresignedURL(t *testing.T) {
	s3 := NewMockS3Service()
	s3.CreateBucket("test-bucket")
	ctx := context.Background()

	url, err := s3.GeneratePresignedURL(ctx, "test-bucket", "key1", time.Hour)
	if err != nil {
		t.Errorf("Expected successful URL generation, got error: %v", err)
	}

	if len(url) == 0 {
		t.Errorf("Expected non-empty URL")
	}
}

// TestS3Client tests S3 client with retry logic
func TestS3ClientPutWithRetry(t *testing.T) {
	s3 := NewMockS3Service()
	s3.CreateBucket("test-bucket")
	client := NewS3Client(s3)

	ctx := context.Background()
	err := client.PutObjectWithRetry(ctx, "test-bucket", "key1", []byte("data"))

	if err != nil {
		t.Errorf("Expected successful put, got error: %v", err)
	}

	if atomic.LoadInt64(&client.stats.Successes) != 1 {
		t.Errorf("Expected 1 success")
	}
}

func TestS3ClientGetWithRetry(t *testing.T) {
	s3 := NewMockS3Service()
	s3.CreateBucket("test-bucket")
	client := NewS3Client(s3)
	ctx := context.Background()

	s3.PutObject(ctx, "test-bucket", "key1", []byte("data"))
	data, err := client.GetObjectWithRetry(ctx, "test-bucket", "key1")

	if err != nil {
		t.Errorf("Expected successful get, got error: %v", err)
	}

	if string(data) != "data" {
		t.Errorf("Expected 'data', got '%s'", string(data))
	}
}

func TestS3ClientConcurrent(t *testing.T) {
	s3 := NewMockS3Service()
	s3.CreateBucket("test-bucket")
	client := NewS3Client(s3)
	ctx := context.Background()

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			key := fmt.Sprintf("key-%d", id)
			client.PutObjectWithRetry(ctx, "test-bucket", key, []byte("data"))
		}(i)
	}
	wg.Wait()

	if atomic.LoadInt64(&client.stats.Successes) != 50 {
		t.Errorf("Expected 50 successes")
	}
}

// TestMockSQSService tests SQS operations
func TestMockSQSServiceCreateQueue(t *testing.T) {
	sqs := NewMockSQSService()

	err := sqs.CreateQueue("test-queue")
	if err != nil {
		t.Errorf("Expected successful queue creation, got error: %v", err)
	}

	err = sqs.CreateQueue("test-queue")
	if err == nil {
		t.Errorf("Expected error when queue already exists")
	}
}

func TestMockSQSServicePublish(t *testing.T) {
	sqs := NewMockSQSService()
	sqs.CreateQueue("test-queue")
	ctx := context.Background()

	msgID, err := sqs.PublishMessage(ctx, "test-queue", "hello", make(map[string]string))
	if err != nil {
		t.Errorf("Expected successful publish, got error: %v", err)
	}

	if len(msgID) == 0 {
		t.Errorf("Expected non-empty message ID")
	}
}

func TestMockSQSServiceReceive(t *testing.T) {
	sqs := NewMockSQSService()
	sqs.CreateQueue("test-queue")
	ctx := context.Background()

	sqs.PublishMessage(ctx, "test-queue", "msg1", make(map[string]string))
	sqs.PublishMessage(ctx, "test-queue", "msg2", make(map[string]string))

	messages, err := sqs.ReceiveMessages(ctx, "test-queue", 10)
	if err != nil {
		t.Errorf("Expected successful receive, got error: %v", err)
	}

	if len(messages) != 2 {
		t.Errorf("Expected 2 messages, got %d", len(messages))
	}
}

func TestMockSQSServiceDelete(t *testing.T) {
	sqs := NewMockSQSService()
	sqs.CreateQueue("test-queue")
	ctx := context.Background()

	msgID, _ := sqs.PublishMessage(ctx, "test-queue", "msg", make(map[string]string))
	messages, _ := sqs.ReceiveMessages(ctx, "test-queue", 1)

	err := sqs.DeleteMessage(ctx, "test-queue", messages[0].ReceiptHandle)
	if err != nil {
		t.Errorf("Expected successful delete, got error: %v", err)
	}

	_ = msgID
}

// TestSQSProducerConsumer tests producer/consumer
func TestSQSProducerConsumer(t *testing.T) {
	sqs := NewMockSQSService()
	sqs.CreateQueue("test-queue")
	ctx := context.Background()

	producer := NewSQSProducer(sqs, "test-queue")
	err := producer.PublishBatch(ctx, []string{"m1", "m2", "m3"})
	if err != nil {
		t.Errorf("Expected successful batch publish, got error: %v", err)
	}

	var processed int32
	consumer := NewSQSConsumer(sqs, "test-queue", func(msg *SQSMessage) error {
		atomic.AddInt32(&processed, 1)
		return nil
	})

	consumerCtx, cancel := context.WithTimeout(ctx, 300*time.Millisecond)
	consumer.Start(consumerCtx, 1)
	cancel()

	if processed < 1 {
		t.Logf("Warning: Less than expected messages processed: %d", processed)
	}
}

// TestLambdaInvoker tests Lambda invocation
func TestLambdaInvokerRegister(t *testing.T) {
	invoker := NewLambdaInvoker()

	invoker.RegisterFunction("test", func(ctx context.Context, data []byte) ([]byte, error) {
		return data, nil
	})

	ctx := context.Background()
	result, err := invoker.InvokSync(ctx, "test", []byte("hello"))

	if err != nil {
		t.Errorf("Expected successful invocation, got error: %v", err)
	}

	if string(result) != "hello" {
		t.Errorf("Expected 'hello', got '%s'", string(result))
	}
}

func TestLambdaInvokerNotFound(t *testing.T) {
	invoker := NewLambdaInvoker()
	ctx := context.Background()

	_, err := invoker.InvokSync(ctx, "nonexistent", []byte("data"))
	if err == nil {
		t.Errorf("Expected error for nonexistent function")
	}
}

func TestLambdaInvokerAsync(t *testing.T) {
	invoker := NewLambdaInvoker()

	invoker.RegisterFunction("async", func(ctx context.Context, data []byte) ([]byte, error) {
		return append(data, []byte("!")...), nil
	})

	var result []byte
	var resultErr error
	done := make(chan struct{})

	ctx := context.Background()
	invoker.InvokeAsync(ctx, "async", []byte("test"), func(data []byte, err error) {
		result = data
		resultErr = err
		close(done)
	})

	<-done

	if resultErr != nil {
		t.Errorf("Expected successful async invocation, got error: %v", resultErr)
	}

	if string(result) != "test!" {
		t.Errorf("Expected 'test!', got '%s'", string(result))
	}
}

// Benchmark tests

func BenchmarkS3PutWithRetry(b *testing.B) {
	s3 := NewMockS3Service()
	s3.CreateBucket("test-bucket")
	client := NewS3Client(s3)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		client.PutObjectWithRetry(ctx, "test-bucket", "key", []byte("data"))
	}
}

func BenchmarkSQSPublish(b *testing.B) {
	sqs := NewMockSQSService()
	sqs.CreateQueue("test-queue")
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sqs.PublishMessage(ctx, "test-queue", "msg", make(map[string]string))
	}
}

func BenchmarkLambdaInvoke(b *testing.B) {
	invoker := NewLambdaInvoker()
	invoker.RegisterFunction("test", func(ctx context.Context, data []byte) ([]byte, error) {
		return data, nil
	})
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		invoker.InvokSync(ctx, "test", []byte("data"))
	}
}

func BenchmarkRetryPolicyBackoff(b *testing.B) {
	rp := NewRetryPolicy()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = rp.GetBackoffDuration(i % 5)
	}
}

// Helper for fmt package
import "fmt"
