# Stage 1: Builder
FROM golang:1.25.0-alpine AS builder

# Install git (may be needed for go mod download)
RUN apk add --no-cache git

# Set working directory
WORKDIR /app

# Copy go.mod and go.sum
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN go build -o bin/kinetria cmd/kinetria/api/main.go

# Stage 2: Runtime
FROM alpine:latest

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/bin/kinetria .

# Expose port
EXPOSE 8080

# Run the application
CMD ["./kinetria"]
