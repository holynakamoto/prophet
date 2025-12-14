# Testing K8sGPT Integration in Prophet

This guide helps you verify that **K8sGPT** is working with your Prophet setup.

Notes:
- Prophet includes a simple K8sGPT deployment manifest at `clusters/common/aiops/k8sgpt/k8sgpt-operator.yaml` (runs `k8sgpt server` in-cluster).
- Some Prophet-side consumers of K8sGPT output (e.g. certain MCP tool calls) may still be **placeholders**; the steps below validate that K8sGPT itself is working and reachable.

### Quick test (CLI — no cluster changes)

This is the fastest way to confirm your **AI backend connectivity** and that K8sGPT can analyze your cluster through your current `kubectl` context.

1) Install K8sGPT CLI:

```bash
brew tap k8sgpt-ai/k8sgpt
brew install k8sgpt
```

2) Configure an AI backend (recommended: local Ollama via OpenAI-compatible API).

- Start Ollama and load a model:

```bash
ollama run phi3
```

- Configure K8sGPT to use the OpenAI-compatible endpoint that Ollama exposes:

```bash
k8sgpt auth add --backend localai --model phi3 --baseurl http://localhost:11434/v1
```

3) Run analysis:

```bash
k8sgpt analyze --explain
k8sgpt analyze --explain --namespace default
k8sgpt analyze --explain --filter Pod
```

### In-cluster test (Prophet manifest)

This validates the **in-cluster K8sGPT service** Prophet ships, and is the closest thing to “Prophet-style integration” today.

1) Deploy K8sGPT (Prophet manifest):

```bash
kubectl apply -f clusters/common/aiops/k8sgpt/k8sgpt-operator.yaml
```

2) (Optional) Provide an API key secret for cloud backends.

The manifest expects:
- Namespace: `k8sgpt`
- Secret: `k8sgpt-secrets`
- Key: `openai-api-key`

```bash
kubectl create secret generic k8sgpt-secrets -n k8sgpt \
  --from-literal=openai-api-key='YOUR_KEY_HERE' \
  --dry-run=client -o yaml | kubectl apply -f -
```

If you are using a local backend instead, update `clusters/common/aiops/k8sgpt/k8sgpt-operator.yaml` to set `K8SGPT_PROVIDER=localai` and configure the base URL/model accordingly.

3) Create a broken workload to diagnose:

```bash
kubectl apply -f - <<'EOF'
apiVersion: v1
kind: Pod
metadata:
  name: broken-pod
  namespace: default
spec:
  containers:
  - name: broken
    image: nginx:invalid-tag
EOF
```

4) Trigger a manual analysis from inside the K8sGPT pod (simple smoke test):

```bash
kubectl exec -n k8sgpt deploy/k8sgpt-operator -- \
  k8sgpt analyze --explain --namespace default --config /etc/k8sgpt/config.yaml
```

Expected: output referencing `ErrImagePull` / `ImagePullBackOff` with a plain-English explanation.

### Alert integration (optional)

Prophet also includes an example Alertmanager webhook integration manifest:

- `aiops/diagnostics/k8sgpt-alert-integration.yaml`

You can apply it as a starting point, then confirm webhook delivery by checking K8sGPT logs:

```bash
kubectl logs -n k8sgpt deploy/k8sgpt-operator -f
```

### Success indicators

- K8sGPT returns plain-English explanations (CLI or in-cluster run).
- No backend connectivity errors (provider auth/config is valid).
- Common issues like `CrashLoopBackOff`, `ErrImagePull`, and OOM are explained accurately.


