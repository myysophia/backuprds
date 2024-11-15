# syntax=docker/dockerfile:1

# Step 1: build golang binary
FROM golang:1.21-alpine AS builder

# Step 2: install base tools
RUN apk add --no-cache git make

# Step 3: set working directory
WORKDIR /app

# Step 4: copy dependency files
COPY go.mod go.sum ./

# Step 5: download dependencies
RUN go mod download

# Step 6: copy source code
COPY . .

# Step 7: build application
RUN CGO_ENABLED=0 GOOS=linux go build -o rdsbackup

# Step 8: run stage
FROM alpine:3.19

# Step 9: install base tools and certificates
RUN apk add --no-cache ca-certificates tzdata \
    && cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
    && echo "Asia/Shanghai" > /etc/timezone

# Step 10: create non-root user
RUN adduser -D -u 1000 appuser

# Step 11: create necessary directories
WORKDIR /app
RUN mkdir -p /app/config && chown -R appuser:appuser /app

# Step 12: copy binary from step1
COPY --from=builder /app/rdsbackup /app/
COPY --from=builder /app/config/config.yaml /app/config/

# Step 13: switch to non-root user
USER appuser

# Step 14: expose port
EXPOSE 8080

# Step 15: health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget -q --spider http://localhost:8080/health || exit 1

# Step 16: start application
ENTRYPOINT ["/app/rdsbackup"]

