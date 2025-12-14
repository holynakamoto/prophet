# Helm-Based Operator Deployment Workflow

This document outlines the standardized workflow for building, packaging, and deploying Prophet operators as Helm charts, following the PRD requirements.

## Overview

All Prophet operators are packaged as Helm charts for:
- **Versioned releases** with semantic versioning
- **Configurable deployments** via `values.yaml`
- **Safe upgrades** with CRD handling
- **GitOps compatibility** with ArgoCD/Flux
- **Environment-specific customization**

## Chart Structure

Each operator chart follows this structure:

```
operators/<operator>/helm/<operator>/
├── Chart.yaml          # Chart metadata and versioning
├── values.yaml         # Default configuration values
├── crds/               # CustomResourceDefinitions (not templated)
│   └── aiops.prophet.io_<resources>.yaml
├── templates/          # Kubernetes manifests (templated)
│   ├── _helpers.tpl    # Template helpers
│   └── <operator>.yaml # Deployment, RBAC, ServiceAccount
└── .helmignore         # Files to ignore during packaging
```

## Configuration Values

### Required Values

```yaml
# Image configuration
image:
  repository: ghcr.io/prophet-aiops/prophet-<operator>
  tag: "latest"  # Should match chart appVersion
  pullPolicy: IfNotPresent

# Namespace to watch (empty = all namespaces)
watchNamespace: ""

# Feature flags
metrics:
  enabled: true

webhooks:
  enabled: false
```

### Operator-Specific Values

Additional values are available in each operator's `values.yaml` for:
- Resource limits and requests
- Node selectors and tolerations
- Service account configuration
- Controller-specific settings

## Development Workflow

### 1. Local Development

```bash
# Modify operator code
vim controllers/<operator>_controller.go

# Test locally
make run

# Build and push image
make docker-build docker-push IMG=ghcr.io/prophet-aiops/prophet-<operator>:dev
```

### 2. Chart Development

```bash
# Lint chart for issues
make helm-lint

# Show rendered templates
make helm-template

# Test installation (dry-run)
helm install <operator> ./helm/<operator> --dry-run
```

### 3. Release Process

```bash
# Update version in Chart.yaml (semantic versioning)
# appVersion should match image tag
version: 1.0.0
appVersion: "v1.0.0"

# Package chart
make helm-package

# Push to chart repository (OCI registry, ChartMuseum, etc.)
helm push <operator>-1.0.0.tgz oci://registry.example.com/charts
```

## Deployment Methods

### Method 1: Direct Helm Install

```bash
# Install with defaults
helm install prophet-label-enforcer operators/label-enforcer/helm/label-enforcer

# Install with custom values
helm install prophet-label-enforcer operators/label-enforcer/helm/label-enforcer \
  --set image.tag=v1.0.0 \
  --set watchNamespace=default \
  --set metrics.enabled=false

# Upgrade existing release
helm upgrade prophet-label-enforcer operators/label-enforcer/helm/label-enforcer
```

### Method 2: GitOps with ArgoCD

Create an ArgoCD Application:

```yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: prophet-label-enforcer
  namespace: argocd
spec:
  project: default
  source:
    repoURL: https://github.com/prophet-aiops/prophet
    targetRevision: main
    path: operators/label-enforcer/helm/label-enforcer
  destination:
    server: https://kubernetes.default.svc
    namespace: prophet-operators
  syncPolicy:
    automated:
      prune: true
      selfHeal: true
```

### Method 3: Flux

```yaml
apiVersion: helm.toolkit.fluxcd.io/v2beta1
kind: HelmRelease
metadata:
  name: prophet-label-enforcer
  namespace: prophet-operators
spec:
  interval: 5m
  chart:
    spec:
      chart: operators/label-enforcer/helm/label-enforcer
      version: "1.0.0"
      sourceRef:
        kind: GitRepository
        name: prophet
  values:
    image:
      tag: v1.0.0
    watchNamespace: default
```

## CRD Management

### Important: CRDs in `crds/` Directory

CRDs are placed in the `crds/` directory (not `templates/`) because:
- They are installed **only once** on initial install
- They are **never updated** during upgrades (prevents data loss)
- They follow Helm best practices for CRD handling

### Manual CRD Installation

If deploying without Helm:

```bash
# Install CRDs first (one-time)
kubectl apply -f operators/label-enforcer/helm/label-enforcer/crds/

# Then deploy the operator
kubectl apply -f operators/label-enforcer/helm/label-enforcer/templates/
```

## Versioning Strategy

### Semantic Versioning

- **Chart version**: Follows operator functionality changes
- **App version**: Matches the operator container image tag
- **Examples**:
  - `version: 1.0.0, appVersion: "v1.0.0"` - Major release
  - `version: 1.0.1, appVersion: "v1.0.0"` - Chart fix, same operator
  - `version: 1.1.0, appVersion: "v1.1.0"` - New features

### Release Tags

```bash
# Tag both chart and image with same version
git tag v1.0.0
docker build -t ghcr.io/prophet-aiops/prophet-label-enforcer:v1.0.0
helm package operators/label-enforcer/helm/label-enforcer
```

## Testing Strategy

### Local Testing

```bash
# Create kind cluster
kind create cluster --name prophet-test

# Install cert-manager (if webhooks enabled)
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.13.0/cert-manager.yaml

# Install operator
helm install prophet-label-enforcer operators/label-enforcer/helm/label-enforcer

# Verify installation
kubectl get pods -n prophet-operators
kubectl get crd | grep prophet
```

### CI/CD Integration

```yaml
# Example GitHub Actions
- name: Lint Helm charts
  run: |
    for chart in operators/*/helm/*/; do
      helm lint $chart
    done

- name: Package and push charts
  run: |
    for chart in operators/*/helm/*/; do
      helm package $chart
      helm push *.tgz oci://ghcr.io/prophet-aiops/charts
    done
```

## Troubleshooting

### Chart Installation Fails

```bash
# Check chart syntax
helm lint operators/label-enforcer/helm/label-enforcer

# Debug template rendering
helm template prophet-label-enforcer operators/label-enforcer/helm/label-enforcer --debug

# Check Kubernetes events
kubectl get events --sort-by=.metadata.creationTimestamp
```

### CRD Conflicts

```bash
# Never update CRDs during upgrade - this can cause data loss
# If CRD changes are needed, do manual migration

# Check CRD status
kubectl get crd labelenforcers.aiops.prophet.io -o yaml
```

### Image Pull Issues

```bash
# Verify image exists
docker pull ghcr.io/prophet-aiops/prophet-label-enforcer:v1.0.0

# Check imagePullSecrets if using private registry
kubectl create secret docker-registry regcred \
  --docker-server=ghcr.io \
  --docker-username=$GITHUB_USERNAME \
  --docker-password=$GITHUB_TOKEN
```

## Best Practices

### Chart Development
- Keep `values.yaml` well-documented with comments
- Use `_helpers.tpl` for common template functions
- Test charts with `helm template` before committing
- Include `.helmignore` to avoid packaging unnecessary files

### Deployment
- Use `helm upgrade --install` for idempotent deployments
- Set resource limits and requests appropriately
- Enable metrics collection for monitoring
- Use namespace isolation for multi-tenant deployments

### Security
- Run operators with minimal RBAC permissions
- Use image pull secrets for private registries
- Enable network policies for pod communication
- Regularly update base images and dependencies

## Migration from kubectl apply

### Before (kubectl apply)
```bash
kubectl apply -f clusters/common/aiops/operators/label-enforcer.yaml
```

### After (Helm)
```bash
# One-time setup
helm install prophet-label-enforcer operators/label-enforcer/helm/label-enforcer

# Upgrades
helm upgrade prophet-label-enforcer operators/label-enforcer/helm/label-enforcer

# Rollbacks
helm rollback prophet-label-enforcer 1

# Customization
helm upgrade prophet-label-enforcer operators/label-enforcer/helm/label-enforcer \
  --set image.tag=v1.1.0 \
  --set watchNamespace=production
```

This Helm-based approach provides production-ready operator deployments with proper versioning, configuration management, and upgrade safety.