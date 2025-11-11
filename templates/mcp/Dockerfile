# MCP Ultra Multi-stage Dockerfile
# Build stage
FROM golang:1.22-alpine AS builder

RUN apk add --no-cache git make ca-certificates tzdata

WORKDIR /build

# Copy dependency files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s -X 'github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/pkg/version.Version=v1.0.0' -X 'github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/pkg/version.GitCommit=$(git rev-parse HEAD 2>/dev/null || echo 'unknown')' -X 'github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/pkg/version.BuildDate=$(date -u +%Y-%m-%dT%H:%M:%SZ)'" \
    -o mcp-ultra-wasm cmd/mcp-model-ultra/main.go

# Runtime stage
FROM alpine:3.19

# Install required packages and create user
RUN apk add --no-cache ca-certificates tzdata wget curl && \
    adduser -D -g '' appuser

WORKDIR /app

# Copy binary and configuration
COPY --from=builder /build/mcp-ultra-wasm .
COPY --from=builder /build/config ./config
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Set ownership
RUN chown -R appuser:appuser /app

# Switch to non-root user
USER appuser

# Expose ports
EXPOSE 9655 9656

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --quiet --tries=1 --spider http://localhost:9655/healthz || exit 1

# Set entrypoint
ENTRYPOINT ["./mcp-ultra-wasm"]