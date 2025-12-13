# Prophet Self-Healing Demo - Grafana Dashboard

## Overview

The Prophet Self-Healing Demo dashboard provides real-time visualization of autonomous remediation in action. Watch chaos injection, pod failures, and automatic recovery all in one stunning Grafana dashboard.

## Features

### Key Panels

1. **Pod Restart Rate (Chaos → Recovery)**
   - Shows restart spikes during chaos injection
   - Visualizes recovery as restarts drop to zero
   - Color-coded: Green (healthy) → Yellow (warning) → Red (critical)

2. **Anomalies Detected** (Stat Panel)
   - Real-time counter of anomalies detected
   - Updates automatically as operator detects issues

3. **Remediations Executed** (Stat Panel)
   - Total number of remediation actions taken
   - Shows operator activity in real-time

4. **Pod Status Over Time**
   - Tracks pod phases (Running, CrashLoopBackOff, Error, Pending)
   - Shows transition from failure to recovery

5. **Remediation Rate**
   - Rate of remediation actions per second
   - Grouped by remediation type (restart, scale, etc.)

6. **Pod Restarts by AnomalyAction**
   - Shows which AnomalyAction resources are triggering restarts
   - Useful for multi-action scenarios

7. **Current Pod Status** (Table)
   - Real-time table of all pods
   - Color-coded by status (green=Running, red=Error/CrashLoopBackOff)

8. **Remediation Duration**
   - P50 and P95 latency for remediation actions
   - Performance metrics for operator efficiency

## Setup

### Prerequisites

- Prometheus scraping operator metrics (port 8080)
- Grafana with Prometheus datasource configured
- kube-state-metrics for pod metrics

### Import Dashboard

**Option 1: Using the import script**
```bash
cd demo/remediation-chaos
./import-dashboard.sh
```

**Option 2: Manual import**
1. Port-forward Grafana: `kubectl port-forward -n monitoring svc/grafana 3000:3000`
2. Open http://localhost:3000
3. Go to Dashboards → Import
4. Upload `monitoring/grafana/dashboards/prophet-self-healing-demo.json`

### Metrics Required

The dashboard uses these Prometheus metrics:

**From AnomalyRemediator Operator:**
- `anomaly_remediator_actionable_anomalies_total` (used by the dashboard for anomaly↔remediation parity)
- `anomaly_remediator_anomalies_detected_total` (raw detections; may exceed remediations due to cooldown / frequent reconciles)
- `anomaly_remediator_remediations_executed_total`
- `anomaly_remediator_remediation_duration_seconds`
- `anomaly_remediator_pod_restarts_total`

**From kube-state-metrics:**
- `kube_pod_container_status_restarts_total`
- `kube_pod_info`
- `kube_pod_status_phase`

**Prometheus ServiceMonitor:**
Ensure Prometheus is scraping the operator:
```yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: anomaly-remediator
  namespace: prophet-operators
spec:
  selector:
    matchLabels:
      control-plane: controller-manager
  endpoints:
  - port: http-metrics
    interval: 15s
```

## Demo Flow

1. **Before Chaos**: Dashboard shows stable state (green, low restarts)
2. **Chaos Injection**: Restart rate spikes (red), pods show CrashLoopBackOff
3. **Remediation**: Operator detects → Anomalies counter increases → Remediations counter increases
4. **Recovery**: Restart rate drops, pods return to Running (green)
5. **Stabilization**: All metrics return to baseline

## Customization

### Change Namespace

If using a different namespace, update queries:
- Replace `namespace="demo-prophet"` with your namespace
- Or use Grafana variables for dynamic selection

### Add Annotations

The dashboard includes annotation support for:
- Chaos injection events
- Remediation events

Add annotations via Grafana UI or API to mark key moments.

## Troubleshooting

### No Metrics Showing

1. Check Prometheus is scraping operator:
   ```bash
   curl http://localhost:9090/api/v1/targets | jq '.data.activeTargets[] | select(.labels.job | contains("anomaly"))'
   ```

2. Verify metrics endpoint:
   ```bash
   kubectl port-forward -n prophet-operators svc/anomaly-remediator-controller-manager-metrics 8080:8080
   curl http://localhost:8080/metrics | grep anomaly_remediator
   ```

3. Check ServiceMonitor exists:
   ```bash
   kubectl get servicemonitor -n prophet-operators
   ```

### Dashboard Not Loading

1. Verify Prometheus datasource is configured in Grafana
2. Check dashboard JSON is valid: `jq . monitoring/grafana/dashboards/prophet-self-healing-demo.json`
3. Ensure all required metrics are available

## Tips for Demos

1. **Set Time Range**: Use "Last 15 minutes" for focused view
2. **Auto-Refresh**: Enable 5s refresh for real-time updates
3. **Full Screen**: Use F11 or Grafana's fullscreen mode
4. **Side-by-Side**: Open Grafana and `kubectl get pods -w` side-by-side
5. **Screenshots**: Capture before/after states for presentations

---

**This dashboard transforms the demo from terminal output to visual storytelling. Customers see the magic happen in real-time!** 🚀📊

