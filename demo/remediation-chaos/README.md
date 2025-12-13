# Prophet Remediation Chaos Demo

Watch the **AnomalyRemediator** operator detect failures and auto-remediate via pod restarts—perfect for proving "agentic infrastructure" without needing full AIOps/ML setup.

## Quick Start

```bash
cd demo/remediation-chaos
./demo.sh
```

That's it! The script will:
1. Deploy a crashing app (`crashy-app`)
2. Apply autonomous restart policy (`AnomalyAction`)
3. Inject chaos (repeated pod kills via Chaos Mesh)
4. Watch the operator auto-restart pods in real-time
5. Clean up everything when done

## What You'll See

### Before Chaos
```
NAME                          READY   STATUS    RESTARTS   AGE
crashy-app-7d4f8b9c6-xxxxx    1/1     Running   0          10s
crashy-app-7d4f8b9c6-yyyyy    1/1     Running   0          10s
crashy-app-7d4f8b9c6-zzzzz    1/1     Running   0          10s
```

### During Chaos (Pods Killed)
```
NAME                          READY   STATUS              RESTARTS   AGE
crashy-app-7d4f8b9c6-xxxxx    0/1     Terminating         0          30s
crashy-app-7d4f8b9c6-yyyyy    0/1     Terminating         0          30s
```

### After Remediation (Auto-Restarted)
```
NAME                          READY   STATUS    RESTARTS   AGE
crashy-app-7d4f8b9c6-new1     1/1     Running   0          5s
crashy-app-7d4f8b9c6-new2     1/1     Running   0          5s
```

## Requirements

- `kubectl` configured with cluster access
- Kubernetes cluster (kind, minikube, EKS, GKE, AKS, etc.)
- AnomalyRemediator operator deployed (script will attempt to deploy it)
- Chaos Mesh (optional - script will install if not present)

## Stabilizing the App After Demo

After the demo shows remediation in action, you can stabilize the app so it stops crashing:

```bash
# Option 1: Use the fix script
./fix-app.sh

# Option 2: Manually patch the ConfigMap
kubectl patch configmap crashy-app-config -n demo-prophet --type merge -p '{"data":{"behavior":"stable"}}'
```

The app will switch from crash mode to stable mode, and pods will run normally.

## Manual Steps

If you prefer to run manually:

### 1. Deploy the App
```bash
kubectl apply -f vulnerable-app.yaml
```

### 2. Deploy the Remediation Policy
```bash
kubectl apply -f anomaly-action.yaml
```

### 3. Watch in Real-Time
```bash
# Terminal 1: Watch pods
kubectl get pods -n demo-prophet -w

# Terminal 2: Watch events
kubectl get events -n demo-prophet --sort-by='.lastTimestamp' -w

# Terminal 3: Watch AnomalyAction status
kubectl get anomalyaction crashy-remediation -n demo-prophet -o yaml -w
```

### 4. Inject Chaos (Optional)
```bash
# Install Chaos Mesh first
kubectl apply -f https://mirrors.chaos-mesh.org/latest/crd.yaml
kubectl apply -f https://mirrors.chaos-mesh.org/latest/chaos-mesh.yaml -n chaos-testing

# Inject chaos
kubectl apply -f chaos-pod-kill.yaml
```

### 5. Clean Up
```bash
kubectl delete -f chaos-pod-kill.yaml --ignore-not-found
kubectl delete -f anomaly-action.yaml
kubectl delete -f vulnerable-app.yaml
kubectl delete ns demo-prophet
```

## How It Works

1. **Vulnerable App**: The `crashy-app` deployment runs pods that intentionally crash after 10 seconds
2. **AnomalyAction**: The `AnomalyAction` CR watches for pods in `Failed` or `CrashLoopBackOff` state
3. **Detection**: When pods fail, the AnomalyRemediator operator detects the anomaly
4. **Remediation**: The operator automatically restarts the failed pods (no human intervention)
5. **Chaos**: Chaos Mesh kills pods every 20 seconds to demonstrate continuous remediation

## Troubleshooting

### Operator Not Found
```bash
# Deploy the operator manually
kubectl apply -f ../../clusters/common/aiops/operators/anomaly-remediator.yaml

# Check if it's running
kubectl get pods -n prophet-operators
```

### Pods Not Restarting
```bash
# Check operator logs
kubectl logs -n prophet-operators -l control-plane=controller-manager -f

# Check AnomalyAction status
kubectl describe anomalyaction crashy-remediation -n demo-prophet
```

### Chaos Mesh Not Working
```bash
# Check if CRDs are installed
kubectl get crd podchaos.chaos-mesh.org

# Check Chaos Mesh operator
kubectl get pods -n chaos-testing
```

## Extending the Demo

### Add More Chaos
Edit `chaos-pod-kill.yaml` to:
- Change kill frequency: `cron: "@every 10s"`
- Target specific pods: Add node selectors
- Add network chaos: Use `NetworkChaos` instead

### Test Different Remediations
Edit `anomaly-action.yaml` to:
- Scale instead of restart: `type: scale`
- Require approval: `requireApproval: true`
- Change cooldown: `cooldownSeconds: 60`

### Monitor with K8sGPT
If K8sGPT is deployed:
```bash
# Get AI analysis of the failures
kubectl exec -n k8sgpt deployment/k8sgpt-operator -- \
  k8sgpt analyze --namespace demo-prophet
```

## Grafana Dashboard

Watch the self-healing in real-time with the Prophet Self-Healing Demo dashboard:

```bash
# 1. Port-forward Grafana
kubectl port-forward -n monitoring svc/grafana 3000:3000 &

# 2. Import dashboard (automatically cleans up duplicates)
./import-dashboard.sh

# 3. Open the dashboard (stable URL):
#    http://localhost:3000/d/prophet-self-healing-demo/prophet-self-healing-demo
```

**If you see duplicate dashboards:**
- Run `./cleanup-dashboards.sh` to remove duplicates
- Or manually delete duplicates via Grafana UI
- See `FIX-DUPLICATE-DASHBOARDS.md` for details

The dashboard shows:
- **Pod Restart Rate**: Spikes during chaos, drops after remediation
- **Anomalies Detected**: Real-time counter
- **Remediations Executed**: Total actions taken
- **Pod Status**: Current state of all pods
- **Remediation Duration**: Performance metrics

## Next Steps

- Try the **PredictiveScaler** demo (coming soon)
- Test **AutonomousAgent** with LLM reasoning (see `aiops/agents/`)
- Explore full Prophet capabilities in the main [README](../../README.md)

---

**This demo showcases Prophet's autonomous remediation capabilities. The AnomalyRemediator detects failures and acts without human intervention—the future of infrastructure operations.** 🚀

