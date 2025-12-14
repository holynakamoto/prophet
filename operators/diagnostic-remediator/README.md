# DiagnosticRemediator Operator

A Golang operator that performs **deep diagnostics** on Kubernetes workloads and **automatically fixes** configuration issues, resource constraints, and dependency problems.

## Overview

Unlike simple pod restart operators, DiagnosticRemediator:

1. **Diagnoses root causes**: Checks resources, environment variables, ConfigMaps/Secrets, service dependencies, image pull policies
2. **Automatically fixes issues**: Adds missing resources, environment variables, updates configurations
3. **Validates dependencies**: Verifies services are reachable, ConfigMaps/Secrets exist
4. **Tracks remediation history**: Records all fixes applied with timestamps

## CRD: DiagnosticRemediation

```yaml
apiVersion: aiops.prophet.io/v1alpha1
kind: DiagnosticRemediation
metadata:
  name: rancher-diagnostic-remediation
  namespace: cattle-system
spec:
  target:
    namespace: cattle-system
    kind: Deployment
    name: rancher
  
  diagnostics:
    resources: true              # Check CPU/memory limits
    environment: true            # Check env vars
    configReferences: true       # Check ConfigMap/Secret refs
    serviceDependencies:         # Check service connectivity
      - name: postgres
        port: 5432
        protocol: TCP
  
  remediation:
    fixResources: true           # Add default resources
    fixEnvironment: true         # Add required env vars
    fixImagePullPolicy: true     # Fix pull policy
    restartOnConfigChange: true  # Restart after fixes
    
    defaultResources:
      cpuRequest: "100m"
      cpuLimit: "1000m"
      memoryRequest: "512Mi"
      memoryLimit: "2Gi"
    
    requiredEnvVars:
      - name: DATABASE_URL
        value: "postgres://..."
  
  autoFix: true
  cooldownSeconds: 300
```

## Diagnostic Checks

| Check | What It Does | Example Issues Found |
|-------|--------------|---------------------|
| `resources` | Checks CPU/memory requests/limits | Missing resource requests, no limits set |
| `environment` | Validates required env vars | Missing DATABASE_URL, API_KEY not set |
| `configReferences` | Verifies ConfigMaps/Secrets exist | Referenced ConfigMap missing, Secret key not found |
| `serviceDependencies` | Tests service connectivity | Database service unreachable, API endpoint down |
| `imagePull` | Validates image pull policy | Using `latest` tag without PullAlways |
| `persistentVolumes` | Checks PVC availability | PVC not bound, storage class missing |
| `networkPolicies` | Validates network policies | Blocking policies preventing connectivity |

## Remediation Actions

| Action | What It Does |
|--------|--------------|
| `fixResources` | Adds default CPU/memory requests/limits if missing |
| `fixEnvironment` | Adds required environment variables |
| `fixImagePullPolicy` | Updates image pull policy to recommended value |
| `createMissingConfigs` | Creates placeholder ConfigMaps/Secrets (use with caution) |
| `restartOnConfigChange` | Restarts pods after configuration updates |
| `scaleUp` | Scales up deployment if resources insufficient |

## Status Fields

```yaml
status:
  phase: Resolved                    # Pending | Diagnosing | IssuesFound | Remediating | Resolved | Failed
  lastDiagnosed: "2025-12-13T..."
  lastRemediated: "2025-12-13T..."
  issues:                            # Found issues
    - type: MissingResources
      severity: Warning
      description: "Container has no resource requests"
      suggestedFix: "Add resource requests"
  remediations:                      # Applied fixes
    - type: AddedResources
      description: "Added default resource requests and limits"
      timestamp: "2025-12-13T..."
      success: true
  remediationCount: 3
```

## Example: Fixing Rancher

```yaml
apiVersion: aiops.prophet.io/v1alpha1
kind: DiagnosticRemediation
metadata:
  name: rancher-fix
  namespace: cattle-system
spec:
  target:
    namespace: cattle-system
    kind: Deployment
    name: rancher
  
  diagnostics:
    resources: true
    environment: true
    configReferences: true
    serviceDependencies:
      - name: rancher
        port: 80
        protocol: HTTP
  
  remediation:
    fixResources: true
    fixEnvironment: true
    defaultResources:
      cpuRequest: "500m"
      cpuLimit: "2000m"
      memoryRequest: "1Gi"
      memoryLimit: "4Gi"
  
  autoFix: true
```

## Development

```bash
# Generate code
make generate manifests

# Run locally
make run

# Build and deploy
make docker-build docker-push
make deploy
```

## Files

```
diagnostic-remediator/
├── api/v1alpha1/
│   ├── diagnosticremediation_types.go  # CRD definitions
│   └── groupversion_info.go
├── controllers/
│   └── diagnosticremediation_controller.go  # Main logic
├── config/
│   ├── crd/bases/                    # Generated CRD
│   ├── rbac/                         # RBAC manifests
│   └── samples/                      # Example resources
└── cmd/
    └── main.go                       # Entrypoint
```

## See Also

- [AnomalyRemediator](../anomaly-remediator/) - Simple pod restart remediation
- [HealthCheck](../health-check/) - Health probe monitoring
- [Rancher Test Scenario](../../resilience/chaos-experiments/README-RANCHER-TEST.md)

