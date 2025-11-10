package main

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// Operation represents a file processing operation
type Operation string

const (
	OperationCompress Operation = "compress"
	OperationEncrypt  Operation = "encrypt"
	OperationAnalyze  Operation = "analyze"
	OperationHash     Operation = "hash"
)

// Job represents a file processing job
type Job struct {
	Filename  string
	Operation Operation
}

// Result represents the result of processing a file
type Result struct {
	Filename  string
	Operation Operation
	Success   bool
	Error     error
	Data      map[string]interface{}
	Duration  time.Duration
}

// Stats represents processing statistics
type Stats struct {
	TotalFiles     int
	ProcessedFiles int
	FailedFiles    int
	TotalDuration  time.Duration
	mu             sync.Mutex
}

func (s *Stats) Update(result Result) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.ProcessedFiles++
	s.TotalDuration += result.Duration
	if !result.Success {
		s.FailedFiles++
	}
}

func (s *Stats) String() string {
	s.mu.Lock()
	defer s.mu.Unlock()

	avgDuration := time.Duration(0)
	if s.ProcessedFiles > 0 {
		avgDuration = s.TotalDuration / time.Duration(s.ProcessedFiles)
	}

	return fmt.Sprintf(
		"Total: %d | Processed: %d | Failed: %d | Avg Time: %v",
		s.TotalFiles, s.ProcessedFiles, s.FailedFiles, avgDuration,
	)
}

// FileProcessor handles concurrent file processing
type FileProcessor struct {
	workerCount int
	stats       *Stats
}

// NewFileProcessor creates a new file processor
func NewFileProcessor(workerCount int) *FileProcessor {
	if workerCount <= 0 {
		workerCount = 1
	}

	return &FileProcessor{
		workerCount: workerCount,
		stats:       &Stats{},
	}
}

// Process processes files concurrently
func (fp *FileProcessor) Process(ctx context.Context, files []string, operation Operation) ([]Result, error) {
	fp.stats.TotalFiles = len(files)

	if len(files) == 0 {
		return nil, fmt.Errorf("no files to process")
	}

	// Create job queue
	jobs := make(chan Job, len(files))
	results := make(chan Result, len(files))

	// Start workers
	var wg sync.WaitGroup
	for i := 0; i < fp.workerCount; i++ {
		wg.Add(1)
		go fp.worker(ctx, &wg, jobs, results)
	}

	// Send jobs
	go func() {
		for _, file := range files {
			select {
			case <-ctx.Done():
				return
			case jobs <- Job{Filename: file, Operation: operation}:
			}
		}
		close(jobs)
	}()

	// Wait for workers to finish
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results
	allResults := make([]Result, 0, len(files))
	for result := range results {
		fp.stats.Update(result)
		allResults = append(allResults, result)
	}

	if ctx.Err() != nil {
		return allResults, ctx.Err()
	}

	return allResults, nil
}

// worker processes jobs from the queue
func (fp *FileProcessor) worker(ctx context.Context, wg *sync.WaitGroup, jobs <-chan Job, results chan<- Result) {
	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case job, ok := <-jobs:
			if !ok {
				return
			}

			result := fp.processFile(job)
			select {
			case <-ctx.Done():
				return
			case results <- result:
			}
		}
	}
}

// processFile processes a single file
func (fp *FileProcessor) processFile(job Job) Result {
	start := time.Now()

	result := Result{
		Filename:  job.Filename,
		Operation: job.Operation,
		Data:      make(map[string]interface{}),
		Duration:  0,
	}

	var err error
	switch job.Operation {
	case OperationCompress:
		err = fp.compressFile(job.Filename, result.Data)
	case OperationEncrypt:
		err = fp.encryptFile(job.Filename, result.Data)
	case OperationAnalyze:
		err = fp.analyzeFile(job.Filename, result.Data)
	case OperationHash:
		err = fp.hashFile(job.Filename, result.Data)
	default:
		err = fmt.Errorf("unknown operation: %s", job.Operation)
	}

	result.Duration = time.Since(start)
	result.Success = err == nil
	result.Error = err

	return result
}

// compressFile compresses a file using gzip
func (fp *FileProcessor) compressFile(filename string, data map[string]interface{}) error {
	input, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	if _, err := gw.Write(input); err != nil {
		return err
	}
	if err := gw.Close(); err != nil {
		return err
	}

	outputFile := filename + ".gz"
	if err := os.WriteFile(outputFile, buf.Bytes(), 0644); err != nil {
		return err
	}

	data["original_size"] = len(input)
	data["compressed_size"] = buf.Len()
	data["compression_ratio"] = float64(buf.Len()) / float64(len(input))
	data["output_file"] = outputFile

	return nil
}

// encryptFile encrypts a file using XOR cipher (simple demonstration)
func (fp *FileProcessor) encryptFile(filename string, data map[string]interface{}) error {
	input, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	// Simple XOR encryption with key
	key := []byte("encryption-key-example")
	encrypted := make([]byte, len(input))
	for i := range input {
		encrypted[i] = input[i] ^ key[i%len(key)]
	}

	outputFile := filename + ".enc"
	if err := os.WriteFile(outputFile, encrypted, 0644); err != nil {
		return err
	}

	data["original_size"] = len(input)
	data["encrypted_size"] = len(encrypted)
	data["output_file"] = outputFile

	return nil
}

// analyzeFile analyzes file content
func (fp *FileProcessor) analyzeFile(filename string, data map[string]interface{}) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	var lines, words, chars int
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		line := scanner.Text()
		lines++
		chars += len(line) + 1 // +1 for newline
		words += len(strings.Fields(line))
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	data["lines"] = lines
	data["words"] = words
	data["characters"] = chars

	return nil
}

// hashFile calculates file hashes
func (fp *FileProcessor) hashFile(filename string, data map[string]interface{}) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// MD5
	md5Hash := md5.New()
	if _, err := io.Copy(md5Hash, file); err != nil {
		return err
	}

	// SHA256
	file.Seek(0, 0)
	sha256Hash := sha256.New()
	if _, err := io.Copy(sha256Hash, file); err != nil {
		return err
	}

	data["md5"] = hex.EncodeToString(md5Hash.Sum(nil))
	data["sha256"] = hex.EncodeToString(sha256Hash.Sum(nil))

	return nil
}

// GetStats returns current processing statistics
func (fp *FileProcessor) GetStats() *Stats {
	return fp.stats
}

// ProcessDirectory processes all files in a directory
func (fp *FileProcessor) ProcessDirectory(ctx context.Context, dir string, operation Operation) ([]Result, error) {
	files, err := filepath.Glob(filepath.Join(dir, "*"))
	if err != nil {
		return nil, err
	}

	// Filter out directories
	regularFiles := make([]string, 0)
	for _, file := range files {
		info, err := os.Stat(file)
		if err != nil {
			continue
		}
		if !info.IsDir() {
			regularFiles = append(regularFiles, file)
		}
	}

	return fp.Process(ctx, regularFiles, operation)
}

func main() {
	// Example usage
	processor := NewFileProcessor(4)

	// Create test files
	testDir := "test_files"
	os.MkdirAll(testDir, 0755)
	defer os.RemoveAll(testDir)

	testFiles := []string{
		filepath.Join(testDir, "test1.txt"),
		filepath.Join(testDir, "test2.txt"),
		filepath.Join(testDir, "test3.txt"),
	}

	for i, file := range testFiles {
		content := fmt.Sprintf("This is test file number %d.\nIt contains multiple lines.\nFor testing purposes.\n", i+1)
		os.WriteFile(file, []byte(content), 0644)
	}

	fmt.Println("File Processor Demo")
	fmt.Println("===================")

	// Analyze files
	fmt.Println("\nAnalyzing files...")
	results, err := processor.Process(context.Background(), testFiles, OperationAnalyze)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	for _, result := range results {
		if result.Success {
			fmt.Printf("✓ %s: %d lines, %d words, %d chars (took %v)\n",
				filepath.Base(result.Filename),
				result.Data["lines"],
				result.Data["words"],
				result.Data["characters"],
				result.Duration,
			)
		} else {
			fmt.Printf("✗ %s: %v\n", filepath.Base(result.Filename), result.Error)
		}
	}

	// Compress files
	fmt.Println("\nCompressing files...")
	results, err = processor.Process(context.Background(), testFiles, OperationCompress)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	for _, result := range results {
		if result.Success {
			fmt.Printf("✓ %s: %.2f compression ratio (took %v)\n",
				filepath.Base(result.Filename),
				result.Data["compression_ratio"],
				result.Duration,
			)
		} else {
			fmt.Printf("✗ %s: %v\n", filepath.Base(result.Filename), result.Error)
		}
	}

	// Hash files
	fmt.Println("\nHashing files...")
	results, err = processor.Process(context.Background(), testFiles, OperationHash)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	for _, result := range results {
		if result.Success {
			fmt.Printf("✓ %s:\n  MD5: %s\n  SHA256: %s\n",
				filepath.Base(result.Filename),
				result.Data["md5"],
				result.Data["sha256"][:16]+"...",
			)
		} else {
			fmt.Printf("✗ %s: %v\n", filepath.Base(result.Filename), result.Error)
		}
	}

	fmt.Printf("\n%s\n", processor.GetStats())
}
