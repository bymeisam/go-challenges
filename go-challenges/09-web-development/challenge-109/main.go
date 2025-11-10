package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

type Book struct {
	ID     int    `json:"id"`
	Title  string `json:"title"`
	Author string `json:"author"`
	ISBN   string `json:"isbn"`
}

type BookAPI struct {
	mux     *http.ServeMux
	books   map[int]*Book
	nextID  int
	mu      sync.RWMutex
}

func NewBookAPI() *BookAPI {
	api := &BookAPI{
		mux:    http.NewServeMux(),
		books:  make(map[int]*Book),
		nextID: 1,
	}
	api.routes()
	api.seedData()
	return api
}

func (api *BookAPI) seedData() {
	api.books[1] = &Book{ID: 1, Title: "The Go Programming Language", Author: "Alan Donovan", ISBN: "978-0134190440"}
	api.books[2] = &Book{ID: 2, Title: "Go in Action", Author: "William Kennedy", ISBN: "978-1617291784"}
	api.nextID = 3
}

func (api *BookAPI) routes() {
	api.mux.HandleFunc("/api/books", api.handleBooks())
	api.mux.HandleFunc("/api/books/", api.handleBook())
}

func (api *BookAPI) handleBooks() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			api.listBooks(w, r)
		case http.MethodPost:
			api.createBook(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func (api *BookAPI) handleBook() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract ID from path
		idStr := strings.TrimPrefix(r.URL.Path, "/api/books/")
		if idStr == "" {
			http.Error(w, "Book ID required", http.StatusBadRequest)
			return
		}

		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "Invalid book ID", http.StatusBadRequest)
			return
		}

		switch r.Method {
		case http.MethodGet:
			api.getBook(w, r, id)
		case http.MethodPut:
			api.updateBook(w, r, id)
		case http.MethodDelete:
			api.deleteBook(w, r, id)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func (api *BookAPI) listBooks(w http.ResponseWriter, r *http.Request) {
	api.mu.RLock()
	defer api.mu.RUnlock()

	books := make([]*Book, 0, len(api.books))
	for _, book := range api.books {
		books = append(books, book)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(books)
}

func (api *BookAPI) getBook(w http.ResponseWriter, r *http.Request, id int) {
	api.mu.RLock()
	defer api.mu.RUnlock()

	book, exists := api.books[id]
	if !exists {
		http.Error(w, "Book not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(book)
}

func (api *BookAPI) createBook(w http.ResponseWriter, r *http.Request) {
	var book Book
	if err := json.NewDecoder(r.Body).Decode(&book); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	api.mu.Lock()
	book.ID = api.nextID
	api.nextID++
	api.books[book.ID] = &book
	api.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(book)
}

func (api *BookAPI) updateBook(w http.ResponseWriter, r *http.Request, id int) {
	api.mu.Lock()
	defer api.mu.Unlock()

	if _, exists := api.books[id]; !exists {
		http.Error(w, "Book not found", http.StatusNotFound)
		return
	}

	var book Book
	if err := json.NewDecoder(r.Body).Decode(&book); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	book.ID = id
	api.books[id] = &book

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(book)
}

func (api *BookAPI) deleteBook(w http.ResponseWriter, r *http.Request, id int) {
	api.mu.Lock()
	defer api.mu.Unlock()

	if _, exists := api.books[id]; !exists {
		http.Error(w, "Book not found", http.StatusNotFound)
		return
	}

	delete(api.books, id)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": fmt.Sprintf("Book %d deleted successfully", id),
	})
}

func (api *BookAPI) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	api.mux.ServeHTTP(w, r)
}

func main() {}
