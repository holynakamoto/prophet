# Helm-Aware Remediation Validation Guide

## Quick Validation Checklist

Once the `diagnostic-remediator` operator is running, verify the Helm-aware remediation flow:

### 1. Check Operator is Running

```bash
kubectl get pods -n prophet-operators -l app=diagnostic-remediator
# Should show: READY 1/1, STATUS Running
```

### 2. Verify Helm Release Health

```bash
# Check Helm release status (should show deployed, not failed)
helm status rancher -n cattle-system

# Check release details
helm get all rancher -n cattle-system

# List all releases
helm list -n cattle-system
```

**Expected:** Helm release should show `STATUS: deployed` with healthy pods, not `failed` or broken hooks.

### 3. Inspect Deployment Annotations

After remediation, check that the rollout restart annotation was set:

```bash
kubectl get deployment rancher -n cattle-system -o yaml | grep -A 10 annotations
```

**Expected annotations:**
```yaml
spec:
  template:
    metadata:
      annotations:
        kubectl.kubernetes.io/restartedAt: "2025-12-13T..."
        prophet.aiops.io/restartedAt: "2025-12-13T..."
        prophet.aiops.io/restartReason: "pod-health-remediation"
        prophet.aiops.io/restartedBy: "diagnostic-remediator"
```

### 4. Watch Pod Rollout Behavior

```bash
# Watch pods during remediation
kubectl get pods -n cattle-system -l app=rancher -w
```

**Expected behavior:**
1. Old pods transition to `Terminating`
2. New pods are created with new ReplicaSet
3. New pods progress: `Pending` → `ContainerCreating` → `Running` → `Ready`
4. All pods eventually become `Ready`
5. Old pods are fully terminated

### 5. Check DiagnosticRemediation Status

```bash
kubectl get diagnosticremediations -n cattle-system rancher-comprehensive-fix -o yaml
```

**Expected status progression:**
```yaml
status:
  phase: Resolved  # or Remediating
  issues:
    - type: PodCrashLoopBackOff
      severity: Critical
      description: "Pod rancher-xxx is in CrashLoopBackOff state"
  remediations:
    - type: PodRestart
      description: "Remediated pod health issue: Pod rancher-xxx is in CrashLoopBackOff state"
      success: true
      timestamp: "2025-12-13T..."
  remediationCount: 1
  lastRemediated: "2025-12-13T..."
```

## Comprehensive Validation Script

```bash
#!/bin/bash
# validate-helm-remediation.sh

set -e

NAMESPACE="cattle-system"
RELEASE="rancher"
OPERATOR_NS="prophet-operators"

echo "🔍 Validating Helm-Aware Remediation"
echo ""

# 1. Check operator is running
echo "1. Checking operator status..."
OPERATOR_POD=$(kubectl get pods -n $OPERATOR_NS -l app=diagnostic-remediator -o jsonpath='{.items[0].metadata.name}' 2>/dev/null || echo "")
if [ -z "$OPERATOR_POD" ]; then
    echo "   ❌ Operator pod not found"
    exit 1
fi
OPERATOR_STATUS=$(kubectl get pod -n $OPERATOR_NS $OPERATOR_POD -o jsonpath='{.status.phase}')
if [ "$OPERATOR_STATUS" != "Running" ]; then
    echo "   ❌ Operator pod is not Running (status: $OPERATOR_STATUS)"
    exit 1
fi
echo "   ✅ Operator is Running"

# 2. Check Helm release status
echo ""
echo "2. Checking Helm release status..."
HELM_STATUS=$(helm status $RELEASE -n $NAMESPACE -o json 2>/dev/null | jq -r '.info.status' || echo "unknown")
if [ "$HELM_STATUS" != "deployed" ]; then
    echo "   ⚠️  Helm release status: $HELM_STATUS (expected: deployed)"
else
    echo "   ✅ Helm release is deployed"
fi

# 3. Check deployment annotations
echo ""
echo "3. Checking deployment annotations..."
RESTART_ANNOTATION=$(kubectl get deployment $RELEASE -n $NAMESPACE -o jsonpath='{.spec.template.metadata.annotations.kubectl\.kubernetes\.io/restartedAt}' 2>/dev/null || echo "")
if [ -n "$RESTART_ANNOTATION" ]; then
    echo "   ✅ Rollout restart annotation found: $RESTART_ANNOTATION"
else
    echo "   ℹ️  No rollout restart annotation (operator hasn't remediated yet or no issues found)"
fi

# 4. Check pod status
echo ""
echo "4. Checking Rancher pods..."
POD_COUNT=$(kubectl get pods -n $NAMESPACE -l app=$RELEASE --no-headers 2>/dev/null | wc -l | tr -d ' ')
READY_COUNT=$(kubectl get pods -n $NAMESPACE -l app=$RELEASE --field-selector=status.phase=Running --no-headers 2>/dev/null | grep -c "1/1" || echo "0")
echo "   Pods: $POD_COUNT total, $READY_COUNT ready"

# 5. Check DiagnosticRemediation status
echo ""
echo "5. Checking DiagnosticRemediation status..."
DR_PHASE=$(kubectl get diagnosticremediations -n $NAMESPACE rancher-comprehensive-fix -o jsonpath='{.status.phase}' 2>/dev/null || echo "unknown")
DR_REMEDIATIONS=$(kubectl get diagnosticremediations -n $NAMESPACE rancher-comprehensive-fix -o jsonpath='{.status.remediationCount}' 2>/dev/null || echo "0")
echo "   Phase: $DR_PHASE"
echo "   Remediations: $DR_REMEDIATIONS"

echo ""
echo "✅ Validation complete!"
```

## Chaos Testing Validation

When running chaos experiments, verify both:

1. **Application Recovery**: Rancher pods recover and become healthy
2. **Helm Release Health**: `helm status rancher -n cattle-system` remains `deployed`

### Example Chaos Test Assertion

```bash
# After chaos injection
CHAOS_INJECTED=$(date +%s)

# Wait for remediation
sleep 60

# Verify Helm release is still healthy
HELM_STATUS=$(helm status rancher -n cattle-system -o json | jq -r '.info.status')
if [ "$HELM_STATUS" != "deployed" ]; then
    echo "❌ FAIL: Helm release status is $HELM_STATUS (expected: deployed)"
    exit 1
fi

# Verify pods are healthy
UNHEALTHY_PODS=$(kubectl get pods -n cattle-system -l app=rancher --field-selector=status.phase!=Running --no-headers 2>/dev/null | wc -l | tr -d ' ')
if [ "$UNHEALTHY_PODS" -gt 0 ]; then
    echo "❌ FAIL: $UNHEALTHY_PODS unhealthy pods found"
    exit 1
fi

echo "✅ PASS: Application recovered and Helm release is healthy"
```

## Guardrails Verification

### Rate Limiting

Check that the operator respects max remediations per hour:

```bash
# Set max remediations per hour
kubectl annotate diagnosticremediations -n cattle-system rancher-comprehensive-fix \
  prophet.aiops.io/maxRemediationsPerHour=3

# Watch operator logs for rate limiting
kubectl logs -n prophet-operators -l app=diagnostic-remediator -f | grep -i "max remediations"
```

**Expected:** After 3 successful remediations in an hour, operator should log "Max remediations per hour reached" and skip further remediations until the window resets.

### Idempotency

Verify that rapid successive remediations don't trigger unnecessary restarts:

```bash
# Check deployment annotations for recent restart
kubectl get deployment rancher -n cattle-system -o jsonpath='{.spec.template.metadata.annotations.prophet\.aiops\.io/restartedAt}'

# If restart was < 2 minutes ago, operator should skip
# Check operator logs
kubectl logs -n prophet-operators -l app=diagnostic-remediator --tail=50 | grep -i "recent restart"
```

**Expected:** Operator should log "Skipping rollout restart - recent restart detected" if a restart happened within the last 2 minutes.

## Troubleshooting

### Helm Release Shows as Failed

If `helm status` shows `failed`:

1. Check release hooks:
   ```bash
   helm get hooks rancher -n cattle-system
   ```

2. Check release notes:
   ```bash
   helm get notes rancher -n cattle-system
   ```

3. Check deployment status:
   ```bash
   kubectl get deployment rancher -n cattle-system
   kubectl describe deployment rancher -n cattle-system
   ```

### Rollout Restart Not Working

1. Check operator has update permissions:
   ```bash
   kubectl get clusterrole diagnostic-remediator-manager-role -o yaml | grep -A 5 deployments
   ```

2. Check operator logs:
   ```bash
   kubectl logs -n prophet-operators -l app=diagnostic-remediator | grep -i rollout
   ```

3. Check deployment annotations:
   ```bash
   kubectl get deployment rancher -n cattle-system -o yaml | grep -A 10 annotations
   ```

### Pods Not Recovering

1. Check pod events:
   ```bash
   kubectl describe pod -n cattle-system -l app=rancher
   ```

2. Check operator detected issues:
   ```bash
   kubectl get diagnosticremediations -n cattle-system rancher-comprehensive-fix -o yaml | grep -A 20 "issues:"
   ```

3. Check remediation actions:
   ```bash
   kubectl get diagnosticremediations -n cattle-system rancher-comprehensive-fix -o yaml | grep -A 20 "remediations:"
   ```

