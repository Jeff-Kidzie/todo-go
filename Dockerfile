# syntax=docker/dockerfile:1

# Stage 1: Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
# CGO_ENABLED=0 for static binary (optional but recommended)
RUN CGO_ENABLED=0 go build -o main .

# Stage 2: Runtime stage (minimal image)
FROM alpine:latest

# Install CA certificates if your app makes HTTPS requests
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy only the compiled binary from builder stage
COPY --from=builder /app/main .

# Expose your app port (adjust 8080 to your port)
EXPOSE 8080

# Run the application
CMD ["./main"]
