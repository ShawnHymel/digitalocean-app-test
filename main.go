package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/google/uuid"
)

//------------------------------------------------------------------------------
// Structs and Types

// Configuration struct to hold all configurable parameters
type Config struct {
	Port                string
	MaxFileSize         int64         // MB
	GradingTimeout      time.Duration // minutes
	CleanupInterval     time.Duration // hours
	CompletedJobTTL     time.Duration // hours
	FailedJobTTL        time.Duration // hours
	OldFileTTL          time.Duration // hours
	QueueBufferSize     int
	UploadsDir          string
	GradingScriptPath   string
}

// Job represents a grading job
type Job struct {
	ID       string    `json:"id"`
	Filename string    `json:"filename"`
	FilePath string    `json:"-"` // Don't expose file path in JSON
	Size     int64     `json:"size"`
	Status   string    `json:"status"` // "queued", "processing", "completed", "failed"
	Result   *JobResult `json:"result,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// JobResult represents the grading result
type JobResult struct {
	Score    float64 `json:"score"`
	Feedback string  `json:"feedback"`
	Error    string  `json:"error,omitempty"`
}

// Queue responses
type SubmitResponse struct {
	JobID   string `json:"job_id"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

type StatusResponse struct {
	Job *Job `json:"job"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

// Simple in-memory queue
type JobQueue struct {
	jobs      map[string]*Job
	queue     chan string
	mutex     sync.RWMutex
	isRunning bool
}

// Global variables
var (
	config   *Config
	jobQueue *JobQueue
)

//------------------------------------------------------------------------------
// Functions

//  Load server configuration settings
func loadConfig() *Config {

	// Load configuration from environment variables with defaults
	config := &Config{
		Port:                getEnv("PORT", "8080"),
		MaxFileSize:         getEnvInt64("MAX_FILE_SIZE_MB", 50),
		GradingTimeout:      time.Duration(getEnvInt("GRADING_TIMEOUT_MIN", 5)) * time.Minute,
		CleanupInterval:     time.Duration(getEnvInt("CLEANUP_INTERVAL_HOURS", 1)) * time.Hour,
		CompletedJobTTL:     time.Duration(getEnvInt("COMPLETED_JOB_TTL_HOURS", 24)) * time.Hour,
		FailedJobTTL:        time.Duration(getEnvInt("FAILED_JOB_TTL_HOURS", 24)) * time.Hour,
		OldFileTTL:          time.Duration(getEnvInt("OLD_FILE_TTL_HOURS", 48)) * time.Hour,
		QueueBufferSize:     getEnvInt("QUEUE_BUFFER_SIZE", 100),
		UploadsDir:          getEnv("UPLOADS_DIR", "/tmp/uploads"),
		GradingScriptPath:   getEnv("GRADING_SCRIPT_PATH", "/usr/local/bin/grade.py"),
	}
	
	// Convert MB to bytes for file size
	config.MaxFileSize = config.MaxFileSize * 1024 * 1024
	
	return config
}

// Helper functions to get environment variables with defaults
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// Helper functions to get environment variables as integers
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// Helper functions to get environment variables as int64
func getEnvInt64(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.ParseInt(value, 10, 64); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// Accept file uploads and queues them for processing
func submitHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Only POST method allowed"})
		return
	}

	// Parse the multipart form with configured max file size
	err := r.ParseMultipartForm(config.MaxFileSize)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Unable to parse form - file may be too large"})
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

	// Check file size against configured limit
	if header.Size > config.MaxFileSize {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error: fmt.Sprintf("File too large. Maximum size: %d MB", config.MaxFileSize/(1024*1024)),
		})
		return
	}

	// Create job ID and file path
	jobID := generateJobID()
	
	// Create uploads directory if it doesn't exist
	os.MkdirAll(config.UploadsDir, 0755)
	
	// Save file to disk
	filePath := fmt.Sprintf("%s/%s_%s", config.UploadsDir, jobID, header.Filename)
	
	// Read and save file
	fileContents, err := io.ReadAll(file)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Unable to read file"})
		return
	}
	
	err = os.WriteFile(filePath, fileContents, 0644)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Unable to save file"})
		return
	}

	// Create job (no file contents in RAM)
	job := &Job{
		ID:        jobID,
		Filename:  header.Filename,
		Size:      header.Size,
		Status:    "queued",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		FilePath:  filePath, // Store path for processing
	}

	// Add to queue
	jobQueue.addJob(job)

	fmt.Printf("📁 File saved: %s (Job: %s)\n", filePath, jobID)

	// Return job ID immediately
	response := SubmitResponse{
		JobID:   jobID,
		Status:  "queued",
		Message: "File submitted for grading. Use job_id to check status.",
	}

	json.NewEncoder(w).Encode(response)
}

// Return the status of a specific job
func statusHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Extract job ID from URL path
	jobID := r.URL.Path[len("/status/"):]
	if jobID == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Job ID required"})
		return
	}

	job := jobQueue.getJob(jobID)
	if job == nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Job not found"})
		return
	}

	response := StatusResponse{Job: job}
	json.NewEncoder(w).Encode(response)
}

// Return overall queue information
func queueStatusHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	queueLength := len(jobQueue.queue)
	totalJobs := len(jobQueue.jobs)
	
	response := map[string]interface{}{
		"queue_length": queueLength,
		"total_jobs":   totalJobs,
		"worker_running": jobQueue.isRunning,
	}
	
	json.NewEncoder(w).Encode(response)
}

// Perform a health check
func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// cReturn current configuration (for debugging/monitoring)
func configHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	configInfo := map[string]interface{}{
		"max_file_size_mb":        config.MaxFileSize / (1024 * 1024),
		"grading_timeout_minutes": int(config.GradingTimeout.Minutes()),
		"cleanup_interval_hours":  int(config.CleanupInterval.Hours()),
		"completed_job_ttl_hours": int(config.CompletedJobTTL.Hours()),
		"failed_job_ttl_hours":    int(config.FailedJobTTL.Hours()),
		"old_file_ttl_hours":      int(config.OldFileTTL.Hours()),
		"queue_buffer_size":       config.QueueBufferSize,
		"uploads_dir":             config.UploadsDir,
		"grading_script_path":     config.GradingScriptPath,
	}
	
	json.NewEncoder(w).Encode(configInfo)
}

// Add a job to the queue and map it to its job ID
func (q *JobQueue) addJob(job *Job) {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	
	q.jobs[job.ID] = job
	q.queue <- job.ID
	
	fmt.Printf("Job %s queued (%s)\n", job.ID, job.Filename)
}

// Get a job by ID from the queue
func (q *JobQueue) getJob(jobID string) *Job {
	q.mutex.RLock()
	defer q.mutex.RUnlock()
	
	return q.jobs[jobID]
}

// Update job status and result in the queue
func (q *JobQueue) updateJob(jobID string, status string, result *JobResult) {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	
	if job, exists := q.jobs[jobID]; exists {
		job.Status = status
		job.Result = result
		job.UpdatedAt = time.Now()
	}
}

// Worker that processes jobs one by one
func (q *JobQueue) startWorker() {
	q.isRunning = true
	fmt.Println("🔄 Worker started - processing jobs...")
	
	for jobID := range q.queue {
		fmt.Printf("⚡ Processing job %s...\n", jobID)
		
		// Update status to processing
		q.updateJob(jobID, "processing", nil)
		
		// Simulate grading work (replace with actual grading logic)
		result := q.processJob(jobID)
		
		// Update with result and cleanup file if failed
		if result.Error != "" {
			q.updateJob(jobID, "failed", result)
			fmt.Printf("❌ Job %s failed: %s\n", jobID, result.Error)
			
			// Clean up file for failed jobs
			job := q.getJob(jobID)
			if job != nil {
				q.cleanupFile(job.FilePath, jobID, "job failed")
			}
		} else {
			q.updateJob(jobID, "completed", result)
			fmt.Printf("✅ Job %s completed (Score: %.1f)\n", jobID, result.Score)
			// File cleanup happens in processJob for successful jobs
		}
	}
}

// Calls the Python script for grading
func (q *JobQueue) processJob(jobID string) *JobResult {
	job := q.getJob(jobID)
	if job == nil {
		return &JobResult{Error: "Job not found"}
	}
	
	// Create isolated grading directory
	tempDir := fmt.Sprintf("/tmp/grading_%s", jobID)
	err := os.MkdirAll(tempDir, 0755)
	if err != nil {
		return &JobResult{Error: "Failed to create grading directory"}
	}
	defer os.RemoveAll(tempDir) // Always cleanup

	fmt.Printf("🔬 Starting grading in %s\n", tempDir)

	// Copy student submission to grading directory
	submissionPath := filepath.Join(tempDir, "submission.zip")
	err = q.copyFile(job.FilePath, submissionPath)
	if err != nil {
		return &JobResult{Error: "Failed to copy submission for grading"}
	}

	// Run Python grading script with timeout
	return q.runPythonGrader(tempDir, submissionPath, job.Filename)
}

// Copy a file from src to dst
func (q *JobQueue) copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}

// Execute the Python grading script
func (q *JobQueue) runPythonGrader(tempDir, submissionPath, originalFilename string) *JobResult {
	// Set up context with configured timeout
	ctx, cancel := context.WithTimeout(context.Background(), config.GradingTimeout)
	defer cancel()

	// Create the grading command using configured script path
	cmd := exec.CommandContext(ctx, "python3", config.GradingScriptPath, 
		submissionPath, originalFilename, tempDir)
	
	// Set working directory
	cmd.Dir = tempDir
	
	// Set up process isolation
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true, // New process group for clean termination
	}

	// Capture output
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	fmt.Printf("⚡ Running Python grader: python3 %s %s (timeout: %v)\n", 
		config.GradingScriptPath, originalFilename, config.GradingTimeout)
	
	// Execute grading
	err := cmd.Run()
	
	// Handle timeout
	if ctx.Err() == context.DeadlineExceeded {
		// Kill the entire process group
		syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
		return &JobResult{
			Error: fmt.Sprintf("Grading timeout - submission took longer than %v to grade", config.GradingTimeout),
		}
	}

	// Handle other errors
	if err != nil {
		return &JobResult{
			Error: fmt.Sprintf("Grading script failed: %s\nStderr: %s", err.Error(), stderr.String()),
		}
	}

	// Parse grading output (expects JSON from Python script)
	return q.parseGradingOutput(stdout.String(), stderr.String())
}

// Convert Python script JSON output to JobResult
func (q *JobQueue) parseGradingOutput(stdout, stderr string) *JobResult {
	fmt.Printf("🔍 Debug - Python stdout: %s\n", stdout)
	fmt.Printf("🔍 Debug - Python stderr: %s\n", stderr)
	
	// Try to parse JSON output from Python script
	var gradingResult struct {
		Score       float64 `json:"score"`
		MaxScore    float64 `json:"max_score"`
		Feedback    string  `json:"feedback"`
		TestResults []struct {
			Name    string  `json:"name"`
			Passed  bool    `json:"passed"`
			Message string  `json:"message"`
			Points  float64 `json:"points"`
		} `json:"test_results"`
		Error string `json:"error"`
	}
	
	// Clean up stdout and try to find JSON
	stdout = strings.TrimSpace(stdout)
	
	// Try to parse the entire stdout as JSON first
	err := json.Unmarshal([]byte(stdout), &gradingResult)
	if err == nil {
		fmt.Printf("✅ Successfully parsed JSON directly\n")
		if gradingResult.Error != "" {
			return &JobResult{Error: gradingResult.Error}
		}
		
		return &JobResult{
			Score:    gradingResult.Score,
			Feedback: gradingResult.Feedback,
		}
	}
	
	fmt.Printf("❌ Failed to parse stdout as JSON directly: %v\n", err)
	
	// If that fails, look for JSON line by line
	lines := strings.Split(stdout, "\n")
	for i, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "{") && strings.HasSuffix(line, "}") {
			fmt.Printf("🔍 Trying to parse line %d: %s\n", i, line)
			err := json.Unmarshal([]byte(line), &gradingResult)
			if err == nil {
				fmt.Printf("✅ Successfully parsed JSON from line %d\n", i)
				if gradingResult.Error != "" {
					return &JobResult{Error: gradingResult.Error}
				}
				
				return &JobResult{
					Score:    gradingResult.Score,
					Feedback: gradingResult.Feedback,
				}
			} else {
				fmt.Printf("❌ Failed to parse line %d: %v\n", i, err)
			}
		}
	}
	
	// Fallback: if no valid JSON found, return error with output
	return &JobResult{
		Error: fmt.Sprintf("Python grader did not return valid JSON.\nStdout: %s\nStderr: %s", stdout, stderr),
	}
}

// Remove a file and log the action
func (q *JobQueue) cleanupFile(filePath, jobID, reason string) {
	if filePath == "" {
		return
	}
	
	err := os.Remove(filePath)
	if err != nil {
		fmt.Printf("⚠️  Failed to cleanup file %s (Job: %s): %v\n", filePath, jobID, err)
	} else {
		fmt.Printf("🗑️  Cleaned up file %s (Job: %s) - %s\n", filePath, jobID, reason)
	}
}

// Run periodic cleanup of old files and jobs
func (q *JobQueue) startCleanup() {
	fmt.Printf("🧹 Cleanup service started - checking every %v...\n", config.CleanupInterval)
	
	ticker := time.NewTicker(config.CleanupInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			q.performCleanup()
		}
	}
}

// Remove old files and stale job records
func (q *JobQueue) performCleanup() {
	fmt.Println("🧹 Starting cleanup routine...")
	
	q.mutex.Lock()
	defer q.mutex.Unlock()
	
	now := time.Now()
	cleanedFiles := 0
	cleanedJobs := 0
	
	// Check all jobs for cleanup candidates
	for jobID, job := range q.jobs {
		shouldCleanup := false
		reason := ""
		
		// Cleanup criteria using configured TTLs
		if job.CreatedAt.Before(now.Add(-config.OldFileTTL)) {
			shouldCleanup = true
			reason = fmt.Sprintf("older than %v", config.OldFileTTL)
		} else if job.Status == "failed" && job.UpdatedAt.Before(now.Add(-config.FailedJobTTL)) {
			shouldCleanup = true
			reason = fmt.Sprintf("failed job older than %v", config.FailedJobTTL)
		} else if job.Status == "completed" && job.UpdatedAt.Before(now.Add(-config.CompletedJobTTL)) {
			shouldCleanup = true
			reason = fmt.Sprintf("completed job older than %v", config.CompletedJobTTL)
		}
		
		if shouldCleanup {
			// Remove file if it exists
			if job.FilePath != "" {
				err := os.Remove(job.FilePath)
				if err == nil {
					cleanedFiles++
					fmt.Printf("🗑️  Cleaned up old file: %s (Job: %s) - %s\n", job.FilePath, jobID, reason)
				}
			}
			
			// Remove job from memory
			delete(q.jobs, jobID)
			cleanedJobs++
		}
	}
	
	fmt.Printf("🧹 Cleanup complete: %d files removed, %d jobs removed\n", cleanedFiles, cleanedJobs)
}

// Generate UUID7-based job ID
func generateJobID() string {

	// Try UUID7 first (best option)
	if u, err := uuid.NewV7(); err == nil {

		// Encode as Base64 for shorter representation (22 chars vs 36)
		return base64.RawURLEncoding.EncodeToString(u[:])
	}
	
	// Fallback to UUID4 if UUID7 fails
	u := uuid.New()

	return base64.RawURLEncoding.EncodeToString(u[:])
}

//------------------------------------------------------------------------------
// Main entry point

// Initializes the server, loads configuration, and starts the API
func main() {

	// Load configuration (from environment variables or defaults)
	config = loadConfig()
	
	// Print configuration on startup
	fmt.Printf("⚙️ ByteGrader API starting with configuration:\n")
	fmt.Printf("   Port: %s\n", config.Port)
	fmt.Printf("   Max file size: %d MB\n", config.MaxFileSize/(1024*1024))
	fmt.Printf("   Grading timeout: %v\n", config.GradingTimeout)
	fmt.Printf("   Cleanup interval: %v\n", config.CleanupInterval)
	fmt.Printf("   Completed job TTL: %v\n", config.CompletedJobTTL)
	fmt.Printf("   Failed job TTL: %v\n", config.FailedJobTTL)
	fmt.Printf("   Old file TTL: %v\n", config.OldFileTTL)
	fmt.Printf("   Queue buffer size: %d\n", config.QueueBufferSize)
	fmt.Printf("   Uploads directory: %s\n", config.UploadsDir)
	fmt.Printf("   Grading script: %s\n", config.GradingScriptPath)
	fmt.Println()

	// Initialize queue with configured buffer size
	jobQueue = &JobQueue{
		jobs:  make(map[string]*Job),
		queue: make(chan string, config.QueueBufferSize),
	}

	// Start the worker and cleanup routines
	go jobQueue.startWorker()
	go jobQueue.startCleanup()

	// Create uploads directory
	os.MkdirAll(config.UploadsDir, 0755)

	// API endpoints
	http.HandleFunc("/submit", submitHandler)
	http.HandleFunc("/status/", statusHandler)
	http.HandleFunc("/queue", queueStatusHandler)
	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/config", configHandler)

	// Print server startup information
	fmt.Printf("🌐 Starting Bytegrader API on port %s...\n", config.Port)
	fmt.Println("Endpoints:")
	fmt.Println("  POST /submit - Submit file for grading (returns job_id)")
	fmt.Println("  GET  /status/{job_id} - Check job status")
	fmt.Println("  GET  /queue - View queue status")
	fmt.Println("  GET  /config - View current configuration")
	fmt.Println("  GET  /health - Health check")
	
	// Start the server
	log.Fatal(http.ListenAndServe(":"+config.Port, nil))
}