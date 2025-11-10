package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const (
	MaxFileSize = 10 << 20 // 10 MB
	UploadDir   = "./uploads"
)

type FileUploadHandler struct {
	mux       *http.ServeMux
	uploadDir string
}

type UploadResponse struct {
	Success  bool   `json:"success"`
	Message  string `json:"message"`
	Filename string `json:"filename,omitempty"`
	Size     int64  `json:"size,omitempty"`
}

func NewFileUploadHandler(uploadDir string) *FileUploadHandler {
	if uploadDir == "" {
		uploadDir = UploadDir
	}

	handler := &FileUploadHandler{
		mux:       http.NewServeMux(),
		uploadDir: uploadDir,
	}

	// Create upload directory if it doesn't exist
	os.MkdirAll(uploadDir, 0755)

	handler.routes()
	return handler
}

func (h *FileUploadHandler) routes() {
	h.mux.HandleFunc("/upload", h.handleUpload())
}

func (h *FileUploadHandler) handleUpload() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			h.respondJSON(w, http.StatusMethodNotAllowed, UploadResponse{
				Success: false,
				Message: "Method not allowed",
			})
			return
		}

		// Parse multipart form with max memory
		if err := r.ParseMultipartForm(MaxFileSize); err != nil {
			h.respondJSON(w, http.StatusBadRequest, UploadResponse{
				Success: false,
				Message: "File too large or invalid form data",
			})
			return
		}

		// Get the file from the form
		file, header, err := r.FormFile("file")
		if err != nil {
			h.respondJSON(w, http.StatusBadRequest, UploadResponse{
				Success: false,
				Message: "No file uploaded",
			})
			return
		}
		defer file.Close()

		// Validate file size
		if header.Size > MaxFileSize {
			h.respondJSON(w, http.StatusBadRequest, UploadResponse{
				Success: false,
				Message: fmt.Sprintf("File size exceeds maximum of %d bytes", MaxFileSize),
			})
			return
		}

		// Validate file type (optional - check extension)
		allowedExtensions := []string{".jpg", ".jpeg", ".png", ".gif", ".pdf", ".txt", ".doc", ".docx"}
		ext := strings.ToLower(filepath.Ext(header.Filename))
		if !contains(allowedExtensions, ext) {
			h.respondJSON(w, http.StatusBadRequest, UploadResponse{
				Success: false,
				Message: fmt.Sprintf("File type %s not allowed", ext),
			})
			return
		}

		// Create destination file
		destPath := filepath.Join(h.uploadDir, header.Filename)
		destFile, err := os.Create(destPath)
		if err != nil {
			h.respondJSON(w, http.StatusInternalServerError, UploadResponse{
				Success: false,
				Message: "Failed to save file",
			})
			return
		}
		defer destFile.Close()

		// Copy file content
		size, err := io.Copy(destFile, file)
		if err != nil {
			h.respondJSON(w, http.StatusInternalServerError, UploadResponse{
				Success: false,
				Message: "Failed to save file",
			})
			return
		}

		h.respondJSON(w, http.StatusOK, UploadResponse{
			Success:  true,
			Message:  "File uploaded successfully",
			Filename: header.Filename,
			Size:     size,
		})
	}
}

func (h *FileUploadHandler) respondJSON(w http.ResponseWriter, status int, data UploadResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func (h *FileUploadHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mux.ServeHTTP(w, r)
}

func main() {}
