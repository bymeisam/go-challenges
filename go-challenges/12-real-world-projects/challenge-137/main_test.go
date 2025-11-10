package main

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/PuerkitoBio/goquery"
)

const testHTML = `
<!DOCTYPE html>
<html>
<head>
    <title>Test Page Title</title>
    <meta name="description" content="Test description">
    <meta name="keywords" content="test, keywords">
</head>
<body>
    <h1>Main Heading</h1>
    <h2>Subheading 1</h2>
    <p>Some content</p>
    <h2>Subheading 2</h2>
    <a href="/page1">Link 1</a>
    <a href="/page2">Link 2</a>
    <a href="http://external.com">External</a>
    <img src="/img1.jpg" alt="Image 1">
    <img src="/img2.jpg" alt="Image 2">
</body>
</html>
`

func TestNewScraper(t *testing.T) {
	tests := []struct {
		name   string
		config ScraperConfig
	}{
		{
			name: "Default config",
			config: ScraperConfig{
				Workers:    0,
				RateLimit:  0,
				MaxRetries: 0,
			},
		},
		{
			name: "Custom config",
			config: ScraperConfig{
				Workers:    10,
				RateLimit:  50 * time.Millisecond,
				MaxRetries: 5,
				UserAgent:  "TestBot/1.0",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scraper := NewScraper(tt.config)
			defer scraper.Close()

			if scraper == nil {
				t.Fatal("Scraper should not be nil")
			}

			if scraper.config.Workers <= 0 {
				t.Error("Workers should be positive")
			}

			if scraper.config.RateLimit == 0 {
				t.Error("RateLimit should be set")
			}

			if scraper.config.MaxRetries <= 0 {
				t.Error("MaxRetries should be positive")
			}

			if scraper.config.UserAgent == "" {
				t.Error("UserAgent should be set")
			}
		})
	}
}

func TestExtractMetadata(t *testing.T) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(testHTML))
	if err != nil {
		t.Fatalf("Failed to parse HTML: %v", err)
	}

	scraper := NewScraper(ScraperConfig{})
	defer scraper.Close()

	result := PageData{}
	scraper.extractMetadata(doc, &result)

	if result.Title != "Test Page Title" {
		t.Errorf("Expected title 'Test Page Title', got '%s'", result.Title)
	}

	if result.Description != "Test description" {
		t.Errorf("Expected description 'Test description', got '%s'", result.Description)
	}

	if result.Keywords != "test, keywords" {
		t.Errorf("Expected keywords 'test, keywords', got '%s'", result.Keywords)
	}
}

func TestExtractHeadings(t *testing.T) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(testHTML))
	if err != nil {
		t.Fatalf("Failed to parse HTML: %v", err)
	}

	scraper := NewScraper(ScraperConfig{})
	defer scraper.Close()

	result := PageData{}
	scraper.extractHeadings(doc, &result)

	if len(result.Headings) != 3 {
		t.Errorf("Expected 3 headings, got %d", len(result.Headings))
	}

	expectedHeadings := []string{"Main Heading", "Subheading 1", "Subheading 2"}
	for i, expected := range expectedHeadings {
		if i >= len(result.Headings) || result.Headings[i] != expected {
			t.Errorf("Expected heading %d to be '%s'", i, expected)
		}
	}
}

func TestExtractLinks(t *testing.T) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(testHTML))
	if err != nil {
		t.Fatalf("Failed to parse HTML: %v", err)
	}

	scraper := NewScraper(ScraperConfig{})
	defer scraper.Close()

	result := PageData{}
	scraper.extractLinks(doc, &result, "http://example.com")

	if len(result.Links) != 3 {
		t.Errorf("Expected 3 links, got %d", len(result.Links))
	}

	// Check that links were resolved to absolute URLs
	for _, link := range result.Links {
		if !strings.HasPrefix(link, "http://") && !strings.HasPrefix(link, "https://") {
			t.Errorf("Link should be absolute URL, got: %s", link)
		}
	}
}

func TestExtractImages(t *testing.T) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(testHTML))
	if err != nil {
		t.Fatalf("Failed to parse HTML: %v", err)
	}

	scraper := NewScraper(ScraperConfig{})
	defer scraper.Close()

	result := PageData{}
	scraper.extractImages(doc, &result, "http://example.com")

	if len(result.Images) != 2 {
		t.Errorf("Expected 2 images, got %d", len(result.Images))
	}

	for i, img := range result.Images {
		if img.URL == "" {
			t.Errorf("Image %d should have URL", i)
		}
		if img.Alt == "" {
			t.Errorf("Image %d should have alt text", i)
		}
	}
}

func TestNormalizeURL(t *testing.T) {
	scraper := NewScraper(ScraperConfig{})
	defer scraper.Close()

	tests := []struct {
		input    string
		expected string
	}{
		{"http://example.com", "http://example.com"},
		{"https://example.com", "https://example.com"},
		{"http://example.com#fragment", "http://example.com"},
		{"ftp://example.com", ""}, // Should reject non-HTTP(S)
		{"invalid-url", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := scraper.normalizeURL(tt.input)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestResolveURL(t *testing.T) {
	scraper := NewScraper(ScraperConfig{})
	defer scraper.Close()

	tests := []struct {
		base     string
		href     string
		expected string
	}{
		{"http://example.com", "/page", "http://example.com/page"},
		{"http://example.com", "page", "http://example.com/page"},
		{"http://example.com/dir/", "page", "http://example.com/dir/page"},
		{"http://example.com", "http://other.com", "http://other.com"},
	}

	for _, tt := range tests {
		t.Run(tt.href, func(t *testing.T) {
			result := scraper.resolveURL(tt.base, tt.href)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestMarkVisited(t *testing.T) {
	scraper := NewScraper(ScraperConfig{})
	defer scraper.Close()

	url := "http://example.com"

	// First visit should succeed
	if !scraper.markVisited(url) {
		t.Error("First visit should return true")
	}

	// Second visit should fail
	if scraper.markVisited(url) {
		t.Error("Second visit should return false")
	}

	// Different URL should succeed
	if !scraper.markVisited("http://example.com/other") {
		t.Error("Different URL should return true")
	}
}

func TestScrapePage(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, testHTML)
	}))
	defer server.Close()

	scraper := NewScraper(ScraperConfig{
		MaxRetries: 2,
	})
	defer scraper.Close()

	result := scraper.scrapePage(context.Background(), server.URL)

	if result.Error != "" {
		t.Errorf("Scraping failed: %s", result.Error)
	}

	if result.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", result.StatusCode)
	}

	if result.Title == "" {
		t.Error("Title should be extracted")
	}

	if len(result.Headings) == 0 {
		t.Error("Headings should be extracted")
	}

	if len(result.Links) == 0 {
		t.Error("Links should be extracted")
	}
}

func TestScrapePageError(t *testing.T) {
	scraper := NewScraper(ScraperConfig{
		MaxRetries: 1,
	})
	defer scraper.Close()

	// Try to scrape non-existent server
	result := scraper.scrapePage(context.Background(), "http://localhost:9999")

	if result.Error == "" {
		t.Error("Expected error for non-existent server")
	}
}

func TestScrapePageNotFound(t *testing.T) {
	// Create test server that returns 404
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	scraper := NewScraper(ScraperConfig{
		MaxRetries: 1,
	})
	defer scraper.Close()

	result := scraper.scrapePage(context.Background(), server.URL)

	if result.StatusCode != 404 {
		t.Errorf("Expected status 404, got %d", result.StatusCode)
	}

	if result.Error == "" {
		t.Error("Expected error for 404 response")
	}
}

func TestScrapeWithContext(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		fmt.Fprint(w, testHTML)
	}))
	defer server.Close()

	scraper := NewScraper(ScraperConfig{
		Workers: 1,
	})
	defer scraper.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err := scraper.Scrape(ctx, []string{server.URL})
	if err != context.DeadlineExceeded && err != context.Canceled {
		t.Errorf("Expected context error, got: %v", err)
	}
}

func TestScrapeMultipleURLs(t *testing.T) {
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		fmt.Fprint(w, testHTML)
	}))
	defer server.Close()

	scraper := NewScraper(ScraperConfig{
		Workers:    2,
		RateLimit:  10 * time.Millisecond,
		MaxRetries: 1,
	})
	defer scraper.Close()

	urls := []string{
		server.URL + "/page1",
		server.URL + "/page2",
		server.URL + "/page3",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	results, err := scraper.Scrape(ctx, urls)
	if err != nil && err != context.DeadlineExceeded {
		t.Fatalf("Scrape failed: %v", err)
	}

	if len(results) == 0 {
		t.Error("Expected at least one result")
	}

	// Check that requests were made
	if requestCount == 0 {
		t.Error("Expected requests to be made")
	}
}

func TestRateLimiting(t *testing.T) {
	requestTimes := make([]time.Time, 0)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestTimes = append(requestTimes, time.Now())
		fmt.Fprint(w, testHTML)
	}))
	defer server.Close()

	rateLimit := 100 * time.Millisecond
	scraper := NewScraper(ScraperConfig{
		Workers:    1, // Single worker to ensure sequential processing
		RateLimit:  rateLimit,
		MaxRetries: 1,
	})
	defer scraper.Close()

	urls := []string{
		server.URL + "/page1",
		server.URL + "/page2",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	scraper.Scrape(ctx, urls)

	if len(requestTimes) >= 2 {
		timeBetween := requestTimes[1].Sub(requestTimes[0])
		if timeBetween < rateLimit/2 {
			t.Errorf("Rate limiting not working: requests only %v apart", timeBetween)
		}
	}
}

func TestCustomUserAgent(t *testing.T) {
	receivedUA := ""
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedUA = r.UserAgent()
		fmt.Fprint(w, testHTML)
	}))
	defer server.Close()

	customUA := "CustomBot/1.0"
	scraper := NewScraper(ScraperConfig{
		UserAgent:  customUA,
		MaxRetries: 1,
	})
	defer scraper.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	scraper.Scrape(ctx, []string{server.URL})

	if receivedUA != customUA {
		t.Errorf("Expected User-Agent '%s', got '%s'", customUA, receivedUA)
	}
}

func TestEmptyURLList(t *testing.T) {
	scraper := NewScraper(ScraperConfig{})
	defer scraper.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	results, err := scraper.Scrape(ctx, []string{})
	if err != nil {
		// Empty list might return error or empty results
		t.Logf("Got error for empty list: %v", err)
	}

	if len(results) != 0 {
		t.Errorf("Expected 0 results for empty URL list, got %d", len(results))
	}
}

func TestPageDataTimestamp(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, testHTML)
	}))
	defer server.Close()

	scraper := NewScraper(ScraperConfig{})
	defer scraper.Close()

	before := time.Now()
	result := scraper.scrapePage(context.Background(), server.URL)
	after := time.Now()

	if result.ScrapedAt.Before(before) || result.ScrapedAt.After(after) {
		t.Error("ScrapedAt timestamp is out of expected range")
	}
}

func TestHTMLWithoutMetaTags(t *testing.T) {
	minimalHTML := `<html><body><h1>Title</h1></body></html>`
	doc, _ := goquery.NewDocumentFromReader(strings.NewReader(minimalHTML))

	scraper := NewScraper(ScraperConfig{})
	defer scraper.Close()

	result := PageData{}
	scraper.extractMetadata(doc, &result)

	// Should not crash, just have empty metadata
	if result.Title != "" {
		t.Logf("Title: %s", result.Title)
	}
}

func TestHTMLWithNoLinks(t *testing.T) {
	noLinksHTML := `<html><body><h1>No Links Here</h1><p>Just text</p></body></html>`
	doc, _ := goquery.NewDocumentFromReader(strings.NewReader(noLinksHTML))

	scraper := NewScraper(ScraperConfig{})
	defer scraper.Close()

	result := PageData{}
	scraper.extractLinks(doc, &result, "http://example.com")

	if len(result.Links) != 0 {
		t.Errorf("Expected 0 links, got %d", len(result.Links))
	}
}

func TestConcurrentScraping(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping concurrent test in short mode")
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(50 * time.Millisecond)
		fmt.Fprint(w, testHTML)
	}))
	defer server.Close()

	scraper := NewScraper(ScraperConfig{
		Workers:    5,
		RateLimit:  10 * time.Millisecond,
		MaxRetries: 1,
	})
	defer scraper.Close()

	// Create 10 URLs
	urls := make([]string, 10)
	for i := 0; i < 10; i++ {
		urls[i] = fmt.Sprintf("%s/page%d", server.URL, i)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	start := time.Now()
	results, err := scraper.Scrape(ctx, urls)
	duration := time.Since(start)

	if err != nil && err != context.DeadlineExceeded {
		t.Logf("Scrape error: %v", err)
	}

	t.Logf("Scraped %d pages in %v with 5 workers", len(results), duration)

	// With 5 workers, should be faster than sequential
	if duration > 2*time.Second {
		t.Logf("Warning: Scraping took longer than expected: %v", duration)
	}
}
