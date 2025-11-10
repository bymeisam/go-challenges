# Challenge 136: File Processor

**Difficulty:** ⭐⭐⭐ Hard
**Time Estimate:** 45 minutes

## Description

Build a concurrent file processing system that processes multiple files in parallel using worker pools. This demonstrates Go's concurrency patterns, worker pools, and file I/O operations.

## Features

- **Worker Pool Pattern**: Configurable number of concurrent workers
- **Job Queue**: Channel-based task distribution
- **Multiple Operations**: Compress, encrypt, analyze files
- **Progress Tracking**: Real-time processing statistics
- **Error Handling**: Graceful error reporting per file
- **Batch Processing**: Process entire directories
- **Result Aggregation**: Collect and summarize results
- **Cancellation**: Context-based cancellation support

## Operations

- **Compress**: Gzip compression of files
- **Encrypt**: Simple XOR encryption (for demonstration)
- **Analyze**: Count lines, words, characters
- **Hash**: Calculate MD5/SHA256 checksums

## Requirements

1. Implement worker pool pattern
2. Support concurrent file processing
3. Track progress and statistics
4. Handle errors gracefully
5. Support context cancellation
6. Aggregate results from all workers
7. Configurable worker count

## Example Usage

```go
processor := NewFileProcessor(4) // 4 workers
files := []string{"file1.txt", "file2.txt", "file3.txt"}

results, err := processor.Process(context.Background(), files, OperationCompress)
for _, result := range results {
    fmt.Printf("%s: %v\n", result.Filename, result.Status)
}
```

## Learning Objectives

- Worker pool implementation
- Channel-based task distribution
- Concurrent file I/O
- Error handling in concurrent systems
- Context-based cancellation
- Result aggregation patterns
- Performance optimization with concurrency
