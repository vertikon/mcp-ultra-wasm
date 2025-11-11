#!/usr/bin/env bash
set -euo pipefail
VERSION=${1:-"latest"}
echo "🚀 Deploying version $VERSION"
kubectl -n production set image deploy/mcp-model-ultra app=example/mcp-model-ultra:$VERSION
kubectl -n production rollout status deploy/mcp-model-ultra
