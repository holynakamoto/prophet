# Prophet v4 - Custom Go Operators Guide

This document provides detailed information about the custom Golang operators in Prophet v4.

## Overview

Prophet v4 includes three custom Kubernetes operators built with Kubebuilder (controller-runtime):

1. **AnomalyRemediator**: Detects anomalies and performs automated remediation
2. **PredictiveScaler**: Uses Grafana ML forecasts to proactively scale Karpenter NodePools
3. **SLOEnforcer**: Monitors SLOs and enforces policies when violations occur

## Architecture

All operators follow the standard Kubebuilder pattern:

```
operators/
├── <operator-name>/
│   ├── api/v1alpha1/          # CRD definitions
│   ├── controllers/            # Reconciliation logic
│   ├── cmd/                    # Main entry point
│   ├── config/                 # Deployment manifests
│   ├── Dockerfile             # Container image
│   ├── Makefile               # Build commands
│   └── go.mod                 # Go dependencies
```

## AnomalyRemediator Operator

### Purpose
Automatically detects anomalies from Prometheus/Grafana ML and performs remediation actions (restart pods, scale deployments, etc.).

### CRD: AnomalyAction

```yaml
apiVersion: aiops.prophet.io/v1alpha1
kind: AnomalyAction
metadata:
  name: backend-memory-anomaly
  namespace: default
spec:
  source: prometheus              # Source: "prometheus", "grafana-ml", "otel"
  metric: container_memory_working_set_bytes
  threshold: "> 1Gi"
  remediation:
    type: restart                 # "restart", "scale", "alert"
    podSelector:
      app: backend
    requireApproval: false
    cooldownSeconds: 300
  target:
    namespace: default
    labels:
      app: backend
    resourceType: Pod
  k8sgpt:
    enabled: true
    endpoint: http://k8sgpt-operator.k8sgpt.svc.cluster.local:8080
  webhookUrl: ""                  # Optional: PagerDuty, Slack, etc.
```

### Status Fields

- `phase`: "Pending", "Detected", "Remediating", "Resolved", "Failed"
- `lastDetected`: Timestamp of last anomaly detection
- `lastRemediated`: Timestamp of last remediation
- `remediationCount`: Number of remediations performed
- `k8sgptAnalysis`: K8sGPT analysis result (if enabled)

### Example Use Cases

1. **Memory Leak Detection**: Restart pods when memory exceeds threshold
2. **CPU Spikes**: Scale deployment when CPU anomaly detected
3. **Pod Failures**: Auto-restart failed pods with K8sGPT analysis

## PredictiveScaler Operator

### Purpose
Uses Grafana ML forecasts to proactively adjust Karpenter NodePools before demand spikes.

### CRD: PredictiveScale

```yaml
apiVersion: aiops.prophet.io/v1alpha1
kind: PredictiveScale
metadata:
  name: cpu-forecast-scaling
  namespace: default
spec:
  forecastQuery: ml_forecast(sum(rate(container_cpu_usage_seconds_total[5m])), 1h)
  nodePoolRef:
    name: default
  horizon: 1h                      # Forecast horizon
  thresholdPercent: 20.0          # Trigger scaling if forecast > 20% increase
  action: provision                # "provision", "consolidate", "adjust"
  grafanaEndpoint: http://grafana.monitoring.svc.cluster.local:3000
```

### Status Fields

- `phase`: "Monitoring", "Scaling", "Complete", "Failed"
- `lastForecast`: Last forecast value
- `lastScaled`: Timestamp of last scaling action
- `scalingCount`: Number of scaling actions performed

### Example Use Cases

1. **Pre-provision GPU Nodes**: Forecast GPU demand and provision ahead
2. **Spot Instance Optimization**: Adjust NodePool capacity types based on forecasts
3. **Cost Optimization**: Consolidate nodes when forecast shows low demand

## SLOEnforcer Operator

### Purpose
Monitors SLOs and automatically enforces policies when violations occur or are predicted.

### CRD: SLOViolation

```yaml
apiVersion: aiops.prophet.io/v1alpha1
kind: SLOViolation
metadata:
  name: backend-availability-slo
  namespace: default
spec:
  sloName: backend-availability
  sloTarget: "99.9%"
  errorBudgetThreshold: 0.1       # Trigger action when budget < 10%
  actions:
  - type: scale
    value: "increase"
  - type: alert
  hpaRef:
    name: backend-hpa
    namespace: default
  enableChaos: false               # Enable chaos testing on violations
```

### Status Fields

- `phase`: "Monitoring", "Violated", "Remediating", "Resolved"
- `errorBudgetRemaining`: Current error budget percentage
- `timeToExhaustion`: Predicted days until budget exhaustion
- `lastViolated`: Timestamp of last violation
- `violationCount`: Number of violations detected

### Example Use Cases

1. **Availability SLO**: Auto-scale HPA when error budget drops
2. **Latency SLO**: Trigger rollback when latency SLO violated
3. **Error Rate SLO**: Enable chaos testing to improve resilience

## Development

### Building Operators

```bash
# Build operator binary
cd operators/anomaly-remediator
make build

# Build Docker image
make docker-build IMG=ghcr.io/prophet-aiops/prophet-anomaly-remediator:latest

# Push to registry
make docker-push IMG=ghcr.io/prophet-aiops/prophet-anomaly-remediator:latest
```

### Running Locally

```bash
# Install CRDs
make install

# Run controller locally (requires kubeconfig)
make run
```

### Testing

```bash
# Run unit tests
make test

# Run with coverage
go test ./... -coverprofile cover.out
```

### Generating Manifests

```bash
# Generate CRDs and RBAC
make manifests

# Generate code (DeepCopy, etc.)
make generate
```

## Deployment

### Via Kustomize

```bash
# Deploy all operators
kubectl apply -f clusters/common/aiops/operators/

# Or individually
kubectl apply -f clusters/common/aiops/operators/anomaly-remediator.yaml
```

### Via ArgoCD

Add to your ArgoCD Application:

```yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: prophet-operators
spec:
  source:
    repoURL: https://github.com/prophet-aiops/prophet
    path: clusters/common/aiops/operators
    targetRevision: main
```

## Monitoring

### Operator Metrics

All operators expose Prometheus metrics on port 8080:

- `controller_runtime_reconcile_total`: Total reconciliations
- `controller_runtime_reconcile_errors_total`: Reconciliation errors
- `controller_runtime_reconcile_time_seconds`: Reconciliation duration

### Logging

View operator logs:

```bash
kubectl logs -n prophet-operators deployment/anomaly-remediator-controller-manager
kubectl logs -n prophet-operators deployment/predictive-scaler-controller-manager
kubectl logs -n prophet-operators deployment/slo-enforcer-controller-manager
```

### Health Checks

All operators expose health endpoints:

- `/healthz`: Liveness probe
- `/readyz`: Readiness probe
- `/metrics`: Prometheus metrics

## Troubleshooting

### Operator Not Starting

1. Check pod status:
```bash
kubectl get pods -n prophet-operators
kubectl describe pod <pod-name> -n prophet-operators
```

2. Check RBAC:
```bash
kubectl auth can-i <verb> <resource> --as=system:serviceaccount:prophet-operators:<operator>-controller-manager
```

3. Check logs for errors:
```bash
kubectl logs -n prophet-operators deployment/<operator>-controller-manager --tail=100
```

### CRD Not Found

```bash
# Verify CRD is installed
kubectl get crd anomalyactions.aiops.prophet.io

# Reinstall if needed
kubectl apply -f clusters/common/aiops/operators/anomaly-remediator.yaml
```

### Reconciliation Not Working

1. Check resource status:
```bash
kubectl get anomalyactions -A
kubectl describe anomalyaction <name> -n <namespace>
```

2. Check events:
```bash
kubectl get events -n <namespace> --sort-by='.lastTimestamp'
```

3. Enable debug logging:
```yaml
# In operator deployment, add:
args:
- --leader-elect
- --zap-log-level=debug
```

## Best Practices

1. **Start with Approval Required**: Set `requireApproval: true` for production
2. **Use Cooldown Periods**: Prevent rapid remediation loops
3. **Monitor Operator Metrics**: Track reconciliation rates and errors
4. **Test in Staging First**: Validate operator behavior before production
5. **Use K8sGPT Integration**: Enable diagnostics for better insights
6. **Set Resource Limits**: Prevent operators from consuming excessive resources

## Future Enhancements

- [ ] Webhook-based admission control
- [ ] Multi-cluster operator federation
- [ ] Operator metrics dashboards
- [ ] Advanced anomaly detection algorithms
- [ ] Integration with more external systems (PagerDuty, Slack, etc.)

