package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestBookAPI(t *testing.T) {
	api := NewBookAPI()

	t.Run("ListBooks", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/books", nil)
		w := httptest.NewRecorder()

		api.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var books []*Book
		json.NewDecoder(w.Body).Decode(&books)

		if len(books) != 2 {
			t.Errorf("Expected 2 books, got %d", len(books))
		}
	})

	t.Run("GetBook", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/books/1", nil)
		w := httptest.NewRecorder()

		api.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var book Book
		json.NewDecoder(w.Body).Decode(&book)

		if book.ID != 1 {
			t.Errorf("Expected book ID 1, got %d", book.ID)
		}
		if book.Title != "The Go Programming Language" {
			t.Errorf("Unexpected book title: %s", book.Title)
		}
	})

	t.Run("GetBookNotFound", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/books/999", nil)
		w := httptest.NewRecorder()

		api.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status 404, got %d", w.Code)
		}
	})

	t.Run("CreateBook", func(t *testing.T) {
		newBook := Book{
			Title:  "Concurrency in Go",
			Author: "Katherine Cox-Buday",
			ISBN:   "978-1491941294",
		}
		jsonData, _ := json.Marshal(newBook)

		req := httptest.NewRequest("POST", "/api/books", bytes.NewReader(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		api.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("Expected status 201, got %d", w.Code)
		}

		var createdBook Book
		json.NewDecoder(w.Body).Decode(&createdBook)

		if createdBook.ID == 0 {
			t.Error("Expected book to have an ID")
		}
		if createdBook.Title != newBook.Title {
			t.Errorf("Expected title '%s', got '%s'", newBook.Title, createdBook.Title)
		}
	})

	t.Run("UpdateBook", func(t *testing.T) {
		updatedBook := Book{
			Title:  "Updated Title",
			Author: "Updated Author",
			ISBN:   "978-0000000000",
		}
		jsonData, _ := json.Marshal(updatedBook)

		req := httptest.NewRequest("PUT", "/api/books/1", bytes.NewReader(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		api.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var book Book
		json.NewDecoder(w.Body).Decode(&book)

		if book.Title != "Updated Title" {
			t.Errorf("Expected title 'Updated Title', got '%s'", book.Title)
		}
		if book.ID != 1 {
			t.Errorf("Expected ID to remain 1, got %d", book.ID)
		}
	})

	t.Run("UpdateBookNotFound", func(t *testing.T) {
		updatedBook := Book{
			Title:  "Test",
			Author: "Test",
			ISBN:   "000",
		}
		jsonData, _ := json.Marshal(updatedBook)

		req := httptest.NewRequest("PUT", "/api/books/999", bytes.NewReader(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		api.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status 404, got %d", w.Code)
		}
	})

	t.Run("DeleteBook", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/api/books/2", nil)
		w := httptest.NewRecorder()

		api.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		// Verify book is deleted
		req = httptest.NewRequest("GET", "/api/books/2", nil)
		w = httptest.NewRecorder()
		api.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status 404 after deletion, got %d", w.Code)
		}
	})

	t.Run("DeleteBookNotFound", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/api/books/999", nil)
		w := httptest.NewRecorder()

		api.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status 404, got %d", w.Code)
		}
	})

	t.Log("✓ JSON REST API works!")
}
