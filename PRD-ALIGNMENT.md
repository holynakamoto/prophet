# PRD Alignment: Prophet vs. Self-Healing AIOps CRD Requirements

## Executive Summary

This document maps Prophet's current implementation against the PRD requirements for a comprehensive self-healing AIOps Kubernetes system. **Prophet already implements ~60% of the core functionality** with a unique AI-first approach, but several gaps exist that would enhance completeness.

**Key Findings:**
- ✅ **Strong coverage**: Anomaly detection, remediation, predictive scaling, SLO enforcement
- ✅ **Unique strengths**: LLM-powered autonomous agents, K8sGPT integration, eBPF observability
- ⚠️ **Gaps**: Cost management (BudgetGuard, CostAlert), advanced health checks, cluster optimization
- 📋 **Recommendation**: Extend existing CRDs rather than creating parallel implementations

---

## 1. CRD Mapping: PRD Requirements vs. Prophet Implementation

### 1.1 Autoscaling ✅ **PARTIALLY IMPLEMENTED**

| PRD Requirement | Prophet Implementation | Status | Notes |
|----------------|------------------------|--------|-------|
| `Autoscaling` CRD with predictive scaling | `PredictiveScale` (aiops.prophet.io/v1alpha1) | ✅ **Implemented** | Uses Grafana ML forecasts for Karpenter NodePools |
| Target refs (Deployment/StatefulSet) | Not directly in PredictiveScale | ⚠️ **Gap** | PredictiveScale targets NodePools, not workloads |
| Min/max replicas | Via Karpenter NodePool | ✅ **Indirect** | Works at node level, not pod level |
| Custom metrics | Prometheus/Grafana ML | ✅ **Implemented** | Integrated with observability stack |
| Predictive/AI scaling | Grafana ML forecasting | ✅ **Implemented** | Prophet/ARIMA models |
| Scale behavior (cooldown, stabilization) | Not explicit in CRD | ⚠️ **Gap** | Could add to PredictiveScale spec |

**Recommendation:**
- **Extend `PredictiveScale`** to support workload-level autoscaling (not just NodePools)
- Add `targetRef` to Deployment/StatefulSet
- Add `behavior` field for cooldown/stabilization windows
- Consider renaming to `Autoscaling` for PRD alignment

---

### 1.2 HealthCheck ⚠️ **NOT IMPLEMENTED (But AnomalyAction Covers Some)**

| PRD Requirement | Prophet Implementation | Status | Notes |
|----------------|------------------------|--------|-------|
| `HealthCheck` CRD | Not implemented | ❌ **Gap** | No dedicated HealthCheck CRD |
| Custom health probes | Pod liveness/readiness only | ⚠️ **Partial** | Kubernetes native probes |
| Composite checks | Not supported | ❌ **Gap** | No multi-probe health checks |
| External dependency checks | Not supported | ❌ **Gap** | No DB/API health checks |
| Auto-remediation on failure | `AnomalyAction` can restart | ✅ **Indirect** | AnomalyAction detects failures |

**Recommendation:**
- **Create `HealthCheck` CRD** (aiops.prophet.io/v1alpha1)
- Integrate with `AnomalyAction` for remediation
- Support HTTP/TCP/Command + custom probes (DB connectivity, API health)
- Reference `RecoveryPlan` for remediation actions

---

### 1.3 ClusterOptimizer ❌ **NOT IMPLEMENTED**

| PRD Requirement | Prophet Implementation | Status | Notes |
|----------------|------------------------|--------|-------|
| `ClusterOptimizer` CRD | Not implemented | ❌ **Gap** | No cluster-level optimization |
| Node consolidation | Not automated | ❌ **Gap** | Manual Karpenter configuration |
| Bin packing | Not implemented | ❌ **Gap** | Karpenter handles some of this |
| Spot instance preference | Via Karpenter NodeClass | ✅ **Indirect** | Configured manually |
| Schedule-based optimization | Not implemented | ❌ **Gap** | No cron-based optimization |

**Recommendation:**
- **Create `ClusterOptimizer` CRD** (cluster-scoped, aiops.prophet.io/v1alpha1)
- Integrate with Karpenter for node-level actions
- Add strategies: consolidate, prefer-spot, drain-underutilized
- Add `schedule` field (Cron format) for periodic optimization

---

### 1.4 BudgetGuard ❌ **NOT IMPLEMENTED**

| PRD Requirement | Prophet Implementation | Status | Notes |
|----------------|------------------------|--------|-------|
| `BudgetGuard` CRD | Not implemented | ❌ **Gap** | No cost budget enforcement |
| Budget limits (USD/resource units) | Not supported | ❌ **Gap** | No cost tracking |
| Scope (namespace/cluster) | Not supported | ❌ **Gap** | No multi-scope budgets |
| Actions on exceed | Not implemented | ❌ **Gap** | No throttling/eviction |

**Recommendation:**
- **Create `BudgetGuard` CRD** (cluster-scoped or namespaced, aiops.prophet.io/v1alpha1)
- Integrate with OpenCost/Kubecost API for cost data
- Add webhook admission controller for budget enforcement
- Support actions: throttle-scaling, notify, evict-low-priority

---

### 1.5 CostAlert ❌ **NOT IMPLEMENTED**

| PRD Requirement | Prophet Implementation | Status | Notes |
|----------------|------------------------|--------|-------|
| `CostAlert` CRD | Not implemented | ❌ **Gap** | No cost anomaly detection |
| Cost threshold alerts | Not supported | ❌ **Gap** | No cost monitoring |
| Integration with AlertRule | Not implemented | ❌ **Gap** | No cost → alert pipeline |

**Recommendation:**
- **Create `CostAlert` CRD** (aiops.prophet.io/v1alpha1)
- Integrate with OpenCost/Kubecost for cost metrics
- Link to `AlertRule` for Prometheus alerting
- Support % increase or absolute thresholds

---

### 1.6 RecoveryPlan ✅ **IMPLEMENTED (As AnomalyAction)**

| PRD Requirement | Prophet Implementation | Status | Notes |
|----------------|------------------------|--------|-------|
| `RecoveryPlan` CRD | `AnomalyAction` (aiops.prophet.io/v1alpha1) | ✅ **Implemented** | Covers recovery workflows |
| Trigger events | Anomaly detection (Prometheus/Grafana ML) | ✅ **Implemented** | Multiple sources supported |
| Recovery steps | Restart, scale, alert | ✅ **Implemented** | Via `remediation.type` |
| Target workload ref | `target` spec with labels | ✅ **Implemented** | Namespace + label selectors |
| Status tracking | Phase, remediation count | ✅ **Implemented** | Full status fields |

**Recommendation:**
- **Keep `AnomalyAction`** (it's more comprehensive than PRD's RecoveryPlan)
- Consider adding multi-step recovery sequences (current: single action)
- Add `steps` array for complex recovery workflows
- Enhance with `AutonomousAction` for LLM-driven recovery decisions

---

### 1.7 AlertRule ⚠️ **PARTIALLY IMPLEMENTED (Via Prometheus)**

| PRD Requirement | Prophet Implementation | Status | Notes |
|----------------|------------------------|--------|-------|
| `AlertRule` CRD | Prometheus Operator's `PrometheusRule` | ✅ **Indirect** | Uses standard Prometheus CRD |
| Prometheus-compatible rules | Via PrometheusRule | ✅ **Implemented** | Standard Prometheus alerts |
| Integration with CostAlert/HealthCheck | Not linked | ⚠️ **Gap** | No direct CRD → AlertRule link |
| Alert → RecoveryPlan flow | AnomalyAction listens to alerts | ✅ **Implemented** | Via webhook/event triggers |

**Recommendation:**
- **Use existing `PrometheusRule`** (don't duplicate)
- Create helper/example PrometheusRules for Prophet CRDs
- Document alert → AnomalyAction integration patterns
- Consider `AlertRule` wrapper CRD that references PrometheusRule + links to Prophet CRDs

---

### 1.8 Observability ✅ **IMPLEMENTED (Via Stack Integration)**

| PRD Requirement | Prophet Implementation | Status | Notes |
|----------------|------------------------|--------|-------|
| `Observability` CRD | Not a CRD, but full stack | ✅ **Implemented** | Prometheus + Grafana + OTel |
| ServiceMonitor/PodMonitor | Via Prometheus Operator | ✅ **Implemented** | Standard Prometheus CRDs |
| Grafana dashboards | Predefined dashboards | ✅ **Implemented** | ML forecasting, anomalies |
| Anomaly detection | Grafana ML + AnomalyAction | ✅ **Implemented** | AI-powered anomaly detection |
| eBPF observability | Cilium + Hubble | ✅ **Implemented** | Unique to Prophet |

**Recommendation:**
- **Create `Observability` CRD** as a convenience wrapper
- Auto-generate ServiceMonitor/PodMonitor from Observability CRD
- Bundle Grafana dashboard provisioning
- Document eBPF integration patterns

---

## 2. Unique Prophet Features (Beyond PRD)

Prophet includes several capabilities **not in the PRD** that enhance self-healing:

### 2.1 AutonomousAgent Operator ✅
- **LLM-powered decision making**: In-cluster inference (Ollama) for autonomous remediation
- **MCP Server**: Model Context Protocol for AI agent integration
- **Safety modes**: Autonomous, human-in-loop, dry-run
- **Status**: Fully implemented (`AutonomousAction` CRD)

### 2.2 K8sGPT Integration ✅
- **AI diagnostics**: Plain-English explanations of cluster issues
- **Auto-analysis**: Triggered on anomalies/alerts
- **Status**: Integrated with `AnomalyAction`

### 2.3 eBPF Observability ✅
- **Cilium + Hubble**: Kernel-level network visibility
- **Zero overhead**: <1% performance impact
- **Status**: Deployed via `clusters/common/network/cilium.yaml`

### 2.4 SLOEnforcer Operator ✅
- **Error budget monitoring**: Tracks SLO violations
- **Predictive exhaustion**: Forecasts when budgets will be exhausted
- **Status**: Implemented (`SLOViolation` CRD)

---

## 3. Implementation Roadmap

### Phase 1: Core Gaps (High Priority)
1. **HealthCheck CRD** (2-3 weeks)
   - Define CRD spec with probe types
   - Controller for health evaluation
   - Integration with AnomalyAction

2. **BudgetGuard CRD** (3-4 weeks)
   - Define CRD with budget limits
   - Integrate OpenCost/Kubecost API
   - Webhook admission controller
   - Actions on budget exceed

3. **CostAlert CRD** (2 weeks)
   - Define CRD with thresholds
   - Cost metric collection
   - Link to AlertRule/PrometheusRule

### Phase 2: Enhancements (Medium Priority)
4. **ClusterOptimizer CRD** (3-4 weeks)
   - Define CRD with optimization strategies
   - Karpenter integration
   - Cron-based scheduling

5. **Autoscaling Enhancement** (2 weeks)
   - Extend PredictiveScale for workload-level scaling
   - Add targetRef to Deployments/StatefulSets
   - Add behavior fields (cooldown, stabilization)

6. **Observability CRD** (1-2 weeks)
   - Convenience wrapper for ServiceMonitor/PodMonitor
   - Dashboard provisioning

### Phase 3: Advanced Features (Low Priority)
7. **Multi-step Recovery Plans**
   - Enhance AnomalyAction with `steps` array
   - Sequential recovery workflows

8. **AlertRule Wrapper**
   - CRD that references PrometheusRule
   - Links to Prophet CRDs

---

## 4. CRD Design Recommendations

### 4.1 Naming Convention
- **Keep existing group**: `aiops.prophet.io/v1alpha1` (don't use `aiops.example.com`)
- **Consistent naming**: Use `PascalCase` for CRD kinds
- **Plural forms**: `anomalyactions`, `healthchecks`, `budgetguards`, etc.

### 4.2 Common Patterns
All new CRDs should follow Prophet's existing patterns:
- **Status subresource**: Always include `.status` with conditions
- **Print columns**: Use kubebuilder markers for `kubectl get` output
- **Finalizers**: For cleanup/cleanup hooks
- **Owner references**: Link to target workloads

### 4.3 Integration Points
- **AnomalyAction**: Central remediation hub (link HealthCheck, CostAlert → AnomalyAction)
- **AutonomousAction**: LLM-powered decision layer (can reason about all CRDs)
- **Prometheus**: Metrics/alerts source (all CRDs should emit metrics)

---

## 5. Example: HealthCheck CRD Spec

```yaml
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: healthchecks.aiops.prophet.io
spec:
  group: aiops.prophet.io
  versions:
  - name: v1alpha1
    served: true
    storage: true
    schema:
      openAPIV3Schema:
        type: object
        properties:
          spec:
            type: object
            properties:
              targetRef:
                type: object
                properties:
                  apiVersion:
                    type: string
                  kind:
                    type: string
                  name:
                    type: string
                  namespace:
                    type: string
              probes:
                type: array
                items:
                  type: object
                  properties:
                    type:
                      type: string
                      enum: [http, tcp, command, custom]
                    httpGet:
                      type: object
                    tcpSocket:
                      type: object
                    exec:
                      type: object
                    custom:
                      type: object
                      description: "Custom probe (e.g., DB connectivity check)"
              failureThreshold:
                type: integer
                default: 3
              periodSeconds:
                type: integer
                default: 10
              remediation:
                type: object
                properties:
                  action:
                    type: string
                    enum: [restart, trigger-recovery-plan, alert]
                  recoveryPlanRef:
                    type: object
                    description: "Reference to AnomalyAction/RecoveryPlan"
          status:
            type: object
            properties:
              healthy:
                type: boolean
              lastCheckTime:
                type: string
                format: date-time
              failureCount:
                type: integer
              conditions:
                type: array
                items:
                  type: object
  scope: Namespaced
  names:
    plural: healthchecks
    singular: healthcheck
    kind: HealthCheck
```

---

## 6. Success Metrics Alignment

| PRD Metric | Prophet Implementation | Status |
|-----------|------------------------|--------|
| Reduction in manual interventions (MTTR) | AnomalyAction auto-remediation | ✅ **Tracked** (remediationCount) |
| Cost savings | Not tracked | ❌ **Gap** (needs BudgetGuard) |
| Alert resolution time | Not explicitly tracked | ⚠️ **Partial** (via Prometheus) |
| Cluster uptime >99.99% | Not tracked | ⚠️ **Gap** (needs SLO tracking) |

**Recommendation:**
- Add metrics to all CRDs (Prometheus counters/gauges)
- Create Grafana dashboard for Prophet success metrics
- Track MTTR, cost savings, alert resolution time

---

## 7. Conclusion

**Prophet is ~60% aligned with the PRD**, with strong coverage in:
- ✅ Anomaly detection & remediation
- ✅ Predictive scaling
- ✅ SLO enforcement
- ✅ AI-powered diagnostics

**Key gaps to address:**
- ❌ Cost management (BudgetGuard, CostAlert)
- ❌ Advanced health checks
- ❌ Cluster optimization

**Recommendation:**
1. **Implement HealthCheck, BudgetGuard, CostAlert** (Phase 1)
2. **Enhance existing CRDs** rather than creating duplicates
3. **Maintain Prophet's AI-first approach** (don't lose LLM/autonomous capabilities)
4. **Document integration patterns** between new and existing CRDs

Prophet's unique strengths (LLM agents, K8sGPT, eBPF) should be **preserved and enhanced**, not replaced by PRD requirements.

