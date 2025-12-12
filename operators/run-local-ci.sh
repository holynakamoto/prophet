#!/bin/bash
# Run local CI for all operators
# Usage: ./run-local-ci.sh [operator-name] [--full]
#   operator-name: Run CI for specific operator only
#   --full: Include image build and security scan

set -e

OPERATORS=("anomaly-remediator" "predictive-scaler" "slo-enforcer" "autonomous-agent")
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
FULL_CI=false

# Parse arguments
if [ "$1" == "--full" ]; then
    FULL_CI=true
    shift
elif [ "$2" == "--full" ]; then
    FULL_CI=true
fi

if [ -n "$1" ]; then
    # Run CI for specific operator
    if [[ " ${OPERATORS[@]} " =~ " ${1} " ]]; then
        echo "Running local CI for $1..."
        cd "$SCRIPT_DIR/$1"
        if [ "$FULL_CI" = true ]; then
            make local-ci-full
        else
            make local-ci
        fi
        echo "✅ $1 CI passed!"
    else
        echo "Error: Unknown operator '$1'"
        echo "Available operators: ${OPERATORS[*]}"
        exit 1
    fi
else
    # Run CI for all operators
    echo "Running local CI for all operators..."
    echo ""
    
    FAILED=()
    for op in "${OPERATORS[@]}"; do
        echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
        echo "🔍 $op"
        echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    cd "$SCRIPT_DIR/$op"
    
        if [ "$FULL_CI" = true ]; then
            CI_CMD="local-ci-full"
        else
            CI_CMD="local-ci"
        fi
        
        if make $CI_CMD; then
            echo "✅ $op passed!"
        else
            echo "❌ $op failed!"
            FAILED+=("$op")
        fi
        echo ""
    done
    
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    if [ ${#FAILED[@]} -eq 0 ]; then
        echo "✅ All operators passed local CI!"
        exit 0
    else
        echo "❌ Failed operators: ${FAILED[*]}"
        exit 1
    fi
fi

