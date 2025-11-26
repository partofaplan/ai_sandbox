# Stage 1: Build the Go application
FROM golang:1.24-alpine AS builder

# Install necessary tools
RUN apk --no-cache add git

# Set the working directory
WORKDIR /app

# Copy go.mod and go.sum first (for dependency caching)
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the rest of the application
COPY . .

# Build binaries
RUN go build -o zachbot . && go build -o penpal ./cmd/penpal

# Stage 2: Create a minimal runtime image
FROM alpine:latest

# Install certificates for HTTPS
RUN apk --no-cache add ca-certificates

# Set working directory
WORKDIR /root/

# Copy the compiled binary from the builder stage
COPY --from=builder /app/zachbot .
COPY --from=builder /app/penpal .

# Copy templates and static files
COPY --from=builder /app/templates ./templates
COPY --from=builder /app/static ./static

# Expose the application port
EXPOSE 6600

# Command to run the executable
CMD ["./zachbot"]
