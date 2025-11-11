#!/usr/bin/env bash
set -euo pipefail

echo "════════════════════════════════════════════════════════════"
echo "  🚀 CI Pipeline - mcp-ultra-wasm (Linux/macOS)"
echo "════════════════════════════════════════════════════════════"
echo ""

# 1) Garantir saúde do módulo (conserta go.sum e previne erros de export data)
echo "📦 [1/4] Executando go mod tidy..."
go mod tidy

echo "✓ [2/4] Verificando go.sum..."
go mod verify

# 2) Lint padrão com gomodguard (nova configuração)
echo ""
echo "🔍 [3/4] Executando golangci-lint (gomodguard)..."
golangci-lint run --config=.golangci-new.yml --timeout=5m

# 3) Compilar e executar vettool nativo (depguard-lite)
echo ""
echo "🔨 [4/4] Compilando e executando depguard-lite..."
mkdir -p vettools
go build -o vettools/depguard-lite ./cmd/depguard-lite
go vet -vettool=$(pwd)/vettools/depguard-lite ./...

echo ""
echo "════════════════════════════════════════════════════════════"
echo "  ✅ CI PASSED - Todas as verificações OK"
echo "════════════════════════════════════════════════════════════"
