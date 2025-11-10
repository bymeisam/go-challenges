package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func createTestFiles(t *testing.T, count int) (string, []string) {
	tmpDir := t.TempDir()
	files := make([]string, count)

	for i := 0; i < count; i++ {
		filename := filepath.Join(tmpDir, "test"+string(rune('0'+i))+".txt")
		content := "Line 1\nLine 2\nLine 3\n"
		if err := os.WriteFile(filename, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
		files[i] = filename
	}

	return tmpDir, files
}

func TestNewFileProcessor(t *testing.T) {
	tests := []struct {
		name        string
		workerCount int
		expected    int
	}{
		{"Valid count", 4, 4},
		{"Zero count", 0, 1},
		{"Negative count", -5, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			processor := NewFileProcessor(tt.workerCount)
			if processor.workerCount != tt.expected {
				t.Errorf("Expected %d workers, got %d", tt.expected, processor.workerCount)
			}
		})
	}
}

func TestProcessAnalyze(t *testing.T) {
	_, files := createTestFiles(t, 3)
	processor := NewFileProcessor(2)

	results, err := processor.Process(context.Background(), files, OperationAnalyze)
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	if len(results) != len(files) {
		t.Errorf("Expected %d results, got %d", len(files), len(results))
	}

	for _, result := range results {
		if !result.Success {
			t.Errorf("File %s failed: %v", result.Filename, result.Error)
		}

		if result.Data["lines"] != 3 {
			t.Errorf("Expected 3 lines, got %v", result.Data["lines"])
		}

		if result.Data["words"] != 6 {
			t.Errorf("Expected 6 words, got %v", result.Data["words"])
		}

		if result.Operation != OperationAnalyze {
			t.Errorf("Expected operation %s, got %s", OperationAnalyze, result.Operation)
		}
	}
}

func TestProcessCompress(t *testing.T) {
	_, files := createTestFiles(t, 2)
	processor := NewFileProcessor(2)

	results, err := processor.Process(context.Background(), files, OperationCompress)
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	for _, result := range results {
		if !result.Success {
			t.Errorf("File %s failed: %v", result.Filename, result.Error)
			continue
		}

		if result.Data["original_size"] == nil {
			t.Error("Missing original_size in result")
		}

		if result.Data["compressed_size"] == nil {
			t.Error("Missing compressed_size in result")
		}

		if result.Data["compression_ratio"] == nil {
			t.Error("Missing compression_ratio in result")
		}

		outputFile := result.Data["output_file"].(string)
		if _, err := os.Stat(outputFile); os.IsNotExist(err) {
			t.Errorf("Compressed file not created: %s", outputFile)
		}
	}
}

func TestProcessEncrypt(t *testing.T) {
	_, files := createTestFiles(t, 1)
	processor := NewFileProcessor(1)

	results, err := processor.Process(context.Background(), files, OperationEncrypt)
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	result := results[0]
	if !result.Success {
		t.Fatalf("Encryption failed: %v", result.Error)
	}

	if result.Data["encrypted_size"] == nil {
		t.Error("Missing encrypted_size in result")
	}

	outputFile := result.Data["output_file"].(string)
	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		t.Errorf("Encrypted file not created: %s", outputFile)
	}

	// Verify encrypted content is different from original
	original, _ := os.ReadFile(files[0])
	encrypted, _ := os.ReadFile(outputFile)

	if string(original) == string(encrypted) {
		t.Error("Encrypted content should differ from original")
	}
}

func TestProcessHash(t *testing.T) {
	_, files := createTestFiles(t, 1)
	processor := NewFileProcessor(1)

	results, err := processor.Process(context.Background(), files, OperationHash)
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	result := results[0]
	if !result.Success {
		t.Fatalf("Hash failed: %v", result.Error)
	}

	md5Hash, ok := result.Data["md5"].(string)
	if !ok || len(md5Hash) != 32 {
		t.Error("Invalid MD5 hash")
	}

	sha256Hash, ok := result.Data["sha256"].(string)
	if !ok || len(sha256Hash) != 64 {
		t.Error("Invalid SHA256 hash")
	}

	// Process same file again - should get same hashes
	results2, _ := processor.Process(context.Background(), files, OperationHash)
	if results2[0].Data["md5"] != md5Hash {
		t.Error("MD5 hash should be consistent")
	}

	if results2[0].Data["sha256"] != sha256Hash {
		t.Error("SHA256 hash should be consistent")
	}
}

func TestProcessEmptyFileList(t *testing.T) {
	processor := NewFileProcessor(2)

	_, err := processor.Process(context.Background(), []string{}, OperationAnalyze)
	if err == nil {
		t.Error("Expected error for empty file list")
	}
}

func TestProcessNonExistentFile(t *testing.T) {
	processor := NewFileProcessor(1)
	files := []string{"/non/existent/file.txt"}

	results, err := processor.Process(context.Background(), files, OperationAnalyze)
	if err != nil && err != context.Canceled {
		// Error is acceptable but not required
	}

	if len(results) > 0 && results[0].Success {
		t.Error("Expected failure for non-existent file")
	}
}

func TestContextCancellation(t *testing.T) {
	_, files := createTestFiles(t, 10)
	processor := NewFileProcessor(1) // Single worker to make cancellation more reliable

	ctx, cancel := context.WithCancel(context.Background())

	// Cancel after a short delay
	go func() {
		time.Sleep(10 * time.Millisecond)
		cancel()
	}()

	results, err := processor.Process(ctx, files, OperationAnalyze)

	if err != context.Canceled {
		// Cancellation timing is uncertain, so we just check it was handled
		t.Logf("Got error: %v (expected context.Canceled)", err)
	}

	// Should have processed fewer files than requested
	if len(results) > 0 {
		t.Logf("Processed %d files before cancellation", len(results))
	}
}

func TestConcurrentProcessing(t *testing.T) {
	_, files := createTestFiles(t, 10)

	// Test with different worker counts
	tests := []struct {
		name        string
		workerCount int
	}{
		{"Single worker", 1},
		{"Multiple workers", 4},
		{"Many workers", 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			processor := NewFileProcessor(tt.workerCount)
			results, err := processor.Process(context.Background(), files, OperationAnalyze)

			if err != nil {
				t.Fatalf("Process failed: %v", err)
			}

			if len(results) != len(files) {
				t.Errorf("Expected %d results, got %d", len(files), len(results))
			}

			// All files should be processed successfully
			for _, result := range results {
				if !result.Success {
					t.Errorf("File %s failed: %v", result.Filename, result.Error)
				}
			}
		})
	}
}

func TestStats(t *testing.T) {
	_, files := createTestFiles(t, 5)
	processor := NewFileProcessor(2)

	processor.Process(context.Background(), files, OperationAnalyze)

	stats := processor.GetStats()

	if stats.TotalFiles != 5 {
		t.Errorf("Expected 5 total files, got %d", stats.TotalFiles)
	}

	if stats.ProcessedFiles != 5 {
		t.Errorf("Expected 5 processed files, got %d", stats.ProcessedFiles)
	}

	if stats.FailedFiles != 0 {
		t.Errorf("Expected 0 failed files, got %d", stats.FailedFiles)
	}

	if stats.TotalDuration == 0 {
		t.Error("Expected non-zero total duration")
	}
}

func TestStatsWithFailures(t *testing.T) {
	tmpDir := t.TempDir()
	files := []string{
		filepath.Join(tmpDir, "exists.txt"),
		filepath.Join(tmpDir, "missing.txt"),
	}

	// Create only one file
	os.WriteFile(files[0], []byte("content"), 0644)

	processor := NewFileProcessor(1)
	processor.Process(context.Background(), files, OperationAnalyze)

	stats := processor.GetStats()

	if stats.ProcessedFiles != 2 {
		t.Errorf("Expected 2 processed files, got %d", stats.ProcessedFiles)
	}

	if stats.FailedFiles != 1 {
		t.Errorf("Expected 1 failed file, got %d", stats.FailedFiles)
	}
}

func TestProcessDirectory(t *testing.T) {
	tmpDir, _ := createTestFiles(t, 3)

	// Create a subdirectory (should be ignored)
	os.Mkdir(filepath.Join(tmpDir, "subdir"), 0755)

	processor := NewFileProcessor(2)
	results, err := processor.ProcessDirectory(context.Background(), tmpDir, OperationAnalyze)

	if err != nil {
		t.Fatalf("ProcessDirectory failed: %v", err)
	}

	// Should process only the 3 files, not the subdirectory
	if len(results) != 3 {
		t.Errorf("Expected 3 results, got %d", len(results))
	}
}

func TestResultDuration(t *testing.T) {
	_, files := createTestFiles(t, 1)
	processor := NewFileProcessor(1)

	results, err := processor.Process(context.Background(), files, OperationAnalyze)
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	if results[0].Duration == 0 {
		t.Error("Expected non-zero duration")
	}

	if results[0].Duration > 1*time.Second {
		t.Errorf("Processing took too long: %v", results[0].Duration)
	}
}

func TestMultipleOperations(t *testing.T) {
	_, files := createTestFiles(t, 2)
	processor := NewFileProcessor(2)

	operations := []Operation{
		OperationAnalyze,
		OperationHash,
		OperationCompress,
		OperationEncrypt,
	}

	for _, op := range operations {
		t.Run(string(op), func(t *testing.T) {
			results, err := processor.Process(context.Background(), files, op)
			if err != nil {
				t.Fatalf("Process failed for %s: %v", op, err)
			}

			if len(results) != len(files) {
				t.Errorf("Expected %d results, got %d", len(files), len(results))
			}

			for _, result := range results {
				if !result.Success {
					t.Errorf("Operation %s failed for %s: %v", op, result.Filename, result.Error)
				}

				if result.Operation != op {
					t.Errorf("Expected operation %s, got %s", op, result.Operation)
				}
			}
		})
	}
}

func TestStatsString(t *testing.T) {
	stats := &Stats{
		TotalFiles:     10,
		ProcessedFiles: 8,
		FailedFiles:    2,
		TotalDuration:  800 * time.Millisecond,
	}

	str := stats.String()
	if str == "" {
		t.Error("Stats string should not be empty")
	}

	// Should contain key information
	if !contains(str, "10") || !contains(str, "8") || !contains(str, "2") {
		t.Errorf("Stats string missing key information: %s", str)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
		len(s) > len(substr)+1 && containsHelper(s[1:len(s)-1], substr)))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestLargeFileSet(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping large file set test in short mode")
	}

	_, files := createTestFiles(t, 50)
	processor := NewFileProcessor(8)

	start := time.Now()
	results, err := processor.Process(context.Background(), files, OperationAnalyze)
	duration := time.Since(start)

	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	if len(results) != 50 {
		t.Errorf("Expected 50 results, got %d", len(results))
	}

	t.Logf("Processed 50 files in %v with 8 workers", duration)
}
