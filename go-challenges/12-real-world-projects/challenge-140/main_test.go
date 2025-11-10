package main

import (
	"encoding/json"
	"testing"
	"time"
)

func TestNewLogAggregator(t *testing.T) {
	aggregator := NewLogAggregator()

	if aggregator == nil {
		t.Fatal("Aggregator should not be nil")
	}

	if aggregator.logs == nil {
		t.Error("Logs slice should be initialized")
	}

	if aggregator.idCounter != 1 {
		t.Errorf("ID counter should start at 1, got %d", aggregator.idCounter)
	}
}

func TestIngest(t *testing.T) {
	aggregator := NewLogAggregator()

	entry := LogEntry{
		Level:   LevelInfo,
		Message: "Test message",
		Source:  "test",
	}

	id := aggregator.Ingest(entry)

	if id == "" {
		t.Error("ID should be generated")
	}

	if len(aggregator.logs) != 1 {
		t.Errorf("Expected 1 log, got %d", len(aggregator.logs))
	}

	storedLog := aggregator.logs[0]
	if storedLog.Message != "Test message" {
		t.Errorf("Expected message 'Test message', got '%s'", storedLog.Message)
	}

	if storedLog.Level != LevelInfo {
		t.Errorf("Expected level INFO, got %s", storedLog.Level)
	}

	if storedLog.Timestamp.IsZero() {
		t.Error("Timestamp should be set")
	}
}

func TestIngestWithID(t *testing.T) {
	aggregator := NewLogAggregator()

	entry := LogEntry{
		ID:      "custom-id",
		Level:   LevelError,
		Message: "Test",
	}

	id := aggregator.Ingest(entry)

	if id != "custom-id" {
		t.Errorf("Expected ID 'custom-id', got '%s'", id)
	}
}

func TestIngestDefaultLevel(t *testing.T) {
	aggregator := NewLogAggregator()

	entry := LogEntry{
		Message: "Test",
	}

	aggregator.Ingest(entry)

	if aggregator.logs[0].Level != LevelInfo {
		t.Errorf("Expected default level INFO, got %s", aggregator.logs[0].Level)
	}
}

func TestSearch(t *testing.T) {
	aggregator := NewLogAggregator()

	// Add test logs
	entries := []LogEntry{
		{Level: LevelInfo, Message: "Info message", Source: "app"},
		{Level: LevelError, Message: "Error message", Source: "db"},
		{Level: LevelWarn, Message: "Warning message", Source: "app"},
	}

	for _, entry := range entries {
		aggregator.Ingest(entry)
	}

	// Search by level
	query := SearchQuery{Level: "ERROR"}
	result := aggregator.Search(query)

	if result.Total != 1 {
		t.Errorf("Expected 1 result, got %d", result.Total)
	}

	if len(result.Logs) != 1 {
		t.Errorf("Expected 1 log, got %d", len(result.Logs))
	}

	if result.Logs[0].Level != LevelError {
		t.Error("Expected ERROR level log")
	}
}

func TestSearchBySource(t *testing.T) {
	aggregator := NewLogAggregator()

	aggregator.Ingest(LogEntry{Level: LevelInfo, Message: "App log", Source: "app"})
	aggregator.Ingest(LogEntry{Level: LevelInfo, Message: "DB log", Source: "db"})
	aggregator.Ingest(LogEntry{Level: LevelInfo, Message: "App log 2", Source: "app"})

	query := SearchQuery{Source: "app"}
	result := aggregator.Search(query)

	if result.Total != 2 {
		t.Errorf("Expected 2 results, got %d", result.Total)
	}
}

func TestSearchByMessage(t *testing.T) {
	aggregator := NewLogAggregator()

	aggregator.Ingest(LogEntry{Message: "Database connection failed"})
	aggregator.Ingest(LogEntry{Message: "User logged in"})
	aggregator.Ingest(LogEntry{Message: "Database query executed"})

	query := SearchQuery{Message: "database"}
	result := aggregator.Search(query)

	if result.Total != 2 {
		t.Errorf("Expected 2 results for 'database', got %d", result.Total)
	}
}

func TestSearchByTimeRange(t *testing.T) {
	aggregator := NewLogAggregator()

	now := time.Now()
	past := now.Add(-1 * time.Hour)
	future := now.Add(1 * time.Hour)

	aggregator.logs = []LogEntry{
		{ID: "1", Timestamp: past, Message: "Old log"},
		{ID: "2", Timestamp: now, Message: "Current log"},
		{ID: "3", Timestamp: future, Message: "Future log"},
	}

	query := SearchQuery{
		StartTime: past.Add(30 * time.Minute),
		EndTime:   future.Add(-30 * time.Minute),
	}

	result := aggregator.Search(query)

	if result.Total != 1 {
		t.Errorf("Expected 1 result in time range, got %d", result.Total)
	}

	if result.Logs[0].ID != "2" {
		t.Error("Expected current log in results")
	}
}

func TestSearchPagination(t *testing.T) {
	aggregator := NewLogAggregator()

	// Add 15 logs
	for i := 0; i < 15; i++ {
		aggregator.Ingest(LogEntry{Message: "Log message"})
	}

	// First page
	query := SearchQuery{Limit: 5, Offset: 0}
	result := aggregator.Search(query)

	if len(result.Logs) != 5 {
		t.Errorf("Expected 5 logs, got %d", len(result.Logs))
	}

	if result.Total != 15 {
		t.Errorf("Expected total 15, got %d", result.Total)
	}

	// Second page
	query.Offset = 5
	result = aggregator.Search(query)

	if len(result.Logs) != 5 {
		t.Errorf("Expected 5 logs on second page, got %d", len(result.Logs))
	}

	// Third page (partial)
	query.Offset = 10
	result = aggregator.Search(query)

	if len(result.Logs) != 5 {
		t.Errorf("Expected 5 logs on third page, got %d", len(result.Logs))
	}
}

func TestSearchDefaultLimit(t *testing.T) {
	aggregator := NewLogAggregator()

	query := SearchQuery{}
	result := aggregator.Search(query)

	if result.Limit != 100 {
		t.Errorf("Expected default limit 100, got %d", result.Limit)
	}
}

func TestGetStats(t *testing.T) {
	aggregator := NewLogAggregator()

	// Add various logs
	aggregator.Ingest(LogEntry{Level: LevelInfo, Source: "app"})
	aggregator.Ingest(LogEntry{Level: LevelError, Source: "db"})
	aggregator.Ingest(LogEntry{Level: LevelError, Source: "app"})
	aggregator.Ingest(LogEntry{Level: LevelWarn, Source: "app"})

	stats := aggregator.GetStats()

	if stats.TotalLogs != 4 {
		t.Errorf("Expected 4 total logs, got %d", stats.TotalLogs)
	}

	if stats.ByLevel[LevelInfo] != 1 {
		t.Errorf("Expected 1 INFO log, got %d", stats.ByLevel[LevelInfo])
	}

	if stats.ByLevel[LevelError] != 2 {
		t.Errorf("Expected 2 ERROR logs, got %d", stats.ByLevel[LevelError])
	}

	if stats.BySource["app"] != 3 {
		t.Errorf("Expected 3 logs from 'app', got %d", stats.BySource["app"])
	}

	if stats.BySource["db"] != 1 {
		t.Errorf("Expected 1 log from 'db', got %d", stats.BySource["db"])
	}
}

func TestGetStatsEmpty(t *testing.T) {
	aggregator := NewLogAggregator()

	stats := aggregator.GetStats()

	if stats.TotalLogs != 0 {
		t.Errorf("Expected 0 total logs, got %d", stats.TotalLogs)
	}

	if len(stats.ByLevel) != 0 {
		t.Error("ByLevel should be empty")
	}
}

func TestGetStatsTimeRange(t *testing.T) {
	aggregator := NewLogAggregator()

	now := time.Now()
	aggregator.logs = []LogEntry{
		{Timestamp: now.Add(-2 * time.Hour)},
		{Timestamp: now},
		{Timestamp: now.Add(-1 * time.Hour)},
	}

	stats := aggregator.GetStats()

	if stats.TimeRange.Earliest.After(now.Add(-2 * time.Hour)) {
		t.Error("Earliest timestamp is incorrect")
	}

	if stats.TimeRange.Latest.Before(now) {
		t.Error("Latest timestamp is incorrect")
	}
}

func TestGetStatsRecentLogs(t *testing.T) {
	aggregator := NewLogAggregator()

	// Add 15 logs
	for i := 0; i < 15; i++ {
		aggregator.Ingest(LogEntry{Message: "Log " + string(rune('0'+i))})
	}

	stats := aggregator.GetStats()

	if len(stats.RecentLogs) != 10 {
		t.Errorf("Expected 10 recent logs, got %d", len(stats.RecentLogs))
	}

	// Recent logs should be newest first
	// The last ingested log should be first in recent logs
}

func TestAggregate(t *testing.T) {
	aggregator := NewLogAggregator()

	aggregator.Ingest(LogEntry{Level: LevelInfo})
	aggregator.Ingest(LogEntry{Level: LevelError})
	aggregator.Ingest(LogEntry{Level: LevelError})
	aggregator.Ingest(LogEntry{Level: LevelWarn})

	result := aggregator.Aggregate("level")

	if result.Field != "level" {
		t.Errorf("Expected field 'level', got '%s'", result.Field)
	}

	if len(result.Groups) != 3 {
		t.Errorf("Expected 3 groups, got %d", len(result.Groups))
	}

	// Should be sorted by count (descending)
	if result.Groups[0].Count < result.Groups[1].Count {
		t.Error("Groups should be sorted by count descending")
	}

	// Check ERROR has count 2
	var errorCount int
	for _, group := range result.Groups {
		if group.Value == "ERROR" {
			errorCount = group.Count
		}
	}

	if errorCount != 2 {
		t.Errorf("Expected ERROR count 2, got %d", errorCount)
	}
}

func TestAggregateBySource(t *testing.T) {
	aggregator := NewLogAggregator()

	aggregator.Ingest(LogEntry{Source: "app"})
	aggregator.Ingest(LogEntry{Source: "app"})
	aggregator.Ingest(LogEntry{Source: "db"})

	result := aggregator.Aggregate("source")

	if len(result.Groups) != 2 {
		t.Errorf("Expected 2 groups, got %d", len(result.Groups))
	}

	// app should have count 2
	var appCount int
	for _, group := range result.Groups {
		if group.Value == "app" {
			appCount = group.Count
		}
	}

	if appCount != 2 {
		t.Errorf("Expected app count 2, got %d", appCount)
	}
}

func TestAggregateByCustomField(t *testing.T) {
	aggregator := NewLogAggregator()

	aggregator.Ingest(LogEntry{
		Message: "Test",
		Fields:  map[string]interface{}{"environment": "prod"},
	})
	aggregator.Ingest(LogEntry{
		Message: "Test",
		Fields:  map[string]interface{}{"environment": "dev"},
	})
	aggregator.Ingest(LogEntry{
		Message: "Test",
		Fields:  map[string]interface{}{"environment": "prod"},
	})

	result := aggregator.Aggregate("environment")

	if len(result.Groups) != 2 {
		t.Errorf("Expected 2 groups, got %d", len(result.Groups))
	}
}

func TestParseJSONLog(t *testing.T) {
	jsonData := `{
		"level": "ERROR",
		"message": "Test error",
		"source": "app",
		"timestamp": "2024-01-01T12:00:00Z"
	}`

	entry, err := ParseJSONLog([]byte(jsonData))
	if err != nil {
		t.Fatalf("ParseJSONLog failed: %v", err)
	}

	if entry.Level != LevelError {
		t.Errorf("Expected level ERROR, got %s", entry.Level)
	}

	if entry.Message != "Test error" {
		t.Errorf("Expected message 'Test error', got '%s'", entry.Message)
	}

	if entry.Source != "app" {
		t.Errorf("Expected source 'app', got '%s'", entry.Source)
	}
}

func TestParseJSONLogInvalid(t *testing.T) {
	invalidJSON := `{invalid json`

	_, err := ParseJSONLog([]byte(invalidJSON))
	if err == nil {
		t.Error("Expected error for invalid JSON")
	}
}

func TestParseCommonLogFormat(t *testing.T) {
	line := `127.0.0.1 - - [10/Oct/2000:13:55:36 -0700] "GET /index.html HTTP/1.0" 200 1234`

	entry, err := ParseCommonLogFormat(line)
	if err != nil {
		t.Fatalf("ParseCommonLogFormat failed: %v", err)
	}

	if entry.Level != LevelInfo {
		t.Errorf("Expected level INFO, got %s", entry.Level)
	}

	if entry.Source != "clf" {
		t.Errorf("Expected source 'clf', got '%s'", entry.Source)
	}

	if entry.Fields == nil {
		t.Fatal("Fields should not be nil")
	}

	if entry.Fields["ip"] != "127.0.0.1" {
		t.Errorf("Expected IP 127.0.0.1, got %v", entry.Fields["ip"])
	}

	if entry.Fields["method"] != "GET" {
		t.Errorf("Expected method GET, got %v", entry.Fields["method"])
	}

	if entry.Fields["status"] != "200" {
		t.Errorf("Expected status 200, got %v", entry.Fields["status"])
	}
}

func TestParseCommonLogFormatInvalid(t *testing.T) {
	invalidLine := "not a valid log line"

	_, err := ParseCommonLogFormat(invalidLine)
	if err == nil {
		t.Error("Expected error for invalid CLF")
	}
}

func TestLogEntrySerialization(t *testing.T) {
	entry := LogEntry{
		ID:        "test-1",
		Timestamp: time.Now(),
		Level:     LevelError,
		Message:   "Test message",
		Source:    "app",
		Fields: map[string]interface{}{
			"key": "value",
		},
	}

	data, err := json.Marshal(entry)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var decoded LogEntry
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if decoded.ID != entry.ID {
		t.Error("ID mismatch")
	}

	if decoded.Level != entry.Level {
		t.Error("Level mismatch")
	}

	if decoded.Message != entry.Message {
		t.Error("Message mismatch")
	}
}

func TestConcurrentIngest(t *testing.T) {
	aggregator := NewLogAggregator()

	done := make(chan bool)
	count := 100

	for i := 0; i < count; i++ {
		go func(id int) {
			entry := LogEntry{
				Level:   LevelInfo,
				Message: "Concurrent log",
			}
			aggregator.Ingest(entry)
			done <- true
		}(i)
	}

	for i := 0; i < count; i++ {
		<-done
	}

	if len(aggregator.logs) != count {
		t.Errorf("Expected %d logs, got %d", count, len(aggregator.logs))
	}
}

func TestConcurrentSearchAndIngest(t *testing.T) {
	aggregator := NewLogAggregator()

	// Pre-populate with some logs
	for i := 0; i < 10; i++ {
		aggregator.Ingest(LogEntry{Level: LevelInfo, Message: "Log"})
	}

	done := make(chan bool)

	// Concurrent ingests
	go func() {
		for i := 0; i < 10; i++ {
			aggregator.Ingest(LogEntry{Level: LevelError, Message: "Error"})
		}
		done <- true
	}()

	// Concurrent searches
	go func() {
		for i := 0; i < 10; i++ {
			aggregator.Search(SearchQuery{})
		}
		done <- true
	}()

	<-done
	<-done

	// No panic = success
}

func TestSearchResultMetadata(t *testing.T) {
	aggregator := NewLogAggregator()

	for i := 0; i < 5; i++ {
		aggregator.Ingest(LogEntry{Message: "Test"})
	}

	result := aggregator.Search(SearchQuery{Limit: 3})

	if result.QueryTime == "" {
		t.Error("QueryTime should be set")
	}

	if result.Limit != 3 {
		t.Errorf("Expected limit 3, got %d", result.Limit)
	}

	if result.Total != 5 {
		t.Errorf("Expected total 5, got %d", result.Total)
	}
}

func TestLogLevels(t *testing.T) {
	levels := []LogLevel{LevelDebug, LevelInfo, LevelWarn, LevelError, LevelFatal}

	aggregator := NewLogAggregator()

	for _, level := range levels {
		aggregator.Ingest(LogEntry{Level: level, Message: "Test"})
	}

	stats := aggregator.GetStats()

	if len(stats.ByLevel) != 5 {
		t.Errorf("Expected 5 different levels, got %d", len(stats.ByLevel))
	}

	for _, level := range levels {
		if stats.ByLevel[level] != 1 {
			t.Errorf("Expected 1 log for level %s, got %d", level, stats.ByLevel[level])
		}
	}
}
