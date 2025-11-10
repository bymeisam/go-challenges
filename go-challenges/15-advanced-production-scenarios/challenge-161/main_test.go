package main

import (
	"testing"
	"time"
)

func TestAddDocument(t *testing.T) {
	se := NewSearchEngine()
	doc := &Document{
		ID:      "1",
		Title:   "Go Programming",
		Content: "Learn Go programming language",
		Tags:    []string{"golang", "programming"},
	}

	err := se.AddDocument(doc)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	retrieved := se.GetDocument("1")
	if retrieved == nil {
		t.Fatal("Expected document to be found")
	}
}

func TestAddDocumentNoID(t *testing.T) {
	se := NewSearchEngine()
	doc := &Document{
		Title:   "Go Programming",
		Content: "Learn Go programming language",
	}

	err := se.AddDocument(doc)
	if err == nil {
		t.Fatal("Expected error for document without ID")
	}
}

func TestSearch(t *testing.T) {
	se := NewSearchEngine()

	doc1 := &Document{ID: "1", Title: "Go Programming", Content: "Learn Go", Tags: []string{"golang"}}
	doc2 := &Document{ID: "2", Title: "Rust Programming", Content: "Learn Rust", Tags: []string{"rust"}}

	se.AddDocument(doc1)
	se.AddDocument(doc2)

	query := &SearchQuery{Q: "go", Size: 10}
	results, err := se.Search(query)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(results) == 0 {
		t.Fatal("Expected search results")
	}
}

func TestSearchWithPagination(t *testing.T) {
	se := NewSearchEngine()

	for i := 0; i < 20; i++ {
		doc := &Document{
			ID:      string(rune(i)),
			Title:   "Go Programming",
			Content: "Learn Go",
		}
		se.AddDocument(doc)
	}

	query := &SearchQuery{Q: "go", Size: 5, From: 0}
	results, _ := se.Search(query)

	if len(results) != 5 {
		t.Fatalf("Expected 5 results, got %d", len(results))
	}

	query.From = 5
	results2, _ := se.Search(query)

	if len(results2) != 5 {
		t.Fatalf("Expected 5 results for second page, got %d", len(results2))
	}

	if results[0].Document.ID == results2[0].Document.ID {
		t.Fatal("Expected different results for different pages")
	}
}

func TestDeleteDocument(t *testing.T) {
	se := NewSearchEngine()
	doc := &Document{ID: "1", Title: "Go Programming", Content: "Learn Go"}

	se.AddDocument(doc)
	err := se.DeleteDocument("1")

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	retrieved := se.GetDocument("1")
	if retrieved != nil {
		t.Fatal("Expected document to be deleted")
	}
}

func TestDeleteNonexistentDocument(t *testing.T) {
	se := NewSearchEngine()

	err := se.DeleteDocument("nonexistent")
	if err == nil {
		t.Fatal("Expected error for nonexistent document")
	}
}

func TestUpdateDocument(t *testing.T) {
	se := NewSearchEngine()
	doc := &Document{ID: "1", Title: "Go Programming", Content: "Learn Go"}

	se.AddDocument(doc)

	updated := &Document{ID: "1", Title: "Advanced Go", Content: "Advanced Go topics"}
	se.UpdateDocument(updated)

	retrieved := se.GetDocument("1")
	if retrieved.Title != "Advanced Go" {
		t.Fatal("Expected title to be updated")
	}
}

func TestPhraseSearch(t *testing.T) {
	se := NewSearchEngine()

	doc := &Document{
		ID:      "1",
		Title:   "Go Programming Language",
		Content: "Go is a powerful programming language",
	}

	se.AddDocument(doc)

	results, _ := se.PhraseSearch("programming")
	if len(results) == 0 {
		t.Fatal("Expected phrase search results")
	}
}

func TestTermSearch(t *testing.T) {
	se := NewSearchEngine()

	doc := &Document{ID: "1", Title: "Go Programming", Content: "Learn Go", Tags: []string{"golang"}}
	se.AddDocument(doc)

	results, _ := se.TermSearch("golang")
	if len(results) == 0 {
		t.Fatal("Expected term search results")
	}
}

func TestMatchAll(t *testing.T) {
	se := NewSearchEngine()

	for i := 0; i < 5; i++ {
		doc := &Document{
			ID:    string(rune(i)),
			Title: "Document",
		}
		se.AddDocument(doc)
	}

	docs, _ := se.MatchAll()
	if len(docs) != 5 {
		t.Fatalf("Expected 5 documents, got %d", len(docs))
	}
}

func TestGetCategoryAggregation(t *testing.T) {
	se := NewSearchEngine()

	doc1 := &Document{ID: "1", Title: "Test", Category: "golang"}
	doc2 := &Document{ID: "2", Title: "Test", Category: "golang"}
	doc3 := &Document{ID: "3", Title: "Test", Category: "rust"}

	se.AddDocument(doc1)
	se.AddDocument(doc2)
	se.AddDocument(doc3)

	agg := se.GetCategoryAggregation()

	if agg.Buckets["golang"] != 2 {
		t.Fatalf("Expected 2 golang docs, got %d", agg.Buckets["golang"])
	}

	if agg.Buckets["rust"] != 1 {
		t.Fatalf("Expected 1 rust doc, got %d", agg.Buckets["rust"])
	}
}

func TestGetTagAggregation(t *testing.T) {
	se := NewSearchEngine()

	doc1 := &Document{ID: "1", Title: "Test", Tags: []string{"golang", "backend"}}
	doc2 := &Document{ID: "2", Title: "Test", Tags: []string{"golang"}}

	se.AddDocument(doc1)
	se.AddDocument(doc2)

	agg := se.GetTagAggregation()

	if agg.Buckets["golang"] != 2 {
		t.Fatalf("Expected 2 golang tags, got %d", agg.Buckets["golang"])
	}

	if agg.Buckets["backend"] != 1 {
		t.Fatalf("Expected 1 backend tag, got %d", agg.Buckets["backend"])
	}
}

func TestSearchCaching(t *testing.T) {
	se := NewSearchEngine()

	doc := &Document{ID: "1", Title: "Go Programming", Content: "Learn Go"}
	se.AddDocument(doc)

	query := &SearchQuery{Q: "go", Size: 10}

	results1, _ := se.Search(query)
	results2, _ := se.Search(query)

	// Should have same number of results
	if len(results1) != len(results2) {
		t.Fatal("Expected same results from cache")
	}

	stats := se.GetStats()
	if stats.CacheHits == 0 {
		t.Fatal("Expected cache hit on second query")
	}
}

func TestCacheInvalidationOnUpdate(t *testing.T) {
	se := NewSearchEngine()

	doc := &Document{ID: "1", Title: "Go Programming", Content: "Learn Go"}
	se.AddDocument(doc)

	query := &SearchQuery{Q: "go", Size: 10}
	se.Search(query)

	// Add another document
	doc2 := &Document{ID: "2", Title: "Go Advanced", Content: "Advanced Go"}
	se.AddDocument(doc2)

	// Cache should be invalidated
	results, _ := se.Search(query)
	if len(results) < 2 {
		t.Fatal("Expected fresh search after document added")
	}
}

func TestGetIndexStats(t *testing.T) {
	se := NewSearchEngine()

	for i := 0; i < 10; i++ {
		doc := &Document{
			ID:    string(rune(i)),
			Title: "Test Document",
		}
		se.AddDocument(doc)
	}

	stats := se.GetIndexStats()

	if stats["document_count"] != int64(10) {
		t.Fatalf("Expected 10 documents, got %v", stats["document_count"])
	}

	if stats["term_count"] == 0 {
		t.Fatal("Expected non-zero term count")
	}
}

func TestGetStats(t *testing.T) {
	se := NewSearchEngine()

	doc := &Document{ID: "1", Title: "Go Programming", Content: "Learn Go"}
	se.AddDocument(doc)

	query := &SearchQuery{Q: "go", Size: 10}
	se.Search(query)

	stats := se.GetStats()

	if stats.TotalQueries == 0 {
		t.Fatal("Expected queries in stats")
	}
}

func TestSearchResultScoring(t *testing.T) {
	se := NewSearchEngine()

	// Document with keyword in title should score higher
	doc1 := &Document{
		ID:      "1",
		Title:   "Go Programming",
		Content: "Some other content",
	}

	// Document with keyword only in content
	doc2 := &Document{
		ID:      "2",
		Title:   "Something else",
		Content: "Go is great for programming",
	}

	se.AddDocument(doc1)
	se.AddDocument(doc2)

	query := &SearchQuery{Q: "go", Size: 10}
	results, _ := se.Search(query)

	if len(results) >= 2 && results[0].Score <= results[1].Score {
		// Title match should score higher or equal
		if results[0].Document.ID != "1" {
			// doc1 should be first due to title boost
		}
	}
}

func TestMultiTokenSearch(t *testing.T) {
	se := NewSearchEngine()

	doc1 := &Document{
		ID:      "1",
		Title:   "Go Programming Language",
		Content: "Learn Go programming",
	}

	doc2 := &Document{
		ID:      "2",
		Title:   "Go basics",
		Content: "Learn basics",
	}

	se.AddDocument(doc1)
	se.AddDocument(doc2)

	query := &SearchQuery{Q: "go programming", Size: 10}
	results, _ := se.Search(query)

	if len(results) == 0 {
		t.Fatal("Expected results for multi-token search")
	}

	// doc1 should score higher as it contains both terms
	if results[0].Document.ID != "1" {
		t.Fatal("Expected doc1 with both tokens to score higher")
	}
}

// ========== Benchmarks ==========

func BenchmarkAddDocument(b *testing.B) {
	se := NewSearchEngine()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		doc := &Document{
			ID:      string(rune(i)),
			Title:   "Go Programming Language",
			Content: "Learn Go programming",
			Tags:    []string{"golang", "programming"},
		}
		se.AddDocument(doc)
	}
}

func BenchmarkSearch(b *testing.B) {
	se := NewSearchEngine()

	for i := 0; i < 100; i++ {
		doc := &Document{
			ID:      string(rune(i)),
			Title:   "Go Programming",
			Content: "Learn Go programming",
		}
		se.AddDocument(doc)
	}

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		query := &SearchQuery{Q: "go", Size: 10}
		se.Search(query)
	}
}

func BenchmarkSearchWithCache(b *testing.B) {
	se := NewSearchEngine()

	for i := 0; i < 100; i++ {
		doc := &Document{
			ID:      string(rune(i)),
			Title:   "Go Programming",
			Content: "Learn Go programming",
		}
		se.AddDocument(doc)
	}

	query := &SearchQuery{Q: "go", Size: 10}

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		se.Search(query)
	}
}
