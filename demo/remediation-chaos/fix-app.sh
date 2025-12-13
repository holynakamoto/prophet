#!/usr/bin/env bash
# Quick script to stabilize the crashy app after demo
# Usage: ./fix-app.sh [namespace]

NS=${1:-demo-prophet}

echo "🔧 Stabilizing crashy app in namespace: $NS"
kubectl patch configmap crashy-app-config -n $NS --type merge -p '{"data":{"behavior":"stable"}}' 2>/dev/null || {
    echo "⚠️  ConfigMap not found. Creating it..."
    kubectl create configmap crashy-app-config -n $NS --from-literal=behavior=stable --dry-run=client -o yaml | kubectl apply -f -
}

echo "⏳ Waiting for pods to restart in stable mode..."
sleep 10

echo ""
echo "✅ App stabilized! Current status:"
kubectl get pods -n $NS -l app=crashy-app
echo ""
kubectl get anomalyaction crashy-remediation -n $NS -o jsonpath='Status: {.status.phase} | Remediations: {.status.remediationCount}{"\n"}' 2>/dev/null || echo "AnomalyAction not found"

