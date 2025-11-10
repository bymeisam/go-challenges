package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// LogLevel represents log severity
type LogLevel string

const (
	LevelDebug LogLevel = "DEBUG"
	LevelInfo  LogLevel = "INFO"
	LevelWarn  LogLevel = "WARN"
	LevelError LogLevel = "ERROR"
	LevelFatal LogLevel = "FATAL"
)

// LogEntry represents a parsed log entry
type LogEntry struct {
	ID        string                 `json:"id"`
	Timestamp time.Time              `json:"timestamp"`
	Level     LogLevel               `json:"level"`
	Message   string                 `json:"message"`
	Source    string                 `json:"source"`
	Fields    map[string]interface{} `json:"fields,omitempty"`
}

// SearchQuery represents a log search query
type SearchQuery struct {
	Level     string    `json:"level,omitempty"`
	Source    string    `json:"source,omitempty"`
	Message   string    `json:"message,omitempty"`
	StartTime time.Time `json:"start_time,omitempty"`
	EndTime   time.Time `json:"end_time,omitempty"`
	Limit     int       `json:"limit,omitempty"`
	Offset    int       `json:"offset,omitempty"`
}

// SearchResult represents search results
type SearchResult struct {
	Logs       []LogEntry `json:"logs"`
	Total      int        `json:"total"`
	Limit      int        `json:"limit"`
	Offset     int        `json:"offset"`
	QueryTime  string     `json:"query_time"`
}

// Stats represents log statistics
type Stats struct {
	TotalLogs   int                `json:"total_logs"`
	ByLevel     map[LogLevel]int   `json:"by_level"`
	BySource    map[string]int     `json:"by_source"`
	TimeRange   TimeRange          `json:"time_range"`
	RecentLogs  []LogEntry         `json:"recent_logs"`
}

// TimeRange represents a time range
type TimeRange struct {
	Earliest time.Time `json:"earliest"`
	Latest   time.Time `json:"latest"`
}

// AggregateResult represents aggregation results
type AggregateResult struct {
	Field  string         `json:"field"`
	Groups []GroupCount   `json:"groups"`
}

// GroupCount represents a group with count
type GroupCount struct {
	Value string `json:"value"`
	Count int    `json:"count"`
}

// LogAggregator manages log collection and analysis
type LogAggregator struct {
	logs      []LogEntry
	mu        sync.RWMutex
	idCounter int
}

// NewLogAggregator creates a new log aggregator
func NewLogAggregator() *LogAggregator {
	return &LogAggregator{
		logs:      make([]LogEntry, 0),
		idCounter: 1,
	}
}

// Ingest adds a log entry
func (la *LogAggregator) Ingest(entry LogEntry) string {
	la.mu.Lock()
	defer la.mu.Unlock()

	// Generate ID if not provided
	if entry.ID == "" {
		entry.ID = fmt.Sprintf("log-%d", la.idCounter)
		la.idCounter++
	}

	// Set timestamp if not provided
	if entry.Timestamp.IsZero() {
		entry.Timestamp = time.Now()
	}

	// Default level
	if entry.Level == "" {
		entry.Level = LevelInfo
	}

	la.logs = append(la.logs, entry)
	return entry.ID
}

// Search searches logs based on query
func (la *LogAggregator) Search(query SearchQuery) SearchResult {
	start := time.Now()

	la.mu.RLock()
	defer la.mu.RUnlock()

	// Filter logs
	filtered := make([]LogEntry, 0)
	for _, log := range la.logs {
		if la.matchesQuery(log, query) {
			filtered = append(filtered, log)
		}
	}

	// Sort by timestamp (newest first)
	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].Timestamp.After(filtered[j].Timestamp)
	})

	// Apply pagination
	total := len(filtered)
	offset := query.Offset
	limit := query.Limit

	if limit <= 0 {
		limit = 100
	}

	if offset >= total {
		offset = total
	}

	end := offset + limit
	if end > total {
		end = total
	}

	if offset < total {
		filtered = filtered[offset:end]
	} else {
		filtered = []LogEntry{}
	}

	return SearchResult{
		Logs:      filtered,
		Total:     total,
		Limit:     limit,
		Offset:    offset,
		QueryTime: time.Since(start).String(),
	}
}

// matchesQuery checks if a log matches the query
func (la *LogAggregator) matchesQuery(log LogEntry, query SearchQuery) bool {
	// Level filter
	if query.Level != "" && string(log.Level) != query.Level {
		return false
	}

	// Source filter
	if query.Source != "" && log.Source != query.Source {
		return false
	}

	// Message filter (contains)
	if query.Message != "" && !strings.Contains(strings.ToLower(log.Message), strings.ToLower(query.Message)) {
		return false
	}

	// Time range filter
	if !query.StartTime.IsZero() && log.Timestamp.Before(query.StartTime) {
		return false
	}

	if !query.EndTime.IsZero() && log.Timestamp.After(query.EndTime) {
		return false
	}

	return true
}

// GetStats returns log statistics
func (la *LogAggregator) GetStats() Stats {
	la.mu.RLock()
	defer la.mu.RUnlock()

	stats := Stats{
		TotalLogs: len(la.logs),
		ByLevel:   make(map[LogLevel]int),
		BySource:  make(map[string]int),
	}

	if len(la.logs) == 0 {
		return stats
	}

	// Calculate statistics
	earliest := la.logs[0].Timestamp
	latest := la.logs[0].Timestamp

	for _, log := range la.logs {
		// Count by level
		stats.ByLevel[log.Level]++

		// Count by source
		stats.BySource[log.Source]++

		// Track time range
		if log.Timestamp.Before(earliest) {
			earliest = log.Timestamp
		}
		if log.Timestamp.After(latest) {
			latest = log.Timestamp
		}
	}

	stats.TimeRange = TimeRange{
		Earliest: earliest,
		Latest:   latest,
	}

	// Recent logs (last 10)
	recentCount := 10
	if len(la.logs) < recentCount {
		recentCount = len(la.logs)
	}

	// Get last N logs
	stats.RecentLogs = make([]LogEntry, recentCount)
	copy(stats.RecentLogs, la.logs[len(la.logs)-recentCount:])

	// Reverse to show newest first
	for i := 0; i < len(stats.RecentLogs)/2; i++ {
		j := len(stats.RecentLogs) - 1 - i
		stats.RecentLogs[i], stats.RecentLogs[j] = stats.RecentLogs[j], stats.RecentLogs[i]
	}

	return stats
}

// Aggregate aggregates logs by a field
func (la *LogAggregator) Aggregate(field string) AggregateResult {
	la.mu.RLock()
	defer la.mu.RUnlock()

	counts := make(map[string]int)

	for _, log := range la.logs {
		var value string

		switch field {
		case "level":
			value = string(log.Level)
		case "source":
			value = log.Source
		default:
			// Try to get from fields
			if log.Fields != nil {
				if v, ok := log.Fields[field]; ok {
					value = fmt.Sprintf("%v", v)
				}
			}
		}

		if value != "" {
			counts[value]++
		}
	}

	// Convert to slice and sort
	groups := make([]GroupCount, 0, len(counts))
	for value, count := range counts {
		groups = append(groups, GroupCount{
			Value: value,
			Count: count,
		})
	}

	sort.Slice(groups, func(i, j int) bool {
		return groups[i].Count > groups[j].Count
	})

	return AggregateResult{
		Field:  field,
		Groups: groups,
	}
}

// ParseJSONLog parses a JSON log entry
func ParseJSONLog(data []byte) (*LogEntry, error) {
	var entry LogEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil, err
	}
	return &entry, nil
}

// ParseCommonLogFormat parses Common Log Format
func ParseCommonLogFormat(line string) (*LogEntry, error) {
	// Simple CLF parser: 127.0.0.1 - - [10/Oct/2000:13:55:36 -0700] "GET /apache_pb.gif HTTP/1.0" 200 2326
	pattern := `^(\S+) \S+ \S+ \[([^\]]+)\] "(\S+) (\S+) \S+" (\d+) (\d+)`
	re := regexp.MustCompile(pattern)
	matches := re.FindStringSubmatch(line)

	if len(matches) < 7 {
		return nil, fmt.Errorf("invalid CLF format")
	}

	// Parse timestamp (simplified)
	timestamp := time.Now() // In real implementation, parse matches[2]

	entry := &LogEntry{
		Timestamp: timestamp,
		Level:     LevelInfo,
		Message:   fmt.Sprintf("%s %s - Status: %s", matches[3], matches[4], matches[5]),
		Source:    "clf",
		Fields: map[string]interface{}{
			"ip":     matches[1],
			"method": matches[3],
			"path":   matches[4],
			"status": matches[5],
			"size":   matches[6],
		},
	}

	return entry, nil
}

// HTTP Handlers

func (la *LogAggregator) handleIngest(w http.ResponseWriter, r *http.Request) {
	var entry LogEntry
	if err := json.NewDecoder(r.Body).Decode(&entry); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	id := la.Ingest(entry)
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (la *LogAggregator) handleSearch(w http.ResponseWriter, r *http.Request) {
	query := SearchQuery{
		Level:   r.URL.Query().Get("level"),
		Source:  r.URL.Query().Get("source"),
		Message: r.URL.Query().Get("message"),
		Limit:   parseIntParam(r.URL.Query().Get("limit"), 100),
		Offset:  parseIntParam(r.URL.Query().Get("offset"), 0),
	}

	// Parse time range
	if start := r.URL.Query().Get("start"); start != "" {
		if t, err := time.Parse(time.RFC3339, start); err == nil {
			query.StartTime = t
		}
	}

	if end := r.URL.Query().Get("end"); end != "" {
		if t, err := time.Parse(time.RFC3339, end); err == nil {
			query.EndTime = t
		}
	}

	result := la.Search(query)
	respondJSON(w, http.StatusOK, result)
}

func (la *LogAggregator) handleStats(w http.ResponseWriter, r *http.Request) {
	stats := la.GetStats()
	respondJSON(w, http.StatusOK, stats)
}

func (la *LogAggregator) handleAggregate(w http.ResponseWriter, r *http.Request) {
	field := r.URL.Query().Get("field")
	if field == "" {
		field = "level"
	}

	result := la.Aggregate(field)
	respondJSON(w, http.StatusOK, result)
}

// Helper functions

func parseIntParam(s string, defaultValue int) int {
	if s == "" {
		return defaultValue
	}
	var value int
	fmt.Sscanf(s, "%d", &value)
	if value <= 0 {
		return defaultValue
	}
	return value
}

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]string{"error": message})
}

func main() {
	aggregator := NewLogAggregator()

	// Add some sample logs
	sampleLogs := []LogEntry{
		{
			Level:   LevelInfo,
			Message: "Application started",
			Source:  "app",
			Fields:  map[string]interface{}{"version": "1.0.0"},
		},
		{
			Level:   LevelWarn,
			Message: "High memory usage detected",
			Source:  "monitor",
			Fields:  map[string]interface{}{"memory_mb": 1024},
		},
		{
			Level:   LevelError,
			Message: "Database connection failed",
			Source:  "db",
			Fields:  map[string]interface{}{"retry_count": 3},
		},
	}

	for _, log := range sampleLogs {
		aggregator.Ingest(log)
	}

	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)

	// Routes
	r.Route("/api/logs", func(r chi.Router) {
		r.Post("/", aggregator.handleIngest)
		r.Get("/", aggregator.handleSearch)
		r.Get("/stats", aggregator.handleStats)
		r.Get("/aggregate", aggregator.handleAggregate)
	})

	// Home page
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		html := `
<!DOCTYPE html>
<html>
<head>
    <title>Log Aggregator</title>
    <style>
        body { font-family: Arial, sans-serif; max-width: 1000px; margin: 50px auto; padding: 20px; }
        pre { background: #f4f4f4; padding: 15px; border-radius: 5px; overflow-x: auto; }
        h2 { color: #333; border-bottom: 2px solid #333; padding-bottom: 10px; }
    </style>
</head>
<body>
    <h1>Log Aggregator API</h1>

    <h2>Ingest Log</h2>
    <pre>POST /api/logs
Content-Type: application/json

{
  "level": "ERROR",
  "message": "Something went wrong",
  "source": "app",
  "fields": {"key": "value"}
}</pre>

    <h2>Search Logs</h2>
    <pre>GET /api/logs?level=ERROR&limit=10&offset=0</pre>

    <h2>Get Statistics</h2>
    <pre>GET /api/logs/stats</pre>

    <h2>Aggregate Logs</h2>
    <pre>GET /api/logs/aggregate?field=level</pre>

    <h2>Example</h2>
    <pre>curl -X POST http://localhost:8080/api/logs \
  -H "Content-Type: application/json" \
  -d '{"level":"ERROR","message":"Test error","source":"test"}'

curl "http://localhost:8080/api/logs?level=ERROR"

curl http://localhost:8080/api/logs/stats</pre>
</body>
</html>
`
		fmt.Fprint(w, html)
	})

	log.Println("Log Aggregator starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
