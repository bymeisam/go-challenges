package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type FormData struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Age   string `json:"age"`
}

type JSONData struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type RequestParser struct {
	mux *http.ServeMux
}

func NewRequestParser() *RequestParser {
	p := &RequestParser{
		mux: http.NewServeMux(),
	}
	p.routes()
	return p
}

func (p *RequestParser) routes() {
	p.mux.HandleFunc("/parse-form", p.handleParseForm())
	p.mux.HandleFunc("/parse-json", p.handleParseJSON())
	p.mux.HandleFunc("/parse-multipart", p.handleParseMultipart())
}

func (p *RequestParser) handleParseForm() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Parse form data
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Failed to parse form", http.StatusBadRequest)
			return
		}

		formData := FormData{
			Name:  r.FormValue("name"),
			Email: r.FormValue("email"),
			Age:   r.FormValue("age"),
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "Form parsed successfully",
			"data":    formData,
		})
	}
}

func (p *RequestParser) handleParseJSON() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var jsonData JSONData
		if err := json.NewDecoder(r.Body).Decode(&jsonData); err != nil {
			http.Error(w, "Failed to parse JSON", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "JSON parsed successfully",
			"data":    jsonData,
		})
	}
}

func (p *RequestParser) handleParseMultipart() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Parse multipart form with max memory of 10 MB
		if err := r.ParseMultipartForm(10 << 20); err != nil {
			http.Error(w, "Failed to parse multipart form", http.StatusBadRequest)
			return
		}

		// Get form values
		title := r.FormValue("title")
		description := r.FormValue("description")

		// Get uploaded file
		file, header, err := r.FormFile("file")
		var fileInfo map[string]interface{}
		if err == nil {
			defer file.Close()

			// Read file content
			content, _ := io.ReadAll(file)

			fileInfo = map[string]interface{}{
				"filename": header.Filename,
				"size":     header.Size,
				"content":  string(content),
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message":     "Multipart form parsed successfully",
			"title":       title,
			"description": description,
			"file":        fileInfo,
		})
	}
}

func (p *RequestParser) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	p.mux.ServeHTTP(w, r)
}

func main() {}
