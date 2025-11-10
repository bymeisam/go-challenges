package main

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

type Product struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Price int    `json:"price"`
}

func NewRouter() *mux.Router {
	r := mux.NewRouter()
	
	// Routes
	r.HandleFunc("/", HandleHome).Methods("GET")
	r.HandleFunc("/products", HandleGetProducts).Methods("GET")
	r.HandleFunc("/products", HandleCreateProduct).Methods("POST")
	r.HandleFunc("/products/{id}", HandleGetProduct).Methods("GET")
	r.HandleFunc("/products/{id}", HandleUpdateProduct).Methods("PUT")
	r.HandleFunc("/products/{id}", HandleDeleteProduct).Methods("DELETE")
	
	// Middleware
	r.Use(loggingMiddleware)
	
	return r
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simple logging middleware
		next.ServeHTTP(w, r)
	})
}

func HandleHome(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]string{"message": "Welcome"})
}

func HandleGetProducts(w http.ResponseWriter, r *http.Request) {
	products := []Product{
		{ID: "1", Name: "Product 1", Price: 100},
		{ID: "2", Name: "Product 2", Price: 200},
	}
	json.NewEncoder(w).Encode(products)
}

func HandleCreateProduct(w http.ResponseWriter, r *http.Request) {
	var product Product
	json.NewDecoder(r.Body).Decode(&product)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(product)
}

func HandleGetProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	product := Product{ID: id, Name: "Product " + id, Price: 100}
	json.NewEncoder(w).Encode(product)
}

func HandleUpdateProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	var product Product
	json.NewDecoder(r.Body).Decode(&product)
	product.ID = id
	json.NewEncoder(w).Encode(product)
}

func HandleDeleteProduct(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

func main() {}
