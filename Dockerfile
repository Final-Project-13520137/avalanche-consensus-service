FROM golang:1.23-alpine AS builder

WORKDIR /app

# Copy go.mod first for better caching
COPY go.mod ./
RUN go mod download

# Copy the source code
COPY src/ ./src/

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/avalanche-consensus-service ./src/cmd/main.go

# Create final minimal image
FROM alpine:latest

WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/bin/avalanche-consensus-service /app/
# Copy the config file
COPY src/config.json /app/

# Create a non-root user and set permissions
RUN adduser -D -h /app appuser && \
    chown -R appuser:appuser /app

USER appuser

# Expose the app port
EXPOSE 8080

# Run the application
CMD ["/app/avalanche-consensus-service", "--config", "config.json"] 