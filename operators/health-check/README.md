# HealthCheck Operator

The HealthCheck operator provides advanced health checking capabilities for Kubernetes workloads, extending beyond native liveness/readiness probes.

## Features

- **Multiple Probe Types**: HTTP, TCP, Command, and Custom probes
- **Composite Health Checks**: Multiple probes per workload
- **Custom Probes**: Database connectivity, external API checks, etc.
- **Auto-Remediation**: Automatic restart or recovery plan triggering on failure
- **Integration with AnomalyAction**: Link health checks to recovery workflows

## CRD: HealthCheck

```yaml
apiVersion: aiops.prophet.io/v1alpha1
kind: HealthCheck
metadata:
  name: backend-health-check
  namespace: default
spec:
  targetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: backend
    namespace: default
  probes:
    - name: http-readiness
      type: http
      httpGet:
        path: /health
        port: 8080
    - name: db-connectivity
      type: custom
      custom:
        description: "Check database connectivity"
        script: |
          psql -h $DB_HOST -U $DB_USER -d $DB_NAME -c "SELECT 1"
        env:
          - name: DB_HOST
            value: "postgres.default.svc.cluster.local"
  failureThreshold: 3
  periodSeconds: 10
  remediation:
    action: restart
    cooldownSeconds: 300
```

## Probe Types

### HTTP Probe
Standard HTTP health check:
```yaml
probes:
  - name: api-health
    type: http
    httpGet:
      path: /health
      port: 8080
      scheme: HTTP
      httpHeaders:
        - name: X-Custom-Header
          value: "value"
```

### TCP Probe
TCP connectivity check:
```yaml
probes:
  - name: database-port
    type: tcp
    tcpSocket:
      port: 5432
```

### Command Probe
Execute a command in the container:
```yaml
probes:
  - name: disk-space
    type: command
    exec:
      command:
        - /bin/sh
        - -c
        - df -h / | awk 'NR==2 {exit ($5+0 > 90)}'
```

### Custom Probe
Custom health check (e.g., database connectivity):
```yaml
probes:
  - name: db-connectivity
    type: custom
    custom:
      description: "Check PostgreSQL connectivity"
      script: |
        #!/bin/sh
        psql -h $DB_HOST -U $DB_USER -d $DB_NAME -c "SELECT 1" || exit 1
      env:
        - name: DB_HOST
          value: "postgres.default.svc.cluster.local"
        - name: DB_USER
          valueFrom:
            secretKeyRef:
              name: db-credentials
              key: username
```

## Remediation Actions

When health checks fail, the operator can:

1. **Restart**: Delete pods to trigger restart
2. **Trigger Recovery Plan**: Reference an `AnomalyAction` for complex recovery
3. **Alert**: Create Kubernetes events for external alerting
4. **None**: Just monitor without action

## Status Fields

- `healthy`: Boolean indicating current health status
- `lastCheckTime`: Timestamp of last health check
- `failureCount`: Consecutive failure count
- `probeResults`: Results of each probe
- `remediationCount`: Number of remediation actions performed

## Integration with AnomalyAction

HealthCheck can trigger `AnomalyAction` resources for recovery:

```yaml
spec:
  remediation:
    action: trigger-recovery-plan
    recoveryPlanRef:
      name: app-recovery-plan
      namespace: default
```

## Deployment

Deploy the operator:

```bash
kubectl apply -f clusters/common/aiops/operators/health-check.yaml
```

## Example Use Cases

1. **Database-Dependent Apps**: Check DB connectivity before marking healthy
2. **External API Dependencies**: Verify external services are reachable
3. **Composite Health**: Multiple checks (API + DB + cache) must all pass
4. **Custom Business Logic**: Script-based health checks for complex scenarios

## See Also

- [PRD Alignment](../PRD-ALIGNMENT.md) - How HealthCheck fits into the PRD requirements
- [AnomalyAction](../anomaly-remediator/README.md) - Recovery workflows
- [Prophet README](../../README.md) - Overall Prophet documentation

