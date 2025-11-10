package main

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"sync"
	"time"
)

// ========== Document Models ==========

type Document struct {
	ID        string                 `json:"id"`
	Title     string                 `json:"title"`
	Content   string                 `json:"content"`
	Category  string                 `json:"category"`
	Tags      []string               `json:"tags"`
	Score     float64                `json:"score"`
	Timestamp time.Time              `json:"timestamp"`
	Fields    map[string]interface{} `json:"fields"`
}

type SearchResult struct {
	Document      *Document              `json:"document"`
	Score         float64                `json:"score"`
	HighlightText string                 `json:"highlight_text"`
	Matches       map[string][]string    `json:"matches"`
}

type SearchQuery struct {
	Q          string                 `json:"q"`
	Fields     []string               `json:"fields"`
	Filter     map[string]interface{} `json:"filter"`
	Facets     []string               `json:"facets"`
	Size       int                    `json:"size"`
	From       int                    `json:"from"`
	Boost      map[string]float64     `json:"boost"`
}

// ========== Index Structures ==========

type InvertedIndex struct {
	terms    map[string]map[string]bool // term -> docID -> true
	termsMu  sync.RWMutex
	docMeta  map[string]*Document
	metaMu   sync.RWMutex
	docCount int64
}

type SearchEngine struct {
	index        *InvertedIndex
	documents    map[string]*Document
	docsMu       sync.RWMutex
	queryCache   map[string][]*SearchResult
	cacheMu      sync.RWMutex
	searchStats  *SearchStats
	statsMu      sync.RWMutex
	facetIndex   map[string]map[string]int // field -> value -> count
	facetMu      sync.RWMutex
}

type SearchStats struct {
	TotalQueries    int64
	TotalResults    int64
	AvgQueryTime    time.Duration
	CacheHits       int64
	CacheMisses     int64
}

type Aggregation struct {
	Field   string               `json:"field"`
	Type    string               `json:"type"` // terms, histogram, range
	Buckets map[string]int       `json:"buckets"`
	Total   int                  `json:"total"`
}

// ========== Index Operations ==========

func NewSearchEngine() *SearchEngine {
	return &SearchEngine{
		index:       &InvertedIndex{terms: make(map[string]map[string]bool), docMeta: make(map[string]*Document)},
		documents:   make(map[string]*Document),
		queryCache:  make(map[string][]*SearchResult),
		searchStats: &SearchStats{},
		facetIndex:  make(map[string]map[string]int),
	}
}

// AddDocument adds a document to the index
func (se *SearchEngine) AddDocument(doc *Document) error {
	if doc.ID == "" {
		return fmt.Errorf("document must have an ID")
	}

	se.docsMu.Lock()
	se.documents[doc.ID] = doc
	se.docsMu.Unlock()

	// Index text
	se.indexDocument(doc)

	// Update facets
	se.updateFacets(doc)

	// Invalidate cache
	se.invalidateCache()

	return nil
}

func (se *SearchEngine) indexDocument(doc *Document) {
	se.index.termsMu.Lock()
	defer se.index.termsMu.Unlock()

	se.index.metaMu.Lock()
	se.index.docMeta[doc.ID] = doc
	se.index.docCount++
	se.index.metaMu.Unlock()

	// Index title
	for _, token := range tokenize(doc.Title) {
		if _, exists := se.index.terms[token]; !exists {
			se.index.terms[token] = make(map[string]bool)
		}
		se.index.terms[token][doc.ID] = true
	}

	// Index content
	for _, token := range tokenize(doc.Content) {
		if _, exists := se.index.terms[token]; !exists {
			se.index.terms[token] = make(map[string]bool)
		}
		se.index.terms[token][doc.ID] = true
	}

	// Index tags
	for _, tag := range doc.Tags {
		if _, exists := se.index.terms[tag]; !exists {
			se.index.terms[tag] = make(map[string]bool)
		}
		se.index.terms[tag][doc.ID] = true
	}
}

func (se *SearchEngine) updateFacets(doc *Document) {
	se.facetMu.Lock()
	defer se.facetMu.Unlock()

	// Index category
	if _, exists := se.facetIndex["category"]; !exists {
		se.facetIndex["category"] = make(map[string]int)
	}
	se.facetIndex["category"][doc.Category]++

	// Index tags
	if _, exists := se.facetIndex["tags"]; !exists {
		se.facetIndex["tags"] = make(map[string]int)
	}
	for _, tag := range doc.Tags {
		se.facetIndex["tags"][tag]++
	}
}

// ========== Search Operations ==========

// Search performs a search query
func (se *SearchEngine) Search(query *SearchQuery) ([]*SearchResult, error) {
	queryKey := query.Q

	// Check cache
	se.cacheMu.RLock()
	if cached, exists := se.queryCache[queryKey]; exists {
		se.cacheMu.RUnlock()
		se.recordStatsCacheHit()
		return cached, nil
	}
	se.cacheMu.RUnlock()

	startTime := time.Now()

	// Parse and execute query
	docIDs := se.executeQuery(query)

	// Score results
	results := se.scoreResults(docIDs, query)

	// Apply pagination
	if query.From > 0 {
		if query.From >= len(results) {
			results = []*SearchResult{}
		} else {
			results = results[query.From:]
		}
	}

	if query.Size > 0 && len(results) > query.Size {
		results = results[:query.Size]
	}

	// Cache results
	se.cacheMu.Lock()
	se.queryCache[queryKey] = results
	se.cacheMu.Unlock()

	se.recordStats(len(results), time.Since(startTime))

	return results, nil
}

func (se *SearchEngine) executeQuery(query *SearchQuery) map[string]bool {
	tokens := tokenize(query.Q)
	matchingDocs := make(map[string]bool)

	se.index.termsMu.RLock()
	defer se.index.termsMu.RUnlock()

	// Find documents matching all tokens (AND search)
	for i, token := range tokens {
		if docs, exists := se.index.terms[token]; exists {
			if i == 0 {
				// Initialize with first token
				for docID := range docs {
					matchingDocs[docID] = true
				}
			} else {
				// Intersect with previous matches
				newMatches := make(map[string]bool)
				for docID := range matchingDocs {
					if docs[docID] {
						newMatches[docID] = true
					}
				}
				matchingDocs = newMatches
			}
		} else if i == 0 {
			// No matches for first token
			return make(map[string]bool)
		}
	}

	return matchingDocs
}

func (se *SearchEngine) scoreResults(docIDs map[string]bool, query *SearchQuery) []*SearchResult {
	var results []*SearchResult

	se.docsMu.RLock()
	defer se.docsMu.RUnlock()

	tokens := tokenize(query.Q)

	for docID := range docIDs {
		if doc, exists := se.documents[docID]; exists {
			score := calculateScore(doc, tokens, query)
			results = append(results, &SearchResult{
				Document: doc,
				Score:    score,
			})
		}
	}

	// Sort by score
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	return results
}

// ========== Aggregations ==========

// GetFacets returns faceted search results
func (se *SearchEngine) GetFacets(field string, query *SearchQuery) *Aggregation {
	se.facetMu.RLock()
	defer se.facetMu.RUnlock()

	agg := &Aggregation{
		Field:   field,
		Type:    "terms",
		Buckets: make(map[string]int),
	}

	if facets, exists := se.facetIndex[field]; exists {
		for value, count := range facets {
			agg.Buckets[value] = count
			agg.Total += count
		}
	}

	return agg
}

// GetCategoryAggregation returns category aggregation
func (se *SearchEngine) GetCategoryAggregation() *Aggregation {
	return se.GetFacets("category", nil)
}

// GetTagAggregation returns tag aggregation
func (se *SearchEngine) GetTagAggregation() *Aggregation {
	return se.GetFacets("tags", nil)
}

// ========== Document Management ==========

// GetDocument retrieves a document by ID
func (se *SearchEngine) GetDocument(docID string) *Document {
	se.docsMu.RLock()
	defer se.docsMu.RUnlock()

	return se.documents[docID]
}

// DeleteDocument removes a document
func (se *SearchEngine) DeleteDocument(docID string) error {
	se.docsMu.Lock()
	doc := se.documents[docID]
	delete(se.documents, docID)
	se.docsMu.Unlock()

	if doc == nil {
		return fmt.Errorf("document not found")
	}

	// Remove from index
	se.index.termsMu.Lock()
	defer se.index.termsMu.Unlock()

	for _, docs := range se.index.terms {
		delete(docs, docID)
	}

	se.index.metaMu.Lock()
	delete(se.index.docMeta, docID)
	se.index.docCount--
	se.index.metaMu.Unlock()

	// Invalidate cache
	se.invalidateCache()

	return nil
}

// UpdateDocument updates a document
func (se *SearchEngine) UpdateDocument(doc *Document) error {
	se.DeleteDocument(doc.ID)
	return se.AddDocument(doc)
}

// ========== Query Types ==========

// PhraseSearch performs a phrase search
func (se *SearchEngine) PhraseSearch(phrase string) ([]*SearchResult, error) {
	query := &SearchQuery{
		Q:    phrase,
		Size: 10,
	}
	return se.Search(query)
}

// MatchAll returns all documents
func (se *SearchEngine) MatchAll() ([]*Document, error) {
	se.docsMu.RLock()
	defer se.docsMu.RUnlock()

	var docs []*Document
	for _, doc := range se.documents {
		docs = append(docs, doc)
	}

	return docs, nil
}

// TermSearch searches for exact term matches
func (se *SearchEngine) TermSearch(term string) ([]*SearchResult, error) {
	se.index.termsMu.RLock()
	docIDs := se.index.terms[term]
	se.index.termsMu.RUnlock()

	results := make(map[string]bool)
	for docID := range docIDs {
		results[docID] = true
	}

	query := &SearchQuery{Q: term, Size: 10}
	return se.scoreResults(results, query), nil
}

// ========== Statistics ==========

func (se *SearchEngine) recordStats(resultCount int, queryTime time.Duration) {
	se.statsMu.Lock()
	defer se.statsMu.Unlock()

	se.searchStats.TotalQueries++
	se.searchStats.TotalResults += int64(resultCount)
	se.searchStats.AvgQueryTime = (se.searchStats.AvgQueryTime + queryTime) / 2
	se.searchStats.CacheMisses++
}

func (se *SearchEngine) recordStatsCacheHit() {
	se.statsMu.Lock()
	defer se.statsMu.Unlock()

	se.searchStats.TotalQueries++
	se.searchStats.CacheHits++
}

// GetStats returns search statistics
func (se *SearchEngine) GetStats() *SearchStats {
	se.statsMu.RLock()
	defer se.statsMu.RUnlock()

	return se.searchStats
}

// ========== Index Management ==========

// GetIndexStats returns index statistics
func (se *SearchEngine) GetIndexStats() map[string]interface{} {
	se.index.metaMu.RLock()
	docCount := se.index.docCount
	se.index.metaMu.RUnlock()

	se.index.termsMu.RLock()
	termCount := len(se.index.terms)
	se.index.termsMu.RUnlock()

	return map[string]interface{}{
		"document_count": docCount,
		"term_count":     termCount,
	}
}

// InvalidateCache clears the query cache
func (se *SearchEngine) invalidateCache() {
	se.cacheMu.Lock()
	se.queryCache = make(map[string][]*SearchResult)
	se.cacheMu.Unlock()
}

func main() {
	// Example usage
	engine := NewSearchEngine()
	doc := &Document{
		ID:      "doc-1",
		Title:   "Go Programming",
		Content: "Learn Go programming language",
		Tags:    []string{"golang", "programming"},
	}
	engine.AddDocument(doc)
}

// ========== Helper Functions ==========

func tokenize(text string) []string {
	text = strings.ToLower(text)
	words := strings.FieldsFunc(text, func(r rune) bool {
		return r < 'a' || r > 'z' && r < '0' || r > '9'
	})

	var tokens []string
	for _, word := range words {
		if len(word) > 2 {
			tokens = append(tokens, word)
		}
	}

	return tokens
}

func calculateScore(doc *Document, tokens []string, query *SearchQuery) float64 {
	score := 0.0

	// TF-IDF-like scoring
	titleTokens := tokenize(doc.Title)
	contentTokens := tokenize(doc.Content)

	for _, token := range tokens {
		// Title match (boost)
		for _, t := range titleTokens {
			if t == token {
				score += 2.0
			}
		}

		// Content match
		for _, t := range contentTokens {
			if t == token {
				score += 1.0
			}
		}
	}

	return math.Min(score, 100.0)
}
