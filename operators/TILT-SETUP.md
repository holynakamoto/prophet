# Tilt Setup for Live Operator Development

Tilt provides **instant feedback** for Kubernetes operator development: code changes → auto-rebuild → live pod update → test immediately.

## Prerequisites

```bash
# Install Tilt
brew install tilt  # macOS
# or: https://docs.tilt.dev/install.html

# Create local Kind cluster (recommended)
brew install kind
kind create cluster --name prophet-local

# Or use existing cluster
# kubectl config current-context  # Verify your context
```

## Quick Start

### Single Operator Development

```bash
cd operators/anomaly-remediator
tilt up
```

Open http://localhost:10350 → Watch your operator:
- Code changes → Image rebuilds → Pod updates automatically
- View logs, resources, and metrics in real-time
- Apply test CRs and see reconciliation instantly

### Multi-Operator Development

From repo root:

```bash
# Start all operators
tilt up --file operators/Tiltfile -- --operator=anomaly-remediator
tilt up --file operators/Tiltfile -- --operator=predictive-scaler
# ... etc

# Or use Tilt's multi-resource setup (see advanced section)
```

## How It Works

1. **Code Change**: Edit `controllers/anomalyaction_controller.go`
2. **Auto-Build**: Tilt detects change → rebuilds Docker image
3. **Live Update**: Updates pod with new image (or hot-reloads binary)
4. **Instant Test**: Apply test CR → See reconciliation in logs

## Tiltfile Locations

- **Root Tiltfile**: `operators/Tiltfile` - Multi-operator support
- **Per-Operator**: `operators/<operator>/Tiltfile` - Single operator focus

## Common Commands

```bash
tilt up              # Start Tilt (opens UI)
tilt down            # Stop and clean up
tilt logs            # Stream logs
tilt trigger <resource>  # Manually trigger rebuild
tilt args --operator=autonomous-agent  # Switch operator
```

## Advanced: Multi-Operator Setup

Create `operators/multi-operator.Tiltfile`:

```starlark
# Run all operators in one Tilt instance
load('./anomaly-remediator/Tiltfile', 'setup_operator')
load('./predictive-scaler/Tiltfile', 'setup_operator')
load('./slo-enforcer/Tiltfile', 'setup_operator')
load('./autonomous-agent/Tiltfile', 'setup_operator')
```

## Integration with Local CI

```bash
# Before starting Tilt
make local-ci        # Ensure code is clean

# During Tilt development
# Tilt handles rebuilds automatically

# Before committing
make local-ci-full   # Full CI + image scan
```

## Troubleshooting

**Tilt not detecting changes?**
```bash
tilt down
tilt up  # Restart
```

**Image not updating?**
```bash
tilt trigger docker_build
```

**Port conflicts?**
Edit Tiltfile `port_forwards` to use different ports.

**Kind cluster issues?**
```bash
kind delete cluster --name prophet-local
kind create cluster --name prophet-local
```

## Benefits

✅ **Instant feedback** - See changes in seconds, not minutes  
✅ **Real cluster** - Test in actual Kubernetes, not mocks  
✅ **Live logs** - Watch reconciliation happen in real-time  
✅ **Resource view** - See all CRs, pods, events in one UI  
✅ **Zero cost** - Everything runs locally  

## Next Steps

1. Install Tilt and create Kind cluster
2. Run `tilt up` in an operator directory
3. Make a code change → Watch the magic happen! ✨

