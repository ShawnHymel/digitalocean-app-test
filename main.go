package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
)

// AddResponse represents what we send back
type AddResponse struct {
	A      float64 `json:"a"`
	B      float64 `json:"b"`
	Result float64 `json:"result"`
}

func main() {
	// Get port from environment or use 8080
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Set up our single endpoint
	http.HandleFunc("/add", addHandler)

	fmt.Printf("Server running on port %s\n", port)
	fmt.Println("Try: curl 'http://localhost:8080/add?a=3.5&b=2.1'")
	
	// Start the server
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

// addHandler does the actual addition
func addHandler(w http.ResponseWriter, r *http.Request) {
	// Get the numbers from URL parameters
	aStr := r.URL.Query().Get("a")
	bStr := r.URL.Query().Get("b")

	// Check if both parameters exist
	if aStr == "" || bStr == "" {
		http.Error(w, "Need both 'a' and 'b' parameters", http.StatusBadRequest)
		return
	}

	// Convert strings to numbers
	a, err := strconv.ParseFloat(aStr, 64)
	if err != nil {
		http.Error(w, "Parameter 'a' must be a number", http.StatusBadRequest)
		return
	}

	b, err := strconv.ParseFloat(bStr, 64)
	if err != nil {
		http.Error(w, "Parameter 'b' must be a number", http.StatusBadRequest)
		return
	}

	// Do the math!
	result := a + b

	// Create response
	response := AddResponse{
		A:      a,
		B:      b,
		Result: result,
	}

	// Send JSON response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
