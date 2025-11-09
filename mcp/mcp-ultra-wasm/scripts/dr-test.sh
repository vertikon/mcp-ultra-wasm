#!/usr/bin/env bash
set -euo pipefail
echo "🚨 DR drill start"

echo "💥 Scaling down primary deployment to simulate outage"
kubectl -n production scale deploy mcp-model-ultra --replicas=0

sleep 10

echo "🔄 Scaling up to recover"
kubectl -n production scale deploy mcp-model-ultra --replicas=2

echo "🔍 Checking health"
kubectl -n production rollout status deploy/mcp-model-ultra

echo "✅ DR drill finished"
