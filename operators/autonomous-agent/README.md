# Autonomous Agent Operator

The Autonomous Agent operator implements MCP-powered remediation agents for Prophet v6.

## Features

- **MCP Server**: Exposes Kubernetes operations as MCP tools
- **Agent Workflow**: Observe → Reason → Decide → Act
- **Safety Gates**: Rate limiting, approval gates, audit logging
- **LLM Integration**: Supports Ollama, OpenAI, and other providers

## Architecture

```
autonomous-agent/
├── mcp-server/          # MCP protocol implementation
│   ├── server.go        # MCP server (HTTP/WebSocket)
│   └── tools/           # Kubernetes tool definitions
│       └── tools.go     # Tool executor
├── controllers/         # Kubernetes controllers
│   ├── autonomousaction_controller.go  # Main reconciler
│   └── actions.go       # Action executor with safety gates
├── llm-inference/       # LLM client implementations
│   └── client.go       # Ollama, OpenAI clients
└── api/v1alpha1/        # CRD definitions
    └── autonomousaction_types.go
```

## MCP Tools

### Read-Only (Auto-Approved)
- `k8s_get_pods`
- `k8s_get_nodes`
- `k8s_get_deployments`
- `k8s_get_events`
- `k8s_get_metrics`
- `k8s_get_k8sgpt_analysis`
- `k8s_get_forecast`

### Actions (Require Approval)
- `k8s_scale_deployment`
- `k8s_restart_pod`
- `k8s_cordon_node`
- `k8s_drain_node`
- `k8s_apply_network_policy`

## Usage

### 1. Build

```bash
make docker-build docker-push
```

### 2. Deploy

```bash
make deploy
```

### 3. Create AutonomousAction

See `../../aiops/agents/autonomous-remediation-agent.yaml` for examples.

## Development

### Local Testing with Tilt

```bash
tilt up
```

### MCP Server Endpoints

- HTTPS: `https://localhost:8443`
- WebSocket (TLS): `wss://localhost:8443/mcp`
- HTTP (legacy): `http://localhost:8082`
- Tools List: `GET /mcp/tools/list`
- Tool Call: `POST /mcp/tools/call`
- Health: `GET /health`

## Safety Features

1. **Rate Limiting**: Max 10 actions per 5 minutes (configurable)
2. **Approval Gates**: autonomous / human-in-loop / dry-run
3. **Audit Logging**: All actions logged as Kubernetes Events
4. **Constraints**: Allowed actions, forbidden namespaces, cooldowns

## See Also

- [Prophet v6 Documentation](../../V6-AGENTIC-AUTONOMY.md)
- [MCP Client Configuration](../../clusters/common/aiops/mcp/client-config.yaml)
- [Agent Workflows](../../aiops/agents/autonomous-remediation-agent.yaml)

