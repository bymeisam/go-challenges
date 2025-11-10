package main

import (
	"context"
	"testing"

	"github.com/graphql-go/graphql"
)

func setupTestStore() *Store {
	store := NewStore()

	// Create test authors
	author1, _ := store.CreateAuthor("Test Author 1", strPtr("Bio 1"))
	author2, _ := store.CreateAuthor("Test Author 2", nil)

	// Create test books
	rating1 := 4.5
	store.CreateBook("Book 1", "ISBN-001", 2020, author1.ID, GenreFiction, &rating1)

	rating2 := 3.8
	store.CreateBook("Book 2", "ISBN-002", 2021, author1.ID, GenreSciFi, &rating2)

	rating3 := 4.9
	store.CreateBook("Book 3", "ISBN-003", 2019, author2.ID, GenreFantasy, &rating3)

	return store
}

func TestCreateBook(t *testing.T) {
	store := NewStore()
	author, _ := store.CreateAuthor("Author", nil)

	tests := []struct {
		name          string
		title         string
		isbn          string
		publishedYear int
		authorID      string
		genre         Genre
		rating        *float64
		wantErr       bool
	}{
		{
			name:          "valid book",
			title:         "Test Book",
			isbn:          "978-0134190440",
			publishedYear: 2020,
			authorID:      author.ID,
			genre:         GenreFiction,
			rating:        floatPtr(4.5),
			wantErr:       false,
		},
		{
			name:          "empty title",
			title:         "",
			isbn:          "978-0134190440",
			publishedYear: 2020,
			authorID:      author.ID,
			genre:         GenreFiction,
			wantErr:       true,
		},
		{
			name:          "invalid year",
			title:         "Test Book",
			isbn:          "978-0134190440",
			publishedYear: 999,
			authorID:      author.ID,
			genre:         GenreFiction,
			wantErr:       true,
		},
		{
			name:          "invalid rating",
			title:         "Test Book",
			isbn:          "978-0134190440",
			publishedYear: 2020,
			authorID:      author.ID,
			genre:         GenreFiction,
			rating:        floatPtr(6.0),
			wantErr:       true,
		},
		{
			name:          "invalid author",
			title:         "Test Book",
			isbn:          "978-0134190440",
			publishedYear: 2020,
			authorID:      "999",
			genre:         GenreFiction,
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			book, err := store.CreateBook(tt.title, tt.isbn, tt.publishedYear, tt.authorID, tt.genre, tt.rating)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateBook() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if book.Title != tt.title {
					t.Errorf("CreateBook() title = %v, want %v", book.Title, tt.title)
				}
				if book.ISBN != tt.isbn {
					t.Errorf("CreateBook() isbn = %v, want %v", book.ISBN, tt.isbn)
				}
			}
		})
	}
}

func TestGetBook(t *testing.T) {
	store := setupTestStore()

	tests := []struct {
		name    string
		id      string
		wantErr bool
	}{
		{"existing book", "1", false},
		{"non-existing book", "999", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			book, err := store.GetBook(tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetBook() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && book.ID != tt.id {
				t.Errorf("GetBook() id = %v, want %v", book.ID, tt.id)
			}
		})
	}
}

func TestGetBooksWithFilter(t *testing.T) {
	store := setupTestStore()

	tests := []struct {
		name        string
		filter      *BookFilter
		wantCount   int
		description string
	}{
		{
			name:        "no filter",
			filter:      nil,
			wantCount:   3,
			description: "should return all books",
		},
		{
			name: "filter by genre",
			filter: &BookFilter{
				Genre: genrePtr(GenreFiction),
			},
			wantCount:   1,
			description: "should return only fiction books",
		},
		{
			name: "filter by min rating",
			filter: &BookFilter{
				MinRating: floatPtr(4.0),
			},
			wantCount:   2,
			description: "should return books with rating >= 4.0",
		},
		{
			name: "filter by title",
			filter: &BookFilter{
				Title: strPtr("Book 1"),
			},
			wantCount:   1,
			description: "should return books matching title",
		},
		{
			name: "filter by author",
			filter: &BookFilter{
				AuthorID: strPtr("1"),
			},
			wantCount:   2,
			description: "should return books by author 1",
		},
		{
			name: "multiple filters",
			filter: &BookFilter{
				Genre:     genrePtr(GenreFiction),
				MinRating: floatPtr(4.0),
			},
			wantCount:   1,
			description: "should apply all filters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			connection, err := store.GetBooks(tt.filter, nil)
			if err != nil {
				t.Errorf("GetBooks() error = %v", err)
				return
			}
			if len(connection.Edges) != tt.wantCount {
				t.Errorf("GetBooks() count = %v, want %v (%s)", len(connection.Edges), tt.wantCount, tt.description)
			}
		})
	}
}

func TestPagination(t *testing.T) {
	store := setupTestStore()

	// Add more books for pagination testing
	author, _ := store.GetAuthor("1")
	for i := 4; i <= 10; i++ {
		rating := 4.0
		store.CreateBook("Book "+string(rune(i+'0')), "ISBN-00"+string(rune(i+'0')), 2020, author.ID, GenreFiction, &rating)
	}

	tests := []struct {
		name         string
		pagination   *PaginationInput
		wantEdges    int
		wantNextPage bool
	}{
		{
			name:         "default pagination",
			pagination:   nil,
			wantEdges:    10,
			wantNextPage: false,
		},
		{
			name: "first 5",
			pagination: &PaginationInput{
				First: intPtr(5),
			},
			wantEdges:    5,
			wantNextPage: true,
		},
		{
			name: "first 3 after cursor",
			pagination: &PaginationInput{
				First: intPtr(3),
				After: strPtr(encodeCursor("2")),
			},
			wantEdges:    3,
			wantNextPage: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			connection, err := store.GetBooks(nil, tt.pagination)
			if err != nil {
				t.Errorf("GetBooks() error = %v", err)
				return
			}
			if len(connection.Edges) != tt.wantEdges {
				t.Errorf("GetBooks() edges = %v, want %v", len(connection.Edges), tt.wantEdges)
			}
			if connection.PageInfo.HasNextPage != tt.wantNextPage {
				t.Errorf("GetBooks() hasNextPage = %v, want %v", connection.PageInfo.HasNextPage, tt.wantNextPage)
			}
		})
	}
}

func TestUpdateBook(t *testing.T) {
	store := setupTestStore()

	newYear := 2022
	newRating := 5.0

	book, err := store.UpdateBook("1", "Updated Title", "New-ISBN", &newYear, &newRating)
	if err != nil {
		t.Fatalf("UpdateBook() error = %v", err)
	}

	if book.Title != "Updated Title" {
		t.Errorf("UpdateBook() title = %v, want %v", book.Title, "Updated Title")
	}
	if book.ISBN != "New-ISBN" {
		t.Errorf("UpdateBook() isbn = %v, want %v", book.ISBN, "New-ISBN")
	}
	if book.PublishedYear != newYear {
		t.Errorf("UpdateBook() year = %v, want %v", book.PublishedYear, newYear)
	}
	if book.Rating == nil || *book.Rating != newRating {
		t.Errorf("UpdateBook() rating = %v, want %v", book.Rating, newRating)
	}
}

func TestDeleteBook(t *testing.T) {
	store := setupTestStore()

	// Delete existing book
	deleted, err := store.DeleteBook("1")
	if err != nil {
		t.Fatalf("DeleteBook() error = %v", err)
	}
	if !deleted {
		t.Error("DeleteBook() should return true for existing book")
	}

	// Try to get deleted book
	_, err = store.GetBook("1")
	if err == nil {
		t.Error("GetBook() should return error for deleted book")
	}

	// Delete non-existing book
	_, err = store.DeleteBook("999")
	if err == nil {
		t.Error("DeleteBook() should return error for non-existing book")
	}
}

func TestCreateAuthor(t *testing.T) {
	store := NewStore()

	tests := []struct {
		name    string
		authName string
		bio     *string
		wantErr bool
	}{
		{"valid author with bio", "Author Name", strPtr("Bio"), false},
		{"valid author without bio", "Author Name", nil, false},
		{"empty name", "", nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			author, err := store.CreateAuthor(tt.authName, tt.bio)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateAuthor() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && author.Name != tt.authName {
				t.Errorf("CreateAuthor() name = %v, want %v", author.Name, tt.authName)
			}
		})
	}
}

func TestGetBooksByAuthor(t *testing.T) {
	store := setupTestStore()

	books, err := store.GetBooksByAuthor("1")
	if err != nil {
		t.Fatalf("GetBooksByAuthor() error = %v", err)
	}

	if len(books) != 2 {
		t.Errorf("GetBooksByAuthor() count = %v, want %v", len(books), 2)
	}

	for _, book := range books {
		if book.AuthorID != "1" {
			t.Errorf("GetBooksByAuthor() returned book with authorID = %v, want %v", book.AuthorID, "1")
		}
	}
}

func TestGraphQLQueryBooks(t *testing.T) {
	store := setupTestStore()
	schema, err := buildSchema(store)
	if err != nil {
		t.Fatal(err)
	}

	query := `
		query {
			books(pagination: {first: 10}) {
				edges {
					node {
						id
						title
						author {
							name
						}
					}
				}
				pageInfo {
					hasNextPage
				}
			}
		}
	`

	result := ExecuteQuery(schema, query, nil)
	if len(result.Errors) > 0 {
		t.Fatalf("GraphQL query failed: %v", result.Errors)
	}

	data := result.Data.(map[string]interface{})
	books := data["books"].(map[string]interface{})
	edges := books["edges"].([]interface{})

	if len(edges) != 3 {
		t.Errorf("Expected 3 books, got %d", len(edges))
	}
}

func TestGraphQLQueryBookWithFilter(t *testing.T) {
	store := setupTestStore()
	schema, err := buildSchema(store)
	if err != nil {
		t.Fatal(err)
	}

	query := `
		query {
			books(filter: {genre: FICTION}) {
				edges {
					node {
						id
						title
						genre
					}
				}
			}
		}
	`

	result := ExecuteQuery(schema, query, nil)
	if len(result.Errors) > 0 {
		t.Fatalf("GraphQL query failed: %v", result.Errors)
	}

	data := result.Data.(map[string]interface{})
	books := data["books"].(map[string]interface{})
	edges := books["edges"].([]interface{})

	if len(edges) != 1 {
		t.Errorf("Expected 1 fiction book, got %d", len(edges))
	}
}

func TestGraphQLMutationCreateBook(t *testing.T) {
	store := setupTestStore()
	schema, err := buildSchema(store)
	if err != nil {
		t.Fatal(err)
	}

	mutation := `
		mutation {
			createBook(input: {
				title: "New Book"
				isbn: "978-1234567890"
				publishedYear: 2023
				authorId: "1"
				genre: FANTASY
				rating: 4.7
			}) {
				id
				title
				isbn
				genre
				author {
					name
				}
			}
		}
	`

	result := ExecuteQuery(schema, mutation, nil)
	if len(result.Errors) > 0 {
		t.Fatalf("GraphQL mutation failed: %v", result.Errors)
	}

	data := result.Data.(map[string]interface{})
	book := data["createBook"].(map[string]interface{})

	if book["title"] != "New Book" {
		t.Errorf("Expected title 'New Book', got %v", book["title"])
	}
	if book["genre"] != "FANTASY" {
		t.Errorf("Expected genre 'FANTASY', got %v", book["genre"])
	}
}

func TestGraphQLMutationUpdateBook(t *testing.T) {
	store := setupTestStore()
	schema, err := buildSchema(store)
	if err != nil {
		t.Fatal(err)
	}

	mutation := `
		mutation {
			updateBook(id: "1", input: {
				title: "Updated Book Title"
				rating: 5.0
			}) {
				id
				title
				rating
			}
		}
	`

	result := ExecuteQuery(schema, mutation, nil)
	if len(result.Errors) > 0 {
		t.Fatalf("GraphQL mutation failed: %v", result.Errors)
	}

	data := result.Data.(map[string]interface{})
	book := data["updateBook"].(map[string]interface{})

	if book["title"] != "Updated Book Title" {
		t.Errorf("Expected title 'Updated Book Title', got %v", book["title"])
	}
	if book["rating"] != 5.0 {
		t.Errorf("Expected rating 5.0, got %v", book["rating"])
	}
}

func TestGraphQLMutationDeleteBook(t *testing.T) {
	store := setupTestStore()
	schema, err := buildSchema(store)
	if err != nil {
		t.Fatal(err)
	}

	mutation := `
		mutation {
			deleteBook(id: "1")
		}
	`

	result := ExecuteQuery(schema, mutation, nil)
	if len(result.Errors) > 0 {
		t.Fatalf("GraphQL mutation failed: %v", result.Errors)
	}

	data := result.Data.(map[string]interface{})
	deleted := data["deleteBook"].(bool)

	if !deleted {
		t.Error("Expected deleteBook to return true")
	}

	// Verify book is deleted
	_, err = store.GetBook("1")
	if err == nil {
		t.Error("Book should be deleted")
	}
}

func TestGraphQLQueryAuthorWithBooks(t *testing.T) {
	store := setupTestStore()
	schema, err := buildSchema(store)
	if err != nil {
		t.Fatal(err)
	}

	query := `
		query {
			author(id: "1") {
				id
				name
				bio
				books {
					id
					title
				}
			}
		}
	`

	result := ExecuteQuery(schema, query, nil)
	if len(result.Errors) > 0 {
		t.Fatalf("GraphQL query failed: %v", result.Errors)
	}

	data := result.Data.(map[string]interface{})
	author := data["author"].(map[string]interface{})
	books := author["books"].([]interface{})

	if len(books) != 2 {
		t.Errorf("Expected author to have 2 books, got %d", len(books))
	}
}

func TestGraphQLErrorHandling(t *testing.T) {
	store := setupTestStore()
	schema, err := buildSchema(store)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name          string
		query         string
		expectError   bool
		errorContains string
	}{
		{
			name: "invalid book id",
			query: `
				query {
					book(id: "999") {
						id
						title
					}
				}
			`,
			expectError:   true,
			errorContains: "book not found",
		},
		{
			name: "invalid author id in mutation",
			query: `
				mutation {
					createBook(input: {
						title: "Test"
						isbn: "123"
						publishedYear: 2020
						authorId: "999"
						genre: FICTION
					}) {
						id
					}
				}
			`,
			expectError:   true,
			errorContains: "author not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExecuteQuery(schema, tt.query, nil)
			hasErrors := len(result.Errors) > 0

			if hasErrors != tt.expectError {
				t.Errorf("Expected error: %v, got errors: %v", tt.expectError, result.Errors)
			}

			if tt.expectError && hasErrors {
				errorFound := false
				for _, err := range result.Errors {
					if contains(err.Message, tt.errorContains) {
						errorFound = true
						break
					}
				}
				if !errorFound {
					t.Errorf("Expected error containing '%s', got: %v", tt.errorContains, result.Errors)
				}
			}
		})
	}
}

func TestConcurrentAccess(t *testing.T) {
	store := setupTestStore()

	const goroutines = 10
	done := make(chan bool, goroutines)

	// Concurrent reads and writes
	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer func() { done <- true }()

			// Read
			store.GetBooks(nil, nil)

			// Write
			rating := 4.0
			store.CreateBook("Concurrent Book", "ISBN", 2020, "1", GenreFiction, &rating)

			// Update
			store.UpdateBook("1", "Updated", "", nil, nil)
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < goroutines; i++ {
		<-done
	}
}

func BenchmarkGetBooks(b *testing.B) {
	store := setupTestStore()

	// Add more books
	author, _ := store.GetAuthor("1")
	for i := 0; i < 100; i++ {
		rating := 4.0
		store.CreateBook("Book", "ISBN", 2020, author.ID, GenreFiction, &rating)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		store.GetBooks(nil, nil)
	}
}

func BenchmarkGraphQLQuery(b *testing.B) {
	store := setupTestStore()
	schema, _ := buildSchema(store)

	query := `
		query {
			books(pagination: {first: 10}) {
				edges {
					node {
						id
						title
						author {
							name
						}
					}
				}
			}
		}
	`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ExecuteQuery(schema, query, nil)
	}
}

// Helper functions
func floatPtr(f float64) *float64 {
	return &f
}

func intPtr(i int) *int {
	return &i
}

func genrePtr(g Genre) *Genre {
	return &g
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || (len(s) > 0 && (s[0:len(substr)] == substr || contains(s[1:], substr))))
}
