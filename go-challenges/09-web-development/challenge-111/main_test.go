package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestStaticFileServer(t *testing.T) {
	// Create temporary static directory
	tempDir := t.TempDir()
	server := NewStaticFileServer(tempDir)

	// Create test files
	testFiles := map[string]string{
		"index.html":  "<!DOCTYPE html><html><body>Test HTML</body></html>",
		"style.css":   "body { color: red; }",
		"script.js":   "console.log('test');",
		"data.json":   `{"test": "data"}`,
		"image.png":   "fake png data",
		"document.pdf": "fake pdf data",
	}

	for filename, content := range testFiles {
		filePath := filepath.Join(tempDir, filename)
		os.WriteFile(filePath, []byte(content), 0644)
	}

	t.Run("HomeRoute", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		w := httptest.NewRecorder()

		server.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		contentType := w.Header().Get("Content-Type")
		if !strings.Contains(contentType, "text/html") {
			t.Errorf("Expected Content-Type to contain 'text/html', got '%s'", contentType)
		}

		body := w.Body.String()
		if !strings.Contains(body, "Static File Server") {
			t.Error("Expected home page content")
		}
	})

	t.Run("ServeHTMLFile", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/static/index.html", nil)
		w := httptest.NewRecorder()

		server.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		contentType := w.Header().Get("Content-Type")
		if !strings.Contains(contentType, "text/html") {
			t.Errorf("Expected Content-Type 'text/html', got '%s'", contentType)
		}

		body := w.Body.String()
		if !strings.Contains(body, "Test HTML") {
			t.Error("Expected HTML file content")
		}
	})

	t.Run("ServeCSSFile", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/static/style.css", nil)
		w := httptest.NewRecorder()

		server.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		contentType := w.Header().Get("Content-Type")
		if !strings.Contains(contentType, "text/css") {
			t.Errorf("Expected Content-Type 'text/css', got '%s'", contentType)
		}
	})

	t.Run("ServeJavaScriptFile", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/static/script.js", nil)
		w := httptest.NewRecorder()

		server.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		contentType := w.Header().Get("Content-Type")
		if !strings.Contains(contentType, "javascript") {
			t.Errorf("Expected Content-Type 'javascript', got '%s'", contentType)
		}
	})

	t.Run("ServeJSONFile", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/static/data.json", nil)
		w := httptest.NewRecorder()

		server.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		contentType := w.Header().Get("Content-Type")
		if !strings.Contains(contentType, "application/json") {
			t.Errorf("Expected Content-Type 'application/json', got '%s'", contentType)
		}
	})

	t.Run("CachingHeaders", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/static/index.html", nil)
		w := httptest.NewRecorder()

		server.ServeHTTP(w, req)

		cacheControl := w.Header().Get("Cache-Control")
		if !strings.Contains(cacheControl, "max-age=3600") {
			t.Errorf("Expected Cache-Control with max-age=3600, got '%s'", cacheControl)
		}

		etag := w.Header().Get("ETag")
		if etag == "" {
			t.Error("Expected ETag header to be set")
		}
	})

	t.Run("FileNotFound", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/static/notfound.html", nil)
		w := httptest.NewRecorder()

		server.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status 404, got %d", w.Code)
		}
	})

	t.Run("DirectoryTraversalPrevention", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/static/../../../etc/passwd", nil)
		w := httptest.NewRecorder()

		server.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400 for directory traversal, got %d", w.Code)
		}
	})

	t.Run("EmptyPath", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/static/", nil)
		w := httptest.NewRecorder()

		server.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400 for empty path, got %d", w.Code)
		}
	})

	t.Run("NotFoundRoute", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/notfound", nil)
		w := httptest.NewRecorder()

		server.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status 404, got %d", w.Code)
		}
	})

	t.Log("✓ Static file server works!")
}
