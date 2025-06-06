# This is a multi-stage Dockerfile for building a Go addition API app.
# Stage 1: Build the Go binary
# Stage 2: Create a minimal runtime image

# -- Stage 1: Build Stage -- #
FROM golang:1.21-alpine AS builder

# Install git (needed for go modules)
RUN apk --no-cache add git ca-certificates

# Set working directory
WORKDIR /app

# Copy go mod files first (for better caching)
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the binary
# -ldflags="-s -w" removes debug info to reduce binary size
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o bin/addition-api .

# -- Stage 2: Runtime Stage -- #
FROM alpine:latest

# Install ca-certificates for HTTPS requests (if needed)
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user for security
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Set working directory
WORKDIR /app

# Copy the binary from builder stage
COPY --from=builder /app/bin/addition-api .

# Change ownership to non-root user
RUN chown -R appuser:appgroup /app

# Switch to non-root user
USER appuser

# Expose port (DigitalOcean will set PORT environment variable)
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:${PORT:-8080}/health || exit 1

# Run the binary
CMD ["./addition-api"]
