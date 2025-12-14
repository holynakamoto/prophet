# Cluster Configurations

GitOps-ready Kustomize overlays for deploying Prophet across AWS, GCP, and Azure.

## Directory Structure

```
clusters/
├── common/                     # Shared components (all clusters)
│   ├── aiops/
│   │   ├── operators/          # Prophet operator manifests
│   │   ├── k8sgpt/             # K8sGPT deployment
│   │   ├── grafana-ml/         # Grafana ML integration
│   │   ├── ai-agents/          # Autonomous agent configs
│   │   └── mcp/                # Model Context Protocol server
│   ├── network/
│   │   └── cilium.yaml         # Cilium CNI + Hubble
│   ├── opentelemetry/
│   │   └── collector.yaml      # OTel collector config
│   └── ui/
│       └── headlamp/           # Headlamp dashboard
│
├── aws/                        # AWS-specific
│   ├── base/
│   │   ├── karpenter/          # Karpenter autoscaler
│   │   ├── ingress/            # ALB ingress
│   │   └── kustomization.yaml
│   └── overlays/
│       ├── prod/               # Production overlay
│       └── staging/            # Staging overlay
│
├── gcp/                        # GCP-specific
│   ├── base/
│   └── overlays/
│       ├── prod/
│       └── staging/
│
├── azure/                      # Azure-specific
│   ├── base/
│   └── overlays/
│       ├── prod/
│       └── staging/
│
└── federation/                 # Multi-cluster management
    ├── management-cluster/     # Central control plane
    └── workload-clusters/      # Managed cluster configs
```

## Environments

| Path | Purpose | Typical Use |
|------|---------|-------------|
| `common/` | Shared across all clouds | Prophet operators, monitoring, UI |
| `*/base/` | Cloud-specific base configs | Karpenter, cloud ingress |
| `*/overlays/staging/` | Pre-production | Testing before prod |
| `*/overlays/prod/` | Production | Full self-healing enabled |

## Bootstrap with ArgoCD

### 1. Install ArgoCD

```bash
kubectl create namespace argocd
kubectl apply -n argocd -f https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml
```

### 2. Create Root Application

Create an ArgoCD Application pointing to this repo:

```yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: prophet-root
  namespace: argocd
spec:
  project: default
  source:
    repoURL: https://github.com/YOUR-ORG/prophet.git
    targetRevision: main
    path: apps/argo-apps
  destination:
    server: https://kubernetes.default.svc
    namespace: argocd
  syncPolicy:
    automated:
      prune: true
      selfHeal: true
```

Apply it:

```bash
kubectl apply -f prophet-root-app.yaml
```

### 3. ArgoCD Syncs the Stack

ArgoCD will sync:
1. `apps/argo-apps/root-app.yaml` → Syncs child apps
2. Child apps deploy operators, monitoring, etc.

## Manual Deployment (Without ArgoCD)

### Deploy Common Components

```bash
# Prophet operators
kubectl apply -f clusters/common/aiops/operators/

# Monitoring
kubectl apply -f monitoring/grafana/
kubectl apply -f monitoring/prometheus/

# UI
kubectl apply -f clusters/common/ui/headlamp/
```

### Deploy Cloud-Specific Components

```bash
# AWS example
kubectl apply -k clusters/aws/overlays/staging/

# GCP example
kubectl apply -k clusters/gcp/overlays/prod/
```

## Kustomize Overlays

Each overlay patches the base for environment-specific settings:

```yaml
# clusters/aws/overlays/prod/kustomization.yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
  - ../../base
  - ../../../common/aiops/operators

patches:
  - path: karpenter-patch.yaml
    target:
      kind: Provisioner
      name: default
```

### Preview Changes

```bash
# See what will be applied
kubectl kustomize clusters/aws/overlays/staging/

# Diff against cluster
kubectl diff -k clusters/aws/overlays/staging/
```

## Environment Variables

Some manifests use environment substitution. Set before applying:

```bash
export CLUSTER_NAME=prophet-prod
export AWS_REGION=us-west-2
export KARPENTER_VERSION=v0.32.0

envsubst < clusters/aws/base/karpenter/provisioner.yaml | kubectl apply -f -
```

## Adding a New Cluster

1. Create overlay directory:
   ```bash
   mkdir -p clusters/aws/overlays/new-env
   ```

2. Create `kustomization.yaml`:
   ```yaml
   apiVersion: kustomize.config.k8s.io/v1beta1
   kind: Kustomization
   resources:
     - ../../base
     - ../../../common/aiops/operators
   patches:
     - path: patches.yaml
   ```

3. Add any environment-specific patches

4. Test:
   ```bash
   kubectl kustomize clusters/aws/overlays/new-env/
   ```

5. Add to ArgoCD (or apply manually)

## Validation

```bash
# Validate all manifests
find clusters/ -name "*.yaml" -exec kubectl apply --dry-run=client -f {} \;

# Lint with kubeconform
kubeconform -kubernetes-version 1.29 clusters/common/aiops/operators/*.yaml
```

## See Also

- [operators/README.md](../operators/README.md) - Operator documentation
- [monitoring/](../monitoring/) - Prometheus, Grafana configs
- [demo/](../demo/) - Demo environment setup

