# Prophet Operators

Custom Kubernetes operators that power Prophet's self-healing and AIOps capabilities.

## Overview

| Operator | CRD | Purpose | Status |
|----------|-----|---------|--------|
| [health-check](./health-check/) | `HealthCheck` | Multi-probe health monitoring | ✅ Production |
| [budget-guard](./budget-guard/) | `BudgetGuard` | Cost budget enforcement | ✅ Production |
| [cost-alert](./cost-alert/) | `CostAlert` | Cost anomaly alerting | ✅ Production |
| [diagnostic-remediator](./diagnostic-remediator/) | `DiagnosticRemediation` | Application-specific remediation | ✅ Production |
| [label-enforcer](./label-enforcer/) | `LabelEnforcer` | Enforce required labels/annotations | ✅ Production |

## Quick Start

### Prerequisites

- Go 1.22+
- Docker
- kubectl configured to a cluster
- (Optional) Tilt for live development

### Run an Operator Locally

```bash
cd operators/anomaly-remediator

# Install dependencies and generate code
make generate manifests

# Run against current kubeconfig context
make run
```

### Run All Operators with Tilt

```bash
cd operators
tilt up
```

This starts all operators with live reload on code changes.

### Build and Push Images

```bash
cd operators/anomaly-remediator
make docker-build docker-push IMG=your-registry/anomaly-remediator:latest
```

## Development Workflow

### 1. Modify CRD Types

Edit `api/v1alpha1/*_types.go`, then regenerate:

```bash
make generate manifests
```

### 2. Run Tests

```bash
make test
```

### 3. Local CI (Lint + Test + Validate)

```bash
make local-ci
```

### 4. Build & Package for Release

```bash
# Build Docker image
make docker-build docker-push IMG=ghcr.io/prophet-aiops/prophet-label-enforcer:v1.0.0

# Lint and package Helm chart
make helm-lint
make helm-package

# Test chart installation
make helm-template
make helm-install  # In test environment
```

## Deployment

Operators can be deployed in multiple ways:

### Option 1: Helm Charts (Recommended)

Each operator includes a production-ready Helm chart:

```bash
# Install via Helm
helm install prophet-label-enforcer operators/label-enforcer/helm/label-enforcer
helm install prophet-health-check operators/health-check/helm/health-check

# Customize with values
helm install prophet-label-enforcer operators/label-enforcer/helm/label-enforcer \
  --set image.tag=v1.0.0 \
  --set watchNamespace=default
```

### Option 2: GitOps with ArgoCD/Flux

Use the provided overlays in `clusters/common/aiops/operators/` or point your GitOps tool to the Helm charts.

### Option 3: Direct Manifests

Apply the generated manifests directly:

```bash
kubectl apply -f clusters/common/aiops/operators/health-check.yaml
```

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                     Prophet Control Plane                        │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐           │
│  │ Anomaly      │  │ Predictive   │  │ SLO          │           │
│  │ Remediator   │  │ Scaler       │  │ Enforcer     │           │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘           │
│         │                 │                 │                    │
│         ▼                 ▼                 ▼                    │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │                   Kubernetes API Server                   │   │
│  └──────────────────────────────────────────────────────────┘   │
│         │                 │                 │                    │
│         ▼                 ▼                 ▼                    │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐           │
│  │ Prometheus   │  │ Grafana ML   │  │ K8sGPT       │           │
│  │ Metrics      │  │ Forecasts    │  │ Diagnostics  │           │
│  └──────────────┘  └──────────────┘  └──────────────┘           │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

## API Group

All Prophet CRDs use the `aiops.prophet.io` API group:

```yaml
apiVersion: aiops.prophet.io/v1alpha1
kind: AnomalyAction
```

## Metrics

Each operator exposes Prometheus metrics on `:8080/metrics`. Available metrics vary by operator.

## Troubleshooting

### View Operator Logs

```bash
kubectl logs -n prophet-operators -l control-plane=controller-manager -f
```

### Check CRD Status

```bash
kubectl get healthchecks,budgetguards,costalerts,diagnosticremediations,labelenforcers -A
```

### Common Issues

| Issue | Solution |
|-------|----------|
| CRD not found | `make manifests && kubectl apply -f config/crd/bases/` |
| RBAC denied | Check ClusterRole/ClusterRoleBinding in operator manifest |
| Operator not reconciling | Check operator logs, verify webhook connectivity |
