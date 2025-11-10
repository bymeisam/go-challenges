package main

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestRequestParser(t *testing.T) {
	parser := NewRequestParser()

	t.Run("ParseForm", func(t *testing.T) {
		formData := url.Values{}
		formData.Set("name", "John Doe")
		formData.Set("email", "john@example.com")
		formData.Set("age", "30")

		req := httptest.NewRequest("POST", "/parse-form", strings.NewReader(formData.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()

		parser.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var resp map[string]interface{}
		json.NewDecoder(w.Body).Decode(&resp)

		if resp["message"] != "Form parsed successfully" {
			t.Errorf("Unexpected message: %v", resp["message"])
		}

		data := resp["data"].(map[string]interface{})
		if data["name"] != "John Doe" {
			t.Errorf("Expected name 'John Doe', got %v", data["name"])
		}
		if data["email"] != "john@example.com" {
			t.Errorf("Expected email 'john@example.com', got %v", data["email"])
		}
	})

	t.Run("ParseJSON", func(t *testing.T) {
		jsonData := map[string]string{
			"username": "alice",
			"password": "secret123",
		}
		jsonBytes, _ := json.Marshal(jsonData)

		req := httptest.NewRequest("POST", "/parse-json", bytes.NewReader(jsonBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		parser.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var resp map[string]interface{}
		json.NewDecoder(w.Body).Decode(&resp)

		if resp["message"] != "JSON parsed successfully" {
			t.Errorf("Unexpected message: %v", resp["message"])
		}

		data := resp["data"].(map[string]interface{})
		if data["username"] != "alice" {
			t.Errorf("Expected username 'alice', got %v", data["username"])
		}
	})

	t.Run("ParseMultipart", func(t *testing.T) {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)

		// Add form fields
		writer.WriteField("title", "Test Document")
		writer.WriteField("description", "This is a test")

		// Add file
		part, _ := writer.CreateFormFile("file", "test.txt")
		part.Write([]byte("test file content"))

		writer.Close()

		req := httptest.NewRequest("POST", "/parse-multipart", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		w := httptest.NewRecorder()

		parser.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var resp map[string]interface{}
		json.NewDecoder(w.Body).Decode(&resp)

		if resp["message"] != "Multipart form parsed successfully" {
			t.Errorf("Unexpected message: %v", resp["message"])
		}

		if resp["title"] != "Test Document" {
			t.Errorf("Expected title 'Test Document', got %v", resp["title"])
		}

		fileInfo := resp["file"].(map[string]interface{})
		if fileInfo["filename"] != "test.txt" {
			t.Errorf("Expected filename 'test.txt', got %v", fileInfo["filename"])
		}
		if fileInfo["content"] != "test file content" {
			t.Errorf("Expected content 'test file content', got %v", fileInfo["content"])
		}
	})

	t.Log("✓ Request parsing works!")
}
