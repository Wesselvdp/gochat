# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Copy go mod files first for better layer caching
COPY go.* ./
RUN go mod download

# Copy the entire project
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/server ./cmd/main.go

# Final stage
FROM alpine:latest

# Add CA certificates for HTTPS
RUN apk --no-cache add ca-certificates

# Copy binary from builder
COPY --from=builder /app/server /server

# Expose port
EXPOSE 8080

# Run the binary
CMD ["/server"]