package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Custom handler type
type GreetHandler struct {
	greeting string
}

func (h *GreetHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	if name == "" {
		name = "Guest"
	}
	fmt.Fprintf(w, "%s, %s!", h.greeting, name)
}

// Handler function
func TimeHandler(w http.ResponseWriter, r *http.Request) {
	currentTime := time.Now().Format(time.RFC3339)
	json.NewEncoder(w).Encode(map[string]string{
		"time": currentTime,
	})
}

// Handler that returns a handler
func MessageHandler(message string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{
			"message": message,
		})
	}
}

// Method-based routing handler
func MethodHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		json.NewEncoder(w).Encode(map[string]string{"method": "GET"})
	case http.MethodPost:
		json.NewEncoder(w).Encode(map[string]string{"method": "POST"})
	case http.MethodPut:
		json.NewEncoder(w).Encode(map[string]string{"method": "PUT"})
	case http.MethodDelete:
		json.NewEncoder(w).Encode(map[string]string{"method": "DELETE"})
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func NewMux() *http.ServeMux {
	mux := http.NewServeMux()
	
	// Custom handler
	mux.Handle("/greet", &GreetHandler{greeting: "Welcome"})
	
	// Handler function
	mux.HandleFunc("/time", TimeHandler)
	
	// Handler with closure
	mux.HandleFunc("/custom", MessageHandler("This is a custom message"))
	
	// Method-based handler
	mux.HandleFunc("/method", MethodHandler)
	
	return mux
}

func main() {}
