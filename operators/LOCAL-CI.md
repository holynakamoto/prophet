# Local CI Setup Guide

This guide shows you how to run the same CI checks locally that run in GitHub Actions, giving you fast, zero-cost feedback before pushing.

## Prerequisites (One-Time Setup)

Install these tools on your machine:

```bash
# golangci-lint (static analysis) - will be auto-installed by Makefile
# But you can install globally if preferred:
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.61.0

# kubeconform (schema validation for Kubernetes manifests)
brew install kubeconform  # macOS
# or download from https://github.com/yannh/kubeconform/releases

# kube-linter (best practices linting)
brew install kube-linter  # macOS
# or: go install golang.stackrox.io/kube-linter/cmd/kube-linter@latest

# Optional: act (to run full GitHub Actions workflows locally)
brew install act
```

## Quick Start

### For a Single Operator

```bash
cd operators/anomaly-remediator

# Full CI pipeline (mirrors GitHub Actions)
make local-ci

# Quick check (lint + tests only, ~10 seconds)
make quick-check

# Individual checks
make lint          # Static analysis
make test          # Unit tests
make manifests     # Generate CRDs
make validate-crds # Validate CRDs with kubeconform + kube-linter
```

### For All Operators

```bash
# Run local CI for all operators
cd operators
for op in anomaly-remediator predictive-scaler slo-enforcer autonomous-agent; do
  echo "Running CI for $op..."
  cd $op && make local-ci && cd ..
done
```

## Recommended Daily Workflow

```bash
# 1. Make code changes
vim controllers/mycontroller.go

# 2. Quick feedback loop (fast, ~10 seconds)
make quick-check

# 3. Before committing CRD changes
make manifests && make validate-crds

# 4. Before pushing (full confidence, ~30 seconds)
make local-ci
```

## What Each Target Does

### `make local-ci`
Runs the complete CI pipeline:
- ✅ `lint` - golangci-lint static analysis
- ✅ `manifests` - Generate CRDs and RBAC
- ✅ `validate-crds` - Schema validation (kubeconform) + best practices (kube-linter)
- ✅ `test` - Unit tests with coverage

### `make quick-check`
Fast feedback loop:
- ✅ `lint` - Static analysis
- ✅ `test` - Unit tests

### `make lint`
Runs golangci-lint with configuration from `.golangci.yml`:
- Error checking
- Code quality
- Security issues
- Style violations

### `make validate-crds`
Validates generated CRDs:
- **kubeconform**: Kubernetes API schema validation
- **kube-linter**: Best practices and security checks

## Running Full GitHub Actions Locally with `act`

If you want to mirror the **exact** GitHub Actions experience:

```bash
# Install act
brew install act

# List available workflows
cd /Users/nickmoore/prophet
act -l

# Run the validation workflow
act pull_request -W .github/workflows/ci-validate.yaml

# Run with debug output
act -v pull_request
```

**Note**: `act` uses Docker to simulate GitHub runners. It's slower than `make local-ci` but catches environment differences.

## Troubleshooting

### golangci-lint not found
The Makefile will auto-install it to `bin/golangci-lint`. If you want it globally:
```bash
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.61.0
```

### kubeconform/kube-linter not found
Install with:
```bash
brew install kubeconform kube-linter
```

### CRD validation fails
Make sure you've generated CRDs first:
```bash
make manifests
```

### Tests fail
Ensure dependencies are downloaded:
```bash
go mod tidy
```

## CI Configuration Files

- **`.golangci.yml`** - golangci-lint configuration (shared across operators)
- **`Makefile`** - Local CI targets in each operator directory
- **`.github/workflows/ci-validate.yaml`** - GitHub Actions workflow (mirrors local CI)

## Benefits

✅ **Instant feedback** - Most checks complete in <10 seconds  
✅ **Zero cost** - No GitHub Actions minutes consumed  
✅ **Full confidence** - Same checks that run in CI  
✅ **Faster iteration** - Catch issues before pushing  
✅ **Professional workflow** - Industry-standard tooling

## Next Steps

1. Install prerequisites (one-time)
2. Run `make local-ci` before every push
3. Use `make quick-check` during development
4. Enjoy faster, cheaper, better development! 🚀

