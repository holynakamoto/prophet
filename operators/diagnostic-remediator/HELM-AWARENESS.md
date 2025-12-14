# Helm-Aware Remediation

## Overview

The DiagnosticRemediator operator is designed to work safely with Helm-managed resources, particularly Rancher deployments.

## Key Principles

### 1. Prefer Rollout Restart Over Pod Deletion

For Helm-managed resources, the operator **always** uses rollout restart instead of deleting individual pods:

- **Rollout restart**: Updates deployment annotation (`kubectl.kubernetes.io/restartedAt`)
  - Equivalent to `kubectl rollout restart deployment/rancher -n cattle-system`
  - Preserves Helm ownership
  - Triggers controlled pod replacement
  - Safer for production workloads

- **Pod deletion**: Only used for non-Helm resources in specific cases
  - Can cause issues with Helm-managed resources
  - May conflict with Helm's reconciliation

### 2. Helm Detection

The operator automatically detects Helm-managed resources by checking labels:

```yaml
labels:
  app.kubernetes.io/managed-by: Helm
  heritage: Helm
  release: <release-name>
  chart: <chart-name>
```

If these labels are present, the operator will:
- Use rollout restart for all pod health issues
- Log Helm release information
- Respect Helm ownership

### 3. Rancher-Specific Configuration

For Rancher deployments, the `DiagnosticRemediation` CR should include:

```yaml
spec:
  target:
    namespace: cattle-system
    kind: Deployment
    name: rancher
    labels:
      app: rancher
      app.kubernetes.io/managed-by: Helm
      release: rancher
```

This ensures:
- Correct pod selection via label matching
- Helm-aware remediation
- Proper tracking of Helm release health

## Remediation Flow

### For Helm-Managed Resources

1. **Detect Issue**: Operator detects pod health problem (CrashLoopBackOff, high restarts, stuck)
2. **Check Helm Labels**: Operator checks if deployment is Helm-managed
3. **Trigger Rollout Restart**: Updates deployment annotation to trigger restart
4. **Track in Status**: Records remediation action in DiagnosticRemediation status

### For Non-Helm Resources

1. **Detect Issue**: Operator detects pod health problem
2. **Choose Strategy**:
   - **CrashLoopBackOff / High Restarts**: Delete pod (triggers immediate recreation)
   - **Stuck Pods**: Rollout restart (controlled replacement)
3. **Track in Status**: Records remediation action

## Verification

### Check Helm Release Status

```bash
# Check Helm release health
helm status rancher -n cattle-system

# List Helm releases
helm list -n cattle-system

# Check deployment annotations (should show restart annotation)
kubectl get deployment rancher -n cattle-system -o yaml | grep -A 5 annotations
```

### Verify Rollout Restart

After remediation, you should see:

1. **Deployment annotation updated**:
   ```yaml
   spec:
     template:
       metadata:
         annotations:
           kubectl.kubernetes.io/restartedAt: "2025-12-13T..."
           prophet.aiops.io/restartedAt: "2025-12-13T..."
           prophet.aiops.io/restartReason: "pod-health-remediation"
   ```

2. **New ReplicaSet created**: Kubernetes creates a new ReplicaSet with the updated annotation

3. **Pods replaced**: Old pods are terminated, new pods are created

4. **Helm release remains healthy**: Helm still recognizes the deployment as managed

## Benefits

✅ **Helm Compatibility**: Works seamlessly with Helm-managed resources  
✅ **Safer Remediation**: Rollout restart is less disruptive than pod deletion  
✅ **Controlled Replacement**: Kubernetes manages the rollout process  
✅ **Helm Awareness**: Operator respects Helm ownership and labels  
✅ **Production Ready**: Suitable for production workloads managed by Helm  

## Troubleshooting

### Rollout Restart Not Working

1. Check deployment is Helm-managed:
   ```bash
   kubectl get deployment rancher -n cattle-system -o yaml | grep -i helm
   ```

2. Verify operator has update permissions:
   ```bash
   kubectl get clusterrole diagnostic-remediator-manager-role -o yaml | grep -A 5 deployments
   ```

3. Check operator logs:
   ```bash
   kubectl logs -n prophet-operators -l app=diagnostic-remediator | grep -i rollout
   ```

### Helm Release Shows as Failed

If Helm release status is "failed" after remediation:

1. Check deployment status:
   ```bash
   kubectl get deployment rancher -n cattle-system
   ```

2. Check pod status:
   ```bash
   kubectl get pods -n cattle-system -l app=rancher
   ```

3. Check Helm release details:
   ```bash
   helm status rancher -n cattle-system
   helm get all rancher -n cattle-system
   ```

The operator focuses on fixing pod health issues. If the Helm release is failing due to configuration issues, those need to be addressed separately (e.g., via Helm values or ConfigMap updates).

