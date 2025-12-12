# Enhanced Local CI - Image Building & Scanning

## Overview

The enhanced local CI adds Docker image building and Trivy security scanning to catch vulnerabilities before pushing to production.

## Prerequisites

```bash
# Install Trivy (one-time)
brew install trivy  # macOS
# or: curl -sfL https://raw.githubusercontent.com/aquasecurity/trivy/main/contrib/install.sh | sh -s -- -b /usr/local/bin
```

## New Makefile Targets

### Image Building

```bash
# Build operator image locally
make build-image

# Scan image for vulnerabilities
make scan-image

# Full CI including image build and scan
make local-ci-full
```

### What Gets Scanned

- Base image vulnerabilities
- Go dependencies (if included in image)
- System packages
- Critical and High severity issues only

## Usage

### During Development

```bash
# Standard CI (fast, no image)
make local-ci

# Full CI with image (before pushing)
make local-ci-full
```

### Batch Testing

```bash
# All operators, standard CI
./run-local-ci.sh

# All operators, full CI (with images)
./run-local-ci.sh --full

# Specific operator, full CI
./run-local-ci.sh anomaly-remediator --full
```

## Trivy Scan Output

Trivy will:
- ✅ **Pass silently** if no HIGH/CRITICAL vulnerabilities found
- ❌ **Fail loudly** if critical issues detected
- ⚠️ **Skip gracefully** if Trivy not installed (non-blocking)

## Example Output

```bash
$ make scan-image
Scanning image ghcr.io/prophet-aiops/prophet-anomaly-remediator:latest with Trivy...
2024-01-15T10:30:00.000Z  INFO    Detected OS: alpine
2024-01-15T10:30:00.000Z  INFO    Detecting Alpine vulnerabilities...
2024-01-15T10:30:01.000Z  INFO    Number of language-specific files: 1
✅ No critical vulnerabilities found!
```

## Integration with GitHub Actions

This mirrors the exact same checks that run in `.github/workflows/ci-operator-build.yaml`:
- Same Trivy version
- Same severity levels (HIGH, CRITICAL)
- Same exit codes

**Result**: If `make local-ci-full` passes, your GitHub Actions will too.

## Performance

- **First build**: ~30-60 seconds (downloads base image)
- **Cached builds**: ~5-10 seconds
- **Trivy scan**: ~10-20 seconds
- **Total**: ~40-80 seconds for full CI

## Troubleshooting

**Trivy not found?**
```bash
brew install trivy
# Verify: trivy --version
```

**Docker not running?**
```bash
# Start Docker Desktop or Docker daemon
docker ps  # Should work
```

**Image build fails?**
```bash
# Check Dockerfile exists
ls Dockerfile

# Build manually to see errors
docker build -t test-image .
```

**Scan finds vulnerabilities?**
```bash
# Review findings
trivy image --severity HIGH,CRITICAL <image-name>

# Update base image in Dockerfile
# Rebuild and rescan
```

## Best Practices

1. **Run `local-ci-full` before pushing** - Catch image issues early
2. **Keep base images updated** - Alpine/scratch images get security patches
3. **Review Trivy output** - Some false positives possible, but most are real
4. **Use in CI/CD** - Same checks run in GitHub Actions

## Next: Tilt Integration

For live development with instant feedback, see [TILT-SETUP.md](./TILT-SETUP.md).

