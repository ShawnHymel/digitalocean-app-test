# This is a multi-stage Dockerfile for building a Go addition API app.
# Stage 1: Build the Go binary
# Stage 2: Create a minimal runtime image

# --- Stage 1: Build the Go binary ---

# Use a lightweight Go image for building the application
FROM golang:1.21-alpine AS builder
WORKDIR /app

# Copy Go module files first (for better caching)
COPY go.mod go.sum ./
RUN go mod download

# Copy only the Go source file
COPY main.go ./
RUN go build -o bytegrader-api .

# --- Stage 2: Create a minimal runtime image ---

# Production image with Python for grading
FROM python:3.12-slim
RUN apt-get update && apt-get install -y \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# Copy the Go binary from builder stage
WORKDIR /app
COPY --from=builder /app/bytegrader-api .

# Copy only the Python grading script
COPY grade.py /usr/local/bin/grade.py
RUN chmod +x /usr/local/bin/grade.py

# Create uploads directory
RUN mkdir -p /tmp/uploads

EXPOSE 8080
CMD ["./bytegrader-api"]
