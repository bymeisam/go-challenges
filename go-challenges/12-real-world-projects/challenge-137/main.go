package main

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// ScraperConfig holds configuration for the scraper
type ScraperConfig struct {
	Workers    int
	RateLimit  time.Duration
	MaxDepth   int
	MaxRetries int
	UserAgent  string
}

// PageData holds extracted data from a page
type PageData struct {
	URL         string            `json:"url"`
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Keywords    string            `json:"keywords"`
	Headings    []string          `json:"headings"`
	Links       []string          `json:"links"`
	Images      []ImageData       `json:"images"`
	StatusCode  int               `json:"status_code"`
	Error       string            `json:"error,omitempty"`
	ScrapedAt   time.Time         `json:"scraped_at"`
	CustomData  map[string]string `json:"custom_data,omitempty"`
}

// ImageData holds image information
type ImageData struct {
	URL string `json:"url"`
	Alt string `json:"alt"`
}

// Job represents a scraping job
type Job struct {
	URL   string
	Depth int
}

// Scraper performs web scraping
type Scraper struct {
	config      ScraperConfig
	client      *http.Client
	visited     map[string]bool
	visitedMu   sync.RWMutex
	rateLimiter *time.Ticker
}

// NewScraper creates a new scraper
func NewScraper(config ScraperConfig) *Scraper {
	if config.Workers <= 0 {
		config.Workers = 5
	}
	if config.RateLimit == 0 {
		config.RateLimit = 100 * time.Millisecond
	}
	if config.MaxRetries <= 0 {
		config.MaxRetries = 3
	}
	if config.UserAgent == "" {
		config.UserAgent = "GoWebScraper/1.0"
	}

	return &Scraper{
		config: config,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		visited:     make(map[string]bool),
		rateLimiter: time.NewTicker(config.RateLimit),
	}
}

// Scrape scrapes the given URLs
func (s *Scraper) Scrape(ctx context.Context, urls []string) ([]PageData, error) {
	jobs := make(chan Job, len(urls)*10)
	results := make(chan PageData, len(urls)*10)

	// Start workers
	var wg sync.WaitGroup
	for i := 0; i < s.config.Workers; i++ {
		wg.Add(1)
		go s.worker(ctx, &wg, jobs, results)
	}

	// Add initial jobs
	go func() {
		for _, u := range urls {
			if normalized := s.normalizeURL(u); normalized != "" {
				select {
				case <-ctx.Done():
					return
				case jobs <- Job{URL: normalized, Depth: 0}:
				}
			}
		}
	}()

	// Collect results
	allResults := make([]PageData, 0)
	done := make(chan struct{})

	go func() {
		wg.Wait()
		close(results)
		close(done)
	}()

	// Process results and potentially add new jobs
	go func() {
		for result := range results {
			allResults = append(allResults, result)

			// Add linked pages if within depth limit
			if result.Error == "" && result.StatusCode == 200 {
				for _, link := range result.Links {
					if normalized := s.normalizeURL(link); normalized != "" {
						depth := s.getDepth(result.URL) + 1
						if depth <= s.config.MaxDepth {
							select {
							case <-ctx.Done():
								return
							case jobs <- Job{URL: normalized, Depth: depth}:
							}
						}
					}
				}
			}
		}
	}()

	// Wait for initial URLs to be processed
	timeout := time.After(30 * time.Second)
	select {
	case <-done:
		close(jobs)
	case <-timeout:
		close(jobs)
	case <-ctx.Done():
		close(jobs)
		return allResults, ctx.Err()
	}

	<-done
	return allResults, nil
}

// worker processes scraping jobs
func (s *Scraper) worker(ctx context.Context, wg *sync.WaitGroup, jobs <-chan Job, results chan<- PageData) {
	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case job, ok := <-jobs:
			if !ok {
				return
			}

			// Check if already visited
			if !s.markVisited(job.URL) {
				continue
			}

			// Rate limiting
			select {
			case <-ctx.Done():
				return
			case <-s.rateLimiter.C:
			}

			// Scrape page with retries
			result := s.scrapePage(ctx, job.URL)
			select {
			case <-ctx.Done():
				return
			case results <- result:
			}
		}
	}
}

// scrapePage scrapes a single page
func (s *Scraper) scrapePage(ctx context.Context, pageURL string) PageData {
	result := PageData{
		URL:        pageURL,
		ScrapedAt:  time.Now(),
		CustomData: make(map[string]string),
	}

	var doc *goquery.Document
	var resp *http.Response
	var err error

	// Retry logic
	for attempt := 0; attempt < s.config.MaxRetries; attempt++ {
		if attempt > 0 {
			backoff := time.Duration(attempt) * time.Second
			time.Sleep(backoff)
		}

		req, err := http.NewRequestWithContext(ctx, "GET", pageURL, nil)
		if err != nil {
			result.Error = err.Error()
			return result
		}

		req.Header.Set("User-Agent", s.config.UserAgent)

		resp, err = s.client.Do(req)
		if err != nil {
			if attempt == s.config.MaxRetries-1 {
				result.Error = err.Error()
				return result
			}
			continue
		}

		result.StatusCode = resp.StatusCode
		if resp.StatusCode != 200 {
			resp.Body.Close()
			if attempt == s.config.MaxRetries-1 {
				result.Error = fmt.Sprintf("HTTP %d", resp.StatusCode)
				return result
			}
			continue
		}

		doc, err = goquery.NewDocumentFromReader(resp.Body)
		resp.Body.Close()
		if err != nil {
			if attempt == s.config.MaxRetries-1 {
				result.Error = err.Error()
				return result
			}
			continue
		}

		break
	}

	if doc == nil {
		return result
	}

	// Extract data
	s.extractMetadata(doc, &result)
	s.extractHeadings(doc, &result)
	s.extractLinks(doc, &result, pageURL)
	s.extractImages(doc, &result, pageURL)

	return result
}

// extractMetadata extracts page metadata
func (s *Scraper) extractMetadata(doc *goquery.Document, result *PageData) {
	result.Title = doc.Find("title").First().Text()

	doc.Find("meta").Each(func(i int, sel *goquery.Selection) {
		name, _ := sel.Attr("name")
		property, _ := sel.Attr("property")
		content, _ := sel.Attr("content")

		switch {
		case name == "description" || property == "og:description":
			if result.Description == "" {
				result.Description = content
			}
		case name == "keywords":
			result.Keywords = content
		}
	})
}

// extractHeadings extracts all headings
func (s *Scraper) extractHeadings(doc *goquery.Document, result *PageData) {
	headings := make([]string, 0)

	for _, tag := range []string{"h1", "h2", "h3", "h4", "h5", "h6"} {
		doc.Find(tag).Each(func(i int, sel *goquery.Selection) {
			text := strings.TrimSpace(sel.Text())
			if text != "" {
				headings = append(headings, text)
			}
		})
	}

	result.Headings = headings
}

// extractLinks extracts all links
func (s *Scraper) extractLinks(doc *goquery.Document, result *PageData, baseURL string) {
	links := make([]string, 0)
	seen := make(map[string]bool)

	doc.Find("a[href]").Each(func(i int, sel *goquery.Selection) {
		href, exists := sel.Attr("href")
		if !exists {
			return
		}

		absoluteURL := s.resolveURL(baseURL, href)
		if absoluteURL != "" && !seen[absoluteURL] {
			seen[absoluteURL] = true
			links = append(links, absoluteURL)
		}
	})

	result.Links = links
}

// extractImages extracts image information
func (s *Scraper) extractImages(doc *goquery.Document, result *PageData, baseURL string) {
	images := make([]ImageData, 0)

	doc.Find("img[src]").Each(func(i int, sel *goquery.Selection) {
		src, exists := sel.Attr("src")
		if !exists {
			return
		}

		absoluteURL := s.resolveURL(baseURL, src)
		if absoluteURL != "" {
			alt, _ := sel.Attr("alt")
			images = append(images, ImageData{
				URL: absoluteURL,
				Alt: alt,
			})
		}
	})

	result.Images = images
}

// resolveURL resolves a relative URL against a base URL
func (s *Scraper) resolveURL(baseURL, href string) string {
	base, err := url.Parse(baseURL)
	if err != nil {
		return ""
	}

	ref, err := url.Parse(href)
	if err != nil {
		return ""
	}

	resolved := base.ResolveReference(ref)
	return resolved.String()
}

// normalizeURL normalizes a URL
func (s *Scraper) normalizeURL(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}

	// Only HTTP/HTTPS
	if u.Scheme != "http" && u.Scheme != "https" {
		return ""
	}

	// Remove fragment
	u.Fragment = ""

	return u.String()
}

// markVisited marks a URL as visited
func (s *Scraper) markVisited(urlStr string) bool {
	s.visitedMu.Lock()
	defer s.visitedMu.Unlock()

	if s.visited[urlStr] {
		return false
	}

	s.visited[urlStr] = true
	return true
}

// getDepth gets the depth of a URL (for demonstration, returns 0)
func (s *Scraper) getDepth(urlStr string) int {
	// In a real implementation, this would track URL depths
	return 0
}

// Close closes the scraper
func (s *Scraper) Close() {
	s.rateLimiter.Stop()
}

func main() {
	// Example usage - scraping a local HTML page
	testHTML := `
<!DOCTYPE html>
<html>
<head>
    <title>Test Page</title>
    <meta name="description" content="A test page for web scraping">
    <meta name="keywords" content="test, scraping, golang">
</head>
<body>
    <h1>Welcome to Test Page</h1>
    <h2>Subsection 1</h2>
    <p>This is a test paragraph.</p>
    <h2>Subsection 2</h2>
    <p>Another paragraph with <a href="/page1">a link</a>.</p>
    <img src="/image1.jpg" alt="Test Image 1">
    <img src="/image2.jpg" alt="Test Image 2">
    <ul>
        <li><a href="/page2">Link 2</a></li>
        <li><a href="/page3">Link 3</a></li>
    </ul>
</body>
</html>
`

	// For demonstration, we'll parse the HTML directly
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(testHTML))
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Println("Web Scraper Demo")
	fmt.Println("================")

	// Extract title
	title := doc.Find("title").First().Text()
	fmt.Printf("\nTitle: %s\n", title)

	// Extract meta tags
	fmt.Println("\nMeta Tags:")
	doc.Find("meta").Each(func(i int, sel *goquery.Selection) {
		if name, exists := sel.Attr("name"); exists {
			content, _ := sel.Attr("content")
			fmt.Printf("  %s: %s\n", name, content)
		}
	})

	// Extract headings
	fmt.Println("\nHeadings:")
	doc.Find("h1, h2, h3").Each(func(i int, sel *goquery.Selection) {
		fmt.Printf("  %s: %s\n", goquery.NodeName(sel), sel.Text())
	})

	// Extract links
	fmt.Println("\nLinks:")
	doc.Find("a[href]").Each(func(i int, sel *goquery.Selection) {
		href, _ := sel.Attr("href")
		text := sel.Text()
		fmt.Printf("  %s -> %s\n", text, href)
	})

	// Extract images
	fmt.Println("\nImages:")
	doc.Find("img[src]").Each(func(i int, sel *goquery.Selection) {
		src, _ := sel.Attr("src")
		alt, _ := sel.Attr("alt")
		fmt.Printf("  %s (alt: %s)\n", src, alt)
	})

	// Real scraper example (commented out to avoid actual web requests)
	/*
	config := ScraperConfig{
		Workers:    3,
		RateLimit:  200 * time.Millisecond,
		MaxDepth:   1,
		MaxRetries: 2,
	}

	scraper := NewScraper(config)
	defer scraper.Close()

	urls := []string{"https://example.com"}
	results, err := scraper.Scrape(context.Background(), urls)

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	for _, result := range results {
		fmt.Printf("\nURL: %s\n", result.URL)
		fmt.Printf("Title: %s\n", result.Title)
		fmt.Printf("Links: %d\n", len(result.Links))
		fmt.Printf("Images: %d\n", len(result.Images))
	}
	*/
}
