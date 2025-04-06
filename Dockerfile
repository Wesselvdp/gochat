# Frontend build stage
FROM node:20-alpine AS frontend-builder
WORKDIR /app
COPY views /views
WORKDIR /app

COPY frontend/package*.json ./
COPY frontend/ ./
RUN npm install
RUN npm run build

# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Add build dependencies for SQLite
RUN apk add --no-cache gcc musl-dev

RUN go install github.com/a-h/templ/cmd/templ@latest

# Install golang-migrate
RUN go install -tags 'sqlite3' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Copy go mod files first for better layer caching
COPY go.* ./
RUN go mod download

# Copy the entire project
COPY . .
COPY  --from=frontend-builder /app/dist /frontend/dist
# Build the application
RUN templ generate
# Note: CGO_ENABLED=1 for SQLite support
RUN CGO_ENABLED=1 GOOS=linux go build -o /app/server ./cmd/main.go

# Final stage
FROM alpine:latest

# Add runtime dependencies for SQLite and CA certificates
RUN apk add ca-certificates sqlite

# Copy binary from builder
COPY --from=builder /app/server /server
RUN chmod +x /server  # Add this line to make the server executable
# Copy the frontend dist files (only once)
COPY --from=frontend-builder /app/dist /frontend/dist
# Copy migrations
COPY db/migrations /db/migrations

COPY --from=builder /go/bin/migrate /usr/local/bin/migrate

# Create directory for SQLite database
RUN mkdir -p /data
# Ensure the app has write permissions
RUN chmod 755 /data

# Set environment variable for database path
ENV DB_PATH=/data/database.db

# Expose port
EXPOSE 8080


# Add an entrypoint script to run migrations before starting the server
COPY scripts/entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
