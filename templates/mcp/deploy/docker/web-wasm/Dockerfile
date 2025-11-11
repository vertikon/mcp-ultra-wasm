# Multi-stage build para Web WASM Server
FROM --platform=linux/amd64 golang:1.24-alpine AS builder

# Instalar dependências necessárias
RUN apk add --no-cache git ca-certificates tzdata

# Configurar timezone
ENV TZ=UTC

WORKDIR /app

# Copiar go mod files
COPY go.mod go.sum ./

# Baixar dependências
RUN go mod download

# Copiar código fonte
COPY . .

# Build do servidor web
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o wasm-server \
    ./cmd/wasm-server

# Build do módulo WASM
RUN mkdir -p /tmp/wasm-build
WORKDIR /tmp/wasm-build

# Copiar código WASM
COPY --from=builder /app/wasm/wasm/ ./

# Build do módulo WASM
RUN GOOS=js GOARCH=wasm go build \
    -ldflags='-s -w' \
    -o main.wasm \
    .

# Copiar wasm_exec.js do Go
RUN cp "$(go env GOROOT)/misc/wasm/wasm_exec.js" ./

# Imagem final
FROM alpine:latest

# Instalar dependências de runtime
RUN apk --no-cache add ca-certificates tzdata curl

# Criar usuário não-root
RUN addgroup -g 1001 -S webwasm && \
    adduser -u 1001 -S webwasm -G webwasm

WORKDIR /app

# Copiar binário do servidor
COPY --from=builder /app/wasm-server .

# Copiar arquivos estáticos
COPY --from=builder /app/wasm/static ./static
COPY --from=builder /app/wasm/templates ./templates

# Copiar arquivos WASM
RUN mkdir -p ./wasm
COPY --from=builder /tmp/wasm-build/main.wasm ./wasm/
COPY --from=builder /tmp/wasm-build/wasm_exec.js ./wasm/

# Copiar arquivos de configuração
COPY --from=builder /app/api ./api
COPY --from=builder /app/deploy/docker/wasm/config.yaml ./config.yaml

# Criar diretórios necessários
RUN mkdir -p logs data

# Configurar permissões
RUN chown -R webwasm:webwasm /app
USER webwasm

# Expor porta
EXPOSE 8080

# Variáveis de ambiente
ENV GIN_MODE=release
ENV LOG_LEVEL=info
ENV PORT=8080
ENV STATIC_PATH=/app/static
ENV WASM_PATH=/app/wasm

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1

# Comando de execução
CMD ["./wasm-server"]


