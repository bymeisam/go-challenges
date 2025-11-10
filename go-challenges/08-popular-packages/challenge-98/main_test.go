package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGorillaMux(t *testing.T) {
	router := NewRouter()

	// Test GET /products
	req := httptest.NewRequest("GET", "/products", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var products []Product
	json.NewDecoder(w.Body).Decode(&products)
	if len(products) != 2 {
		t.Errorf("Expected 2 products, got %d", len(products))
	}

	// Test GET /products/{id}
	req = httptest.NewRequest("GET", "/products/123", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var product Product
	json.NewDecoder(w.Body).Decode(&product)
	if product.ID != "123" {
		t.Errorf("Expected product ID '123', got '%s'", product.ID)
	}

	// Test POST /products
	newProduct := Product{Name: "New Product", Price: 150}
	body, _ := json.Marshal(newProduct)
	req = httptest.NewRequest("POST", "/products", bytes.NewReader(body))
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", w.Code)
	}

	// Test DELETE /products/{id}
	req = httptest.NewRequest("DELETE", "/products/123", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Expected status 204, got %d", w.Code)
	}

	t.Log("✓ gorilla/mux works!")
}
