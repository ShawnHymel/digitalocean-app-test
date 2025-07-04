# This is a multi-stage Dockerfile for building a Go addition API app.
# Stage 1: Build the Go binary
# Stage 2: Create a minimal runtime image

# --- Stage 1: Build the Go binary ---

# Use a lightweight Go image for building the application
FROM golang:1.24-alpine AS builder
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
        libmagic1 \
    && rm -rf /var/lib/apt/lists/*

# Install Python dependencies
RUN python3 -m pip install --no-cache-dir \
        python-magic

# Copy the Go binary from builder stage
WORKDIR /app
COPY --from=builder /app/bytegrader-api .

# Copy the grading scripts
COPY graders/ /usr/local/bin/graders

# Create uploads directory
RUN mkdir -p /tmp/uploads

EXPOSE 8080
CMD ["./bytegrader-api"]
