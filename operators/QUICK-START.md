# Quick Start - Local CI

## One-Time Setup

```bash
# Install tools (macOS)
brew install kubeconform kube-linter

# Or install golangci-lint globally (optional - Makefile will auto-install)
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.61.0
```

## Daily Usage

### Single Operator
```bash
cd operators/anomaly-remediator

# Full CI (before pushing)
make local-ci

# Quick check (during development)
make quick-check
```

### All Operators
```bash
cd operators

# Run CI for all operators
./run-local-ci.sh

# Run CI for specific operator
./run-local-ci.sh anomaly-remediator
```

## What Gets Checked

- ✅ **Linting** - golangci-lint static analysis
- ✅ **CRD Generation** - Ensures manifests are up to date
- ✅ **Schema Validation** - kubeconform validates Kubernetes API schemas
- ✅ **Best Practices** - kube-linter checks for security and best practices
- ✅ **Tests** - Unit tests with coverage

## Time Estimates

- `make quick-check`: ~10 seconds
- `make local-ci`: ~30 seconds
- `./run-local-ci.sh`: ~2 minutes (all operators)

## Troubleshooting

**golangci-lint not found?**  
The Makefile auto-installs it. First run will download it.

**kubeconform/kube-linter not found?**  
```bash
brew install kubeconform kube-linter
```

**CRD validation fails?**  
```bash
make manifests  # Generate CRDs first
```

See [LOCAL-CI.md](./LOCAL-CI.md) for detailed documentation.

