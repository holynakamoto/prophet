# Prophet v6 – Full Agentic Autonomy with MCP-Powered Remediation Agents

## Overview

Prophet v6 completes the journey to **full agentic autonomy** by implementing the Model Context Protocol (MCP) as the backbone for autonomous remediation agents. Building on v5's in-operator LLM inference, eBPF observability, multi-cluster federation, and custom Golang operators, v6 introduces **MCP-powered remediation agents** that enable secure, standardized interaction between AI models and the Kubernetes environment.

## Key Features

### 1. MCP-Powered Remediation Agents

- **MCP Server Implementation** in-operator (`autonomous-agent` operator)
  - Exposes Kubernetes resources, metrics, logs, and events as MCP tools
  - Supports tool discovery, invocation, and result streaming
  - Secure by default: RBAC-mapped permissions, request scoping, audit logging

- **Agent Workflow**:
  1. **Trigger** (anomaly, SLO violation, forecast threshold, or manual query)
  2. **Context Gathering** (live cluster state, K8sGPT analysis, Grafana ML forecasts, Hubble flows)
  3. **LLM Reasoning** (local inference or external via MCP client)
  4. **Action Proposal** (e.g., scale deployment, drain node, rollback, create ticket)
  5. **Execution**:
     - **Auto-mode**: Approved actions executed directly
     - **Review-mode**: Proposal sent for human approval (Slack, email, PR)
     - **Dry-run**: Simulated execution with impact preview

### 2. Supported Actions

Bounded set of safe, auditable actions:
- Scale HPA-managed deployments
- Cordon/drain nodes
- Restart failing pods
- Apply emergency network policies
- Create GitOps PRs for configuration changes
- Open incidents in external systems

### 3. Safety & Governance

- **Approval Gates**: Configurable per-action (auto/require-review/deny)
- **Dry-Run Mode**: Default for high-impact actions
- **Audit Trail**: All agent decisions and executions logged as Kubernetes Events + external sink
- **Rate Limiting**: Prevent runaway loops (configurable per agent)
- **Rollback Integration**: Automatic reversion on failed remediation

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    External MCP Clients                     │
│  (Claude Desktop, Cursor, Custom Agents)                     │
└──────────────────────┬──────────────────────────────────────┘
                       │ MCP Protocol (HTTP/WebSocket)
                       │
┌──────────────────────▼──────────────────────────────────────┐
│              MCP Server (autonomous-agent)                    │
│  ┌──────────────────────────────────────────────────────┐  │
│  │  MCP Tools:                                           │  │
│  │  - k8s_get_pods, k8s_get_nodes, k8s_get_deployments  │  │
│  │  - k8s_scale_deployment, k8s_restart_pod            │  │
│  │  - k8s_cordon_node, k8s_drain_node                   │  │
│  │  - k8s_get_k8sgpt_analysis, k8s_get_forecast        │  │
│  └──────────────────────────────────────────────────────┘  │
└──────────────────────┬──────────────────────────────────────┘
                       │
┌──────────────────────▼──────────────────────────────────────┐
│         AutonomousAction Controller                          │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐     │
│  │   Trigger    │→ │   Context    │→ │    LLM       │     │
│  │   Detection  │  │   Gathering   │  │  Reasoning   │     │
│  └──────────────┘  └──────────────┘  └──────────────┘     │
│         │                  │                  │             │
│         └──────────────────┴──────────────────┘             │
│                            │                                  │
│                   ┌────────▼────────┐                        │
│                   │ Action Executor │                        │
│                   │  - Safety Gates │                        │
│                   │  - Rate Limits │                        │
│                   │  - Audit Log   │                        │
│                   └─────────────────┘                        │
└──────────────────────────────────────────────────────────────┘
```

## Usage

### 1. Deploy Autonomous Agent

```yaml
apiVersion: aiops.prophet.io/v1alpha1
kind: AutonomousAction
metadata:
  name: autonomous-remediation-agent
  namespace: default
spec:
  trigger:
    type: anomaly
    anomalyScoreThreshold: 0.8
  
  llm:
    provider: ollama
    model: llama-3.2
    endpoint: http://ollama:11434
  
  context:
    includeK8sGPT: true
    includeMetrics: true
    includeEvents: true
    namespaces:
      - default
  
  approvalMode: human-in-loop  # Options: autonomous, human-in-loop, dry-run
  
  constraints:
    allowedActions:
      - scale
      - restart
    forbiddenNamespaces:
      - kube-system
    maxConcurrent: 3
    cooldownSeconds: 300
```

### 2. Connect External MCP Client

For Claude Desktop, add to `claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "prophet-kubernetes": {
      "command": "curl",
      "args": [
        "-X", "POST",
            "--cacert", "/path/to/prophet-mcp-ca-bundle.pem",
            "https://autonomous-agent-service.default.svc.cluster.local:8082/mcp/tools/call"
      ]
    }
  }
}
```

### 3. Agent Workflows

Three built-in workflows:

1. **Autonomous Recovery**: Fully automated recovery from common failures
2. **Predictive Intervention**: Proactive scaling based on forecasts
3. **Human-Assisted Fix**: Complex issues requiring human review

See `aiops/agents/autonomous-remediation-agent.yaml` for examples.

## Safety Features

### Rate Limiting

Prevents runaway action loops:
- Default: Max 10 actions per 5 minutes per agent
- Configurable per `AutonomousAction` resource
- Circuit breaker pattern for repeated failures

### Approval Gates

Three approval modes:

1. **autonomous**: Actions executed immediately (use with caution)
2. **human-in-loop**: Actions require manual approval
3. **dry-run**: Actions simulated, not executed

### Audit Trail

All agent decisions logged:
- Kubernetes Events (searchable via `kubectl get events`)
- In-memory audit log (last 1000 events)
- External sink support (configurable)

### Constraints

Per-agent constraints:
- `allowedActions`: Whitelist of action types
- `forbiddenNamespaces`: Blacklist of protected namespaces
- `maxConcurrent`: Maximum concurrent actions
- `cooldownSeconds`: Minimum time between actions

## MCP Tools Reference

### Read-Only Tools (Auto-Approved)

- `k8s_get_pods`: Get pods with optional namespace/label filters
- `k8s_get_nodes`: Get cluster nodes with status
- `k8s_get_deployments`: Get deployments
- `k8s_get_events`: Get recent events
- `k8s_get_metrics`: Get resource metrics
- `k8s_get_k8sgpt_analysis`: Get K8sGPT diagnostic analysis
- `k8s_get_forecast`: Get Grafana ML forecasts

### Action Tools (Require Approval)

- `k8s_scale_deployment`: Scale deployment to target replicas
- `k8s_restart_pod`: Restart a pod
- `k8s_cordon_node`: Cordon a node
- `k8s_drain_node`: Drain a node (high-impact)
- `k8s_apply_network_policy`: Apply emergency network policy

## Use Cases

### 1. Autonomous Recovery

**Scenario**: Pod OOM loop detected

**Flow**:
1. Anomaly detected (memory spike)
2. Agent gathers context (pod status, events, K8sGPT analysis)
3. LLM reasons: "Pod is OOMing, restart won't help. Scale deployment."
4. Action proposed: Scale deployment +1 replica
5. Executed (if approved) → Cluster stable

### 2. Predictive Intervention

**Scenario**: Forecast shows CPU spike in 1 hour

**Flow**:
1. Forecast threshold exceeded
2. Agent pre-scales HPA target
3. No pending pods when demand arrives

### 3. Security Response

**Scenario**: Suspicious traffic via Hubble

**Flow**:
1. Security event detected
2. Agent quarantines namespace with network policy
3. Alerts team for investigation

### 4. Human-Assisted Fix

**Scenario**: Complex multi-step issue

**Flow**:
1. Issue detected
2. Agent proposes multi-step plan
3. SRE reviews/approves via Slack
4. Executed step-by-step

## Success Metrics

- **Autonomous remediation success rate**: >90% for common failure modes
- **Mean Time to Recovery (MTTR)**: <3 minutes (from detection to resolution)
- **Zero unauthorized high-impact actions**: All actions logged and approved
- **Full audit coverage**: 100% of agent decisions logged
- **External MCP client integration**: Successfully tested with Claude Desktop

## Local Development

### Prerequisites

- Kubernetes cluster (kind, minikube, or cloud)
- Ollama for local LLM inference
- kubectl configured

### Setup

1. **Deploy Ollama**:
```bash
kubectl apply -f - <<EOF
apiVersion: apps/v1
kind: Deployment
metadata:
  name: ollama
spec:
  replicas: 1
  selector:
    matchLabels:
      app: ollama
  template:
    metadata:
      labels:
        app: ollama
    spec:
      containers:
      - name: ollama
        image: ollama/ollama:latest
        ports:
        - containerPort: 11434
EOF
```

2. **Build and deploy autonomous-agent operator**:
```bash
cd operators/autonomous-agent
make docker-build docker-push
make deploy
```

3. **Create AutonomousAction resource**:
```bash
kubectl apply -f aiops/agents/autonomous-remediation-agent.yaml
```

4. **Test MCP server**:
```bash
curl --cacert /path/to/prophet-mcp-ca-bundle.pem https://localhost:8443/mcp/tools/list
```

### Tilt Integration

The operator includes Tilt integration for local development:

```bash
cd operators/autonomous-agent
tilt up
```

Safety gates default to "review required" in dev mode.

## Security Considerations

1. **RBAC**: MCP server runs with minimal RBAC; tool permissions explicitly configured
2. **Network Policies**: MCP server only accessible within cluster (or via ingress with auth)
3. **Audit Logging**: All actions logged; external audit sink recommended for production
4. **Rate Limiting**: Prevents abuse and runaway loops
5. **Action Bounding**: Only predefined actions allowed; no arbitrary code execution

## Roadmap

- [ ] Integration with GitOps (create PRs for config changes)
- [ ] External incident management (PagerDuty, Opsgenie)
- [ ] Multi-cluster agent coordination
- [ ] Advanced forecasting integration (Grafana ML)
- [ ] Hubble flow analysis for security
- [ ] Cost optimization actions (right-sizing, spot instance management)

## References

- [Model Context Protocol Specification](https://modelcontextprotocol.io)
- [Prophet v5 Documentation](./V5-ASCENSION.md)
- [Operator Documentation](./OPERATORS.md)

---

**Prophet v6 represents the current frontier of agentic infrastructure**—a practical, secure implementation of autonomous remediation agents using the emerging MCP standard.

This is no longer just a repository.  
It's a working blueprint for the next generation of self-managing Kubernetes platforms.

Ready when you are. Let's build the future. 🚀

