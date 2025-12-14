# K8sGPT Extension for Rancher

A Rancher UI Extension that displays K8sGPT diagnostic results directly in the Rancher dashboard.

## Features

- **K8sGPT Results List**: View all K8sGPT diagnostic results across your clusters
- **Detailed View**: Drill into individual results with AI-powered explanations
- **Severity Badges**: Visual indicators for error, warning, and healthy states
- **Integrated UI**: Native Rancher look and feel

## Requirements

- Rancher 2.9.0+
- UI Extensions 3.0.0+
- K8sGPT operator installed in your cluster

## Building

```bash
# Install dependencies
npm install --legacy-peer-deps

# Build the extension
npm run build

# Output is in dist-pkg/k8sgpt-1.0.0/
```

## Deployment Options

### Option 1: Developer Load (Fastest for Testing)

1. Build the extension:
   ```bash
   npm run build
   ```

2. Serve the built packages:
   ```bash
   npm run serve-pkgs
   ```
   This serves on http://localhost:4500

3. In Rancher UI:
   - Go to **☰ > Local > Extensions**
   - Click **⋮ > Enable Extension Developer Features**
   - Click **Developer Load**
   - Enter: `http://localhost:4500/k8sgpt-1.0.0/k8sgpt-1.0.0.umd.min.js`
   - Click **Load**

4. The K8sGPT extension should now appear in the left sidebar.

### Option 2: Helm Chart Distribution

1. Package the Helm chart:
   ```bash
   helm package charts/k8sgpt-extension
   ```

2. Publish to a Helm repository (GitHub Pages, OCI, etc.)

3. In Rancher UI:
   - Go to **☰ > Local > Extensions**
   - Add your Helm repository
   - Install the extension from the catalog

### Option 3: Direct Extension Package

1. Create a tarball:
   ```bash
   cd dist-pkg/k8sgpt-1.0.0
   tar -czf ../k8sgpt-1.0.0.tgz .
   ```

2. Host the tarball and serve it via HTTPS

3. Add as an extension repository in Rancher

## Development

```bash
# Start development server
npm run dev

# Clean build artifacts
npm run clean
```

## Project Structure

```
rancher-k8sgpt-extension/
├── pkg/k8sgpt/              # Extension source code
│   ├── index.js             # Main entry point
│   ├── product.js           # Product registration (sidebar)
│   ├── list/                # List view component
│   ├── detail/              # Detail view component
│   ├── components/          # Shared components
│   └── models/              # Data models
├── charts/k8sgpt-extension/ # Helm chart for distribution
├── dist-pkg/                # Build output (generated)
└── package.json
```

## K8sGPT Setup

To see results in the extension, you need K8sGPT installed and running:

```bash
# Install K8sGPT operator
kubectl apply -f https://raw.githubusercontent.com/k8sgpt-ai/k8sgpt-operator/main/config/crd/bases/core.k8sgpt.ai_results.yaml

# The operator will create Result CRDs that this extension displays
```

## Compatibility

| Component | Version |
|-----------|---------|
| Rancher | 2.9.0+ |
| UI Extensions | 3.0.0+ |
| Node.js | 16+ (tested with 25.x using --openssl-legacy-provider) |
| @rancher/shell | 2.0.3 |

## Known Issues

- Build requires `NODE_OPTIONS=--openssl-legacy-provider` on Node.js 17+ (already configured in package.json scripts)
- Vue 2 and webpack 4 are used (per Rancher 2.9 requirements)

## License

Apache-2.0

