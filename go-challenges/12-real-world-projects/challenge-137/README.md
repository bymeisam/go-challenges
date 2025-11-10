# Challenge 137: Web Scraper

**Difficulty:** ⭐⭐⭐⭐ Hard
**Time Estimate:** 50 minutes

## Description

Build a concurrent web scraper that extracts data from HTML pages using goquery. This demonstrates HTTP clients, HTML parsing, concurrency for web scraping, and rate limiting.

## Features

- **Concurrent Scraping**: Parallel fetching with worker pools
- **HTML Parsing**: Extract data using CSS selectors
- **Rate Limiting**: Respect server limits
- **Retry Logic**: Handle temporary failures
- **URL Deduplication**: Avoid scraping same URL twice
- **Link Following**: Optionally follow links to depth N
- **Data Extraction**: Structured data extraction
- **Export Results**: Save to JSON/CSV

## Scraping Targets

- **Links**: Extract all links from pages
- **Metadata**: Title, description, keywords
- **Content**: Headings, paragraphs, lists
- **Images**: Image URLs and alt text
- **Custom Selectors**: User-defined CSS selectors

## Requirements

1. Use goquery for HTML parsing
2. Implement concurrent fetching with worker pool
3. Rate limiting to avoid overwhelming servers
4. Retry failed requests with exponential backoff
5. Deduplication of URLs
6. Configurable depth for link following
7. Structured output format

## Example Usage

```go
scraper := NewScraper(ScraperConfig{
    Workers:      5,
    RateLimit:    100 * time.Millisecond,
    MaxDepth:     2,
    MaxRetries:   3,
})

results, err := scraper.Scrape([]string{"https://example.com"})
for _, result := range results {
    fmt.Printf("%s: %s\n", result.URL, result.Title)
}
```

## Learning Objectives

- HTTP client usage
- HTML parsing with goquery
- Concurrent web requests
- Rate limiting patterns
- Error handling and retries
- URL normalization
- Data extraction and structuring
- Politeness in web scraping
