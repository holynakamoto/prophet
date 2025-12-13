#!/usr/bin/env bash
# Quick setup script for Grafana and Prometheus for the demo
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$SCRIPT_DIR/../.."

echo "📊 Setting up Grafana and Prometheus for Prophet Demo..."
echo ""

# Create monitoring namespace
kubectl create namespace monitoring --dry-run=client -o yaml | kubectl apply -f -

# Deploy kube-state-metrics (required for kube_* pod metrics used by the dashboard)
echo "📦 Deploying kube-state-metrics..."
if [ -f "$REPO_ROOT/monitoring/kube-state-metrics/kube-state-metrics.yaml" ]; then
    kubectl apply -f "$REPO_ROOT/monitoring/kube-state-metrics/kube-state-metrics.yaml"
else
    echo "⚠️  kube-state-metrics manifest not found, dashboard pod panels may be empty"
fi

# Deploy Prometheus (simplified for demo)
echo "📈 Deploying Prometheus..."
if [ -f "$REPO_ROOT/monitoring/prometheus/prometheus.yaml" ]; then
    kubectl apply -f "$REPO_ROOT/monitoring/prometheus/prometheus.yaml" || echo "⚠️  Prometheus deployment may need additional configuration"
else
    echo "⚠️  Prometheus manifest not found, skipping..."
fi

# Deploy Grafana
echo "📊 Deploying Grafana..."
kubectl apply -f "$REPO_ROOT/monitoring/grafana/datasources.yaml"
kubectl apply -f "$REPO_ROOT/monitoring/grafana/grafana.yaml"

echo "⏳ Waiting for kube-state-metrics to be ready..."
kubectl wait --for=condition=available deployment/kube-state-metrics -n monitoring --timeout=120s || true

echo "⏳ Waiting for Prometheus to be ready..."
kubectl wait --for=condition=available deployment/prometheus -n monitoring --timeout=120s || true

echo "⏳ Waiting for Grafana to be ready..."
kubectl wait --for=condition=ready pod -l app=grafana -n monitoring --timeout=120s || {
    echo "⚠️  Grafana not ready yet, but continuing..."
}

echo ""
echo "✅ Grafana deployed!"
echo ""
echo "🔗 Access Grafana:"
echo "   1. Port-forward: kubectl port-forward -n monitoring svc/grafana 3000:3000"
echo "   2. Open: http://localhost:3000"
echo "   3. Login: admin/admin"
echo ""
echo "📊 Import dashboard:"
echo "   ./import-dashboard.sh"
echo ""

