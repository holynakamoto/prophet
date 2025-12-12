#!/bin/bash
# Setup Go modules for all operators
# Run this script to initialize and download dependencies

set -e

echo "Setting up Go modules for Prophet operators..."

for operator in anomaly-remediator predictive-scaler slo-enforcer autonomous-agent; do
    if [ -d "$operator" ]; then
        echo "Setting up $operator..."
        cd "$operator"
        if [ -f "go.mod" ]; then
            go mod tidy
            echo "✓ $operator dependencies downloaded"
        else
            echo "⚠ $operator has no go.mod file"
        fi
        cd ..
    fi
done

echo "Done! All operator dependencies are ready."

