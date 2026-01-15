# Multi-stage build for CodexDB HTTP Service
# Stage 1: Build
FROM golang:1.24-alpine AS builder

WORKDIR /build

# Install build dependencies
RUN apk add --no-cache git ca-certificates

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build binary (static, no CGO)
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -a -installsuffix cgo \
    -ldflags="-w -s" \
    -o codex-service \
    ./cmd/codex-service

# Stage 2: Runtime
FROM alpine:3.19

# Add ca-certificates for HTTPS (if needed in future)
RUN apk --no-cache add ca-certificates

# Create non-root user
RUN addgroup -g 1000 codex && \
    adduser -D -u 1000 -G codex codex

# Create data directory with proper permissions
RUN mkdir -p /data && chown -R codex:codex /data

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder --chown=codex:codex /build/codex-service .

# Switch to non-root user
USER codex

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run service
ENTRYPOINT ["/app/codex-service"]
CMD []

# Labels for container metadata
LABEL maintainer="CodexDB Project"
LABEL description="CodexDB HTTP Service - File-based Key-Value Database"
LABEL version="1.0"
