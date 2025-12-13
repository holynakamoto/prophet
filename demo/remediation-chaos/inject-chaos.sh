#!/usr/bin/env bash
# Repeatedly kill pods to simulate chaos for the demo
# Usage: ./inject-chaos.sh [namespace] [duration_seconds]

NS=${1:-demo-prophet}
DURATION=${2:-120}  # Default 2 minutes
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo "🔥 Injecting repeated pod kills for $DURATION seconds..."
echo "   This will kill pods every 30 seconds to demonstrate remediation"
echo ""

# Ensure app is in crash mode (not stable)
kubectl patch configmap crashy-app-config -n $NS --type merge -p '{"data":{"behavior":"crash"}}' 2>/dev/null || true
echo "   Switched app to crash mode"
sleep 5

# Function to kill all pods
kill_pods() {
    kubectl get pods -n $NS -l app=crashy-app -o name | xargs -r kubectl delete -n $NS
}

# Kill pods repeatedly (slower for better demo visibility)
START_TIME=$(date +%s)
ITERATION=0

while [ $(($(date +%s) - START_TIME)) -lt $DURATION ]; do
    ITERATION=$((ITERATION + 1))
    echo "[$(date +%H:%M:%S)] --- Chaos iteration $ITERATION ---"
    echo "🔥 Killing all crashy-app pods..."
    kill_pods
    echo "⏳ Waiting 30 seconds for operator to detect and remediate..."
    echo "   (This gives time to see the failure state before remediation)"
    sleep 30
done

echo ""
echo "✅ Chaos injection complete!"
echo "   The operator should have detected and remediated failures throughout."

