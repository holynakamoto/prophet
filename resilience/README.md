# Resilience Testing

Chaos experiments to validate Prophet's self-healing capabilities.

## Overview

Prophet uses [Chaos Mesh](https://chaos-mesh.org/) for controlled fault injection. Each experiment validates that operators correctly detect and remediate failures.

## Experiments

| Experiment | Validates | Operator |
|------------|-----------|----------|
| [pod-failure.yaml](./chaos-experiments/pod-failure.yaml) | Pod crash recovery | AnomalyRemediator |
| [karpenter-node-failure.yaml](./chaos-experiments/karpenter-node-failure.yaml) | Node replacement | Karpenter |

## Prerequisites

Before running experiments:

1. **Deploy Chaos Mesh**
   ```bash
   kubectl apply -f https://raw.githubusercontent.com/chaos-mesh/chaos-mesh/master/manifests/crd.yaml
   helm install chaos-mesh chaos-mesh/chaos-mesh -n chaos-mesh --create-namespace
   ```

2. **Deploy Prophet Operators**
   ```bash
   kubectl apply -f clusters/common/aiops/operators/
   ```

3. **Deploy Monitoring Stack** (for observability)
   ```bash
   kubectl apply -f monitoring/prometheus/
   kubectl apply -f monitoring/grafana/
   ```

4. **Deploy Target Workload**
   ```bash
   kubectl apply -f demo/remediation-chaos/vulnerable-app.yaml
   ```

---

## Experiment: Pod Failure

**File:** `chaos-experiments/pod-failure.yaml`

### What It Does

Injects pod failures into workloads labeled `app: backend` in the `default` namespace.

### Preconditions

- [ ] Chaos Mesh installed and running
- [ ] AnomalyRemediator operator deployed
- [ ] Target deployment exists with label `app: backend`
- [ ] AnomalyAction CR configured for the target namespace

### Expected Signals

| Signal | Source | Expected Value |
|--------|--------|----------------|
| Pod restart count | `kubectl get pods` | Increases during experiment |
| `prophet_anomaly_detected_total` | Prometheus | Increments |
| `prophet_remediation_executed_total` | Prometheus | Increments |
| AnomalyAction status | `kubectl get anomalyactions` | Phase cycles: Detected → Resolved |
| K8sGPT analysis | AnomalyAction status | Contains failure explanation |

### Pass Criteria

✅ **Pass** if:
- Anomaly detected within 30s of pod failure
- Pod restarted automatically (not manual intervention)
- AnomalyAction status shows `phase: Resolved`
- Prometheus metrics recorded the event

❌ **Fail** if:
- No anomaly detected after 60s
- Pod remains in Failed/CrashLoopBackOff for >2 minutes
- Operator logs show errors
- Metrics not recorded

### How to Run

```bash
# 1. Deploy experiment
kubectl apply -f resilience/chaos-experiments/pod-failure.yaml

# 2. Watch pods
kubectl get pods -w -l app=backend

# 3. Watch AnomalyAction
kubectl get anomalyactions -w

# 4. Check metrics (if Prometheus deployed)
curl -s "http://localhost:9090/api/v1/query?query=prophet_remediation_executed_total"
```

### Cleanup

```bash
kubectl delete -f resilience/chaos-experiments/pod-failure.yaml
```

---

## Experiment: Karpenter Node Failure

**File:** `chaos-experiments/karpenter-node-failure.yaml`

### What It Does

Terminates a Karpenter-managed node to validate automatic node replacement.

### Preconditions

- [ ] Karpenter installed and configured
- [ ] At least 2 nodes in the cluster
- [ ] Workloads with PodDisruptionBudgets (to prevent full outage)

### Expected Signals

| Signal | Source | Expected Value |
|--------|--------|----------------|
| Node count | `kubectl get nodes` | Temporarily decreases, then recovers |
| Pod rescheduling | `kubectl get pods -o wide` | Pods move to new node |
| Karpenter provisioner | Karpenter logs | Shows node creation |

### Pass Criteria

✅ **Pass** if:
- New node provisioned within 2 minutes
- All pods rescheduled successfully
- No persistent pod failures

❌ **Fail** if:
- Node not replaced within 5 minutes
- Pods stuck in Pending
- Karpenter errors in logs

### Cleanup

```bash
kubectl delete -f resilience/chaos-experiments/karpenter-node-failure.yaml
```

---

## Running All Experiments

```bash
# Run pod failure test
kubectl apply -f resilience/chaos-experiments/pod-failure.yaml
sleep 300  # Wait for experiment duration
kubectl delete -f resilience/chaos-experiments/pod-failure.yaml

# Check results
kubectl get anomalyactions -A -o wide
```

## AI Validation

Each experiment includes optional AI validation:

1. **K8sGPT Analysis**: Post-experiment diagnostics
2. **Grafana ML Check**: Anomaly detection verification
3. **Report Generation**: Markdown summary of findings

See the `ai-validate-*` Job in each experiment file.

## Writing New Experiments

1. Create YAML in `chaos-experiments/`
2. Add documentation section in this README
3. Include:
   - Preconditions checklist
   - Expected signals table
   - Pass/fail criteria
4. Test locally before committing

## See Also

- [demo/remediation-chaos/](../demo/remediation-chaos/) - Interactive demo
- [operators/anomaly-remediator/](../operators/anomaly-remediator/) - Operator docs
- [Chaos Mesh Documentation](https://chaos-mesh.org/docs/)

