# Prophet Operators

This directory contains custom Golang Kubernetes operators built with Kubebuilder.

## Operators

- **anomaly-remediator**: Detects anomalies and performs automated remediation
- **predictive-scaler**: Uses Grafana ML forecasts to proactively scale Karpenter NodePools
- **slo-enforcer**: Monitors SLOs and enforces policies when violations occur
- **autonomous-agent**: LLM-powered autonomous remediation with MCP server (v5)

## Setup

### Prerequisites

- Go 1.21 or later
- kubebuilder (optional, for code generation)
- Access to a Kubernetes cluster (for testing)

### Initialize Dependencies

Run the setup script to download all Go module dependencies:

```bash
./setup-go-modules.sh
```

Or manually for each operator:

```bash
cd anomaly-remediator
go mod tidy
cd ../predictive-scaler
go mod tidy
cd ../slo-enforcer
go mod tidy
cd ../autonomous-agent
go mod tidy
```

### Build

Build an operator:

```bash
cd anomaly-remediator
make build
# or
go build -o bin/manager cmd/main.go
```

### Run Locally

```bash
# Install CRDs
make install

# Run controller (requires kubeconfig)
make run
```

### Docker Build

```bash
make docker-build IMG=ghcr.io/prophet-aiops/prophet-<operator-name>:latest
make docker-push IMG=ghcr.io/prophet-aiops/prophet-<operator-name>:latest
```

## Development

### Generate Code

```bash
# Generate CRDs and RBAC
make manifests

# Generate DeepCopy methods
make generate
```

### Testing

```bash
# Run unit tests
make test

# Run with coverage
go test ./... -coverprofile cover.out
```

## Local CI

Run the same CI checks locally that run in GitHub Actions:

```bash
# Full CI pipeline (mirrors GitHub Actions)
make local-ci

# Full CI + image build & security scan
make local-ci-full

# Quick check (lint + tests, ~10 seconds)
make quick-check

# Individual checks
make lint          # Static analysis with golangci-lint
make test          # Unit tests
make validate-crds # Validate CRDs with kubeconform + kube-linter
make build-image   # Build Docker image
make scan-image    # Scan image with Trivy
```

See [LOCAL-CI.md](./LOCAL-CI.md) for detailed setup instructions.
See [ENHANCED-CI.md](./ENHANCED-CI.md) for image building & scanning.
See [TILT-SETUP.md](./TILT-SETUP.md) for live development with Tilt.

## Linting

The linter may show errors about missing packages until `go mod tidy` is run. This is expected - the dependencies need to be downloaded first.

After running `go mod tidy`, the lint errors should resolve.

Run linting with:
```bash
make lint  # Uses golangci-lint with .golangci.yml config
```

## Deployment

See the main README.md for deployment instructions. Operators are deployed via:

```bash
kubectl apply -f ../clusters/common/aiops/operators/<operator-name>.yaml
```

