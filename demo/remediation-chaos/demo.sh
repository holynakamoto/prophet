#!/usr/bin/env bash
set -euo pipefail

NS=demo-prophet
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo "🚀 Setting up Prophet Remediation Chaos Demo in namespace: $NS"
echo ""

# Create namespace
kubectl create ns $NS --dry-run=client -o yaml | kubectl apply -f -

echo "🧼 Resetting demo resources (so failure/remediation is obvious)..."
kubectl delete anomalyaction crashy-remediation -n $NS --ignore-not-found=true >/dev/null 2>&1 || true

echo "📦 Deploying vulnerable app..."
kubectl apply -f "$SCRIPT_DIR/vulnerable-app.yaml" -n $NS

echo "⏳ Letting the app start... (it is designed to fail)"
sleep 5

echo ""
echo "💥 Phase 1: Show the failure (before Prophet remediation is enabled)"
echo "   Waiting until at least one pod is failing (CrashLoopBackOff/Error)..."
FAIL_TIMEOUT_SECONDS=90
START_TS=$(date +%s)
while true; do
    FAILING_COUNT=$(kubectl get pods -n $NS -l app=crashy-app --no-headers 2>/dev/null | awk '$3 != "Running" {c++} END{print c+0}')
    if [ "${FAILING_COUNT:-0}" -gt 0 ]; then
        break
    fi
    NOW_TS=$(date +%s)
    if [ $((NOW_TS-START_TS)) -ge $FAIL_TIMEOUT_SECONDS ]; then
        echo "   ⚠️  Timed out waiting to observe a failure. Continuing anyway..."
        break
    fi
    sleep 3
done

echo ""
echo "📊 Pod status (note the failures):"
kubectl get pods -n $NS -l app=crashy-app || true

echo "🤖 Deploying AnomalyRemediator operator..."
# Check if operator is already deployed
if ! kubectl get deployment -n prophet-operators anomaly-remediator-controller-manager &>/dev/null; then
    echo "   Operator not found in prophet-operators namespace. Deploying from repo..."
    if [ -f "$SCRIPT_DIR/../../clusters/common/aiops/operators/anomaly-remediator.yaml" ]; then
        kubectl apply -f "$SCRIPT_DIR/../../clusters/common/aiops/operators/anomaly-remediator.yaml"
    else
        echo "   ⚠️  Operator manifest not found. Please deploy the operator manually:"
        echo "      kubectl apply -f clusters/common/aiops/operators/anomaly-remediator.yaml"
        echo "   Continuing with demo anyway..."
    fi
else
    echo "   ✓ Operator already deployed"
fi

echo ""
echo "🔧 Ensuring the operator image includes the demo metrics (local build for kind)..."
if command -v kind >/dev/null 2>&1 && command -v docker >/dev/null 2>&1 && kind get clusters 2>/dev/null | grep -q '^prophet-local$'; then
    IMG="ghcr.io/prophet-aiops/prophet-anomaly-remediator:demo-metrics"
    echo "   Building: $IMG"
    (cd "$SCRIPT_DIR/../../operators/anomaly-remediator" && make docker-build IMG="$IMG") >/dev/null
    echo "   Loading image into kind cluster: prophet-local"
    kind load docker-image "$IMG" --name prophet-local >/dev/null
    echo "   Updating deployment to use local image (imagePullPolicy: Never)"
    kubectl -n prophet-operators set image deployment/anomaly-remediator-controller-manager manager="$IMG" >/dev/null
    kubectl patch deployment -n prophet-operators anomaly-remediator-controller-manager --type='json' -p='[{\"op\":\"add\",\"path\":\"/spec/template/spec/containers/0/imagePullPolicy\",\"value\":\"Never\"}]' >/dev/null 2>&1 || true
    kubectl rollout status deployment/anomaly-remediator-controller-manager -n prophet-operators --timeout=120s >/dev/null || true
else
    echo "   Skipping local build (requires kind+docker, cluster name prophet-local)."
fi

echo ""
if ! kubectl get deployment -n monitoring prometheus &>/dev/null; then
    echo "📈 Grafana/Prometheus not detected (recommended for a visual demo)."
    read -p "Press Enter to install monitoring (Prometheus + Grafana + kube-state-metrics)..." 
    echo ""
    "$SCRIPT_DIR/setup-grafana.sh" || true
    echo ""
else
    echo "📈 Monitoring detected (Prometheus/Grafana). Dashboard should have data during the demo."
fi
echo ""

echo "🔌 Phase 2: Enable Prophet remediation (Golang operator will start fixing the cluster)"
read -p "Press Enter to enable auto-remediation (apply AnomalyAction)..." 
echo ""
echo "📋 Applying autonomous remediation policy..."
kubectl apply -f "$SCRIPT_DIR/anomaly-action.yaml" -n $NS

echo "   🔧 Ensuring operator metrics are scrapeable by Prometheus..."
kubectl patch deployment -n prophet-operators anomaly-remediator-controller-manager --type merge -p \
  '{\"spec\":{\"template\":{\"metadata\":{\"annotations\":{\"prometheus.io/scrape\":\"true\",\"prometheus.io/port\":\"8080\",\"prometheus.io/path\":\"/metrics\"}}}}}' >/dev/null 2>&1 || true

echo ""
echo "✅ Remediation enabled. Watch the operator log for detection + remediation:"
echo "   kubectl logs -n prophet-operators deploy/anomaly-remediator-controller-manager -f --tail=50"

echo ""
if kubectl get deployment -n monitoring prometheus &>/dev/null; then
    echo "📈 Prometheus is available. (These will become non-zero once Prophet detects/remediates.)"
    DETECTED=$(kubectl exec -n monitoring deploy/prometheus -- wget -qO- 'http://localhost:9090/api/v1/query?query=sum(anomaly_remediator_anomalies_detected_total{namespace=\"demo-prophet\"})' \
      | jq -r '.data.result[0].value[1] // \"0\"' 2>/dev/null || echo "0")
    REMEDIATED=$(kubectl exec -n monitoring deploy/prometheus -- wget -qO- 'http://localhost:9090/api/v1/query?query=sum(anomaly_remediator_remediations_executed_total{namespace=\"demo-prophet\"})' \
      | jq -r '.data.result[0].value[1] // \"0\"' 2>/dev/null || echo "0")
    echo "   anomaly_detected_total:      $DETECTED"
    echo "   remediation_executed_total:  $REMEDIATED"
else
    echo "📈 Tip: run $SCRIPT_DIR/setup-grafana.sh to get Prometheus+Grafana dashboards."
fi

echo "🔧 Checking for Chaos Mesh..."
if ! kubectl get crd podchaos.chaos-mesh.org &>/dev/null; then
    echo "   Installing Chaos Mesh CRDs..."
    kubectl apply -f https://mirrors.chaos-mesh.org/latest/crd.yaml || {
        echo "   ⚠️  Failed to install Chaos Mesh CRDs. Continuing without chaos injection..."
        CHAOS_AVAILABLE=false
    }
    CHAOS_AVAILABLE=true
else
    echo "   ✓ Chaos Mesh CRDs already installed"
    CHAOS_AVAILABLE=true
fi

if [ "$CHAOS_AVAILABLE" = true ]; then
    if ! kubectl get namespace chaos-testing &>/dev/null; then
        echo "   Installing Chaos Mesh operator..."
        kubectl create namespace chaos-testing --dry-run=client -o yaml | kubectl apply -f -
        kubectl apply -f https://mirrors.chaos-mesh.org/latest/chaos-mesh.yaml -n chaos-testing || {
            echo "   ⚠️  Failed to install Chaos Mesh. Continuing without chaos injection..."
            CHAOS_AVAILABLE=false
        }
        echo "   ⏳ Waiting for Chaos Mesh to be ready..."
        sleep 15
    fi
fi

echo ""
echo "✅ Demo ready! Watch remediation in real-time:"
echo ""
echo "   📊 Pods:    kubectl get pods -n $NS -w"
echo "   📝 Events:  kubectl get events -n $NS --sort-by='.lastTimestamp' -w"
echo "   📋 Status:  kubectl get anomalyaction crashy-remediation -n $NS -o yaml -w"
echo "   📜 Logs:    kubectl logs -n prophet-operators -l control-plane=controller-manager -f --tail=50"
echo ""
echo "   📈 Grafana Dashboard:"
echo "     Quick setup: ./setup-grafana.sh"
echo "     Access: ./access-grafana.sh"
echo "     Or manually: kubectl port-forward -n monitoring svc/grafana 3000:3000"
echo "     Then open: http://localhost:3000 (admin/admin)"
echo ""

if [ "$CHAOS_AVAILABLE" = true ]; then
    echo ""
    echo "Choose chaos injection method:"
    echo "  1) Chaos Mesh PodChaos (one-time kill)"
    echo "  2) Repeated manual kills (better for demo visualization)"
    read -p "Enter choice [1 or 2, default: 2]: " CHAOS_METHOD
    CHAOS_METHOD=${CHAOS_METHOD:-2}
    
    if [ "$CHAOS_METHOD" = "1" ]; then
        read -p "Press Enter to inject chaos (Chaos Mesh PodChaos)..." 
        echo ""
        echo "🔥 Injecting chaos: Killing pods with Chaos Mesh..."
        kubectl apply -f "$SCRIPT_DIR/chaos-pod-kill.yaml" -n $NS
        
        echo ""
        echo "⏱️  Chaos active for 2 minutes... Watch the operator auto-restart pods!"
        echo ""
        
        # Show live status
        for i in {1..12}; do
            echo "--- Status check $i/12 ---"
            kubectl get pods -n $NS -l app=crashy-app
            kubectl get anomalyaction crashy-remediation -n $NS -o jsonpath='{.status.phase}' 2>/dev/null || echo "Pending"
            echo ""
            sleep 10
        done
        
        echo "🧹 Cleaning up chaos experiment..."
        kubectl delete podchaos pod-kill-demo -n $NS --ignore-not-found=true || true
    else
        read -p "Press Enter to start repeated pod kills (runs in background)..." 
        echo ""
        echo "🔥 Starting repeated pod kills (runs for 2 minutes)..."
        echo "   Watch Grafana dashboard or run: kubectl get pods -n $NS -w"
        echo ""
        
        # Run chaos injection in background
        "$SCRIPT_DIR/inject-chaos.sh" "$NS" 120 > /tmp/chaos-injection.log 2>&1 &
        CHAOS_PID=$!
        echo "   Chaos injection running (PID: $CHAOS_PID)"
        echo "   Logs: tail -f /tmp/chaos-injection.log"
        echo ""
        
        # Show live status while chaos runs (slower for better visibility)
        echo "   Watching remediation in real-time (slower updates for visibility)..."
        for i in {1..8}; do
            echo ""
            echo "============================================================"
            echo "--- Status check $i/8 [$(date +%H:%M:%S)] ---"
            echo "📊 Failing pods (what Prophet is fixing):"
            FAILING_LINES=$(kubectl get pods -n $NS -l app=crashy-app --no-headers 2>/dev/null | awk '$3 != "Running" {printf "  %s: %s\n", $1, $3}')
            if [ -n "${FAILING_LINES:-}" ]; then
                echo "$FAILING_LINES"
            else
                echo "  (none)"
            fi
            echo ""
            echo "📊 All pods:"
            kubectl get pods -n $NS -l app=crashy-app --no-headers 2>/dev/null | awk '{printf "  %s: %s\n", $1, $3}' || true
            echo ""
            echo "🤖 AnomalyAction Status:"
            kubectl get anomalyaction crashy-remediation -n $NS -o jsonpath='  Phase: {.status.phase} | Remediations: {.status.remediationCount} | Last Detected: {.status.lastDetected}{"\n"}' 2>/dev/null || echo "  Pending"
            echo ""
            echo "🧠 Operator (recent remediation logs):"
            kubectl logs -n prophet-operators deploy/anomaly-remediator-controller-manager --tail=200 2>/dev/null | egrep -i 'anomaly detected|restarting pod|remediation|resolved' | tail -5 || true
            echo ""
            echo "⏱️  Next update in 15 seconds... (Watch Grafana for real-time metrics!)"
            sleep 15
        done
        
        # Wait for chaos to finish
        wait $CHAOS_PID 2>/dev/null || true
        echo "✅ Chaos injection complete!"
    fi
else
    echo "⚠️  Chaos Mesh not available. Pods will crash naturally due to the app design."
    echo "   Watch for the AnomalyRemediator to detect and restart them:"
    echo ""
    for i in {1..6}; do
        echo "--- Status check $i/6 ---"
        kubectl get pods -n $NS -l app=crashy-app
        kubectl get anomalyaction crashy-remediation -n $NS -o jsonpath='{.status.phase}' 2>/dev/null || echo "Pending"
        echo ""
        sleep 10
    done
fi

echo ""
read -p "Press Enter to stabilize the app (stop crashing)..." 

echo ""
echo "🔧 Stabilizing app (switching to stable mode)..."
kubectl patch configmap crashy-app-config -n $NS --type merge -p '{"data":{"behavior":"stable"}}' || true
echo "   Waiting for pods to restart in stable mode..."
sleep 15

echo ""
echo "✅ App is now stable! Showing final state:"
kubectl get pods -n $NS -l app=crashy-app
kubectl get anomalyaction crashy-remediation -n $NS -o jsonpath='Status: {.status.phase} | Remediations: {.status.remediationCount}{"\n"}'

echo ""
read -p "Press Enter to clean up demo resources..." 

echo ""
echo "🧹 Cleaning up..."
kubectl delete -f "$SCRIPT_DIR/anomaly-action.yaml" -n $NS --ignore-not-found=true || true
kubectl delete -f "$SCRIPT_DIR/vulnerable-app.yaml" -n $NS --ignore-not-found=true || true
kubectl delete ns $NS --force --grace-period=0 --ignore-not-found=true || true

echo ""
echo "✅ Demo complete!"
echo ""
echo "📊 Summary:"
echo "   The AnomalyRemediator detected pod failures and restarted them autonomously."
echo "   Check the operator logs to see the remediation actions:"
echo "   kubectl logs -n prophet-operators -l control-plane=controller-manager -f --tail=50"
echo ""
echo "🚀 Want to see it again? Just run ./demo.sh"
echo ""

