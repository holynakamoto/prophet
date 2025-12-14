# Label Enforcer Operator

A simple Kubernetes operator that enforces required labels and annotations on resources, providing self-healing for configuration compliance.

## Overview

The LabelEnforcer operator ensures that Kubernetes resources always have required labels and annotations. This is useful for:

- **Security policies**: Ensure all pods have required security labels
- **Compliance**: Enforce organizational labeling standards
- **Automation**: Automatically apply annotations for monitoring, networking, etc.
- **Governance**: Prevent misconfiguration drift

## How It Works

1. Define a `LabelEnforcer` CR specifying which resources to watch and what labels/annotations are required
2. The operator reconciles by finding resources that don't have the required metadata
3. Missing labels/annotations are automatically added
4. Status shows how many resources were corrected

## Example Use Case: Security Labels

```yaml
apiVersion: aiops.prophet.io/v1alpha1
kind: LabelEnforcer
metadata:
  name: security-label-enforcer
  namespace: default
spec:
  targetResource: pods
  namespace: default
  requiredLabels:
    security.alpha.kubernetes.io/scc: restricted
    pod-security.kubernetes.io/enforce: restricted
  requiredAnnotations:
    security.alpha.kubernetes.io/validate: "true"
```

This ensures all pods in the `default` namespace have the required security labels and annotations.

## Supported Resources

- `pods` - Kubernetes Pods
- `deployments` - Kubernetes Deployments
- `services` - Kubernetes Services
- `configmaps` - Kubernetes ConfigMaps
- `secrets` - Kubernetes Secrets

## Installation

1. **Apply the CRD:**
   ```bash
   kubectl apply -f config/crd/bases/aiops.prophet.io_labelenforcers.yaml
   ```

2. **Run the operator:**
   ```bash
   make run
   ```

3. **Or build and deploy:**
   ```bash
   make build
   # Then deploy the built image to your cluster
   ```

## Usage Examples

### 1. Enforce Security Labels on All Pods

```yaml
apiVersion: aiops.prophet.io/v1alpha1
kind: LabelEnforcer
metadata:
  name: pod-security-enforcer
spec:
  targetResource: pods
  requiredLabels:
    pod-security.kubernetes.io/enforce: restricted
    pod-security.kubernetes.io/warn: restricted
```

### 2. Add Monitoring Annotations to Deployments

```yaml
apiVersion: aiops.prophet.io/v1alpha1
kind: LabelEnforcer
metadata:
  name: monitoring-annotations
spec:
  targetResource: deployments
  requiredAnnotations:
    prometheus.io/scrape: "true"
    prometheus.io/port: "8080"
```

### 3. Enforce Labels on Specific Resources

```yaml
apiVersion: aiops.prophet.io/v1alpha1
kind: LabelEnforcer
metadata:
  name: app-label-enforcer
spec:
  targetResource: pods
  labelSelector:
    app.kubernetes.io/name: my-app
  requiredLabels:
    team: platform
    environment: production
```

## Status and Monitoring

Check the status of your LabelEnforcer:

```bash
kubectl get labelenforcer security-label-enforcer -o yaml
```

The status shows:
- `correctedResources`: How many resources were fixed
- `lastCorrected`: When the last correction happened

## Development

### Prerequisites

- Go 1.24+
- Kubernetes cluster or kind
- kubectl configured

### Local Development

```bash
# Run locally against your cluster
make run

# Build the operator
make build

# Run tests
make test
```

### Building for Production

```bash
# Build Docker image
make docker-build IMG=your-registry/label-enforcer:v1.0.0

# Push to registry
make docker-push IMG=your-registry/label-enforcer:v1.0.0
```

## Architecture

The operator uses the standard Kubernetes controller pattern:

1. **Watch**: Monitors `LabelEnforcer` CRs and target resources
2. **Reconcile**: Compares current state vs. desired state
3. **Act**: Updates resources that are missing required metadata
4. **Report**: Updates status with correction counts

## RBAC Permissions

The operator requires permissions to:
- Get, list, watch `LabelEnforcer` resources
- Get, list, watch, update target resources (pods, deployments, etc.)
- Create events for logging

## Why This is "Low Hanging Fruit"

This operator demonstrates self-healing with:
- ✅ Simple reconciliation logic
- ✅ No external dependencies
- ✅ Immediate value for compliance
- ✅ Easy to understand and extend
- ✅ Builds on basic Kubernetes concepts
- ✅ Can be implemented in hours, not days

Start with enforcing security labels, then expand to monitoring annotations, team ownership labels, etc.