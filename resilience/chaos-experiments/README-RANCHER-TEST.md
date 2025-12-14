# Rancher Pod Restart Storm Remediation Test

## Overview

This test validates that the **AnomalyRemediator operator** automatically detects and remediates pods experiencing restart storms (excessive restarts).

## Test Scenario

**Target**: Rancher pod in `cattle-system` namespace with >3 restarts  
**Detection**: Pod status monitoring (restart count, phase, container status)  
**Remediation**: Automatic pod restart (delete pod, ReplicaSet recreates)  
**Cooldown**: 60 seconds between remediations

## Prerequisites

```bash
# 1. Deploy AnomalyRemediator operator
kubectl apply -f clusters/common/aiops/operators/anomaly-remediator.yaml

# 2. Verify operator is running
kubectl get pods -n prophet-operators -l app=anomaly-remediator

# 3. Verify Rancher is deployed (or use any pod with restart issues)
kubectl get pods -n cattle-system -l app=rancher
```

## Running the Test

### Step 1: Deploy AnomalyAction

```bash
kubectl apply -f resilience/chaos-experiments/rancher-pod-restart-remediation.yaml
```

### Step 2: Monitor Remediation

```bash
# Watch AnomalyAction status
kubectl get anomalyactions rancher-pod-remediation -n cattle-system -w

# Watch Rancher pods
kubectl get pods -n cattle-system -l app=rancher -w

# Watch operator logs
kubectl logs -n prophet-operators -l app=anomaly-remediator -f
```

### Step 3: Verify Remediation

```bash
# Check AnomalyAction status
kubectl get anomalyactions rancher-pod-remediation -n cattle-system -o yaml | grep -A 10 "status:"

# Expected output:
# status:
#   phase: Resolved
#   lastDetected: "2025-12-13T20:53:10Z"
#   lastRemediated: "2025-12-13T20:53:10Z"
#   remediationCount: 1
```

## Expected Behavior

### Timeline

1. **0s**: AnomalyAction created, operator starts watching
2. **~30s**: Operator detects pod with >3 restarts
3. **~30s**: AnomalyAction status.phase = "Detected"
4. **~30s**: Operator deletes the failing pod
5. **~30s**: ReplicaSet creates new pod
6. **~30s**: AnomalyAction status.phase = "Resolved", remediationCount = 1

### Signals to Watch

| Signal | Source | Expected Value |
|--------|--------|----------------|
| AnomalyAction phase | `kubectl get anomalyactions` | Detected → Resolved |
| remediationCount | AnomalyAction status | Increments from 0 |
| Pod status | `kubectl get pods` | Old pod: Terminating, New pod: Running |
| Operator logs | `kubectl logs` | "Anomaly detected", "Restarting pod" |

## Pass Criteria

✅ **Pass** if:
- Anomaly detected within 30s
- Pod automatically restarted (old pod Terminating, new pod Created)
- `remediationCount > 0`
- `phase = Resolved`
- No operator errors in logs

❌ **Fail** if:
- No detection after 60s
- Pod not restarted after 2 minutes
- `remediationCount` remains 0
- Operator logs show errors
- AnomalyAction stuck in "Pending" or "Failed"

## Metrics

The operator exposes Prometheus metrics:

```bash
# Check metrics (if Prometheus deployed)
curl http://<operator-pod>:8080/metrics | grep prophet_remediation
```

Expected metrics:
- `prophet_anomaly_detected_total` - increments on detection
- `prophet_remediation_executed_total` - increments on restart
- `prophet_remediation_duration_seconds` - time to execute

## Troubleshooting

### Operator Not Running

```bash
# Check deployment
kubectl get deployment anomaly-remediator-controller-manager -n prophet-operators

# Check pod status
kubectl describe pod -n prophet-operators -l app=anomaly-remediator

# Check logs
kubectl logs -n prophet-operators -l app=anomaly-remediator --tail=50
```

### AnomalyAction Not Detecting

```bash
# Verify pod has >3 restarts
kubectl get pods -n cattle-system -l app=rancher -o jsonpath='{.items[0].status.containerStatuses[0].restartCount}'

# Check operator is watching the namespace
kubectl get anomalyactions -n cattle-system

# Check RBAC permissions
kubectl auth can-i get pods -n cattle-system --as=system:serviceaccount:prophet-operators:anomaly-remediator-controller-manager
```

### Remediation Not Executing

```bash
# Check cooldown period
kubectl get anomalyactions rancher-pod-remediation -n cattle-system -o jsonpath='{.status.lastRemediated}'

# Check if approval required
kubectl get anomalyactions rancher-pod-remediation -n cattle-system -o jsonpath='{.spec.remediation.requireApproval}'

# Verify RBAC for pod deletion
kubectl auth can-i delete pods -n cattle-system --as=system:serviceaccount:prophet-operators:anomaly-remediator-controller-manager
```

## Cleanup

```bash
# Remove AnomalyAction
kubectl delete anomalyaction rancher-pod-remediation -n cattle-system

# Optional: Remove operator
kubectl delete -f clusters/common/aiops/operators/anomaly-remediator.yaml
```

## Related Tests

- [Pod Failure Chaos Experiment](./pod-failure.yaml) - General pod failure remediation
- [Demo Script](../demo/remediation-chaos/demo.sh) - Interactive self-healing demo

## See Also

- [AnomalyRemediator Operator README](../../operators/anomaly-remediator/README.md)
- [Resilience Testing README](../README.md)

