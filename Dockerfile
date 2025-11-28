FROM golang:1.25-trixie AS builder

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -o backend .

# Final stage
FROM debian:trixie-slim

# Install ca-certificates for HTTPS
RUN apt update && apt install -y ca-certificates curl

# Create non-root user
RUN groupadd -g 1001 nishiki && \
    useradd -u 1001 -g nishiki -m nishiki

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/backend .

# Copy configuration files
COPY --from=builder /app/app.toml .

# Create certs directory for TLS certificates
RUN mkdir -p certs && chown -R nishiki:nishiki /app

# Switch to non-root user
USER nishiki

# Expose port
EXPOSE 3001

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:3001/health || exit 1

# Start the application
CMD ["./backend"]
