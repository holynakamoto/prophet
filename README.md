# Prophet 🚀

## The AIOps-Powered Self-Healing Multi-Cloud Karpenter Platform

[![CI - Manifest Validation](https://github.com/prophet-aiops/prophet/actions/workflows/ci-validate.yaml/badge.svg)](https://github.com/prophet-aiops/prophet/actions/workflows/ci-validate.yaml)
[![CI - Operator Build](https://github.com/prophet-aiops/prophet/actions/workflows/ci-operator-build.yaml/badge.svg)](https://github.com/prophet-aiops/prophet/actions/workflows/ci-operator-build.yaml)

**Prophet** is an open-source GitOps monorepo that makes Kubernetes clusters across AWS, GCP, and Azure **predictive and self-healing**.

At its core are **custom Golang operators** that automatically detect issues, reason about them with AI, and remediate common failures without human intervention.

### Core Self-Healing & AIOps Capabilities

- **AnomalyRemediator Operator**  
  Detects anomalies via OpenTelemetry + Grafana ML → automatically restarts pods, scales deployments, or drains nodes.

- **PredictiveScaler Operator**  
  Uses Grafana ML forecasts to pre-provision Karpenter nodes before demand spikes.

- **SLOEnforcer Operator**  
  Monitors error budgets and takes action when SLOs are at risk.

- **AutonomousAgent Operator**  
  In-cluster LLM inference + Model Context Protocol (MCP) server for agentic autonomy: diagnose → decide → act.

- **K8sGPT Integration**  
  AI-powered plain-English diagnostics on every alert.

- **eBPF Observability**  
  Cilium + Hubble for kernel-level network visibility with zero overhead.

All managed declaratively through ArgoCD and Kustomize overlays.

### Quick Demo: Watch Self-Healing in Action (<5 minutes)

```bash
git clone https://github.com/prophet-aiops/prophet.git
cd prophet/demo/remediation-chaos
./demo.sh
```

See a crashing app get automatically detected and healed by our operators.

### Key Stack

- **Autoscaling**: Karpenter (multi-cloud)
- **GitOps**: ArgoCD + Kustomize
- **Observability**: OpenTelemetry → Prometheus → Grafana (with ML forecasting)
- **Diagnostics**: K8sGPT
- **Networking**: Cilium + Hubble (eBPF)
- **AI Reasoning**: Local LLM inference + MCP protocol
- **CI**: GitHub Actions (manifest validation, security scanning, operator builds)

### Getting Started

1. Clone and explore
2. Try the quick demo above
3. Deploy operators using Helm charts or GitOps overlays

### Testing K8sGPT Integration

See `aiops/diagnostics/K8SGPT-TESTING.md`.

### UI: Headlamp (Prophet AIOps Console)

Deploy Headlamp to browse Prophet CRDs and “trust-but-verify” self-healing actions:

- `clusters/common/ui/headlamp/README.md`

Great as a starter kit, learning resource, or foundation for production AIOps.

---

## This Repository

This is an active development fork focused on:

- **Custom Go operators** for self-healing automation (`operators/`)
- **Multi-cloud Kustomize overlays** for AWS, GCP, Azure (`clusters/`)
- **Chaos engineering experiments** with AI validation (`resilience/`, `demo/`)
- **Rancher & Headlamp UI extensions** for K8sGPT diagnostics (`rancher-k8sgpt-extension/`, `headlamp-k8sgpt/`)

See the operator-specific READMEs for current status and documentation.

---

## For Contributors

**Quick Links:**
- [Contributing Guide](./CONTRIBUTING.md) - Dev setup, PR guidelines
- [Operator Reference](./operators/README.md) - How to build and run operators
- [Cluster Overlays](./clusters/README.md) - GitOps structure and deployment

**Tech Stack:**
- Go 1.22+ (operators, controller-runtime)
- Kubebuilder v3.x (CRD scaffolding)
- Kustomize + ArgoCD (GitOps)
- Tilt (local operator development)

**Local Development:**
```bash
# Run all operators locally (requires kind cluster)
make dev-up

# Run specific operator
cd operators/label-enforcer && make run

# Run demo
cd demo/remediation-chaos && ./demo.sh
```

**Helm Deployment:**
```bash
# Add Helm repository (if using hosted charts)
# helm repo add prophet https://charts.prophet-aiops.dev
# helm repo update

# Install operators via Helm
helm install prophet-label-enforcer operators/label-enforcer/helm/label-enforcer
helm install prophet-health-check operators/health-check/helm/health-check

# Or use GitOps with ArgoCD/Flux pointing to chart directories
```

---

## Design Documents

| Document | Description |
|----------|-------------|
| [PRD-ALIGNMENT.md](./PRD-ALIGNMENT.md) | Requirements mapping |
| [AIOPS-UPGRADE.md](./AIOPS-UPGRADE.md) | V5 upgrade path |
| [V6-AGENTIC-AUTONOMY.md](./V6-AGENTIC-AUTONOMY.md) | Agentic architecture |

---

GitHub: `https://github.com/prophet-aiops/prophet`

Interested in self-healing operators or agentic Kubernetes? Let's connect!
