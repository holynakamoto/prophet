# Quick Reference - Prophet Operators

## Daily Commands

### Local CI
```bash
# Fast feedback (during development)
make quick-check          # ~10 seconds

# Full CI (before committing)
make local-ci             # ~30 seconds

# Full CI + image scan (before pushing)
make local-ci-full        # ~60 seconds

# All operators
./run-local-ci.sh         # Standard CI
./run-local-ci.sh --full  # With images
```

### Image Operations
```bash
make build-image    # Build Docker image
make scan-image     # Security scan with Trivy
```

### Live Development (Tilt)
```bash
# Start Tilt
tilt up

# Access UI
open http://localhost:10350

# Stop Tilt
tilt down
```

## Prerequisites Checklist

- [ ] Go 1.21+
- [ ] Docker
- [ ] kubectl
- [ ] golangci-lint (auto-installs)
- [ ] kubeconform (optional, for CRD validation)
- [ ] kube-linter (optional, for CRD validation)
- [ ] trivy (optional, for image scanning)
- [ ] tilt (optional, for live development)
- [ ] kind (optional, for local cluster)

## Install Optional Tools

```bash
# macOS
brew install kubeconform kube-linter trivy tilt kind

# Verify
kubeconform --version
kube-linter version
trivy --version
tilt version
kind version
```

## Troubleshooting

**Lint errors?**
```bash
go mod tidy
make generate
```

**Build fails?**
```bash
make clean  # If exists
make build
```

**Tilt not working?**
```bash
kubectl config current-context  # Check cluster
tilt down && tilt up            # Restart
```

**Image scan fails?**
```bash
docker ps  # Check Docker is running
trivy image <image-name>  # Manual scan
```

## File Locations

- **Makefiles**: `operators/<operator>/Makefile`
- **Tiltfiles**: `operators/<operator>/Tiltfile` or `operators/Tiltfile`
- **Test CRs**: `operators/<operator>/config/samples/`
- **CI Scripts**: `operators/run-local-ci.sh`

## Workflow Examples

### Standard Development
```bash
# 1. Make changes
vim controllers/mycontroller.go

# 2. Quick check
make quick-check

# 3. Full CI before commit
make local-ci

# 4. Commit & push
git commit -m "Fix controller logic"
git push
```

### With Image Testing
```bash
# 1. Full CI with image
make local-ci-full

# 2. Test image locally
docker run --rm <image-name> --help

# 3. Push if all good
make docker-push
```

### Live Development
```bash
# 1. Start Tilt
tilt up

# 2. Make changes (auto-rebuilds)
vim controllers/mycontroller.go

# 3. Apply test CR
kubectl apply -f config/samples/

# 4. Watch logs in Tilt UI
# 5. Stop when done
tilt down
```

## See Also

- [LOCAL-CI.md](./LOCAL-CI.md) - Detailed CI setup
- [ENHANCED-CI.md](./ENHANCED-CI.md) - Image building & scanning
- [TILT-SETUP.md](./TILT-SETUP.md) - Live development guide
- [README.md](./README.md) - General documentation

