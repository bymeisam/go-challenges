package main

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// CronScheduler parses and manages cron expressions
type CronScheduler struct {
	expression string
	minute     []int
	hour       []int
	day        []int
	month      []int
	dayOfWeek  []int
	location   *time.Location
}

// JobScheduler manages job queue and execution
type JobScheduler struct {
	jobs          map[string]*Job
	queue         chan *Job
	workers       int
	running       atomic.Bool
	activeJobs    atomic.Int32
	completedJobs atomic.Int64
	failedJobs    atomic.Int64
	mu            sync.RWMutex
}

// Job represents a schedulable job
type Job struct {
	ID              string
	Name            string
	CronExpression  string
	Handler         JobHandler
	Priority        int
	MaxRetries      int
	Timeout         time.Duration
	CreatedAt       time.Time
	LastRun         time.Time
	NextRun         time.Time
	Status          JobStatus
	RetryCount      int
	Dependencies    []string
	LastError       error
	ExecutionTime   time.Duration
}

// JobStatus represents job execution status
type JobStatus string

const (
	StatusPending   JobStatus = "pending"
	StatusRunning   JobStatus = "running"
	StatusCompleted JobStatus = "completed"
	StatusFailed    JobStatus = "failed"
)

// JobHandler is the function executed by a job
type JobHandler func(ctx context.Context) error

// BatchProcessor processes data in batches
type BatchProcessor struct {
	batchSize      int
	workers        int
	dataSource     DataSource
	transformer    Transformer
	loader         Loader
	validator      Validator
	checkpointer   Checkpointer
	metrics        *BatchMetrics
	deadLetterQueue *DeadLetterQueue
}

// Batch represents a batch of data
type Batch struct {
	ID           string
	Items        []interface{}
	CreatedAt    time.Time
	ProcessedAt  time.Time
	StartIndex   int
	EndIndex     int
	ErrorCount   int
	Status       BatchStatus
}

// BatchStatus represents batch processing status
type BatchStatus string

const (
	BatchPending     BatchStatus = "pending"
	BatchProcessing  BatchStatus = "processing"
	BatchCompleted   BatchStatus = "completed"
	BatchFailed      BatchStatus = "failed"
	BatchCheckpointed BatchStatus = "checkpointed"
)

// ETLPipeline orchestrates Extract-Transform-Load process
type ETLPipeline struct {
	name           string
	extractor      DataExtractor
	transformers   []DataTransformer
	loader         DataLoader
	validator      DataValidator
	errorHandler   ErrorHandler
	checkpointer   PipelineCheckpointer
	metrics        *ETLMetrics
	deadLetters    []*DeadLetter
	mu             sync.RWMutex
}

// DataExtractor extracts data from source
type DataExtractor interface {
	Extract(ctx context.Context) ([]interface{}, error)
}

// DataTransformer transforms data
type DataTransformer interface {
	Transform(item interface{}) (interface{}, error)
}

// DataLoader loads data to destination
type DataLoader interface {
	Load(ctx context.Context, items []interface{}) error
}

// DataValidator validates data
type DataValidator interface {
	Validate(item interface{}) error
}

// ErrorHandler handles errors in pipeline
type ErrorHandler interface {
	HandleError(item interface{}, err error)
}

// DataSource represents data source
type DataSource interface {
	Read() ([]interface{}, error)
}

// Transformer transforms items
type Transformer interface {
	Transform(item interface{}) (interface{}, error)
}

// Loader loads items
type Loader interface {
	Load(items []interface{}) error
}

// Validator validates items
type Validator interface {
	Validate(item interface{}) error
}

// Checkpointer saves processing checkpoint
type Checkpointer interface {
	SaveCheckpoint(batchID string, index int) error
	LoadCheckpoint(batchID string) (int, error)
}

// PipelineCheckpointer saves pipeline state
type PipelineCheckpointer interface {
	SaveState(pipelineID string, state interface{}) error
	LoadState(pipelineID string) (interface{}, error)
}

// DeadLetterQueue stores failed items
type DeadLetterQueue struct {
	items []interface{}
	mu    sync.RWMutex
}

// DeadLetter represents a failed item
type DeadLetter struct {
	Item      interface{}
	Error     string
	Timestamp time.Time
	PipelineID string
}

// BatchMetrics tracks batch processing metrics
type BatchMetrics struct {
	totalBatches     atomic.Int64
	processedBatches atomic.Int64
	failedBatches    atomic.Int64
	totalItems       atomic.Int64
	processedItems   atomic.Int64
	failedItems      atomic.Int64
	totalDuration    atomic.Int64
}

// ETLMetrics tracks ETL pipeline metrics
type ETLMetrics struct {
	extractedCount  atomic.Int64
	transformedCount atomic.Int64
	loadedCount     atomic.Int64
	errorCount      atomic.Int64
	startTime       time.Time
	endTime         time.Time
	totalDuration   atomic.Int64
}

// JobDAG represents a directed acyclic graph of jobs
type JobDAG struct {
	jobs        map[string]*Job
	dependencies map[string][]string
	mu          sync.RWMutex
}

// NewCronScheduler creates a new cron scheduler
func NewCronScheduler(expression string) (*CronScheduler, error) {
	cs := &CronScheduler{
		expression: expression,
		location:   time.Local,
	}

	if err := cs.parse(); err != nil {
		return nil, err
	}

	return cs, nil
}

// parse parses a cron expression
func (cs *CronScheduler) parse() error {
	parts := strings.Fields(cs.expression)
	if len(parts) != 5 {
		return errors.New("cron expression must have 5 fields")
	}

	var err error
	cs.minute, err = parseField(parts[0], 0, 59)
	if err != nil {
		return fmt.Errorf("invalid minute: %v", err)
	}

	cs.hour, err = parseField(parts[1], 0, 23)
	if err != nil {
		return fmt.Errorf("invalid hour: %v", err)
	}

	cs.day, err = parseField(parts[2], 1, 31)
	if err != nil {
		return fmt.Errorf("invalid day: %v", err)
	}

	cs.month, err = parseField(parts[3], 1, 12)
	if err != nil {
		return fmt.Errorf("invalid month: %v", err)
	}

	cs.dayOfWeek, err = parseField(parts[4], 0, 6)
	if err != nil {
		return fmt.Errorf("invalid day of week: %v", err)
	}

	return nil
}

// parseField parses a cron field
func parseField(field string, min, max int) ([]int, error) {
	if field == "*" {
		result := make([]int, 0)
		for i := min; i <= max; i++ {
			result = append(result, i)
		}
		return result, nil
	}

	if strings.Contains(field, "/") {
		return parseInterval(field, min, max)
	}

	if strings.Contains(field, ",") {
		return parseList(field, min, max)
	}

	if strings.Contains(field, "-") {
		return parseRange(field, min, max)
	}

	val, err := strconv.Atoi(field)
	if err != nil {
		return nil, err
	}

	if val < min || val > max {
		return nil, fmt.Errorf("value %d out of range [%d-%d]", val, min, max)
	}

	return []int{val}, nil
}

// parseInterval parses a cron interval (*/5)
func parseInterval(field string, min, max int) ([]int, error) {
	parts := strings.Split(field, "/")
	if len(parts) != 2 {
		return nil, errors.New("invalid interval format")
	}

	interval, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, err
	}

	result := make([]int, 0)
	start := min
	if parts[0] != "*" {
		start, err = strconv.Atoi(parts[0])
		if err != nil {
			return nil, err
		}
	}

	for i := start; i <= max; i += interval {
		result = append(result, i)
	}

	return result, nil
}

// parseRange parses a cron range (1-5)
func parseRange(field string, min, max int) ([]int, error) {
	parts := strings.Split(field, "-")
	if len(parts) != 2 {
		return nil, errors.New("invalid range format")
	}

	start, err := strconv.Atoi(parts[0])
	if err != nil {
		return nil, err
	}

	end, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, err
	}

	result := make([]int, 0)
	for i := start; i <= end; i++ {
		result = append(result, i)
	}

	return result, nil
}

// parseList parses a cron list (1,3,5)
func parseList(field string, min, max int) ([]int, error) {
	parts := strings.Split(field, ",")
	result := make([]int, 0)

	for _, part := range parts {
		val, err := strconv.Atoi(part)
		if err != nil {
			return nil, err
		}

		if val < min || val > max {
			return nil, fmt.Errorf("value %d out of range [%d-%d]", val, min, max)
		}

		result = append(result, val)
	}

	sort.Ints(result)
	return result, nil
}

// NextRun calculates the next execution time
func (cs *CronScheduler) NextRun(from time.Time) time.Time {
	next := from.Add(1 * time.Minute)
	next = next.Truncate(time.Minute)

	for {
		if intContains(cs.month, int(next.Month())) &&
			intContains(cs.day, next.Day()) &&
			(intContains(cs.dayOfWeek, int(next.Weekday())) || intContains(cs.day, next.Day())) &&
			intContains(cs.hour, next.Hour()) &&
			intContains(cs.minute, next.Minute()) {
			return next
		}

		next = next.Add(1 * time.Minute)
		if next.After(from.Add(366 * 24 * time.Hour)) {
			return from
		}
	}
}

func intContains(slice []int, val int) bool {
	for _, v := range slice {
		if v == val {
			return true
		}
	}
	return false
}

// NewJobScheduler creates a new job scheduler
func NewJobScheduler(workers int) *JobScheduler {
	return &JobScheduler{
		jobs:    make(map[string]*Job),
		queue:   make(chan *Job, workers*10),
		workers: workers,
	}
}

// RegisterJob registers a new job
func (js *JobScheduler) RegisterJob(job *Job) error {
	if job.ID == "" {
		return errors.New("job ID cannot be empty")
	}

	js.mu.Lock()
	defer js.mu.Unlock()

	js.jobs[job.ID] = job
	job.Status = StatusPending

	return nil
}

// Start starts the job scheduler
func (js *JobScheduler) Start(ctx context.Context) {
	if js.running.Load() {
		return
	}

	js.running.Store(true)

	// Start worker goroutines
	for i := 0; i < js.workers; i++ {
		go js.worker(ctx)
	}

	// Start scheduler goroutine
	go js.scheduler(ctx)
}

// worker processes jobs from the queue
func (js *JobScheduler) worker(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case job := <-js.queue:
			if job == nil {
				return
			}

			js.executeJob(ctx, job)
		}
	}
}

// executeJob executes a single job
func (js *JobScheduler) executeJob(ctx context.Context, job *Job) {
	js.activeJobs.Add(1)
	defer js.activeJobs.Add(-1)

	job.Status = StatusRunning
	start := time.Now()

	// Create context with timeout
	jobCtx, cancel := context.WithTimeout(ctx, job.Timeout)
	defer cancel()

	err := job.Handler(jobCtx)
	job.ExecutionTime = time.Since(start)

	if err != nil {
		job.LastError = err
		job.RetryCount++

		if job.RetryCount < job.MaxRetries {
			job.Status = StatusPending
			js.queue <- job
		} else {
			job.Status = StatusFailed
			js.failedJobs.Add(1)
		}
	} else {
		job.Status = StatusCompleted
		job.LastRun = start
		js.completedJobs.Add(1)
	}
}

// scheduler schedules jobs based on cron expressions
func (js *JobScheduler) scheduler(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			js.checkAndSchedule()
		}
	}
}

// checkAndSchedule checks jobs and schedules them
func (js *JobScheduler) checkAndSchedule() {
	js.mu.RLock()
	defer js.mu.RUnlock()

	now := time.Now()

	for _, job := range js.jobs {
		if job.CronExpression == "" {
			continue
		}

		scheduler, err := NewCronScheduler(job.CronExpression)
		if err != nil {
			continue
		}

		nextRun := scheduler.NextRun(job.LastRun)
		if now.After(nextRun) && job.Status != StatusRunning {
			js.queue <- job
		}
	}
}

// Stop stops the job scheduler
func (js *JobScheduler) Stop() {
	js.running.Store(false)
	close(js.queue)
}

// GetJob retrieves a job by ID
func (js *JobScheduler) GetJob(id string) *Job {
	js.mu.RLock()
	defer js.mu.RUnlock()
	return js.jobs[id]
}

// GetJobHistory returns execution history
func (js *JobScheduler) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"completed_jobs": js.completedJobs.Load(),
		"failed_jobs":    js.failedJobs.Load(),
		"active_jobs":    js.activeJobs.Load(),
	}
}

// NewBatchProcessor creates a new batch processor
func NewBatchProcessor(batchSize, workers int) *BatchProcessor {
	return &BatchProcessor{
		batchSize:       batchSize,
		workers:         workers,
		metrics:         &BatchMetrics{},
		deadLetterQueue: &DeadLetterQueue{items: make([]interface{}, 0)},
	}
}

// SetDataSource sets the data source
func (bp *BatchProcessor) SetDataSource(ds DataSource) {
	bp.dataSource = ds
}

// SetTransformer sets the transformer
func (bp *BatchProcessor) SetTransformer(t Transformer) {
	bp.transformer = t
}

// SetLoader sets the loader
func (bp *BatchProcessor) SetLoader(l Loader) {
	bp.loader = l
}

// SetValidator sets the validator
func (bp *BatchProcessor) SetValidator(v Validator) {
	bp.validator = v
}

// SetCheckpointer sets the checkpointer
func (bp *BatchProcessor) SetCheckpointer(c Checkpointer) {
	bp.checkpointer = c
}

// Process processes data in batches
func (bp *BatchProcessor) Process(ctx context.Context) error {
	if bp.dataSource == nil {
		return errors.New("data source not set")
	}

	data, err := bp.dataSource.Read()
	if err != nil {
		return fmt.Errorf("failed to read data: %v", err)
	}

	bp.metrics.totalItems.Store(int64(len(data)))

	// Create batches
	for i := 0; i < len(data); i += bp.batchSize {
		end := i + bp.batchSize
		if end > len(data) {
			end = len(data)
		}

		batch := &Batch{
			ID:         fmt.Sprintf("batch-%d", i/bp.batchSize),
			Items:      data[i:end],
			CreatedAt:  time.Now(),
			StartIndex: i,
			EndIndex:   end,
			Status:     BatchPending,
		}

		if err := bp.processBatch(ctx, batch); err != nil {
			batch.Status = BatchFailed
			bp.metrics.failedBatches.Add(1)
			return fmt.Errorf("batch processing failed: %v", err)
		}

		batch.Status = BatchCompleted
		bp.metrics.processedBatches.Add(1)

		// Save checkpoint
		if bp.checkpointer != nil {
			bp.checkpointer.SaveCheckpoint(batch.ID, batch.EndIndex)
		}
	}

	return nil
}

// processBatch processes a single batch
func (bp *BatchProcessor) processBatch(ctx context.Context, batch *Batch) error {
	batch.Status = BatchProcessing

	for _, item := range batch.Items {
		// Validate
		if bp.validator != nil {
			if err := bp.validator.Validate(item); err != nil {
				batch.ErrorCount++
				bp.deadLetterQueue.Add(item)
				continue
			}
		}

		// Transform
		transformed := item
		if bp.transformer != nil {
			var err error
			transformed, err = bp.transformer.Transform(item)
			if err != nil {
				batch.ErrorCount++
				bp.deadLetterQueue.Add(item)
				continue
			}
		}

		// Load
		if bp.loader != nil {
			if err := bp.loader.Load([]interface{}{transformed}); err != nil {
				batch.ErrorCount++
				bp.deadLetterQueue.Add(item)
				continue
			}
		}

		bp.metrics.processedItems.Add(1)
	}

	batch.ProcessedAt = time.Now()
	batch.Status = BatchCheckpointed

	if batch.ErrorCount > 0 {
		bp.metrics.failedItems.Add(int64(batch.ErrorCount))
	}

	return nil
}

// GetMetrics returns batch processing metrics
func (bp *BatchProcessor) GetMetrics() map[string]interface{} {
	return map[string]interface{}{
		"total_batches":      bp.metrics.totalBatches.Load(),
		"processed_batches":  bp.metrics.processedBatches.Load(),
		"failed_batches":     bp.metrics.failedBatches.Load(),
		"total_items":        bp.metrics.totalItems.Load(),
		"processed_items":    bp.metrics.processedItems.Load(),
		"failed_items":       bp.metrics.failedItems.Load(),
	}
}

// Add adds an item to the dead letter queue
func (dlq *DeadLetterQueue) Add(item interface{}) {
	dlq.mu.Lock()
	defer dlq.mu.Unlock()
	dlq.items = append(dlq.items, item)
}

// GetItems returns all items in the dead letter queue
func (dlq *DeadLetterQueue) GetItems() []interface{} {
	dlq.mu.RLock()
	defer dlq.mu.RUnlock()

	result := make([]interface{}, len(dlq.items))
	copy(result, dlq.items)
	return result
}

// Clear clears the dead letter queue
func (dlq *DeadLetterQueue) Clear() {
	dlq.mu.Lock()
	defer dlq.mu.Unlock()
	dlq.items = make([]interface{}, 0)
}

// NewETLPipeline creates a new ETL pipeline
func NewETLPipeline(name string) *ETLPipeline {
	return &ETLPipeline{
		name:        name,
		transformers: make([]DataTransformer, 0),
		metrics:     &ETLMetrics{startTime: time.Now()},
		deadLetters: make([]*DeadLetter, 0),
	}
}

// SetExtractor sets the data extractor
func (ep *ETLPipeline) SetExtractor(e DataExtractor) {
	ep.extractor = e
}

// AddTransformer adds a transformer to the pipeline
func (ep *ETLPipeline) AddTransformer(t DataTransformer) {
	ep.transformers = append(ep.transformers, t)
}

// SetLoader sets the data loader
func (ep *ETLPipeline) SetLoader(l DataLoader) {
	ep.loader = l
}

// SetValidator sets the validator
func (ep *ETLPipeline) SetValidator(v DataValidator) {
	ep.validator = v
}

// SetErrorHandler sets the error handler
func (ep *ETLPipeline) SetErrorHandler(eh ErrorHandler) {
	ep.errorHandler = eh
}

// SetCheckpointer sets the checkpointer
func (ep *ETLPipeline) SetCheckpointer(c PipelineCheckpointer) {
	ep.checkpointer = c
}

// Execute executes the ETL pipeline
func (ep *ETLPipeline) Execute(ctx context.Context) error {
	if ep.extractor == nil || ep.loader == nil {
		return errors.New("extractor and loader must be set")
	}

	// Extract
	data, err := ep.extractor.Extract(ctx)
	if err != nil {
		return fmt.Errorf("extraction failed: %v", err)
	}

	ep.metrics.extractedCount.Store(int64(len(data)))

	// Transform
	transformed := make([]interface{}, 0)
	for _, item := range data {
		current := item

		// Validate
		if ep.validator != nil {
			if err := ep.validator.Validate(current); err != nil {
				if ep.errorHandler != nil {
					ep.errorHandler.HandleError(item, err)
				}
				continue
			}
		}

		// Apply transformers
		for _, transformer := range ep.transformers {
			var err error
			current, err = transformer.Transform(current)
			if err != nil {
				if ep.errorHandler != nil {
					ep.errorHandler.HandleError(item, err)
				}
				continue
			}
		}

		transformed = append(transformed, current)
		ep.metrics.transformedCount.Add(1)
	}

	// Load
	if err := ep.loader.Load(ctx, transformed); err != nil {
		return fmt.Errorf("loading failed: %v", err)
	}

	ep.metrics.loadedCount.Store(int64(len(transformed)))
	ep.metrics.endTime = time.Now()
	ep.metrics.totalDuration.Store(ep.metrics.endTime.Sub(ep.metrics.startTime).Milliseconds())

	return nil
}

// GetMetrics returns ETL metrics
func (ep *ETLPipeline) GetMetrics() map[string]interface{} {
	return map[string]interface{}{
		"extracted_count":  ep.metrics.extractedCount.Load(),
		"transformed_count": ep.metrics.transformedCount.Load(),
		"loaded_count":     ep.metrics.loadedCount.Load(),
		"error_count":      ep.metrics.errorCount.Load(),
		"duration_ms":      ep.metrics.totalDuration.Load(),
	}
}

// NewJobDAG creates a new job DAG
func NewJobDAG() *JobDAG {
	return &JobDAG{
		jobs:         make(map[string]*Job),
		dependencies: make(map[string][]string),
	}
}

// AddJob adds a job to the DAG
func (jd *JobDAG) AddJob(job *Job) {
	jd.mu.Lock()
	defer jd.mu.Unlock()
	jd.jobs[job.ID] = job
}

// AddDependency adds a dependency between jobs
func (jd *JobDAG) AddDependency(jobID, dependsOn string) {
	jd.mu.Lock()
	defer jd.mu.Unlock()

	if _, exists := jd.jobs[jobID]; !exists {
		return
	}
	if _, exists := jd.jobs[dependsOn]; !exists {
		return
	}

	jd.dependencies[jobID] = append(jd.dependencies[jobID], dependsOn)
}

// Execute executes the DAG
func (jd *JobDAG) Execute(ctx context.Context) error {
	jd.mu.RLock()
	defer jd.mu.RUnlock()

	// Topological sort
	sorted := jd.topologicalSort()

	for _, jobID := range sorted {
		job := jd.jobs[jobID]

		if err := job.Handler(ctx); err != nil {
			job.Status = StatusFailed
			job.LastError = err

			// Skip dependent jobs
			for depID, deps := range jd.dependencies {
				for _, dep := range deps {
					if dep == jobID {
						jd.jobs[depID].Status = StatusFailed
					}
				}
			}

			return fmt.Errorf("job %s failed: %v", jobID, err)
		}

		job.Status = StatusCompleted
	}

	return nil
}

// topologicalSort performs topological sort on the DAG
func (jd *JobDAG) topologicalSort() []string {
	visited := make(map[string]bool)
	var result []string

	var visit func(string)
	visit = func(jobID string) {
		if visited[jobID] {
			return
		}

		visited[jobID] = true

		for _, depID := range jd.dependencies[jobID] {
			visit(depID)
		}

		result = append(result, jobID)
	}

	for jobID := range jd.jobs {
		visit(jobID)
	}

	return result
}

// Simple implementations for testing

// SimpleDataSource implements DataSource
type SimpleDataSource struct {
	data []interface{}
}

func (sds *SimpleDataSource) Read() ([]interface{}, error) {
	return sds.data, nil
}

// SimpleTransformer implements Transformer
type SimpleTransformer struct{}

func (st *SimpleTransformer) Transform(item interface{}) (interface{}, error) {
	if m, ok := item.(map[string]interface{}); ok {
		m["transformed"] = true
		return m, nil
	}
	return item, nil
}

// SimpleLoader implements Loader
type SimpleLoader struct {
	loaded []interface{}
	mu     sync.Mutex
}

func (sl *SimpleLoader) Load(items []interface{}) error {
	sl.mu.Lock()
	defer sl.mu.Unlock()
	sl.loaded = append(sl.loaded, items...)
	return nil
}

// SimpleValidator implements Validator
type SimpleValidator struct{}

func (sv *SimpleValidator) Validate(item interface{}) error {
	if item == nil {
		return errors.New("item is nil")
	}
	return nil
}

// SimpleCheckpointer implements Checkpointer
type SimpleCheckpointer struct {
	checkpoints map[string]int
	mu          sync.Mutex
}

func (sc *SimpleCheckpointer) SaveCheckpoint(batchID string, index int) error {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	sc.checkpoints[batchID] = index
	return nil
}

func (sc *SimpleCheckpointer) LoadCheckpoint(batchID string) (int, error) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	return sc.checkpoints[batchID], nil
}

func main() {
	// Example usage
	scheduler := NewJobScheduler(4)

	job := &Job{
		ID:             "example-job",
		Name:           "Example Job",
		CronExpression: "0 0 * * *",
		Handler: func(ctx context.Context) error {
			fmt.Println("Job executed")
			return nil
		},
		Priority:   1,
		MaxRetries: 3,
		Timeout:    5 * time.Minute,
	}

	scheduler.RegisterJob(job)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	scheduler.Start(ctx)

	time.Sleep(2 * time.Second)
	scheduler.Stop()

	stats := scheduler.GetStats()
	fmt.Printf("Stats: %v\n", stats)
}
