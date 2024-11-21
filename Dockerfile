# syntax=docker/dockerfile:1

# Build stage
FROM golang:1.22-alpine AS builder

# Install build tools
RUN apk add --no-cache git make build-base

# Set working directory
WORKDIR /src

# Copy go mod files first
COPY go.mod go.sum ./

# Download dependencies with retry and verbose logging
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download -x && \
    go mod verify

# Copy source code
COPY . .

# Build with optimizations
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" \
    -o /app/backuprds

# Runtime stage
FROM alpine:3.19

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata && \
    cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    echo "Asia/Shanghai" > /etc/timezone

# Create non-root user
RUN adduser -D -u 1000 appuser

# Create app directory
WORKDIR /app

# Copy binary and config
COPY --from=builder /app/backuprds .
COPY --from=builder /src/static ./static
COPY --from=builder /src/config/config.yaml ./config/

# Set permissions
RUN chown -R appuser:appuser /app

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget -q --spider http://localhost:8080/health || exit 1

# Start application
ENTRYPOINT ["/app/backuprds"]

