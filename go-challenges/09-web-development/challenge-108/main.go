package main

import (
	"encoding/json"
	"encoding/xml"
	"html/template"
	"net/http"
)

type User struct {
	ID       int    `json:"id" xml:"id"`
	Name     string `json:"name" xml:"name"`
	Email    string `json:"email" xml:"email"`
	Username string `json:"username" xml:"username"`
}

type ResponseWriter struct {
	mux *http.ServeMux
}

func NewResponseWriter() *ResponseWriter {
	rw := &ResponseWriter{
		mux: http.NewServeMux(),
	}
	rw.routes()
	return rw
}

func (rw *ResponseWriter) routes() {
	rw.mux.HandleFunc("/json", rw.handleJSON())
	rw.mux.HandleFunc("/xml", rw.handleXML())
	rw.mux.HandleFunc("/html", rw.handleHTML())
}

func (rw *ResponseWriter) handleJSON() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := User{
			ID:       1,
			Name:     "John Doe",
			Email:    "john@example.com",
			Username: "johndoe",
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		if err := json.NewEncoder(w).Encode(user); err != nil {
			http.Error(w, "Failed to encode JSON", http.StatusInternalServerError)
		}
	}
}

func (rw *ResponseWriter) handleXML() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := User{
			ID:       2,
			Name:     "Jane Smith",
			Email:    "jane@example.com",
			Username: "janesmith",
		}

		w.Header().Set("Content-Type", "application/xml")
		w.WriteHeader(http.StatusOK)

		// Write XML declaration
		w.Write([]byte(xml.Header))

		if err := xml.NewEncoder(w).Encode(user); err != nil {
			http.Error(w, "Failed to encode XML", http.StatusInternalServerError)
		}
	}
}

func (rw *ResponseWriter) handleHTML() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := User{
			ID:       3,
			Name:     "Bob Johnson",
			Email:    "bob@example.com",
			Username: "bobjohnson",
		}

		htmlTemplate := `<!DOCTYPE html>
<html>
<head>
    <title>User Profile</title>
</head>
<body>
    <h1>User Profile</h1>
    <div>
        <p><strong>ID:</strong> {{.ID}}</p>
        <p><strong>Name:</strong> {{.Name}}</p>
        <p><strong>Email:</strong> {{.Email}}</p>
        <p><strong>Username:</strong> {{.Username}}</p>
    </div>
</body>
</html>`

		tmpl, err := template.New("user").Parse(htmlTemplate)
		if err != nil {
			http.Error(w, "Failed to parse template", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)

		if err := tmpl.Execute(w, user); err != nil {
			http.Error(w, "Failed to execute template", http.StatusInternalServerError)
		}
	}
}

func (rw *ResponseWriter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rw.mux.ServeHTTP(w, r)
}

func main() {}
