package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

// FileResponse represents what we send back for file uploads
type FileResponse struct {
	Filename string `json:"filename"`
	Contents string `json:"contents"`
	Size     int64  `json:"size"`
}

// ErrorResponse for error handling
type ErrorResponse struct {
	Error string `json:"error"`
}

func main() {
	// Get port from environment or use 8080
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// API endpoint
	http.HandleFunc("/upload", uploadHandler)
	
	// Simple status endpoint
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	fmt.Printf("File Upload API running on port %s\n", port)
	fmt.Println("Endpoints:")
	fmt.Println("  POST /upload - Upload file and get contents")
	fmt.Println("  GET  /health - Health check")
	
	// Start the server
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

// uploadHandler handles file uploads and returns contents
func uploadHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Only accept POST requests
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Only POST method allowed"})
		return
	}

	// Parse the multipart form (10MB max)
	err := r.ParseMultipartForm(10 << 20) // 10 MB limit
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Unable to parse form"})
		return
	}

	// Get the file from form data
	file, header, err := r.FormFile("file")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Unable to get file from form"})
		return
	}
	defer file.Close()

	// Read file contents
	fileContents, err := io.ReadAll(file)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Unable to read file"})
		return
	}

	// Create response with file info and contents
	response := FileResponse{
		Filename: header.Filename,
		Contents: string(fileContents),
		Size:     header.Size,
	}

	// Send JSON response
	json.NewEncoder(w).Encode(response)
}
