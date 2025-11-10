package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type StaticFileServer struct {
	mux       *http.ServeMux
	staticDir string
}

func NewStaticFileServer(staticDir string) *StaticFileServer {
	if staticDir == "" {
		staticDir = "./static"
	}

	server := &StaticFileServer{
		mux:       http.NewServeMux(),
		staticDir: staticDir,
	}

	// Create static directory if it doesn't exist
	os.MkdirAll(staticDir, 0755)

	server.routes()
	return server
}

func (s *StaticFileServer) routes() {
	s.mux.HandleFunc("/", s.handleHome())
	s.mux.HandleFunc("/static/", s.handleStatic())
}

func (s *StaticFileServer) handleHome() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintf(w, `<!DOCTYPE html>
<html>
<head>
    <title>Static File Server</title>
</head>
<body>
    <h1>Welcome to Static File Server</h1>
    <p>Access static files at <a href="/static/">/static/</a></p>
</body>
</html>`)
	}
}

func (s *StaticFileServer) handleStatic() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Remove /static/ prefix
		path := strings.TrimPrefix(r.URL.Path, "/static/")
		if path == "" || path == "/" {
			http.Error(w, "File path required", http.StatusBadRequest)
			return
		}

		// Prevent directory traversal
		if strings.Contains(path, "..") {
			http.Error(w, "Invalid file path", http.StatusBadRequest)
			return
		}

		// Build full file path
		filePath := filepath.Join(s.staticDir, path)

		// Check if file exists
		info, err := os.Stat(filePath)
		if err != nil {
			http.NotFound(w, r)
			return
		}

		// Don't serve directories
		if info.IsDir() {
			http.Error(w, "Cannot serve directory", http.StatusForbidden)
			return
		}

		// Set Content-Type based on file extension
		contentType := getContentType(filePath)
		w.Header().Set("Content-Type", contentType)

		// Set caching headers (cache for 1 hour)
		w.Header().Set("Cache-Control", "public, max-age=3600")
		w.Header().Set("ETag", fmt.Sprintf(`"%d-%d"`, info.Size(), info.ModTime().Unix()))

		// Serve the file
		http.ServeFile(w, r, filePath)
	}
}

func getContentType(filePath string) string {
	ext := strings.ToLower(filepath.Ext(filePath))

	contentTypes := map[string]string{
		".html": "text/html; charset=utf-8",
		".htm":  "text/html; charset=utf-8",
		".css":  "text/css; charset=utf-8",
		".js":   "application/javascript; charset=utf-8",
		".json": "application/json",
		".xml":  "application/xml",
		".txt":  "text/plain; charset=utf-8",
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".png":  "image/png",
		".gif":  "image/gif",
		".svg":  "image/svg+xml",
		".pdf":  "application/pdf",
		".zip":  "application/zip",
	}

	if ct, ok := contentTypes[ext]; ok {
		return ct
	}

	return "application/octet-stream"
}

func (s *StaticFileServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func main() {}
