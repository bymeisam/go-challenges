package main

import (
	"encoding/json"
	"net/http"
	"strings"
)

type CORSConfig struct {
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	ExposedHeaders   []string
	AllowCredentials bool
	MaxAge           int // in seconds
}

func DefaultCORSConfig() *CORSConfig {
	return &CORSConfig{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Content-Type", "Authorization"},
		ExposedHeaders:   []string{},
		AllowCredentials: false,
		MaxAge:           3600, // 1 hour
	}
}

type CORSServer struct {
	mux    *http.ServeMux
	config *CORSConfig
}

func NewCORSServer(config *CORSConfig) *CORSServer {
	if config == nil {
		config = DefaultCORSConfig()
	}

	server := &CORSServer{
		mux:    http.NewServeMux(),
		config: config,
	}
	server.routes()
	return server
}

func (s *CORSServer) routes() {
	s.mux.HandleFunc("/api/data", s.corsMiddleware(s.handleData()))
	s.mux.HandleFunc("/api/public", s.corsMiddleware(s.handlePublic()))
}

func (s *CORSServer) corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		// Check if origin is allowed
		if origin != "" && s.isOriginAllowed(origin) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		} else if s.isOriginAllowed("*") {
			w.Header().Set("Access-Control-Allow-Origin", "*")
		}

		// Set allowed methods
		if len(s.config.AllowedMethods) > 0 {
			w.Header().Set("Access-Control-Allow-Methods", strings.Join(s.config.AllowedMethods, ", "))
		}

		// Set allowed headers
		if len(s.config.AllowedHeaders) > 0 {
			w.Header().Set("Access-Control-Allow-Headers", strings.Join(s.config.AllowedHeaders, ", "))
		}

		// Set exposed headers
		if len(s.config.ExposedHeaders) > 0 {
			w.Header().Set("Access-Control-Expose-Headers", strings.Join(s.config.ExposedHeaders, ", "))
		}

		// Set credentials
		if s.config.AllowCredentials {
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}

		// Set max age
		if s.config.MaxAge > 0 {
			w.Header().Set("Access-Control-Max-Age", string(rune(s.config.MaxAge)))
		}

		// Handle preflight request
		if r.Method == http.MethodOptions {
			// Check if this is a valid preflight request
			requestMethod := r.Header.Get("Access-Control-Request-Method")
			if requestMethod != "" {
				// Validate the requested method
				if !s.isMethodAllowed(requestMethod) {
					w.WriteHeader(http.StatusForbidden)
					return
				}

				// Validate requested headers
				requestHeaders := r.Header.Get("Access-Control-Request-Headers")
				if requestHeaders != "" && !s.areHeadersAllowed(requestHeaders) {
					w.WriteHeader(http.StatusForbidden)
					return
				}

				w.WriteHeader(http.StatusNoContent)
				return
			}
		}

		// Call next handler
		next(w, r)
	}
}

func (s *CORSServer) isOriginAllowed(origin string) bool {
	for _, allowed := range s.config.AllowedOrigins {
		if allowed == "*" || allowed == origin {
			return true
		}
	}
	return false
}

func (s *CORSServer) isMethodAllowed(method string) bool {
	for _, allowed := range s.config.AllowedMethods {
		if strings.EqualFold(allowed, method) {
			return true
		}
	}
	return false
}

func (s *CORSServer) areHeadersAllowed(headers string) bool {
	requestedHeaders := strings.Split(headers, ",")
	for _, header := range requestedHeaders {
		header = strings.TrimSpace(header)
		if !s.isHeaderAllowed(header) {
			return false
		}
	}
	return true
}

func (s *CORSServer) isHeaderAllowed(header string) bool {
	for _, allowed := range s.config.AllowedHeaders {
		if strings.EqualFold(allowed, header) {
			return true
		}
	}
	return false
}

func (s *CORSServer) handleData() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{
				"message": "GET data",
			})
		case http.MethodPost:
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{
				"message": "POST data created",
			})
		case http.MethodPut:
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{
				"message": "PUT data updated",
			})
		case http.MethodDelete:
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{
				"message": "DELETE data deleted",
			})
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func (s *CORSServer) handlePublic() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Public endpoint",
		})
	}
}

func (s *CORSServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func main() {}
