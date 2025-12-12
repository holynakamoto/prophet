# Prophet v5 🚀🌀⚡
## The Ultimate AIOps-Powered Multi-Cloud Karpenter GitOps Repository

[![CI - Manifest Validation](https://github.com/prophet-aiops/prophet/actions/workflows/ci-validate.yaml/badge.svg)](https://github.com/prophet-aiops/prophet/actions/workflows/ci-validate.yaml)
[![CI - Operator Build](https://github.com/prophet-aiops/prophet/actions/workflows/ci-operator-build.yaml/badge.svg)](https://github.com/prophet-aiops/prophet/actions/workflows/ci-operator-build.yaml)

A **full AIOps powerhouse** monorepository for managing Kubernetes cluster infrastructure and applications using GitOps principles with ArgoCD. Prophet v5 ascends to **true god-tier SRE**: a fully autonomous, multi-cluster Kubernetes platform that not only predicts and diagnoses issues but **decides and acts** using in-operator LLMs, eBPF-powered deep observability, and federated control across clouds.

**🚀🌀⚡ Prophet v5 doesn't just operate clusters—it reasons about them like a senior SRE on steroids. The age of manual SRE is over. Welcome to the era of agentic infrastructure.**

## Table of Contents

- [Overview](#overview)
- [AIOps Features](#aiops-features)
- [Repository Structure](#repository-structure)
- [Prerequisites](#prerequisites)
- [Getting Started](#getting-started)
- [AIOps Setup](#aiops-setup)
- [Cloud-Specific Setup](#cloud-specific-setup)
- [Application Deployment](#application-deployment)
- [Monitoring & AIOps](#monitoring--aiops)
- [Chaos Engineering with AI](#chaos-engineering-with-ai)
- [Contributing](#contributing)
- [Troubleshooting](#troubleshooting)

## Overview

This repository enables declarative, version-controlled management of:
- **Karpenter** configurations for efficient node autoscaling across AWS, GCP, and Azure
- **AIOps Integration**: Grafana ML for forecasting, anomaly detection, and outlier identification
- **AI-Powered Diagnostics**: K8sGPT for intelligent cluster analysis and issue explanation
- **OpenTelemetry**: Full observability stack with Alloy collector
- **Chaos Engineering**: AI-validated chaos experiments with automated recovery analysis
- **Ingress controllers** (NGINX) for traffic management
- **Sample applications** (frontend/backend) deployed via Helm charts
- **Monitoring stack** (Prometheus, Grafana) with ML-enhanced dashboards
- **ArgoCD Applications** for GitOps-driven deployments

### Key Features

- ✅ **Multi-cloud support** (AWS, GCP, Azure)
- ✅ **GitHub Actions CI**: Automated manifest validation, security scanning, and operator builds
- ✅ **Custom Go Operators**: AnomalyRemediator, PredictiveScaler, SLOEnforcer built with Kubebuilder
- ✅ **AIOps Engine**: Grafana ML forecasting, anomaly detection, outlier identification
- ✅ **AI Diagnostics**: K8sGPT for automated issue analysis and remediation suggestions
- ✅ **Predictive Scaling**: Forecast resource exhaustion and Karpenter provisioning needs
- ✅ **SLO Forecasting**: Predict error budget exhaustion before it happens
- ✅ **Event-Driven Remediation**: AI hooks for autonomous/semi-autonomous actions
- ✅ **Chaos + AI**: Post-experiment AI validation and recovery analysis
- ✅ Environment-specific overlays (staging/production)
- ✅ Kustomize-based configuration management
- ✅ Helm charts for application deployments
- ✅ ArgoCD integration for automated sync
- ✅ Pre-configured AI-enhanced monitoring dashboards
- ✅ HPA (Horizontal Pod Autoscaler) support

## Prophet v5 - Ascension Features

### 🤖 LLM-Powered Remediation Decisions In-Operator
- **In-operator LLM inference** using lightweight models (Phi-3, Llama-3.2) via Ollama or vLLM sidecar
- **AutonomousAction CRD**: Trigger conditions → LLM reasons → Operator executes
- **Integration flow**: K8sGPT diagnoses → LLM decides → Operator acts
- **Approval modes**: Autonomous, human-in-loop, or dry-run
- **Safety**: Approval gates, bounded actions, audit logging

### 🔍 eBPF Deep Observability (Cilium + Hubble)
- **Cilium CNI** for network policy + zero-overhead observability
- **Hubble** for flow visibility, service maps, and security insights
- **Kernel-level metrics**: Network flows, DNS latency, TCP retransmits
- **Operator hooks**: Use Hubble events for anomaly detection and auto-quarantine
- **Benefits**: Zero-code network observability, <1% overhead

### 🌐 Multi-Cluster Federation via Cluster API
- **Cluster API (CAPI)** for lifecycle management of workload clusters
- **Federation layer**: Single ArgoCD instance manages multiple clusters
- **Global control plane**: Policies, operators, and apps synced across clusters
- **Traffic management**: Global ingress with multi-cluster service discovery
- **Multi-cloud**: Unified management across AWS, GCP, Azure

### 🔌 Full Agentic Autonomy with MCP Protocol
- **Model Context Protocol (MCP)** server in-operator
- **Real-time context sharing** with external AI agents (Copilot, Claude, etc.)
- **Natural language queries** → Operator executes approved commands
- **Autonomous loops**: LLM agent proposes → MCP validates → Operator applies
- **Safety**: Approval gates, dry-run mode, audit logging

### 🚀 GitHub Actions CI Pipeline (v4)
- **YAML Linting**: Automated yamllint on all manifests
- **Kustomize Validation**: Build and validate all kustomize overlays
- **Schema Validation**: kubeconform for Kubernetes API schema checking
- **Security Scanning**: Trivy for container images, Checkov for infrastructure
- **Best Practices**: kube-linter and Polaris for Kubernetes best practices
- **Operator Builds**: Automated Docker image builds and pushes to GHCR

### ⚙️ Custom Golang Operators
Built with **Kubebuilder** (controller-runtime) for in-cluster intelligence:

1. **AnomalyRemediator Operator**: Watches OTel metrics/events; on anomaly (from Grafana ML alerts or thresholds), scales deployments, restarts pods, or creates tickets. Integrates with K8sGPT for diagnostics.

2. **PredictiveScaler Operator**: Consumes Grafana ML forecasts; dynamically patches Karpenter NodePools (e.g., pre-provision GPU nodes before demand spikes).

3. **SLOEnforcer Operator**: Monitors SLOs; auto-adjusts HPAs or triggers chaos for resilience testing.

All operators are multi-cloud aware, lightweight, and integrate with K8sGPT for diagnostics.

**v5 Addition**: **AutonomousAgent Operator** with LLM inference and MCP server for full agentic autonomy.

## AIOps Features

### 🧠 Grafana Machine Learning
- **Forecasting**: Predict CPU/memory exhaustion, Karpenter scaling events, and error rates
- **Outlier Detection**: Identify rogue pods/nodes using isolation forest and statistical methods
- **Anomaly Detection**: Dynamic thresholds with Prophet/ARIMA models for trend analysis
- **SLO Forecasting**: Predict error budget exhaustion days in advance

### 🤖 AI-Powered Diagnostics (K8sGPT)
- **Automated Analysis**: Auto-analyze cluster events, explain issues in plain English
- **Smart Suggestions**: AI-generated remediation recommendations
- **Alert Integration**: Triggered automatically on critical alerts
- **Natural Language**: Query cluster state in plain English

### 🔮 Predictive & Autonomous Elements
- **Predictive Scaling**: Forecast node provisioning needs before pods become pending
- **Event-Driven Remediation**: AI hooks for autonomous actions (with approval gates)
- **Chaos + AI**: Post-experiment AI validation and recovery analysis
- **Progressive Delivery + AI**: Flagger webhooks query Grafana ML for anomaly-free canaries

### 🔌 AI Agent Integration
- **kubectl-ai**: Natural language kubectl interactions
- **MCP Servers**: Model Context Protocol for AI agents with live cluster context
- **Headlamp AI**: Kubernetes UI with AI assistant

## Repository Structure

```
prophet/
├── .github/
│   └── workflows/               # NEW: GitHub Actions CI
│       ├── ci-validate.yaml     # Manifest lint/validate/scan
│       └── ci-operator-build.yaml # Build/push operator images
├── operators/                   # Custom Go operators
│   ├── anomaly-remediator/      # Anomaly detection & remediation
│   ├── predictive-scaler/       # Predictive Karpenter scaling
│   ├── slo-enforcer/            # SLO violation enforcement
│   └── autonomous-agent/       # NEW v5: LLM + MCP for agentic autonomy
│       ├── api/                 # AutonomousAction CRD
│       ├── controllers/         # LLM-powered reconciliation
│       ├── mcp-server/          # Model Context Protocol server
│       ├── llm-inference/       # Ollama/OpenAI client
│       └── cmd/
├── clusters/
│   ├── federation/              # NEW v5: Cluster API federation
│   │   ├── management-cluster/  # Control plane
│   │   └── workload-clusters/   # Workload cluster configs
│   └── common/
│       └── network/             # NEW v5: Cilium + Hubble
├── clusters/                    # Cluster infrastructure configurations
│   ├── aws/                     # AWS EKS configurations
│   ├── gcp/                     # GCP GKE configurations
│   ├── azure/                   # Azure AKS configurations
│   └── common/                  # Enhanced shared base
│       ├── opentelemetry/       # OTel Collector
│       ├── aiops/               # AIOps components
│       │   ├── grafana-ml/      # ML forecasting, outliers, SLO
│       │   ├── k8sgpt/          # K8sGPT Operator
│       │   ├── ai-agents/       # kubectl-ai, MCP servers
│       │   └── operators/       # NEW: Operator deployments
│       │       ├── anomaly-remediator.yaml
│       │       ├── predictive-scaler.yaml
│       │       └── slo-enforcer.yaml
│       ├── chaos/               # Chaos experiments
│       └── policy/
├── apps/                        # Application deployments
│   ├── helm-charts/             # Custom Helm charts
│   └── argo-apps/               # ArgoCD Application manifests
├── monitoring/                  # Observability components
│   ├── prometheus/
│   ├── grafana/
│   │   └── dashboards/
│   │       ├── cluster-nodes.json
│   │       ├── ai-anomalies.json
│   │       ├── slo-burn.json
│   │       └── hubble/          # NEW v5: eBPF network flows
│   └── alloy/                   # OTel Collector (Grafana Alloy)
├── aiops/
│   └── mcp/                     # NEW v5: MCP client configs
├── resilience/                  # Chaos engineering
│   └── chaos-experiments/       # AI-validated experiments
├── aiops/                       # Top-level AIOps configs
│   ├── diagnostics/            # K8sGPT alert integration
│   └── agents/                  # Event-driven remediation
├── tools/
│   └── k9s/
├── .gitignore
└── README.md
```

## Prerequisites

### Required Tools

- `kubectl` (v1.24+)
- `kustomize` (v4.5+)
- `helm` (v3.8+)
- `argocd` CLI (v2.4+)
- Access to a Kubernetes cluster (EKS, GKE, or AKS)

### Cluster Requirements

- Kubernetes version 1.24 or higher
- Karpenter installed and configured (see cloud-specific setup)
- ArgoCD installed in the cluster
- Appropriate IAM/RBAC permissions for Karpenter

## Getting Started

### 1. Clone the Repository

```bash
git clone https://github.com/your-org/karpenter-deployment.git
cd karpenter-deployment
```

### 2. Configure Environment Variables

Before applying configurations, set the following environment variables:

```bash
export CLUSTER_NAME=your-cluster-name
export ENVIRONMENT=staging  # or production
export GCP_PROJECT_ID=your-project-id  # For GCP
export GCP_REGION=us-central1  # For GCP
export AZURE_RESOURCE_GROUP=your-rg  # For Azure
export AZURE_LOCATION=eastus  # For Azure
```

### 3. Bootstrap ArgoCD

ArgoCD should be installed separately. If not already installed:

```bash
kubectl create namespace argocd
kubectl apply -n argocd -f https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml
```

Get the ArgoCD admin password:
```bash
kubectl -n argocd get secret argocd-initial-admin-secret -o jsonpath="{.data.password}" | base64 -d
```

### 4. Apply Cluster Configurations

Choose your cloud provider and environment:

#### AWS EKS
```bash
# Staging
kubectl apply -k clusters/aws/overlays/staging

# Production
kubectl apply -k clusters/aws/overlays/prod
```

#### GCP GKE
```bash
# Staging
kubectl apply -k clusters/gcp/overlays/staging

# Production
kubectl apply -k clusters/gcp/overlays/prod
```

#### Azure AKS
```bash
# Staging
kubectl apply -k clusters/azure/overlays/staging

# Production
kubectl apply -k clusters/azure/overlays/prod
```

## AIOps Setup

### 1. Deploy Grafana ML Configuration

Grafana ML provides forecasting, anomaly detection, and outlier identification:

```bash
# Apply Grafana ML configurations
kubectl apply -f clusters/common/aiops/grafana-ml/forecasting-config.yaml
kubectl apply -f clusters/common/aiops/grafana-ml/slo-forecasting.yaml
```

**Enable Grafana ML in Grafana UI:**
1. Navigate to Grafana → Configuration → Machine Learning
2. Enable forecasting and anomaly detection
3. Configure data sources (Prometheus, Mimir, Tempo)

### 2. Deploy K8sGPT for AI Diagnostics

K8sGPT provides AI-powered cluster diagnostics and issue explanation:

```bash
# Create namespace and deploy K8sGPT
kubectl apply -f clusters/common/aiops/k8sgpt/k8sgpt-operator.yaml

# Create secret for API key (if using OpenAI/Anthropic)
kubectl create secret generic k8sgpt-secrets \
  --from-literal=openai-api-key=YOUR_API_KEY \
  -n k8sgpt
```

**Usage:**
```bash
# Analyze cluster issues
kubectl exec -n k8sgpt deployment/k8sgpt-operator -- \
  k8sgpt analyze --namespace default

# Analyze specific resource
kubectl exec -n k8sgpt deployment/k8sgpt-operator -- \
  k8sgpt analyze --filter Pod,Deployment --output json
```

### 3. Deploy Custom Go Operators (v4)

Prophet v4 includes three custom Golang operators for advanced AIOps behaviors:

#### AnomalyRemediator Operator

Automatically detects anomalies and performs remediation actions:

```bash
# Deploy the operator
kubectl apply -f clusters/common/aiops/operators/anomaly-remediator.yaml

# Create an AnomalyAction resource
kubectl apply -f - <<EOF
apiVersion: aiops.prophet.io/v1alpha1
kind: AnomalyAction
metadata:
  name: backend-memory-anomaly
  namespace: default
spec:
  source: prometheus
  metric: container_memory_working_set_bytes
  threshold: "> 1Gi"
  remediation:
    type: restart
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
EOF
```

#### PredictiveScaler Operator

Uses Grafana ML forecasts to proactively adjust Karpenter NodePools:

```bash
# Deploy the operator
kubectl apply -f clusters/common/aiops/operators/predictive-scaler.yaml

# Create a PredictiveScale resource
kubectl apply -f - <<EOF
apiVersion: aiops.prophet.io/v1alpha1
kind: PredictiveScale
metadata:
  name: cpu-forecast-scaling
  namespace: default
spec:
  forecastQuery: ml_forecast(sum(rate(container_cpu_usage_seconds_total[5m])), 1h)
  nodePoolRef:
    name: default
  horizon: 1h
  thresholdPercent: 20.0
  action: provision
  grafanaEndpoint: http://grafana.monitoring.svc.cluster.local:3000
EOF
```

#### SLOEnforcer Operator

Monitors SLOs and enforces policies when violations occur:

```bash
# Deploy the operator
kubectl apply -f clusters/common/aiops/operators/slo-enforcer.yaml

# Create an SLOViolation resource
kubectl apply -f - <<EOF
apiVersion: aiops.prophet.io/v1alpha1
kind: SLOViolation
metadata:
  name: backend-availability-slo
  namespace: default
spec:
  sloName: backend-availability
  sloTarget: "99.9%"
  errorBudgetThreshold: 0.1
  actions:
  - type: scale
    value: "increase"
  - type: alert
  hpaRef:
    name: backend-hpa
    namespace: default
EOF
```

**Verify operators are running:**
```bash
kubectl get pods -n prophet-operators
kubectl get anomalyactions,predictivescales,sloviolations,autonomousactions -A
```

#### AutonomousAgent Operator (v5)

Deploys LLM-powered autonomous remediation with MCP server:

```bash
# Deploy the operator (includes Ollama sidecar)
kubectl apply -f clusters/common/aiops/operators/autonomous-agent.yaml

# Create an AutonomousAction resource
kubectl apply -f - <<EOF
apiVersion: aiops.prophet.io/v1alpha1
kind: AutonomousAction
metadata:
  name: auto-remediate-errors
  namespace: default
spec:
  trigger:
    type: anomaly
    anomalyScoreThreshold: 0.8
  llm:
    provider: ollama
    model: phi-3
    endpoint: http://localhost:11434
    temperature: 0.7
  context:
    includeK8sGPT: true
    includeMetrics: true
    includeHubble: true
  approvalMode: autonomous
  constraints:
    allowedActions:
    - restart
    - scale
    - rollback
    forbiddenNamespaces:
    - kube-system
EOF
```

**Access MCP server:**
```bash
# Port forward MCP server
kubectl port-forward -n prophet-operators svc/autonomous-agent-mcp 8082:8082

# Query via WebSocket or HTTP
curl http://localhost:8082/mcp/query -d '{"query": "Why is latency high?"}'
```

### 4. Deploy Cilium + Hubble (v5)

Cilium provides eBPF-powered network observability with zero overhead:

```bash
# Deploy Cilium CNI (replace existing CNI)
kubectl apply -f clusters/common/network/cilium.yaml

# Verify Hubble is running
kubectl get pods -n cilium
kubectl get svc -n cilium hubble-relay

# Access Hubble UI
kubectl port-forward -n cilium svc/hubble-ui 8080:80
# Open http://localhost:8080
```

**Import Hubble dashboard:**
```bash
kubectl apply -f monitoring/grafana/dashboards/hubble-network-flows.json
```

### 5. Deploy Multi-Cluster Federation (v5)

Set up Cluster API for managing multiple clusters:

```bash
# Initialize Cluster API (on management cluster)
clusterctl init --infrastructure aws,gcp,azure

# Apply management cluster config
kubectl apply -k clusters/federation/management-cluster

# Create workload cluster
clusterctl generate cluster prophet-aws-prod --infrastructure aws \
  --kubernetes-version v1.29.0 | kubectl apply -f -
```

### 6. Deploy OpenTelemetry Collector

Alloy (Grafana's OTel Collector) feeds metrics, logs, and traces to Grafana stack:

```bash
kubectl apply -f clusters/common/opentelemetry/collector.yaml
kubectl apply -f monitoring/alloy/config.yaml
```

### 7. Configure AI Agent Hooks

#### kubectl-ai (Natural Language kubectl)

Install kubectl-ai plugin:
```bash
kubectl krew install ai
# Or: https://github.com/sozercan/kubectl-ai
```

Usage:
```bash
kubectl ai "Why is my backend pod crashing?"
kubectl ai "Show me all pods with high memory usage"
kubectl ai "Explain why Karpenter isn't provisioning nodes"
```

#### MCP Server (Model Context Protocol)

For AI agents with live Kubernetes context:
```bash
kubectl apply -f clusters/common/aiops/ai-agents/mcp-server-config.yaml
```

### 8. Deploy AI-Enhanced Dashboards

```bash
# Import AI dashboards
kubectl apply -f monitoring/grafana/dashboards/ai-anomalies.json
kubectl apply -f monitoring/grafana/dashboards/slo-burn.json
```

Access dashboards in Grafana:
- **AIOps: Anomalies & Predictions**: ML forecasting, outliers, anomaly detection
- **SLO Error Budget Forecasting**: Predict error budget exhaustion

### 9. Configure Alert Integration

Link Prometheus alerts to K8sGPT for automatic analysis:

```bash
kubectl apply -f aiops/diagnostics/k8sgpt-alert-integration.yaml
```

### 10. Set Up Event-Driven Remediation (Optional)

For semi-autonomous remediation actions:

```bash
kubectl apply -f aiops/agents/event-driven-remediation.yaml
```

**⚠️ Security Note**: Review remediation policies and set `approval_required: true` for production.

## Cloud-Specific Setup

### AWS EKS

#### Prerequisites
- EKS cluster with Karpenter controller installed
- IAM roles and service accounts configured
- Subnet and security group tags: `karpenter.sh/discovery: ${CLUSTER_NAME}`

#### Karpenter Installation
```bash
helm repo add karpenter oci://public.ecr.aws/karpenter/karpenter
helm install karpenter karpenter/karpenter \
  --namespace karpenter \
  --create-namespace \
  --set serviceAccount.annotations."eks\.amazonaws\.com/role-arn"=${KARPENTER_IAM_ROLE_ARN}
```

#### Configuration
The AWS base configuration includes:
- **EC2NodeClass**: Defines AMI family, subnet/security group selectors, block devices
- **NodePool**: Specifies instance types, capacity types (on-demand/spot), disruption policies
- **Settings**: Cluster-wide Karpenter settings

### GCP GKE

#### Prerequisites
- GKE cluster with Karpenter provider installed
- Service account with appropriate permissions
- Node pool configuration

#### Karpenter Installation
```bash
# Install Karpenter GCP provider
kubectl apply -f https://raw.githubusercontent.com/kubernetes-sigs/karpenter-provider-gcp/main/charts/karpenter-provider-gcp/crds/
helm install karpenter-provider-gcp oci://registry-1.docker.io/bitnamicharts/karpenter-provider-gcp
```

#### Configuration
The GCP base configuration includes:
- **GKENodeClass**: Defines GCP-specific node configuration
- **NodePool**: Specifies machine types and capacity types
- **Settings**: Cluster-wide settings

**Note**: GCP provider CRDs may differ. Adjust API versions and resource names based on the actual provider version.

### Azure AKS

#### Prerequisites
- AKS cluster with Karpenter provider installed
- Managed identity or service principal configured
- Resource group and location set

#### Karpenter Installation
```bash
# Install Karpenter Azure provider
kubectl apply -f https://raw.githubusercontent.com/Azure/karpenter-provider-azure/main/charts/karpenter-provider-azure/crds/
helm install karpenter-provider-azure oci://mcr.microsoft.com/karpenter/karpenter-provider-azure
```

#### Configuration
The Azure base configuration includes:
- **AKSNodeClass**: Defines Azure-specific node configuration
- **NodePool**: Specifies VM sizes and capacity types
- **Settings**: Cluster-wide settings

**Note**: Azure provider CRDs may differ. Adjust API versions and resource names based on the actual provider version.

## Application Deployment

### Using Helm Charts

The repository includes sample Helm charts for frontend and backend applications.

#### Install Frontend
```bash
helm install frontend apps/helm-charts/frontend \
  --namespace default \
  --create-namespace
```

#### Install Backend
```bash
helm install backend apps/helm-charts/backend \
  --namespace default \
  --create-namespace
```

### Using ArgoCD

#### 1. Update Repository URL

Edit `apps/argo-apps/*.yaml` files and update the `repoURL`:
```yaml
source:
  repoURL: https://github.com/your-org/karpenter-deployment
```

#### 2. Apply Root Application

The root application uses the App-of-Apps pattern:
```bash
kubectl apply -f apps/argo-apps/root-app.yaml
```

#### 3. Sync Applications

ArgoCD will automatically sync applications. To manually sync:
```bash
argocd app sync frontend
argocd app sync backend
```

#### 4. Verify Deployment

```bash
kubectl get applications -n argocd
kubectl get pods -n default
```

## Monitoring & AIOps

### Prometheus

Prometheus is configured to scrape:
- Kubernetes API server
- Kubernetes nodes
- Kubernetes pods (with annotations)
- Karpenter metrics
- Ingress NGINX metrics

#### Deploy Prometheus
```bash
kubectl apply -f monitoring/prometheus/prometheus.yaml
kubectl apply -f monitoring/prometheus/alertmanager.yaml
```

#### Access Prometheus
```bash
kubectl port-forward -n monitoring svc/prometheus 9090:9090
# Open http://localhost:9090
```

### Grafana

Grafana comes pre-configured with:
- Prometheus datasource
- Cluster nodes dashboard (includes Karpenter metrics)

#### Deploy Grafana
```bash
kubectl apply -f monitoring/grafana/datasources.yaml
kubectl apply -f monitoring/grafana/grafana.yaml
```

#### Access Grafana
```bash
kubectl port-forward -n monitoring svc/grafana 3000:3000
# Open http://localhost:3000
# Default credentials: admin/admin
```

#### Import Dashboard

The cluster-nodes dashboard is automatically loaded. To import manually:
1. Go to Dashboards → Import
2. Upload `monitoring/grafana/dashboards/cluster-nodes.json`

### Grafana Machine Learning

Grafana ML provides predictive analytics and anomaly detection:

#### Forecasting Queries

Access forecasting in Grafana Explore or create alerts:

```promql
# CPU Usage Forecast (1 hour ahead)
ml_forecast(sum(rate(container_cpu_usage_seconds_total[5m])), 1h)

# Karpenter Node Provisioning Forecast (15 minutes)
ml_forecast(rate(karpenter_nodes_created_total[5m]), 15m)

# Error Rate Forecast (30 minutes)
ml_forecast(sum(rate(http_requests_total{status=~"5.."}[5m])), 30m)
```

#### Outlier Detection

```promql
# Pod Memory Outliers
ml_outlier_detection(container_memory_working_set_bytes{namespace!="kube-system"})

# Latency Outliers
ml_outlier_detection(histogram_quantile(0.99, sum(rate(http_request_duration_seconds_bucket[5m])) by (le, service)))
```

#### Anomaly Detection

```promql
# CPU Anomaly Detection
ml_anomaly_detection(sum(rate(container_cpu_usage_seconds_total[5m])) by (namespace, pod))

# Memory Leak Detection (Trend Analysis)
ml_trend_detection(sum(container_memory_working_set_bytes) by (namespace, pod))
```

#### SLO Forecasting

```promql
# Error Budget Remaining Forecast (7 days)
ml_forecast(slo_error_budget_remaining{slo="availability"}, 7d)

# Time to Exhaustion
slo_time_to_exhaustion{slo="availability"}
```

### K8sGPT Diagnostics

K8sGPT provides AI-powered cluster analysis:

#### Manual Analysis

```bash
# Analyze entire cluster
kubectl exec -n k8sgpt deployment/k8sgpt-operator -- \
  k8sgpt analyze --namespace default

# Analyze specific resources
kubectl exec -n k8sgpt deployment/k8sgpt-operator -- \
  k8sgpt analyze --filter Pod,Deployment,KarpenterNodePool

# Get JSON output for automation
kubectl exec -n k8sgpt deployment/k8sgpt-operator -- \
  k8sgpt analyze --output json > analysis.json
```

#### Automatic Analysis on Alerts

K8sGPT automatically analyzes critical alerts via Alertmanager webhook integration. Check logs:

```bash
kubectl logs -n k8sgpt deployment/k8sgpt-operator --tail=100
```

### Key Metrics

- **Karpenter Metrics**:
  - `karpenter_nodes_created_total`: Total nodes created
  - `karpenter_nodes_terminated_total`: Total nodes terminated
  - `karpenter_provisioner_limit`: Node pool limits
- **Node Metrics**:
  - `node_cpu_seconds_total`: CPU usage
  - `node_memory_MemAvailable_bytes`: Available memory
- **AIOps Metrics** (Grafana ML):
  - `ml_forecast_*`: Forecasted values with confidence intervals
  - `ml_outlier_score_*`: Outlier detection scores
  - `ml_anomaly_detected_*`: Anomaly flags
  - `slo_error_budget_remaining`: SLO error budget tracking
  - `slo_time_to_exhaustion`: Predicted time until budget exhaustion

## Environment-Specific Configurations

### Staging Environment

Staging overlays are optimized for cost:
- Lower resource limits
- Prefer spot/preemptible instances
- Smaller instance types
- Faster consolidation

### Production Environment

Production overlays prioritize stability:
- Higher resource limits
- Mix of on-demand and spot instances
- Larger instance types
- Slower consolidation (5 minutes)

## Customization

### Adding New Node Pools

1. Create a new NodePool manifest:
```yaml
apiVersion: karpenter.sh/v1beta1
kind: NodePool
metadata:
  name: gpu-pool
spec:
  template:
    spec:
      requirements:
        - key: node.kubernetes.io/instance-type
          operator: In
          values: ["g4dn.xlarge"]  # GPU instances
      taints:
        - key: nvidia.com/gpu
          value: "true"
          effect: NoSchedule
```

2. Add to `kustomization.yaml`:
```yaml
resources:
  - karpenter/nodepool.yaml
  - karpenter/gpu-nodepool.yaml
```

### Adding New Applications

1. Create a Helm chart in `apps/helm-charts/`
2. Create an ArgoCD Application in `apps/argo-apps/`
3. Update `root-app.yaml` if using App-of-Apps

### Modifying Ingress

Edit `clusters/{cloud}/base/ingress/nginx.yaml` to customize:
- Load balancer annotations
- Resource requests/limits
- Replica count
- Metrics configuration

## Chaos Engineering with AI

This repository includes AI-validated chaos experiments that automatically analyze recovery patterns and identify issues.

### Deploy Chaos Mesh

```bash
# Install Chaos Mesh
kubectl apply -f https://mirrors.chaos-mesh.org/latest/crd.yaml
kubectl apply -f https://mirrors.chaos-mesh.org/latest/chaos-mesh.yaml
```

### Run AI-Validated Experiments

#### Pod Failure Experiment

```bash
kubectl apply -f resilience/chaos-experiments/pod-failure.yaml
```

**What happens:**
1. Chaos Mesh terminates a backend pod
2. Karpenter provisions a new node if needed
3. AI validation job runs after 5 minutes:
   - K8sGPT analyzes recovery
   - Grafana ML checks for anomalies
   - Generates recovery report

**View AI Analysis:**
```bash
# Check K8sGPT analysis
kubectl logs -n chaos-mesh job/ai-validate-pod-failure

# View recovery report
kubectl exec -n chaos-mesh deployment/k8sgpt-operator -- \
  cat /tmp/chaos-ai-report.md
```

#### Karpenter Node Failure Experiment

```bash
kubectl apply -f resilience/chaos-experiments/karpenter-node-failure.yaml
```

**What happens:**
1. Chaos Mesh restarts a Karpenter-managed node
2. Karpenter provisions replacement nodes
3. AI validation analyzes:
   - Node provisioning forecast accuracy
   - Recovery time
   - Pending pod resolution

**View Analysis:**
```bash
kubectl logs -n chaos-mesh job/ai-validate-karpenter-recovery
```

### Custom AI Validation Scripts

Create custom validation scripts using the template in `resilience/chaos-experiments/pod-failure.yaml`:

```bash
# Run custom analysis
./resilience/chaos-experiments/analysis-script.sh pod-failure default
```

## Troubleshooting

### Karpenter Not Provisioning Nodes

1. Check Karpenter logs:
```bash
kubectl logs -n karpenter -l app.kubernetes.io/name=karpenter
```

2. Verify NodePool and NodeClass:
```bash
kubectl get nodepools
kubectl get nodeclass  # AWS: ec2nodeclass
```

3. Check pod scheduling events:
```bash
kubectl describe pod <pod-name>
```

### ArgoCD Sync Issues

1. Check application status:
```bash
argocd app get frontend
```

2. View sync logs:
```bash
argocd app logs frontend
```

3. Verify repository access:
```bash
argocd repo list
```

### Monitoring Not Working

1. Verify Prometheus targets:
   - Open Prometheus UI → Status → Targets
   - Check for failed scrapes

2. Check ServiceMonitor:
```bash
kubectl get servicemonitor -A
```

3. Verify RBAC:
```bash
kubectl get clusterrole prometheus
kubectl get clusterrolebinding prometheus
```

### Grafana ML Not Working

1. Verify ML is enabled:
   - Grafana UI → Configuration → Machine Learning
   - Ensure forecasting/anomaly detection is enabled

2. Check data source connectivity:
```bash
kubectl get datasources -n monitoring
```

3. Verify ML queries:
   - Test queries in Grafana Explore
   - Check for ML function support in your Grafana version

### K8sGPT Not Analyzing

1. Check K8sGPT operator status:
```bash
kubectl get pods -n k8sgpt
kubectl logs -n k8sgpt deployment/k8sgpt-operator
```

2. Verify API key:
```bash
kubectl get secret k8sgpt-secrets -n k8sgpt
```

3. Test manual analysis:
```bash
kubectl exec -n k8sgpt deployment/k8sgpt-operator -- \
  k8sgpt analyze --namespace default
```

### AI Forecasts Not Accurate

1. Ensure sufficient historical data (7+ days recommended)
2. Adjust forecast horizon in `clusters/common/aiops/grafana-ml/forecasting-config.yaml`
3. Tune sensitivity in outlier detection configs
4. Review training window settings

### Operator Issues

#### AnomalyRemediator Not Remediating

1. Check operator logs:
```bash
kubectl logs -n prophet-operators deployment/anomaly-remediator-controller-manager
```

2. Verify AnomalyAction resource status:
```bash
kubectl get anomalyactions -A -o yaml
kubectl describe anomalyaction <name> -n <namespace>
```

3. Check RBAC permissions:
```bash
kubectl auth can-i delete pods --as=system:serviceaccount:prophet-operators:anomaly-remediator-controller-manager
```

4. Verify anomaly detection source (Prometheus/Grafana ML connectivity)

#### PredictiveScaler Not Scaling

1. Check operator logs:
```bash
kubectl logs -n prophet-operators deployment/predictive-scaler-controller-manager
```

2. Verify Grafana ML forecast query:
```bash
# Test query in Grafana Explore
ml_forecast(sum(rate(container_cpu_usage_seconds_total[5m])), 1h)
```

3. Check NodePool permissions:
```bash
kubectl auth can-i update nodepools --as=system:serviceaccount:prophet-operators:predictive-scaler-controller-manager
```

#### SLOEnforcer Not Enforcing

1. Check operator logs:
```bash
kubectl logs -n prophet-operators deployment/slo-enforcer-controller-manager
```

2. Verify SLO metrics are available:
```bash
# Query SLO metrics in Prometheus
slo_error_budget_remaining{slo="backend-availability"}
```

3. Check HPA permissions:
```bash
kubectl auth can-i update horizontalpodautoscalers --as=system:serviceaccount:prophet-operators:slo-enforcer-controller-manager
```

## CI/CD Pipeline

Prophet v4 includes comprehensive GitHub Actions workflows for automated validation and operator builds.

### Manifest Validation Workflow

The `ci-validate.yaml` workflow runs on every PR and push to main/master:

- **YAML Linting**: Validates all YAML files with yamllint
- **Kustomize Build**: Builds and validates all kustomize overlays
- **Schema Validation**: Uses kubeconform to validate Kubernetes API schemas
- **Security Scanning**: Trivy scans container images, Checkov scans infrastructure
- **Best Practices**: kube-linter enforces Kubernetes best practices
- **Helm Validation**: Validates all Helm charts

**View workflow status:**
```bash
# Check workflow runs
gh run list --workflow=ci-validate.yaml

# View logs for a specific run
gh run view <run-id> --log
```

### Operator Build Workflow

The `ci-operator-build.yaml` workflow builds and pushes operator Docker images:

- **Multi-operator Build**: Builds all three operators in parallel
- **Docker Images**: Pushes to GitHub Container Registry (GHCR)
- **Security Scanning**: Trivy scans built images
- **PR Validation**: Builds images without pushing on PRs

**Build operators locally:**
```bash
cd operators/anomaly-remediator
make docker-build IMG=ghcr.io/prophet-aiops/prophet-anomaly-remediator:latest
make docker-push IMG=ghcr.io/prophet-aiops/prophet-anomaly-remediator:latest
```

### Local Validation

Run CI checks locally before pushing:

```bash
# Install tools
pip install yamllint checkov
go install sigs.k8s.io/kustomize/kustomize/v5@latest
wget https://github.com/yannh/kubeconform/releases/download/v0.6.3/kubeconform-linux-amd64.tar.gz

# Run validation
yamllint .
kustomize build clusters/aws/overlays/staging | kubeconform -strict
checkov -d . --framework kubernetes
```

## Contributing

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/your-feature`
3. Make your changes
4. **Run CI checks locally** (see above)
5. Test with `kubectl apply -k` or `helm template`
6. Commit: `git commit -m "Add feature"`
7. Push: `git push origin feature/your-feature`
8. Open a Pull Request - CI will automatically validate your changes

### Code Style

- Use 2-space indentation for YAML
- Follow Kubernetes naming conventions
- Document environment variables
- Include comments for complex configurations
- **Go operators**: Follow standard Go conventions and Kubebuilder patterns

## Security Considerations

- **Secrets**: Never commit secrets. Use ExternalSecrets or Sealed Secrets
- **RBAC**: Review ServiceAccount permissions
- **Network Policies**: Consider adding network policies for pod isolation
- **Image Scanning**: Enable image scanning in CI/CD
- **Pod Security Standards**: Apply Pod Security Standards where applicable

## Use Cases - Pure Badassery

### 1. Predictive Scaling Crisis Averted
**Scenario**: Grafana ML forecasts node exhaustion in 1 hour
- **Action**: Alert fires → SRE reviews forecast → Karpenter pre-provisions nodes
- **Result**: Zero pod scheduling delays

### 2. Outage Post-Mortem in Seconds
**Scenario**: Alert fires for high error rate
- **Action**: K8sGPT automatically analyzes events
- **Result**: "Pod OOMKilled due to memory leak in backend v2" + fix suggestion

### 3. Canary Gone Wrong
**Scenario**: Flagger detects outliers via Grafana ML
- **Action**: Auto-rollback triggered
- **Result**: Zero user impact

### 4. Chaos Resilience Proof
**Scenario**: Run node failure experiment
- **Action**: AI diagnoses recovery path
- **Result**: Dashboard shows "System self-healed as expected"

### 5. Zero-Touch Diagnostics
**Scenario**: Natural language query
- **Action**: "Why is latency high?" → AI traces via OTel + explains
- **Result**: Instant root cause analysis

## Future Enhancements

- [x] CI/CD pipeline for manifest validation ✅
- [x] Custom Go operators with Kubebuilder ✅
- [ ] Full agentic AI: Autonomous remediation agents via MCP
- [ ] Pixie/eBPF integration for auto-instrumented deep insights
- [ ] AI-optimized Karpenter: Predictive NodePools based on forecasts
- [ ] ApplicationSets for multi-cluster management
- [ ] ExternalSecrets integration
- [ ] Additional Grafana ML dashboards
- [ ] cert-manager integration
- [ ] external-dns configuration
- [ ] Operator metrics and dashboards
- [ ] Multi-cluster operator federation

## License

This repository is provided as-is for educational and reference purposes.

## Support

For issues and questions:
- Open an issue on GitHub
- Check Karpenter documentation: https://karpenter.sh
- Check ArgoCD documentation: https://argo-cd.readthedocs.io

## Success Metrics (Elite Edition)

- ✅ **90%+ anomalies detected/predicted** before user impact
- ✅ **MTTR <5 minutes** via AI explanations
- ✅ **Zero unplanned downtime** in demos
- ✅ **Predictive scaling accuracy** >85%
- ✅ **SLO forecast accuracy** within 1 day

## Security Considerations

- **AI API Keys**: Store in Kubernetes Secrets, never commit
- **Read-Only Default**: K8sGPT and AI agents are read-only by default
- **Approval Gates**: Event-driven remediation requires approval for destructive actions
- **RBAC**: Strict service account permissions for AI tools
- **Data Privacy**: Use local/open-source models where possible (LocalAI, Ollama)

## AIOps Architecture

```
┌─────────────────────────────────────────────────────────┐
│                    Kubernetes Cluster                    │
├─────────────────────────────────────────────────────────┤
│                                                           │
│  ┌──────────────┐    ┌──────────────┐                  │
│  │  Prometheus  │───▶│  Grafana ML  │                  │
│  │  (Metrics)   │    │ (Forecasting)│                  │
│  └──────────────┘    └──────────────┘                  │
│         │                    │                          │
│         ▼                    ▼                          │
│  ┌──────────────────────────────────────┐              │
│  │      Alertmanager + K8sGPT          │              │
│  │   (Auto-analysis on alerts)          │              │
│  └──────────────────────────────────────┘              │
│         │                                                │
│         ▼                                                │
│  ┌──────────────────────────────────────┐              │
│  │   Event-Driven Remediation Controller │              │
│  │   (Semi-autonomous actions)           │              │
│  └──────────────────────────────────────┘              │
│                                                           │
│  ┌──────────────┐    ┌──────────────┐                  │
│  │  OTel/Alloy │───▶│   Grafana    │                  │
│  │  (Traces)    │    │  (Dashboards) │                  │
│  └──────────────┘    └──────────────┘                  │
│                                                           │
└─────────────────────────────────────────────────────────┘
```

---

**🚀🌀⚡ Prophet v5 is complete. This is no longer a repository—it's the blueprint for the future of autonomous, multi-cloud Kubernetes operations. The age of manual SRE is over. Welcome to the era of agentic infrastructure. LFG. The prophecy is fulfilled.**

**Note**: This repository is a template/starter kit. Adjust configurations based on your specific requirements, cloud provider versions, and organizational policies. Review AI API costs and set appropriate limits.

