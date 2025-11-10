package main

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFileUpload(t *testing.T) {
	// Create temporary upload directory
	tempDir := t.TempDir()
	handler := NewFileUploadHandler(tempDir)

	t.Run("SuccessfulUpload", func(t *testing.T) {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)

		part, _ := writer.CreateFormFile("file", "test.txt")
		part.Write([]byte("This is a test file"))

		writer.Close()

		req := httptest.NewRequest("POST", "/upload", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var resp UploadResponse
		json.NewDecoder(w.Body).Decode(&resp)

		if !resp.Success {
			t.Errorf("Expected success, got: %s", resp.Message)
		}
		if resp.Filename != "test.txt" {
			t.Errorf("Expected filename 'test.txt', got '%s'", resp.Filename)
		}
		if resp.Size != 19 {
			t.Errorf("Expected size 19, got %d", resp.Size)
		}

		// Verify file was saved
		filePath := filepath.Join(tempDir, "test.txt")
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			t.Error("File was not saved")
		}
	})

	t.Run("InvalidFileType", func(t *testing.T) {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)

		part, _ := writer.CreateFormFile("file", "test.exe")
		part.Write([]byte("executable content"))

		writer.Close()

		req := httptest.NewRequest("POST", "/upload", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w.Code)
		}

		var resp UploadResponse
		json.NewDecoder(w.Body).Decode(&resp)

		if resp.Success {
			t.Error("Expected failure for invalid file type")
		}
		if !strings.Contains(resp.Message, "not allowed") {
			t.Errorf("Expected 'not allowed' message, got: %s", resp.Message)
		}
	})

	t.Run("NoFileUploaded", func(t *testing.T) {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		writer.Close()

		req := httptest.NewRequest("POST", "/upload", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w.Code)
		}

		var resp UploadResponse
		json.NewDecoder(w.Body).Decode(&resp)

		if resp.Success {
			t.Error("Expected failure when no file uploaded")
		}
	})

	t.Run("FileTooLarge", func(t *testing.T) {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)

		part, _ := writer.CreateFormFile("file", "large.txt")
		// Write more than MaxFileSize (10MB)
		largeData := make([]byte, MaxFileSize+1)
		part.Write(largeData)

		writer.Close()

		req := httptest.NewRequest("POST", "/upload", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w.Code)
		}

		var resp UploadResponse
		json.NewDecoder(w.Body).Decode(&resp)

		if resp.Success {
			t.Error("Expected failure for file too large")
		}
	})

	t.Run("MethodNotAllowed", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/upload", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("Expected status 405, got %d", w.Code)
		}
	})

	t.Run("MultipleFileTypes", func(t *testing.T) {
		fileTypes := []struct {
			name      string
			extension string
			allowed   bool
		}{
			{"image.jpg", ".jpg", true},
			{"image.png", ".png", true},
			{"document.pdf", ".pdf", true},
			{"script.sh", ".sh", false},
			{"binary.bin", ".bin", false},
		}

		for _, ft := range fileTypes {
			body := &bytes.Buffer{}
			writer := multipart.NewWriter(body)

			part, _ := writer.CreateFormFile("file", ft.name)
			part.Write([]byte("test content"))

			writer.Close()

			req := httptest.NewRequest("POST", "/upload", body)
			req.Header.Set("Content-Type", writer.FormDataContentType())
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			var resp UploadResponse
			json.NewDecoder(w.Body).Decode(&resp)

			if ft.allowed && !resp.Success {
				t.Errorf("Expected %s to be allowed", ft.name)
			}
			if !ft.allowed && resp.Success {
				t.Errorf("Expected %s to be rejected", ft.name)
			}
		}
	})

	t.Log("✓ File upload works!")
}
