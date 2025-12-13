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

See a crashing app get automatically detected and healed by the AnomalyRemediator operator.

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
3. Deploy to your cluster using the provided overlays

Great as a starter kit, learning resource, or foundation for production AIOps.

GitHub: `https://github.com/prophet-aiops/prophet`

Interested in self-healing operators or agentic Kubernetes? Let's connect!
