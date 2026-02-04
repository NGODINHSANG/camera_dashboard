# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install build dependencies for CGO (SQLite)
RUN apk add --no-cache gcc musl-dev sqlite-dev

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build with CGO enabled for SQLite
RUN CGO_ENABLED=1 GOOS=linux go build -o server ./cmd/server

# Runtime stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata sqlite ffmpeg

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/server .

# Create directories for data and recordings
RUN mkdir -p /app/data /app/recordings

# Expose port
EXPOSE 8080

# Run the application
CMD ["./server"]
