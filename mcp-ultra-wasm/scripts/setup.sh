#!/usr/bin/env bash
set -euo pipefail
echo "🔧 Setting up MCP Model Ultra environment"

kubectl create namespace production --dry-run=client -o yaml | kubectl apply -f -
kubectl create namespace monitoring --dry-run=client -o yaml | kubectl apply -f -

echo "📦 Applying K8s manifests"
kubectl -n production apply -f deploy/k8s/deployment.yaml
kubectl -n production apply -f deploy/flagger/canary.yaml
kubectl -n monitoring apply -f deploy/monitoring/servicemonitor.yaml
kubectl -n monitoring apply -f deploy/monitoring/alerts.yaml

echo "✅ Setup done"
