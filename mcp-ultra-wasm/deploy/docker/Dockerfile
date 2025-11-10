# Build stage
FROM golang:1.22-alpine AS builder

RUN apk add --no-cache git make ca-certificates tzdata

WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s -X 'github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/pkg/version.Version=v1.0.0' -X 'github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/pkg/version.GitCommit=$(git rev-parse HEAD)' -X 'github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/pkg/version.BuildDate=$(date -u +%Y-%m-%dT%H:%M:%SZ)'" \
    -o mcp-ultra-wasm cmd/mcp-model-ultra/main.go

# Runtime stage
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata wget && \
    adduser -D -g '' appuser

WORKDIR /app

COPY --from=builder /build/mcp-ultra-wasm .
COPY --from=builder /build/config ./config

# Create non-root user
USER appuser

EXPOSE 9655

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --quiet --tries=1 --spider http://localhost:9655/healthz || exit 1

ENTRYPOINT ["./mcp-ultra-wasm"]
