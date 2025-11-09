#!/usr/bin/env bash
set -euo pipefail

echo "[mcp-ultra-wasm] Regenerando mocks com GoMock..."

# Exemplo — ajuste os caminhos conforme seu repo
mockgen -source=internal/services/task_repository.go -destination=internal/mocks/task_repository_mock.go -package=mocks
mockgen -source=internal/compliance/framework.go -destination=internal/mocks/compliance_mock.go -package=mocks || true

echo "OK"