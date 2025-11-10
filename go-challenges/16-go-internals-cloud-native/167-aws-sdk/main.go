package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"math"
	"math/rand"
	"net/url"
	"sync"
	"sync/atomic"
	"time"
)

// Challenge 167: AWS SDK Integration
// S3, SQS, Lambda operations with error handling and retries

// ===== 1. Retry Policy & Backoff Strategy =====

type RetryPolicy struct {
	maxRetries      int
	initialBackoff  time.Duration
	maxBackoff      time.Duration
	backoffMultiplier float64
	jitterFraction  float64
}

func NewRetryPolicy() *RetryPolicy {
	return &RetryPolicy{
		maxRetries:       3,
		initialBackoff:   100 * time.Millisecond,
		maxBackoff:       10 * time.Second,
		backoffMultiplier: 2.0,
		jitterFraction:   0.1,
	}
}

func (rp *RetryPolicy) GetBackoffDuration(attempt int) time.Duration {
	backoff := time.Duration(float64(rp.initialBackoff.Nanoseconds()) * math.Pow(rp.backoffMultiplier, float64(attempt)))
	if backoff > rp.maxBackoff {
		backoff = rp.maxBackoff
	}

	// Add jitter
	jitterAmount := time.Duration(float64(backoff.Nanoseconds()) * rp.jitterFraction * rand.Float64())
	return backoff + jitterAmount
}

type RetryableError struct {
	Code      string
	Message   string
	Transient bool
}

func (re *RetryableError) Error() string {
	return fmt.Sprintf("[%s] %s (transient=%v)", re.Code, re.Message, re.Transient)
}

func NewTransientError(code, message string) *RetryableError {
	return &RetryableError{Code: code, Message: message, Transient: true}
}

func NewPermanentError(code, message string) *RetryableError {
	return &RetryableError{Code: code, Message: message, Transient: false}
}

// ===== 2. Mock S3 Service =====

type MockS3Object struct {
	Key      string
	Data     []byte
	Metadata map[string]string
	ETag     string
	Created  time.Time
	Modified time.Time
}

type MockS3Bucket struct {
	Name    string
	Objects map[string]*MockS3Object
	mu      sync.RWMutex
}

type MockS3Service struct {
	Buckets map[string]*MockS3Bucket
	mu      sync.RWMutex
	rp      *RetryPolicy
	stats   *S3Stats
}

type S3Stats struct {
	PutCount     int64
	GetCount     int64
	DeleteCount  int64
	ListCount    int64
	PresignCount int64
	Errors       int64
}

func NewMockS3Service() *MockS3Service {
	return &MockS3Service{
		Buckets: make(map[string]*MockS3Bucket),
		rp:      NewRetryPolicy(),
		stats:   &S3Stats{},
	}
}

func (m *MockS3Service) CreateBucket(bucketName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.Buckets[bucketName]; exists {
		return NewPermanentError("BucketAlreadyExists", "Bucket already exists")
	}

	m.Buckets[bucketName] = &MockS3Bucket{
		Name:    bucketName,
		Objects: make(map[string]*MockS3Object),
	}
	return nil
}

func (m *MockS3Service) PutObject(ctx context.Context, bucketName, key string, data []byte) error {
	m.mu.RLock()
	bucket, exists := m.Buckets[bucketName]
	m.mu.RUnlock()

	if !exists {
		atomic.AddInt64(&m.stats.Errors, 1)
		return NewPermanentError("NoSuchBucket", "Bucket does not exist")
	}

	bucket.mu.Lock()
	defer bucket.mu.Unlock()

	etag := base64.StdEncoding.EncodeToString(generateETag(data))

	bucket.Objects[key] = &MockS3Object{
		Key:      key,
		Data:     data,
		Metadata: make(map[string]string),
		ETag:     etag,
		Created:  time.Now(),
		Modified: time.Now(),
	}

	atomic.AddInt64(&m.stats.PutCount, 1)
	return nil
}

func (m *MockS3Service) GetObject(ctx context.Context, bucketName, key string) ([]byte, error) {
	m.mu.RLock()
	bucket, exists := m.Buckets[bucketName]
	m.mu.RUnlock()

	if !exists {
		atomic.AddInt64(&m.stats.Errors, 1)
		return nil, NewPermanentError("NoSuchBucket", "Bucket does not exist")
	}

	bucket.mu.RLock()
	obj, exists := bucket.Objects[key]
	bucket.mu.RUnlock()

	if !exists {
		atomic.AddInt64(&m.stats.Errors, 1)
		return nil, NewPermanentError("NoSuchKey", "Object does not exist")
	}

	atomic.AddInt64(&m.stats.GetCount, 1)
	return obj.Data, nil
}

func (m *MockS3Service) DeleteObject(ctx context.Context, bucketName, key string) error {
	m.mu.RLock()
	bucket, exists := m.Buckets[bucketName]
	m.mu.RUnlock()

	if !exists {
		atomic.AddInt64(&m.stats.Errors, 1)
		return NewPermanentError("NoSuchBucket", "Bucket does not exist")
	}

	bucket.mu.Lock()
	defer bucket.mu.Unlock()

	if _, exists := bucket.Objects[key]; exists {
		delete(bucket.Objects, key)
	}

	atomic.AddInt64(&m.stats.DeleteCount, 1)
	return nil
}

func (m *MockS3Service) ListObjects(ctx context.Context, bucketName, prefix string) ([]string, error) {
	m.mu.RLock()
	bucket, exists := m.Buckets[bucketName]
	m.mu.RUnlock()

	if !exists {
		return nil, NewPermanentError("NoSuchBucket", "Bucket does not exist")
	}

	bucket.mu.RLock()
	defer bucket.mu.RUnlock()

	var keys []string
	for key := range bucket.Objects {
		if prefix == "" || len(key) >= len(prefix) && key[:len(prefix)] == prefix {
			keys = append(keys, key)
		}
	}

	atomic.AddInt64(&m.stats.ListCount, 1)
	return keys, nil
}

func (m *MockS3Service) GeneratePresignedURL(ctx context.Context, bucketName, key string, expiration time.Duration) (string, error) {
	m.mu.RLock()
	_, exists := m.Buckets[bucketName]
	m.mu.RUnlock()

	if !exists {
		return "", NewPermanentError("NoSuchBucket", "Bucket does not exist")
	}

	// Generate a mock presigned URL
	expiryTime := time.Now().Add(expiration).Unix()
	query := url.Values{}
	query.Set("X-Amz-Expires", "3600")
	query.Set("X-Amz-Date", fmt.Sprintf("%d", time.Now().Unix()))

	presignedURL := fmt.Sprintf("https://%s.s3.amazonaws.com/%s?%s&ExpirationTime=%d",
		bucketName, key, query.Encode(), expiryTime)

	atomic.AddInt64(&m.stats.PresignCount, 1)
	return presignedURL, nil
}

// ===== 3. S3 Client with Retry Logic =====

type S3Client struct {
	service *MockS3Service
	rp      *RetryPolicy
	stats   *RetryStats
}

type RetryStats struct {
	Attempts    int64
	Successes   int64
	Failures    int64
	Retries     int64
	TotalTime   time.Duration
}

func NewS3Client(service *MockS3Service) *S3Client {
	return &S3Client{
		service: service,
		rp:      NewRetryPolicy(),
		stats:   &RetryStats{},
	}
}

func (c *S3Client) PutObjectWithRetry(ctx context.Context, bucketName, key string, data []byte) error {
	atomic.AddInt64(&c.stats.Attempts, 1)
	start := time.Now()
	defer func() { c.stats.TotalTime += time.Since(start) }()

	for attempt := 0; attempt <= c.rp.maxRetries; attempt++ {
		err := c.service.PutObject(ctx, bucketName, key, data)
		if err == nil {
			atomic.AddInt64(&c.stats.Successes, 1)
			return nil
		}

		if retErr, ok := err.(*RetryableError); ok && !retErr.Transient {
			atomic.AddInt64(&c.stats.Failures, 1)
			return err
		}

		if attempt < c.rp.maxRetries {
			atomic.AddInt64(&c.stats.Retries, 1)
			backoff := c.rp.GetBackoffDuration(attempt)
			select {
			case <-time.After(backoff):
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}

	atomic.AddInt64(&c.stats.Failures, 1)
	return NewTransientError("MaxRetriesExceeded", "Maximum retries exceeded")
}

func (c *S3Client) GetObjectWithRetry(ctx context.Context, bucketName, key string) ([]byte, error) {
	atomic.AddInt64(&c.stats.Attempts, 1)
	start := time.Now()
	defer func() { c.stats.TotalTime += time.Since(start) }()

	for attempt := 0; attempt <= c.rp.maxRetries; attempt++ {
		data, err := c.service.GetObject(ctx, bucketName, key)
		if err == nil {
			atomic.AddInt64(&c.stats.Successes, 1)
			return data, nil
		}

		if retErr, ok := err.(*RetryableError); ok && !retErr.Transient {
			atomic.AddInt64(&c.stats.Failures, 1)
			return nil, err
		}

		if attempt < c.rp.maxRetries {
			atomic.AddInt64(&c.stats.Retries, 1)
			backoff := c.rp.GetBackoffDuration(attempt)
			select {
			case <-time.After(backoff):
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}
	}

	atomic.AddInt64(&c.stats.Failures, 1)
	return nil, NewTransientError("MaxRetriesExceeded", "Maximum retries exceeded")
}

// ===== 4. Mock SQS Service =====

type SQSMessage struct {
	MessageID     string
	Body          string
	Attributes    map[string]string
	ReceiptHandle string
	Timestamp     time.Time
}

type SQSQueue struct {
	Name           string
	Messages       []*SQSMessage
	DeadLetterQueueName string
	VisibilityTimeout   time.Duration
	mu             sync.RWMutex
}

type MockSQSService struct {
	Queues map[string]*SQSQueue
	mu     sync.RWMutex
	stats  *SQSStats
}

type SQSStats struct {
	PublishCount   int64
	ConsumeCount   int64
	DeleteCount    int64
	ReceiveCount   int64
	DLQCount       int64
}

func NewMockSQSService() *MockSQSService {
	return &MockSQSService{
		Queues: make(map[string]*SQSQueue),
		stats:  &SQSStats{},
	}
}

func (m *MockSQSService) CreateQueue(queueName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.Queues[queueName]; exists {
		return NewPermanentError("QueueAlreadyExists", "Queue already exists")
	}

	m.Queues[queueName] = &SQSQueue{
		Name:              queueName,
		Messages:          make([]*SQSMessage, 0),
		VisibilityTimeout: 30 * time.Second,
	}
	return nil
}

func (m *MockSQSService) PublishMessage(ctx context.Context, queueName, body string, attributes map[string]string) (string, error) {
	m.mu.RLock()
	queue, exists := m.Queues[queueName]
	m.mu.RUnlock()

	if !exists {
		return "", NewPermanentError("QueueDoesNotExist", "Queue does not exist")
	}

	messageID := generateMessageID()

	queue.mu.Lock()
	defer queue.mu.Unlock()

	msg := &SQSMessage{
		MessageID:     messageID,
		Body:          body,
		Attributes:    attributes,
		ReceiptHandle: generateReceiptHandle(),
		Timestamp:     time.Now(),
	}

	queue.Messages = append(queue.Messages, msg)
	atomic.AddInt64(&m.stats.PublishCount, 1)

	return messageID, nil
}

func (m *MockSQSService) ReceiveMessages(ctx context.Context, queueName string, maxMessages int) ([]*SQSMessage, error) {
	m.mu.RLock()
	queue, exists := m.Queues[queueName]
	m.mu.RUnlock()

	if !exists {
		return nil, NewPermanentError("QueueDoesNotExist", "Queue does not exist")
	}

	queue.mu.Lock()
	defer queue.mu.Unlock()

	if len(queue.Messages) == 0 {
		return []*SQSMessage{}, nil
	}

	count := maxMessages
	if count > len(queue.Messages) {
		count = len(queue.Messages)
	}

	messages := queue.Messages[:count]
	queue.Messages = queue.Messages[count:]

	atomic.AddInt64(&m.stats.ReceiveCount, 1)
	return messages, nil
}

func (m *MockSQSService) DeleteMessage(ctx context.Context, queueName, receiptHandle string) error {
	atomic.AddInt64(&m.stats.DeleteCount, 1)
	return nil
}

// ===== 5. SQS Producer/Consumer =====

type SQSProducer struct {
	service *MockSQSService
	queue   string
}

type SQSConsumer struct {
	service *MockSQSService
	queue   string
	handler func(*SQSMessage) error
}

func NewSQSProducer(service *MockSQSService, queue string) *SQSProducer {
	return &SQSProducer{
		service: service,
		queue:   queue,
	}
}

func (p *SQSProducer) PublishBatch(ctx context.Context, messages []string) error {
	for _, msg := range messages {
		_, err := p.service.PublishMessage(ctx, p.queue, msg, make(map[string]string))
		if err != nil {
			return err
		}
	}
	return nil
}

func NewSQSConsumer(service *MockSQSService, queue string, handler func(*SQSMessage) error) *SQSConsumer {
	return &SQSConsumer{
		service: service,
		queue:   queue,
		handler: handler,
	}
}

func (c *SQSConsumer) Start(ctx context.Context, concurrency int) {
	var wg sync.WaitGroup
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c.pollMessages(ctx)
		}()
	}
	wg.Wait()
}

func (c *SQSConsumer) pollMessages(ctx context.Context) {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			messages, err := c.service.ReceiveMessages(ctx, c.queue, 10)
			if err != nil {
				continue
			}

			for _, msg := range messages {
				if err := c.handler(msg); err == nil {
					c.service.DeleteMessage(ctx, c.queue, msg.ReceiptHandle)
				}
			}
		}
	}
}

// ===== 6. Lambda Invocation Handler =====

type LambdaInvoker struct {
	functions map[string]func(context.Context, []byte) ([]byte, error)
	mu        sync.RWMutex
	stats     *LambdaStats
}

type LambdaStats struct {
	Invocations int64
	Successes   int64
	Errors      int64
	TotalTime   time.Duration
}

func NewLambdaInvoker() *LambdaInvoker {
	return &LambdaInvoker{
		functions: make(map[string]func(context.Context, []byte) ([]byte, error)),
		stats:     &LambdaStats{},
	}
}

func (li *LambdaInvoker) RegisterFunction(name string, fn func(context.Context, []byte) ([]byte, error)) {
	li.mu.Lock()
	defer li.mu.Unlock()
	li.functions[name] = fn
}

func (li *LambdaInvoker) InvokSync(ctx context.Context, functionName string, payload []byte) ([]byte, error) {
	atomic.AddInt64(&li.stats.Invocations, 1)
	start := time.Now()
	defer func() { li.stats.TotalTime += time.Since(start) }()

	li.mu.RLock()
	fn, exists := li.functions[functionName]
	li.mu.RUnlock()

	if !exists {
		atomic.AddInt64(&li.stats.Errors, 1)
		return nil, NewPermanentError("FunctionNotFound", "Function not found")
	}

	result, err := fn(ctx, payload)
	if err != nil {
		atomic.AddInt64(&li.stats.Errors, 1)
		return nil, err
	}

	atomic.AddInt64(&li.stats.Successes, 1)
	return result, nil
}

func (li *LambdaInvoker) InvokeAsync(ctx context.Context, functionName string, payload []byte, callback func([]byte, error)) error {
	go func() {
		result, err := li.InvokSync(ctx, functionName, payload)
		callback(result, err)
	}()
	return nil
}

// ===== 7. Helper Functions =====

func generateETag(data []byte) []byte {
	hash := sha256.Sum256(data)
	return hash[:]
}

func generateMessageID() string {
	return fmt.Sprintf("msg-%d-%d", time.Now().Unix(), rand.Int63())
}

func generateReceiptHandle() string {
	return fmt.Sprintf("receipt-%d", rand.Int63())
}

// ===== Main Demo =====

func main() {
	fmt.Println("=== AWS SDK Integration ===\n")

	// 1. S3 Operations
	fmt.Println("1. S3 Operations with Retry")
	s3Service := NewMockS3Service()
	s3Service.CreateBucket("test-bucket")

	s3Client := NewS3Client(s3Service)

	ctx := context.Background()

	// Upload
	data := []byte("Hello, AWS S3!")
	err := s3Client.PutObjectWithRetry(ctx, "test-bucket", "test-key", data)
	fmt.Printf("Put object: error=%v\n", err)

	// Download
	retrieved, err := s3Client.GetObjectWithRetry(ctx, "test-bucket", "test-key")
	fmt.Printf("Get object: data=%s, error=%v\n", string(retrieved), err)

	// List
	keys, err := s3Service.ListObjects(ctx, "test-bucket", "")
	fmt.Printf("List objects: keys=%v, error=%v\n", keys, err)

	// Presigned URL
	presignedURL, err := s3Service.GeneratePresignedURL(ctx, "test-bucket", "test-key", time.Hour)
	fmt.Printf("Presigned URL: %s\n\n", presignedURL)

	// S3 Stats
	fmt.Printf("S3 Stats: Puts=%d, Gets=%d, Deletes=%d, Lists=%d\n",
		atomic.LoadInt64(&s3Service.stats.PutCount),
		atomic.LoadInt64(&s3Service.stats.GetCount),
		atomic.LoadInt64(&s3Service.stats.DeleteCount),
		atomic.LoadInt64(&s3Service.stats.ListCount))
	fmt.Printf("Retry Stats: Attempts=%d, Successes=%d, Retries=%d, Failures=%d\n\n",
		atomic.LoadInt64(&s3Client.stats.Attempts),
		atomic.LoadInt64(&s3Client.stats.Successes),
		atomic.LoadInt64(&s3Client.stats.Retries),
		atomic.LoadInt64(&s3Client.stats.Failures))

	// 2. SQS Operations
	fmt.Println("2. SQS Producer/Consumer")
	sqsService := NewMockSQSService()
	sqsService.CreateQueue("test-queue")

	producer := NewSQSProducer(sqsService, "test-queue")
	producer.PublishBatch(ctx, []string{"msg1", "msg2", "msg3"})

	processed := int32(0)
	consumer := NewSQSConsumer(sqsService, "test-queue", func(msg *SQSMessage) error {
		atomic.AddInt32(&processed, 1)
		return nil
	})

	consumerCtx, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
	consumer.Start(consumerCtx, 2)
	cancel()

	fmt.Printf("SQS Stats: Published=%d, Received=%d, Processed=%d\n\n",
		atomic.LoadInt64(&sqsService.stats.PublishCount),
		atomic.LoadInt64(&sqsService.stats.ReceiveCount),
		processed)

	// 3. Lambda Invocation
	fmt.Println("3. Lambda Invocation")
	invoker := NewLambdaInvoker()

	invoker.RegisterFunction("echo", func(ctx context.Context, payload []byte) ([]byte, error) {
		return payload, nil
	})

	result, err := invoker.InvokSync(ctx, "echo", []byte("Hello Lambda"))
	fmt.Printf("Lambda result: %s, error=%v\n", string(result), err)

	// Async invocation
	var asyncResult []byte
	var asyncErr error
	invoker.InvokeAsync(ctx, "echo", []byte("Async"), func(data []byte, err error) {
		asyncResult = data
		asyncErr = err
	})

	time.Sleep(100 * time.Millisecond)
	fmt.Printf("Async Lambda result: %s, error=%v\n", string(asyncResult), asyncErr)

	fmt.Printf("Lambda Stats: Invocations=%d, Successes=%d, Errors=%d\n\n",
		atomic.LoadInt64(&invoker.stats.Invocations),
		atomic.LoadInt64(&invoker.stats.Successes),
		atomic.LoadInt64(&invoker.stats.Errors))

	// 4. Retry Backoff Strategy
	fmt.Println("4. Retry Backoff Strategy")
	rp := NewRetryPolicy()
	for i := 0; i < 3; i++ {
		backoff := rp.GetBackoffDuration(i)
		fmt.Printf("Attempt %d backoff: %v\n", i, backoff)
	}

	fmt.Println("\n=== Complete ===")
}
