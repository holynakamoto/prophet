# Contributing to Prophet

Thank you for your interest in contributing to Prophet! This guide will help you get started.

## Development Setup

### Prerequisites

| Tool | Version | Purpose |
|------|---------|---------|
| Go | 1.22+ | Operator development |
| Docker | Latest | Container builds |
| kubectl | 1.28+ | Cluster interaction |
| kind | 0.20+ | Local Kubernetes cluster |
| Tilt | Latest | Live operator development |
| kubeconform | Latest | Manifest validation |
| golangci-lint | 1.61+ | Go linting |

### Quick Setup

```bash
# Clone the repo
git clone https://github.com/YOUR-USERNAME/prophet.git
cd prophet

# Install Go tools (from any operator directory)
cd operators/anomaly-remediator
make controller-gen kustomize golangci-lint
cd ../..

# Create local cluster
kind create cluster --name prophet-dev

# Deploy monitoring stack (optional, for full testing)
kubectl apply -f monitoring/prometheus/
kubectl apply -f monitoring/grafana/

# Run an operator locally
cd operators/anomaly-remediator
make run
```

### Using Tilt for Live Development

Tilt provides hot-reload for operators:

```bash
cd operators
tilt up
```

Open http://localhost:10350 to see all operators running with live logs.

## Project Structure

```
prophet/
├── operators/           # Go operators (main code)
│   ├── anomaly-remediator/
│   ├── predictive-scaler/
│   ├── slo-enforcer/
│   └── ...
├── clusters/            # Kustomize overlays
├── monitoring/          # Prometheus, Grafana configs
├── demo/                # Demo scripts and manifests
├── resilience/          # Chaos experiments
├── aiops/               # AI/ML integration configs
└── apps/                # ArgoCD app definitions
```

## Workflow

### 1. Create a Branch

```bash
git checkout -b feature/my-feature
# or
git checkout -b fix/my-bugfix
```

Branch naming:
- `feature/` - New functionality
- `fix/` - Bug fixes
- `docs/` - Documentation only
- `refactor/` - Code restructuring

### 2. Make Changes

For operator changes:
```bash
cd operators/anomaly-remediator

# Edit code...

# Regenerate if you changed CRD types
make generate manifests

# Run tests
make test

# Lint
make lint
```

### 3. Test Locally

```bash
# Full local CI pipeline
make local-ci

# Or run specific checks
make lint
make test
make validate-crds
```

### 4. Commit

Write clear commit messages:

```
feat(anomaly-remediator): add Prometheus query support

- Add prometheus client to controller
- Query configurable metrics for anomaly detection
- Fall back to pod status if Prometheus unavailable

Closes #123
```

Prefixes:
- `feat` - New feature
- `fix` - Bug fix
- `docs` - Documentation
- `refactor` - Code change that doesn't add features or fix bugs
- `test` - Adding tests
- `chore` - Maintenance tasks

### 5. Push and Create PR

```bash
git push origin feature/my-feature
```

Then create a Pull Request on GitHub.

## Pull Request Guidelines

### Before Submitting

- [ ] All tests pass (`make test`)
- [ ] Linting passes (`make lint`)
- [ ] CRDs validate (`make validate-crds`)
- [ ] Documentation updated (if applicable)
- [ ] Commit messages follow conventions

### PR Description Template

```markdown
## What

Brief description of the change.

## Why

Why this change is needed.

## How

How the change works.

## Testing

How you tested this change.

## Checklist

- [ ] Tests added/updated
- [ ] Documentation updated
- [ ] IMPLEMENTATION-PROGRESS.md updated (if applicable)
```

## Coding Standards

### Go

- Follow [Effective Go](https://go.dev/doc/effective_go)
- Use `gofmt` (enforced by CI)
- Write table-driven tests
- Add godoc comments for exported symbols

### YAML/Kubernetes Manifests

- Use 2-space indentation
- Add comments for non-obvious configurations
- Validate with `kubeconform`

### Commits

- Keep commits atomic (one logical change per commit)
- Write descriptive commit messages
- Reference issues when applicable

## Running CI Locally

Before pushing, run the full CI locally:

```bash
# Per-operator CI
cd operators/anomaly-remediator
make local-ci

# Full CI including image build + security scan
make local-ci-full
```

## Adding a New Operator

1. Scaffold with kubebuilder:
   ```bash
   cd operators
   mkdir my-operator && cd my-operator
   kubebuilder init --domain prophet.io --repo github.com/prophet-aiops/my-operator
   kubebuilder create api --group aiops --version v1alpha1 --kind MyResource
   ```

2. Copy Makefile patterns from existing operators

3. Add deployment manifest to `clusters/common/aiops/operators/`

4. Add to operators Tiltfile

5. Document in `operators/README.md`

## Chaos Experiments

When adding chaos experiments:

1. Create in `resilience/chaos-experiments/` or `demo/remediation-chaos/`

2. Document:
   - Preconditions (what must be deployed)
   - Expected behavior (what signals indicate pass/fail)
   - Cleanup steps

3. Add CI validation if appropriate

## Getting Help

- Open an issue for bugs or feature requests
- Check existing issues before creating new ones
- Join discussions in PRs

## License

By contributing, you agree that your contributions will be licensed under the Apache 2.0 License.

