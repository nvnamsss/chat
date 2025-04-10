FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Download dependencies first (leveraging Docker cache)
COPY src/go.mod src/go.sum* ./src/
RUN cd src && go mod download

# Copy source code
COPY src/ ./src/

# Build the application
RUN cd src && CGO_ENABLED=0 GOOS=linux go build -o /app/chat-service ./cmd/main

# Create a minimal production image
FROM alpine:3.18

WORKDIR /app

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Copy the binary from the builder stage
COPY --from=builder /app/chat-service /app/chat-service
# Copy migrations
COPY --from=builder /app/src/migrations /app/src/migrations

# Set environment variables
ENV GIN_MODE=release \
    APP_ENV=production

# Expose the service port
EXPOSE 8080

# Run the service
CMD ["/app/chat-service"]