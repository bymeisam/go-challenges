# Challenge 140: Log Aggregator

**Difficulty:** ⭐⭐⭐⭐ Hard
**Time Estimate:** 60 minutes

## Description

Build a log collection and analysis system that aggregates logs from multiple sources, parses different formats, and provides search and analytics capabilities. This demonstrates log processing, parsing, indexing, and querying.

## Features

- **Multiple Log Sources**: File tailing, HTTP endpoints, stdin
- **Format Parsing**: JSON, Common Log Format, custom formats
- **Real-time Processing**: Stream processing with workers
- **Indexing**: Fast log searching and filtering
- **Aggregation**: Count, group, and analyze logs
- **Time-based Queries**: Search by time range
- **Level Filtering**: Filter by log level (INFO, WARN, ERROR)
- **Pattern Matching**: Regex-based log filtering
- **Export**: Export results to JSON/CSV

## Log Formats Supported

- **JSON**: Structured JSON logs
- **Common Log Format**: Apache/Nginx style
- **Custom**: Configurable regex patterns

## Requirements

1. Concurrent log processing with worker pool
2. Parse multiple log formats
3. In-memory or persistent storage
4. Search and filter capabilities
5. Aggregation and statistics
6. Real-time log tailing
7. HTTP API for queries

## API Endpoints

```
POST   /api/logs           - Ingest logs
GET    /api/logs           - Search logs
GET    /api/logs/stats     - Get statistics
GET    /api/logs/aggregate - Aggregate by field
```

## Example Usage

```bash
# Ingest log
curl -X POST http://localhost:8080/api/logs \
  -H "Content-Type: application/json" \
  -d '{"level":"ERROR","message":"Database connection failed","timestamp":"2024-01-01T12:00:00Z"}'

# Search logs
curl "http://localhost:8080/api/logs?level=ERROR&limit=10"

# Get stats
curl http://localhost:8080/api/logs/stats

# Aggregate
curl "http://localhost:8080/api/logs/aggregate?field=level"
```

## Learning Objectives

- Log parsing and normalization
- Stream processing patterns
- Full-text search implementation
- Time-series data handling
- Aggregation and analytics
- Worker pool for log processing
- Real-time data ingestion
- Query language implementation
