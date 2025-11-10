package main

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"
)

// Cron Scheduler Tests

func TestCronSchedulerParsing(t *testing.T) {
	tests := []struct {
		name       string
		expression string
		shouldFail bool
	}{
		{"valid every minute", "* * * * *", false},
		{"valid interval", "*/5 * * * *", false},
		{"valid range", "0-30 * * * *", false},
		{"valid list", "0,15,30,45 * * * *", false},
		{"invalid fields", "* * * *", true},
		{"invalid too many fields", "* * * * * *", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cs, err := NewCronScheduler(tt.expression)
			if tt.shouldFail && err == nil {
				t.Error("expected error")
			}
			if !tt.shouldFail && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if !tt.shouldFail && cs == nil {
				t.Error("scheduler should not be nil")
			}
		})
	}
}

func TestParseFieldInterval(t *testing.T) {
	result, err := parseField("*/5", 0, 59)
	if err != nil {
		t.Fatalf("parseField failed: %v", err)
	}

	expected := []int{0, 5, 10, 15, 20, 25, 30, 35, 40, 45, 50, 55}
	if len(result) != len(expected) {
		t.Errorf("expected %d values, got %d", len(expected), len(result))
	}

	for i, v := range result {
		if v != expected[i] {
			t.Errorf("value at index %d: expected %d, got %d", i, expected[i], v)
		}
	}
}

func TestParseFieldRange(t *testing.T) {
	result, err := parseField("1-5", 0, 59)
	if err != nil {
		t.Fatalf("parseField failed: %v", err)
	}

	expected := []int{1, 2, 3, 4, 5}
	if len(result) != len(expected) {
		t.Errorf("expected %d values, got %d", len(expected), len(result))
	}
}

func TestParseFieldList(t *testing.T) {
	result, err := parseField("1,3,5", 0, 59)
	if err != nil {
		t.Fatalf("parseField failed: %v", err)
	}

	expected := []int{1, 3, 5}
	if len(result) != len(expected) {
		t.Errorf("expected %d values, got %d", len(expected), len(result))
	}
}

func TestNextRun(t *testing.T) {
	cs, err := NewCronScheduler("*/5 * * * *")
	if err != nil {
		t.Fatalf("NewCronScheduler failed: %v", err)
	}

	now := time.Date(2025, 1, 1, 12, 0, 0, 0, time.Local)
	next := cs.NextRun(now)

	if next.Minute() != 5 && next.Minute() != 0 {
		t.Errorf("expected next run at minute 0 or 5, got %d", next.Minute())
	}
}

func TestNextRunDaily(t *testing.T) {
	cs, err := NewCronScheduler("0 0 * * *")
	if err != nil {
		t.Fatalf("NewCronScheduler failed: %v", err)
	}

	now := time.Date(2025, 1, 1, 12, 0, 0, 0, time.Local)
	next := cs.NextRun(now)

	if next.Hour() != 0 || next.Minute() != 0 {
		t.Errorf("expected next run at 00:00, got %02d:%02d", next.Hour(), next.Minute())
	}
}

// Job Scheduler Tests

func TestJobSchedulerCreation(t *testing.T) {
	js := NewJobScheduler(4)

	if js.workers != 4 {
		t.Errorf("expected 4 workers, got %d", js.workers)
	}

	if js.jobs == nil {
		t.Error("jobs map should be initialized")
	}
}

func TestJobRegistration(t *testing.T) {
	js := NewJobScheduler(2)

	job := &Job{
		ID:   "test-job",
		Name: "Test Job",
		Handler: func(ctx context.Context) error {
			return nil
		},
	}

	err := js.RegisterJob(job)
	if err != nil {
		t.Fatalf("RegisterJob failed: %v", err)
	}

	retrieved := js.GetJob("test-job")
	if retrieved == nil {
		t.Error("job not found")
	}

	if retrieved.ID != "test-job" {
		t.Errorf("expected ID test-job, got %s", retrieved.ID)
	}
}

func TestJobRegistrationError(t *testing.T) {
	js := NewJobScheduler(2)

	job := &Job{
		ID: "",
		Handler: func(ctx context.Context) error {
			return nil
		},
	}

	err := js.RegisterJob(job)
	if err == nil {
		t.Error("expected error for empty job ID")
	}
}

func TestJobExecution(t *testing.T) {
	js := NewJobScheduler(1)

	executed := false
	job := &Job{
		ID:   "test-job",
		Name: "Test Job",
		Handler: func(ctx context.Context) error {
			executed = true
			return nil
		},
		MaxRetries: 1,
		Timeout:    5 * time.Second,
	}

	js.RegisterJob(job)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	js.Start(ctx)
	js.executeJob(ctx, job)
	js.Stop()

	if !executed {
		t.Error("job was not executed")
	}

	if job.Status != StatusCompleted {
		t.Errorf("expected status completed, got %s", job.Status)
	}
}

func TestJobExecutionFailure(t *testing.T) {
	js := NewJobScheduler(1)

	job := &Job{
		ID:   "test-job",
		Name: "Test Job",
		Handler: func(ctx context.Context) error {
			return errors.New("job failed")
		},
		MaxRetries: 0,
		Timeout:    5 * time.Second,
	}

	js.RegisterJob(job)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	js.Start(ctx)
	js.executeJob(ctx, job)
	js.Stop()

	if job.Status != StatusFailed {
		t.Errorf("expected status failed, got %s", job.Status)
	}

	if job.LastError == nil {
		t.Error("last error should be set")
	}
}

func TestJobRetry(t *testing.T) {
	js := NewJobScheduler(1)

	attempts := 0
	job := &Job{
		ID:         "test-job",
		Name:       "Test Job",
		MaxRetries: 2,
		Timeout:    5 * time.Second,
		Handler: func(ctx context.Context) error {
			attempts++
			if attempts < 2 {
				return errors.New("fail")
			}
			return nil
		},
	}

	js.RegisterJob(job)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	js.Start(ctx)

	// First execution
	js.executeJob(ctx, job)
	if job.RetryCount != 1 {
		t.Errorf("expected retry count 1, got %d", job.RetryCount)
	}

	// Retry
	js.executeJob(ctx, job)
	if job.Status != StatusCompleted {
		t.Errorf("expected status completed, got %s", job.Status)
	}

	js.Stop()
}

func TestJobStats(t *testing.T) {
	js := NewJobScheduler(1)

	job := &Job{
		ID:   "test-job",
		Name: "Test Job",
		Handler: func(ctx context.Context) error {
			return nil
		},
		MaxRetries: 0,
		Timeout:    5 * time.Second,
	}

	js.RegisterJob(job)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	js.Start(ctx)
	js.executeJob(ctx, job)
	js.Stop()

	stats := js.GetStats()

	if stats["completed_jobs"] != int64(1) {
		t.Errorf("expected 1 completed job, got %v", stats["completed_jobs"])
	}
}

// Batch Processor Tests

func TestBatchProcessorCreation(t *testing.T) {
	bp := NewBatchProcessor(100, 4)

	if bp.batchSize != 100 {
		t.Errorf("expected batch size 100, got %d", bp.batchSize)
	}

	if bp.workers != 4 {
		t.Errorf("expected 4 workers, got %d", bp.workers)
	}

	if bp.deadLetterQueue == nil {
		t.Error("dead letter queue should be initialized")
	}
}

func TestBatchProcessing(t *testing.T) {
	bp := NewBatchProcessor(10, 1)

	data := make([]interface{}, 25)
	for i := range data {
		data[i] = i
	}

	ds := &SimpleDataSource{data: data}
	loader := &SimpleLoader{loaded: make([]interface{}, 0)}
	validator := &SimpleValidator{}
	transformer := &SimpleTransformer{}

	bp.SetDataSource(ds)
	bp.SetLoader(loader)
	bp.SetValidator(validator)
	bp.SetTransformer(transformer)

	ctx := context.Background()
	err := bp.Process(ctx)

	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	metrics := bp.GetMetrics()
	if metrics["total_items"].(int64) != 25 {
		t.Errorf("expected 25 total items, got %v", metrics["total_items"])
	}
}

func TestBatchProcessingWithCheckpoint(t *testing.T) {
	bp := NewBatchProcessor(10, 1)

	data := make([]interface{}, 25)
	for i := range data {
		data[i] = i
	}

	ds := &SimpleDataSource{data: data}
	loader := &SimpleLoader{loaded: make([]interface{}, 0)}
	checkpointer := &SimpleCheckpointer{checkpoints: make(map[string]int)}

	bp.SetDataSource(ds)
	bp.SetLoader(loader)
	bp.SetCheckpointer(checkpointer)

	ctx := context.Background()
	err := bp.Process(ctx)

	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	if len(checkpointer.checkpoints) == 0 {
		t.Error("checkpoints should be saved")
	}
}

func TestBatchProcessingNoDataSource(t *testing.T) {
	bp := NewBatchProcessor(10, 1)

	ctx := context.Background()
	err := bp.Process(ctx)

	if err == nil {
		t.Error("expected error when data source not set")
	}
}

func TestDeadLetterQueue(t *testing.T) {
	dlq := &DeadLetterQueue{items: make([]interface{}, 0)}

	dlq.Add("item1")
	dlq.Add("item2")

	items := dlq.GetItems()
	if len(items) != 2 {
		t.Errorf("expected 2 items, got %d", len(items))
	}

	dlq.Clear()
	items = dlq.GetItems()
	if len(items) != 0 {
		t.Error("dead letter queue should be empty after clear")
	}
}

// ETL Pipeline Tests

func TestETLPipelineCreation(t *testing.T) {
	pipeline := NewETLPipeline("test-pipeline")

	if pipeline.name != "test-pipeline" {
		t.Errorf("expected name test-pipeline, got %s", pipeline.name)
	}

	if pipeline.metrics == nil {
		t.Error("metrics should be initialized")
	}
}

func TestETLPipelineExecution(t *testing.T) {
	pipeline := NewETLPipeline("test-pipeline")

	data := []interface{}{
		map[string]interface{}{"name": "John"},
		map[string]interface{}{"name": "Jane"},
	}

	extractor := &TestExtractor{data: data}
	loader := &TestLoader{loaded: make([]interface{}, 0)}

	pipeline.SetExtractor(extractor)
	pipeline.SetLoader(loader)

	ctx := context.Background()
	err := pipeline.Execute(ctx)

	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	metrics := pipeline.GetMetrics()
	if metrics["extracted_count"].(int64) != 2 {
		t.Errorf("expected 2 extracted items, got %v", metrics["extracted_count"])
	}
}

func TestETLPipelineWithTransformer(t *testing.T) {
	pipeline := NewETLPipeline("test-pipeline")

	data := []interface{}{
		map[string]interface{}{"name": "John"},
	}

	extractor := &TestExtractor{data: data}
	loader := &TestLoader{loaded: make([]interface{}, 0)}
	transformer := &SimpleTransformer{}

	pipeline.SetExtractor(extractor)
	pipeline.AddTransformer(transformer)
	pipeline.SetLoader(loader)

	ctx := context.Background()
	err := pipeline.Execute(ctx)

	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	metrics := pipeline.GetMetrics()
	if metrics["transformed_count"].(int64) != 1 {
		t.Errorf("expected 1 transformed item, got %v", metrics["transformed_count"])
	}
}

func TestETLPipelineValidation(t *testing.T) {
	pipeline := NewETLPipeline("test-pipeline")

	data := []interface{}{
		nil,
		map[string]interface{}{"name": "John"},
	}

	extractor := &TestExtractor{data: data}
	loader := &TestLoader{loaded: make([]interface{}, 0)}
	validator := &SimpleValidator{}

	pipeline.SetExtractor(extractor)
	pipeline.SetValidator(validator)
	pipeline.SetLoader(loader)

	ctx := context.Background()
	err := pipeline.Execute(ctx)

	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// Should skip invalid item
	metrics := pipeline.GetMetrics()
	if metrics["loaded_count"].(int64) != 1 {
		t.Errorf("expected 1 loaded item, got %v", metrics["loaded_count"])
	}
}

func TestETLPipelineNoExtractor(t *testing.T) {
	pipeline := NewETLPipeline("test-pipeline")
	pipeline.SetLoader(&TestLoader{})

	ctx := context.Background()
	err := pipeline.Execute(ctx)

	if err == nil {
		t.Error("expected error when extractor not set")
	}
}

// Job DAG Tests

func TestJobDAGCreation(t *testing.T) {
	dag := NewJobDAG()

	if dag.jobs == nil {
		t.Error("jobs map should be initialized")
	}

	if dag.dependencies == nil {
		t.Error("dependencies map should be initialized")
	}
}

func TestJobDAGAddJob(t *testing.T) {
	dag := NewJobDAG()

	job := &Job{
		ID: "job-1",
		Handler: func(ctx context.Context) error {
			return nil
		},
	}

	dag.AddJob(job)

	if len(dag.jobs) != 1 {
		t.Errorf("expected 1 job, got %d", len(dag.jobs))
	}
}

func TestJobDAGDependency(t *testing.T) {
	dag := NewJobDAG()

	job1 := &Job{ID: "job-1", Handler: func(ctx context.Context) error { return nil }}
	job2 := &Job{ID: "job-2", Handler: func(ctx context.Context) error { return nil }}

	dag.AddJob(job1)
	dag.AddJob(job2)
	dag.AddDependency("job-2", "job-1")

	deps := dag.dependencies["job-2"]
	if len(deps) != 1 || deps[0] != "job-1" {
		t.Error("dependency not set correctly")
	}
}

func TestJobDAGExecution(t *testing.T) {
	dag := NewJobDAG()

	executed := []string{}

	job1 := &Job{
		ID: "job-1",
		Handler: func(ctx context.Context) error {
			executed = append(executed, "job-1")
			return nil
		},
	}

	job2 := &Job{
		ID: "job-2",
		Handler: func(ctx context.Context) error {
			executed = append(executed, "job-2")
			return nil
		},
	}

	dag.AddJob(job1)
	dag.AddJob(job2)
	dag.AddDependency("job-2", "job-1")

	ctx := context.Background()
	err := dag.Execute(ctx)

	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if len(executed) != 2 {
		t.Errorf("expected 2 executed jobs, got %d", len(executed))
	}

	// job-1 should execute before job-2
	if executed[0] != "job-1" {
		t.Errorf("expected job-1 to execute first, got %s", executed[0])
	}
}

func TestJobDAGExecutionFailure(t *testing.T) {
	dag := NewJobDAG()

	job1 := &Job{
		ID: "job-1",
		Handler: func(ctx context.Context) error {
			return errors.New("job failed")
		},
	}

	job2 := &Job{
		ID: "job-2",
		Handler: func(ctx context.Context) error {
			return nil
		},
	}

	dag.AddJob(job1)
	dag.AddJob(job2)
	dag.AddDependency("job-2", "job-1")

	ctx := context.Background()
	err := dag.Execute(ctx)

	if err == nil {
		t.Error("expected error from failed job")
	}

	if job1.Status != StatusFailed {
		t.Errorf("expected job-1 status failed, got %s", job1.Status)
	}

	if job2.Status != StatusFailed {
		t.Errorf("expected job-2 status failed (skipped), got %s", job2.Status)
	}
}

// Test Helpers

type TestExtractor struct {
	data []interface{}
}

func (te *TestExtractor) Extract(ctx context.Context) ([]interface{}, error) {
	return te.data, nil
}

type TestLoader struct {
	loaded []interface{}
}

func (tl *TestLoader) Load(ctx context.Context, items []interface{}) error {
	tl.loaded = append(tl.loaded, items...)
	return nil
}

type TestTransformer struct {
	transformCount int
}

func (tt *TestTransformer) Transform(item interface{}) (interface{}, error) {
	tt.transformCount++
	return item, nil
}

// Benchmark Tests

func BenchmarkCronSchedulerNextRun(b *testing.B) {
	cs, _ := NewCronScheduler("*/5 * * * *")
	now := time.Now()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cs.NextRun(now)
	}
}

func BenchmarkJobSchedulerExecute(b *testing.B) {
	js := NewJobScheduler(4)

	job := &Job{
		ID:         "bench-job",
		Handler:    func(ctx context.Context) error { return nil },
		MaxRetries: 1,
		Timeout:    5 * time.Second,
	}

	js.RegisterJob(job)

	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		js.executeJob(ctx, job)
	}
}

func BenchmarkBatchProcessing(b *testing.B) {
	data := make([]interface{}, 1000)
	for i := range data {
		data[i] = i
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bp := NewBatchProcessor(100, 4)
		bp.SetDataSource(&SimpleDataSource{data: data})
		bp.SetLoader(&SimpleLoader{})

		ctx := context.Background()
		bp.Process(ctx)
	}
}

func BenchmarkETLPipelineExecution(b *testing.B) {
	data := make([]interface{}, 1000)
	for i := range data {
		data[i] = map[string]interface{}{"id": i}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pipeline := NewETLPipeline("bench")
		pipeline.SetExtractor(&TestExtractor{data: data})
		pipeline.SetLoader(&TestLoader{})

		ctx := context.Background()
		pipeline.Execute(ctx)
	}
}

func BenchmarkJobDAGExecution(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dag := NewJobDAG()

		for j := 0; j < 10; j++ {
			job := &Job{
				ID:      fmt.Sprintf("job-%d", j),
				Handler: func(ctx context.Context) error { return nil },
			}
			dag.AddJob(job)
		}

		ctx := context.Background()
		dag.Execute(ctx)
	}
}

