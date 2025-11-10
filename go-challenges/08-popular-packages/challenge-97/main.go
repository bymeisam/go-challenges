package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Response struct {
	Message string `json:"message"`
	UserID  string `json:"user_id,omitempty"`
}

func NewRouter() chi.Router {
	r := chi.NewRouter()
	
	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Timeout(60 * time.Second))
	
	// Routes
	r.Get("/", HandleHome)
	r.Get("/users/{userID}", HandleUser)
	r.Post("/users", HandleCreateUser)
	
	// Subrouter
	r.Route("/api", func(r chi.Router) {
		r.Get("/health", HandleHealth)
	})
	
	return r
}

func HandleHome(w http.ResponseWriter, r *http.Request) {
	resp := Response{Message: "Welcome to chi router"}
	json.NewEncoder(w).Encode(resp)
}

func HandleUser(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userID")
	resp := Response{
		Message: "User details",
		UserID:  userID,
	}
	json.NewEncoder(w).Encode(resp)
}

func HandleCreateUser(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusCreated)
	resp := Response{Message: "User created"}
	json.NewEncoder(w).Encode(resp)
}

func HandleHealth(w http.ResponseWriter, r *http.Request) {
	resp := Response{Message: "OK"}
	json.NewEncoder(w).Encode(resp)
}

func main() {}
