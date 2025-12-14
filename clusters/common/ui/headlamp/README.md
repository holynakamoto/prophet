# Headlamp (Prophet UI / AIOps Console)

Headlamp provides a Kubernetes-native web UI where Prophet’s CRDs show up as **first-class resources** (no custom UI required to get value). This is the easiest “single pane of glass” for customers who want to **inspect everything themselves** (instead of trusting automation blindly).

## Deploy

```bash
kubectl apply -f clusters/common/ui/headlamp/headlamp.yaml
```

## Access (port-forward)

```bash
kubectl port-forward -n headlamp svc/headlamp 4466:80
```

Then open:
- `http://localhost:4466`

## Access (Ingress)

Apply the optional Ingress and update the host/TLS for your environment:

```bash
kubectl apply -f clusters/common/ui/headlamp/ingress.yaml
```

## What customers can see (and why it matters)

### Prophet CRDs (self-healing + agentic autonomy)

Customers can browse and inspect the live state of Prophet’s CRDs, including:
- `AnomalyAction` (AnomalyRemediator policies + status)
- `AutonomousAction` (AutonomousAgent workflows + decisions)

This gives them:
- A transparent view of **what policy is configured**
- The **current status** and recent changes
- A path to “trust but verify” without reading YAML dumps

### K8sGPT output (independent diagnostics)

If you enable K8sGPT in an `AnomalyAction`, the operator stores the diagnostic text in:
- `.status.k8sgptAnalysis`

Customers can also validate K8sGPT independently via:
- `kubectl logs -n k8sgpt deploy/k8sgpt-operator -f`
- `kubectl exec -n k8sgpt deploy/k8sgpt-operator -- k8sgpt analyze --explain ...`

See `aiops/diagnostics/K8SGPT-TESTING.md`.

### Monitoring links

Headlamp is a great “front door,” but dashboards still live in Grafana. Most teams add:
- A link to Grafana (ML forecasting/anomalies/SLO)
- A link to Hubble UI (eBPF network flows), if enabled

## Security model (customer-friendly)

Headlamp does **not** need cluster-admin to be useful.
- End users authenticate with their own Kubernetes identity (OIDC/token) and only see what RBAC allows.
- The Headlamp service account is only granted minimal permissions for token validation/RBAC checks.


