# Build stage
FROM golang:1.24.1-alpine AS builder

# Set working directory
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies
RUN go mod download

# Copy the source code
COPY *.go ./

# Build the application with optimizations
# Changed from building all .go files to specifically building the main package file
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-w -s" -o jokeapi jokeapi.go

# Final stage
FROM alpine:latest  

# Add CA certificates for HTTPS
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/jokeapi .

# Expose the application port
EXPOSE 8080

# Run the binary
CMD ["./jokeapi"]